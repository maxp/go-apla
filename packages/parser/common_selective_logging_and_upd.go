// Copyright 2016 The go-daylight Authors
// This file is part of the go-daylight library.
//
// The go-daylight library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-daylight library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-daylight library. If not, see <http://www.gnu.org/licenses/>.

package parser

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/AplaProject/go-apla/packages/consts"
	"github.com/AplaProject/go-apla/packages/converter"
	"github.com/AplaProject/go-apla/packages/model"

	log "github.com/sirupsen/logrus"
)

// selectiveLoggingAndUpd changes DB and writes all DB changes for rollbacks
// do not use for comments
func (p *Parser) selectiveLoggingAndUpd(fields []string, ivalues []interface{}, table string, whereFields, whereValues []string, generalRollback bool) (int64, string, error) {
	logger := p.GetLogger()
	var (
		tableID string
		err     error
		cost    int64
	)

	if generalRollback && p.BlockData == nil {
		logger.WithFields(log.Fields{"type": consts.EmptyObject}).Error("Block is undefined")
		return 0, ``, fmt.Errorf(`It is impossible to write to DB when Block is undefined`)
	}

	isBytea := getBytea(table)
	for i, v := range ivalues {
		if len(fields) > i && isBytea[fields[i]] {
			var vlen int
			switch v.(type) {
			case []byte:
				vlen = len(v.([]byte))
			case string:
				if vbyte, err := hex.DecodeString(v.(string)); err == nil {
					ivalues[i] = vbyte
					vlen = len(vbyte)
				} else {
					vlen = len(v.(string))
				}
			}
			if vlen > 64 {
				if isCustom, err := IsCustomTable(table); err != nil {
					return 0, ``, err
				} else if isCustom {
					log.WithFields(log.Fields{"type": consts.ParameterExceeded}).Error("hash value cannot be larger than 64 bytes")
					return 0, ``, fmt.Errorf(`hash value cannot be larger than 64 bytes`)
				}
			}
		}
	}

	values := converter.InterfaceSliceToStr(ivalues)

	addSQLFields := p.AllPkeys[table]
	if len(addSQLFields) > 0 {
		addSQLFields += `,`
	}
	for i, field := range fields {
		field = strings.TrimSpace(field)
		fields[i] = field
		if field[:1] == "+" || field[:1] == "-" {
			addSQLFields += field[1:len(field)] + ","
		} else if strings.HasPrefix(field, `timestamp `) {
			addSQLFields += field[len(`timestamp `):] + `,`
		} else {
			addSQLFields += field + ","
		}
	}

	addSQLWhere := ""
	if whereFields != nil && whereValues != nil {
		for i := 0; i < len(whereFields); i++ {
			if val := converter.StrToInt64(whereValues[i]); val != 0 {
				addSQLWhere += whereFields[i] + "= " + whereValues[i] + " AND "
			} else {
				addSQLWhere += whereFields[i] + "= '" + whereValues[i] + "' AND "
			}
		}
	}
	if len(addSQLWhere) > 0 {
		addSQLWhere = " WHERE " + addSQLWhere[0:len(addSQLWhere)-5]
	}

	// if there is something to log
	selectQuery := `SELECT ` + addSQLFields + ` rb_id FROM "` + table + `" ` + addSQLWhere
	selectCost, err := model.GetQueryTotalCost(selectQuery)
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err, "query": selectQuery}).Error("getting query total cost")
		return 0, tableID, err
	}
	logData, err := model.GetOneRowTransaction(p.DbTransaction, selectQuery).String()
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err, "query": selectQuery}).Error("getting one row transaction")
		return 0, tableID, err
	}
	cost += selectCost

	if whereFields != nil && len(logData) > 0 {
		jsonMap := make(map[string]string)
		for k, v := range logData {
			if k == p.AllPkeys[table] {
				continue
			}
			if (isBytea[k] || converter.InSliceString(k, []string{"hash", "tx_hash", "pub", "tx_hash", "public_key_0", "node_public_key"})) && v != "" {
				jsonMap[k] = string(converter.BinToHex([]byte(v)))
			} else {
				jsonMap[k] = v
			}
			if k == "rb_id" {
				k = "prev_rb_id"
			}
			if k[:1] == "+" || k[:1] == "-" {
				addSQLFields += k[1:len(k)] + ","
			} else if strings.HasPrefix(k, `timestamp `) {
				addSQLFields += k[len(`timestamp `):] + `,`
			} else {
				addSQLFields += k + ","
			}
		}
		jsonData, _ := json.Marshal(jsonMap)
		if err != nil {
			logger.WithFields(log.Fields{"type": consts.JSONMarshallError, "error": err}).Error("marshalling rollback info to json")
			return 0, tableID, err
		}
		rollback := &model.Rollback{Data: string(jsonData), BlockID: p.BlockData.BlockID}
		err = rollback.Create(p.DbTransaction)
		if err != nil {
			logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("creating rollback")
			return 0, tableID, err
		}
		addSQLUpdate := ""
		for i := 0; i < len(fields); i++ {
			if isBytea[fields[i]] && len(values[i]) != 0 {
				addSQLUpdate += fields[i] + `=decode('` + hex.EncodeToString([]byte(values[i])) + `','HEX'),`
			} else if fields[i][:1] == "+" {
				addSQLUpdate += fields[i][1:len(fields[i])] + `=` + fields[i][1:len(fields[i])] + `+` + values[i] + `,`
			} else if fields[i][:1] == "-" {
				addSQLUpdate += fields[i][1:len(fields[i])] + `=` + fields[i][1:len(fields[i])] + `-` + values[i] + `,`
			} else if values[i] == `NULL` {
				addSQLUpdate += fields[i] + `= NULL,`
			} else if strings.HasPrefix(fields[i], `timestamp `) {
				addSQLUpdate += fields[i][len(`timestamp `):] + `= to_timestamp('` + values[i] + `'),`
			} else if strings.HasPrefix(values[i], `timestamp `) {
				addSQLUpdate += fields[i] + `= timestamp '` + values[i][len(`timestamp `):] + `',`
			} else {
				addSQLUpdate += fields[i] + `='` + strings.Replace(values[i], `'`, `''`, -1) + `',`
			}
		}
		updateQuery := `UPDATE "` + table + `" SET ` + addSQLUpdate + fmt.Sprintf(` rb_id = '%d'`, rollback.RbID) + addSQLWhere
		updateCost, err := model.GetQueryTotalCost(updateQuery)
		if err != nil {
			logger.WithFields(log.Fields{"type": consts.DBError, "error": err, "query": updateQuery}).Error("getting query total cost for update query")
			return 0, tableID, err
		}
		cost += updateCost
		err = model.Update(p.DbTransaction, table, addSQLUpdate+fmt.Sprintf("rb_id = %d", rollback.RbID), addSQLWhere)
		if err != nil {
			logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting update query")
			return 0, tableID, err
		}
		tableID = logData[p.AllPkeys[table]]
	} else {
		isID := false
		addSQLIns0 := ""
		addSQLIns1 := ""
		for i := 0; i < len(fields); i++ {
			if fields[i] == `id` {
				isID = true
			}
			if fields[i][:1] == "+" || fields[i][:1] == "-" {
				addSQLIns0 += fields[i][1:len(fields[i])] + `,`
			} else if strings.HasPrefix(fields[i], `timestamp `) {
				addSQLIns0 += fields[i][len(`timestamp `):] + `,`
			} else {
				addSQLIns0 += fields[i] + `,`
			}
			// || utils.InSliceString(fields[i], []string{"hash", "tx_hash", "public_key", "public_key_0", "node_public_key"}))
			if isBytea[fields[i]] && len(values[i]) != 0 {
				addSQLIns1 += `decode('` + hex.EncodeToString([]byte(values[i])) + `','HEX'),`
			} else if values[i] == `NULL` {
				addSQLIns1 += `NULL,`
			} else if strings.HasPrefix(fields[i], `timestamp `) {
				addSQLIns1 += `to_timestamp('` + values[i] + `'),`
			} else if strings.HasPrefix(values[i], `timestamp `) {
				addSQLIns1 += `timestamp '` + values[i][len(`timestamp `):] + `',`
			} else {
				addSQLIns1 += `'` + strings.Replace(values[i], `'`, `''`, -1) + `',`
			}
		}
		if whereFields != nil && whereValues != nil {
			for i := 0; i < len(whereFields); i++ {
				if whereFields[i] == `id` {
					isID = true
				}
				addSQLIns0 += `` + whereFields[i] + `,`
				addSQLIns1 += `'` + whereValues[i] + `',`
			}
		}
		addSQLIns0 = addSQLIns0[0 : len(addSQLIns0)-1]
		addSQLIns1 = addSQLIns1[0 : len(addSQLIns1)-1]
		if !isID {
			id, err := model.GetNextID(p.DbTransaction, table)
			if err != nil {
				logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting next id for table")
				return 0, ``, err
			}
			tableID = converter.Int64ToStr(id)
			addSQLIns0 += `,id`
			addSQLIns1 += `,'` + tableID + `'`
		}

		insertQuery := `INSERT INTO "` + table + `" (` + addSQLIns0 + `) VALUES (` + addSQLIns1 + `)`
		insertCost, err := model.GetQueryTotalCost(insertQuery)
		if err != nil {
			logger.WithFields(log.Fields{"type": consts.DBError, "error": err, "query": insertQuery}).Error("getting total query cost for insert query")
			return 0, tableID, err
		}
		cost += insertCost
		err = model.GetDB(p.DbTransaction).Exec(insertQuery).Error
		if err != nil {
			logger.WithFields(log.Fields{"type": consts.DBError, "error": err, "query": insertQuery}).Error("executing insert query")
		}
	}
	if err != nil {
		return 0, tableID, err
	}

	if generalRollback {
		rollbackTx := &model.RollbackTx{
			BlockID:   p.BlockData.BlockID,
			TxHash:    p.TxHash,
			NameTable: table,
			TableID:   tableID,
		}

		err = rollbackTx.Create(p.DbTransaction)
		if err != nil {
			logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("creating rollback tx")
			return 0, tableID, err
		}
	}
	return cost, tableID, nil
}

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

package smart

import (
	"database/sql"
	"fmt"
	"reflect"
	"strings"

	"github.com/AplaProject/go-apla/packages/converter"
	"github.com/AplaProject/go-apla/packages/model"
	"github.com/AplaProject/go-apla/packages/script"
	"github.com/AplaProject/go-apla/packages/templatev2"
	"github.com/AplaProject/go-apla/packages/utils/tx"
)

type SmartContract struct {
	vde     bool
	TxSmart *tx.SmartContract
}

var (
	funcCallsDB = map[string]struct{}{
		"DBSelect": struct{}{},
	}
	extendCost = map[string]int64{
		"EcosystemParam": 10,
	}
)

func getCost(name string) int64 {
	if val, ok := extendCost[name]; ok {
		return val
	}
	return -1
}

func EmbedFuncs(vm *script.VM) {
	vmExtend(vm, &script.ExtendData{Objects: map[string]interface{}{
		"DBSelect":       DBSelect,
		"EcosystemParam": EcosystemParam,
	}, AutoPars: map[string]string{
		`*SmartContract`: `sc`,
	}})
	vmExtendCost(vm, getCost)
	vmFuncCallsDB(vm, funcCallsDB)
}

// DBSelect returns an array of values of the specified columns when there is selection of data 'offset', 'limit', 'where'
func DBSelect(sc *SmartContract, tblname string, columns string, id int64, order string, offset, limit, ecosystem int64,
	where string, params []interface{}) (int64, []interface{}, error) {

	var (
		err  error
		rows *sql.Rows
	)
	/*	if err = checkReport(tblname); err != nil {
		return 0, nil, err
	}*/
	if len(columns) == 0 {
		columns = `*`
	}
	if len(order) == 0 {
		order = `id`
	}
	where = strings.Replace(converter.Escape(where), `$`, `?`, -1)
	if id > 0 {
		where = fmt.Sprintf(`id='%d'`, id)
	}
	if limit == 0 {
		limit = 25
	}
	if limit < 0 || limit > 250 {
		limit = 250
	}
	if ecosystem == 0 {
		ecosystem = sc.TxSmart.StateID
	}
	if tblname[0] < '1' || tblname[0] > '9' || !strings.Contains(tblname, `_`) {
		tblname = fmt.Sprintf(`%d_%s`, ecosystem, tblname)
	}

	rows, err = model.DBConn.Table(tblname).Select(columns).Where(where, params...).Order(order).
		Offset(offset).Limit(limit).Rows()
	if err != nil {
		return 0, nil, err
	}
	defer rows.Close()
	cols, err := rows.Columns()
	if err != nil {
		return 0, nil, err
	}
	values := make([][]byte, len(cols))
	scanArgs := make([]interface{}, len(values))
	for i := range values {
		scanArgs[i] = &values[i]
	}
	result := make([]interface{}, 0, 50)
	for rows.Next() {
		err = rows.Scan(scanArgs...)
		if err != nil {
			return 0, nil, err
		}
		row := make(map[string]string)
		for i, col := range values {
			var value string
			if col != nil {
				value = string(col)
			}
			row[cols[i]] = value
		}
		result = append(result, reflect.ValueOf(row).Interface())
	}
	return 0, result, nil
}

// EcosystemParam returns the value of the specified parameter for the ecosystem
func EcosystemParam(sc *SmartContract, name string) string {
	val, _ := templatev2.StateParam(sc.TxSmart.StateID, name)
	return val
}

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
	"github.com/AplaProject/go-apla/packages/consts"
	"github.com/AplaProject/go-apla/packages/model"
	"github.com/AplaProject/go-apla/packages/utils"

	log "github.com/sirupsen/logrus"
)

func (p *Parser) autoRollback() error {
	logger := p.GetLogger()
	rollbackTx := &model.RollbackTx{}
	txs, err := rollbackTx.GetRollbackTransactions(p.DbTransaction, p.TxHash)
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting rollback transactions")
		return utils.ErrInfo(err)
	}
	for _, tx := range txs {
		err := p.selectiveRollback(tx["table_name"], p.AllPkeys[tx["table_name"]]+"='"+tx["table_id"]+`'`)
		if err != nil {
			return p.ErrInfo(err)
		}
	}
	txForDelete := &model.RollbackTx{TxHash: p.TxHash}
	err = txForDelete.DeleteByHash(p.DbTransaction)
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("deleting rollback transaction by hash")
		return p.ErrInfo(err)
	}
	return nil
}

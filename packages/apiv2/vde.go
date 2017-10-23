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

package apiv2

import (
	"fmt"
	"net/http"

	"github.com/AplaProject/go-apla/packages/converter"
	"github.com/AplaProject/go-apla/packages/model"
)

type vdeCreateResult struct {
	Result bool `json:"result"`
}

func vdeCreate(w http.ResponseWriter, r *http.Request, data *apiData) error {
	if model.IsTable(fmt.Sprintf(`%d_local_tables`, data.state)) {
		return errorAPI(w, `E_LOCALCREATED`, http.StatusBadRequest)
	}
	sp := &model.StateParameter{}
	if err := sp.SetTablePrefix(converter.Int64ToStr(data.state)).GetByName(`founder_account`); err != nil {
		return errorAPI(w, err, http.StatusBadRequest)
	}
	if converter.StrToInt64(sp.Value) != data.wallet {
		return errorAPI(w, `E_PERMISSION`, http.StatusUnauthorized)
	}
	if err := model.ExecSchemaLocalData(int(data.state), data.wallet); err != nil {
		return errorAPI(w, err, http.StatusInternalServerError)
	}
	data.result = vdeCreateResult{Result: true}
	return nil
}

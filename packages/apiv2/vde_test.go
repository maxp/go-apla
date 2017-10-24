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
	"net/url"
	"testing"

	"github.com/AplaProject/go-apla/packages/crypto"
)

func TestVDECreate(t *testing.T) {
	var (
		err error
		ret vdeCreateResult
	)
	if err = keyLogin(1); err != nil {
		t.Error(err)
		return
	}
	if err = sendPost(`vde/create`, nil, &ret); err != nil &&
		err.Error() != `400 {"error": "E_VDECREATED", "msg": "Virtual Dedicated Ecosystem is already created" }` {
		t.Error(err)
		return
	}

	rnd := `rnd` + crypto.RandSeq(6)
	form := url.Values{`Value`: {`contract ` + rnd + ` {
		    data {
				Par string
			}
			action { Test("active",  $Par)}}`}, `Conditions`: {`true`}, `vde`: {`true`}}

	if err := postTx(`NewContract`, &form); err != nil {
		t.Error(err)
		return
	}
	fmt.Println(`OK`)
}

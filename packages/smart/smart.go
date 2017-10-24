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
	"encoding/hex"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/AplaProject/go-apla/packages/converter"
	"github.com/AplaProject/go-apla/packages/model"
	"github.com/AplaProject/go-apla/packages/script"

	"github.com/op/go-logging"
)

// Contract contains the information about the contract.
type Contract struct {
	Name          string
	Called        uint32
	FreeRequest   bool
	TxPrice       int64   // custom price for citizens
	TxGovAccount  int64   // state wallet
	EGSRate       float64 // money/EGS rate
	TableAccounts string
	StackCont     []string // Stack of called contracts
	Extend        *map[string]interface{}
	Block         *script.Block
}

const (
	// CallInit is a flag for calling init function of the contract
	CallInit = 0x01
	// CallCondition is a flag for calling condition function of the contract
	CallCondition = 0x02
	// CallAction is a flag for calling action function of the contract
	CallAction = 0x04
	// CallRollback is a flag for calling rollback function of the contract
	CallRollback = 0x08
)

var (
	smartVM   *script.VM
	smartVDE  map[int64]*script.VM
	smartTest = make(map[string]string)
	log       = logging.MustGetLogger("daemons")
)

func testValue(name string, v ...interface{}) {
	smartTest[name] = fmt.Sprint(v...)
}

// GetTestValue returns the test value of the specified key
func GetTestValue(name string) string {
	return smartTest[name]
}

func newVM() *script.VM {
	vm := script.NewVM()
	vm.Extern = true
	vm.Extend(&script.ExtendData{Objects: map[string]interface{}{
		"Println": fmt.Println,
		"Sprintf": fmt.Sprintf,
		"TxJson":  TxJSON,
		"Float":   Float,
		"Money":   script.ValueToDecimal,
		`Test`:    testValue,
	}, AutoPars: map[string]string{
		`*smart.Contract`: `contract`,
	}})
	return vm
}

func init() {
	smartVM = newVM()
	smartVDE = make(map[int64]*script.VM)
}

/*func pref2state(prefix string) (state uint32) {
	if prefix != `global` {
		if val, err := strconv.ParseUint(prefix, 10, 32); err == nil {
			state = uint32(val)
		}
	}
	return
}*/

func GetVM(vde bool, ecosystemID int64) *script.VM {
	if vde {
		if v, ok := smartVDE[ecosystemID]; ok {
			return v
		}
		return nil
	}
	return smartVM
}

func vmExternOff(vm *script.VM) {
	vm.FlushExtern()
}

func vmCompile(vm *script.VM, src string, owner *script.OwnerInfo) error {
	return vm.Compile([]rune(src), owner)
}

func vmCompileBlock(vm *script.VM, src string, owner *script.OwnerInfo) (*script.Block, error) {
	return vm.CompileBlock([]rune(src), owner)
}

func vmCompileEval(vm *script.VM, src string, prefix uint32) error {
	return vm.CompileEval(src, prefix)
}

func vmEvalIf(vm *script.VM, src string, state uint32, extend *map[string]interface{}) (bool, error) {
	return vm.EvalIf(src, state, extend)
}

func vmFlushBlock(vm *script.VM, root *script.Block) {
	vm.FlushBlock(root)
}

func vmExtend(vm *script.VM, ext *script.ExtendData) {
	vm.Extend(ext)
}

func vmRun(vm *script.VM, block *script.Block, params []interface{}, extend *map[string]interface{}) (ret []interface{}, err error) {
	var extcost int64
	cost := script.CostDefault
	if ecost, ok := (*extend)[`txcost`]; ok {
		cost = ecost.(int64)
	}
	rt := vm.RunInit(cost)
	ret, err = rt.Run(block, params, extend)
	if ecost, ok := (*extend)[`txcost`]; ok && cost > ecost.(int64) {
		extcost = cost - ecost.(int64)
	}
	(*extend)[`txcost`] = rt.Cost() - extcost
	return
}

func GetContractVM(vm *script.VM, name string, state uint32) *Contract {
	name = script.StateName(state, name)
	obj, ok := vm.Objects[name]
	if ok && obj.Type == script.ObjContract {
		return &Contract{Name: name, Block: obj.Value.(*script.Block)}
	}
	return nil
}

func vmGetUsedContracts(vm *script.VM, name string, state uint32, full bool) []string {
	contract := GetContractVM(vm, name, state)
	if contract == nil || contract.Block.Info.(*script.ContractInfo).Used == nil {
		return nil
	}
	ret := make([]string, 0)
	used := make(map[string]bool)
	for key := range contract.Block.Info.(*script.ContractInfo).Used {
		ret = append(ret, key)
		used[key] = true
		if full {
			sub := vmGetUsedContracts(vm, key, state, full)
			for _, item := range sub {
				if _, ok := used[item]; !ok {
					ret = append(ret, item)
					used[item] = true
				}
			}
		}
	}
	return ret
}

func vmGetContractByID(vm *script.VM, id int32) *Contract {
	idcont := id // - CNTOFF
	if len(vm.Children) <= int(idcont) || vm.Children[idcont].Type != script.ObjContract {
		return nil
	}
	return &Contract{Name: vm.Children[idcont].Info.(*script.ContractInfo).Name,
		Block: vm.Children[idcont]}
}

func vmExtendCost(vm *script.VM, ext func(string) int64) {
	vm.ExtCost = ext
}

func vmFuncCallsDB(vm *script.VM, funcCallsDB map[string]struct{}) {
	vm.FuncCallsDB = funcCallsDB
}

// ExternOff switches off the extern compiling mode in smartVM
func ExternOff() {
	vmExternOff(smartVM)
}

// Compile compiles contract source code in smartVM
func Compile(src string, owner *script.OwnerInfo) error {
	return vmCompile(smartVM, src, owner)
}

// CompileBlock calls CompileBlock for smartVM
func CompileBlock(src string, owner *script.OwnerInfo) (*script.Block, error) {
	return vmCompileBlock(smartVM, src, owner)
}

// CompileEval calls CompileEval for smartVM
func CompileEval(src string, prefix uint32) error {
	return vmCompileEval(smartVM, src, prefix)
}

// EvalIf calls EvalIf for smartVM
func EvalIf(src string, state uint32, extend *map[string]interface{}) (bool, error) {
	return vmEvalIf(smartVM, src, state, extend)
}

// FlushBlock calls FlushBlock for smartVM
func FlushBlock(root *script.Block) {
	vmFlushBlock(smartVM, root)
}

// ExtendCost sets the cost of calling extended obj in smartVM
func ExtendCost(ext func(string) int64) {
	vmExtendCost(smartVM, ext)
}

func FuncCallsDB(funcCallsDB map[string]struct{}) {
	vmFuncCallsDB(smartVM, funcCallsDB)
}

// Extend set extended variable and functions in smartVM
func Extend(ext *script.ExtendData) {
	vmExtend(smartVM, ext)
}

// Run executes Block in smartVM
func Run(block *script.Block, params []interface{}, extend *map[string]interface{}) (ret []interface{}, err error) {
	return vmRun(smartVM, block, params, extend)
}

// ActivateContract sets Active status of the contract in smartVM
func ActivateContract(tblid, state int64, active bool) {
	for i, item := range smartVM.Block.Children {
		if item != nil && item.Type == script.ObjContract {
			cinfo := item.Info.(*script.ContractInfo)
			if cinfo.Owner.TableID == tblid && cinfo.Owner.StateID == uint32(state) {
				smartVM.Children[i].Info.(*script.ContractInfo).Owner.Active = active
			}
		}
	}
}

// GetContract returns true if the contract exists in smartVM
func GetContract(name string, state uint32) *Contract {
	return GetContractVM(smartVM, name, state)
}

// GetUsedContracts returns the list of contracts which are called from the specified contract
func GetUsedContracts(name string, state uint32, full bool) []string {
	return vmGetUsedContracts(smartVM, name, state, full)
}

// GetContractByID returns true if the contract exists
func GetContractByID(id int32) *Contract {
	return vmGetContractByID(smartVM, id)
}

// GetFunc returns the block of the specified function in the contract
func (contract *Contract) GetFunc(name string) *script.Block {
	if block, ok := (*contract).Block.Objects[name]; ok && block.Type == script.ObjFunc {
		return block.Value.(*script.Block)
	}
	return nil
}

// TxJSON returns JSON data which has been generated from Tx data and extended variables
func TxJSON(contract *Contract) string {
	lines := make([]string, 0)
	for _, fitem := range *(*contract).Block.Info.(*script.ContractInfo).Tx {
		switch fitem.Type.String() {
		case `string`:
			lines = append(lines, fmt.Sprintf(`"%s": "%s"`, fitem.Name, (*(*contract).Extend)[fitem.Name]))
		case `int64`:
			lines = append(lines, fmt.Sprintf(`"%s": %d`, fitem.Name, (*(*contract).Extend)[fitem.Name]))
		case `[]uint8`:
			lines = append(lines, fmt.Sprintf(`"%s": "%s"`, fitem.Name,
				hex.EncodeToString((*(*contract).Extend)[fitem.Name].([]byte))))
		}
	}
	return `{` + strings.Join(lines, ",\r\n") + `}`
}

// Float converts int64, string to float64
func Float(v interface{}) (ret float64) {
	switch value := v.(type) {
	case int64:
		ret = float64(value)
	case string:
		if val, err := strconv.ParseFloat(value, 64); err == nil {
			ret = val
		}
	}
	return
}

func ContractsList(value string) []string {
	list := make([]string, 0)
	re := regexp.MustCompile(`contract[\s]*([\d\w_]+)[\s]*{`)
	for _, item := range re.FindAllStringSubmatch(value, -1) {
		if len(item) > 1 {
			list = append(list, item[1])
		}
	}
	return list
}

// LoadContracts reads and compiles contracts from smart_contracts tables
func LoadContracts(transaction *model.DbTransaction) (err error) {
	var states []map[string]string
	var prefix []string
	prefix = []string{`system`}
	states, err = model.GetAll(`select id from system_states order by id`, -1)
	if err != nil {
		return err
	}
	for _, istate := range states {
		prefix = append(prefix, istate[`id`])
	}
	for _, ipref := range prefix {
		if err = LoadContract(transaction, ipref); err != nil {
			break
		}
	}
	ExternOff()
	return
}

// LoadContract reads and compiles contract of new state
func LoadContract(transaction *model.DbTransaction, prefix string) (err error) {
	var contracts []map[string]string
	contracts, err = model.GetAllTransaction(transaction, `select * from "`+prefix+`_contracts" order by id`, -1)
	if err != nil {
		return err
	}
	state := uint32(converter.StrToInt64(prefix))
	for _, item := range contracts {
		names := strings.Join(ContractsList(item[`value`]), `,`)
		owner := script.OwnerInfo{
			StateID:  state,
			Active:   item[`active`] == `1`,
			TableID:  converter.StrToInt64(item[`id`]),
			WalletID: converter.StrToInt64(item[`wallet_id`]),
			TokenID:  converter.StrToInt64(item[`token_id`]),
		}
		if err = Compile(item[`value`], &owner); err != nil {
			log.Error("Load Contract", names, err)
			fmt.Println("Error Load Contract", names, err)
			//return
		} else {
			fmt.Println("OK Load Contract", names, item[`id`], item[`active`] == `1`)
		}
	}
	LoadVDEContracts(transaction, prefix)
	return
}

func LoadVDEContracts(transaction *model.DbTransaction, prefix string) (err error) {
	var contracts []map[string]string

	if !model.IsTable(prefix + `_vde_contracts`) {
		return
	}
	contracts, err = model.GetAllTransaction(transaction, `select * from "`+prefix+`_vde_contracts" order by id`, -1)
	if err != nil {
		return err
	}
	state := converter.StrToInt64(prefix)
	vm := newVM()
	EmbedFuncs(vm)
	smartVDE[state] = vm
	for _, item := range contracts {
		names := strings.Join(ContractsList(item[`value`]), `,`)
		owner := script.OwnerInfo{
			StateID:  uint32(state),
			Active:   false,
			TableID:  converter.StrToInt64(item[`id`]),
			WalletID: 0,
			TokenID:  0,
		}
		if err = vmCompile(vm, item[`value`], &owner); err != nil {
			log.Error("Load VDE Contract", names, err)
			fmt.Println("Error Load VDE Contract", names, err)
		} else {
			fmt.Println("OK Load VDE Contract", names, item[`id`])
		}
	}

	return
}

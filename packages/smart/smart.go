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
	"bytes"
	"encoding/hex"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/AplaProject/go-apla/packages/config/syspar"
	"github.com/AplaProject/go-apla/packages/converter"
	"github.com/AplaProject/go-apla/packages/model"
	"github.com/AplaProject/go-apla/packages/script"
	"github.com/AplaProject/go-apla/packages/templatev2"
	"github.com/AplaProject/go-apla/packages/utils"

	"github.com/jinzhu/gorm"
	"github.com/op/go-logging"
	"github.com/shopspring/decimal"
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
		`*smart.Contract`: `sc`,
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

func VMCompileEval(vm *script.VM, src string, prefix uint32) error {
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

func VMRun(vm *script.VM, block *script.Block, params []interface{}, extend *map[string]interface{}) (ret []interface{}, err error) {
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

func VMGetContractByID(vm *script.VM, id int32) *Contract {
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
	return VMCompileEval(smartVM, src, prefix)
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
	return VMRun(smartVM, block, params, extend)
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
	return VMGetContractByID(smartVM, id)
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

// GetContractLimit returns the default maximal cost of contract
func (sc *SmartContract) GetContractLimit() (ret int64) {
	// default maximum cost of F
	if len(sc.TxSmart.MaxSum) > 0 {
		sc.TxCost = converter.StrToInt64(sc.TxSmart.MaxSum)
	} else {
		cost, _ := templatev2.StateParam(sc.TxSmart.StateID, `max_sum`)
		if len(cost) > 0 {
			sc.TxCost = converter.StrToInt64(cost)
		}
	}
	if sc.TxCost == 0 {
		sc.TxCost = script.CostDefault // fuel
	}
	return sc.TxCost
}

func (sc *SmartContract) getExtend() *map[string]interface{} {
	head := sc.TxSmart
	var citizenID, walletID int64
	citizenID = int64(head.UserID)
	walletID = int64(head.UserID)
	block := int64(0)
	blockTime := int64(0)
	walletBlock := int64(0)
	if sc.BlockData != nil {
		block = sc.BlockData.BlockID
		walletBlock = sc.BlockData.WalletID
		blockTime = sc.BlockData.Time
	}
	extend := map[string]interface{}{`type`: head.Type, `time`: head.Time, `state`: head.StateID,
		`block`: block, `citizen`: citizenID, `wallet`: walletID, `wallet_block`: walletBlock,
		`parent`: ``, `txcost`: sc.GetContractLimit(), `txhash`: sc.TxHash, `result`: ``,
		`parser`: sc, `sc`: sc, `contract`: sc.TxContract, `block_time`: blockTime}
	for key, val := range sc.TxData {
		extend[key] = val
	}

	return &extend
}

// StackCont adds an element to the stack of contract call or removes the top element when name is empty
func StackCont(sc interface{}, name string) {
	cont := sc.(*SmartContract).TxContract
	if len(name) > 0 {
		cont.StackCont = append(cont.StackCont, name)
	} else {
		cont.StackCont = cont.StackCont[:len(cont.StackCont)-1]
	}
	return
}

// CallContract calls the contract functions according to the specified flags
func (sc *SmartContract) CallContract(flags int) (err error) {
	var (
		public                     []byte
		sizeFuel /*toID,*/, fromID int64
		fuelRate                   decimal.Decimal
	)
	payWallet := &model.Key{}
	sc.TxContract.Extend = sc.getExtend()
	var price int64

	methods := []string{`init`, `conditions`, `action`, `rollback`}
	sc.TxContract.StackCont = []string{sc.TxContract.Name}
	(*sc.TxContract.Extend)[`stack_cont`] = StackCont
	sc.VM = GetVM(sc.VDE, sc.TxSmart.StateID)
	if (flags&CallRollback) == 0 && (flags&CallAction) != 0 {
		/*if sc.BlockData != nil {
			toID = sc.BlockData.WalletID
		}*/
		fromID = sc.TxSmart.UserID
		if len(sc.TxSmart.PublicKey) > 0 && string(sc.TxSmart.PublicKey) != `null` {
			public = sc.TxSmart.PublicKey
		}
		wallet := &model.Key{}
		wallet.SetTablePrefix(sc.TxSmart.StateID)
		err := wallet.Get(sc.TxSmart.UserID)
		if err != nil && err != gorm.ErrRecordNotFound {
			return err
		}
		if len(wallet.PublicKey) > 0 {
			public = wallet.PublicKey
		}
		if sc.TxSmart.Type == 258 { // UpdFullNodes
			node := syspar.GetNode(sc.TxSmart.UserID)
			if node == nil {
				return fmt.Errorf("unknown node id")
			}
			public = node.Public
		}
		if len(public) == 0 {
			return fmt.Errorf("empty public key")
		}
		sc.PublicKeys = append(sc.PublicKeys, public)
		//		fmt.Println(`CALL CONTRACT`, sc.TxData[`forsign`].(string))
		CheckSignResult, err := utils.CheckSign(sc.PublicKeys, sc.TxData[`forsign`].(string), sc.TxSmart.BinSignatures, false)
		if err != nil {
			fmt.Println(`ForSign`, sc.TxData[`forsign`].(string))
			return err
		}
		if !CheckSignResult {
			return fmt.Errorf("incorrect sign")
		}
		if sc.TxSmart.StateID > 0 {
			if sc.TxSmart.TokenEcosystem == 0 {
				sc.TxSmart.TokenEcosystem = 1
			}
			fuelRate, err = decimal.NewFromString(syspar.GetFuelRate(sc.TxSmart.TokenEcosystem))
			if err != nil {
				return err
			}
			if fuelRate.Cmp(decimal.New(0, 0)) <= 0 {
				return fmt.Errorf(`Fuel rate must be greater than 0`)
			}
			if len(sc.TxSmart.PayOver) > 0 {
				payOver, err := decimal.NewFromString(sc.TxSmart.PayOver)
				if err != nil {
					return err
				}
				fuelRate = fuelRate.Add(payOver)
			}
			if sc.TxContract.Block.Info.(*script.ContractInfo).Owner.Active {
				fromID = sc.TxContract.Block.Info.(*script.ContractInfo).Owner.WalletID
				sc.TxSmart.TokenEcosystem = sc.TxContract.Block.Info.(*script.ContractInfo).Owner.TokenID
			} else if len(sc.TxSmart.PayOver) > 0 {
				payOver, err := decimal.NewFromString(sc.TxSmart.PayOver)
				if err != nil {
					return err
				}
				fuelRate = fuelRate.Add(payOver)
			}
			payWallet.SetTablePrefix(sc.TxSmart.TokenEcosystem)
			if err = payWallet.Get(fromID); err != nil {
				if err == gorm.ErrRecordNotFound {
					return fmt.Errorf(`current balance is not enough`)
				}
				return err
			}
			if !bytes.Equal(wallet.PublicKey, payWallet.PublicKey) && !bytes.Equal(sc.TxSmart.PublicKey, payWallet.PublicKey) {
				return fmt.Errorf(`Token and user public keys are different`)
			}
			amount, err := decimal.NewFromString(payWallet.Amount)
			if err != nil {
				return err
			}
			if cprice := sc.TxContract.GetFunc(`price`); cprice != nil {
				var ret []interface{}
				if ret, err = VMRun(sc.VM, cprice, nil, sc.TxContract.Extend); err != nil {
					return err
				} else if len(ret) == 1 {
					if _, ok := ret[0].(int64); !ok {
						return fmt.Errorf(`Wrong result type of price function`)
					}
					price = ret[0].(int64)
				} else {
					return fmt.Errorf(`Wrong type of price function`)
				}
			}
			sizeFuel = syspar.GetSizeFuel() * int64(len(sc.TxSmart.Data)) / 1024
			if amount.Cmp(decimal.New(sizeFuel+price, 0).Mul(fuelRate)) <= 0 {
				return fmt.Errorf(`current balance is not enough`)
			}
		}
	}
	before := (*sc.TxContract.Extend)[`txcost`].(int64) + price

	// Payment for the size
	(*sc.TxContract.Extend)[`txcost`] = (*sc.TxContract.Extend)[`txcost`].(int64) - sizeFuel

	sc.TxContract.FreeRequest = false
	for i := uint32(0); i < 4; i++ {
		if (flags & (1 << i)) > 0 {
			cfunc := sc.TxContract.GetFunc(methods[i])
			if cfunc == nil {
				continue
			}
			sc.TxContract.Called = 1 << i
			_, err = VMRun(sc.VM, cfunc, nil, sc.TxContract.Extend)
			if err != nil {
				before -= price
				break
			}
		}
	}
	sc.TxUsedCost = decimal.New(before-(*sc.TxContract.Extend)[`txcost`].(int64), 0)
	sc.TxContract.TxPrice = price
	/*	if (flags&CallAction) != 0 && sc.TxSmart.StateID > 0 && !sc.VDE {
		apl := sc.TxUsedCost.Mul(fuelRate)
		fmt.Printf("Pay fuel=%v fromID=%d toID=%d cost=%v apl=%v", fuelRate, fromID, toID, sc.TxUsedCost, apl)
		wltAmount, err := decimal.NewFromString(payWallet.Amount)
		if err != nil {
			return err
		}
		if wltAmount.Cmp(apl) < 0 {
			apl = wltAmount
		}
		commission := apl.Mul(decimal.New(syspar.SysInt64(`commission_size`), 0)).Div(decimal.New(100, 0)).Floor()
		walletTable := fmt.Sprintf(`%d_keys`, sc.TxSmart.TokenEcosystem)
				if _, _, err := sc.selectiveLoggingAndUpd([]string{`-amount`}, []interface{}{apl}, walletTable, []string{`id`},
					[]string{converter.Int64ToStr(fromID)}, true); err != nil {
					return err
				}
				if _, _, err := sc.selectiveLoggingAndUpd([]string{`+amount`}, []interface{}{apl.Sub(commission)}, walletTable, []string{`id`},
					[]string{converter.Int64ToStr(toID)}, true); err != nil {
					return err
				}
				if _, _, err := sc.selectiveLoggingAndUpd([]string{`+amount`}, []interface{}{commission}, walletTable, []string{`id`},
					[]string{syspar.GetCommissionWallet(sc.TxSmart.TokenEcosystem)}, true); err != nil {
					return err
				}
		fmt.Printf(" Paid commission %v\r\n", commission)
	} */
	return
}

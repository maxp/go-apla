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

package utils

import (
	"archive/zip"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/AplaProject/go-apla/packages/consts"
	"github.com/AplaProject/go-apla/packages/converter"
	"github.com/AplaProject/go-apla/packages/crypto"
	"github.com/EGaaS/go-egaas-mvp/packages/model"
	"github.com/EGaaS/go-egaas-mvp/packages/utils"
	"github.com/kardianos/osext"
	"github.com/mcuadros/go-version"
	"github.com/op/go-logging"
)

var log = logging.MustGetLogger("daemons")

// BlockData is a structure of the block's header
type BlockData struct {
	BlockID  int64
	Time     int64
	WalletID int64
	StateID  int64
	Sign     []byte
	Hash     []byte
	Version  int
}

type Update struct {
	Version string
	Hash    string
	Sign    string
	URL     string
}

var (
	// FirstBlockDir is a folder where 1block file will be stored
	FirstBlockDir = flag.String("firstBlockDir", "", "FirstBlockDir")
	// FirstBlockPublicKey is the private key
	FirstBlockPublicKey = flag.String("firstBlockPublicKey", "", "FirstBlockPublicKey")
	// FirstBlockNodePublicKey is the node private key
	FirstBlockNodePublicKey = flag.String("firstBlockNodePublicKey", "", "FirstBlockNodePublicKey")
	// FirstBlockHost is the host of the first block
	FirstBlockHost = flag.String("firstBlockHost", "", "FirstBlockHost")
	// WalletAddress is a wallet address for forging
	WalletAddress = flag.String("walletAddress", "", "walletAddress for forging ")
	// TCPHost is the tcp host
	TCPHost = flag.String("tcpHost", "", "tcpHost (e.g. 127.0.0.1)")
	// ListenHTTPPort is HTTP port
	ListenHTTPPort = flag.String("listenHttpPort", "7079", "ListenHTTPPort")
	// GenerateFirstBlock show if the first block must be generated
	GenerateFirstBlock = flag.Int64("generateFirstBlock", 0, "generateFirstBlock")
	// OldVersion is the number of the old version
	OldVersion = flag.String("oldVersion", "", "")
	// TestRollBack equals 1 for testing rollback
	TestRollBack = flag.Int64("testRollBack", 0, "testRollBack")
	// Dir is apla folder
	Dir = flag.String("dir", GetCurrentDir(), "DayLight directory")
	// OldFileName is the old file name
	OldFileName = flag.String("oldFileName", "", "")
	// LogLevel is the log level
	LogLevel = flag.String("logLevel", "", "DayLight LogLevel")
	// Console equals 1 for starting in console
	Console = flag.Int64("console", 0, "Start from console")
	// StartBlockID is the start block
	StartBlockID = flag.Int64("startBlockId", 0, "Start block for blockCollection daemon")
	// EndBlockID is the end block
	EndBlockID = flag.Int64("endBlockId", 0, "End block for blockCollection daemon")
	// RollbackToBlockID is the target block for rollback
	RollbackToBlockID = flag.Int64("rollbackToBlockId", 0, "Rollback to block_id")
	// TLS is a directory for .well-known and keys. It is required for https
	TLS = flag.String("tls", "", "Support https. Specify directory for .well-known")
	// DevTools switches on dev tools in thrust shell
	DevTools = flag.Int64("devtools", 0, "Devtools in thrust-shell")
	// BoltDir is the edir for BoltDb folder
	BoltDir = flag.String("boltDir", GetCurrentDir(), "Bolt directory")
	// BoltPsw is the password for BoltDB
	BoltPsw = flag.String("boltPsw", "", "Bolt password")
	// APIToken is an api token for exchange api
	APIToken = flag.String("apiToken", "", "API Token")
	// OneCountry is the country which is supported
	OneCountry int64
	// PrivCountry is protect system from registering
	PrivCountry bool
	//	OutFile            *os.File

	// LogoExt is the extension of the logotype
	LogoExt = `png`
	// DltWalletID is the wallet identifier
	DltWalletID = flag.Int64("dltWalletId", 0, "DltWalletID")

	ReturnCh     chan string
	CancelFunc   context.CancelFunc
	DaemonsCount int
	// Thrust is true for thrust shell
	Thrust bool
)

func init() {
	flag.Parse()
}

// IOS checks if the app runs on iOS
func IOS() bool {
	if (runtime.GOARCH == "arm" || runtime.GOARCH == "arm64") && runtime.GOOS == "darwin" {
		return true
	}
	return false
}

// Desktop checks if the app runs on the desktop with thrust_shell
func Desktop() bool {
	thrustShell := "thrust_shell"
	if runtime.GOOS == "windows" {
		thrustShell = "thrust_shell.exe"
	} else if runtime.GOOS == "darwin" {
		thrustShell = "ThrustShell"
	}
	if _, err := os.Stat(*Dir + "/" + thrustShell); err == nil {
		return true
	}
	return false
}

// Mobile checks if the app runs on Android or iOS
func Mobile() bool {
	if IOS() || runtime.GOOS == "android" {
		return true
	}
	return false
}

// Android checks if the app runs on Android
func Android() bool {
	if runtime.GOOS == "android" {
		return true
	}
	return false
}

// CheckInputData checks the input data
func CheckInputData(idata interface{}, dataType string) bool {
	var data string
	switch idata.(type) {
	case int:
		data = converter.IntToStr(idata.(int))
	case int64:
		data = converter.Int64ToStr(idata.(int64))
	case float64:
		data = converter.Float64ToStr(idata.(float64))
	case string:
		data = idata.(string)
	case []byte:
		data = string(idata.([]byte))
	}
	log.Debug("CheckInputData:" + data)
	log.Debug("dataType:" + dataType)
	switch dataType {
	case "arbitration_trust_list":
		if ok, _ := regexp.MatchString(`^\[[0-9]{1,10}(,[0-9]{1,10}){0,100}\]$`, data); ok {
			return true
		}
	case "abuse_comment", "vote_comment":
		if ok, _ := regexp.MatchString(`^(?i)[0-9a-z\,\s\.\-]{1,255}$`, data); ok {
			return true
		}
	case "private_key":
		if ok, _ := regexp.MatchString(`^[0-9a-fA-F]+$`, data); ok {
			if len(data) == 64 {
				return true
			}
		}
	case "votes_comment", "cf_comment":
		if ok, _ := regexp.MatchString(`^[\pL0-9\,\s\.\-\:\=\;\?\!\%\)\(\@\/\n\r]{1,140}$`, data); ok {
			return true
		}
	case "type":
		if ok, _ := regexp.MatchString(`^[\w]+$`, data); ok {
			if converter.StrToInt(data) <= 30 {
				return true
			}
		}
	case "word":
		if ok, _ := regexp.MatchString(`^(?i)[a-z]+$`, data); ok {
			if converter.StrToInt(data) <= 1024 {
				return true
			}
		}
	case "currency_name", "state_name":
		if ok, _ := regexp.MatchString(`^[\pL0-9\,\s\.\-\:\=\;\?\!\%\)\(\@\/\n\r]{1,20}$`, data); ok {
			if converter.StrToInt(data) <= 1024 {
				return true
			}
		}
	case "string":
		if ok, _ := regexp.MatchString(`^[\w]+$`, data); ok {
			if converter.StrToInt(data) <= 1024 {
				return true
			}
		}
	case "referral":
		if ok, _ := regexp.MatchString(`^[0-9]{1,2}$`, data); ok {
			if converter.StrToInt(data) <= 30 {
				return true
			}
		}
	case "currency_id":
		if ok, _ := regexp.MatchString(`^[0-9]{1,3}$`, data); ok {
			if converter.StrToInt(data) <= 255 {
				return true
			}
		}
	case "system_commission":
		if ok, _ := regexp.MatchString(`^[0-9]{1,3}$`, data); ok {
			if converter.StrToInt(data) <= 15 && converter.StrToInt(data) >= 5 {
				return true
			}
		}
	case "tinyint":
		if ok, _ := regexp.MatchString(`^[0-9]{1,3}$`, data); ok {
			if converter.StrToInt(data) <= 127 {
				return true
			}
		}
	case "smallint":
		if ok, _ := regexp.MatchString(`^[0-9]{1,5}$`, data); ok {
			if converter.StrToInt(data) <= 65535 {
				return true
			}
		}
	case "column_type":
		if ok, _ := regexp.MatchString(`^(text|int64|time|hash|money|double)$`, data); ok {
			return true
		}
	case "avatar":
		regex := `https?\:\/\/`        // SCHEME
		regex += `[\w-.]*\.[a-z]{2,4}` // Host or IP
		regex += `(\:[0-9]{2,5})?`     // Port
		regex += `(\/[\w_-]+)*\/?`     // Path
		regex += `\.(png|jpg)`         // Img
		if ok, _ := regexp.MatchString(`^`+regex+`$`, data); ok {
			if len(data) < 100 {
				return true
			}
		}
	case "img_url":
		regex := `https?\:\/\/`        // SCHEME
		regex += `[\w-.]*\.[a-z]{2,4}` // Host or IP
		regex += `(\:[0-9]{2,5})?`     // Port
		regex += `(\/[\w_-]+)*\/?`     // Path
		regex += `\.(png|jpg)`         // Img
		if ok, _ := regexp.MatchString(`^`+regex+`$`, data); ok {
			if len(data) < 50 {
				return true
			}
		}
	case "ca_url", "arbitrator_url":
		regex := `https?\:\/\/`        // SCHEME
		regex += `[\w-.]*\.[a-z]{2,4}` // Host or IP
		regex += `(\:[0-9]{2,5})?`     // Port
		regex += `(\/[\w_-]+)*\/?`     // Path
		if ok, _ := regexp.MatchString(`^`+regex+`$`, data); ok {
			if len(data) <= 30 {
				return true
			}
		}
	case "credit_pct", "pct":
		if ok, _ := regexp.MatchString(`^[0-9]{1,3}(\.[0-9]{2})?$`, data); ok {
			return true
		}
	case "user_name":
		if ok, _ := regexp.MatchString(`^[\w\s]{1,30}$`, data); ok {
			return true
		}
	case "admin_currency_list":
		if ok, _ := regexp.MatchString(`^((\d{1,3}\,){0,9}\d{1,3}|ALL)$`, data); ok {
			return true
		}
	case "users_ids":
		if ok, _ := regexp.MatchString(`^([0-9]{1,12},){0,1000}[0-9]{1,12}$`, data); ok {
			return true
		}
	case "version":
		if ok, _ := regexp.MatchString(`^[0-9]{1,2}\.[0-9]{1,2}\.[0-9]{1,2}([a-z]{1,2}[0-9]{1,2})?$`, data); ok {
			return true
		}
	case "soft_type":
		if ok, _ := regexp.MatchString(`^[a-z]{3,10}$`, data); ok {
			return true
		}
	case "currency_full_name":
		if ok, _ := regexp.MatchString(`^[a-zA-Z\s]{3,50}$`, data); ok {
			return true
		}
	case "currency_commission":
		if ok, _ := regexp.MatchString(`^[0-9]{1,7}(\.[0-9]{1,2})?$`, data); ok {
			return true
		}
	case "sell_rate":
		if ok, _ := regexp.MatchString(`^[0-9]{0,10}(\.[0-9]{0,10})?$`, data); ok {
			return true
		}
	case "amount":
		if ok, _ := regexp.MatchString(`^[0-9]{0,10}(\.[0-9]{0,2})?$`, data); ok {
			return true
		}
	case "amount_btc":
		if ok, _ := regexp.MatchString(`^[0-9]{0,10}(\.[0-9]{0,5})?$`, data); ok {
			return true
		}
	case "tpl_name":
		if ok, _ := regexp.MatchString("^[\\w]{1,30}$", data); ok {
			return true
		}
	case "example_spots":
		r1 := `"\d{1,2}":\["\d{1,3}","\d{1,3}",(\[("[a-z_]{1,30}",?){0,20}\]|""),"\d{1,2}","\d{1,2}"\]`
		reg := `^\{(\"(face|profile)\":\{(` + r1 + `,?){1,20}\},?){2}}$`
		if ok, _ := regexp.MatchString(reg, data); ok {
			return true
		}
	case "segments":
		r1 := `"\d{1,2}":\["\d{1,2}","\d{1,2}"\]`
		face := `"face":\{(` + r1 + `\,){1,20}` + r1 + `\}`
		profile := `"profile":\{(` + r1 + `\,){1,20}` + r1 + `\}`
		reg := `^\{` + face + `,` + profile + `\}$`
		if ok, _ := regexp.MatchString(reg, data); ok {
			return true
		}
	case "tolerances":
		r1 := `"\d{1,2}":"0\.\d{1,2}"`
		face := `"face":\{(` + r1 + `\,){1,50}` + r1 + `\}`
		profile := `"profile":\{(` + r1 + `\,){1,50}` + r1 + `\}`
		reg := `^\{` + face + `,` + profile + `\}$`
		if ok, _ := regexp.MatchString(reg, data); ok {
			return true
		}
	case "compatibility":
		if ok, _ := regexp.MatchString(`^\[(\d{1,5},)*\d{1,5}\]$`, data); ok {
			return true
		}
	case "race":
		if ok, _ := regexp.MatchString("^[1-3]$", data); ok {
			return true
		}
	case "country":
		if ok, _ := regexp.MatchString("^[0-9]{1,3}$", data); ok {
			return true
		}
	case "vote", "boolean":
		if ok, _ := regexp.MatchString(`^0|1$`, data); ok {
			return true
		}
	case "coordinate":
		if ok, _ := regexp.MatchString(`^\-?[0-9]{1,3}(\.[0-9]{1,5})?$`, data); ok {
			return true
		}
	case "cf_links":
		regex := `\["https?\:\/\/(goo\.gl|bit\.ly|t\.co)\/[\w-]+",[0-9]+,[0-9]+,[0-9]+,[0-9]+\]`
		if ok, _ := regexp.MatchString(`^\[`+regex+`(\,`+regex+`)*\]$`, data); ok {
			if len(data) < 512 {
				return true
			}
		}
	case "http_host":
		if ok, _ := regexp.MatchString(`^https?:\/\/[0-9a-z\_\.\-\/:]{1,100}[\/]$`, data); ok {
			return true
		}
	case "e_host":
		if ok, _ := regexp.MatchString(`^https?:\/\/[0-9a-z\_\.\-\/:]{1,100}[\/]$`, data); ok || data == "0" {
			return true
		}
	case "host":
		if ok, _ := regexp.MatchString(`^(?i)[0-9a-z\_\.\-]{1,100}$`, data); ok {
			return true
		}
	case "tcp_host":
		if ok, _ := regexp.MatchString(`^(?i)[0-9a-z\_\.\-]{1,100}:[0-9]+$`, data); ok {
			return true
		}
	case "coords":
		xy := `\[\d{1,3}\,\d{1,3}\]`
		r := `^\[(` + xy + `\,){}` + xy + `\]$`
		if ok, _ := regexp.MatchString(r, data); ok {
			return true
		}
	case "lang":
		if ok, _ := regexp.MatchString("^(en|ru)$", data); ok {
			return true
		}
	case "payment_systems_ids":
		if ok, _ := regexp.MatchString("^([0-9]{1,4},){0,4}[0-9]{1,4}$", data); ok {
			return true
		}
	case "video_type":
		if ok, _ := regexp.MatchString("^(youtube|vimeo|youku|null)$", data); ok {
			return true
		}
	case "video_url_id", "sn_url_id":
		if ok, _ := regexp.MatchString("^(?i)(null|[0-9a-z_\\-\\.]{2,32})$", data); ok {
			return true
		}
	case "sn_type":
		if ok, _ := regexp.MatchString("^(vk|fb|qq)$", data); ok {
			return true
		}
	case "sha1":
		if ok, _ := regexp.MatchString("^[0-9a-z]{40}$", data); ok {
			return true
		}

	case "walletAddress":
		if ok, _ := regexp.MatchString("^(?i)[0-9]{20}$", strings.Replace(data, `-`, ``, -1)); ok {
			return true
		}
	case "photo_hash", "sha256":
		if ok, _ := regexp.MatchString("^[0-9a-z]{64}$", data); ok {
			return true
		}
	case "cash_code":
		if ok, _ := regexp.MatchString("^[0-9a-z]{32}$", data); ok {
			return true
		}
	case "alert":
		if ok, _ := regexp.MatchString("^[\\pL0-9\\,\\s\\.\\-\\:\\=\\;\\?\\!\\%\\)\\(\\@\\/]{1,512}$", data); ok {
			return true
		}
	case "int":
		if ok, _ := regexp.MatchString("^[0-9]{1,10}$", data); ok {
			return true
		}
	case "float":
		if ok, _ := regexp.MatchString(`^[0-9]{1,5}(\.[0-9]{1,5})?$`, data); ok {
			return true
		}
	case "sleep_var":
		if ok, _ := regexp.MatchString(`^\{\"is_ready\"\:\[([0-9]{1,5},){1,100}[0-9]{1,5}\],\"generator\"\:\[([0-9]{1,5},){1,100}[0-9]{1,5}\]\}$`, data); ok {
			return true
		}
	case "int64", "bigint", "user_id":
		if ok, _ := regexp.MatchString("^-?[0-9]{1,20}$", data); ok {
			return true
		}
	case "decimal": // 1.2345678e+25
		if ok, _ := regexp.MatchString(`^([0-9]{1,30})|([0-9]+\.[0-9]+[e]\+\[0-9]+)$`, data); ok {
			return true
		}
	case "level":
		if converter.StrToInt(data) >= 0 && converter.StrToInt(data) <= 34 {
			return true
		}
	case "comment":
		if len(data) >= 1 && len(data) <= 512 {
			return true
		}
	case "conditions":
		if len(data) <= 1024 {
			return true
		}
	case "hex_sign", "hex", "public_key":
		if ok, _ := regexp.MatchString("^[0-9a-z]+$", data); ok {
			if len(data) < 2048 {
				return true
			}
		}
	case "account":
		if ok, _ := regexp.MatchString(`^[0-9a-zA-Z\-\s_\+\#\:]{1,50}$`, data); ok {
			return true
		}
	case "method":
		if ok, _ := regexp.MatchString(`^[0-9a-zA-Z\-\_]{1,30}$`, data); ok {
			return true
		}
	}

	return false
}

// GetHTTPTextAnswer returns HTTP answer as a string
func GetHTTPTextAnswer(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	htmlData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode == 404 {
		err = fmt.Errorf(`404`)
	}
	return string(htmlData), err
}

// ErrInfoFmt fomats the error message
func ErrInfoFmt(err string, a ...interface{}) error {
	return fmt.Errorf("%s (%s)", fmt.Sprintf(err, a...), Caller(1))
}

// ErrInfo formats the error message
func ErrInfo(verr interface{}, additionally ...string) error {
	var err error
	switch verr.(type) {
	case error:
		err = verr.(error)
	case string:
		err = errors.New(verr.(string))
	}
	if err != nil {
		if len(additionally) > 0 {
			return fmt.Errorf("%s # %s (%s)", err, additionally, Caller(1))
		}
		return fmt.Errorf("%s (%s)", err, Caller(1))
	}
	return err
}

// CallMethod calls the function by its name
func CallMethod(i interface{}, methodName string) interface{} {
	var ptr reflect.Value
	var value reflect.Value
	var finalMethod reflect.Value

	value = reflect.ValueOf(i)

	// if we start with a pointer, we need to get value pointed to
	// if we start with a value, we need to get a pointer to that value
	if value.Type().Kind() == reflect.Ptr {
		ptr = value
		value = ptr.Elem()
	} else {
		ptr = reflect.New(reflect.TypeOf(i))
		temp := ptr.Elem()
		temp.Set(value)
	}

	// check for method on value
	method := value.MethodByName(methodName)
	if method.IsValid() {
		finalMethod = method
	}
	// check for method on pointer
	method = ptr.MethodByName(methodName)
	if method.IsValid() {
		finalMethod = method
	}

	if finalMethod.IsValid() {
		return finalMethod.Call([]reflect.Value{})[0].Interface()
	}

	// return or panic, method not found of either type
	return fmt.Errorf("method %s not found", methodName)
}

// Caller returns the name of the latest function
func Caller(steps int) string {
	name := "?"
	if pc, _, num, ok := runtime.Caller(steps + 1); ok {
		name = fmt.Sprintf("%s :  %d", filepath.Base(runtime.FuncForPC(pc).Name()), num)
	}
	return name
}

// CopyFileContents copy files
func CopyFileContents(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return ErrInfo(err)
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return ErrInfo(err)
	}
	defer func() {
		cerr := out.Close()
		if err == nil {
			err = cerr
		}
	}()
	if _, err = io.Copy(out, in); err != nil {
		return ErrInfo(err)
	}
	err = out.Sync()
	return ErrInfo(err)
}

// CheckSign checks the signature
func CheckSign(publicKeys [][]byte, forSign string, signs []byte, nodeKeyOrLogin bool) (bool, error) {
	defer func() {
		if r := recover(); r != nil {
			log.Error("Panic CheckECDSA %v", r)
		}
	}()

	var signsSlice [][]byte
	if len(forSign) == 0 {
		return false, ErrInfoFmt("len(forSign) == 0")
	}
	if len(publicKeys) == 0 {
		return false, ErrInfoFmt("len(publicKeys) == 0")
	}
	if len(signs) == 0 {
		return false, ErrInfoFmt("len(signs) == 0")
	}

	// node always has olny one signature
	if nodeKeyOrLogin {
		signsSlice = append(signsSlice, signs)
	} else {
		length, err := converter.DecodeLength(&signs)
		if err != nil {
			return false, err
		}
		if length > 0 {
			signsSlice = append(signsSlice, converter.BytesShift(&signs, length))
		}
		if len(publicKeys) != len(signsSlice) {
			return false, fmt.Errorf("sign error %d!=%d", len(publicKeys), len(signsSlice))
		}
	}
	return crypto.CheckSign(publicKeys[0], forSign, signsSlice[0])
}

// MerkleTreeRoot rertun Merkle value
func MerkleTreeRoot(dataArray [][]byte) []byte {
	log.Debug("dataArray: %s", dataArray)
	result := make(map[int32][][]byte)
	for _, v := range dataArray {
		hash, err := crypto.DoubleHash(v)
		if err != nil {
			log.Fatal(err)
		}
		hash = converter.BinToHex(hash)
		result[0] = append(result[0], hash)
	}
	var j int32
	for len(result[j]) > 1 {
		for i := 0; i < len(result[j]); i = i + 2 {
			if len(result[j]) <= (i + 1) {
				if _, ok := result[j+1]; !ok {
					result[j+1] = [][]byte{result[j][i]}
				} else {
					result[j+1] = append(result[j+1], result[j][i])
				}
			} else {
				if _, ok := result[j+1]; !ok {
					hash, err := crypto.DoubleHash(append(result[j][i], result[j][i+1]...))
					if err != nil {
						log.Fatal(err)
					}
					hash = converter.BinToHex(hash)
					result[j+1] = [][]byte{hash}
				} else {
					hash, err := crypto.DoubleHash([]byte(append(result[j][i], result[j][i+1]...)))
					if err != nil {
						log.Fatal(err)
					}
					hash = converter.BinToHex(hash)
					result[j+1] = append(result[j+1], hash)
				}
			}
		}
		j++
	}

	log.Debug("result: %s", result)
	ret := result[int32(len(result)-1)]
	log.Debug("result_: %s", ret)
	return []byte(ret[0])
}

// GetMrklroot returns MerkleTreeRoot
func GetMrklroot(binaryData []byte, first bool, maxTxSize int64, maxTxCount int) ([]byte, error) {
	var mrklSlice [][]byte
	var txSize int64
	// parse [error] after the calling of a function
	if len(binaryData) > 0 {
		for {

			// to exclude an attack on memory overflow
			if !first {
				if txSize > maxTxSize {
					return nil, ErrInfoFmt("[error] MAX_TX_SIZE")
				}
			}
			txSize, err := converter.DecodeLength(&binaryData)
			if err != nil {
				panic(err)
			}

			// separate one transaction from the list of transactions
			if txSize > 0 {
				transactionBinaryData := converter.BytesShift(&binaryData, txSize)
				dSha256Hash, err := crypto.DoubleHash(transactionBinaryData)
				if err != nil {
					log.Fatal(err)
				}
				dSha256Hash = converter.BinToHex(dSha256Hash)
				mrklSlice = append(mrklSlice, dSha256Hash)
			}

			// to exclude an attack on memory overflow
			if !first {
				if len(mrklSlice) > maxTxCount {
					return nil, ErrInfo(fmt.Errorf("[error] MAX_TX_COUNT (%v > %v)", len(mrklSlice), maxTxCount))
				}
			}
			if len(binaryData) == 0 {
				break
			}
		}
	} else {
		mrklSlice = append(mrklSlice, []byte("0"))
	}
	log.Debug("mrklSlice: %s", mrklSlice)
	if len(mrklSlice) == 0 {
		mrklSlice = append(mrklSlice, []byte("0"))
	}
	log.Debug("mrklSlice: %s", mrklSlice)
	return MerkleTreeRoot(mrklSlice), nil
}

// TypeInt returns the identifier of the embedded transaction
func TypeInt(txType string) int64 {
	for k, v := range consts.TxTypes {
		if v == txType {
			return int64(k)
		}
	}
	return 0
}

// GetNetworkTime returns the network time
func GetNetworkTime() (*time.Time, error) {

	ntpAddr := []string{"0.pool.ntp.org", "europe.pool.ntp.org", "asia.pool.ntp.org", "oceania.pool.ntp.org", "north-america.pool.ntp.org", "south-america.pool.ntp.org", "africa.pool.ntp.org"}
	for i := 0; i < len(ntpAddr); i++ {
		host := ntpAddr[i]
		raddr, err := net.ResolveUDPAddr("udp", host+":123")
		if err != nil {
			continue
		}

		data := make([]byte, 48)
		data[0] = 3<<3 | 3

		con, err := net.DialUDP("udp", nil, raddr)
		if err != nil {
			continue
		}

		defer con.Close()

		_, err = con.Write(data)
		if err != nil {
			continue
		}

		con.SetDeadline(time.Now().Add(5 * time.Second))

		_, err = con.Read(data)
		if err != nil {
			continue
		}

		var sec, frac uint64
		sec = uint64(data[43]) | uint64(data[42])<<8 | uint64(data[41])<<16 | uint64(data[40])<<24
		frac = uint64(data[47]) | uint64(data[46])<<8 | uint64(data[45])<<16 | uint64(data[44])<<24

		nsec := sec * 1e9
		nsec += (frac * 1e9) >> 32

		t := time.Date(1900, 1, 1, 0, 0, 0, 0, time.UTC).Add(time.Duration(nsec)).Local()
		return &t, nil
	}
	return nil, errors.New("unable connect to NTP")

}

// TCPConn connects to the address
func TCPConn(Addr string) (net.Conn, error) {
	conn, err := net.DialTimeout("tcp", Addr, 10*time.Second)
	if err != nil {
		return nil, ErrInfo(err)
	}
	conn.SetReadDeadline(time.Now().Add(consts.READ_TIMEOUT * time.Second))
	conn.SetWriteDeadline(time.Now().Add(consts.WRITE_TIMEOUT * time.Second))
	return conn, nil
}

// WriteSizeAndData writes []byte to the connection
func WriteSizeAndData(binaryData []byte, conn net.Conn) error {
	size := converter.DecToBin(len(binaryData), 4)
	_, err := conn.Write(size)
	if err != nil {
		return ErrInfo(err)
	}
	if len(binaryData) > 0 {
		_, err = conn.Write(binaryData)
		if err != nil {
			return ErrInfo(err)
		}
	}
	return nil
}

// GetCurrentDir returns the current directory
func GetCurrentDir() string {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		return "."
	}
	return dir
}

// GetBlockBody gets the block data
func GetBlockBody(host string, blockID int64, dataTypeBlockBody int64) ([]byte, error) {
	conn, err := TCPConn(host)
	if err != nil {
		return nil, ErrInfo(err)
	}
	defer conn.Close()

	log.Debug("dataTypeBlockBody: %v", dataTypeBlockBody)
	// send the type of data
	_, err = conn.Write(converter.DecToBin(dataTypeBlockBody, 2))
	if err != nil {
		return nil, ErrInfo(err)
	}

	log.Debug("blockID: %v", blockID)

	// send the number of a block
	_, err = conn.Write(converter.DecToBin(blockID, 4))
	if err != nil {
		return nil, ErrInfo(err)
	}

	// recieve the data size as a response that server wants to transfer
	buf := make([]byte, 4)
	n, err := conn.Read(buf)
	if err != nil {
		return nil, ErrInfo(err)
	}
	log.Debug("dataSize buf: %x / get: %v", buf, n)

	// if the data size is less than 10mb, we will receive them
	dataSize := converter.BinToDec(buf)
	var binaryBlock []byte
	if dataSize < 10485760 && dataSize > 0 {
		binaryBlock = make([]byte, dataSize)

		_, err = io.ReadFull(conn, binaryBlock)
		if err != nil {
			return nil, ErrInfo(err)
		}
	} else {
		return nil, ErrInfo("null block")
	}
	return binaryBlock, nil

}

// GetUpdVerAndURL downloads the information about the version
func GetUpdVerAndURL(host string) (updinfo *Update, err error) {

	update, err := GetHTTPTextAnswer(host + "/update.json")
	if len(update) > 0 {
		updateData := make(map[string]Update)
		err = json.Unmarshal([]byte(update), &updateData)
		if err != nil {
			return
		}
		if upd, ok := updateData[runtime.GOOS+`_`+runtime.GOARCH]; ok && version.Compare(upd.Version, consts.VERSION, ">") {
			updinfo = &upd
		}
	}
	return
}

// ShellExecute runs cmdline
func ShellExecute(cmdline string) {
	time.Sleep(500 * time.Millisecond)
	switch runtime.GOOS {
	case "linux":
		exec.Command("xdg-open", cmdline).Start()
	case "windows":
		exec.Command(`rundll32.exe`, `url.dll,FileProtocolHandler`, cmdline).Start()
	case "darwin":
		exec.Command("open", cmdline).Start()
	}
}

// AplaUpdate decompresses and updates executable file
func AplaUpdate(url string) error {
	//	GetUpdVerAndURL(host string) (updinfo *lib.Update, err error)

	zipfile := filepath.Join(*Dir, "apla.zip")
	/*	_, err := DownloadToFile(url, zipfile, 3600, nil, nil, "upd")
		if err != nil {
			return ErrInfo(err)
		}
		fmt.Println(zipfile)*/
	reader, err := zip.OpenReader(zipfile)
	if err != nil {
		return ErrInfo(err)
	}
	appname := filepath.Base(os.Args[0])
	tmpname := filepath.Join(*Dir, `tmp_`+appname)

	ftemp := reader.Reader.File
	f := ftemp[0]
	zipped, err := f.Open()
	if err != nil {
		return ErrInfo(err)
	}

	writer, err := os.OpenFile(tmpname, os.O_WRONLY|os.O_CREATE, f.Mode())
	if err != nil {
		return ErrInfo(err)
	}

	if _, err = io.Copy(writer, zipped); err != nil {
		return ErrInfo(err)
	}
	reader.Close()
	zipped.Close()
	writer.Close()

	/*	pwd, err := os.Getwd()
		if err != nil {
			return ErrInfo(err)
		}
		fmt.Print(pwd)*/

	folderPath, err := osext.ExecutableFolder()
	if err != nil {
		return ErrInfo(err)
	}

	old := ""
	if _, err := os.Stat(os.Args[0]); err == nil {
		old = os.Args[0]
	} else if _, err := os.Stat(filepath.Join(folderPath, appname)); err == nil {
		old = filepath.Join(folderPath, appname)
	} else {
		old = filepath.Join(*Dir, appname)
	}
	err = exec.Command(tmpname, "-oldFileName", old, "-dir", *Dir, "-oldVersion", consts.VERSION).Start()
	if err != nil {
		return ErrInfo(err)
	}
	return nil
}

// GetPrefix returns the prefix of the table. In this case it is checked that the prefix was global or matched
// with the identifier of the state
func GetPrefix(tableName, stateID string) (string, error) {
	s := strings.Split(tableName, "_")
	if len(s) < 2 {
		return "", ErrInfo("incorrect table name")
	}
	prefix := s[0]
	if prefix != "global" && prefix != stateID {
		return "", ErrInfo("incorrect table name")
	}
	return prefix, nil
}

// GetParent returns the information where the call of function happened
func GetParent() string {
	parent := ""
	for i := 2; ; i++ {
		name := ""
		if pc, _, num, ok := runtime.Caller(i); ok {
			name = filepath.Base(runtime.FuncForPC(pc).Name())
			file, line := runtime.FuncForPC(pc).FileLine(pc)
			if i > 5 || name == "runtime.goexit" {
				break
			} else {
				parent += fmt.Sprintf("%s:%d -> %s:%d / ", filepath.Base(file), line, name, num)
			}
		}
	}
	return parent
}

func GetTcpPort(config map[string]string) string {
	if port, ok := config["tcp_port"]; ok {
		return port
	}
	return consts.TCP_PORT
}

func Stop() {
	log.Debug("Stop()")
	log.Debug("DayLight Stop : %v", model.DBConn)

	stopDaemons := &model.StopDaemon{StopTime: time.Now().Unix()}
	err := stopDaemons.Create()
	if err != nil {
		log.Error("%v", utils.ErrInfo(err))
	}
	log.Debug("DayLight Stop")
}

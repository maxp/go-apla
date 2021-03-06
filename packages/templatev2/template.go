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

package templatev2

import (
	"encoding/json"
	"html"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/AplaProject/go-apla/packages/consts"
	"github.com/AplaProject/go-apla/packages/converter"
	"github.com/AplaProject/go-apla/packages/language"
	"github.com/AplaProject/go-apla/packages/model"

	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
)

const (
	tagText = `text`
	tagData = `data`
)

type node struct {
	Tag      string                 `json:"tag"`
	Attr     map[string]interface{} `json:"attr,omitempty"`
	Text     string                 `json:"text,omitempty"`
	Children []*node                `json:"children,omitempty"`
	Tail     []*node                `json:"tail,omitempty"`
}

type parFunc struct {
	Owner *node
	Node  *node
	Vars  *map[string]string
	Pars  *map[string]string
	Tails *[]*[]string
}

type nodeFunc func(par parFunc) string

type tplFunc struct {
	Func   nodeFunc // process function
	Full   nodeFunc // full process function
	Tag    string   // HTML tag
	Params string   // names of parameters
}

type tailInfo struct {
	tplFunc
	Last bool
}

type forTails struct {
	Tails map[string]tailInfo
}

func setAttr(par parFunc, name string) {
	if len((*par.Pars)[name]) > 0 {
		par.Node.Attr[strings.ToLower(name)] = (*par.Pars)[name]
	}
}

func setAllAttr(par parFunc) {
	for key, v := range *par.Pars {
		if key == `Params` || key == `PageParams` {
			imap := make(map[string]interface{})
			re := regexp.MustCompile(`(?is)(.*)\((.*)\)`)
			for _, parval := range strings.Split(v, `,`) {
				parval = strings.TrimSpace(parval)
				if len(parval) > 0 {
					if off := strings.IndexByte(parval, '='); off == -1 {
						imap[parval] = map[string]interface{}{
							`type`: `text`, `text`: parval}
					} else {
						val := strings.TrimSpace(parval[off+1:])
						if ret := re.FindStringSubmatch(val); len(ret) == 3 {
							plist := strings.Split(ret[2], `,`)
							for i, ilist := range plist {
								plist[i] = strings.TrimSpace(ilist)
							}
							imap[strings.TrimSpace(parval[:off])] = map[string]interface{}{
								`type`: ret[1], `params`: plist}
						} else {
							imap[strings.TrimSpace(parval[:off])] = map[string]interface{}{
								`type`: `text`, `text`: val}
						}
					}
				}
			}
			if len(imap) > 0 {
				par.Node.Attr[strings.ToLower(key)] = imap
			}
		} else if key != `Body` && key != `Data` && len(v) > 0 {
			par.Node.Attr[strings.ToLower(key)] = v
		}
	}
}

func ifValue(val string, vars *map[string]string) bool {
	var (
		sep   string
		owner node
	)

	if strings.IndexByte(val, '(') != -1 {
		process(val, &owner, vars)
		if len(owner.Children) > 0 {
			inode := owner.Children[0]
			if inode.Tag == tagText {
				val = inode.Text
			}
		} else {
			val = ``
		}

	}
	if strings.Index(val, `;base64`) < 0 {
		for _, item := range []string{`==`, `!=`, `<=`, `>=`, `<`, `>`} {
			if strings.Index(val, item) >= 0 {
				sep = item
				break
			}
		}
	}
	cond := []string{val}
	if len(sep) > 0 {
		cond = strings.SplitN(val, sep, 2)
		cond[0], cond[1] = strings.Trim(cond[0], `"`), strings.Trim(cond[1], `"`)
	}
	switch sep {
	case ``:
		return len(val) > 0 && val != `0` && val != `false`
	case `==`:
		return len(cond) == 2 && strings.TrimSpace(cond[0]) == strings.TrimSpace(cond[1])
	case `!=`:
		return len(cond) == 2 && strings.TrimSpace(cond[0]) != strings.TrimSpace(cond[1])
	case `>`, `<`, `<=`, `>=`:
		ret0, err := decimal.NewFromString(strings.TrimSpace(cond[0]))
		if err != nil {
			log.WithFields(log.Fields{"type": consts.ConvertionError, "error": err, "value": strings.TrimSpace(cond[0])}).Error("converting left condition from string to decimal")
		}
		ret1, err := decimal.NewFromString(strings.TrimSpace(cond[1]))
		if err != nil {
			log.WithFields(log.Fields{"type": consts.ConvertionError, "error": err, "value": strings.TrimSpace(cond[1])}).Error("converting right condition from string to decimal")
		}
		if len(cond) == 2 {
			var bin bool
			if sep == `>` || sep == `<=` {
				bin = ret0.Cmp(ret1) > 0
			} else {
				bin = ret0.Cmp(ret1) < 0
			}
			if sep == `<=` || sep == `>=` {
				bin = !bin
			}
			return bin
		}
	}
	return false
}

func replace(input string, level int, vars *map[string]string) string {
	if len(input) == 0 {
		return input
	}
	result := make([]rune, 0, utf8.RuneCountInString(input))
	isName := false
	name := make([]rune, 0, 128)
	syschar := '#'
	clearname := func() {
		result = append(append(result, syschar), name...)
		isName = false
		name = name[:0]
	}
	for _, r := range input {
		if r != syschar {
			if isName {
				name = append(name, r)
				if len(name) > 64 || r <= ' ' {
					clearname()
				}
			} else {
				result = append(result, r)
			}
			continue
		}
		if isName {
			if value, ok := (*vars)[string(name)]; ok {
				if level < 10 {
					value = replace(value, level+1, vars)
				}
				result = append(result, []rune(value)...)
				isName = false
			} else {
				result = append(append(result, syschar), name...)
			}
			name = name[:0]
		} else {
			isName = true
		}
	}
	if isName {
		result = append(append(result, syschar), name...)
	}
	return string(result)
}

func macro(input string, vars *map[string]string) string {
	if (*vars)[`_full`] == `1` || strings.IndexByte(input, '#') == -1 {
		return input
	}
	return replace(input, 0, vars)
}

func appendText(owner *node, text string) {
	if len(strings.TrimSpace(text)) == 0 {
		return
	}
	if len(text) > 0 {
		owner.Children = append(owner.Children, &node{Tag: tagText, Text: html.EscapeString(text)})
	}
}

func callFunc(curFunc *tplFunc, owner *node, vars *map[string]string, params *[]string, tailpars *[]*[]string) {
	var (
		out     string
		curNode node
	)
	pars := make(map[string]string)
	parFunc := parFunc{
		Vars: vars,
	}
	if curFunc.Params == `*` {
		for i, v := range *params {
			val := strings.TrimSpace(v)
			off := strings.IndexByte(val, ':')
			if off != -1 {
				pars[val[:off]] = macro(strings.Trim(val[off+1:], "\t\r\n \"`"), vars)
			} else {
				pars[strconv.Itoa(i)] = macro(val, vars)
			}
		}
	} else {
		for i, v := range strings.Split(curFunc.Params, `,`) {
			if i < len(*params) {
				val := macro(strings.TrimSpace((*params)[i]), vars)
				off := strings.IndexByte(val, ':')
				if off != -1 && strings.Contains(curFunc.Params, val[:off]) {
					cut := "\t\r\n \"`"
					if val[:off] == `Data` {
						cut = "\t\r\n "
					}
					pars[val[:off]] = strings.Trim(val[off+1:], cut)
				} else {
					pars[v] = val
				}
			} else if _, ok := pars[v]; !ok {
				pars[v] = ``
			}
		}
	}
	state := int(converter.StrToInt64((*vars)[`ecosystem_id`]))
	for i, v := range pars {
		pars[i] = language.LangMacro(v, state, (*vars)[`accept_lang`])
	}
	if len(curFunc.Tag) > 0 {
		curNode.Tag = curFunc.Tag
		curNode.Attr = make(map[string]interface{})
		if len(pars[`Body`]) > 0 {
			process(pars[`Body`], &curNode, vars)
		}
		parFunc.Owner = owner
		//		owner.Children = append(owner.Children, &curNode)
		parFunc.Node = &curNode
		parFunc.Tails = tailpars
	}
	parFunc.Pars = &pars
	if (*vars)[`_full`] == `1` {
		out = curFunc.Full(parFunc)
	} else {
		out = curFunc.Func(parFunc)
	}
	if len(out) > 0 {
		if len(owner.Children) > 0 && owner.Children[len(owner.Children)-1].Tag == tagText {
			owner.Children[len(owner.Children)-1].Text += out
		} else {
			appendText(owner, out)
		}
	}
}

func getFunc(input string, curFunc tplFunc) (*[]string, int, *[]*[]string) {
	var (
		curp, skip, off, mode, lenParams int
		quote                            bool
		pair, ch                         rune
		tailpar                          *[]*[]string
	)
	params := make([]string, 1)
	if curFunc.Params == `*` {
		lenParams = 0xff
	} else {
		lenParams = len(strings.Split(curFunc.Params, `,`))
	}
	level := 1
	if input[0] == '{' {
		mode = 1
	}
	skip = 1
main:
	for off, ch = range input {
		if skip > 0 {
			skip--
			continue
		}
		if pair > 0 {
			if ch != pair {
				params[curp] += string(ch)
			} else {
				if off+1 == len(input) || rune(input[off+1]) != pair {
					pair = 0
					if quote {
						params[curp] += string(ch)
						quote = false
					}
				} else {
					params[curp] += string(ch)
					skip = 1
				}
			}
			continue
		}
		if len(params[curp]) == 0 && mode == 0 && ch != modes[mode][1] && ch != ',' {
			if ch >= '!' {
				if ch == '"' || ch == '`' {
					pair = ch
				} else {
					params[curp] += string(ch)
				}
			}
			continue
		}

		switch ch {
		case '"', '`':
			if mode == 0 {
				pair = ch
				quote = true
			}
		case ',':
			if mode == 0 && level == 1 && len(params) < lenParams {
				params = append(params, ``)
				curp++
				continue
			}
		case modes[mode][0]:
			level++
		case modes[mode][1]:
			if level > 0 {
				level--
			}
			if level == 0 {
				if mode == 0 && (strings.Contains(curFunc.Params, `Body`) || strings.Contains(curFunc.Params, `Data`)) {
					var isBody bool
					next := off + 1
					for next < len(input) {
						if rune(input[next]) == modes[1][0] {
							isBody = true
							break
						}
						if rune(input[next]) == ' ' || rune(input[next]) == '\t' {
							next++
							continue
						}
						break
					}
					if isBody {
						mode = 1
						for _, keyp := range []string{`Body`, `Data`} {
							if strings.Contains(curFunc.Params, keyp) {
								params = append(params, keyp+`:`)
								break
							}
						}
						curp++
						skip = next - off
						level = 1
						continue
					}
				}
				for tail, ok := tails[curFunc.Tag]; ok && off+2 < len(input) && input[off+1] == '.'; {
					var found bool
					for key, tailFunc := range tail.Tails {
						next := off + 2
						if next < len(input) && strings.HasPrefix(input[next:], key) {
							var isTail bool
							next += len(key)
							for next < len(input) {
								if rune(input[next]) == '(' || rune(input[next]) == '{' {
									isTail = true
									break
								}
								if rune(input[next]) == ' ' || rune(input[next]) == '\t' {
									next++
									continue
								}
								break
							}
							if isTail {
								parTail, shift, _ := getFunc(input[next:], tailFunc.tplFunc)
								off = shift + next
								if tailpar == nil {
									fortail := make([]*[]string, 0)
									tailpar = &fortail
								}
								*parTail = append(*parTail, key)
								*tailpar = append(*tailpar, parTail)
								found = true
								if tailFunc.Last {
									break main
								}
								break
							}
						}
					}
					if !found {
						break
					}
				}
				break main
			}
		}
		params[curp] += string(ch)
		continue
	}
	return &params, utf8.RuneCountInString(input[:off]), tailpar
}

func process(input string, owner *node, vars *map[string]string) {
	var (
		nameOff, shift int
		curFunc        tplFunc
		isFunc         bool
		params         *[]string
		tailpars       *[]*[]string
	)
	name := make([]rune, 0, 128)
	for off, ch := range input {
		if shift > 0 {
			shift--
			continue
		}
		if ch == '(' {
			if curFunc, isFunc = funcs[string(name[nameOff:])]; isFunc {
				appendText(owner, string(name[:nameOff]))
				name = name[:0]
				nameOff = 0
				params, shift, tailpars = getFunc(input[off:], curFunc)
				callFunc(&curFunc, owner, vars, params, tailpars)
				for off+shift+3 < len(input) && input[off+shift+1:off+shift+3] == `.(` {
					var next int
					params, next, tailpars = getFunc(input[off+shift+2:], curFunc)
					callFunc(&curFunc, owner, vars, params, tailpars)
					shift += next + 2
				}
				continue
			}
		}
		if (ch < 'A' || ch > 'Z') && (ch < 'a' || ch > 'z') {
			nameOff = len(name) + 1
		}
		name = append(name, ch)
	}
	appendText(owner, string(name))
}

// Template2JSON converts templates to JSON data
func Template2JSON(input string, full bool, vars *map[string]string) []byte {
	if full {
		(*vars)[`_full`] = `1`
	} else {
		(*vars)[`_full`] = `0`
	}
	root := node{}
	process(input, &root, vars)
	if root.Children == nil {
		return []byte(`[]`)
	}
	out, err := json.Marshal(root.Children)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.JSONMarshallError, "error": err}).Error("marshalling template data to json")
		return []byte(err.Error())
	}
	return out
}

// StateParam returns the value of state parameters
func StateParam(idstate int64, name string) (string, error) {
	val, err := model.Single(`SELECT value FROM "`+converter.Int64ToStr(idstate)+`_parameters" WHERE name = ?`, name).String()
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("selecting state parameter")
	}
	return val, err
}

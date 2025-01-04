// Copyright (c) 2024 The horm-database Authors. All rights reserved.
// This file Author:  CaoHao <18500482693@163.com> .
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package logic

import (
	"fmt"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"

	"github.com/horm-database/common/errs"
	"github.com/horm-database/common/proto"
	"github.com/horm-database/common/structs"
	"github.com/horm-database/common/types"
	"github.com/horm-database/common/util"
	"github.com/horm-database/orm/obj"
	cs "github.com/horm-database/server/consts"
)

// 将引用替换为具体的值
func referHandle(dbType int, unit *proto.Unit, node *obj.Tree) (where, having, data map[string]interface{},
	datas []map[string]interface{}, key string, args []interface{}, isNil bool, err error) {
	if len(unit.Args) > 0 {
		args, isNil, err = argsReferHandle(node, unit.Args, unit.DataType)
		if err != nil || isNil {
			return
		}
	}

	if unit.Key != "" {
		key, isNil, err = keyReferHandle(node, unit.Key)
		if err != nil || isNil {
			return
		}
	}

	if hasReferer(unit.Where) {
		where, isNil, err = mapReferHandle(node, dbType, unit.Where)
		if err != nil || isNil {
			return
		}
	} else {
		where = unit.Where
	}

	if hasReferer(unit.Having) {
		having, isNil, err = mapReferHandle(node, dbType, unit.Having)
		if err != nil || isNil {
			return
		}
	} else {
		having = unit.Having
	}

	if hasReferer(unit.Data) {
		data, isNil, err = mapReferHandle(node, dbType, unit.Data)
		if err != nil || isNil {
			return
		}
	} else {
		data = unit.Data
	}

	var has bool
	hasIndex := map[int]bool{}
	for k, tmp := range unit.Datas {
		if hasReferer(tmp) {
			has = true
			hasIndex[k] = true
		}
	}

	if has {
		datas = make([]map[string]interface{}, len(unit.Datas))

		for k, tmp := range unit.Datas {
			if ok, _ := hasIndex[k]; ok {
				datas[k], isNil, err = mapReferHandle(node, dbType, tmp)
				if err != nil || isNil {
					return
				}
			} else {
				datas[k] = tmp
			}
		}
	} else {
		datas = unit.Datas
	}

	return
}

// mapReferHandle map 类型引用处理
func mapReferHandle(node *obj.Tree, dbType int, data map[string]interface{}) (map[string]interface{}, bool, error) {
	result := map[string]interface{}{}
	for k, v := range data {
		nk := util.RemoveComments(k)
		if types.FirstWord(nk, 1) == "@" { //引用
			referer, ok := v.(string)
			if !ok {
				return nil, false, errs.Newf(errs.ErrRefererMustBeString, "referer %s must be string", nk)
			}
			ret, isNil, err := findReferer(node, referer)
			if err != nil || isNil {
				return nil, isNil, err
			}

			result[k[1:]] = ret
		} else { // 递归处理引用
			rv := reflect.ValueOf(v)
			isRelation, isSliceAndOR, _ := util.GetRelation(dbType, nk, rv)
			if isRelation {
				if isSliceAndOR {
					rvLen := rv.Len()
					subMaps := make([]map[string]interface{}, rvLen)
					for index := 0; index < rvLen; index++ {
						arrVal := rv.Index(index)
						if types.IsNil(arrVal) {
							subMaps[index] = nil
						} else {
							mapV, err := types.ToMap(arrVal.Interface())
							if err != nil {
								return nil, false, err
							}

							subMap, isNil, err := mapReferHandle(node, dbType, mapV)
							if err != nil || isNil {
								return nil, isNil, err
							}
							subMaps[index] = subMap
						}
					}
					result[k] = subMaps
				} else {
					mapV, err := types.ToMap(v)
					if err != nil {
						return nil, false, err
					}

					ret, isNil, err := mapReferHandle(node, dbType, mapV)
					if err != nil || isNil {
						return nil, isNil, err
					}

					result[k] = ret
				}
			} else {
				result[k] = v
			}
		}
	}

	return result, false, nil
}

// key 引用处理
func keyReferHandle(node *obj.Tree, key string) (string, bool, error) {
	if argHasReferer(key) {
		ret, isNil, err := findReferer(node, getRefererParam(key))
		if err != nil || isNil {
			return "", isNil, err
		}

		return fmt.Sprintf("%v", ret), false, nil
	} else {
		return key, false, nil
	}
}

// args 引用处理
func argsReferHandle(node *obj.Tree, args []interface{},
	types map[string]structs.Type) (ret []interface{}, isNil bool, err error) {
	for k, arg := range args {
		str, ok := arg.(string)
		if ok && argHasReferer(str) {
			args[k], isNil, err = findReferer(node, getRefererParam(str))
			if err != nil || isNil {
				return nil, isNil, err
			}
		} else if len(types) > 0 {
			args[k], err = util.GetDataByType(strconv.Itoa(k), arg, types)
			if err != nil {
				return nil, false, err
			}
		}
	}

	return args, false, nil
}

// hasReferer 是否有引用
func hasReferer(data map[string]interface{}) bool {
	if len(data) == 0 {
		return false
	}

	for k, v := range data {
		nk := util.RemoveComments(k)
		if types.FirstWord(nk, 1) == "@" {
			return true
		} else {
			switch rv := v.(type) {
			case map[string]interface{}:
				if hasReferer(rv) {
					return true
				}
			}
		}
	}

	return false
}

// argHasReferer 参数是否有引用
// 如果用户内容首字符、末字符可能也包含 @{...}，请用户务必用 util.ArgRefererEscape 将首个 @ 转义为 \@。
func argHasReferer(arg string) bool {
	arg = strings.TrimSpace(arg)

	l := len(arg)
	if l <= 4 {
		return false
	}

	if arg[:2] == "@{" && arg[l-1:] == "}" {
		return true
	}

	return false
}

// getRefererParam 取出引用参数
func getRefererParam(arg string) string {
	arg = strings.TrimSpace(arg)
	return arg[2 : len(arg)-1]
}

// findReferer 找到被引用节点结果
func findReferer(node *obj.Tree, referer string) (ret interface{}, isNil bool, err error) {
	refererPath, refererField := pathAndField(referer)
	if !filepath.IsAbs(refererPath) { // 相对路径转化为绝对路径
		if len(refererPath) < 3 || refererPath[:3] != "../" {
			if len(refererPath) >= 2 && refererPath[:2] == "./" {
				refererPath = "." + refererPath
			} else {
				refererPath = "../" + refererPath
			}
		}

		path := node.GetReal().GetPath()
		refererPath = filepath.Join(path, refererPath)
	}

	for {
		var cur *obj.Tree
		if node.Last != nil {
			node = node.Last
		} else {
			cur = node
			node = node.Parent
		}

		if node == nil {
			return nil, false, errs.Newf(errs.ErrRefererNotFound, "not find referer unit [%s]", referer)
		}

		nodePath := node.GetReal().GetPath()

		if nodePath == refererPath {
			var result interface{}

			if node.Finished != cs.QueryFinishedYes {
				return nil, false, errs.Newf(errs.ErrRefererUnitNotExec, "referer unit is not execute [%s]", referer)
			}

			if cur != nil {
				if node.IsNil {
					return nil, true, nil
				}

				if !node.IsSuccess() {
					return nil, false, errs.Newf(errs.ErrRefererUnitFailed, "referer unit [%s] failed", referer)
				}

				result = cur.ParentRet
			} else {
				result, isNil, err = getRefererResult(node, referer)
				if err != nil || isNil {
					return nil, isNil, err
				}
			}

			ret, err = getFieldValue(refererField, result)
			if err != nil {
				return nil, false, err
			}

			if ret != nil {
				ret = reflect.Indirect(reflect.ValueOf(ret)).Interface()
			}
			return
		} else if strings.Index(refererPath, nodePath) == 0 {
			if len(node.SubQuery) == 0 {
				return nil, false, errs.Newf(errs.ErrRefererNotFound, "not find referer unit [%s]", referer)
			}

			relativeRefererPath := strings.TrimPrefix(refererPath, nodePath)
			if relativeRefererPath[0] == '/' {
				relativeRefererPath = relativeRefererPath[1:]
			}
			refererPathArr := strings.Split(relativeRefererPath, "/")

			var find bool
			var e error
			var retArr = []interface{}{}

			for _, sub := range node.SubQuery {
				findRefererByPath(sub, referer, refererPathArr, refererField, &retArr, &find, &e)
			}

			if !find {
				return nil, false, errs.Newf(errs.ErrRefererNotFound, "not find referer unit [%s]", referer)
			}

			if e != nil {
				return nil, false, e
			}

			if len(retArr) == 0 {
				return nil, true, nil
			}

			return retArr, false, nil
		}
	}
}

func findRefererByPath(cur *obj.Tree, referer string,
	refererPathArr []string, refererField string, ret *[]interface{}, find *bool, e *error) {
	for {
		dir := refererPathArr[0]
		if cur.GetReal().GetKey() == dir {
			if len(refererPathArr) == 1 {
				if cur.Finished != cs.QueryFinishedYes {
					*e = errs.Newf(errs.ErrRefererUnitNotExec, "referer unit is not execute [%s]", referer)
					return
				}
				*find = true

				if cur.IsNil {
					return
				}

				if cur.Error != nil {
					*e = cur.Error
					return
				}

				result, err := getFieldValue(refererField, cur.Result)
				if err != nil {
					*e = err
				}

				switch v := result.(type) {
				case []interface{}:
					*ret = append(*ret, v...)
				default:
					*ret = append(*ret, v)
				}

				return
			} else {
				if len(cur.SubQuery) == 0 {
					return
				}

				for _, sub := range cur.SubQuery {
					findRefererByPath(sub, referer, refererPathArr[1:], refererField, ret, find, e)
				}

				return
			}
		} else {
			cur = cur.Next
			if cur == nil {
				return
			}
		}
	}
}

func getRefererResult(node *obj.Tree, referer string) (interface{}, bool, error) {
	if node.IsNil {
		return nil, true, nil
	}

	if !node.IsSuccess() {
		return nil, false, errs.Newf(errs.ErrRefererUnitFailed, "referer unit [%s] failed", referer)
	}

	if node.HasSub {
		ret := []interface{}{}
		if len(node.SubQuery) == 1 {
			return node.SubQuery[0].ParentRet, false, nil
		} else {
			for _, subQuery := range node.SubQuery {
				ret = append(ret, subQuery.ParentRet)
			}
			return ret, false, nil
		}
	} else {
		return node.Result, false, nil
	}
}

// getFieldValue 找到被引用的 field
func getFieldValue(field string, refererResult interface{}) (interface{}, error) {
	if field == "" {
		return refererResult, nil
	}

	switch result := refererResult.(type) {
	case map[string]interface{}:
		ret, ok := result[field]
		if !ok {
			return nil, errs.Newf(errs.ErrRefererFieldNotExist, "referer result filed not exist")
		}
		return ret, nil
	case []interface{}:
		var ret []interface{}
		for _, v := range result {
			mv := v.(map[string]interface{})
			tmp, ok := mv[field]
			if !ok {
				return nil, errs.Newf(errs.ErrRefererFieldNotExist, "referer result filed not exist")
			}

			ret = append(ret, tmp)
		}

		if len(ret) == 0 {
			return ret[0], nil
		}
		return ret, nil
	case []map[string]interface{}:
		var ret []interface{}
		for _, v := range result {
			tmp, ok := v[field]
			if !ok {
				return nil, errs.Newf(errs.ErrRefererFieldNotExist, "referer result filed not exist")
			}

			ret = append(ret, tmp)
		}

		if len(ret) == 0 {
			return ret[0], nil
		}
		return ret, nil
	case *proto.ModRet:
		switch field {
		case "id", "last_insert_id":
			return result.ID, nil
		case "rows_affected":
			return result.RowAffected, nil
		case "version":
			return result.Version, nil
		case "status":
			return result.Status, nil
		case "reason":
			return result.Reason, nil
		default:
			return nil, errs.Newf(errs.ErrRefererFieldNotExist, "referer result filed not exist")
		}
	default:
		return nil, errs.Newf(errs.ErrRefererResultType, "referer result type is invalid")
	}
}

// pathAndField 获取路径和字段
func pathAndField(referer string) (refererPath, refererField string) {
	refererPath = referer

	if strings.Index(referer, "../") != -1 {
		referer = strings.Replace(referer, "../", "", -1)
	}

	if strings.Index(referer, "./") != -1 {
		referer = strings.Replace(referer, "./", "", -1)
	}

	index := strings.Index(referer, ".")
	if index != -1 {
		refererField = referer[index+1:]
		refererPath = strings.TrimSuffix(refererPath, "."+refererField)
	}

	return
}

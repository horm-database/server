package logic

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/horm-database/common/consts"
	"github.com/horm-database/common/errs"
	"github.com/horm-database/common/proto"
	"github.com/horm-database/common/types"
	"github.com/horm-database/common/util"
	"github.com/horm-database/orm/obj"
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
			ret, isNil, err := findReferer(node, v.(string))
			if err != nil || isNil {
				return nil, isNil, err
			}

			result[k] = ret
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
							mapV, err := types.InterfaceToMap(arrVal.Interface())
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
					mapV, err := types.InterfaceToMap(v)
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
	typeMap map[string]consts.DataType) (ret []interface{}, isNil bool, err error) {
	for k, arg := range args {
		str, ok := arg.(string)
		if ok && argHasReferer(str) {
			args[k], isNil, err = findReferer(node, getRefererParam(str))
			if err != nil || isNil {
				return nil, isNil, err
			}
		} else if len(typeMap) > 0 {
			args[k], err = util.GetDataByType(strconv.Itoa(k), arg, typeMap)
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
	return arg[2 : len(arg)-2]
}

// findReferer 找到被引用节点结果
func findReferer(node *obj.Tree, referer string) (ret interface{}, isNil bool, err error) {
	path := node.GetReal().GetPath()

	refererPath, refererField := util.PathAndField(referer)
	if types.FirstWord(refererPath, 1) != "/" { //将相对路径转化为引用路径
		lastIndex := strings.LastIndex(path, "/")
		refererPath = path[:lastIndex+1] + refererPath
	}

	for {
		var tmp *obj.Tree
		if node.Last != nil {
			node = node.Last
		} else {
			tmp = node
			node = node.Parent
		}

		if node == nil {
			return nil, false, errs.Newf(errs.RetNotFindReferer, "not find referer unit [%s]", referer)
		}

		if node.GetReal().GetPath() == refererPath {
			var result interface{}
			if tmp != nil {
				if node.IsNil {
					return nil, true, nil
				}

				if !node.IsSuccess() {
					return nil, false, errs.Newf(errs.RetRefererUnitFailed,
						"referer unit [%s] failed", referer)
				}

				result = tmp.ParentRet
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
			return
		}
	}
}

func getRefererResult(node *obj.Tree, referer string) (interface{}, bool, error) {
	if node.IsNil {
		return nil, true, nil
	}

	if !node.IsSuccess() {
		return nil, false, errs.Newf(errs.RetRefererUnitFailed, "referer unit [%s] failed", referer)
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
	switch result := refererResult.(type) {
	case map[string]interface{}:
		ret, ok := result[field]
		if !ok {
			return nil, errs.Newf(errs.RetRefererFieldNotExist, "referer result filed not exist")
		}
		return ret, nil
	case []interface{}:
		var ret []interface{}
		for _, v := range result {
			mv := v.(map[string]interface{})
			tmp, ok := mv[field]
			if !ok {
				return nil, errs.Newf(errs.RetRefererFieldNotExist, "referer result filed not exist")
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
				return nil, errs.Newf(errs.RetRefererFieldNotExist, "referer result filed not exist")
			}

			ret = append(ret, tmp)
		}

		if len(ret) == 0 {
			return ret[0], nil
		}
		return ret, nil
	case *proto.ModResult:
		if field == "last_insert_id" {
			return result.ID, nil
		} else if field == "rows_affected" {
			return result.RowAffected, nil
		} else {
			return nil, errs.Newf(errs.RetRefererFieldNotExist, "referer result filed not exist")
		}
	default:
		return nil, errs.Newf(errs.RetRefererResultType, "referer result type error")
	}
}

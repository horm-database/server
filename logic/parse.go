package logic

import (
	"context"
	"reflect"
	"strings"

	cc "github.com/horm-database/common/consts"
	"github.com/horm-database/common/errs"
	"github.com/horm-database/common/proto"
	"github.com/horm-database/common/proto/plugin"
	"github.com/horm-database/common/types"
	ut "github.com/horm-database/common/util"
	"github.com/horm-database/orm/obj"
	"github.com/horm-database/server/consts"
	"github.com/horm-database/server/model/table"
	"github.com/horm-database/server/util"
)

// Parse 请求解析
func Parse(ctx context.Context, head *proto.RequestHeader, units []*proto.Unit) (resp *proto.QueryResp, err error) {
	tree := &obj.Tree{}
	err = createTree(tree, nil, units, head)
	if err != nil {
		return nil, err
	}

	execute(ctx, head.Appid, tree)

	resp = &proto.QueryResp{}

	if head.QueryMode == cc.QueryModeSingle {
		resp.IsNil, resp.RspData, err = parseResult(tree)
		if err != nil {
			return nil, err
		}
	} else if head.QueryMode == cc.QueryModeParallel {
		resp.RspData, resp.RspNils, resp.RspErrs = parseParallelResult(tree)
	} else {
		resp.RspData = parseCompResult(tree)
	}

	return
}

// createTree 根据 unit 生成分析树
func createTree(head, parent *obj.Tree, units []*proto.Unit, requestHeader *proto.RequestHeader) error {
	var node *obj.Tree

	if parent == nil {
		node = head
	} else {
		node = &obj.Tree{
			Parent: parent,
		}

		parent.Sub = node
	}

	for i, unit := range units {
		if len(unit.Trans) == 1 { //事务下只有一个执行单元，当成普通执行单元
			unit = unit.Trans[0]
		}

		if len(unit.Trans) > 1 {
			if parent != nil && parent.TransInfo != nil { //父节点与所有子节点同属于一个事务，无需在子节点再重复定义事务节点
				return errs.Newf(errs.RetSameTransaction,
					"parent and all child nodes belong to the same transaction, no need to repeat the definition")
			}

			err := InitTree(node, unit, requestHeader)
			if err != nil {
				return err
			}

			err = createTransTree(node, unit.Trans, requestHeader)
			if err != nil {
				return err
			}
		} else {
			if parent != nil && parent.TransInfo != nil { // 父子节点同属于一个事务。
				node.TransInfo = parent.TransInfo
			}

			err := initTree(head, node, unit, requestHeader)
			if err != nil {
				return err
			}
		}

		if i < len(units)-1 { //生成下一个执行单元
			node.Next = &obj.Tree{
				Parent: parent,
				Last:   node,
			}
			node = node.Next
		}
	}

	return nil
}

// 创建事务节点
func createTransTree(head *obj.Tree, units []*proto.Unit, requestHeader *proto.RequestHeader) error {
	// 初始化事务信息
	var transInfo = &obj.TransInfo{}

	head.TransInfo = transInfo

	node := &obj.Tree{
		Parent:    head,
		TransInfo: transInfo,
	}

	transInfo.Trans = node

	for i, unit := range units {
		if len(unit.Trans) != 0 { //所有兄弟节点同属于一个事务，无需重复定义事务
			return errs.Newf(errs.RetSameTransaction,
				"all sibling nodes belong to the same transaction, no need to repeat the definition")
		}

		err := initTree(nil, node, unit, requestHeader)
		if err != nil {
			return err
		}

		if i < len(units)-1 { // 生成下一个执行单元
			node.Next = &obj.Tree{
				TransInfo: transInfo,
				Parent:    head,
				Last:      node,
			}
			node = node.Next
		}
	}

	return nil
}

func initTree(head, node *obj.Tree, unit *proto.Unit, requestHeader *proto.RequestHeader) error {
	err := InitTree(node, unit, requestHeader)
	if err != nil {
		return err
	}

	if len(unit.Sub) > 0 {
		err = createTree(head, node, unit.Sub, requestHeader)
		if err != nil {
			return err
		}
	}
	return nil
}

// execute 执行查询节点
func execute(ctx context.Context, appid uint64, node *obj.Tree) {
	for {
		realNode := node.GetReal()

		if realNode.IsTransaction() {
			if node.IsSub { // 子查询，生成子查询自有的 TransInfo，并生成新的事务节点，Real 共用原有的事务节点。
				node.TransInfo = &obj.TransInfo{
					Trans: &obj.Tree{
						Name:   realNode.TransInfo.Trans.Name,
						Parent: node,
						IsSub:  node.IsSub,
						Real:   realNode.TransInfo.Trans,
					},
				}
				// 事务查询节点的 TransInfo 都指向新生成的 TransInfo
				node.TransInfo.Trans.TransInfo = node.TransInfo
			}

			node.TransInfo.Trans.InTrans = true
			execute(ctx, appid, node.TransInfo.Trans)
			finishTrans(node)
			node.TransInfo.ResetTxClient() // 事务完成，重置事务
		} else {
			var ret interface{}

			if node.InTrans && node.TransInfo.Rollback { //事务需回滚，不再执行 query 语句
				node.Finished = consts.QueryFinishedRollback
			} else {
				ret, node.Detail, node.IsNil, node.Error = query(ctx, appid, node)
				node.Finished = consts.QueryFinishedYes
			}

			if node.InTrans && node.Error != nil {
				node.Finished = consts.QueryFinishedRollback
				node.TransInfo.Rollback = true // 事务待回滚
			}

			if realNode.Sub == nil { //不包含嵌套子查询
				node.Result = ret
			} else {
				if node.Finished == consts.QueryFinishedYes && node.IsSuccess() && !node.IsNil {
					node.HasSub = true

					var rv reflect.Value
					if realNode.IsArray() {
						rv = reflect.ValueOf(ret)
					} else {
						rv = reflect.ValueOf([]interface{}{ret})
					}
					l := rv.Len()

					node.SubQuery = make([]*obj.Tree, l)

					for k := 0; k < l; k++ {
						var result = types.Interface(rv.Index(k))

						node.SubQuery[k] = &obj.Tree{
							Name:      realNode.Sub.Name,
							IsSub:     true,
							InTrans:   node.InTrans,
							Real:      realNode.Sub, //根据查询结果数量，生成对应的子查询节点，Real 共用同一个。
							Parent:    node,
							ParentRet: result,
						}

						if node.InTrans {
							// 事务中的所有生成的新节点 TransInfo 都指向同一个 TransInfo
							node.SubQuery[k].TransInfo = node.TransInfo
						}

						execute(ctx, appid, node.SubQuery[k])
					}
				}
			}
		}

		if realNode.Next == nil {
			break
		}

		if node.IsSub { //每个子查询节点都应该有他自己的 Next 节点，而 Real 共用同一个。
			node.Next = &obj.Tree{
				Name:    realNode.Next.Name,
				Last:    node,
				Parent:  node.Parent,
				IsSub:   node.IsSub,
				Real:    realNode.Next,
				InTrans: node.InTrans,
			}

			if node.InTrans {
				// 事务中的所有生成的新节点 TransInfo 都指向同一个 TransInfo
				node.Next.TransInfo = node.TransInfo
			}
		}

		node = node.Next
	}
}

// 结束事务
func finishTrans(node *obj.Tree) {
	l := len(node.TransInfo.DBs)

	if node.TransInfo.Rollback {
		for i := 0; i < l; i++ {
			txClient := node.TransInfo.GetTxClient(node.TransInfo.DBs[i])
			err := txClient.FinishTx(node.Error)
			if err != nil {
				node.Error = errs.Newf(errs.RetTransaction,
					"rollback error: %v, source error is [%s]", err, node.Error.Error())
			}
		}
	} else {
		for i := 0; i < l; i++ {
			txClient := node.TransInfo.GetTxClient(node.TransInfo.DBs[i])
			err := txClient.FinishTx(nil)
			if err != nil {
				node.Error = errs.Newf(errs.RetTransaction,
					"rollback error: %v, source error is [%s]", err, node.Error.Error())
			}
		}
	}
}

// parseResult 解析单执行单元查询结果
func parseResult(node *obj.Tree) (bool, interface{}, error) {
	if !node.IsSuccess() {
		msg := errs.Msg(node.Error)
		if len(msg) > 5000 { //返回太长，截断
			node.Error = errs.SetErrorMsg(node.Error, msg[0:5000])
		}

		return false, nil, node.Error
	}

	if hasDetail(node.Detail) {
		result := proto.PageResult{Detail: node.Detail}
		if node.Result == nil {
			result.Data = []interface{}{}
		} else {
			result.Data, _ = types.InterfaceToArray(node.Result)
		}
		return node.IsNil, result, nil
	}

	return node.IsNil, node.Result, nil
}

// parseParallelResult 解析并行查询结果
func parseParallelResult(node *obj.Tree) (interface{}, map[string]bool, map[string]*proto.Error) {
	result := map[string]interface{}{}
	rspErrs := map[string]*proto.Error{}
	rspNil := map[string]bool{}

	for {
		key := node.GetReal().GetKey()
		if !node.IsSuccess() {
			rspErrs[key] = util.ErrorToRspError(node.Error)
		} else if node.IsNil {
			rspNil[key] = true
		} else {
			if hasDetail(node.Detail) {
				pageResult := proto.PageResult{Detail: node.Detail}
				if node.Result == nil {
					pageResult.Data = []interface{}{}
				} else {
					pageResult.Data, _ = types.InterfaceToArray(node.Result)
				}
				result[key] = node.Detail
				return result, rspNil, nil
			} else {
				result[key] = node.Result
			}
		}

		if node.Next == nil {
			break
		}

		node = node.Next
	}

	return result, rspNil, rspErrs
}

// parseCompResult 解析混合查询结果
func parseCompResult(node *obj.Tree) map[string]interface{} {
	result := map[string]interface{}{}
	if node.IsSub {
		result, _ = types.InterfaceToMap(node.ParentRet)
	}

	for {
		realNode := node.GetReal()
		key := realNode.GetKey()

		if !node.IsSuccess() {
			result[key] = proto.CompResult{Error: util.ErrorToRspError(node.Error)}
		} else if node.IsNil {
			result[key] = proto.CompResult{IsNil: true}
		} else {
			if realNode.IsTransaction() {
				result[key] = parseCompResult(node.TransInfo.Trans)
			} else if node.HasSub {
				if realNode.IsArray() {
					ret := make([]interface{}, len(node.SubQuery))
					for k, subQuery := range node.SubQuery {
						ret[k] = parseCompResult(subQuery)
					}

					compRet := proto.CompResult{Data: ret}
					if hasDetail(node.Detail) {
						compRet.Detail = node.Detail
					}

					result[key] = compRet
				} else {
					result[key] = proto.CompResult{Data: parseCompResult(node.SubQuery[0])}
				}
			} else {
				compRet := proto.CompResult{Data: node.Result}
				if hasDetail(node.Detail) {
					compRet.Detail = node.Detail
				}

				result[key] = compRet
			}
		}

		if node.Next == nil {
			break
		}

		node = node.Next
	}

	return result
}

// InitTree 初始化分析树
func InitTree(node *obj.Tree, unit *proto.Unit, requestHeader *proto.RequestHeader) error {
	if unit.Extend == nil {
		unit.Extend = map[string]interface{}{}
	}

	if requestHeader != nil {
		unit.Extend["request_header"] = &plugin.Header{
			RequestId: requestHeader.RequestId,
			TraceId:   requestHeader.TraceId,
			Timestamp: requestHeader.Timestamp,
			Timeout:   requestHeader.Timeout,
			Caller:    requestHeader.Caller,
			Appid:     requestHeader.Appid,
			Ip:        requestHeader.Ip,
		}
	}

	node.Name = unit.Name
	node.Unit = unit

	property := obj.Property{}
	property.Op = strings.ToLower(unit.Op)
	property.Name, property.Alias = ut.Alias(unit.Name)
	if property.Name == "" {
		return errs.Newf(errs.RetUnitNameEmpty, "unit name is empty")
	}

	if property.Alias == "" {
		property.Key = property.Name
	} else {
		property.Key = property.Alias
	}

	if node.Parent == nil {
		property.Path = property.Key
	} else {
		property.Path = node.Parent.GetPath() + "/" + property.Key
	}

	if hasDuplicateKey(node, property.Key) {
		return errs.Newf(errs.RetRepeatNameAlias, "has repeat name or alias in same layer, name=[%s]", property.Path)
	}

	if len(unit.Trans) == 0 {
		tables, table, db, ambiguous := table.GetTableAndDB(property.Name, unit.Shard)
		if ambiguous {
			return errs.Newf(errs.RetNameAmbiguity,
				"[%s] there are multiple tables with the same name, please input namespace to separate", property.Path)
		}

		property.Table = table

		if len(tables) == 0 || table == nil || db == nil {
			return errs.Newf(errs.RetNotFindName, "[%s] not find table or db, name is %s", property.Path, unit.Name)
		}

		property.Tables = tables
		property.DB = db
	}

	node.Property = &property

	return nil
}

// hasDuplicateKey 判断同一层级是否有重复 key
func hasDuplicateKey(node *obj.Tree, key string) bool {
	for {
		if node.Last == nil {
			return false
		} else if node.Last.GetKey() == key {
			return true
		}
		node = node.Last
	}
}

func hasDetail(detail *proto.Detail) bool {
	if detail == nil {
		return false
	}

	if detail.Size > 0 || detail.Scroll != nil || len(detail.Extras) > 0 {
		return true
	}

	return false
}

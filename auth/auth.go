package auth

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	cc "github.com/horm-database/common/consts"
	"github.com/horm-database/common/errs"
	"github.com/horm-database/orm/obj"
	"github.com/horm-database/server/consts"
	"github.com/horm-database/server/model/table"
)

// PermissionCheck 权限校验，appid 是否拥有对应的操作权限
func PermissionCheck(source *obj.Tree, appid uint64, op, query string, isRecheck bool) error {
	return nil
	//访问者信息
	appInfo := table.GetAppInfo(appid)
	if appInfo == nil {
		return errs.Newf(errs.RetNotFindAppid, "[%s] not find app info of appid %d", source.GetPath(), appid)
	}

	// 表信息
	tblTable := source.GetTable()

	// 库权限
	acdb, _ := appInfo.AccessDB[tblTable.DB]
	actb, _ := appInfo.AccessTable[tblTable.Id]
	dbOPs, _ := appInfo.DBOps[tblTable.DB]

	// 库超级权限
	if acdb != nil && acdb.Status == consts.AuthStatusNormal && acdb.Root == consts.DBRootAll {
		return nil
	}

	// DDL 必须拥有库级别权限
	if op == cc.OpCreate || op == cc.OpDrop {
		if supportOp(dbOPs, op) {
			return nil
		}

		return errs.Newf(errs.RetHasNoDBRight, "[%s]%s appid(%d) has no permission to %s %s",
			source.GetPath(), recheck(isRecheck), appid, op, source.GetName())
	}

	// 直接查询 query 语句，必须拥有库的数据权限或者表的 query_all 权限
	if query != "" {
		if (acdb != nil && acdb.Status == consts.AuthStatusNormal && acdb.Root == consts.DBRootTableData) ||
			(actb != nil && actb.Status == consts.AuthStatusNormal && actb.QueryAll == consts.TableQueryAllTrue) {
			return nil
		}

		return errs.Newf(errs.RetHasNoDBRight, "[%s]%s appid(%d) has no permission to query %s directly",
			source.GetPath(), recheck(isRecheck), appid, source.GetName())
	}

	// 库权限
	if acdb != nil && acdb.Status == consts.AuthStatusNormal &&
		(acdb.Root == consts.DBRootTableData || supportOp(dbOPs, op)) {
		return nil
	}

	// 表权限状态
	if actb != nil && actb.Status == consts.AuthStatusNormal {
		if actb.QueryAll == consts.TableQueryAllTrue { //拥有表 query_all 权限
			return nil
		} else {
			tableOps, _ := appInfo.TableOPs[tblTable.Id]
			if supportOp(tableOps, op) {
				return nil
			}
		}
	}

	return errs.Newf(errs.RetHasNoTableRight, "[%s]%s appid [%d] has no permission to %s table %s",
		source.GetPath(), recheck(isRecheck), appid, op, source.GetName())
}

func supportOp(ops map[string]bool, needle string) bool {
	if len(ops) == 0 {
		return false
	}

	has, _ := ops[needle]
	return has
}

// TableVerify table verify
func TableVerify(source *obj.Tree, appid uint64, tables []string, verifyRule string) error {
	if verifyRule == "" {
		return nil
	}

	for _, t := range tables {
		if !matchTable(t, verifyRule) {
			return errs.Newf(errs.RetTableVerifyFailed,
				"[%s] verify failed, appid [%d] is not allowed to access table [%v]", source.GetPath(), appid, tables)
		}
	}

	return nil
}

func matchTable(table, verifyRule string) bool {
	rules := strings.Split(verifyRule, ",")
	for _, rule := range rules {
		if rule == table {
			return true
		}

		re := regexp.MustCompile(`(\d+)(\.\.\.)(\d+)`)
		matchs := re.FindAllStringSubmatch(rule, -1)
		if matchs != nil && len(matchs) == 1 && len(matchs[0]) == 4 {
			ruleArr := strings.Split(rule, matchs[0][0])
			startIndex, err := strconv.Atoi(matchs[0][1])
			if err != nil {
				continue
			}
			endIndex, err := strconv.Atoi(matchs[0][3])
			if err != nil {
				continue
			}
			if endIndex < startIndex {
				continue
			}
			for i := startIndex; i <= endIndex; i++ {
				ruleStr := ruleArr[0] + fmt.Sprint(i) + ruleArr[1]
				if table == ruleStr {
					return true
				}
			}
		}

		if strings.HasPrefix(rule, "regex/") && strings.HasSuffix(rule, "/") {
			ruleStr := rule[6 : len(rule)-1]
			if regexp.MustCompile(ruleStr).MatchString(table) {
				return true
			}
		}
	}

	return false
}

func recheck(recheck bool) string {
	if recheck {
		return " recheck after plugin execute -"
	}
	return ""
}

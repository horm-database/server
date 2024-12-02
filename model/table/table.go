package table

import (
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/horm-database/common/errs"
	"github.com/horm-database/common/json"
	"github.com/horm-database/common/log"
	"github.com/horm-database/common/util"
	"github.com/horm-database/orm/obj"
	"github.com/horm-database/server/consts"
	"github.com/horm-database/server/filter/conf"
	sc "github.com/horm-database/server/srv/codec"
)

var (
	filterLock        = new(sync.RWMutex)
	filter            = map[int]*TblFilter{}
	tablePreFilters   = map[int][]*TblTableFilter{}
	tablePostFilters  = map[int][]*TblTableFilter{}
	tableDeferFilters = map[int][]*TblTableFilter{}

	dbLock     = new(sync.RWMutex)
	dbMap      = map[int]*obj.TblDB{}
	dbNameMap  = map[string]*obj.TblDB{}
	tableMap   = map[string]map[int]*obj.TblTable{}
	appInfoMap = map[uint64]*AppInfo{}

	wsLock    = new(sync.RWMutex)
	workspace = TblWorkspace{}
)

// AppInfo 数据访问者信息
type AppInfo struct {
	Info        *TblAppInfo             // 应用信息
	AccessDB    map[int]*TblAccessDB    // 可以访问的仓库
	AccessTable map[int]*TblAccessTable // 可以访问的表
	DBOps       map[int]map[string]bool // 支持的库操作
	TableOPs    map[int]map[string]bool // 支持的表操作
}

// GetTables 获取所有表
func GetTables() map[string]map[int]*obj.TblTable {
	dbLock.RLock()
	defer dbLock.RUnlock()
	return tableMap
}

func GetTablesDB(t *obj.TblTable) *obj.TblDB {
	dbLock.RLock()
	ret := dbMap[t.DB]
	dbLock.RUnlock()
	return ret
}

// GetTableAndDB 根据数据名称（执行单元名）返回表名/索引名/redis、及其数据库信息
func GetTableAndDB(name string, shard []string) (tables []string,
	tblTable *obj.TblTable, db *obj.TblDB, ambiguous bool) {
	dbLock.RLock()
	defer dbLock.RUnlock()

	dbname, tableName := util.Namespace(name)

	tblTables, _ := tableMap[tableName]

	if dbname == "" {
		if len(tables) > 1 { //有多个同名表
			return nil, nil, nil, true
		}

		for _, tmp := range tblTables {
			tblTable = tmp
		}
	} else {
		db, _ := dbNameMap[dbname]
		if db != nil {
			tblTable = tblTables[db.Id]
		}
	}

	if tblTable == nil {
		return
	}

	db = dbMap[tblTable.DB]

	if len(shard) > 0 {
		tables = shard
	} else {
		tables = []string{tableName}
	}

	return
}

func SetWorkspace(ws *TblWorkspace) {
	wsLock.Lock()
	defer wsLock.Unlock()
	workspace = *ws
}

func GetWorkspace() TblWorkspace {
	wsLock.RLock()
	defer wsLock.RUnlock()
	return workspace
}

func GetAppInfo(appid uint64) *AppInfo {
	dbLock.RLock()
	defer dbLock.RUnlock()
	return appInfoMap[appid]
}

func GetTableFilters(tableID int, typ int8) []*TblTableFilter {
	filterLock.RLock()
	defer filterLock.RUnlock()

	switch typ {
	case consts.PreFilter:
		return tablePreFilters[tableID]
	case consts.PostFilter:
		return tablePostFilters[tableID]
	default:
		return tableDeferFilters[tableID]
	}
}

func GetFilter(id int) *TblFilter {
	filterLock.RLock()
	defer filterLock.RUnlock()
	return filter[id]
}

func SetDB(db *obj.TblDB) {
	dbLock.Lock()
	defer dbLock.Unlock()

	var err error

	db.Addr = &util.DBAddress{
		Type:    db.Type,
		Version: db.Version,
		Network: db.Network,
		Address: db.Address,

		WriteTimeout: db.WriteTimeoutTmp,
		ReadTimeout:  db.ReadTimeoutTmp,
		WarnTimeout:  db.WarnTimeoutTmp,
		OmitError:    db.OmitErrorTmp,
		Debug:        db.DebugTmp,
	}

	err = util.ParseConnFromAddress(db.Addr)
	if err != nil {
		log.Errorf(sc.GCtx, errs.RetDBAddressParseError, "parse db %s address error: %v", db.Name, err)
	}

	dbMap[db.Id] = db
	dbNameMap[db.Name] = db
}

func SetTable(table *obj.TblTable) {
	dbLock.Lock()
	defer dbLock.Unlock()

	_, exists := tableMap[table.Name]
	if !exists {
		tableMap[table.Name] = map[int]*obj.TblTable{}
	}

	tableMap[table.Name][table.DB] = table
}

func SetAppInfo(info *TblAppInfo) {
	dbLock.Lock()
	defer dbLock.Unlock()

	appInfoMap[info.Appid] = &AppInfo{
		Info:        info,
		AccessDB:    map[int]*TblAccessDB{},
		AccessTable: map[int]*TblAccessTable{},
		TableOPs:    map[int]map[string]bool{},
		DBOps:       map[int]map[string]bool{},
	}
}

func SetAccessDB(accessDB *TblAccessDB) {
	dbLock.Lock()
	defer dbLock.Unlock()

	appInfo, ok := appInfoMap[accessDB.Appid]
	if ok {
		appInfo.AccessDB[accessDB.DB] = accessDB
		appInfo.DBOps[accessDB.DB] = map[string]bool{}
		ops := strings.Split(accessDB.Op, ",")
		for _, op := range ops {
			appInfo.DBOps[accessDB.DB][op] = true
		}
	}
}

func SetAccessTable(accessTable *TblAccessTable) {
	dbLock.Lock()
	defer dbLock.Unlock()

	appInfo, ok := appInfoMap[accessTable.Appid]
	if ok {
		appInfo.AccessTable[accessTable.TableId] = accessTable
		appInfo.TableOPs[accessTable.TableId] = map[string]bool{}
		ops := strings.Split(accessTable.Op, ",")
		for _, op := range ops {
			appInfo.TableOPs[accessTable.TableId][op] = true
		}
	}
}

func SetFilter(f *TblFilter) {
	filterLock.Lock()
	defer filterLock.Unlock()

	filter[f.Id] = f
}

func InitTableFilter(tableFitlers []*TblTableFilter) error {
	filterLock.Lock()
	defer filterLock.Unlock()

	for _, tf := range tableFitlers {
		tf.Conf = getFilterConfig(tf.FilterId, tf.FilterVersion, tf.Config)
		tf.ScheduleConf = &conf.ScheduleConfig{}
		if tf.ScheduleConfig != "" {
			err := json.Api.Unmarshal([]byte(tf.ScheduleConfig), &tf.ScheduleConf)
			if err != nil {
				log.Errorf(sc.GCtx, errs.RetFilterConfigDecode,
					"unmarshal filter schedule config error=[%v], filter_id=[%d], filter_version=[%d], schedule_config=[%s]",
					err, tf.FilterId, tf.FilterVersion, tf.ScheduleConfig)
			}
		}

		switch tf.Type {
		case consts.PreFilter:
			tablePreFilters[tf.TableId] = append(tablePreFilters[tf.TableId], tf)
		case consts.PostFilter:
			tablePostFilters[tf.TableId] = append(tablePostFilters[tf.TableId], tf)
		case consts.DeferFilter:
			tableDeferFilters[tf.TableId] = append(tableDeferFilters[tf.TableId], tf)
		default:
			return errors.New(
				fmt.Sprintf("not find filter type: %d, table_id=%d and filter_id=%d", tf.Type, tf.TableId, tf.FilterId))
		}
	}

	for k := range tablePreFilters {
		sortedTablePreFilters, err := SortTableFilters("table pre-filter", tablePreFilters[k])
		if err != nil {
			return err
		}
		tablePreFilters[k] = sortedTablePreFilters
	}

	for k := range tablePostFilters {
		sortedTablePostFilters, err := SortTableFilters("table post-filter", tablePostFilters[k])
		if err != nil {
			return err
		}
		tablePostFilters[k] = sortedTablePostFilters
	}

	for k := range tableDeferFilters {
		sortedTableDeferFilters, err := SortTableFilters("table defer-filter", tableDeferFilters[k])
		if err != nil {
			return err
		}
		tableDeferFilters[k] = sortedTableDeferFilters
	}

	return nil
}

func UpdateDBInfo(dbs []*obj.TblDB, tables []*obj.TblTable,
	appInfos []*TblAppInfo, accessDBs []*TblAccessDB, accessTables []*TblAccessTable) {
	dbLock.Lock()
	defer dbLock.Unlock()

	//更新数据库信息
	if len(dbs) > 0 {
		for _, db := range dbs {
			SetDB(db)
		}
	}

	//更新表信息
	if len(tables) > 0 {
		for _, tbl := range tables {
			SetTable(tbl)
		}
	}

	//更新访问者信息
	for _, info := range appInfos {
		if _, ok := appInfoMap[info.Appid]; ok {
			appInfoMap[info.Appid].Info = info
		} else {
			appInfoMap[info.Appid] = &AppInfo{
				Info:        info,
				AccessDB:    map[int]*TblAccessDB{},
				AccessTable: map[int]*TblAccessTable{},
				TableOPs:    map[int]map[string]bool{},
				DBOps:       map[int]map[string]bool{},
			}
		}
	}

	for _, accessDB := range accessDBs {
		appInfo, ok := appInfoMap[accessDB.Appid]
		if ok {
			appInfo.AccessDB[accessDB.DB] = accessDB

			appInfo.DBOps[accessDB.DB] = map[string]bool{}
			ops := strings.Split(accessDB.Op, ",")
			for _, op := range ops {
				appInfo.DBOps[accessDB.DB][op] = true
			}
		}
	}

	for _, accessTable := range accessTables {
		appInfo, ok := appInfoMap[accessTable.Appid]
		if ok {
			appInfo.AccessTable[accessTable.TableId] = accessTable

			appInfo.TableOPs[accessTable.TableId] = map[string]bool{}
			ops := strings.Split(accessTable.Op, ",")
			for _, op := range ops {
				appInfo.TableOPs[accessTable.TableId][op] = true
			}
		}
	}
}

func getFilterConfig(filterID, filterVersion int, config string) map[string]interface{} {
	result := map[string]interface{}{}

	if config == "" {
		return result
	}

	err := json.Api.Unmarshal([]byte(config), &result)
	if err != nil {
		log.Errorf(sc.GCtx, errs.RetFilterConfigDecode,
			"unmarshal filter config error=[%v], filter_id=[%d], filter_version=[%d], config=[%s]",
			err, filterID, filterVersion, config)
		return nil
	}

	return result
}

func SortTableFilters(typ string, tableFilters []*TblTableFilter) ([]*TblTableFilter, error) {
	if len(tableFilters) == 0 {
		return []*TblTableFilter{}, nil
	}

	var head *TblTableFilter

	for _, tableFilter := range tableFilters {
		if tableFilter.Front == 0 {
			head = tableFilter
			break
		}
	}

	if head == nil {
		return nil, errs.Newf(errs.RetFilterFrontNotFind,
			"table_id %d not find head of %s", tableFilters[0].TableId, typ)
	}

	ret := []*TblTableFilter{}
	ret = append(ret, head)

	currentTableFilter := head
	for i := 0; i < len(tableFilters)-1; i++ {
		frontTableFilter := findFrontTableFilter(currentTableFilter, tableFilters)
		if frontTableFilter == nil {
			return nil, errs.Newf(errs.RetFilterFrontNotFind, "%s %d not find front table_filter=%d",
				typ, currentTableFilter.Id, currentTableFilter.Front)
		}

		currentTableFilter = frontTableFilter
		ret = append(ret, currentTableFilter)
	}

	return ret, nil
}

func findFrontTableFilter(currentTableFilter *TblTableFilter, tableFilters []*TblTableFilter) *TblTableFilter {
	for _, tableFilter := range tableFilters {
		if currentTableFilter.TableId == tableFilter.Front {
			return tableFilter
		}
	}
	return nil
}

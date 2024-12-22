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
	"github.com/horm-database/server/plugin/conf"
	sc "github.com/horm-database/server/srv/codec"
)

var (
	pluginLock        = new(sync.RWMutex)
	plugin            = map[int]*TblPlugin{}
	tablePrePlugins   = map[int][]*TblTablePlugin{}
	tablePostPlugins  = map[int][]*TblTablePlugin{}
	tableDeferPlugins = map[int][]*TblTablePlugin{}

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

func GetTablePlugins(tableID int, typ int8) []*TblTablePlugin {
	pluginLock.RLock()
	defer pluginLock.RUnlock()

	switch typ {
	case consts.PrePlugin:
		return tablePrePlugins[tableID]
	case consts.PostPlugin:
		return tablePostPlugins[tableID]
	default:
		return tableDeferPlugins[tableID]
	}
}

func GetPlugin(id int) *TblPlugin {
	pluginLock.RLock()
	defer pluginLock.RUnlock()
	return plugin[id]
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
		log.Errorf(sc.GCtx, errs.ErrDBAddressParse, "parse db %s address error: %v", db.Name, err)
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

func SetPlugin(f *TblPlugin) {
	pluginLock.Lock()
	defer pluginLock.Unlock()

	plugin[f.Id] = f
}

func InitTablePlugin(tableFitlers []*TblTablePlugin) error {
	pluginLock.Lock()
	defer pluginLock.Unlock()

	for _, tf := range tableFitlers {
		tf.Conf = getPluginConfig(tf.PluginID, tf.PluginVersion, tf.Config)
		tf.ScheduleConf = &conf.ScheduleConfig{}
		if tf.ScheduleConfig != "" {
			err := json.Api.Unmarshal([]byte(tf.ScheduleConfig), &tf.ScheduleConf)
			if err != nil {
				log.Errorf(sc.GCtx, errs.ErrPluginConfigDecode,
					"unmarshal plugin schedule config error=[%v], plugin_id=[%d], plugin_version=[%d], schedule_config=[%s]",
					err, tf.PluginID, tf.PluginVersion, tf.ScheduleConfig)
			}
		}

		switch tf.Type {
		case consts.PrePlugin:
			tablePrePlugins[tf.TableId] = append(tablePrePlugins[tf.TableId], tf)
		case consts.PostPlugin:
			tablePostPlugins[tf.TableId] = append(tablePostPlugins[tf.TableId], tf)
		case consts.DeferPlugin:
			tableDeferPlugins[tf.TableId] = append(tableDeferPlugins[tf.TableId], tf)
		default:
			return errors.New(
				fmt.Sprintf("not find plugin type: %d, table_id=%d and plugin_id=%d", tf.Type, tf.TableId, tf.PluginID))
		}
	}

	for k := range tablePrePlugins {
		sortedTablePrePlugins, err := SortTablePlugins("table pre-plugin", tablePrePlugins[k])
		if err != nil {
			return err
		}
		tablePrePlugins[k] = sortedTablePrePlugins
	}

	for k := range tablePostPlugins {
		sortedTablePostPlugins, err := SortTablePlugins("table post-plugin", tablePostPlugins[k])
		if err != nil {
			return err
		}
		tablePostPlugins[k] = sortedTablePostPlugins
	}

	for k := range tableDeferPlugins {
		sortedTableDeferPlugins, err := SortTablePlugins("table defer-plugin", tableDeferPlugins[k])
		if err != nil {
			return err
		}
		tableDeferPlugins[k] = sortedTableDeferPlugins
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

func getPluginConfig(pluginID, pluginVersion int, config string) map[string]interface{} {
	result := map[string]interface{}{}

	if config == "" {
		return result
	}

	err := json.Api.Unmarshal([]byte(config), &result)
	if err != nil {
		log.Errorf(sc.GCtx, errs.ErrPluginConfigDecode,
			"unmarshal plugin config error=[%v], plugin_id=[%d], plugin_version=[%d], config=[%s]",
			err, pluginID, pluginVersion, config)
		return nil
	}

	return result
}

func SortTablePlugins(typ string, tablePlugins []*TblTablePlugin) ([]*TblTablePlugin, error) {
	if len(tablePlugins) == 0 {
		return []*TblTablePlugin{}, nil
	}

	var head *TblTablePlugin

	for _, tablePlugin := range tablePlugins {
		if tablePlugin.Front == 0 {
			head = tablePlugin
			break
		}
	}

	if head == nil {
		return nil, errs.Newf(errs.ErrPrefixPluginNotFount,
			"table_id %d not find head of %s", tablePlugins[0].TableId, typ)
	}

	ret := []*TblTablePlugin{}
	ret = append(ret, head)

	currentTablePlugin := head
	for i := 0; i < len(tablePlugins)-1; i++ {
		frontTablePlugin := findFrontTablePlugin(currentTablePlugin, tablePlugins)
		if frontTablePlugin == nil {
			return nil, errs.Newf(errs.ErrPrefixPluginNotFount, "%s %d not find prefix table_plugin=%d",
				typ, currentTablePlugin.Id, currentTablePlugin.Front)
		}

		currentTablePlugin = frontTablePlugin
		ret = append(ret, currentTablePlugin)
	}

	return ret, nil
}

func findFrontTablePlugin(currentTablePlugin *TblTablePlugin, tablePlugins []*TblTablePlugin) *TblTablePlugin {
	for _, tablePlugin := range tablePlugins {
		if currentTablePlugin.TableId == tablePlugin.Front {
			return tablePlugin
		}
	}
	return nil
}

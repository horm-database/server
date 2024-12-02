package model

import (
	"context"
	"fmt"
	"time"

	"github.com/horm-database/common/snowflake"
	"github.com/horm-database/go-horm/horm"
	"github.com/horm-database/orm"
	"github.com/horm-database/orm/obj"
	"github.com/horm-database/server/consts"
	"github.com/horm-database/server/model/table"
)

var SyncTime time.Time

func Init(ctx context.Context, machineID int) {
	snowflake.SetMachineID(machineID)

	SyncTime = time.Now()

	initWorkspace(ctx)
	initDBConfig(ctx)
	initTableConfig(ctx)
	initAppInfo(ctx)
	initFilter(ctx)
}

// 初始化 workspace 信息
func initWorkspace(ctx context.Context) {
	//初始化执行实例信息
	var workspace table.TblWorkspace
	_, err := orm.NewORM(consts.DBConfigName).
		Name("tbl_workspace").Find().Exec(ctx, &workspace)
	if err != nil {
		panic(fmt.Errorf("initial tbl_db from db error: %s", err))
	}

	if workspace.Id == 0 {
		return
	}

	table.SetWorkspace(&workspace)
}

// 初始化数据库配置表到 body
func initDBConfig(ctx context.Context) {
	//初始化执行实例信息
	dbs := []*obj.TblDB{}
	_, err := orm.NewORM(consts.DBConfigName).
		Name("tbl_db").FindAll().Exec(ctx, &dbs)
	if err != nil {
		panic(fmt.Errorf("initial tbl_db from db error: %s", err))
	}

	if len(dbs) == 0 {
		return
	}

	for _, db := range dbs {
		table.SetDB(db)
	}
}

// 初始化表配置信息到 body
func initTableConfig(ctx context.Context) {
	tables := make([]*obj.TblTable, 0)
	_, err := orm.NewORM(consts.DBConfigName).
		Name("tbl_table").FindAll().Exec(ctx, &tables)
	if err != nil {
		panic(fmt.Errorf("initial tbl_table from db error: %s", err))
	}

	if len(tables) == 0 {
		return
	}

	for _, tbl := range tables {
		table.SetTable(tbl)
	}
}

// 初始化接入者信息
func initAppInfo(ctx context.Context) {
	//初始化执行实例信息
	appInfos := make([]*table.TblAppInfo, 0)
	accessDBs := make([]*table.TblAccessDB, 0)
	accessTables := make([]*table.TblAccessTable, 0)

	c := orm.NewORM(consts.DBConfigName)

	_, err := c.Name("tbl_app_info").FindAll().Exec(ctx, &appInfos)
	if err != nil {
		panic(fmt.Errorf("initial tbl_app_info from db error: %s", err))
	}

	if len(appInfos) == 0 {
		return
	}

	_, err = c.Name("tbl_access_db").FindAll().Exec(ctx, &accessDBs)
	if err != nil {
		panic(fmt.Errorf("initial tbl_access_db from db error: %s", err))
	}

	_, err = c.Name("tbl_access_table").FindAll().Exec(ctx, &accessTables)
	if err != nil {
		panic(fmt.Errorf("initial tbl_access_table from db error: %s", err))
	}

	for _, info := range appInfos {
		table.SetAppInfo(info)
	}

	if len(accessDBs) > 0 {
		for _, accessDB := range accessDBs {
			table.SetAccessDB(accessDB)
		}
	}

	if len(accessTables) > 0 {
		for _, accessTable := range accessTables {
			table.SetAccessTable(accessTable)
		}
	}
}

// 初始化插件
func initFilter(ctx context.Context) {
	filter := make([]*table.TblFilter, 0)
	tableFilter := make([]*table.TblTableFilter, 0)

	c := orm.NewORM(consts.DBConfigName)

	_, err := c.Name("tbl_filter").FindAll().Exec(ctx, &filter)
	if err != nil {
		panic(fmt.Errorf("init tbl_filter from db error: %s", err))
	}

	if len(filter) == 0 {
		return
	}

	_, err = c.Name("tbl_table_filter").FindAll().Exec(ctx, &tableFilter)
	if err != nil {
		panic(fmt.Errorf("init tbl_table_filter from db error: %s", err))
	}

	for _, f := range filter {
		table.SetFilter(f)
	}

	if len(tableFilter) > 0 {
		err = table.InitTableFilter(tableFilter)
		if err != nil {
			panic(fmt.Errorf("init tbl_table_filter error: %s", err))
		}
	}
}

// SyncDbNewToLocal 定时将数据库新数据更新到本地
func SyncDbNewToLocal(ctx context.Context) {
	now := time.Now()

	syncDbNewToLocal(ctx, now)
	syncFilterToLocal(ctx, now)
}

func syncDbNewToLocal(ctx context.Context, now time.Time) {
	c := orm.NewORM(consts.DBConfigName)

	//获取最新配置信息
	dbs := make([]*obj.TblDB, 0)
	tables := make([]*obj.TblTable, 0)
	appInfos := make([]*table.TblAppInfo, 0)
	accessDBs := make([]*table.TblAccessDB, 0)
	accessTables := make([]*table.TblAccessTable, 0)

	where := horm.Where{"updated_at >=": SyncTime.Format("2006-01-02 15:04:05")}

	_, _ = c.Name("tbl_db").FindAll(where).Exec(ctx, &dbs)
	_, _ = c.Name("tbl_table").FindAll(where).Exec(ctx, &tables)
	_, _ = c.Name("tbl_app_info").FindAll(where).Exec(ctx, &appInfos)
	_, _ = c.Name("tbl_access_db").FindAll(where).Exec(ctx, &accessDBs)
	_, _ = c.Name("tbl_access_table").FindAll(where).Exec(ctx, &accessTables)

	SyncTime = now

	table.UpdateDBInfo(dbs, tables, appInfos, accessDBs, accessTables)
}

func syncFilterToLocal(ctx context.Context, now time.Time) {}

// InitTable 表结构获取
func InitTable(ctx context.Context) {
	//var client = horm.NewMySQLClient(config.DBConfig, 30000000)
	//table, err := client.GenerateStructByTable(ctx, "tbl_db")
	//fmt.Println(table)
	//
	//table, err = client.GenerateStructByTable(ctx, "tbl_table")
	//fmt.Println(table)
	//
	//table, err = client.GenerateStructByTable(ctx, "tbl_app_info")
	//fmt.Println(table)
	//
	//table, err = client.GenerateStructByTable(ctx, "tbl_access_db")
	//fmt.Println(table)
	//
	//table, err = client.GenerateStructByTable(ctx, "tbl_access_table")
	//fmt.Println(table)
	//
	//table, err = client.GenerateStructByTable(ctx, "tbl_filter")
	//fmt.Println(table)
	//
	//table, err = client.GenerateStructByTable(ctx, "tbl_filter_config")
	//fmt.Println(table)
	//
	//table, err = client.GenerateStructByTable(ctx, "tbl_table_filter")
	//fmt.Println(table)
	//
	//fmt.Println(err)
}

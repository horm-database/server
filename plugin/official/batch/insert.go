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

package batch

import (
	"context"
	"fmt"
	"math/rand"
	"sort"
	"sync"
	"time"

	"github.com/horm-database/common/compress"
	"github.com/horm-database/common/consts"
	"github.com/horm-database/common/errs"
	co "github.com/horm-database/common/json"
	"github.com/horm-database/common/log"
	"github.com/horm-database/common/util"
	"github.com/horm-database/go-horm/horm"
	"github.com/horm-database/orm"
	"github.com/horm-database/orm/database/sql"
	"github.com/horm-database/orm/database/sql/client"
	"github.com/horm-database/orm/obj"
	"github.com/horm-database/server/model/table"
	"github.com/horm-database/server/srv/codec"
	"github.com/robfig/cron"
)

type InsertItem struct {
	Rand     int                      `json:"rand,omitempty"`
	Time     int64                    `json:"time,omitempty"`
	Table    string                   `json:"proto,omitempty"`
	Retry    int                      `json:"retry,omitempty"`
	Data     map[string]interface{}   `json:"data,omitempty"`
	Datas    []map[string]interface{} `json:"datas,omitempty"`
	DataType map[string]int8          `json:"data_type,omitempty"`
	Errors   map[int]string           `json:"errors,omitempty"`
}

type Cron struct {
	Switch bool          //true-开 false-关
	Cron   *cron.Cron    //定时器
	Table  *obj.TblTable //表信息
}

var (
	tableBatchCron = map[int]*Cron{}
)

// InsertHandle 批量插入处理
func InsertHandle(ctx context.Context) error {
	tblTables := table.GetTables()

	for _, tblTable := range tblTables {
		for _, tableInfo := range tblTable {
			cronMaintain(ctx, tableInfo)
		}
	}

	return nil
}

// FailedCheck 检查批量插入失败数据
func FailedCheck(ctx context.Context) error {
	tblTables := table.GetTables()
	for _, tblTable := range tblTables {
		for _, tableInfo := range tblTable {
			logFailedBatch(ctx, tableInfo)
		}
	}

	return nil
}

// HandleBatchFailed 处理批量插入失败数据
func HandleBatchFailed(ctx context.Context, name, params string, seq int32) error {
	tblTable, db, _, ok := table.GetTableAndDB(params, "")
	if !ok {
		log.Errorf(ctx, RetBatchFailedHandle, "HandleError proto %s not find", params)
		return nil
	}

	for {
		select {
		case <-ctx.Done():
			fmt.Printf("HandleBatchFailed Error context done, proto=%s, err=%v \n", params, ctx.Err())
			return ctx.Err()
		default:
		}

		datas, err := popBatchInsert(ctx, PreBatchInsertFailBuff, tblTable.Id, 1, "")
		if err != nil {
			log.Errorf(ctx, RetBatchFailedHandle,
				"HandleError popBatchInsert Error: [%v], proto=[%s]", err, tblTable.Name)
			return errs.Newf(RetBatchFailedHandle,
				"HandleError popBatchInsert Error: [%v], proto=[%s]", err, tblTable.Name)
		}

		if len(datas) == 0 {
			return nil
		}

		log.Infof(ctx, "HandleError pop %s transaction_num = %d", tblTable.Name, len(datas[0].Datas))

		c := client.NewClient(db)

		parts := map[string]bool{}
		for _, v := range datas[0].Datas {
			if date, ok := v["date"]; ok {
				if dateStr, ok := date.(string); ok {
					parts[dateStr] = true
				}
			}
		}

		log.Infof(ctx, "HandleBatchFailed_pops_parts [%s] [%s] data=%d parts=%d",
			tableInfo.Name, datas[0].Table, len(datas[0].Datas), len(parts))

		insertToDB(ctx, "failed", db, tableInfo, c, datas[0])

		time.Sleep(10 * time.Millisecond)
	}
}

// 有重试3次失败日志，等待手工处理
func logFailedBatch(ctx context.Context, tableInfo *obj.TblTable) {
	l := failedBufferLen(codec.GCtx, PreBatchInsertFailBuff, tableInfo.Id)

	if l > 0 {
		log.Errorf(ctx, RetHasBatchFailed, "logFailedBatch_failed_table %s , num=%d, key=%s",
			tableInfo.Name, l, fmt.Sprintf("%s_%d", PreBatchInsertFailBuff, tableInfo.Id))
	}
}

// Insert 批量插入
func Insert(ctx context.Context, tableInfo *obj.TblTable, subIntervals map[string]int) {
	// pop 最大数量
	popMax := InsertPopMax
	if tableInfo.BatchPopMax > 0 {
		popMax = tableInfo.BatchPopMax
	}

	waitInserts := getWaitInsertTable(tableInfo.Id)
	for tableName := range waitInserts {
		insert(ctx, popMax, tableInfo, tableName, subIntervals)
	}
}

func insert(inputCtx context.Context, popMax int, tableInfo *obj.TblTable,
	tableName string, subIntervals map[string]int) {
	ctx, cancel := context.WithTimeout(inputCtx, time.Second*600)
	defer cancel()

	if isMutex(ctx, tableInfo.Id, tableName) { //间隔指定时间才运行
		return
	}

	interval := tableInfo.BatchInterval

	if subIntervals != nil {
		if subInterval, ok := subIntervals[tableName]; ok {
			interval = subInterval
		}
	}

	setMutex(ctx, tableInfo.Id, tableName, interval)

	code := RetBatchInsert

	start := time.Now()
	datas, err := popBatchInsert(ctx, PreBatchInsertBuff, tableInfo.Id, popMax, tableName)
	during := time.Since(start)

	if err != nil && !horm.IsNil(err) {
		log.Errorf(ctx, code, "popBatchInsert Error: [%v], proto=[%s], during=[%v]", err, tableInfo.Name, during)
		return
	}

	if len(datas) == 0 {
		return
	}

	//按照插入时间排序
	sort.Sort(BatchItems(datas))

	db := table.GetTablesDB(tableInfo)

	c := client.NewClient(db)

	var sepByDate bool     //按照分区隔离
	if db.Version == 223 { //clickhouse v22.3
		sepByDate = true
	}

	// 将不同字段数和字段类型的插入记录放在不同的 database，避免插入字段错位。
	insertSeparate := categoryData(tableInfo.Id, sepByDate, datas)

	var batchNum int
	for _, batchInsertItems := range insertSeparate {
		batchNum = batchNum + len(batchInsertItems)
	}

	if len(datas) > 10000 {
		log.Warnf(ctx, "BatchInsert_pops [%s] [%s] transaction_num = %d batchNum= %d, during = %v",
			tableInfo.Name, tableName, len(datas), batchNum, during)
	} else {
		log.Infof(ctx, "BatchInsert_pops [%s] [%s] transaction_num = %d batchNum= %d, during = %v",
			tableInfo.Name, tableName, len(datas), batchNum, during)
	}

	for _, batchInsertItems := range insertSeparate {
		for _, batchInsertItem := range batchInsertItems {
			insertToDB(ctx, "batch", db, tableInfo, c, batchInsertItem)

			parts := map[string]bool{}
			for _, v := range batchInsertItem.Datas {
				if date, ok := v["date"]; ok {
					if dateStr, ok := date.(string); ok {
						parts[dateStr] = true
					}
				}
			}
			log.Infof(ctx, "BatchInsert_pops_parts [%s] [%s] data=%d parts=%d",
				tableInfo.Name, tableName, len(batchInsertItem.Datas), len(parts))
		}
	}
}

// 互斥运行，例如当 pop 间隔是 10s，则如果有实例已经 pop，则其他实例 6s 内不可以再 pop
func setMutex(ctx context.Context, tableID int, tableName string, expire int) {
	ctx, cancel := context.WithTimeout(ctx, 500*time.Millisecond)
	defer cancel()

	if expire < 1 {
		expire = 1
	}

	key := fmt.Sprintf("%s_%d_%s", PreBatchMutex, tableID, tableName)
	_ = orm.NewORM("cache").SetEX(key, 1, expire).Exec(ctx)
}

func isMutex(ctx context.Context, tableID int, tableName string) bool {
	ctx, cancel := context.WithTimeout(ctx, 500*time.Millisecond)
	defer cancel()

	var buffID int
	key := fmt.Sprintf("%s_%d_%s", PreBatchMutex, tableID, tableName)
	_ = orm.NewORM("cache").Get(key).Exec(ctx, &buffID)

	if buffID == 1 { //已经上锁
		return true
	}
	return false
}

func insertToDB(ctx context.Context, desc string, db *obj.TblDB,
	tableInfo *obj.TblTable, c *client.Client, batchInsertItem *InsertItem) {
	var err error
	batchInsertItem.Datas, err = util.FormatDatas(batchInsertItem.Datas, batchInsertItem.DataType)
	if err != nil {
		log.Errorf(ctx, ErrFormatData, "%s FormatDatas Error: [%v], proto=[%s]", desc, err, tableInfo.Name)
		return
	}

	if db.Type == consts.DBTypeClickHouse {
		retryDatas, failItems, _ := c.InsertToCK(ctx, desc, db.Version, true,
			batchInsertItem.Retry, batchInsertItem.Table, batchInsertItem.Datas)

		if len(failItems) > 0 {
			batchInsertItem.Datas = retryDatas
			batchInsertItem.Errors = failItems
			pushRetryBatchInsert(ctx, tableInfo.Id, batchInsertItem.Table, batchInsertItem)
		}

	} else {
		statement := &sql.Statement{}
		statement.SetDBType(db.Type)
		statement.SetTable(batchInsertItem.Table, "")
		statement.SetMaps(batchInsertItem.Datas)

		querySql := sql.InsertSQL(statement)
		_, _, err = c.Execute(ctx, querySql, statement.GetParams()...)
		if err != nil {
			log.Errorf(ctx, RetBatchInsert, "%s Error: [%v], proto=[%s]", desc, err, tableInfo.Name)
			pushRetryBatchInsert(ctx, tableInfo.Id, batchInsertItem.Table, batchInsertItem)
		}
	}
}

// 将不同字段数和字段类型的插入记录放在不同的 database，避免插入字段错位。
func categoryData(tableID int, sepByDate bool, datas []*InsertItem) map[string][]*InsertItem {
	insertSeparate := map[string][]*InsertItem{}
	for _, data := range datas {
		batchInsertItems, ok := insertSeparate[data.Table]
		if !ok {
			batchInsertItem := newBatchInsertItem(data)
			insertSeparate[data.Table] = []*InsertItem{batchInsertItem}
		} else {
			var find bool
			for _, batchInsertItem := range batchInsertItems {
				if sameField(tableID, sepByDate, data, batchInsertItem) {
					if len(data.Datas) > 0 {
						batchInsertItem.Datas = append(batchInsertItem.Datas, data.Datas...)
					} else if len(data.Data) > 0 {
						batchInsertItem.Datas = append(batchInsertItem.Datas, data.Data)
					}
					find = true
					break
				}
			}

			if !find {
				batchInsertItem := newBatchInsertItem(data)
				insertSeparate[data.Table] = append(insertSeparate[data.Table], batchInsertItem)
			}
		}
	}

	return insertSeparate
}

func newBatchInsertItem(data *InsertItem) *InsertItem {
	batchInsertItem := InsertItem{}
	batchInsertItem.Rand = data.Rand
	batchInsertItem.Time = data.Time
	batchInsertItem.Table = data.Table
	batchInsertItem.Retry = data.Retry
	batchInsertItem.DataType = data.DataType

	if len(data.Datas) > 0 {
		batchInsertItem.Datas = data.Datas
	} else if len(data.Data) > 0 {
		batchInsertItem.Datas = append(batchInsertItem.Datas, data.Data)
	}

	return &batchInsertItem
}

// 需要插入的字段是否全部都一样
func sameField(tableID int, sepByDate bool, data *InsertItem, newBatch *InsertItem) bool {
	dest := newBatch.Datas[0]
	source := data.Data
	if len(data.Datas) > 0 {
		source = data.Datas[0]
	}

	//不同重试次数
	if data.Retry != newBatch.Retry {
		return false
	}

	//字段不一致
	if len(dest) != len(source) {
		return false
	}

	for k := range source {
		_, ok := dest[k]
		if !ok {
			return false
		}
	}

	//日期不一致
	if sepByDate {
		sourceDate, _ := source["date"].(string)
		destDate, _ := dest["date"].(string)

		if sourceDate != destDate {
			return false
		}
	}

	//字段类型不一致
	destTypes := newBatch.DataType
	sourceTypes := data.DataType

	if len(destTypes) != len(sourceTypes) {
		return false
	}

	for k, sourceType := range sourceTypes {
		destType, ok := destTypes[k]
		if !ok || sourceType != destType {
			return false
		}
	}
	return true
}

// PushBatchInsert 批量插入缓冲区
func PushBatchInsert(ctx context.Context, nameSrv string, tableID int, table string,
	data map[string]interface{}, datas []map[string]interface{}, dataType map[string]int8) (n int, err error) {
	ctx, cancel := context.WithTimeout(ctx, 1000*time.Millisecond)
	defer cancel()

	duringLog := log.NewTimeLog(ctx)
	duringLog.SetThreshold(100) //在请求响应时长超过 ms 的时候告警

	now := time.Now().UnixNano() / 1e9
	batchItem := InsertItem{
		Rand:     rand.Intn(99999), // 集合不允许重复数据，防止重复
		Time:     now,
		Table:    table,
		Data:     data,
		Datas:    datas,
		DataType: dataType,
	}

	key := fmt.Sprintf("%s_%d_%s", PreBatchInsertBuff, tableID, table)

	buf, err := compress.JsonMarshalAndCompress(batchItem)
	if err != nil {
		duringLog.Errorf(RetBatchInsert,
			"PushBatchInsert jsonMarshalAndCompress Error: %v, batch item =[%+v]", err, batchItem)
		return 0, err
	}
	defer buf.Free()

	err = orm.NewORM(nameSrv).SAdd(key, buf.Bytes()).Exec(ctx)
	if err != nil {
		duringLog.Errorf(RetBatchInsert, "PushBatchInsert %s Error: %v, batch item =[%+v]", nameSrv, err, batchItem)
	} else {
		setWaitInsertTable(tableID, table)
	}

	return
}

// 尝试最多 3 次重试，之后写入失败集合，等待手动处理
func pushRetryBatchInsert(ctx context.Context, tableID int, tableName string, batchItem *InsertItem) {
	var key string
	var cacheRedis *orm.ORM
	if batchItem.Retry >= 2 { //批量插入失败集合，等待修复数据库之后手动重试
		key = fmt.Sprintf("%s_%d", PreBatchInsertFailBuff, tableID)
		cacheRedis = orm.NewORM("failed_set")
	} else {
		key = fmt.Sprintf("%s_%d_%s", PreBatchInsertBuff, tableID, tableName)
		cacheRedis = orm.NewORM("buffer")
		batchItem.Retry++
	}

	buf, err := compress.JsonMarshalAndCompress(batchItem)
	if err != nil {
		log.Errorf(ctx, 0, "marshal batchItem %+v  failure:%s", batchItem, err)
		return
	}
	defer buf.Free()

	duringLog := log.NewTimeLog(ctx)
	duringLog.SetThreshold(100) //在请求响应时长超过 ms 的时候告警

	err = cacheRedis.SAdd(key, buf.Bytes()).Exec(ctx)
	if err != nil {
		duringLog.Errorf(RetBatchInsert,
			"PushRetryBatchInsert Error: %v, batch item =[%+v]", err, batchItem)
	}
}

// pop 缓冲区数据
func popBatchInsert(ctx context.Context, prefix string,
	tableID, num int, tableName string) (ret []*InsertItem, err error) {
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("%v", e)
		}
	}()

	cacheRedis := orm.NewORM("buffer")

	var key string
	if tableName == "" {
		key = fmt.Sprintf("%s_%d", prefix, tableID)
	} else {
		key = fmt.Sprintf("%s_%d_%s", prefix, tableID, tableName)
	}

	var batchItems []*InsertItem

	var compressBatchItems [][]byte
	err = cacheRedis.SPop(key, num).Exec(ctx, &compressBatchItems)
	if err != nil {
		if horm.IsNil(err) {
			return nil, nil
		}
		return nil, err
	}

	for _, compressBatchItem := range compressBatchItems {
		batchItem := InsertItem{}
		err = compress.DecompressJsonUnmarshal(compressBatchItem, &batchItem)
		if err != nil || batchItem.Time == 0 {
			log.Errorf(ctx, RetBatchDataUnMarshal, "popBatchInsert Json Unmarshal Table %d "+
				"Error: %v, batchItemBytes=[%s]", tableID, err, string(compressBatchItem))
			continue
		}

		batchItems = append(batchItems, &batchItem)
	}

	return batchItems, nil
}

// BufferLen 缓冲区数据个数
func BufferLen(ctx context.Context, tableID int, tableName string) int {
	ctx, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
	defer cancel()

	cacheRedis := orm.NewORM("buffer")

	key := fmt.Sprintf("%s_%d_%s", PreBatchInsertBuff, tableID, tableName)

	var l int
	_ = cacheRedis.SCard(key).Exec(ctx, &l)
	return l
}

// 失败缓冲区长度
func failedBufferLen(ctx context.Context, prefix string, tableID int) int {
	ctx, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
	defer cancel()

	cacheRedis := orm.NewORM("failed_set")

	key := fmt.Sprintf("%s_%d", prefix, tableID)

	var l int
	_ = cacheRedis.SCard(key).Exec(ctx, &l)
	return l
}

// cronMaintain 批量插入定时器维护
func cronMaintain(ctx context.Context, tableInfo *obj.TblTable) {
	defer func() {
		if e := recover(); e != nil {
			log.Errorf(ctx, RetBatchInsert, "Batch Insert Handle Error: %v", e)
			return
		}
	}()

	batchCron, ok := tableBatchCron[tableInfo.Id]

	subInterval := map[string]int{}
	_ = co.JsonAPI.Unmarshal([]byte(tableInfo.BatchSubInterval), &subInterval)

	//关闭批量插入
	if tableInfo.BatchFlush == 0 {
		if ok && batchCron.Switch {
			batchCron.Switch = false
			batchCron.Cron.Stop() //关闭定时器

			Insert(codec.GCtx, tableInfo, subInterval) //插入最后一批数据
		}

		return
	}

	if ok {
		//定时器完全一样，不做任何变更
		if tableInfo.BatchInterval == batchCron.Table.BatchInterval &&
			tableInfo.BatchPopMax == batchCron.Table.BatchPopMax &&
			tableInfo.BatchSubInterval == batchCron.Table.BatchSubInterval {
			return
		}

		batchCron.Cron.Stop() //关闭老的定时器
	}

	newCron := &Cron{
		Cron: cron.New(),
	}

	tableBatchCron[tableInfo.Id] = newCron

	newCron.Switch = true
	newCron.Table = tableInfo

	handler := func() {
		defer func() {
			if e := recover(); e != nil {
				log.Errorf(ctx, RetBatchInsert, "Batch Insert Handle Error: %v", e)
				return
			}
		}()

		Insert(codec.GCtx, tableInfo, subInterval)
	}

	v, _ := co.JsonAPI.Marshal(subInterval)
	log.Infof(ctx, "CreateBatchInsert_Task name: %s sub_interval: %s", tableInfo.Name, string(v))

	_ = newCron.Cron.AddFunc("* * * * * *", handler)
	go newCron.Cron.Start()
}

var (
	WaitInsert       = map[int]map[string]bool{}
	WaitInsertRWLock = new(sync.RWMutex)
)

func setWaitInsertTable(tableID int, table string) {
	WaitInsertRWLock.RLock()
	_, ok := WaitInsert[tableID]
	WaitInsertRWLock.RUnlock()

	if ok {
		WaitInsertRWLock.RLock()
		ok = WaitInsert[tableID][table]
		WaitInsertRWLock.RUnlock()

		if ok {
			return
		}

		WaitInsertRWLock.Lock()
		WaitInsert[tableID][table] = true
		WaitInsertRWLock.Unlock()
	} else {
		WaitInsertRWLock.Lock()
		WaitInsert[tableID] = map[string]bool{table: true}
		WaitInsertRWLock.Unlock()
	}
}

func getWaitInsertTable(tableID int) map[string]bool {
	WaitInsertRWLock.RLock()
	defer WaitInsertRWLock.RUnlock()
	return WaitInsert[tableID]
}

type BatchItems []*InsertItem

func (s BatchItems) Len() int           { return len(s) }
func (s BatchItems) Less(i, j int) bool { return s[i].Time < s[j].Time }
func (s BatchItems) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

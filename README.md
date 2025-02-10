# 数据统一接入协议
数据统一接入协议是为了不同类型数据库的访问而设计的一套包含增删改查等等一系列操作的协议。接入层采用同一套协议， 可以操作数据统一接入服务配置的
多种存储引擎，如：mysql、clickhouse、postgres等类 sql 协议引擎，还有 redis 协议引擎，另外还支持  elastic search 引擎。<br>
在数据统一接入服务，我们可以将数据的采集、更新加工、展示、监控流向一站式的把控。

#  示例表
建表语句：
```sql
CREATE TABLE `student` (
    `id` bigint unsigned NOT NULL AUTO_INCREMENT,
    `identify` bigint NOT NULL COMMENT '学生编号',
    `gender` tinyint NOT NULL DEFAULT '1' COMMENT '1-male 2-female',
    `age` int unsigned NOT NULL DEFAULT '0' COMMENT '年龄',
    `name` varchar(64) NOT NULL COMMENT '名称',
    `score` double DEFAULT NULL COMMENT '分数',
    `image` blob COMMENT 'image',
    `article` text COMMENT 'publish article',
    `exam_time` time DEFAULT NULL COMMENT '考试时间',
    `birthday` date DEFAULT NULL COMMENT '出生日期',
    `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '修改时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `identity` (`identify`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='学生表'

CREATE TABLE `student_course` (
    `id` int NOT NULL AUTO_INCREMENT,
    `identify` bigint NOT NULL COMMENT '学生编号',
    `course` varchar(64) NOT NULL COMMENT '课程',
    `hours` int DEFAULT '0' COMMENT '课时数',
    PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='学生课程表';

CREATE TABLE `course_info` (
    `course` varchar(64) NOT NULL COMMENT '课程',
    `teacher` varchar(64) NOT NULL COMMENT '课程老师',
    `time` time NOT NULL COMMENT '上课时间',
    PRIMARY KEY (`course`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='课程信息';

CREATE TABLE `teacher_info` (
    `teacher` varchar(32) NOT NULL COMMENT '老师',
    `age` int NOT NULL DEFAULT '0' COMMENT '年龄',
    PRIMARY KEY (`teacher`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='老师信息';

CREATE TABLE `score_rank_reward` (
    `rank` int NOT NULL COMMENT '排名',
    `reward` varchar(128) NOT NULL DEFAULT '' COMMENT '奖励',
    PRIMARY KEY (`rank`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='分数排名奖励'
```

# 查询单元（执行单元）
## 数据名称
协议由一组查询单元（执行单元）组成，每个执行单元都是对一张表或者一个es 索引的一个操作，包含增删改查等操作，查询单元（执行单元）
在 数据统一接入服务通过 `name(数据名称)` 找到对应的mysql表/es索引/redis配置信息、及其数据库信息，然后根据协议将执行单元转化
为对应数据库 sql语句、 elastic 请求或 redis 请求，并将执行结果返回到客户端。

请求：
```json
{
  "name": "student",
  "op": "find",
  "where": {
    "name": "caohao"
  }
}
```
如果存在相同的数据名称的时候，我们可以通过增加库名来区分如下，否则会报错，不允许存在相同的库名+数据名。
请求：
```json
{
  "name": "test::student",
  "op": "find",
  "where": {
    "name": "caohao"
  }
}
```

返回结果：
```json
{
  "created_at": "2024-11-30T20:53:57+08:00",
  "id": 234047220842770433,
  "identify": 2024080313,
  "age": 23,
  "score": 91.5,
  "image": "SU1BR0UuUENH",
  "exam_time": "15:30:00",
  "birthday": "1987-08-27T00:00:00+08:00",
  "gender": 1,
  "name": "caohao",
  "article": "groundbreaking work in cryptography and complexity theory",
  "updated_at": "2024-12-12T19:30:37+08:00"
}
```

## 查询单元结构体
一个完整的执行单元包含如下信息：
```go
// github.com/horm-database/common/proto
package proto

import (
	"github.com/horm-database/common/proto/sql"
	"github.com/horm-database/common/types"
)

// Unit 查询单元（执行单元）
type Unit struct {
	// query base info
	Name  string   `json:"name,omitempty"`  // name
	Op    string   `json:"op,omitempty"`    // operation
	Shard []string `json:"shard,omitempty"` // 分片、分表、分库

	// 结构化查询共有
	Column []string               `json:"column,omitempty"` // columns
	Where  map[string]interface{} `json:"where,omitempty"`  // query condition
	Order  []string               `json:"order,omitempty"`  // order by
	Page   int                    `json:"page,omitempty"`   // request pages. when page > 0, the request is returned in pagination.
	Size   int                    `json:"size,omitempty"`   // size per page
	From   uint64                 `json:"from,omitempty"`   // offset

	// data maintain
	Val      interface{}              `json:"val,omitempty"`       // 单条记录 val (not map/[]map)
	Data     map[string]interface{}   `json:"data,omitempty"`      // add/update one map data
	Datas    []map[string]interface{} `json:"datas,omitempty"`     // batch add/update map data
	Args     []interface{}            `json:"args,omitempty"`      // multiple args, 还可用于 query 语句的参数，或者 redis 协议，如 MGET、HMGET、HDEL 等
	DataType map[string]types.Type    `json:"data_type,omitempty"` // 数据类型（主要用于 clickhouse，对于数据类型有强依赖），请求 json 不区分 int8、int16、int32、int64 等，只有 Number 类型，bytes 也会被当成 string 处理。

	// group by
	Group  []string               `json:"group,omitempty"`  // group by
	Having map[string]interface{} `json:"having,omitempty"` // group by condition

	// for databases such as mysql ...
	Join []*sql.Join `json:"join,omitempty"`

	// for databases such as elastic ...
	Type   string  `json:"type,omitempty"`   // type, such as elastic`s type, it can be customized before v7, and unified as _doc after v7
	Scroll *Scroll `json:"scroll,omitempty"` // scroll info

	// for databases such as redis ...
	Prefix string   `json:"prefix,omitempty"` // prefix, It is strongly recommended to bring it to facilitate finer-grained summary statistics, otherwise the statistical granularity can only be cmd ，such as GET、SET、HGET ...
	Key    string   `json:"key,omitempty"`    // key
	Keys   []string `json:"keys,omitempty"`   // keys

	// bytes 字节流
	Bytes []byte `json:"bytes,omitempty"`

	// params 与数据库特性相关的附加参数，例如 redis 的 WITHSCORES，以及 elastic 的 refresh、collapse、runtime_mappings、track_total_hits 等等。
	Params map[string]interface{} `json:"params,omitempty"`

	// 直接送 Query 语句，需要拥有库的 表权限、或 root 权限。具体参数为 args
	Query string        `json:"query,omitempty"`

	// Extend 扩展信息，作用于插件
	Extend map[string]interface{} `json:"extend,omitempty"`

	Sub   []*Unit `json:"sub,omitempty"`   // 子查询
	Trans []*Unit `json:"trans,omitempty"` // 事务，该事务下的所有 Unit 必须同时成功或失败（注意：仅适合支持事务的数据库回滚，如果数据库不支持事务，则操作不会回滚）
}

// Scroll 滚动查询
type Scroll struct {
	ID   string `json:"id,omitempty"`   // 滚动 id
	Info string `json:"info,omitempty"` // 滚动查询信息，如时间
}

type Join struct {
	Type  string            `json:"type,omitempty"`
	Table string            `json:"table,omitempty"`
	Using []string          `json:"using,omitempty"`
	On    map[string]string `json:"on,omitempty"`
}
```

## 基础数据类型
执行单元中的 data、datas、args 等数据参数，可以包含如下一些基础数据类型下，在一般情况下，我们是不需要指定数据的类型的：
```go
package structs

type Type int8

const (
	TypeTime   Type = 1 // 类型是 time.Time
	TypeBytes  Type = 2 // 类型是 []byte
	TypeFloat  Type = 3
	TypeDouble Type = 4
	TypeInt    Type = 5
	TypeUint   Type = 6
	TypeInt8   Type = 7
	TypeInt16  Type = 8
	TypeInt32  Type = 9
	TypeInt64  Type = 10
	TypeUint8  Type = 11
	TypeUint16 Type = 12
	TypeUint32 Type = 13
	TypeUint64 Type = 14
	TypeString Type = 15
	TypeBool   Type = 16
	TypeJSON   Type = 17
)

```

我们发送请求到数据统一调度服务的时候，绝大多数情况下可以不指定数据类型，服务端也可以正常解析并执行 query 语句，但是在某些特殊情况下，
比如 clickhouse 对类型有强限制，又或者字段是一个超大 uint64 整数，json 编码之后请求服务端，由于 json 的基础类型只包含 string、 
number(当成float64)、bool，数字在服务端会被解析为 float64，存在精度丢失问题，一般当类型为 time、[]byte、int、int8~int64、uint、
uint8~uint64 时，需要在执行单元 data_type 字段里将数据类型带上，比如下面对clickhouse的插入：

```json
{
  "name": "student(add)",
  "op": "insert",
  "data": {
    "identify": 2024080313,
    "name": "caohao",
    "score": 91.5,
    "created_at": "2025-01-05T20:14:50.702248+08:00",
    "exam_time": "15:30:00",
    "birthday": "1987-08-27",
    "updated_at": "2025-01-05T20:14:50.702249+08:00",
    "article": "groundbreaking work in cryptography and complexity theory",
    "id": 234047220842770433,
    "image": "SU1BR0UuUENH",
    "gender": 1,
    "age": 23
  },
  "data_type": {
    "id": 14,
    "image": 2,
    "created_at": 1,
    "identify": 10,
    "gender": 7,
    "age": 6,
    "updated_at": 1
  }
}
```
horm 基础类型，会在数据统一接入服务根据指定的数据源引擎映射、解析成对应的类型，例如在 mysql 和 clickhouse 类型映射为：
```go
//github.com/orm/database/sql/type.go

var MySQLTypeMap = map[string]types.Type{
  "INT":                types.TypeInt,
  "TINYINT":            types.TypeInt8,
  "SMALLINT":           types.TypeInt16,
  "MEDIUMINT":          types.TypeInt32,
  "BIGINT":             types.TypeInt64,
  "UNSIGNED INT":       types.TypeUint,
  "UNSIGNED TINYINT":   types.TypeUint8,
  "UNSIGNED SMALLINT":  types.TypeUint16,
  "UNSIGNED MEDIUMINT": types.TypeUint32,
  "UNSIGNED BIGINT":    types.TypeUint64,
  "BIT":                types.TypeBytes,
  "FLOAT":              types.TypeFloat,
  "DOUBLE":             types.TypeDouble,
  "DECIMAL":            types.TypeDouble,
  "VARCHAR":            types.TypeString,
  "CHAR":               types.TypeString,
  "TEXT":               types.TypeString,
  "BLOB":               types.TypeBytes,
  "BINARY":             types.TypeBytes,
  "VARBINARY":          types.TypeBytes,
  "TIME":               types.TypeString,
  "DATE":               types.TypeTime,
  "DATETIME":           types.TypeTime,
  "TIMESTAMP":          types.TypeTime,
  "JSON":               types.TypeJSON,
}

var ClickHouseTypeMap = map[string]types.Type{
  "Int":         types.TypeInt,
  "Int8":        types.TypeInt8,
  "Int16":       types.TypeInt16,
  "Int32":       types.TypeInt32,
  "Int64":       types.TypeInt64,
  "UInt":        types.TypeUint,
  "UInt8":       types.TypeUint8,
  "UInt16":      types.TypeUint16,
  "UInt32":      types.TypeUint32,
  "UInt64":      types.TypeUint64,
  "Float":       types.TypeFloat,
  "Float32":     types.TypeFloat,
  "Float64":     types.TypeDouble,
  "Decimal":     types.TypeDouble,
  "String":      types.TypeString,
  "FixedString": types.TypeString,
  "UUID":        types.TypeString,
  "DateTime":    types.TypeTime,
  "DateTime64":  types.TypeTime,
  "Date":        types.TypeTime,
}
```

## 别名
如果我们用到 mysql 的别名，或者在并发查询、复合查询模式下、同一层级的多个查询单元如果访问同一张表，为了结果的正常，我们必须在括号里加上别名，
如下代码的`student(add)` 和 `student(find)` ，我们都是访问 student。
```json
[
  {
    "name": "student(add)",
    "op": "insert",
    "data": {
      "exam_time": "15:30:00",
      "identify": 2024092316,
      "name": "jerry",
      "score": 82.5,
      "created_at": "2025-01-05T20:30:21.977161+08:00",
      "gender": 1,
      "article": "contributions to deep learning in artificial intelligence",
      "age": 17,
      "updated_at": "2025-01-05T20:30:21.977162+08:00",
      "id": 234049805125431297,
      "image": "SU1BR0UuUENH",
      "birthday": "1995-03-24"
    }
  },
  {
    "name": "student(find)",
    "op": "find",
    "where": {
      "@id": "add.id"
    }
  }
]
```

以下是上面请求的返回结果，是一个 map，其中 map 的 key 就是执行单元的名称或别名，如果都用 student，则无法区分是返回
是哪个执行单元的结果，而且会丢失一个执行单元的结果，这时候需要用别名来区别。
```json
{
  "add": {
    "id": "234049805125431297",
    "rows_affected": 1
  },
  "find": {
    "name": "jerry",
    "image": "SU1BR0UuUENH",
    "article": "contributions to deep learning in artificial intelligence",
    "updated_at": "2025-01-05T20:30:22+08:00",
    "birthday": "1995-03-24T00:00:00+08:00",
    "created_at": "2025-01-05T20:30:22+08:00",
    "id": 234049805125431297,
    "identify": 2024092316,
    "gender": 1,
    "age": 17,
    "score": 82.5,
    "exam_time": "15:30:00"
  }
}
```

另外一种情况就是作为 mysql 的别名存在：
```json
{
  "name": "student_course(sc)",
  "op": "find_all",
  "column": [
    "sc.*",
    "s.name"
  ],
  "size": 100,
  "join": [
    {
      "type": "LEFT",
      "table": "student(s)",
      "on": {
        "identify": "identify"
      }
    }
  ]
}
```

上面的语句生成的对应 sql 语句如下：
```sql
SELECT  `sc`.* , `s`.`name`  FROM `student_course` AS `sc` 
	LEFT JOIN `student` AS `s` ON `sc`.`identify`=`s`.`identify`
```

## 分片、分表、分库
默认情况下，我们的表名就等于数据名，但是，如果有mysql分表、elastic分索引、redis分库的情况，我们需要用到 shard 功能来指定分表，
如下案例我们 student 表，根据 identify % 100 分了100张分表。
```json
[
  {
    "name": "student",
    "op": "find",
    "shard": [
      "student_33"
    ],
    "where": {
      "identify": 2024070733
    }
  }
]
```

在统一接入服务，我们会校验 shard 表是否符合该数据的表校验规则，表校验规则支持单一表名、逗号分隔的多个表名、正则表达式 regex/student_*?/、
还有就是比较常用 `...` 校验， 例如咱们例子中的student_0...99 表示 从 student_0 一直到 student_99。


# 查询模式
数据统一接入协议一共包含3种查询模式，单查询单元，并行执行，嵌套查询。
## 单查询单元
整个查询仅包含一个执行单元。
### 单结果返回
执行单条语句，`isNil`, `error` 直接通过 Exec 函数返回，当查询结果为空时，isNil=true。

查询单条记录：
请求：
```json
{
  "name": "student",
  "op": "find",
  "where": {
    "name": "caohao"
  }
}
```

返回结果：
```json
{
  "created_at": "2024-11-30T20:53:57+08:00",
  "id": 234047220842770433,
  "identify": 2024080313,
  "age": 23,
  "score": 91.5,
  "image": "SU1BR0UuUENH",
  "exam_time": "15:30:00",
  "birthday": "1995-03-23T00:00:00+08:00",
  "gender": 1,
  "name": "caohao",
  "article": "groundbreaking work in cryptography and complexity theory",
  "updated_at": "2024-12-12T19:30:37+08:00"
}
```

查询多条记录
请求：
```json
{
    "name": "student",
    "op": "find_all",
    "where": {
        "name ~": "%cao%"
    },
    "size": 100
}
```

返回结果：
```json
[
  {
    "image": "SU1BR0UuUENH",
    "created_at": "2024-11-30T20:53:57+08:00",
    "updated_at": "2024-12-12T19:30:37+08:00",
    "age": 23,
    "name": "caohao",
    "score": 91.5,
    "article": "groundbreaking work in cryptography and complexity theory",
    "exam_time": "15:30:00",
    "birthday": "1987-08-27T00:00:00+08:00",
    "id": 234047220842770433,
    "identify": 2024080313,
    "gender": 1
  },
  {
    "image": "cDFJHVDwerC",
    "created_at": "2024-11-30T20:53:57+08:00",
    "updated_at": "2024-12-12T19:30:37+08:00",
    "age": 23,
    "name": "hongcao",
    "score": 91.3,
    "article": "",
    "exam_time": "14:30:00",
    "birthday": "1995-03-23T00:00:00+08:00",
    "id": 1,
    "identify": 2024062461,
    "gender": 1
  }
]
```

### 多结果返回
有时候，可能会返回多个结果，例如 redis 的 ZRangeByScore：
请求：
```json
{
  "name": "redis_student",
  "op": "zrangebyscore",
  "key": "student_age_rank",
  "args": [10, 50],
  "params": {
    "with_scores": true
  }
}
```

返回如下，member 和 score 在两个数据内返回，而且 member 是有序集成员，数组下标相同的 score 为成员的分数。：
```json
{
    "member": [
        "{\"age\":23,\"image\":\"SU1BR0UuUENH\",\"article\":\"contributions to deep learning in artificial intelligence\",\"id\":227518753250750465,\"score\":91.5,\"birthday\":\"1987-08-27T00:00:00Z\",\"updated_at\":\"2024-12-18T19:58:17.869141+08:00\",\"identify\":2024080313,\"exam_time\":\"15:30:00\",\"gender\":2,\"name\":\"jerry\",\"created_at\":\"2024-12-18T19:58:17.869147+08:00\"}",
        "{\"image\":\"SU1BR0UuUENH\",\"birthday\":\"0001-01-01T00:00:00Z\",\"name\":\"jerry\",\"article\":\"contributions to deep learning in artificial intelligence\",\"updated_at\":\"2024-12-18T19:40:41.184551+08:00\",\"gender\":2,\"score\":91.5,\"exam_time\":\"15:30:00\",\"id\":227514321192628225,\"created_at\":\"2024-12-18T19:40:41.184549+08:00\",\"identify\":2024080313,\"age\":23}",
        "{\"score\":91.5,\"birthday\":\"0001-01-01T00:00:00Z\",\"name\":\"jerry\",\"article\":\"contributions to deep learning in artificial intelligence\",\"exam_time\":\"15:30:00\",\"updated_at\":\"2024-12-17T20:49:17.568859+08:00\",\"id\":227169198692904961,\"age\":23,\"created_at\":\"2024-12-17T20:49:17.568853+08:00\",\"gender\":2,\"identify\":2024080313,\"image\":\"SU1BR0UuUENH\"}"
    ],
    "score": [
        23,
        23,
        23
    ]
}
```

### 分页返回
当我们请求参数 page > 1 时，返回结果会以分页形式返回：

请求：
```json
{
    "name": "student",
    "op": "find_all",
    "page": 1,
    "size": 10
}
```

返回结果，我们将分页信息放在 detail 中，数据结果放在 data 中：
```json
{
  "detail": {
    "total": 2,
    "total_page": 1,
    "page": 1,
    "size": 10
  },
  "data": [
    {
      "score": 91.5,
      "image": "SU1BR0UuUENH",
      "created_at": "2025-01-05T20:20:06+08:00",
      "gender": 1,
      "age": 23,
      "name": "caohao",
      "article": "groundbreaking work in cryptography and complexity theory",
      "exam_time": "15:30:00",
      "birthday": "1987-08-27T00:00:00+09:00",
      "updated_at": "2025-01-05T20:20:06+08:00",
      "id": 234047220842770433,
      "identify": 2024080313
    },
    {
      "age": 17,
      "name": "jerry",
      "image": "SU1BR0UuUENH",
      "article": "contributions to deep learning in artificial intelligence",
      "birthday": "1995-03-24T00:00:00+08:00",
      "updated_at": "2025-01-05T20:30:22+08:00",
      "id": 234049805125431297,
      "identify": 2024092316,
      "exam_time": "15:30:00",
      "created_at": "2025-01-05T20:30:22+08:00",
      "gender": 1,
      "score": 82.5
    }
  ]
}
```

实际上统一接入服务返回的分页数据结构如下：

```go
// PageResult 当 page > 1 时会返回分页结果
type PageResult struct {
	Detail *Detail       `orm:"detail,omitempty" json:"detail,omitempty"` // 查询细节信息
	Data   []interface{} `orm:"data,omitempty" json:"data,omitempty"`     // 分页结果
}

// Detail 其他查询细节信息，例如 分页信息、滚动翻页信息、其他信息等。
type Detail struct {
	Total     uint64                 `orm:"total" json:"total"`                               // 总数
	TotalPage uint32                 `orm:"total_page,omitempty" json:"total_page,omitempty"` // 总页数
	Page      int                    `orm:"page,omitempty" json:"page,omitempty"`             // 当前分页
	Size      int                    `orm:"size,omitempty" json:"size,omitempty"`             // 每页大小
	Scroll    *Scroll                `orm:"scroll,omitempty" json:"scroll,omitempty"`         // 滚动翻页信息
	Extras    map[string]interface{} `orm:"extras,omitempty" json:"extras,omitempty"`         // 更多详细信息
}
```

## 并行查询
### 并发同时执行
为了高效并发，我们可以将多个语句组织在一起，一同发送到到数据统一接入服务，由数据统一接入服务并发执行，并返回结果。

`注意：如果并行执行访问同一个数据时，为了区别，可以像下面一样在括号里面加别名：redis_student(zadd) 和 redis_student(range)。`<br><br>

`另外我们注意看返回结果，zrangebyscore 仅返回了2条数据，实际上应该有3条数据，也就是 zadd 的数据并未出现在 zrangebyscore 结果中， 这是
因为在并发执行过程中，两个语句是同时执行，我们并不知道哪个语句先执行完，如果 zrangebyscore 先于 zadd 执行完成，就会导致数据还未插入完成就
获取了排序结果，这显然与我们的预期不符，所以当遇到两条执行语句有先后要求时，我们最好拆成两条独立的语句先后执行，而不是放在一个并发执行中。`

```json
[
  {
    "name": "redis_student(zadd)",
    "op": "zadd",
    "key": "student_age_rank",
    "args": [
      17,
      "{\"id\":234051825504890881,\"exam_time\":\"15:30:00\",\"name\":\"jerry\",\"image\":\"SU1BR0UuUENH\",\"updated_at\":\"2025-01-05T20:38:23.673443+08:00\",\"article\":\"contributions to deep learning in artificial intelligence\",\"gender\":1,\"age\":17,\"birthday\":\"1987-08-27\",\"created_at\":\"2025-01-05T20:38:23.67346+08:00\",\"identify\":2024092316,\"score\":82.5}"
    ]
  },
  {
    "name": "redis_student(range)",
    "op": "zrangebyscore",
    "key": "student_age_rank",
    "args": [
      10,
      50
    ],
    "params": {
      "with_scores": true
    }
  }
]
```

### 引用
引用是指的一个查询单元的请求参数来自另外一个查询的返回结果，当出现引用的时候，并行执行会退化为串行执行。引用有多种方式，
如下 {"@identify": "student.identify"} 中 map 的 key 以`@`开头的时候，表示 identify 的值引用自 student 执行单元
的返回结果的 identify 字段。`.` 之前表示引用路径，之后表示引用的 field， 被引用的执行单元必须在引用的执行单元之前被执行，否则就会报错。

请求：
```json
[
    {
        "name": "student",
        "op": "find_all",
        "page": 1,
        "size": 10
    },
    {
        "name": "student_course",
        "op": "find_all",
        "where": {
            "@identify": "student.identify"
        }
    }
]
```

返回：
```json
{
  "student": {
    "detail": {
      "total": 2,
      "total_page": 1,
      "page": 1,
      "size": 10
    },
    "data": [
      {
        "score": 91.5,
        "exam_time": "15:30:00",
        "birthday": "1987-08-27T00:00:00+09:00",
        "updated_at": "2025-01-05T20:20:06+08:00",
        "id": 234047220842770433,
        "name": "caohao",
        "age": 23,
        "image": "SU1BR0UuUENH",
        "article": "groundbreaking work in cryptography and complexity theory",
        "created_at": "2025-01-05T20:20:06+08:00",
        "identify": 2024080313,
        "gender": 1
      },
      {
        "name": "jerry",
        "score": 82.5,
        "updated_at": "2025-01-05T20:30:22+08:00",
        "article": "contributions to deep learning in artificial intelligence",
        "exam_time": "15:30:00",
        "birthday": "1995-03-24T00:00:00+08:00",
        "id": 234049805125431297,
        "identify": 2024092316,
        "gender": 1,
        "age": 17,
        "image": "SU1BR0UuUENH",
        "created_at": "2025-01-05T20:30:22+08:00"
      }
    ]
  },
  "student_course": [
    {
      "course": "Math",
      "hours": 54,
      "id": 1,
      "identify": 2024080313
    },
    {
      "id": 2,
      "identify": 2024080313,
      "course": "Physics",
      "hours": 32
    },
    {
      "id": 3,
      "identify": 2024092316,
      "course": "English",
      "hours": 68
    }
  ]
}
```

当引用参数是 key (string) 或者 args ([]interface{}) 而不是 where (map[string]interface{}) 的时候，
需要 `@{}` 方式，例如 @{student.identify} 来表示该参数来自于引用 student.identify。 例如下面这个例子，
我们需要先查询 name="caohao" 的学生，然后根据学生的 identify 来获取他的排名：

请求：
```json
[
    {
        "name": "student",
        "op": "find",
        "where": {
            "name": "caohao"
        }
    },
    {
        "name": "redis_student(score_rank)",
        "op": "zrank",
        "key": "student_score_rank",
        "args": ["@{student.identify}"]
    }
]
```

返回：
```json
{
    "student": {
        "article": "groundbreaking work in cryptography and complexity theory",
        "identify": 2024080313,
        "score": 91.5,
        "image": "SU1BR0UuUENH",
        "name": "caohao",
        "exam_time": "15:30:00",
        "birthday": "1987-08-27T00:00:00+08:00",
        "created_at": "2024-11-30T20:53:57+08:00",
        "updated_at": "2024-12-12T19:30:37+08:00",
        "id": 234047220842770433,
        "gender": 1,
        "age": 23
    },
    "score_rank": 1
}
```

当被引用的值不是一个 map，而是一个具体数值的时候，我们不需要 `.` 来指定 field，而是直接采用被引用的执行单元即可。 例如下面我们获取了
一个学生的排名， 我们期望在一个并行执行单元中知道该排名的奖励：
```json
[
    {
        "name": "redis_student(score_rank)",
        "op": "zrank",
        "key": "student_score_rank",
        "args": [2024080313]
    },
    {
        "name": "score_rank_reward",
        "op": "find",
        "where": {
            "@rank": "score_rank"
        }
    }
]
```

## 复合查询
### 返回结构
复合执行包含并行执行加上子查询，在复合查询的结果，如果返回的是一个数组，我们会为每个数组结果都执行一遍该查询的子查询，每个复合查询的结果
都包含 error、is_nil、detail 和 data 4个参数，当 error 不存在或者等于 nil 的时候，则结果正常无报错，分页等详情再 detail 中，
如果返回数据为空则 is_nil=true，当 is_nil 不存在，或者等于 false 时，返回数据存在于 data 中。子查询也在父查询的返回 data 中。

```go
package proto // "github.com/horm-database/common/proto"

// CompResult 混合查询返回结果
type CompResult struct {
	RetBase             // 返回基础信息
	Data    interface{} `json:"data"` // 返回数据
}

// RetBase 混合查询返回结果基础信息
type RetBase struct {
	Error  *Error  `json:"error,omitempty"`  // 错误返回
	IsNil  bool    `json:"is_nil,omitempty"` // 是否为空
	Detail *Detail `json:"detail,omitempty"` // 查询细节信息
}
```

请求：
```json
[
    {
        "name": "student",
        "op": "find_all",
        "page": 1,
        "size": 10,
        "sub": [
            {
                "name": "student_course",
                "op": "find_all",
                "where": {
                    "@identify": "/student.identify"
                },
                "size": 100,
                "sub": [
                    {
                        "name": "course_info",
                        "op": "find",
                        "where": {
                            "@course": "../.course"
                        }
                    }
                ]
            },
            {
                "name": "teacher_info",
                "op": "find_all",
                "where": {
                    "@teacher": "student_course/course_info.teacher"
                },
                "size": 100,
                "sub": [
                    {
                        "name": "redis_student(test_nil)",
                        "op": "get",
                        "key": "not_exists"
                    }
                ]
            }
        ]
    },
    {
        "name": "teacher_info(test_error)",
        "op": "find",
        "where": {
            "not_exist_field": 55
        }
    }
]
```

返回：
```json
{
  "student": {
    "detail": {
      "total": 2,
      "total_page": 1,
      "page": 1,
      "size": 10
    },
    "data": [
      {
        "id": 234047220842770433,
        "identify": 2024080313,
        "gender": 1,
        "age": 23,
        "name": "caohao",
        "score": 91.5,
        "image": "SU1BR0UuUENH",
        "article": "groundbreaking work in cryptography and complexity theory",
        "exam_time": "15:30:00",
        "birthday": "1987-08-27T00:00:00+09:00",
        "created_at": "2025-01-05T20:20:06+08:00",
        "updated_at": "2025-01-05T20:20:06+08:00",
        "student_course": {
          "data": [
            {
              "id": 1,
              "identify": 2024080313,
              "course": "Math",
              "hours": 54,
              "course_info": {
                "data": {
                  "course": "Math",
                  "teacher": "Simon",
                  "time": "11:00:00"
                }
              }
            },
            {
              "id": 2,
              "identify": 2024080313,
              "course": "Physics",
              "hours": 32,
              "course_info": {
                "data": {
                  "course": "Physics",
                  "teacher": "Richard",
                  "time": "14:00:00"
                }
              }
            }
          ]
        },
        "teacher_info": {
          "data": [
            {
              "teacher": "Richard",
              "age": 57,
              "test_nil": {
                "is_nil": true
              }
            },
            {
              "teacher": "Simon",
              "age": 61,
              "test_nil": {
                "is_nil": true
              }
            }
          ]
        }
      },
      {
        "id": 234049805125431297,
        "identify": 2024092316,
        "gender": 1,
        "age": 17,
        "name": "jerry",
        "score": 82.5,
        "image": "SU1BR0UuUENH",
        "article": "contributions to deep learning in artificial intelligence",
        "exam_time": "15:30:00",
        "birthday": "1995-03-24T00:00:00+08:00",
        "created_at": "2025-01-05T20:30:22+08:00",
        "updated_at": "2025-01-05T20:30:22+08:00",
        "student_course": {
          "data": [
            {
              "id": 3,
              "identify": 2024092316,
              "course": "English",
              "hours": 68,
              "course_info": {
                "data": {
                  "course": "English",
                  "teacher": "Dennis",
                  "time": "15:30:00"
                }
              }
            }
          ]
        },
        "teacher_info": {
          "data": [
            {
              "teacher": "Dennis",
              "age": 39,
              "test_nil": {
                "is_nil": true
              }
            }
          ]
        }
      }
    ]
  },
  "test_error": {
    "error": {
      "type": 2,
      "code": 1054,
      "msg": "mysql query error: [Unknown column 'not_exist_field' in 'where clause']"
    }
  }
}
```
### 引用路径
不同于并行查询的所有查询单元都在同一个层级，在复合查询中，有了子查询，在不同层级的情况下，引用会变得复杂，我们可以采用相对路径和绝对路径，
来指向我们需要被引用的查询单元。 如果 `/` 开头，则表是该路径属于绝对路径，例如上面实例中的 `/student.identify`，否则，就是相对路径，
相对路径在计算的时候，会把当前层级所在的父查询的绝对路径加在相对路径前，例如上面案例的 `student_course/course_info.teacher` ，
会变成 `/student/student_course/course_info.teacher`如果以 `../` 开头的相对路径，则会把`../` 转化为父查询的绝对路径，
例如上面案例的 `../.course`，会变成 `/student/student_course.course`，在相对路径转化为绝对路径之后，再根据规则获取指定路径的引用结果。

## 返回结果
### 空返回 和 error
当数据源为 mysql、clickhouse、es 等数据库时，如果 find 或者 find_all 查询的数据为空时，返回参数 isNil=true，否则，返回参数为 false，
而当数据源为 redis 时，只有 redis 返回 redigo: nil returned 错误时，才会使得 isNil = true，其他时候都是 isNil = false，
即便如下 ZRangeByScore 去查询一个不存在的有序集时，isNil 也是 false。
```json
[
    {
        "name": "student",
        "op": "find",
        "where": {
            "name": "noexist"
        }
    }
]
```
```go
// is_nil = true
```
```json
[
    {
        "name": "student",
        "op": "find_all",
        "where": {
            "name": "noexists"
        }
    }
]
```
```go
// is_nil = true
```
```json
[
    {
        "name": "redis_student",
        "op": "get",
        "key": "noexists"
    }
]
```
```go
// is_nil = true
```
```json
[
    {
        "name": "redis_student",
        "op": "zrangebyscore",
        "data_type": {
            "0": 5,
            "1": 5
        },
        "key": "noexists",
        "args": [
            70,
            100
        ],
        "params": {
            "with_scores": true
        }
    }
]
```
```go
// is_nil = false
```


上面展示的是单执行单元的返回结果，在单执行单元中，is_nil、error 参数在 ResponseHeader 中返回客户端：
```protobuf
/* ResponseHeader 响应头 */
message ResponseHeader {
  ...
  Error err = 5;                     // 返回错误
  bool is_nil = 6;                   // 返回是否为空（针对单执行单元）
}
```

在并行查询中，一般系统返回，例如请求参数错误、解析失败、网络错误、权限错误等都会在 ResponseHeader 的 err 返回。
每个并行查询单元的 is_nil、error 结果则会在 ResponseHeader 中的 rsp_nils、rsp_errs 中返回给客户端，
这是一个 map，key是请求名(别名)。
```protobuf
/* ResponseHeader 响应头 */
message ResponseHeader {
  ...
  Error err = 5;                     // 返回错误
  map<string, Error> rsp_errs = 7;   // 错误返回（针对多执行单元并发）
  map<string, bool> rsp_nils = 8;    // 是否为空返回（针对多执行单元并发）
}
```
示例：
```json
[
    {
        "name": "student(add)",
        "op": "insert",
        "data": {
            "no_field": null
        }
    },
    {
        "name": "student(find)",
        "op": "find",
        "where": {
            "no_field": "caohao"
        }
    }
]
```

在复合查询中，请求参数错误、解析失败、网络错误、权限错误等依然在 ResponseHeader 的 err 中返回，
每个查询单元的 is_nil、error 则包含在结果里面。
```go
package proto // "github.com/horm-database/common/proto"

// CompResult 混合查询返回结果
type CompResult struct {
	RetBase             // 返回基础信息
	Data    interface{} `json:"data"` // 返回数据
}

// RetBase 混合查询返回结果基础信息
type RetBase struct {
	Error  *Error  `json:"error,omitempty"`  // 错误返回
	IsNil  bool    `json:"is_nil,omitempty"` // 是否为空
	Detail *Detail `json:"detail,omitempty"` // 查询细节信息
}
```

数据统一接入服务的错误结构如下，错误包含：错误类型，错误码，错误信息，异常查询语句组成（sql不仅指代sql语句，elastic语句、redis 命令也包含在内）
```protobuf
message Error {
  int32  type = 1; //错误类型
  int32  code = 2; //错误码
  string msg = 3;  //错误信息
  string sql = 4;  //异常sql语句
}
```

错误类型包含3大类，比如请求参数错误、解析失败、网络错误、权限错误等都属于系统错误，找不到插件、插件未注册、插件执行错误等都属于插件错误。
数据库执行报错都属于数据库错误。
```go
// EType 错误类型
type EType int8

const (
	ETypeSystem   EType = 0 //系统错误
	ETypePlugin   EType = 1 //插件错误
	ETypeDatabase EType = 2 //数据库错误
)
```

### 全部成功
当 Elastic 批量插入新数据时，返回 `[]*proto.ModRet`，我们可以遍历返回结果，`status` 为错误码，当 `status!=0` 则该条记录
插入失败，`reason`为失败原因，这样，我们可以针对失败的记录做特殊处理，比如重试。
```json
[
  {
    "name": "es_student",
    "op": "insert",
    "datas": [
      {
        "image": "SU1BR0UuUENH",
        "gender": 1,
        "age": 67,
        "name": "wigderson",
        "exam_time": "14:30:00",
        "id": 1,
        "article": "enhanced human understanding of the role of randomness and pseudo-randomness in computing.",
        "updated_at": "2025-01-05T20:58:37.585526+08:00",
        "identify": 2024061211,
        "birthday": "1967-08-27",
        "created_at": "2025-01-05T20:58:37.585539+08:00",
        "score": 98.3
      },
      {
        "age": 59,
        "id": 2,
        "article": "practice and theory of programming language and systems design",
        "exam_time": "11:30:00",
        "created_at": "2025-01-05T20:58:37.585541+08:00",
        "gender": 2,
        "updated_at": "2025-01-05T20:58:37.585543+08:00",
        "score": 99.1,
        "image": "SU1BR0UuUENH",
        "birthday": "1967-08-27",
        "identify": 2024070733,
        "name": "liskov"
      }
    ]
  }
]
```

返回结果：
```json
[
  {
    "id": "Ay7DApQBdHFFOkFBRxKQ",
    "rows_affected": 1,
    "version": 1,
    "status": 0
  },
  {
    "id": "BC7DApQBdHFFOkFBRxKQ",
    "rows_affected": 1,
    "version": 1,
    "status": 0
  }
]
```

ModRet 的结构体如下：
```go
// ModRet 新增/更新返回信息
type ModRet struct {
	ID          ID                     `orm:"id,omitempty" json:"id,omitempty"`                       // id 主键，可能是 mysql 的最后自增id，last_insert_id 或 elastic 的 _id 等，类型可能是 int64、string
	RowAffected int64                  `orm:"rows_affected,omitempty" json:"rows_affected,omitempty"` // 影响行数
	Version     int64                  `orm:"version,omitempty" json:"version,omitempty"`             // 数据版本
	Status      int                    `orm:"status,omitempty" json:"status,omitempty"`               // 返回状态码
	Reason      string                 `orm:"reason,omitempty" json:"reason,omitempty"`               // mod 失败原因
	Extras      map[string]interface{} `orm:"extras,omitempty" json:"extras,omitempty"`               // 更多详细信息
}

type ID string

func (id ID) String() string
func (id ID) Float64() float64
func (id ID) Int() int
func (id ID) Int64() int64
func (id ID) Uint()
func (id ID) Uint64()
```

上面语句在 es 插入了两条数据：
```eslint
GET /es_student/_search
{
  "query": {
    "match_all": {}
  }
}
```

```json
{
  "took": 0,
  "timed_out": false,
  "_shards": {
    "total": 1,
    "successful": 1,
    "skipped": 0,
    "failed": 0
  },
  "hits": {
    "total": {
      "value": 2,
      "relation": "eq"
    },
    "max_score": 1,
    "hits": [
      {
        "_index": "es_student",
        "_type": "_doc",
        "_id": "z6SONpQBT1ym-Bx5C67P",
        "_score": 1,
        "_source": {
          "age": 67,
          "article": "enhanced human understanding of the role of randomness and pseudo-randomness in computing.",
          "birthday": "1967-08-27",
          "created_at": "2025-01-05T20:58:37.585539+08:00",
          "exam_time": "14:30:00",
          "gender": 1,
          "id": 1,
          "identify": 2024061211,
          "image": "SU1BR0UuUENH",
          "name": "wigderson",
          "score": 98.3,
          "updated_at": "2025-01-05T20:58:37.585526+08:00"
        }
      },
      {
        "_index": "es_student",
        "_type": "_doc",
        "_id": "0KSONpQBT1ym-Bx5C67P",
        "_score": 1,
        "_source": {
          "age": 59,
          "article": "practice and theory of programming language and systems design",
          "birthday": "1967-08-27",
          "created_at": "2025-01-05T20:58:37.585541+08:00",
          "exam_time": "11:30:00",
          "gender": 2,
          "id": 2,
          "identify": 2024070733,
          "image": "SU1BR0UuUENH",
          "name": "liskov",
          "score": 99.1,
          "updated_at": "2025-01-05T20:58:37.585543+08:00"
        }
      }
    ]
  }
}
```

# 查询
## QUERY
### 数据维护

### 1.2.4 redis 协议
#### 1.2.4.1 基础用法
统一接入协议支持 redis 协议的 NoSQL 数据库操作。下面解释下字段含义：
- `name`：查询名称
- `op`： cmd 操作命令。
- `prefix`：前缀，强烈建议加上
- `key`：键 key
- `args`：参数


#### 1.2.4.2 返回类型
返回类型根据 op 操作决定。然后在客户端，用户可以根据接收类型，做解码。 协议不支持阻塞操作 `BLPOP`、`BRPOP`


## 1.3 事务
我们可以将同一个事务下的多个执行单元放到一个事务单元下。所有执行单元要不全部执行成功，或者全部回滚。

## 1.4 异常返回
### 1.4.1 系统异常
### 1.4.2 执行单元异常

## 1.5 高并发、高可用
### 1.5.1 缓存
### 1.5.2 异步化
#### 1.5.2.1 重试次数
### 1.5.3 限流
### 2.5.4 降级
### 2.5.5 流量切换

## 2.6 统计
### 2.6.1 数据看板
用于查看请求的平均耗时、总执行次数，总执行耗时，可以用于优化，数据对比增量，看是否有请求暴增。用于限流、降级等高可用方案。

### 2.6.2 报错分析
### 2.6.3 流量突增
### 2.6.4 告警




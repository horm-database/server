# 服务统一访问协议
服务统一访问协议，是一套可以访问所有服务而设计的协议，使用统一访问协议访问服务的优势：
* 复杂业务拼写sql/es/redis 语句访问DB，可读性差，可维护性差，开发效率低下，不同的数据库需要拼写语句，通过服务统一访问平台协议，可以极大提升开发效率，目前已经支持 mysql、postgresql、clickhouse、es、redis 等协议。
* 服务统一访问协议可以极大提升跨部门、项目之间协作的效率，降低沟通成本。
* 所有业务模块单独支持数据库访问，开发成本高，权限分散，不易管理。在服务统一访问平台 ，可以做统一配置化管理，包括接入方授权、分表、日志级别等配置。
* 当有兄弟部门服务调用需求，需要单独为其开发接口，效率低下，在服务统一访问平台，可以直接给兄弟部门授权表级别数据权限，兄弟部门可以通过 sdk 接入。也可以避免数据这里存一份、那里存一份，降低存储成本，降本增效。
* 组件解决并发性、高可用问题。比如缓存、多个执行单元并发、异步化，降级方案，针对指定接入方，可以降级为管理平台配置好的返回数据，不执行SQL。
* 高效的异常定位与解决方案，超时重试、失败手动重试功能，数据大盘可以用于 sql 性能分析，优化，可以对错误进行分析，数据暴增、快速定位暴增接入应用。
* 支持 GO、NODE、JAVA、C++、Python 等客户端的 SDK 接入。

[整体介绍](https://github.com/horm-database/doc/blob/master/server_access_platform.pdf)

```go
const ( // 支持的服务类型
    DBTypeElastic    DBType = 1  // elastic search
    DBTypeMongo      DBType = 2  // mongo 暂未支持
    DBTypeRedis      DBType = 3  // redis
    DBTypeMySQL      DBType = 10 // mysql
    DBTypePostgreSQL DBType = 11 // postgresql
    DBTypeClickHouse DBType = 12 // clickhouse
    DBTypeOracle     DBType = 13 // oracle 暂未支持
    DBTypeDB2        DBType = 14 // DB2 暂未支持
    DBTypeSQLite     DBType = 15 // sqlite 暂未支持
    DBTypeRPC        DBType = 40 // rpc 协议，暂未支持，spring cloud 协议可以选 grpc、thrift、tars、dubbo 协议
    DBTypeHTTP       DBType = 50 // http 请求
    DBTypeFunction   DBType = 60 // 函数逻辑
)
```

<img width="1993" height="1349" alt="sever" src="https://github.com/user-attachments/assets/1f25bcc5-0666-40cf-9b46-2aac81336066" />

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

# 执行单元
## 单元名称
请求协议由一个或多个执行单元组成，每个执行单元都是对平台管理的服务访问动作，可以是增删改查表数据、es索引数据、 kv 存储数据，也可以是 http 接口调用，甚至可以是逻辑函数，请求到达服务统一访问平台后，通过 `name(单元名称)` 找到对应的 mysql 表/ es索引 / redis key 、及其所属服务/库配置，然后根据协议将执行单元转化为对应数据库 sql语句或 elastic、redis、http 等请求，执行并将结果返回到客户端。

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
如果存在相同的单元名称的时候，我们可以通过增加服务名来区分，否则会报错，不允许存在相同的服务名::单元名，如下请求：
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
    "id": 234047220842770433,
    "identify": 2024080313,
    "age": 23,
    "name": "caohao",
    "article": "groundbreaking work in cryptography and complexity theory",
    "exam_time": "15:30:00",
    "created_at": "2025-01-05T20:20:06+08:00",
    "gender": 1,
    "score": 91.5,
    "image": "SU1BR0UuUENH",
    "updated_at": "2025-01-08T22:22:06+08:00"
}
```

## 执行单元结构体
一个完整的执行单元包含如下信息：
```go
// git.woa.com/horm-database/common/proto

// Unit 执行单元
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
    Val      interface{}              `json:"val,omitempty"`       // 单条记录 val (not map)
    Data     map[string]interface{}   `json:"data,omitempty"`      // maintain one map data
    Datas    []map[string]interface{} `json:"datas,omitempty"`     // maintain multiple map data
    Args     []interface{}            `json:"args,omitempty"`      // multiple args, 可用于 query 语句的参数，或者 redis 协议，如 MGET、HMGET、HDEL 等
    DataType map[string]types.Type    `json:"data_type,omitempty"` // 数据类型（主要用于 clickhouse，对于数据类型有强依赖），请求 json 不区分 int8、int16、int32、int64 等，只有 Number 类型，bytes 也会被当成 string 处理。

    // group by
    Group  []string               `json:"group,omitempty"`  // group by
    Having map[string]interface{} `json:"having,omitempty"` // group by condition

    // for databases such as mysql ...
    Join     []*Join `json:"join,omitempty"`
    Distinct bool    `json:"distinct,omitempty"`

    // for databases such as redis/elastic ...
    Field  string  `json:"field,omitempty"`  // redis key 或 hash field, 或者 elastic`s type, it can be customized before v7, and unified as _doc after v7
    Scroll *Scroll `json:"scroll,omitempty"` // scroll info

    // bytes 字节流
    Bytes []byte `json:"bytes,omitempty"`

    // params 与数据库特性相关的附加参数，例如 redis 的 WITHSCORES、EX、NX、等，以及 elastic 的 refresh、collapse、runtime_mappings、track_total_hits 等等。
    Params map[string]interface{} `json:"params,omitempty"`

    // 直接送 Query 语句，需要拥有库的表操作权限、或 root 权限。具体参数为 args
    Query string `json:"query,omitempty"`

    // Extend 扩展信息，作用于插件
    Extend map[string]interface{} `json:"extend,omitempty"`

    Sub   []*Unit          `json:"sub,omitempty"`   // 子查询
    Trans []*Unit          `json:"trans,omitempty"` // 事务，该事务下的所有 Unit 必须同时成功或失败（注意：仅适合支持事务的数据库回滚，如果数据库不支持事务，则操作不会回滚）
    Orchs []*Orchestration `json:"orchs,omitempty"` // 结果编排
}

// Orchestration 结果编排
type Orchestration struct {
    Name string                 `json:"name,omitempty"` // 编排名
    Path string                 `json:"path,omitempty"` // 编排对象路径
    Args map[string]interface{} `json:"args,omitempty"` // 编排请求参数
}

// Scroll 滚动查询
type Scroll struct {
    ID   string `json:"id,omitempty"`   // 滚动 id
    Info string `json:"info,omitempty"` // 滚动查询信息，如时间
}

// Join MySQL 表 JOIN
type Join struct {
    Table string            `json:"table,omitempty"`
    Type  string            `json:"type,omitempty"`
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

我们发送请求到数据统一调度服务的时候，绝大多数情况下可以不指定数据类型，服务端也可以正常解析并执行 query 语句，但是在某些特殊情况下，比如 clickhouse 对类型有强限制，又或者字段是一个超大 uint64 整数，json 编码之后请求服务端，由于 json 的基础类型只包含 string、 number(当成float64)、bool，数字在服务端会被解析为 float64，存在精度丢失问题，一般当类型为 time、[]byte、int、int8~int64、uint、
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
horm 基础类型，会在服务统一访问平台根据指定的数据源引擎映射、解析成对应的类型，例如在 mysql 和 clickhouse 类型映射为：
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
如果我们用到 mysql 的别名，或者在并发执行、复合执行模式下、同一层级的多个执行单元如果访问同一张表，为了结果的正常，我们必须在括号里加上别名，
如下代码的`student(add)` 和 `student(find)` ，我们都是访问 student。

```json
[
    {
        "name": "student(add)",
        "op": "insert",
        "data": {
            "identify": 2024092316,
            "score": 82.5,
            "gender": 1,
            "image": "SU1BR0UuUENH",
            "exam_time": "15:30:00",
            "age": 17,
            "name": "jerry",
            "article": "contributions to deep learning in artificial intelligence",
            "id": 377949979861331969
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

以下是上面请求的返回结果，是一个 map，其中 map 的 key 就是执行单元的名称或别名，如果都用 student，则无法区分是返回是哪个执行单元的结果，这样会丢失一个执行单元的结果，这时候需要用别名来区别。
```json
{
    "data": {
        "add": {
            "id": "377952464772542465",
            "rows_affected": 1
        },
        "find": {
            "id": 377952464772542465,
            "gender": 1,
            "age": 17,
            "name": "jerry",
            "score": 82.5,
            "image": "SU1BR0UuUENH",
            "updated_at": "2026-02-06T22:48:10+08:00",
            "identify": 2024092316,
            "article": "contributions to deep learning in artificial intelligence",
            "exam_time": "15:30:00",
            "created_at": "2026-02-06T22:48:10+08:00"
        }
    }
}
```

另外一种情况就是作为 mysql 的别名存在：
```json
{
    "name": "student_course(sc)",
    "op": "find_all",
    "column": ["sc.*", "s.name"],
    "size": 100,
    "join": [
        {
            "table": "student(s)",
            "type": "LEFT",
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
默认情况下，我们的执行单元名就等于表名，但是，如果有mysql分表、elastic分索引、redis分库的情况，我们需要用到 shard 功能来指定分表，
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

在服务统一访问平台，我们会校验 shard 表是否符合该数据的表校验规则，表校验规则支持单一表名、逗号分隔的多个表名、正则表达式 regex/student_*?/、还有就是比较常用的 `...` 校验， 例如咱们例子中的student_0...99 表示 从 student_0 一直到 student_99。

## 事务
```go
// git.woa.com/horm-database/common/proto

// Unit 执行单元
type Unit struct {
    // query base info
    Name  string   `json:"name,omitempty"`  // name
    Op    string   `json:"op,omitempty"`    // operation
    Shard []string `json:"shard,omitempty"` // 分片、分表、分库
	
    ...
	
    Trans []*Unit          `json:"trans,omitempty"` // 事务，该事务下的所有 Unit 必须同时成功或失败（注意：仅适合支持事务的数据库回滚，如果数据库不支持事务，则操作不会回滚）
    ...
}
```

如果某个执行单元 op='transaction' 的时候，则该执行单元为事务，这时候，实际的执行单元则是 `Trans []*Unit`，这些执行单元全部成功或者失败，只针对支持回滚的服务有效，例如 mysql。

# 执行模式
服务统一访问协议一共包含3种执行模式，单执行单元，并行执行，复合执行。
## 单执行单元
整个查询仅包含一个执行单元。
### 单结果返回
执行单条语句，`is_nil`, `error` 直接通过 header 返回，当查询结果为空时，is_nil=true。

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
    "name": "redis_student_zset",
    "op": "zrangebyscore",
    "params": {
        "min": 10,
        "max": 50,
        "WITHSCORES": true
    }
}
```

返回如下，member 和 score 在两个数据内返回，而且 member 是有序集成员，数组下标相同的 score 为成员的分数。：
```json
{
    "member": [
        "{\"age\":23,\"image\":\"SU1BR0UuUENH\",\"article\":\"contributions to deep learning in artificial intelligence\",\"id\":227518753250750465,\"score\":91.5,\"updated_at\":\"2024-12-18T19:58:17.869141+08:00\",\"identify\":2024080313,\"exam_time\":\"15:30:00\",\"gender\":2,\"name\":\"jerry\",\"created_at\":\"2024-12-18T19:58:17.869147+08:00\"}",
        "{\"image\":\"SU1BR0UuUENH\",\"name\":\"jerry\",\"article\":\"contributions to deep learning in artificial intelligence\",\"updated_at\":\"2024-12-18T19:40:41.184551+08:00\",\"gender\":2,\"score\":91.5,\"exam_time\":\"15:30:00\",\"id\":227514321192628225,\"created_at\":\"2024-12-18T19:40:41.184549+08:00\",\"identify\":2024080313,\"age\":23}",
        "{\"score\":91.5,\"name\":\"jerry\",\"article\":\"contributions to deep learning in artificial intelligence\",\"exam_time\":\"15:30:00\",\"updated_at\":\"2024-12-17T20:49:17.568859+08:00\",\"id\":227169198692904961,\"age\":23,\"created_at\":\"2024-12-17T20:49:17.568853+08:00\",\"gender\":2,\"identify\":2024080313,\"image\":\"SU1BR0UuUENH\"}"
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
        "total": 3,
        "total_page": 1,
        "page": 1,
        "size": 10
    },
    "data": [
        {
            "identify": 2024080313,
            "age": 23,
            "score": 91.5,
            "image": "SU1BR0UuUENH",
            "exam_time": "15:30:00",
            "gender": 1,
            "name": "caohao",
            "article": "groundbreaking work in cryptography and complexity theory",
            "created_at": "2025-01-05T20:20:06+08:00",
            "updated_at": "2025-01-08T22:22:06+08:00",
            "id": 234047220842770433
        },
        {
            "identify": 2024070746,
            "gender": 2,
            "name": "emerson",
            "image": "SU1BR0UuUENH",
            "created_at": "2025-01-11T10:40:44+08:00",
            "id": 235842198988452699,
            "age": 36,
            "score": 79.9,
            "article": "develop automated methods to detect design errors in computer hardware and software",
            "exam_time": "15:30:00",
            "updated_at": "2025-01-11T10:40:44+08:00"
        }
    ]
}
```

实际上服务统一访问平台返回的分页数据结构如下：

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

## 并行执行
### 并发同时执行
为了高效并发，我们可以将多个语句组织在一起，一同发送到到服务统一访问平台，由平台并发执行，并返回结果。

`注意：如果并行执行访问同一个数据时，为了区别，可以像下面一样在括号里面加别名：redis_student_zset(zadd) 和 redis_student_zset(range)。`<br><br>

`另外我们注意看返回结果，zrangebyscore 仅返回了2条数据，实际上应该有3条数据，也就是 zadd 的数据并未出现在 zrangebyscore 结果中， 这是因为在并发执行过程中，两个语句是同时执行，我们并不知道哪个语句先执行完，如果 zrangebyscore 先于 zadd 执行完成，就会导致数据还未插入完成就获取了排序结果，这显然与我们的预期不符，所以当遇到两条执行语句有先后要求时，我们最好拆成两条独立的语句先后执行，而不是放在一个并发执行中。`

```json
[
    {
        "name": "redis_student_zset(zadd)",
        "op": "zadd",
        "data": {
            "gender": 1,
            "age": 17,
            "name": "jerry",
            "score": 82.5,
            "image": "SU1BR0UuUENH",
            "article": "contributions to deep learning in artificial intelligence",
            "updated_at": "2026-02-07T12:25:47.665089+08:00",
            "identify": 2024092316,
            "exam_time": "15:30:00"
        },
        "params": {
            "score": 17
        }
    },
    {
        "name": "redis_student_zset(range)",
        "op": "zrangebyscore",
        "params": {
            "min": 10,
            "max": 50,
            "WITHSCORES": true
        }
    }
]
```

### 返回结果
并行执行返回结果包含两个部分，第一部分 base ，是每个查询基础信息，例如 erro、is_nil、detail，第二部分则是返回数据，key 都是执行单元名称。
```go
// ParallelResult 并行查询返回结果
type ParallelResult struct {
    Base map[string]*RetBase    `json:"base"` // 返回基础信息
    Data map[string]interface{} `json:"data"` // 返回数据
}

// RetBase 混合查询返回结果基础信息
type RetBase struct {
    Error  *Error  `json:"error,omitempty"`  // 错误返回
    IsNil  bool    `json:"is_nil,omitempty"` // 是否为空
    Detail *Detail `json:"detail,omitempty"` // 查询细节信息
}

// Detail 其他查询细节信息，例如 分页信息、滚动翻页信息、其他信息等。
type Detail struct {
    Total     uint64                 `orm:"total,omitempty" json:"total"`             // 总数
    TotalPage uint32                 `orm:"total_page,omitempty" json:"total_page"`   // 总页数
    Page      int                    `orm:"page,omitempty" json:"page"`               // 当前分页
    Size      int                    `orm:"size,omitempty" json:"size"`               // 每页大小
    Scroll    *Scroll                `orm:"scroll,omitempty" json:"scroll,omitempty"` // 滚动翻页信息
    Extras    map[string]interface{} `orm:"extras,omitempty" json:"extras,omitempty"` // 更多详细信息
}
```

上述结果返回如下：
```json
{
    "base": {
        "zadd": {},
        "range": {}
    },
    "data": {
        "zadd": 1,
        "range": {
            "member": [
                "{\"article\":\"contributions to deep learning in artificial intelligence\",\"age\":17,\"exam_time\":\"15:30:00\",\"updated_at\":\"2026-02-07T11:13:58.511697+08:00\",\"image\":\"SU1BR0UuUENH\",\"score\":82.5,\"identify\":2024092316,\"name\":\"jerry\",\"gender\":1}",
                "{\"image\":\"SU1BR0UuUENH\",\"name\":\"kitty\",\"article\":\"Artificial Intelligence\",\"updated_at\":\"2024-12-18T19:40:41.184551+08:00\",\"gender\":2,\"score\":91.5,\"exam_time\":\"15:30:00\",\"id\":227514321192628225,\"created_at\":\"2024-12-18T19:40:41.184549+08:00\",\"identify\":2024080313,\"age\":23}"
            ],
            "score": [
                17,
                23
            ]
        }
    }
}
```

### 引用
引用是指的一个执行单元的请求参数来自另外一个查询的返回结果，当出现引用的时候，并行执行会退化为串行执行。引用有多种方式，如下 {"@identify": "student.identify"} 中 map 的 key 以`@`开头的时候，表示 identify 的值引用自 student 执行单元的返回结果的 identify 字段。`.` 之前表示引用路径，之后表示引用的 field， 被引用的执行单元必须在引用的执行单元之前被执行，否则就会报错。

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
    "data": {
        "student": {
            "detail": {
                "total": 3,
                "total_page": 1,
                "page": 1,
                "size": 10
            },
            "data": [
                {
                    "exam_time": "15:30:00",
                    "updated_at": "2025-01-08T22:22:06+08:00",
                    "id": 234047220842770433,
                    "age": 23,
                    "name": "caohao",
                    "image": "SU1BR0UuUENH",
                    "created_at": "2025-01-05T20:20:06+08:00",
                    "identify": 2024080313,
                    "gender": 1,
                    "score": 91.5,
                    "article": "groundbreaking work in cryptography and complexity theory"
                },
                {
                    "image": "SU1BR0UuUENH",
                    "article": "contributions to deep learning in artificial intelligence",
                    "exam_time": "15:30:00",
                    "created_at": "2026-02-06T22:53:07+08:00",
                    "identify": 2024092316,
                    "name": "jerry",
                    "updated_at": "2026-02-06T22:53:07+08:00",
                    "id": 377953711600709633,
                    "gender": 1,
                    "age": 17,
                    "score": 82.5
                }
            ]
        },
        "student_course": [
            {
                "id": 1,
                "identify": 2024080313,
                "course": "Math",
                "hours": 54
            },
            {
                "id": 2,
                "identify": 2024080313,
                "course": "Physics",
                "hours": 32
            },
            {
                "course": "English",
                "hours": 68,
                "id": 3,
                "identify": 2024092316
            }
        ]
    },
    "base": {
        "student": {},
        "student_course": {}
    }
}
```

当引用参数是 val（interface{}）、field (string) 或者 args ([]interface{}) 而不是 horm.Where (map[string]interface{}) 的时候， 需要 `@{}` 方式，例如 @{student.identify} 来表示该参数来自于引用 student.identify。 例如下面这个例子，我们需要先查询 name="caohao" 的学生，然后根据学生返回的 identify 来获取他的排名：

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
        "name": "redis_student_zset",
        "op": "zrank",
        "val": "@{student.identify}"
    }
]
```

返回：
```json
{
    "base": {
        "student": {},
        "redis_student_zset": {}
    },
    "data": {
        "student": {
            "age": 23,
            "image": "SU1BR0UuUENH",
            "article": "groundbreaking work in cryptography and complexity theory",
            "created_at": "2025-01-05T20:20:06+08:00",
            "updated_at": "2025-01-08T22:22:06+08:00",
            "id": 234047220842770433,
            "identify": 2024080313,
            "name": "caohao",
            "score": 91.5,
            "exam_time": "15:30:00",
            "gender": 1
        },
        "redis_student_zset": 16
    }
}
```

当被引用的返回不是一个 map，而是一个具体数值的时候，我们不需要指定 field，而是直接指定被引用的执行单元名。 例如下面我们获取了一个学生的排名， 我们期望在一个并行执行单元中知道该排名的奖励：
```json
[
    {
        "name": "redis_student_zset(score_rank)",
        "op": "zrevrank",
        "val": 2024092316
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

返回结果如下：
```json
{
    "base": {
        "score_rank": {},
        "score_rank_reward": {}
    },
    "data": {
        "score_rank": 1,
        "score_rank_reward": {
            "rank": 1,
            "reward": "iphone 17 pro max 512G"
        }
    }
}
```

## 复合执行
### 返回结构
复合执行包含并行执行加上子查询，在复合执行的结果，如果返回的是一个数组，我们会为每个数组结果都执行一遍该查询的子查询，每个复合执行及其子查询的结果都包含 error、is_nil、detail 和 data 4个参数，当 error 不存在或者等于 nil 的时候，则结果正常无报错，分页等详情再 detail 中，如果返回数据为空则is_nil=true，当 is_nil 不存在，或者等于 false 时，返回数据存在于 data 中。子查询也在父查询的返回 data 中。

```go
package proto // "git.woa.com/horm-database/common/proto"

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
                        "name": "redis_student_string(test_nil)",
                        "op": "get",
                        "field": "not_exists"
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
    "test_error": {
        "error": {
            "type": 2,
            "code": 1054,
            "msg": "mysql query error: [Unknown column 'not_exist_field' in 'where clause']"
        },
        "data": null
    },
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
                "gender": 1,
                "age": 23,
                "image": "SU1BR0UuUENH",
                "exam_time": "15:30:00",
                "updated_at": "2025-01-08T22:22:06+08:00",
                "identify": 2024080313,
                "name": "caohao",
                "score": 91.5,
                "article": "groundbreaking work in cryptography and complexity theory",
                "created_at": "2025-01-05T20:20:06+08:00",
                "student_course": {
                    "data": [
                        {
                            "course_info": {
                                "data": {
                                    "course": "Math",
                                    "teacher": "Simon",
                                    "time": "11:00:00"
                                }
                            },
                            "id": 1,
                            "identify": 2024080313,
                            "course": "Math",
                            "hours": 54
                        },
                        {
                            "id": 2,
                            "identify": 2024080313,
                            "course": "Physics",
                            "hours": 32,
                            "course_info": {
                                "data": {
                                    "time": "14:00:00",
                                    "course": "Physics",
                                    "teacher": "Richard"
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
                                "is_nil": true,
                                "data": null
                            }
                        },
                        {
                            "teacher": "Simon",
                            "age": 61,
                            "test_nil": {
                                "is_nil": true,
                                "data": null
                            }
                        }
                    ]
                }
            },
            {
                "updated_at": "2026-02-06T22:53:07+08:00",
                "id": 377953711600709633,
                "gender": 1,
                "age": 17,
                "image": "SU1BR0UuUENH",
                "exam_time": "15:30:00",
                "created_at": "2026-02-06T22:53:07+08:00",
                "student_course": {
                    "data": [
                        {
                            "identify": 2024092316,
                            "course": "English",
                            "hours": 68,
                            "course_info": {
                                "data": {
                                    "course": "English",
                                    "teacher": "Dennis",
                                    "time": "15:30:00"
                                }
                            },
                            "id": 3
                        }
                    ]
                },
                "teacher_info": {
                    "data": [
                        {
                            "teacher": "Dennis",
                            "age": 39,
                            "test_nil": {
                                "is_nil": true,
                                "data": null
                            }
                        }
                    ]
                },
                "identify": 2024092316,
                "name": "jerry",
                "score": 82.5,
                "article": "contributions to deep learning in artificial intelligence"
            }
        ]
    }
}
```
### 引用路径
不同于并行执行的所有执行单元都在同一个层级，在复合执行中，有了子查询，在不同层级的情况下，引用会变得复杂，我们可以采用相对路径和绝对路径，来指向我们需要被引用的执行单元。 如果 `/` 开头，则表是该路径属于绝对路径，例如上面实例中的 `/student.identify`，否则，就是相对路径，相对路径在计算的时候，会把当前层级所在的父查询的绝对路径加在相对路径前，例如上面案例的`student_course/course_info.teacher` ，会变成 `/student/student_course/course_info.teacher`如果以 `../` 开头的相对路径，则会把`../` 转化为父查询的绝对路径，例如上面案例的 `../.course`，会变成 `/student/student_course.course`，在相对路径转化为绝对路径之后，再根据规则获取指定路径的引用结果。

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
        "name": "redis_student_string",
        "op": "get",
        "field": "noexists"
    }
]
```
```go
// is_nil = true
```
```json
[
    {
        "name": "redis_student_not_exists",
        "op": "zrangebyscore",
        "args": [
            70,
            100
        ],
        "params": {
            "WITHSCORES": true
        }
    }
]
```
```go
// is_nil = false
```

上面展示的是单执行单元的返回结果，在单执行单元中，is_nil、error 参数在 HEADER 中返回客户端：
```protobuf
/* ResponseHeader 响应头 */
message ResponseHeader {
  ...
  Error err = 4;                     // 返回错误
  bool is_nil = 5;                   // 返回是否为空
}
```

在并行执行中，返回结果包含两个部分，第一部分 base ，是每个查询基础信息，例如 erro、is_nil、detail，第二部分则是返回数据，key 都是执行单元名称。

```go
// ParallelResult 并行查询返回结果
type ParallelResult struct {
    Base map[string]*RetBase    `json:"base"` // 返回基础信息
    Data map[string]interface{} `json:"data"` // 返回数据
}

// RetBase 混合查询返回结果基础信息
type RetBase struct {
    Error  *Error  `json:"error,omitempty"`  // 错误返回
    IsNil  bool    `json:"is_nil,omitempty"` // 是否为空
    Detail *Detail `json:"detail,omitempty"` // 查询细节信息
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

返回结果如下：
```json
{
    "base": {
        "add": {
            "error": {
                "type": 2,
                "code": 1054,
                "msg": "mysql query error: [Unknown column 'no_field' in 'field list']"
            }
        },
        "find": {
            "error": {
                "type": 2,
                "code": 1054,
                "msg": "mysql query error: [Unknown column 'no_field' in 'where clause']"
            }
        }
    },
    "data": {}
}
```

在复合执行中，请求参数错误、解析失败、网络错误、权限错误等依然在 ResponseHeader 的 err 中返回，
每个执行单元的 is_nil、error 则包含在结果里面。
```go
package proto // "git.woa.com/horm-database/common/proto"

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

服务统一访问平台的错误结构如下，错误包含：错误类型，错误码，错误信息，异常查询语句组成（sql不仅指代sql语句，elastic语句、redis 命令也包含在内）
```protobuf
message Error {
  int32  type = 1; //错误类型
  int32  code = 2; //错误码
  string msg = 3;  //错误信息
  string sql = 4;  //异常sql语句
}
```

错误类型包含3大类，比如请求参数错误、解析失败、网络错误、权限错误等都属于系统错误；找不到插件、插件未注册、插件执行错误等都属于插件错误；数据库执行报错都属于数据库错误。
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
当 Elastic 批量插入新数据时，返回 `[]*proto.ModRet`，我们可以遍历返回结果，`status` 为错误码，当 `status!=0` 则该条记录插入失败，`reason`为失败原因，这样，我们可以针对失败的记录做特殊处理，比如重试。
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

上面语句在 es 的索引 es_student 插入了两条数据：
```json
{
    "took": 18,
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
                "_id": "0LCGOJwBIpMpryuCG1ge",
                "_score": 1,
                "_source": {
                    "age": 67,
                    "article": "enhanced human understanding of the role of randomness and pseudo-randomness in computing.",
                    "created_at": "2026-02-07T22:33:58.508853+08:00",
                    "exam_time": "14:30:00",
                    "gender": 1,
                    "id": 1,
                    "identify": 2024061211,
                    "image": "SU1BR0UuUENH",
                    "name": "wigderson",
                    "score": 98.3,
                    "updated_at": "2026-02-07T22:33:58.508856+08:00"
                }
            },
            {
                "_index": "es_student",
                "_type": "_doc",
                "_id": "0bCGOJwBIpMpryuCG1gg",
                "_score": 1,
                "_source": {
                    "age": 59,
                    "article": "practice and theory of programming language and systems design",
                    "created_at": "2026-02-07T22:33:58.508857+08:00",
                    "exam_time": "11:30:00",
                    "gender": 2,
                    "id": 2,
                    "identify": 2024070733,
                    "image": "SU1BR0UuUENH",
                    "name": "liskov",
                    "score": 99.1,
                    "updated_at": "2026-02-07T22:33:58.508858+08:00"
                }
            }
        ]
    }
}
```

# 可配置化插件
插件是服务统一访问平台的核心概念，通过插件，我们可以为数据库的执行赋能，并将能力沉淀复用。在统一访问平台，插件拥有很高的优先级，他可以决定是否
运行后面的插件和数据库语句执行，也可以决定是否直接返回错误， 或者已经写好的内容， 所以在使用插件的时候慎用，而插件编写者，尽量提供开关来控制，
在插件异常的时候，是否继续执行后续插件和数据库语句，例如缓存，如果已经获取到数据，则不需要再执行数据库，这是我们预期的事情，再比如日志或者监控、
统计系统等，不应该影响主流程的执行，我们要确保插件流程的 panic 被捕获并得到正确处理。这里推荐将一般的插件函数，逻辑拆成3个逻辑，
前置处理，HandleFunc 执行，后置处理，在前置处理函数和后置处理函数里加上 defer 来做 panic recover，保障 HandleFunc 正确执行。

## 插件配置
## 插件版本

### 版本回滚

# 结果编排
结果编排是在 query 到数据之后，将数据返回结果按照我们的需求，通过编排规则重新组织的一个过程，数据可以通过多个编排组件依次处理，下边是单个结果编排的请求结构：
```json
{
    "name": "setval",
    "path": "student.data(?score<=90|!(is_test=false)|(gender=1|(age+3)*2>=20)&hobby!=null&hobby[0].name='basketball')[0].hobby.weight",
    "args": {
      "type": "int",   
      "val": 30
    }
}
```
## 编排名称与参数
如上 setval 则是我们的一个编排组件，我们通过名字找到在服务统一访问平台注册的编排组件来对结果进行处理，只要是实现了 `Orchestration` 接口，都可以注册为编排组件，编排组件接口如下：
```go
// Orchestration 结果编排接口，对返回结果进行重新组织
type Orchestration interface {
    // Handle 编排处理函数。
    // input param: ctx context 上下文。
    // input param: data 返回结果
    // input param: args 编排参数。
    // output param: result 编排处理后的返回数据。
    // output param: err 插件处理异常，err 非空会直接返回客户端 error，不再执行后续逻辑。
    Handle(ctx context.Context, data interface{}, args types.Config) (interface{}, error)
}

// 注册编排组件
func Register() {
    register("extractor", &extractor.Orch{})
    register("setval", &setval.Orch{})
}
```

- 比如 setval 插件，我们根据 args 传入的参数，将 data 转化为指定类型的 val 值：
```go
type Orch struct{}

// Handle 重新组织返回结果
func (r *Orch) Handle(ctx context.Context, data interface{}, args types.Config) (v interface{}, err error) {
    typ, _ := args.GetString("type")
    switch typ {
    case "int":
       v, _, err = args.GetInt("val")
    case "uint":
       v, _, err = args.GetUint("val")
    case "float":
       v, _, err = args.GetFloat("val")
    case "bool":
       v, _ = args.GetBool("val")
    case "null":
       v = nil
    default:
       v, _ = args.GetString("val")
    }
    return v, err
}
```

## 编排路径
上面 path 表示我们要针对指定路径的结果内容进行处理，如果 path 为空则整个 result 会被传入编排函数，上面路径规则 `student.data(?score<=90|!(is_test=false)|(gender=1|(age+3)*2>=20)&hobby!=null&hobby[0].name='basketball')[0].hobby.weight` 包含如下内容：

- 路径分隔符 `.`

`.` 作为数据提取的路径分隔符，以上提取路径为 student->data->hobby->weight，最终满足条件的 weight 都会被替换为 30。

- 条件表达式 `(?条件表达式)`

如果当前路径的数据为数组，默认数组下面所有元素，都会被重新编排，但是如果包含 (?条件表达式) ，则会过滤筛选数据，满足条件的数据，才会被 setval 改变。例如 `score<=90|!(is_test=false)|(gender=1|(age+3)*2>=20)&hobby!=null&hobby[0].name='basketball'`，条件表达式可以包含标识符、NULL、Bool、String、数字、或者路径符`.`、数组符号`[]`、子条件`()`、子条件取否`!()` 以及逻辑运算符`&`（与）、`|`（或）、加减乘除四则运算（浮点精度仅支持到小数点后第10位）、还有比较运算符 `>`、`>=`、`<`、`<=`、`!=`、`=` 等，注意：条件表达式里面不可以再跟条件表达式。

- 数组筛选 `[:]`

`[0:1]`，如果当前路径为数组，并且经过条件筛选之后，我们还可以带`[0]`、`[1:3]`、`[1:]`、`[:3]`、`[-3:-1]` 等规则用于取数组数据的子集，指定子集元素会被重新编排，采用左右开区间的方式，`[1:3]`即提取下标为1到3的数据,还可以使用`[1:]`表示下标1以后的所有数据,`[:3]`表示下标3以前的所有数据，`[-3:-1]`表示下标倒数第三到倒数第一；如果左边越界，左边界会被置为0，右边越界，则右边界被置为 length-1，如果左边界>右边界，则返回空数组。如果数组筛选为 `[:]`、`[]`、`[0:-1]`、`[0:]`或`[:-1]`都表示整个数组。


## 编排示例
下面这个执行单元，首先会查询 age>10的学生，然后对返回结果依次做 setval 和 extractor 两个编排处理。
```json
{
    "name": "student",
    "op": "find_all",
    "where": {
        "age >": 10
    },
    "orchs": [
        {
            "name": "extractor",
            "path": "",
            "args": {
                "total": {
                    "$": "student.detail.total"
                },
                "courses": {
                    "$[]": "student_course(?hours>40)"
                }
            }
        },
        {
            "name": "setval",
            "path": "courses(?course='Math')[0].hours",
            "args": {
                "type": "int",
                "val": 60
            }
        }
    ]
}
```
假设我们的返回结果如下：
```json
{
    "userid":  2024080311,
    "userids": [2024080311, 2024080313, 2024092316, 2024070177, 2024031518],
    "student": {
        "detail": { "total": 2, "total_page": 1, "page": 1, "size": 10},
        "data": [
            {
                "identify": 2024080313,
                "name": "caohao",
                "age": 23,
                "gender": 1,
                "score": 91.5,
                "hobby": [{"name": "basketball", "weight": 10},
                          {"name": "football", "weight": 7}]
            },
            {
                "identify": 2024092316,
                "name": "jerry",
                "age": 17,
                "gender": 1,
                "score": 82.5,
                "hobby": [{"name": "badminton", "weight": 8 },
                          {"name": "table tennis", "weight": 7}]
            }
        ]
    },
    "sport_student": [
        {"identify": 2024070177, "name": "glad", "age": 19, "gender": 2, "score": 71.2},
        { "identify": 2024031518, "name": "moly", "age": 18, "gender": 2, "score": 67.4}
    ],
    "no1_student": {"identify": 2024080311, "name": "coby", "score": 100, "age": 22, "gender": 1},
    "student_course": [
        {"id": 2024080311, "course": "Computer", "hours": 36},
        {"id": 2024080313, "course": "Math", "hours": 54},
        {"id": 2024080313, "course": "Physics", "hours": 32},
        {"id": 2024092316, "course": "English", "hours": 68 },
        {"id": 2024070177, "course": "Sport", "hours": 30 },
        {"id": 2024031518, "course": "Sport", "hours": 30 }
    ]
}
```

首先 extractor 组件会提取 total 和 所有课时数 >40 的课程，结果如下：
```json
{
    "total": 2,
    "courses": [
        {
            "id": 2024080313,
            "course": "Math",
            "hours": 54
        },
        {
            "course": "English",
            "hours": 68,
            "id": 2024092316
        }
    ]
}
```

然后 setval 编排组件会通过路径 `courses(?course='Math')[0].hours` 将数学课的课时改成 60 ：
```json
{
    "total": 2,
    "courses": [
        {
            "id": 2024080313,
            "course": "Math",
            "hours": 60
        },
        {
            "course": "English",
            "hours": 68,
            "id": 2024092316
        }
    ]
}
```


## 数组与数组下的元素
注意一点下面两个路径的区别：
```json
{
            "name": "setval",
            "path": "student.data",
            "args": {
                "type": "null"
            }
}
```
和 
```json
{
            "name": "setval",
            "path": "student.data[]",
            "args": {
                "type": "null"
            }
}
```

前者是将 student.data 置为 null：
```json
{
    "student": {
        "detail": { "total": 2, "total_page": 1, "page": 1, "size": 10},
        "data": null
    }
}
```
后者会将 student.data 的每个元素置为 null。
```json
{
    "student": {
        "detail": { "total": 2, "total_page": 1, "page": 1, "size": 10},
        "data": [null, null]
    }
}
```


## 内容提取组件
内容提取组件 extractor 是一个将返回结果通过提取规则重新提取组织的一个过程，是结果编排中一个非常重要的组件，如下会提取两个内容，total 和 data，每个内容都由一个 map[string]interface{} 类型的提取规则组成，rule 的每个 kv 对应一个数据提取规则，key 表示提取结果类型以及数据 JOIN 规则，value 表示数据提取规则。

```json
{
    "name": "student",
    "op": "find_all",
    "orchs": [
        {
            "name": "extractor",
            "path": "",
            "args": {
                "total": {
                    "$": "student.detail.total"
                },
                "data": {
                    "$[]": "student.data(?score>80)[0:1].{identify,name}",
                    "courses(identify=id)[]": "student_course"
                }
            }
        }
    ]
}
```

返回结果：
```json
{
    "total": 2,
    "data": [
        {
            "identify": 2024080313,
            "name": "caohao",
            "courses": [
                {"id": 2024080313, "course": "Math", "hours": 54},
                {"id": 2024080313, "course": "Physics", "hours": 32}
            ]
        },
        {
            "identify": 2024092316,
            "name": "jerry",
            "courses": [
                {"id": 2024092316, "course": "English", "hours": 68}
            ]
        }
    ]
}
```

### 提取结果类型
上面规则 map 的 key 表示提取结果类型以及数据 JOIN 规则，key 可以是 `$`、`$[]`、`courses(identify=id)[]`、`courses(name&age&gender)*` 等等：

- 提取结果类型：如果 key 包含 `[]` 表示提取结果为一个数组，否则就是提取 map 数据；
- 数据 JOIN：如果 key 以 `$` 开头，表示该提取结果为主数据，内容提取的第一步首先是提取主数据，然后提取副数据，最后将副数据根据一定规则 JOIN 到主数据。 `()` 包含的就是 JOIN 规则，主副数据如果关联字段不同，比如主数据通过 identify 与副数据的 id 关联到一起，则写法是 `joined_data(identify=id)` ，主副数据如果关联字段相同，则写法可以是 `joined_data(name&age&gender)`，多个关联条件以 `&` 号来组合；
- key 最后还可以带 `*` 号，当我们的副数据提取的是 map，表示将提取到的 map 所有字段 merge 到主数据；

### 提取规则
上面规则 map 的 value 表示数据的提取规则，规则还可能是一个更复杂的表达式 `student.data(?score<=90|!(is_test=false)|(gender=1|(age+3)*2>=20)&hobby!=null&hobby[0].name='basketball')[0:1].{identify,name,age}`，包含如下内容：

- 路径分隔符 `.`

`.` 作为数据提取的路径分隔符，以上提取路径为 student->data->hobby，我们会依次递归处理 student -> data -> hobby 数据，在递归处理函数的最外层定义好结果接收器，然后将接收器指针传递下去，用于接收提取结果，当提取结果不是数组的时候，如果提取过程遇到数组，我们可以只传递满足条件的第一条数据到递归下一层，最终只取一条结果写入结果接收器；

- 条件表达式 `(?条件表达式)`

如果当前路径的数据为数组，后面有可能跟 (?条件表达式) ，主要用于过滤筛选数据，满足条件的数据，才会被保留。例如 `score<=90|!(is_test=false)|(gender=1|(age+3)*2>=20)&hobby!=null&hobby[0].name='basketball'`，条件表达式可以包含标识符、NULL、Bool、String、数字、或者路径符`.`、数组符号`[]`、子条件`()`、子条件取否`!()` 以及逻辑运算符`&`（与）、`|`（或）、加减乘除四则运算（浮点精度仅支持到小数点后第10位）、还有比较运算符 `>`、`>=`、`<`、`<=`、`!=`、`=` 等，注意：条件表达式里面不可以再跟条件表达式。

- 数组筛选 `[:]`

`[0:1]`，如果当前路径为数组，并且经过条件筛选之后，我们还可以带`[0]`、`[1:3]`、`[1:]`、`[:3]`、`[-3:-1]` 等规则用于取数组数据的子集，采用左右开区间的方式，`[1:3]`即提取下标为1到3的数据,还可以使用`[1:]`表示下标1以后的所有数据,`[:3]`表示下标3以前的所有数据，`[-3:-1]`表示下标倒数第三到倒数第一；如果左边越界，左边界会被置为0，右边越界，则右边界被置为 length-1，如果左边界>右边界，则返回空数组。如果数组筛选为 `[:]`、`[]`、`[0:-1]`、`[0:]`或`[:-1]`都表示取所有数据。

- 字段选择符号 `{}`

`{identify,name}` 表示我们仅提取结果里面的 identify、name 字段。


### 提取 map
如果 key 不包含 `[]`，表示提取一个 map 数据。例如：
- 提取规则：提取 gender=1 的数据的 identify,name,age 字段。
```json
{"$": "student.data(?gender=1).{identify,name,age}"}
```
提取结果如下:
```json
{
    "identify": 2024080313,
    "name": "caohao",
    "age": 23
}
```
`student.data(?gender=1)` 提取出来的是一个包含2条记录的数组，但是由于提取结果类型为 `map`，所以上述实际上是提取的满足条件的第一条数据，这个提取规则与下面规则是完全等价的：
```json
{"$": "student.data(?gender=1)[0].{identify,name,age}"}
```
最后的`{identify,name,age}`为字段选择，代表仅选择目标字段进行输出，也可以写为
```json
{"$": "student.data(?name='caohao'&gender=1)[0].{data:{identify,name,age}}"}
```
此时的`{data:{identify,name,age}}`表示以结构`{"data":{xxx}}`的形式输出结果,如下:
提取结果
```json
{
    "data": {
        "identify": 2024080313,
        "name": "caohao",
        "age": 23
    }
}
```

### 提取数组
- 当 key 包含 `[]` 的时候，表示提取一个数组数据，如果最终根据提取规则获取的数据是一个 map，则将该 map append 到数组结果上面去：
```json
{"$[]": "student_course(?hours>20)[1:2].{id,course}"}
```
提取结果如下：
```json
 [
    {"id": 2024080313, "course": "Math"},
    {"id": 2024080313, "course": "Physics"}
]
```

如果我们期望将 `course` 单独提取出来作为一个课程数组，可以去掉最后的 `{}`：:
```json
{"$[]": "student_course(?hours>20)[1:2].course"}
```
提取结果如下：
```json
 ["Math", "Physics"]
```

- 下面表示取下标 index <= 1 的数据：
```json
{"$[]": "student_course(?hours>20)[:1]"}
```
提取结果如下：
```json
 [
    {"id": 2024080311, "course": "Computer", "hours": 36},
    {"id": 2024080313, "course": "Math", "hours": 54}
]
```
- 下面表示取下标 index >=2 的数据：
```json
{"$[]": "student_course(?hours>20)[2:]"}
```
提取结果如下：
```json
 [
    {"id": 2024080313, "course": "Physics", "hours": 32},
    {"id": 2024092316, "course": "English", "hours": 68 },
    {"id": 2024070177, "course": "Sport", "hours": 30 },
    {"id": 2024031518, "course": "Sport", "hours": 30 }
]
```
- 下面表示取倒数第2个之后的数据：
```json
{"$[]": "student_course(?hours>20)[-2:]"}
```
提取结果如下：
```json
 [
    {"id": 2024070177, "course": "Sport", "hours": 30 },
    {"id": 2024031518, "course": "Sport", "hours": 30 }
]
```

- 下面表示取倒数第3个到倒数第2个的数据：
```json
{"$[]": "student_course(?hours>20)[-3:-2]"}
```
提取结果如下：
```json
 [
    {"id": 2024092316, "course": "English", "hours": 68 },
    {"id": 2024070177, "course": "Sport", "hours": 30 }
]
```


### 指定位置插入数据
当主数据为数组时，我们可以通过 key 来指定在对应数组位置插入数据，这个场景应用非常广泛，比如信息流，我们拉取数据之后，希望在指定位置插入广告信息。

- 可以通过 key="$!" 在数组头部插入数据，这种方式只能在头部插入一次：
```json
{
    "$[]": "student.data",
    "$!":"no1_student"
}
```
编排结果:
```json
[
    {"identify": 2024080311, "name": "coby", "score": 100, "age": 22, "gender": 1},
    {"identify": 2024080313, "name": "caohao", "age": 23, "gender": 1, "score": 91.5},
    {"identify": 2024092316, "name": "jerry", "age": 17, "gender": 1, "score": 82.5}
]
```
- 也可以通过 key="$+" 在数组尾部插入数据，用法类似上面例子，这种方式也只能在尾部插入一次。


- 我们可以通过 key="$>1" 在指定数组下标（起始为0）位置插入数据，如果有多个位置插入，则从后往前开始插入，
当 index <= 0 的时候，表示往前插入，我们可以通过负数大小来决定插入先后顺序， 比如 $>0 先于 $>-1 先于 $-2 被插入头部，同理当 index >= len(data)，我们也可以依次往结果尾部插入数据。：
```json
{
    "$[]": "student.data",
    "$>1":"sport_student"
}
```
编排结果:
```json
[
    {"identify": 2024080313, "name": "caohao", "age": 23, "gender": 1, "score": 91.5},
    {"identify": 2024070177, "name": "glad", "age": 19, "gender": 2, "score": 71.2},
    {"identify": 2024031518, "name": "moly", "age": 18, "gender": 2, "score": 67.4},
    {"identify": 2024092316, "name": "jerry", "age": 17, "gender": 1, "score": 82.5}
]
```

### 数据 JOIN
如果key 以 `$` 开头的时候，而且不以 `$>`、`$!`、`$+` 开头，则表示该规则提取出来的数据为主数据，内容提取的第一步首先是提取主数据，然后是提取副数据，副数据会根据一定规则 JOIN 到主数据：
```json
{
    "$[]": "student.data.{identify,name}",
    "courses(identify=id)[]":"student_course"
}
```
上面规则会将数据 `student_course` 根据 id 值与主数据 `student.data` 通过 identify 值关联，并以字段 `courses` JOIN 到主数据中去：
```json
[
    {
        "identify": 2024080313,
        "name": "caohao",
        "courses": [
            {"id": 2024080313, "course": "Math", "hours": 54},
            {"id": 2024080313, "course": "Physics", "hours": 32}
        ]
    },
    {
        "identify": 2024092316,
        "name": "jerry",
        "courses": [
            {"id": 2024092316, "course": "English","hours": 68}
        ]
    }
]
```
如果我们期望将所有的课程作为一个字符串数组提取出来，可以在数组`[]` 里面加上提取的字段名：
```json
{
    "$[]": "student.data.{identify,name}",
    "courses(identify=id)[course]":"student_course"
}
```
提取结果如下：
```json
[
    {
        "identify": 2024080313,
        "name": "caohao",
        "courses": ["Math", "Physics"]
    },
    {
        "identify": 2024092316,
        "name": "jerry",
        "courses": ["English"]
    }
]
```

不同于上面被 JOIN 的是一个数组，如果被 JOIN 的是一个 map，我们在 key 最后加 `*` 号则表示将提取到的结果 merge 到当前结果：
```json
{
    "$[]": "student.data.{identify,name}",
    "first_course(identify=id)*":"student_course"
}
```
注意一下 JOIN 结果，虽然有多条 course 数据根据  identify=id 被匹配到主数据，但是由于被 JOIN 的数据提取结果类型为 map，所以只有第一条数据会被 merge 到主数据。

提取结果如下：
```json
[
    {
        "identify": 2024080313,
        "name": "caohao",
        "id": 2024080313,
        "course": "Math",
        "hours": 54
    },
    {
        "identify": 2024092316,
        "name": "jerry",
        "id": 2024092316,
        "course": "English",
        "hours": 68
    }
]
```
# 数据维护
## INSERT 语句
### 返回结构
- 在 horm，数据插入的返回结构体为 `proto.ModRet` ，该结构体如下：
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

### 插入单条数据
- 往 mysql 插入单条语句
```json
{
    "name": "student",
    "op": "insert",
    "data": {
        "article": "contribution to leading the public into the era of hyper-connectivity",
        "image": "SU1BR0UuUENH",
        "created_at": "2026-02-08T22:25:55.802019+08:00",
        "identify": 2024061211,
        "name": "metcalfe",
        "gender": 2,
        "age": 39,
        "score": 93.8,
        "exam_time": "15:30:00",
        "updated_at": "2026-02-08T22:25:55.802019+08:00",
        "id": 235842198988402689
    }
}
```

返回结果如下：
```json
{
    "id": "235842198988402689",
    "rows_affected": 1
}
```

- 示例2，往 ElasticSearch插入单条记录：
```json
{
    "name": "es_student",
    "op": "insert",
    "data": {
        "gender": 2,
        "article": "contribution to leading the public into the era of hyper-connectivity",
        "exam_time": "15:30:00",
        "image": "SU1BR0UuUENH",
        "created_at": "2026-02-08T22:31:15.71306+08:00",
        "updated_at": "2026-02-08T22:31:15.71306+08:00",
        "_id": 66666,
        "id": 66666,
        "name": "metcalfe",
        "age": 39,
        "score": 93.8,
        "identify": 2024061211
    }
}
```
注意上面请求 `_id` 为指定 Elastic 数据主键，返回结果如下：
```json
{
    "id": "66666",
    "rows_affected": 1,
    "version": 1,
    "status": 0
}
```

### 插入多条数据
- 插入 MySQL 等，则返回 `proto.ModRet`， 包含批量插入最后一条记录的id和影响行数rows_affected：
```json
[
    {
        "name": "student",
        "op": "insert",
        "datas": [
            {
                "identify": 2024061291,
                "name": "metcalfe",
                "gender": 2,
                "article": "contribution to leading the public into the era of hyper-connectivity",
                "exam_time": "15:30:00",
                "image": "SU1BR0UuUENH",
                "created_at": "2026-02-08T23:18:30.990211+08:00",
                "id": 235842198988432689,
                "age": 39,
                "score": 93.8,
                "updated_at": "2026-02-08T23:18:30.990211+08:00"
            },
            {
                "article": "develop automated methods to detect design errors in computer hardware and software",
                "exam_time": "15:30:00",
                "image": "IMAGE.PCG",
                "age": 36,
                "score": 79.9,
                "created_at": "2026-02-08T23:18:30.990212+08:00",
                "updated_at": "2026-02-08T23:18:30.990212+08:00",
                "id": 235842198988452699,
                "identify": 2024070746,
                "name": "emerson",
                "gender": 2
            }
        ]
    }
]
```

返回结果：
```json
{
    "id": "235842198988452699",
    "rows_affected": 2
}
```

- 插入 Elastic 由于支持部分成功，返回的是数组 `[]*proto.ModRet` 表示每条记录的插入结果，插入的具体案例可以参考章节 [全部成功](#全部成功)

### ElasticSearch 主键
- Elastic 插入数据的时候可以通过 `_id` 字段来指定主键：
```json
{
    "name": "es_student",
    "op": "insert",
    "datas": [
        {
            "_id": 378848672248508417,
            "id": 378848672248508417,
            "score": 93.8,
            "image": "SU1BR0UuUENH",
            "created_at": "2026-02-09T10:09:22.132001+08:00",
            "age": 39,
            "article": "contribution to leading the public into the era of hyper-connectivity",
            "gender": 2,
            "name": "metcalfe",
            "exam_time": "15:30:00",
            "updated_at": "2026-02-09T10:09:22.132002+08:00",
            "identify": 2024061211
        },
        {
            "_id": 378848672248508418,
            "id": 378848672248508418,
            "exam_time": "15:30:00",
            "image": "SU1BR0UuUENH",
            "created_at": "2026-02-09T10:09:22.132004+08:00",
            "article": "develop automated methods to detect design errors in computer hardware and software",
            "updated_at": "2026-02-09T10:09:22.132004+08:00",
            "identify": 2024070733,
            "gender": 2,
            "name": "emerson",
            "score": 79.9,
            "age": 36
        }
    ]
}
```
返回结果：
```json
[
    {
        "id": "378848672248508417",
        "rows_affected": 1,
        "version": 1
    },
    {
        "rows_affected": 1,
        "version": 1,
        "id": "378848672248508418"
    }
]
```

- 也可以在 where 中加 `_id` 来指定主键：
```json
[
    {
        "name": "es_student",
        "op": "insert",
        "where": {
            "_id": 888
        },
        "data": {
            "updated_at": "2026-02-09T10:17:31.665456+08:00",
            "identify": 2024061211,
            "gender": 1,
            "image": "SU1BR0UuUENH",
            "age": 78,
            "name": "Alen Joy",
            "exam_time": "16:30:00",
            "created_at": "2026-02-09T10:17:31.665457+08:00",
            "id": 888,
            "score": 99.9,
            "article": "UNIX operating system and C programming language"
        }
    }
]
```

生成的 es 语句如下：
```json
PUT /es_student/_doc/888?op_type=create&refresh=false
{
    "age": 78,
    "article": "UNIX operating system and C programming language",
    "created_at": "2025-01-09T15:37:54.219903+08:00",
    "exam_time": "16:30:00",
    "gender": 1,
    "id": 888,
    "identify": 2024061211,
    "image": "SU1BR0UuUENH",
    "name": "Alen Joy",
    "score": 99.9,
    "updated_at": "2025-01-09T15:37:54.219914+08:00"
}
```
返回结果如下：
```json
{
    "id": "888",
    "rows_affected": 1,
    "version": 1,
    "status": 0
}
```

## REPLACE 语句
replace 和 insert 函数类似，只不过是把 sql 关键词 insert 替换为 replace。
`注意：elastic search 不支持 replace`

## UPDATE 语句
- 示例1：update by id
```json
[
    {
        "name": "student",
        "op": "update",
        "where": {
            "id": 235842198988452699
        },
        "data": {
            "age": 49,
            "updated_at": "2026-02-09T10:42:50.636937+08:00",
            "exam_time": "09:00:00"
        }
    }
]
```
返回结果：
```json
{
  "rows_affected": 2
}
```
- 示例2：update by where
```json
[
    {
        "name": "student",
        "op": "update",
        "where": {
            "age>": 40
        },
        "data": {
            "exam_time": "16:00:00",
            "updated_at": "2026-02-09T10:44:42.04203+08:00"
        },
        "data_type": {
            "updated_at": 1
        }
    }
]
```

- 示例3（Elastic update by _id）：
```json
[
    {
        "name": "es_student",
        "op": "update",
        "where": {
            "_id": 888
        },
        "data": {
            "updated_at": "2026-02-09T10:50:46.335497+08:00",
            "exam_time": "15:45:00"
        },
        "data_type": {
            "updated_at": 1
        }
    }
]
```
生成的请求：
```json
{
    "script": {
        "params": {
            "exam_time": "15:45:00",
            "updated_at": "2026-02-09T10:50:46.335497+08:00"
        },
        "source": "ctx._source.exam_time=params.exam_time;ctx._source.updated_at=params.updated_at"
    }
}
```
返回结果：
```json
{
    "id": "999",
    "rows_affected": 1,
    "version": 3,
    "status": 0
}
```
- 示例4（Elastic update by query）：
```json
[
    {
        "name": "es_student",
        "op": "update",
        "where": {
            "age >": 60
        },
        "data": {
            "exam_time": "16:00:00",
            "updated_at": "2026-02-09T10:56:03.278071+08:00"
        },
        "data_type": {
            "updated_at": 1
        }
    }
]
```

生成的请求：
```json
POST /es_student/_update_by_query?refresh=false
{
  "query": {
    "bool": {
      "filter": {
        "range": {
          "age": {
            "from": 60,
            "include_lower": false,
            "include_upper": true,
            "to": null
          }
        }
      }
    }
  },
  "script": {
    "params": {
      "exam_time": "16:00:00",
      "updated_at": "2025-01-11T22:17:25.282884+08:00"
    },
    "source": "ctx._source.exam_time=params.exam_time;ctx._source.updated_at=params.updated_at"
  }
}
```
返回结果：
```json
{
  "rows_affected": 2
}
```

## DELETE 语句
### mysql 删除
- 示例1，mysql 删除
```json
{
    "name": "student",
    "op": "delete",
    "where": {
        "name": "metcalfe"
    }
}
```
返回结果：
```json
{
  "rows_affected": 1
}
```

### elastic search 删除

- 示例2（Elastic delete by query）
```json
{
    "name": "es_student",
    "op": "delete",
    "where": {
        "name": "metcalfe"
    }
}
```

生成的 es 请求：

```json
POST /es_student/_doc/_delete_by_query?refresh=false
{
    "query": {
        "bool": {
            "filter": {
                "terms": {
                    "name": ["metcalfe"]
                }
            }
        }
    }
}
```

返回结果：
```json
{
    "rows_affected":2
}
```


- 示例3（Elastic delete by 主键）
```json
{
    "name": "es_student",
    "op": "delete",
    "where": {
        "_id": 999
    }
}
```

生成的 es 请求:
```json
DELETE /es_student/_doc/999?refresh=false
```

返回结果：
```json
{
    "_id":"999",
    "version":2,
    "rows_affected":1,
    "status":0
}
```

## refresh
在 Elastic 通过指定 params 参数 `refresh=true` 可以使数据在更新之后立即被刷新，当然，这个会导致 Elastic Search 的压力增大。 
```json
{
    "name": "es_student",
    "op": "update",
    "where": {
        "_id": "0bCGOJwBIpMpryuCG1gg"
    },
    "data": {
        "exam_time": "15:45:00",
        "updated_at": "2026-02-09T11:59:42.415441+08:00"
    },
    "params": {
        "refresh": true
    }
}
```

生成的 es 请求：
```json
POST /es_student/_update/234062949419855874?refresh=true
{
    "script": {
        "params": {
            "exam_time": "15:45:00",
            "updated_at": "2025-01-12T09:19:57.578219+08:00"
        },
        "source": "ctx._source.exam_time=params.exam_time;ctx._source.updated_at=params.updated_at"
    }
}
```


# 查询语句
## 指定查询列
通过 `column` 指定要查询的列。

- 示例1
```json
{
    "name": "student",
    "op": "find_all",
    "column": ["id", "identify", "gender", "age", "name"]
}
```


SQL语句：
```sql
 SELECT `identify` , `gender` , `age` , `name`  FROM `student`
```

返回结果：
```json
[
  {
    "identify": 2024080313,
    "name": "caohao",
    "gender": 1,
    "age": 23
  },
  {
    "identify": 2024070733,
    "name": "jerry",
    "gender": 1,
    "age": 17
  },
  {
    "identify": 2024080313,
    "name": "wigderson",
    "gender": 2,
    "age": 23
  }
]
```

- 示例 2：
```json
{
    "name": "student",
    "op": "find",
    "column": [
        "count(1) as cnt",
        "avg(age) as age",
        "sum(score) as score"
    ]
}
```

SQL语句：
```sql
 SELECT count(1) as cnt , avg(age) as age , sum(score) as score  FROM `student` LIMIT 1
```


返回结果：
```json
{
    "cnt": 5,
    "age": 21,
    "score": 456.5
}
```

## 主键查询
- 示例1，mysql 主键查询：
```json
{
    "name": "student",
    "op": "find",
    "where": {
        "identify": 2024080313
    }
}
```

SQL语句：
```sql
 SELECT * FROM `student` WHERE  `identify` = 2024080313  LIMIT 1
```
返回结果：
```json
{
    "created_at": "2024-11-30T20:53:57+08:00",
    "updated_at": "2024-12-12T19:30:37+08:00",
    "age": 23,
    "name": "caohao",
    "gender": 1,
    "score": 91.5,
    "image": "SU1BR0UuUENH",
    "article": "groundbreaking work in cryptography and complexity theory",
    "exam_time": "15:30:00",
    "id": 234047220842770433,
    "identify": 2024080313
}
```

- 示例2，mysql 主键查询：
```json
{
    "name": "student",
    "op": "find_all",
    "where": {
        "identify": [2024080313,  2024092316]
    }
}
```
SQL语句：
```sql
SELECT * FROM `student` WHERE `identify` IN (2024080313, 2024092316) 
```


- 示例3，elastic search 通过 `_id` 指定主键：
```json
{
    "name": "es_student",
    "op": "find",
    "where": {
        "_id": 999
    }
}
```
结果：
```json
{
    "_elastic": {
        "_score": 0,
        "_index": "es_student",
        "_id": "888"
    },
    "exam_time": "16:30:00",
    "image": "SU1BR0UuUENH",
    "name": "Alen Joy",
    "age": 78,
    "gender": 1,
    "id": 888,
    "identify": 2024061211,
    "score": 99.9,
    "article": "UNIX operating system and C programming language",
    "created_at": "2026-02-09T10:23:56.196973+08:00",
    "updated_at": "2026-02-09T10:23:56.196974+08:00"
}
```

## where 查询条件
### 操作符
```go
const ( // 操作符
	OPEqual          = "="   // 等于
	OPBetween        = "()"  // 在某个区间
	OPNotBetween     = "!()" // 不在某个区间
	OPGt             = ">"   // 大于
	OPGte            = ">="  // 大于等于
	OPLt             = "<"   // 小于
	OPLte            = "<="  // 小于等于
	OPNot            = "!"   // 去反
	OPLike           = "~"   // like语句，（或 es 的部分匹配）
	OPNotLike        = "!~"  // not like 语句，（或 es 的部分匹配排除）
	OPMatchPhrase    = "?"   // es 短语匹配 match_phrase
	OPNotMatchPhrase = "!?"  // es 短语匹配排除 must_not match_phrase
	OPMatch          = "*"   // es 全文搜索 match 语句
	OPNotMatch       = "!*"  // es 全文搜索排除 must_not match
)
```


### 基础用法
由于篇幅问题，下面所有用法都是用 mysql 举例，如果对应库类型为 elastic 则服务统一访问平台会生成对应的 es 请求。
```json
{
    "name": "student",
    "op": "find_all",
    "where": {
        "age": 29,                  // `age` = 29
        "age >": 29,                // `age` > 29
        "age <=": 39,               // `age` <= 29
        "age !": 29,                // `age` != 29
        "age ()": [20, 29],         // `age` BETWEEN 20 AND 29     
        "age !()": [35, 40],        //  NOT ( `age` BETWEEN 35 AND 40)       
        "score": [60, 61, 62],      // `score` IN (60, 61, 62)        
        "score !": [70, 71, 72],    // `score` NOT IN (70, 71, 72)          
        "name": null,               // `name` IS NULL
        "name !": null,             // `name` IS NOT NULL
        "name ! #注释：排除smallhow": "smallhow" //  `name` != 'smallhow'
    }
}
```

### 组合查询
针对快速构建 where 语句方式，我们也支持通过 "AND" 或者 "OR"、"NOT" 来组合更复杂的语句。

- 示例1：
```json
{
    "name": "student",
    "op": "find_all",
    "where": {
        "age >": 18,
        "score >": 85,
        "OR": {
            "id ()": [1, 100],
            "gender": 1
        }
    }
}
```

生成 SQL 语句：
```sql
SELECT * FROM `student` WHERE  `age` > 18 AND `score` > 85  AND (( `id`  BETWEEN 1 AND 100)  OR `gender` = 1) 
```

上述语句如果转化为 elastic search 的 query 条件语句，则为（es 占用篇幅较大，后面都以 MySQL 为例）：
```json
{
  "from": 0,
  "query": {
    "bool": {
      "filter": [
        {
          "bool": {
            "should": [
              {
                "range": {
                  "id": {
                    "from": 1,
                    "include_lower": true,
                    "include_upper": true,
                    "to": 100
                  }
                }
              },
              {
                "terms": {
                  "gender": [
                    1
                  ]
                }
              }
            ]
          }
        },
        {
          "range": {
            "age": {
              "from": 18,
              "include_lower": false,
              "include_upper": true,
              "to": null
            }
          }
        },
        {
          "range": {
            "score": {
              "from": 85,
              "include_lower": false,
              "include_upper": true,
              "to": null
            }
          }
        }
      ]
    }
  }
}
```

- 示例2，用注释来区分相同 key：

注意：由于 horm.Where 是map参数，所以在下面的情况下，第一个 OR 会被覆盖。
```json
{
    "name": "student",
    "op": "find_all",
    "where": {
        "OR": {
            "id >": 3,
            "gender": 1
        },
        "OR": {
            "identify !": 0,
            "age >=": 20
        }
    }
}
```

```sql
[X] SELECT * FROM `student` WHERE (`identify`!=0 OR `age`>=20)
```
通过 `#` 来注释 OR，可以保证两个条件都会工作：

```json
{
    "name": "student",
    "op": "find_all",
    "where": {
        "OR #注释1": {
            "id >": 3,
            "gender": 1
        },
        "OR #注释2": {
            "identify !": 0,
            "age >=": 20
        }
    }
}
```
```sql
[√]  SELECT * FROM `student` WHERE (`id` > 3  OR `gender` = 1)  AND (`identify` != 0 OR `age` >= 20) LIMIT 100
```
也可以用 `#` 来区分多次like
```json
{
    "name": "student",
    "op": "find_all",
    "where": {
        "OR": {
            "article ~ #1": "%computer%",
            "article ~ #2": "%medical%",
            "article ~ #3": "%physic%"
        }
    }
}
```
```sql
[√]  SELECT * FROM `student` WHERE  (`article` LIKE '%computer%' OR `article` LIKE '%medical%' OR `article` LIKE '%physic%')  LIMIT 100
```

上面 like 等同于
```json
{
    "name": "student",
    "op": "find_all",
    "where": {
        "article ~": [
            "%computer%",
            "%medical%",
            "%physic%"
        ]
    }
}
```

- 示例3，NOT 使用：
```json
{
    "name": "student",
    "op": "find_all",
    "where": {
        "NOT": {
            "id >": 3,
            "gender": 1
        }
    }
}
```
查询语句：
```sql
SELECT * FROM `student` WHERE NOT (`id` > 3 AND `gender` = 1)  LIMIT 100
 ```

- 示例 4， OR 下边的 map 数组，map里面各元素默认为 AND：
```json
{
    "name": "student",
    "op": "find_all",
    "where": {
        "OR": [
            {
                "id >": 3,
                "gender": 1
            },
            {
                "id <": 100,
                "gender": 2
            },
            {
                "id >=": 999,
                "gender": 2
            }
        ]
    }
}
```
查询语句：
```sql
SELECT * FROM `student` WHERE ((`id` > 3 AND `gender` = 1) OR (`gender` = 2 AND `id` < 100) OR (`id` >= 999 AND `gender` = 2)) LIMIT 100
```

### 模糊查询
#### SQL LIKE 
在数据库引擎为 sql 相关系统时，`~` 操作符表示 LIKE。
- 示例1
```json
{
    "name": "student",
    "op": "find_all",
    "where": {
        "name ~": "%cao%",  // `name` LIKE '%cao%'
        "name !~": "%cao%", // `name` NOT LIKE '%cao%'
        "birthday ~": ["2019-08%", "2020-01%"], // (`birthday` LIKE '2019-08%' OR `birthday` LIKE '2020-01%')
        "birthday !~": ["2019-08%","2020-01%"] // (`birthday` NOT LIKE '2019-08%' AND `birthday` NOT LIKE '2020-01%')  ## 注意他和 LIKE 的连接词不一样，NOT LIKE 是 AND，而 LIKE 是 OR
    }
}
```
- 示例2
```json
{
    "name": "student",
    "op": "find_all",
    "where": {
        "name ~ #1": "Londo_",  // London, Londox, Londos...
        "name ~ #2": "[BCR]at", // Bat, Cat, Rat
        "name ~ #3": "[!BCR]at" // Eat, Fat, Hat...
    }
}
```

#### Elastic 部分匹配
不同于 sql 相关数据库，在 elastic 中，`~` 操作符表示部分匹配。部分匹配分3种类型，prefix（默认）、wildcard、regexp

- prefix 前缀查询（默认）

如下会匹配 jerry, jerrycao, jerrybao...
```json
{
    "name": "es_student",
    "op": "find_all",
    "where": {
        "name ~": "jerry"
    }
}
```
生成的 elastic query 条件语句 ：
```json
{
  "query": {
    "bool": {
      "must": {
        "prefix": {
          "name": "jerry"
        }
      }
    }
  },
  "from": 0,
  "size": 100
}
```
- wildcard 通配符查询

如下我们为 `~` 操作符加上了 `wildcard` 属性，它使用标准的 shell 通配符查询： `?` 匹配任意字符， `*` 匹配 0 或多个字符，如下会匹配 jerry, jriy, jasteriy...
```json
{
    "name": "es_student",
    "op": "find_all",
    "where": {
        "name ~(wildcard)": "j*r?y"
    },
    "size": 100
}
```
生成的 elastic query 条件语句：
```json
{
  "query": {
    "bool": {
      "must": {
        "wildcard": {
          "name": {
            "value": "j*r?y"
          }
        }
      }
    }
  },
  "from": 0,
  "size": 100
}
```

- regexp 正则表达式查询

这个是正则查询，如下示例的正则表达式要求词必须以 W 开头，紧跟 0 至 9 之间的任何一个数字，然后接一或多个其他字符。
```json
{
    "name": "es_student",
    "op": "find_all",
    "where": {
        "name ~(regexp)": "W[0-9].+"
    },
    "size": 100
}
```
生成的 elastic query 条件语句 ：
```json
{
  "query": {
    "bool": {
      "must": {
        "regexp": {
          "name": {
            "value": "W[0-9].+"
          }
        }
      }
    }
  },
  "from": 0,
  "size": 100
}
```
- NOT 部分匹配排除，如下查询不以 cao 开头的学生：
```json
{
    "name": "es_student",
    "op": "find_all",
    "where": {
        "name !~": "cao"
    },
    "size": 100
}
```

### 短语匹配 match_phrase
在 Elastic Search 中， `match_phrase` 查询首先将查询字符串解析成一个`词项列表`，然后对这些词项进行搜索，
但只保留那些包含`全部搜索词项`，且`位置`与搜索词项相同的文档。在 horm ，我们用 `?` 操作符表示短语匹配。 `!?` 表示短语匹配排除。

```json
{
    "name": "es_student",
    "op": "find_all",
    "where": {
        "article ?": "programming"
    },
    "size": 100
}
```

生成的 elastic query 条件语句 ：
```json
{
  "query": {
    "bool": {
      "must": {
        "match_phrase": {
          "article": {
            "query": "programming"
          }
        }
      }
    }
  },
  "from": 0,
  "size": 100
}
```


#### 灵活度 slop
精确短语匹配或许是过于严格了。也许我们想要包含 “develop automated methods” 的文档也能够匹配 “develop methods”，可以为`?`操作加上 `slop` 属性如下：

```json
{
    "name": "es_student",
    "op": "find_all",
    "where": {
        "article ?(slop=1)": "develop methods"
    },
    "size": 100
}
```
生成的 elastic query 条件语句 ：
```json
{
  "query": {
    "bool": {
      "must": {
        "match_phrase": {
          "article": {
            "query": "develop methods",
            "slop": 1
          }
        }
      }
    }
  },
  "from": 0,
  "size": 100
}
```

#### 提升权重
我们可以通过指定 `boost` 属性来控制任何查询语句的相对的权重，`boost` 的默认值为 `1` ，大于 `1` 会提升一个语句的相对权重。

如下，name 中包含"caohao"的话，权重更高。那么他可能会拥有更高的 `_score`评分。
```json
{
    "name": "es_student",
    "op": "find_all",
    "where": {
        "name ?(boost=3)": "caohao",
        "article ?(boost=2)": "complexity",
        "exam_time ?(boost=1)": "15"
    },
    "order": [
        "_score desc"
    ],
    "size": 100
}
```

生成的 es 查询语句：
```json
{
  "from": 0,
  "query": {
    "bool": {
      "must": [
        {
          "match_phrase": {
            "exam_time": {
              "boost": 1,
              "query": "15"
            }
          }
        },
        {
          "match_phrase": {
            "name": {
              "boost": 3,
              "query": "caohao"
            }
          }
        },
        {
          "match_phrase": {
            "article": {
              "boost": 2,
              "query": "complexity"
            }
          }
        }
      ]
    }
  },
  "size": 100,
  "sort": [
    {
      "_score": {
        "order": "desc"
      }
    }
  ]
}
```

#### 多操作属性
一个 where 操作符可以拥有多个操作属性，通过逗号 `,` 来分隔，如下 article 有 slop 和 boost 两个条件属性。

```json
{
    "name": "es_student",
    "op": "find_all",
    "where": {
        "article ?(slop=2,boost=1)": "develop methods"
    },
    "size": 100
}
```
生成的 elastic query 条件语句 ：
```json
{
  "query": {
    "bool": {
      "must": {
        "match_phrase": {
          "article": {
            "boost": 1,
            "query": "develop methods",
            "slop": 2
          }
        }
      }
    }
  },
  "from": 0,
  "size": 100
}
```
#### 短语匹配排除
操作符 `!?` 表示短语匹配排除
```json
{
    "name": "es_student",
    "op": "find_all",
    "where": {
        "article !?": "develop"
    },
    "size": 100
}
```
生成的 elastic query 条件语句 ：
```json
{
  "query": {
    "bool": {
      "must_not": {
        "match_phrase": {
          "article": {
            "query": "develop"
          }
        }
      }
    }
  },
  "from": 0,
  "size": 100
}
```

### 全文检索 match
当数据库为 Elastic 时，`*` 操作符表示对字段进行全文检索，在全文字段中搜索到最相关的文档。

```json
{
    "name": "es_student",
    "op": "find_all",
    "where": {
        "article *": "contribution to"
    },
    "size": 100
}
```
生成的 elastic query 条件语句 ：
```json
{
  "query": {
    "bool": {
      "must": {
        "match": {
          "article": {
            "query": "contribution to"
          }
        }
      }
    }
  },
  "from": 0,
  "size": 100
}
```

#### 提高精度 operator
上述例子，中文分词会将`contribution to`分为`contribution`、`to`， 用任意查询词项匹配文档可能会导致结果中出现不相关的长尾，这是种散弹式
搜索，可能我们只想搜索包含`所有词项`的文档，也就是说，不去匹配 `contribution OR to` ，而通过匹配 `contribution AND to`找到所有文档。
`*` 可以加上 operator 属性，默认情况下该属性是 `or`。我们可以将它修改成 `and` 让所有指定词项都必须匹配：

```json
{
    "name": "es_student",
    "op": "find_all",
    "where": {
        "article *(and)": "contribution to"
    },
    "size": 100
}
```
生成的 elastic query 条件语句 ：
 ```json
{
  "query": {
    "bool": {
      "must": {
        "match": {
          "article": {
            "operator": "and",
            "query": "contribution to"
          }
        }
      }
    }
  },
  "from": 0,
  "size": 100
}
```
#### 控制精度 minimum_should_match
在所有与任意间二选一有点过于非黑即白。如果用户给定 5 个查询词项，想查找只包含其中 4 个的文档，该如何处理？
在全文搜索的大多数应用场景下，我们既想包含那些可能相关的文档，同时又排除那些不太相关的。换句话说，我们想要处于中间某种结果。
`*` 查询支持 `minimum_should_match` 最小匹配属性，这让我们可以指定必须匹配的词项数用来表示一个文档是否相关。我们可以将其设置为某个具体数字，更常用的做法是将其设置为一个百分数，因为我们无法控制用户搜索时输入的单词数量，如下，我们设置最小匹配参数为 40%，即只需要命中至少2个词，则匹配文档。

```json
{
    "name": "es_student",
    "op": "find_all",
    "where": {
        "article *(minimum_should_match=40%)": "contribution to lead the public"
    },
    "size": 100
}
```
生成的 elastic query 条件语句 ：
```json
{
  "query": {
    "bool": {
      "must": {
        "match": {
          "article": {
            "minimum_should_match": "40%",
            "query": "contribution to lead the public"
          }
        }
      }
    }
  },
  "from": 0,
  "size": 100
}
```

#### 评分计算
`bool` 查询会为每个文档计算相关度评分 `_score`，再将所有匹配的 `must` 和 `should` 语句的分数 `_score` 求和，最后除以 `must` 和  
`should` 语句的总数。`must_not`  语句不会影响评分；它的作用只是将不相关的文档排除。

#### 提升权重
提升权重与 `match_phrase` 里的用法是一样的，也是通过指定 `boost` 来控制任何查询语句的相对的权重，`boost` 的默认值为 `1`，大于 `1` 会提升一个语句的相对权重。如下，name 中包含"caohao"的话，权重更高。那么他可能会拥有更高的 `_score`评分。
```json
{
    "name": "es_student",
    "op": "find_all",
    "where": {
        "name ?(boost=3)": "caohao",
        "article ?(boost=2)": "work in",
        "exam_time ?(boost=1)": "15"
    },
    "size": 100
}
```
生成的 elastic query 条件语句 ：
```json
{
  "query": {
    "bool": {
      "must": [
        {
          "match_phrase": {
            "article": {
              "boost": 2,
              "query": "work in"
            }
          }
        },
        {
          "match_phrase": {
            "exam_time": {
              "boost": 1,
              "query": "15"
            }
          }
        },
        {
          "match_phrase": {
            "name": {
              "boost": 3,
              "query": "caohao"
            }
          }
        }
      ]
    }
  },
  "from": 0,
  "size": 100
}
```


#### 全文搜索排除
```json
{
    "name": "es_student",
    "op": "find_all",
    "where": {
        "article !*": "contribution to"
    },
    "size": 100
}
```
生成的 elastic query 条件语句 ：
```json
{
  "query": {
    "bool": {
      "must_not": {
        "match": {
          "article": {
            "query": "contribution to"
          }
        }
      }
    }
  },
  "from": 0,
  "size": 100
}
```
## 分组、聚合（暂未支持 elastic 的聚合）
### GROUP

通过 group 支持分组查询。如下查看大于10岁的学生中，各个性别、年龄段分组的学生总数和平均分：
```json
{
    "name": "student",
    "op": "find_all",
    "column": [
        "gender",
        "age",
        "count(1) as cnt",
        "avg(score) as score_avg"
    ],
    "where": {
        "age >": 10
    },
    "group": [
        "gender",
        "age"
    ]
}
```

```sql
SELECT `gender`, `age`, count(1) as cnt, avg(score) as score_avg FROM `student` WHERE  `age` > 10  GROUP BY `gender`,`age` 
```

查询结果：
```json
[
  {
    "gender": 1,
    "age": 23,
    "cnt": 1,
    "score_avg": 91.5
  },
  {
    "gender": 1,
    "age": 17,
    "cnt": 1,
    "score_avg": 82.5
  }
]
```

### HAVING
有些场景，我们需要在 group by 分组之后，根据分组聚合数据进行再次过滤，这时候我们需要用到 having，例如下面我们需要查询平均分大于90分的年龄段、性别分组的学生总数和平均分，Having 函数的参数与 where 条件一样，解析规则也是一样。
```json
{
    "name": "student",
    "op": "find_all",
    "column": [
        "gender",
        "age",
        "count(1) as cnt",
        "avg(score) as score_avg"
    ],
    "group": [
        "gender",
        "age"
    ],
    "having": {
        "score_avg >": 90
    }
}
```
```sql
SELECT `gender`, `age`, count(1) as cnt, avg(score) as score_avg FROM `student` GROUP BY `gender`,`age` HAVING  `score_avg` > 90 
```

## 排序与分页
### ORDER 排序
- 通过 `order` 函数指定排序。
```json
{
    "name": "student",
    "op": "find_all",
    "order": ["age"]
}
```
```sql
SELECT * from `student` ORDER BY age
```
```json
{
    "name": "student",
    "op": "find_all",
    "order": ["+age", "-score"]
}
```
```sql
SELECT * from `student` ORDER BY age ASC, score DESC
```
```json
{
    "name": "student",
    "op": "find_all",
    "order": ["age asc", "score desc"]
}
```
```sql
SELECT * from `student` ORDER BY age ASC, score DESC
```

- Elastic 按照相关性评分排序：
```json
{
    "name": "es_student",
    "op": "find_all",
    "where": {
        "article *": "contribution to"
    },
    "order": [
        "_score desc"
    ],
    "size": 100
}
```
请求如下：
```json
{
    "query": {
        "bool": {
            "must": {
                "match": {
                    "article": {
                        "query": "contribution to"
                    }
                }
            }
        }
    },
    "from": 0,
    "size": 100,
    "sort": [
        {
            "_score": {
                "order": "desc"
            }
        }
    ]
}
```


返回如下，根据 _elastic._score 全文检索匹配相关性评分从高到底排序。
```json
[
    {
        "_elastic": {
            "_score": 2.1302004,
            "_index": "es_student",
            "_id": "234062949419855873"
        },
        "exam_time": "15:30:00",
        "age": 39,
        "article": "contribution to leading the public into the era of hyper-connectivity",
        "created_at": "2025-01-05T21:22:35.821669+08:00",
        "gender": 2,
        "score": 93.8,
        "id": 234062949419855873,
        "identify": 2024061211,
        "image": "SU1BR0UuUENH",
        "name": "metcalfe",
        "updated_at": "2025-01-05T21:22:35.821654+08:00"
    },
    {
        "_elastic": {
            "_score": 0.7857686,
            "_index": "es_student",
            "_id": "zqR0NpQBT1ym-Bx53K4b"
        },
        "article": "contributions to deep learning in artificial intelligence",
        "exam_time": "15:30:00",
        "identify": 2024092316,
        "image": "SU1BR0UuUENH",
        "updated_at": "2025-01-05T20:33:33.041235+08:00",
        "gender": 1,
        "id": 234050606505930753,
        "name": "jerry",
        "age": 17,
        "created_at": "2025-01-05T20:33:33.04126+08:00",
        "score": 82.5
    },
    {
        "_elastic": {
            "_score": 0.6358339,
            "_index": "es_student",
            "_id": "234062949419855874"
        },
        "article": "develop automated methods to detect design errors in computer hardware and software",
        "created_at": "2025-01-05T21:22:35.82168+08:00",
        "exam_time": "15:30:00",
        "identify": 2024070733,
        "image": "SU1BR0UuUENH",
        "gender": 2,
        "id": 234062949419855874,
        "age": 36,
        "name": "emerson",
        "score": 79.9,
        "updated_at": "2025-01-05T21:22:35.821675+08:00"
    }
]
```

### LIMIT、OFFSET
通过 `Limit` 函数去指定 limit 、offset 参数。
```json
{
  "name" : "student",
  "op" : "find_all",
  "size" : 10
}
```

```sql
SELECT * from `student` LIMIT 10
```
```json
{
  "name" : "student",
  "op" : "find_all",
  "size" : 10,
  "from" : 30
}
```

```sql
SELECT * from `student` LIMIT 10 OFFSET 30
```

### 分页 PAGE
当 page >= 1 的时候请求分页数据，服务统一访问平台返回的分页数据结构如下，具体案例可以参考执行模式-单执行单元-分页返回章节。

```go
// PageResult 当 page >= 1 时会返回分页结果
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

## 返回结果高亮
在 Elastic Search 中，我们可以请求 es 将我们的检索结果中的关键词打上高亮标签返回，我们可以针对不同的字段打不同的标签，第四个参数 replace 是一个可选参数，在我们不需要原字段返回，而只需要返回带标签的内容时，将 replace 置为 true，可以减少输出内容，避免返回过大，如下：
```json
{
    "name": "es_student",
    "op": "find_all",
    "where": {
        "article *": "contribution to",
        "exam_time *": "15"
    },
    "params": {
        "highlights": [
            {
                "post_tag": "</red>",
                "field": "article",
                "pre_tag": "<red>",
                "replace": true
            },
            {
                "field": "exam_time",
                "pre_tag": "<yellow>",
                "post_tag": "</yellow>"
            }
        ]
    },
    "size": 100
}
```

生成的 Elastic Search 请求：
```json
{
    "query": {
        "bool": {
            "must": [
                {
                    "match": {
                        "article": {
                            "query": "contribution to"
                        }
                    }
                },
                {
                    "match": {
                        "exam_time": {
                            "query": "15"
                        }
                    }
                }
            ]
        }
    },
    "highlight": {
        "fields": {
            "article": {
                "post_tags": ["</red>"],
                "pre_tags": ["<red>"]
            },
            "exam_time": {
                "post_tags": ["</yellow>"],
                "pre_tags": ["<yellow>"]
            }
        }
    },
    "from": 0,
    "size": 100
}
```

返回结果如下，我们会在高亮字段前面加 `highlight_` 表示该字段为高亮结果，他是一个字符串数组，在结果中我们可以看到，原来的 article 并未返回，因为 replace 为 true：
```json
[
  {
    "_elastic": {
      "_score": 3.4667747,
      "_index": "es_student",
      "_id": "234062949419855873"
    },
    "image": "SU1BR0UuUENH",
    "age": 39,
    "created_at": "2025-01-05T21:22:35.821669+08:00",
    "gender": 2,
    "identify": 2024061211,
    "score": 93.8,
    "name": "metcalfe",
    "updated_at": "2025-01-05T21:22:35.821654+08:00",
    "id": 234062949419855873,
    "exam_time": "15:30:00",
    "highlight_article": [
      "<red>contribution</red> <red>to</red> leading the public into the era of hyper-connectivity"
    ],
    "highlight_exam_time": [
      "<yellow>15</yellow>:30:00"
    ]
  },
  {
    "_elastic": {
      "_score": 1.8139606,
      "_index": "es_student",
      "_id": "zqR0NpQBT1ym-Bx53K4b"
    },
    "age": 17,
    "score": 82.5,
    "image": "SU1BR0UuUENH",
    "name": "jerry",
    "id": 234050606505930753,
    "identify": 2024092316,
    "created_at": "2025-01-05T20:33:33.04126+08:00",
    "updated_at": "2025-01-05T20:33:33.041235+08:00",
    "gender": 1,
    "exam_time": "15:30:00",
    "highlight_article": [
      "contributions <red>to</red> deep learning in artificial intelligence"
    ],
    "highlight_exam_time": [
      "<yellow>15</yellow>:30:00"
    ]
  },
  {
    "_elastic": {
      "_score": 1.6168463,
      "_index": "es_student",
      "_id": "234062949419855874"
    },
    "age": 36,
    "created_at": "2025-01-05T21:22:35.82168+08:00",
    "identify": 2024070733,
    "name": "emerson",
    "score": 79.9,
    "gender": 2,
    "id": 234062949419855874,
    "updated_at": "2025-01-05T21:22:35.821675+08:00",
    "image": "SU1BR0UuUENH",
    "exam_time": "15:30:00",
    "highlight_article": [
      "develop automated methods <red>to</red> detect design errors in computer hardware and software"
    ],
    "highlight_exam_time": [
      "<yellow>15</yellow>:30:00"
    ]
  }
]
```
# redis 协议

## redis 键、值、参数
```go
// Unit 执行单元
type Unit struct {
	// query base info
	Name  string   `json:"name,omitempty"`  // name
	Op    string   `json:"op,omitempty"`    // operation
	Shard []string `json:"shard,omitempty"` // 分片、分表、分库

	...

	// data maintain
	Val      interface{}              `json:"val,omitempty"`       // 单条记录 val
	Data     map[string]interface{}   `json:"data,omitempty"`      // maintain one map data
	Datas    []map[string]interface{} `json:"datas,omitempty"`     // maintain multiple map data
	Args     []interface{}            `json:"args,omitempty"`      // multiple args

	// for databases such as redis/elastic ...
	Field  string  `json:"field,omitempty"`  // redis key 或 hash field,

	// params 与数据库特性相关的附加参数，例如 redis 的 WITHSCORES、EX、NX、等
	Params map[string]interface{} `json:"params,omitempty"`
	...
}
```
###  key 
redis 除了键操作和字符串类型会将 key 通过 `field` 字段传入服务端之外，`哈希、列表、集合、有序集` 的 key，一般情况下，都在平台设置。<br>
特殊情况也可以通过 shard 来指定，然后在平台对 shard 做校验，比如每个学生的信息用哈希存储：

```json
{
  "name" : "redis_student_hash",
  "op" : "hmset",
  "shard" : [ "student_2024061211" ],
  "data" : {
    "name" : "metcalfe",
    "image" : "SU1BR0UuUENH",
    "updated_at" : "2026-02-11T20:43:29.886223+08:00",
    "gender" : 2,
    "score" : 93.8,
    "article" : "contribution to leading the public into the era of hyper-connectivity",
    "exam_time" : "15:30:00",
    "identify" : 2024061211,
    "age" : 39
  }
}
```

```shell
hmset student_2024061211 name metcalfe score 93.8 gender 2 age 39 image SU1BR0UuUENH article contribution to leading the public into the era of hyper-connectivity exam_time 15:30:00 identify 2024061211
```

当然，如果你平台设置了 key，那么即便请求指定了 shard ，也会忽略用户指定的 shard，而选择平台的 key

### prefix 前缀
`Prefix` 可以为我们的 key 加上前缀，当我们在管理平台设置了 Prefix: "student_"，如下案例真正的 key 就是 `student_2024080313`，<br>

`强烈建议所有 key 都加上前缀，便于服务统一访问平台根据 Prefix 来对不同的对象的区分统计，如果没有加 Prefix，所有 key 都在一个库里面，
比如有 student_2024080313, teacher_Simon，那么我们如果希望统计每天查询 student 的请求量是多少，则无法统计，也无法更好的定位具体是哪个对象
发生了请求突发暴增的情况。`

```json
{
    "name": "redis_student_string",
    "op": "get",
    "field": "2024080313"
}
```

### value
我们可以将值通过 val、data、datas、args 传输到服务端，他们有一定的区别：<br>

`单条记录`，可以直接通过 val 传到平台，如果是复杂Object则会被序列化之后整体 set 到 redis <br>
`多条记录`，可以通过 args 传到服务端，每个元素都会被序列化<br>
`data`，单个 map 数据，会通过 data 字段传到服务端，json_encode 之后 set 到 redis，<br>
`datas`，多个 map 数据，会通过 datas 字段传到服务端，每条记录 json_encode 之后 set 到 redis <br>

由于集合/有序集的元素唯一性，但是 golang 的 map 类型是无序的，所以我们会将 map 按照 `key` 排序之后再 json_encode，然后 set 到 redis<br>

### params
redis 的 score、min、max、WITHSCORES、EX、NX、等等参数都会通过 params 带到服务端



## 返回结构
horm 所有支持的 redis 操作，一共会返回8种类型的结构：
* 无返回，仅包含 error。这类操作包含 `EXPIRE` 、 `SET` 、  `SETEX` 、 `HSET` 、 `HMSET` ，仅返回 error，如果无 error 则执行成功。
* 返回 `[]byte`，这类操作包含 `GET` 、 `GETSET` 、 `HGET` 、 `LPOP` 、 `RPOP`。
* 返回 `bool`，这类操作包含 `EXISTS` 、 `SETNX` 、 `HEXISTS` 、 `HSETNX` 、 `SISMEMBER`。
* 返回 `int64`，这类操作包含 这类操作包含 `INCR` 、 `DECR` 、 `INCRBY` 、 `HINCRBY` 、 `TTL` 、 `DEL` 、 `HDEL` 、 `HLEN` 、 `LPUSH` 、 `RPUSH` 、 `LLEN` 、 `SADD` 、 `SREM` 、 `SCARD` 、 `SMOVE` 、 `ZADD` 、 `ZREM` 、 `ZREMRANGEBYSCORE` 、 `ZREMRANGEBYRANK` 、 `ZCARD` 、 `ZRANK` 、 `ZREVRANK` 、 `ZCOUNT`。
* 返回 `float64`，这类操作包含 `ZSCORE` 、 `ZINCRBY`。
* 返回 `[][]byte`，这类操作包含 `HKEYS` 、 `SMEMBERS` 、 `SRANDMEMBER` 、 `SPOP` 、 `ZRANGE` 、 `ZRANGEBYSCORE` 、 `ZREVRANGE` 、 `ZREVRANGEBYSCORE`。
* 返回 `map[string]string`，这类操作包含 `HGETALL` 、 `HMGET`。
* 返回 `map[string]float64`，这类操作包含 `ZPOPMIN`
* 返回 `member 和 score（类型为 [][]byte、[]float64）`，这类操作包含 `ZRANGE ... WITHSCORES` 、 `ZRANGEBYSCORE ... WITHSCORES` 、 `ZREVRANGE ... WITHSCORES` 、 `ZREVRANGEBYSCORE ... WITHSCORES`


## 键操作
我们用 field 字段来表示 redis 的 key。

- `EXPIRE`

 设置 key 的过期时间，key 过期后将不再可用。单位以秒计。<br>
 `param`: seconds 到期时间

```json
{
  "name" : "redis_student_string",
  "op" : "expire",
  "field" : "student_2024061211",
  "params" : {
    "seconds" : 3600
  }
}
```
生成 redis 命令：
```shell
expire student_2024061211 3600
```
报错则返回 err，成功不返回。

- `TTL`

以秒为单位返回 key 的剩余过期时间。
```json
{
  "name" : "redis_student_string",
  "op" : "ttl",
  "field" : "student_2024061211"
}
```

生成 redis 命令：
```shell
ttl student_2024061211
```

返回结果：
当 key 不存在时，返回 -2 。 当 key 存在但没有设置剩余生存时间时，返回 -1 。 否则，以秒为单位，返回 key 的剩余生存时间。
```go
3391  
```

- `EXISTS`

查看值是否存在
```json
{
  "name" : "redis_student_string",
  "op" : "exists",
  "field" : "student_2024061211"
}
```

生成 redis 命令：
```shell
exists student_2024061211
```

返回结果：
存在则返回 true，否则返回 false
```json
false
```

- `DEL`
删除已存在的键。不存在的 key 会被忽略。

```json
{
  "name" : "redis_student_string",
  "op" : "del",
  "field" : "student_2024061211"
}
```

生成 redis 命令：
```shell
del student_2024061211
```
返回结果：
被删除 key 的数量。
```go
1
```

## 字符串
- `SET`

设置给定 key 的值。如果 key 已经存储其他值， Set 就覆写旧值。 。<br>
`param`: 其他参数:  包含 [NX | XX] [GET] [EX seconds | PX milliseconds | EXAT unix-time-seconds | PXAT unix-time-milliseconds | KEEPTTL]<br>

`示例1`：
```json
{
  "name" : "redis_student_string",
  "op" : "set",
  "field" : "test_float",
  "val" : 63.2567,
  "params" : {
    "NX" : true,
    "PX" : 5
  }
}
```
生成 redis 命令：
```shell
set test_float 63.2567 NX PX 5
```

`示例2`：
```json
{
  "name" : "redis_student_string",
  "op" : "set",
  "field" : "test_bool",
  "val" : true
}
```
生成 redis 命令：
```shell
set test_bool true
```


`示例3`：
```json
{
  "name" : "redis_student_string",
  "op" : "set",
  "field" : "test_int",
  "val" : 78
}
```
生成 redis 命令：
```shell
set test_int 78
```


`示例4`：
```json
{
  "name" : "redis_student_string",
  "op" : "set",
  "field" : "test_string",
  "val" : "i am ok"
}
```
生成 redis 命令：
```shell
set test_string i am ok
```


`示例5`：
```json
{
  "name" : "redis_student_string",
  "op" : "set",
  "field" : "test_struct",
  "data" : {
    "article" : "contribution to leading the public into the era of hyper-connectivity",
    "name" : "metcalfe",
    "score" : 93.8,
    "image" : "SU1BR0UuUENH",
    "identify" : 2024061211,
    "gender" : 2,
    "updated_at" : "2026-02-11T11:52:57.289878+08:00",
    "age" : 39,
    "exam_time" : "15:30:00"
  }
}
```
生成 redis 命令：
```shell
set test_struct {"score":93.8,"image":"SU1BR0UuUENH","identify":2024061211,"gender":2,"updated_at":"2026-02-11T11:52:57.289878+08:00","article":"contribution to leading the public into the era of hyper-connectivity","name":"metcalfe","age":39,"exam_time":"15:30:00"}
```


- `SETEX`

指定的 key 设置值及其过期时间。如果 key 已经存在， SETEX 命令将会替换旧的值。<br>
`param`: seconds int 到期时间<br>

```json
{
  "name" : "redis_student_string",
  "op" : "setex",
  "field" : "test_float",
  "val" : 63.2567,
  "params" : {
    "seconds" : 3600
  }
}
```

生成 redis 命令：
```shell
setex test_float 3600 63.2567
```

- `SETNX`

key 不存在时，为 key 设置指定的值。

```json
{
  "name" : "redis_student_string",
  "op" : "setnx",
  "field" : "test_struct",
  "data" : {
    "exam_time" : "15:30:00",
    "name" : "metcalfe",
    "score" : 93.8,
    "updated_at" : "2026-02-11T19:14:07.695046+08:00",
    "image" : "SU1BR0UuUENH",
    "article" : "contribution to leading the public into the era of hyper-connectivity",
    "identify" : 2024061211,
    "gender" : 2,
    "age" : 39
  }
}
```

生成 redis 命令：
```shell
setnx test_struct {"score":93.8,"article":"contribution to leading the public into the era of hyper-connectivity","identify":2024061211,"name":"metcalfe","updated_at":"2026-02-11T19:14:07.695046+08:00","image":"SU1BR0UuUENH","gender":2,"age":39,"exam_time":"15:30:00"}
```

- `GET`

获取指定 key 的值。如果 key 不存在，返回 nil 。可用 IsNil(err) 判断是否key不存在，如果key储存的值不是字符串类型，返回一个错误。 <br>

`示例1`：
```json
{
  "name" : "redis_student_string",
  "op" : "get",
  "field" : "test_bool"
}
```

生成 redis 命令：
```shell
get test_bool
```

返回结果：
```go
true
```


`示例2`：
```json
{
  "name" : "redis_student_string",
  "op" : "get",
  "field" : "test_int"
}
```

生成 redis 命令：
```shell
get test_int
```

返回结果：
```go
78
```

`示例3`：
```json
{
  "name" : "redis_student_string",
  "op" : "get",
  "field" : "test_float"
}
```

生成 redis 命令：
```shell
get test_float
```

返回结果：
```go
63.2567
```

`示例4`：
```json
{
  "name" : "redis_student_string",
  "op" : "get",
  "field" : "test_string"
}
```

生成 redis 命令：
```shell
get test_string
```

返回结果：
```go
i am ok
```

`示例5`：
```json
{
  "name" : "redis_student_string",
  "op" : "get",
  "field" : "test_struct"
}
```

生成 redis 命令：
```shell
get test_struct
```

返回结果：
```go
{
  "score" : 93.8,
  "image" : "SU1BR0UuUENH",
  "identify" : 2024061211,
  "gender" : 2,
  "updated_at" : "2026-02-11T11:52:57.289878+08:00",
  "article" : "contribution to leading the public into the era of hyper-connectivity",
  "name" : "metcalfe",
  "age" : 39,
  "exam_time" : "15:30:00"
}
```

- `GETSET`

设置给定 key 的值。如果 key 已经存储其他值， GetSet 就覆写旧值，并返回原来的值。如果原来未设置值，则 is_nil=true <br>

```json
{
  "name" : "redis_student_string",
  "op" : "getset",
  "field" : "test_struct",
  "data" : {
    "age" : 39,
    "score" : 93.8,
    "image" : "SU1BR0UuUENH",
    "article" : "contribution to leading the public into the era of hyper-connectivity",
    "exam_time" : "15:30:00",
    "name" : "metcalfe",
    "gender" : 2,
    "updated_at" : "2026-02-11T19:26:02.644605+08:00",
    "identify" : 2024061211
  }
}
```

生成 redis 命令：
```shell
getset test_struct {"image":"SU1BR0UuUENH","exam_time":"15:30:00","updated_at":"2026-02-11T19:26:02.644605+08:00","article":"contribution to leading the public into the era of hyper-connectivity","name":"metcalfe","gender":2,"identify":2024061211,"age":39,"score":93.8}
```

返回结果：
```go
{
  "score" : 93.8,
  "image" : "SU1BR0UuUENH",
  "identify" : 2024061211,
  "gender" : 2,
  "updated_at" : "2026-02-11T11:52:57.289878+08:00",
  "article" : "contribution to leading the public into the era of hyper-connectivity",
  "name" : "metcalfe",
  "age" : 39,
  "exam_time" : "15:30:00"
}
```

- `INCR`

将 key 中储存的数字值增一。 如果 key 不存在，那么 key 的值会先被初始化为 0 ，然后再执行 INCR 操作。 如果值包含错误的类型，或字符串类型的值不能表示为数字，那么返回一个错误。 <br>

```json
{
  "name" : "redis_student_string",
  "op" : "incr",
  "field" : "test_int"
}
```

生成 redis 命令：
```shell
incr test_int
```

返回结果：
```go
80
```

- `DECR`

将 key 中储存的数字值减一。如果 key 不存在，那么 key 的值会先被初始化为 0 ，然后再执行 DECR 操作。如果值包含错误的类型，或字符串类型的值不能表示为数字，那么返回一个错误。 <br>

```json
{
  "name" : "redis_student_string",
  "op" : "decr",
  "field" : "test_int"
}
```

生成 redis 命令：
```shell
decr test_int
```

返回结果：
```go
79
```

- `INCRBY`

将 key 中储存的数字加上指定的增量值。如果 key 不存在，那么 key 的值会先被初始化为 0 ，然后再执行 INCRBY 命令。如果值包含错误的类型，或字符串类型的值不能表示为数字，那么返回一个错误。<br>

`param`: increment 自增数量
```json
{
  "name" : "redis_student_string",
  "op" : "incrby",
  "field" : "test_int",
  "params" : {
    "increment" : 15
  }
}
```

生成 redis 命令：
```shell
incrby test_int 15
```

返回结果：
```go
94
```

- `MSET`

批量设置一个或多个 key-value 对
`注意：这里所有的 key 都会被加上平台设置的 Prefix`
```json
{
  "name" : "redis_student_string",
  "op" : "mset",
  "data" : {
    "c" : 25,
    "a" : 19,
    "b" : 21
  }
}
```

生成 redis 命令：
```shell
mset c 25 a 19 b 21
```

结果：
```bash
127.0.0.1:6379> get a
"19"
127.0.0.1:6379> get b
"21"
127.0.0.1:6379> get c
"25"
```


- `MGET`

返回多个 key 的 val
```json
{
  "name" : "redis_student_string",
  "op" : "mget",
  "args" : [ "a", "b", "c" ]
}
```

生成 redis 命令：
```shell
mget a b c
```

返回结果，数组三个元素分别对应 a、b、c 的结果。
```go
[ "19", "21", "25" ]
```

- `SETBIT`

设置或清除指定偏移量上的位<br>
`param`: offset uint32 参数必须大于或等于 0 ，小于 2^32 (bit 映射被限制在 512 MB 之内)<br>
`param`: value int 1-设置, 0-清除

```json
{
  "name" : "redis_student_string",
  "op" : "setbit",
  "field" : "test_string",
  "params" : {
    "value" : 1,
    "offset" : 5
  }
}
```

生成 redis 命令：
```shell
setbit test_string 5 1
```

- `GETBIT`

获取指定偏移量上的位<br>
`param`: offset uint32 参数必须大于或等于 0 ，小于 2^32 (bit 映射被限制在 512 MB 之内)<br>

```json
{
  "name" : "redis_student_string",
  "op" : "getbit",
  "field" : "test_string",
  "params" : {
    "offset" : 5
  }
}
```

生成 redis 命令：
```shell
getbit test_string 5
```

- `BITCOUNT`

计算给定字符串中，被设置为 1 的比特位的数量<br>
`param`: key string<br>
`param`: start int 可以使用负数值： 比如 -1 表示最后一个字节， -2 表示倒数第二个字节，以此类推<br>
`param`: end int 可以使用负数值： 比如 -1 表示最后一个字节， -2 表示倒数第二个字节，以此类推<br>
`param`: [BYTE | BIT]<br>

```json
{
  "name" : "redis_student_string",
  "op" : "bitcount",
  "field" : "test_string",
  "params" : {
    "start" : 1,
    "end" : -1,
    "BYTE" : true
  }
}
```

生成 redis 命令：
```shell
bitcount test_string 1 -1 BYTE
```






## 哈希

- `HSET`

为哈希表中的字段赋值 。
```json
{
  "name" : "redis_student_hash",
  "op" : "hset",
  "field" : "age",
  "val" : 39
}
```

生成 redis 命令：
```shell
hset redis_student_hash age 39
```

hset 多条的时候
```json
{
  "name" : "redis_student_hash",
  "op" : "hset",
  "data" : {
    "age" : 39,
    "name" : "smallhowcao"
  }
}
```
生成 redis 命令：
```shell
hset redis_student_hash age 39 name smallhowcao
```

- `HSETNX`
为哈希表中不存在的的字段赋值。
```json
{
  "name" : "redis_student_hash",
  "op" : "hsetnx",
  "field" : "13324",
  "val" : 22
}
```

生成 redis 命令：
```shell
hsetnx redis_student_hash 13324 22
```

返回结果，设置成功，返回true 。 如果给定字段已经存在且没有操作被执行，返回 false：
```json
true
```

- `HMSET`
把map数据设置到哈希表中，此命令会覆盖哈希表中已存在的字段。如果哈希表不存在，会创建一个空哈希表，并执行 HMSET 操作。

```json
{
  "name" : "redis_student_hash",
  "op" : "hmset",
  "data" : {
    "name" : "metcalfe",
    "image" : "SU1BR0UuUENH",
    "updated_at" : "2026-02-11T20:43:29.886223+08:00",
    "gender" : 2,
    "score" : 93.8,
    "article" : "contribution to leading the public into the era of hyper-connectivity",
    "exam_time" : "15:30:00",
    "identify" : 2024061211,
    "age" : 39
  }
}
```

生成 redis 命令：
```shell
hmset redis_student_hash image SU1BR0UuUENH gender 2 exam_time 15:30:00 age 39 name metcalfe updated_at 2026-02-11T20:43:29.886223+08:00 score 93.8 article contribution to leading the public into the era of hyper-connectivity identify 2024061211
```

结果
```bash
127.0.0.1:6379> hgetall redis_student_hash
 1) "age"
 2) "39"
 3) "13324"
 4) "22"
 5) "updated_at"
 6) "2026-02-11T20:43:29.886223+08:00"
 7) "name"
 8) "metcalfe"
 9) "exam_time"
10) "15:30:00"
11) "article"
12) "contribution to leading the public into the era of hyper-connectivity"
13) "image"
14) "SU1BR0UuUENH"
15) "identify"
16) "2024061211"
17) "score"
18) "93.8"
19) "gender"
20) "2"
```

- `HMGET`
返回哈希表中，一个或多个给定字段的值。
```json
{
  "name" : "redis_student_hash",
  "op" : "hmget",
  "args" : [ "identify", "name", "age", "gender", "score" ]
}
```

生成 redis 命令：
```shell
hmget redis_student_hash identify name age gender score
```

返回结果：
```json
{
  "score" : "93.8",
  "identify" : "2024061211",
  "name" : "metcalfe",
  "age" : "39",
  "gender" : "2"
}
```



- `HGET`

数据从redis hget 出来
```json
{
  "name" : "redis_student_hash",
  "op" : "hget",
  "field" : "age"
}
```

生成 redis 命令：
```shell
hget redis_student_hash age
```

返回结果：
```json
39
```


- `HGETALL`

返回哈希表中，所有的字段和值。

```json
{
  "name" : "redis_student_hash",
  "op" : "hgetall"
}
```

生成 redis 命令：
```shell
hgetall redis_student_hash
```

返回结果：
```json
{
  "13324" : "22",
  "name" : "metcalfe",
  "exam_time" : "15:30:00",
  "age" : "39",
  "image" : "SU1BR0UuUENH",
  "identify" : "2024061211",
  "score" : "93.8",
  "updated_at" : "2026-02-11T20:43:29.886223+08:00",
  "article" : "contribution to leading the public into the era of hyper-connectivity",
  "gender" : "2"
}
```



- `HKEYS`

获取哈希表中的所有域（field）。

```json
{
  "name" : "redis_student_hash",
  "op" : "hkeys"
}
```

生成 redis 命令：
```shell
hkeys redis_student_hash
```

返回结果：
```json
[ "updated_at", "name", "exam_time", "article", "age", "image", "identify", "score", "gender" ]
```


- `HINCRBY`

为哈希表中的字段值加上指定增量值。<br>
`param`: increment string 自增数量

```json
{
  "name" : "redis_student_hash",
  "op" : "hincrby",
  "field" : "age",
  "params" : {
    "increment" : 5
  }
}
```

生成 redis 命令：
```shell
hincrby redis_student_hash age 5
```

返回结果：
```json
44
```

- `HDEL`

删除哈希表 key 中的一个或多个指定字段，不存在的字段将被忽略。

```json
{
  "name" : "redis_student_hash",
  "op" : "hdel",
  "args" : [ "19827", "23312", "98322" ]
}
```

生成 redis 命令：
```shell
hdel redis_student_hash 19827 23312 98322
```

返回结果，被删除的字段数：
```json
0
```

- `HEXISTS`

查看哈希表的指定字段是否存在。

```json
{
  "name" : "redis_student_hash",
  "op" : "hexists",
  "field" : "age"
}
```

生成 redis 命令：
```shell
hexists redis_student_hash age
```

返回结果：
```json
true
```


- `HLEN`

获取哈希表中字段的数量。

```json
{
  "name" : "redis_student_hash",
  "op" : "hlen"
}
```

生成 redis 命令：
```shell
hlen redis_student_hash
```

返回结果：
```json
11
```

- `HSTRLEN`

获取哈希表某个字段长度。

```json
{
  "name" : "redis_student_hash",
  "op" : "hstrlen",
  "field" : "article"
}
```

生成 redis 命令：
```shell
hstrlen redis_student_hash article
```

返回结果：
```json
69
```

- `HINCRBYFLOAT`

为哈希表中的字段值加上指定增量浮点数。<br>
`param`: field string<br>
`param`: incr float64 自增数量<br>

```json
{
  "name" : "redis_student_hash",
  "op" : "hincrbyfloat",
  "field" : "score",
  "params" : {
    "increment" : 3.7
  }
}
```

生成 redis 命令：
```shell
hincrbyfloat redis_student_hash score 3.7
```

返回结果：
```json
97.5
```

- `HVALS`

返回所有的 val

```json
{
  "name" : "redis_student_hash",
  "op" : "hvals"
}
```

生成 redis 命令：
```shell
hvals redis_student_hash
```

返回结果：
```json
[
    "22",
    "2026-02-11T20:43:29.886223+08:00",
    "metcalfe",
    "15:30:00",
    "contribution to leading the public into the era of hyper-connectivity",
    "44",
    "SU1BR0UuUENH",
    "2024061211",
    "97.5",
    "0001-01-01",
    "2"
]
```

## 列表
- `LPUSH`

LPush 将一个或多个值插入到列表头部。 如果 key 不存在，一个空列表会被创建并执行 LPUSH 操作。 当 key 存在但不是列表类型时，返回一个错误。

`示例1`：
```json
{
  "name" : "redis_student_list",
  "op" : "lpush",
  "val" : 12345
}
```
生成 redis 命令：
```shell
lpush redis_student_list 12345
```

`示例2`：
```json
{
  "name" : "redis_student_list",
  "op" : "lpush",
  "args" : [ "aaa", "bbb", "ccc" ]
}
```
生成 redis 命令：
```shell
lpush redis_student_list aaa bbb ccc
```

返回结果，执行 LPUSH 命令后，列表的长度：
```json
4
```

`示例3`：
```json
{
  "name" : "redis_student_list",
  "op" : "lpush",
  "data" : {
    "identify" : 2024061211,
    "gender" : 2,
    "age" : 39,
    "name" : "metcalfe",
    "score" : 93.8
  }
}
```

生成 redis 命令：
```shell
lpush redis_student_list {"age":39,"name":"metcalfe","score":93.8,"identify":2024061211,"gender":2}
```

`示例4` PUSH 两个 map：

```json
{
    "name": "redis_student_list",
    "op": "lpush",
    "datas": [
        {
            "a": 11,
            "b": 22
        },
        {
            "a": 33,
            "b": 44
        }
    ]
}
```
生成 redis 命令：
```shell
lpush redis_student_list {"a":11,"b":22} {"a":33,"b":44}
```

`示例5` 如果你要 PUSH 的数据就是 map 数组，则采用如下方式，下面 val 无论是什么类型会被整体 json_encode 之后写入列表：

```json
{
    "name": "redis_student_list",
    "op": "lpush",
    "val": [
        {
            "a": 11,
            "b": 22
        },
        {
            "a": 33,
            "b": 44
        }
    ]
}
```
生成 redis 命令：
```shell
lpush redis_student_list [{"a":11,"b":22},{"a":33,"b":44}]
```


- `RPUSH`

将一个或多个值插入到列表的尾部(最右边)。如果列表不存在，一个空列表会被创建并执行 RPUSH 操作。 当列表存在但不是列表类型时，返回一个错误。<br>

使用示例与 `LPUSH` 一样，唯一插入位置不同，一个左插入，一个右边插入



- `LPOP`

移除并返回列表的第一个元素。

```json
{
  "name" : "redis_student_list",
  "op" : "lpop"
}
```

生成 redis 命令：
```shell
lpop redis_student_list
```

返回结果：
```json
[
    {
        "a": 11,
        "b": 22
    },
    {
        "a": 33,
        "b": 44
    }
]
```

- `RPOP`

移除列表的最后一个元素，返回值为移除的元素。<br>
使用示例与 `LPUSH` 一样，唯一插入位置不同，一个左弹出，一个右弹出

- `LLEN`

返回列表的长度。 如果列表 key 不存在，则 key 被解释为一个空列表，返回 0 。 如果 key 不是列表类型，返回一个错误。

```json
{
  "name" : "redis_student_list",
  "op" : "llen"
}
```

生成 redis 命令：
```shell
llen redis_student_list
```

返回结果：
```json
3
```


## 集合

- `SADD`

将一个或多个成员元素加入到集合中，已经存在于集合的成员元素将被忽略。

`示例1`：
```json
{
  "name" : "redis_student_set",
  "op" : "sadd",
  "val" : 12345
}
```

生成 redis 命令：
```shell
sadd redis_student_set 12345
```

返回结果，被添加到集合中的新元素的数量，不包括被忽略的元素。：
```json
1
```

`示例2`：
```json
{
  "name" : "redis_student_set",
  "op" : "sadd",
  "args" : [ "aaa", "bbb", "ccc" ]
}
```
生成 redis 命令：
```shell
sadd redis_student_set aaa bbb ccc
```

`示例3`，由于集合的元素具有唯一性，所以平台会对 map 按照 key 做排序之后再 json_encode 并写入集合：
```json
{
  "name" : "redis_student_set",
  "op" : "sadd",
  "data" : {
    "identify" : 2024061211,
    "gender" : 2,
    "age" : 39,
    "name" : "metcalfe",
    "score" : 93.8
  }
}
```

生成 redis 命令：
```shell
sadd redis_student_set {"age":39,"gender":2,"identify":2024061211,"name":"metcalfe","score":93.8}
```

`示例4` PUSH 两个 map：

```json
{
    "name": "redis_student_set",
    "op": "sadd",
    "datas": [
        {
            "e": 11,
            "b": 22,
            "a": 33
        },
        {
            "e": 88,
            "a": 99,
            "b": 77
        }
    ]
}
```
生成 redis 命令：
```shell
sadd redis_student_set {"a":33,"b":22,"e":11} {"a":99,"b":77,"e":88}
```

`示例5` 如果你要 PUSH 的数据就是 map 数组，则采用如下方式，下面 val 无论是什么类型会被整体 json_encode 之后写入列表：

```json
{
    "name": "redis_student_set",
    "op": "val",
    "datas": [
        {
            "e": 11,
            "b": 22,
            "a": 33
        },
        {
            "e": 88,
            "a": 99,
            "b": 77
        }
    ]
}
```
生成 redis 命令：
```shell
sadd redis_student_set [{"a":33,"b":22,"e":11},{"a":99,"b":77,"e":88}]
```

- `SMEMBERS`

返回集合中的所有的成员。 不存在的集合 key 被视为空集合。

```json
{
  "name" : "redis_student_set",
  "op" : "smembers"
}
```

生成 redis 命令：
```shell
smembers redis_student_set
```

返回结果：
```json
[
    "bbb",
    "ccc",
    "{\"age\":39,\"gender\":2,\"identify\":2024061211,\"name\":\"metcalfe\",\"score\":93.8}",
    "aaa",
    "{\"a\":99,\"b\":77,\"e\":88}"
]
```


- `SREM`

将一个或多个成员元素加入到集合中，已经存在于集合的成员元素将被忽略。

用法和 SADD 一样，这里仅展示一个示例：

```json
{
  "name" : "redis_student_set",
  "op" : "srem",
  "args" : [ "aaa", "bbb", "ccc" ]
}
```

生成 redis 命令：
```shell
srem redis_student_set aaa bbb ccc
```

返回结果，被移除的元素数量，不包括被忽略的元素。：
```json
3
```


- `SCARD`

返回集合中元素的数量。

```json
{
  "name" : "redis_student_set",
  "op" : "scard"
}
```

生成 redis 命令：
```shell
scard redis_student_set
```

返回结果，集合元素数量：
```json
5
```

- `SISMEMBER`

判断成员元素是否是集合的成员。

```json
{
  "name" : "redis_student_set",
  "op" : "sismember",
  "data" : {
    "identify" : 2024061211,
    "gender" : 2,
    "age" : 39,
    "name" : "metcalfe",
    "score" : 93.8
  }
}
```

生成 redis 命令：
```shell
sismember redis_student_set {"age":39,"gender":2,"identify":2024061211,"name":"metcalfe","score":93.8}
```

返回结果：
```json
true
```

- `SRANDMEMBER`

返回集合中的count个随机元素。<br>
`param`: count int 随机返回元素个数。<br>
如果 count 为正数，且小于集合基数，那么命令返回一个包含 count 个元素的数组，数组中的元素各不相同。<br>
如果 count 大于等于集合基数，那么返回整个集合。<br>
如果 count 为负数，那么命令返回一个数组，数组中的元素可能会重复出现多次，而数组的长度为 count 的绝对值。<br>

```json
{
  "name" : "redis_student_set",
  "op" : "srandmember",
  "params" : {
    "count" : 2
  }
}
```

生成 redis 命令：
```shell
srandmember redis_student_set 2
```

返回结果：
```json
[
    "{\"a\":99,\"b\":77,\"e\":88}",
    "{\"age\":39,\"exam_time\":\"15:30:00\",\"gender\":2,\"identify\":2024061211,\"name\":\"metcalfe\",\"score\":93.8}"
]
```

- `SPOP`

移除集合中的指定 key 的一个或多个随机成员，移除后会返回移除的成员。<br>
`param`: count int 随机返回元素个数。<br>

```json
{
  "name" : "redis_student_set",
  "op" : "spop",
  "params" : {
    "count" : 2
  }
}
```

生成 redis 命令：
```shell
spop redis_student_set 2
```

返回结果：
```json
[
    "{\"a\":99,\"b\":77,\"e\":88}",
    "{\"age\":39,\"exam_time\":\"15:30:00\",\"gender\":2,\"identify\":2024061211,\"name\":\"metcalfe\",\"score\":93.8}"
]
```


- `SMOVE`

将指定成员 member 元素从 source 集合移动到 destination 集合。<br>
`param`: destination 目标集合 key

```json
{
  "name" : "redis_student_set",
  "op" : "smove",
  "data" : {
    "identify" : 2024061211,
    "gender" : 2,
    "age" : 39,
    "name" : "metcalfe",
    "score" : 93.8
  },
  "params" : {
    "destination" : "destination_key"
  }
}
```

生成 redis 命令：
```shell
smove redis_student_set destination_key {"age":39,"gender":2,"identify":2024061211,"name":"metcalfe","score":93.8}
```

返回结果，成功迁移的元素个数：
```json
1
```

## 有序集

- `ZADD`

将成员元素及其分数值加入到有序集当中。如果某个成员已经是有序集的成员，那么更新这个成员的分数值，并通过重新插入这个成员元素，来保证该成员在正确的位置上。分数值可以是整数值或双精度浮点数。<br>
`param`：score/scores/score_field 插入数据的分数。


`示例1`，插入单条数据：
```json
{
  "name" : "redis_student_zset",
  "op" : "zadd",
  "val" : 2024061211,
  "params" : {
    "score" : 97
  }
}
```

生成 redis 命令：
```shell
zadd redis_student_zset 97 2024061211
```

返回结果，被成功添加的新成员的数量，不包括那些被更新的、已经存在的成员。：
```json
1
```

`示例2`，插入多条记录，参数 scores 是一个数组：
```json
{
    "name": "redis_student_zset",
    "op": "zadd",
    "args": [
        2024061211,
        2024061212
    ],
    "params": {
        "scores": [
            97,
            89
        ]
    }
}
```

生成 redis 命令：
```shell
zadd redis_student_zset 97 2024061211 89 2024061212
```

`示例3`，插入 map 数据：
```json
{
    "name": "redis_student_zset",
    "op": "zadd",
    "data": {
        "score": 97,
        "identify": 2024061211,
        "gender": 2,
        "age": 39,
        "name": "metcalfe"
    },
    "params": {
        "score": 97
    }
}
```

`注意：由于有序集合的元素具有唯一性，所以平台会对 map 按照 key 做排序之后再 json_encode 并写入有序集合`

生成 redis 命令：
```shell
zadd redis_student_zset 97 {"age":39,"gender":2,"identify":2024061211,"name":"metcalfe","score":97}
```


`示例4`，插入多个 map 数据：
```json
{
    "name": "redis_student_zset",
    "op": "zadd",
    "datas": [
        {
            "age": 39,
            "name": "metcalfe",
            "score": 97,
            "identify": 2024061211,
            "gender": 2
        },
        {
            "identify": 2024061212,
            "gender": 1,
            "age": 78,
            "name": "Alen Joy",
            "score": 89
        }
    ],
    "params": {
        "scores": [
            97,
            89
        ]
    }
}
```

生成 redis 命令：
```shell
zadd redis_student_zset 97 {"age":39,"gender":2,"identify":2024061211,"name":"metcalfe","score":97} 89 {"age":78,"gender":1,"identify":2024061212,"name":"Alen Joy","score":89}
```

`示例5`，插入多个 map 数据的时候，我们还可以指定 map 的哪个字段为分数：
```json
{
    "name": "redis_student_zset",
    "op": "zadd",
    "datas": [
        {
            "gender": 2,
            "age": 39,
            "name": "metcalfe",
            "score": 97,
            "identify": 2024061211
        },
        {
            "score": 89,
            "identify": 2024061212,
            "gender": 1,
            "age": 78,
            "name": "Alen Joy"
        }
    ],
    "params": {
        "score_field": "age"
    }
}
```

生成 redis 命令：
```shell
zadd redis_student_zset 39 {"age":39,"gender":2,"identify":2024061211,"name":"metcalfe","score":97} 78 {"age":78,"gender":1,"identify":2024061212,"name":"Alen Joy","score":89}
```


- `ZREM`

移除有序集中的一个或多个成员，不存在的成员将被忽略。<br>
用法和 SADD 一样，不需要带 score 参数，这里仅展示一个示例：
```json
{
  "name" : "redis_student_set",
  "op" : "zrem",
  "args" : [ 2024061211, 2024061212 ]
}
```

生成 redis 命令：
```shell
zrem redis_student_zset 2024061211 2024061212
```

返回结果，被成功移除的成员的数量，不包括被忽略的成员。：
```json
2
```


- `ZREMRANGEBYSCORE`

移除有序集中，指定分数（score）区间内的所有成员。<br>
`param`: min max 分数区间，类型为整数或者浮点数

```json
{
  "name" : "redis_student_zset",
  "op" : "zremrangebyscore",
  "params" : {
    "min" : 90,
    "max" : 100
  }
}
```

生成 redis 命令：
```shell
zremrangebyscore redis_student_zset 90 100
```

返回结果，被移除成员的数量。：
```json
3
```

- `ZREMRANGEBYRANK`

移除有序集中，指定排名(rank)区间内的所有成员。<br>
`param`: start stop int 排名区间

```json
{
  "name" : "redis_student_zset",
  "op" : "zremrangebyrank",
  "params" : {
    "start" : 0,
    "stop" : 2
  }
}
```

生成 redis 命令：
```shell
zremrangebyrank redis_student_zset 0 2
```

返回结果：
```json
1
```

- `ZCARD`

返回有序集成员个数

```json
{
  "name" : "redis_student_zset",
  "op" : "zcard"
}
```

生成 redis 命令：
```shell
zcard redis_student_zset
```

返回结果，有序集成员个数：
```json
1
```

- `ZRANK`

返回有序集中指定成员的排名。其中有序集成员按分数值递增(从小到大)顺序排列。

```json
{
  "name" : "redis_student_zset",
  "op" : "zrank",
  "data" : {
    "name" : "metcalfe",
    "score" : 97,
    "identify" : 2024061211,
    "gender" : 2,
    "age" : 39
  }
}
```

生成 redis 命令：
```shell
zrank redis_student_zset {"age":39,"gender":2,"identify":2024061211,"name":"metcalfe","score":97}
```

返回结果：
```json
1
```

- `ZREVRANK`

ZRevRank 返回有序集中指定成员的排名。其中有序集成员按分数值递增(从大到小)顺序排列。<br><br>
用法和 ZRANK 一样。


- `ZCOUNT`

计算有序集合中指定分数区间的成员数量<br>
`param`: min, max 分数区间<br>

```json
{
  "name" : "redis_student_zset",
  "op" : "zcount",
  "params" : {
    "min" : 70,
    "max" : 80
  }
}
```

生成 redis 命令：
```shell
zcount redis_student_zset 70 80
```

返回结果：
```json
1
```

- `ZPOPMIN`

移除并弹出有序集合中分值最小的 count 个元素<br>
`param`: count 不设置count参数时，弹出一个元素

```json
{
  "name" : "redis_student_zset",
  "op" : "zpopmin",
  "params" : {
    "count" : 2
  }
}
```

生成 redis 命令：
```shell
zpopmin redis_student_zset 2
```

返回结果，返回包含集合元素 member 以及元素对应的分数 score：
```json
{
    "member": [
        "{\"age\":39,\"gender\":2,\"identify\":2024061211,\"name\":\"metcalfe\",\"score\":97}",
        "{\"age\":78,\"gender\":1,\"identify\":2024061212,\"name\":\"Alen Joy\",\"score\":89}"
    ],
    "score": [
        39,
        78
    ]
}
```

- `ZPOPMAX`


移除并弹出有序集合中分值最大的 count 个元素<br>
`param`: count 不设置count参数时，弹出一个元素<br><br>

用法和 ZPOPMIN 一样。


- `ZINCRBY`

对有序集合中指定成员的分数加上增量 increment，也可以通过传递一个负数值 increment，让分数减去相应的值，<br>
当 key 不存在，或不是 key 的成员时，相当于新增成员，分数置为 increment。<br>
当 key 不是有序集类型时，返回一个错误。分数值可以是整数值或双精度浮点数。<br>
`param`: increment 增量值，可以为整数或双精度浮点<br>
```json
{
  "name" : "redis_student_zset",
  "op" : "zincrby",
  "val" : "2024061211",
  "params" : {
    "increment" : 9.5
  }
}
```

生成 redis 命令：
```shell
zincrby redis_student_zset 9.5 2024061211
```

返回结果，member 成员的新分数值。：
```json
116
```

- `ZRANGE`

返回有序集中，指定区间内的成员。其中成员的位置按分数值递增(从小到大)来排序。<br>
`param`: int start, stop 以 0 表示有序集第一个成员，以 1 表示有序集第二个成员，你也可以使用负数下标，以 -1 表示最后一个成员， -2 表示倒数第二个成员，以此类推。<br>
`param`: args [BYSCORE | BYLEX] [REV] [LIMIT offset count] [WITHSCORES]<br>
"WITHSCORES" 是否返回有序集的分数，结果分开在两个数组存储，但是数组下标是一一对应的，比如 member[3] 成员的分数是 score[3]

`示例1`
```json
{
  "name" : "redis_student_zset",
  "op" : "zrange",
  "params" : {
    "start" : 0,
    "stop" : 9
  }
}
```

生成 redis 命令：
```shell
zrange redis_student_zset 0 9
```

返回结果，返回集合成员：
```json
[
    "bbbbbb",
    "2024061212",
    "{\"age\":39,\"gender\":2,\"identify\":2024061211,\"name\":\"metcalfe\",\"score\":97}",
    "2024061211"
]
```

`示例2`
```json
{
  "name" : "redis_student_zset",
  "op" : "zrange",
  "params" : {
    "start" : 0,
    "stop" : 9,
    "WITHSCORES" : true
  }
}
```

生成 redis 命令：
```shell
zrange redis_student_zset 0 9 WITHSCORES
```

返回结果，返回集合成员，及其对应分数：
```json
{
    "member": [
        "bbbbbb",
        "2024061212",
        "{\"age\":39,\"gender\":2,\"identify\":2024061211,\"name\":\"metcalfe\",\"score\":97}",
        "2024061211"
    ],
    "score": [
        19,
        89,
        97,
        116
    ]
}
```


`示例3`
```json
{
  "name" : "redis_student_zset",
  "op" : "zrange",
  "params" : {
    "start" : 0,
    "stop" : 10,
    "BYSCORE" : true,
    "REV" : true,
    "LIMIT" : [ 0, 2 ]
  }
}
```

生成 redis 命令：
```shell
zrange redis_student_zset 0 10 BYSCORE REV LIMIT 0 2
```


- `ZRANGEBYSCORE`

根据分数返回有序集中指定区间的成员，顺序从小到大<br>
`param`: int min, max 分数的范围，类型必须为 int, float，但是 -inf +inf 表示负正无穷大<br>
`param`: args [LIMIT offset count] [WITHSCORES]<br>
"WITHSCORES" 是否返回有序集的分数，结果分开在两个数组存储，但是数组下标是一一对应的，比如 member[3] 成员的分数是 score[3]<br>

`示例1`：
```json
{
  "name" : "redis_student_zset",
  "op" : "zrangebyscore",
  "params" : {
    "min" : 70,
    "max" : 100
  }
}
```

生成 redis 命令：
```shell
zrangebyscore redis_student_zset 70 100
```

返回结果，返回集合成员：
```json
[
    "2024061212",
    "{\"age\":39,\"gender\":2,\"identify\":2024061211,\"name\":\"metcalfe\",\"score\":97}"
]
```


`示例2`：
```json
{
  "name" : "redis_student_zset",
  "op" : "zrangebyscore",
  "params" : {
    "min" : 70,
    "max" : 100,
    "WITHSCORES" : true
  }
}
```

生成 redis 命令：
```shell
zrangebyscore redis_student_zset 70 100 WITHSCORES
```

返回结果，返回集合成员，及其对应分数：
```json
{
    "member": [
        "2024061212",
        "{\"age\":39,\"gender\":2,\"identify\":2024061211,\"name\":\"metcalfe\",\"score\":97}"
    ],
    "score": [
        89,
        97
    ]
}
```


`示例3`：
```json
{
  "name" : "redis_student_zset",
  "op" : "zrangebyscore",
  "params" : {
    "min" : 70,
    "max" : 100,
    "LIMIT" : [ 0, 1 ],
    "WITHSCORES" : true
  }
}
```

生成 redis 命令：
```shell
zrangebyscore redis_student_zset 70 100 WITHSCORES LIMIT 0 1
```

返回结果：
```json
{
  "member" : [ "2024061212" ],
  "score" : [ 89 ]
}
```

- `ZREVRANGE`

返回有序集中指定区间的成员，其中成员的位置按分数值递减(从大到小)来排列。<br>
`param`: start, stop 排名区间，以 0 表示有序集第一个成员，以 1 表示有序集第二个成员，你也可以使用负数下标，以 -1 表示最后一个成员， -2 表示倒数第二个成员，以此类推。<br>
`param`: "WITHSCORES" 是否返回有序集的分数，结果分开在两个数组存储，但是数组下标是一一对应的，比如 member[3] 成员的分数是 score[3]<br><br>
用法跟 ZRANGE 类似，只不过是按分数递减，另外只支持 start、stop、WITHSCORES 三个参数。<br>

`示例1`
```json
{
  "name" : "redis_student_zset",
  "op" : "zrevrange",
  "params" : {
    "start" : 0,
    "stop" : 9
  }
}
```

生成 redis 命令：
```shell
zrevrange redis_student_zset 0 9
```


`示例2`
```json
{
  "name" : "redis_student_zset",
  "op" : "zrevrange",
  "params" : {
    "start" : 0,
    "stop" : 9,
    "WITHSCORES" : true
  }
}
```

生成 redis 命令：
```shell
zrevrange redis_student_zset 0 9 WITHSCORES
```

- `ZREVRANGEBYSCORE`

返回有序集中指定分数区间内的所有的成员。有序集成员按分数值递减(从大到小)的次序排列。<br>
`param`: max, min  interface{} 分数区间，类型为整数或双精度浮点数，但是 -inf +inf 表示负正无穷大<br>
`param`: args [LIMIT offset count] [WITHSCORES]<br>
"WITHSCORES" 是否返回有序集的分数，结果分开在两个数组存储，但是数组下标是一一对应的，比如 member[3] 成员的分数是 score[3]<br><br>
用法跟 ZRANGEBYSCORE 类似，只不过是按分数递减<br>



`示例1`：
```json
{
  "name" : "redis_student_zset",
  "op" : "zrevrangebyscore",
  "params" : {
    "min" : 70,
    "max" : 100
  }
}
```

生成 redis 命令：
```shell
zrevrangebyscore redis_student_zset 70 100
```


`示例2`：
```json
{
  "name" : "redis_student_zset",
  "op" : "zrevrangebyscore",
  "params" : {
    "min" : 70,
    "max" : 100,
    "WITHSCORES" : true
  }
}
```

生成 redis 命令：
```shell
zrevrangebyscore redis_student_zset 70 100 WITHSCORES
```


`示例3`：
```json
{
  "name" : "redis_student_zset",
  "op" : "zrevrangebyscore",
  "params" : {
    "min" : 70,
    "max" : 100,
    "LIMIT" : [ 0, 1 ],
    "WITHSCORES" : true
  }
}
```

生成 redis 命令：
```shell
zrevrangebyscore redis_student_zset 70 100 WITHSCORES LIMIT 0 1
```

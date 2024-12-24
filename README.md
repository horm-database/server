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
    PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='学生表';

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
  "id": 1,
  "identify": 2024061211,
  "age": 19,
  "score": 89.7,
  "image": "SU1BR0UuUENH",
  "exam_time": "15:30:00",
  "birthday": "1995-03-23T00:00:00+08:00",
  "gender": 1,
  "name": "caohao",
  "article": "Compilation theory, architecture of large systems, and development of Reduced Instruction Set (RISC) computers",
  "updated_at": "2024-12-12T19:30:37+08:00"
}
```

## 查询单元结构体
一个完整的执行单元包含如下信息：
```go
// github.com/horm-database/common/proto
package proto

import (
	"github.com/horm-database/common/consts"
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

	// 数据更新
	Data     map[string]interface{}     `json:"data,omitempty"`      // add/update one data
	Datas    []map[string]interface{}   `json:"datas,omitempty"`     // batch add/update data
	DataType map[string]consts.DataType `json:"data_type,omitempty"` // 数据类型（主要用于 clickhouse，对于数据类型有强依赖），请求 json 不区分 int8、int16、int32、int64 等，只有 Number 类型，bytes 也会被当成 string 处理。

	// group by
	Group  []string               `json:"group,omitempty"`  // group by
	Having map[string]interface{} `json:"having,omitempty"` // group by condition

	// for databases such as mysql ...
	Join []*Join `json:"join,omitempty"`

	// for databases such as elastic ...
	Type   string  `json:"type,omitempty"`   // type, such as elastic`s type, it can be customized before v7, and unified as _doc after v7
	Scroll *Scroll `json:"scroll,omitempty"` // scroll info

	// for databases such as redis ...
	Prefix string        `json:"prefix,omitempty"` // prefix, It is strongly recommended to bring it to facilitate finer-grained summary statistics, otherwise the statistical granularity can only be cmd ，such as GET、SET、HGET ...
	Key    string        `json:"key,omitempty"`    // key
	Args   []interface{} `json:"args,omitempty"`   // args 参数的数据类型存于 data_type

	// bytes 字节流
	Bytes []byte `json:"bytes,omitempty"`

	// params 与数据库特性相关的附加参数，例如 redis 的 WITHSCORES，以及 elastic 的 refresh、collapse、runtime_mappings、track_total_hits 等等。
	Params map[string]interface{} `json:"params,omitempty"`

	// 直接送 Query 语句，需要拥有库的 表权限、或 root 权限。具体参数为 args
	Query string `json:"query,omitempty"`

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

## 别名
如果我们用到 mysql 的别名，或者在并发查询、复合查询模式下、同一层级的多个查询单元如果访问同一张表，为了结果的正常，我们必须在括号里加上别名，
如下代码的`student(add)` 和 `student(find)` ，我们都是访问 student。
```json
[
  {
    "name": "student(add)",
    "op": "insert",
    "data": {
      "id": 227759629650636801,
      "identify": 2024080313,
      "name": "kitty",
      "image": "SU1BR0UuUENH",
      "article": "Artificial Intelligence",
      "created_at": "2024-12-19T11:55:27.278103+08:00",
      "updated_at": "2024-12-19T11:55:27.278105+08:00",
      "age": 23,
      "birthday": "1987-08-27T00:00:00Z",
      "gender": 2,
      "score": 91.5,
      "exam_time": "15:30:00"
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
    "id": "227759629650636801",
    "rows_affected": 1
  },
  "find": {
    "id": 227759629650636801,
    "name": "kitty",
    "article": "Artificial Intelligence",
    "created_at": "2024-12-19T11:55:27+08:00",
    "birthday": "1987-08-27T00:00:00+09:00",
    "updated_at": "2024-12-19T11:55:27+08:00",
    "identify": 2024080313,
    "gender": 2,
    "age": 23,
    "score": 91.5,
    "image": "SU1BR0UuUENH",
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
  "id": 1,
  "identify": 2024061211,
  "age": 19,
  "score": 89.7,
  "image": "SU1BR0UuUENH",
  "exam_time": "15:30:00",
  "birthday": "1995-03-23T00:00:00+08:00",
  "gender": 1,
  "name": "caohao",
  "article": "Compilation theory, architecture of large systems, and development of Reduced Instruction Set (RISC) computers",
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
    "age": 19,
    "name": "caohao",
    "score": 89.7,
    "article": "Compilation theory, architecture of large systems, and development of Reduced Instruction Set (RISC) computers",
    "exam_time": "15:30:00",
    "birthday": "1995-03-23T00:00:00+08:00",
    "id": 1,
    "identify": 2024061211,
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
        "{\"age\":23,\"image\":\"SU1BR0UuUENH\",\"article\":\"Artificial Intelligence\",\"id\":227518753250750465,\"score\":91.5,\"birthday\":\"1987-08-27T00:00:00Z\",\"updated_at\":\"2024-12-18T19:58:17.869141+08:00\",\"identify\":2024080313,\"exam_time\":\"15:30:00\",\"gender\":2,\"name\":\"kitty\",\"created_at\":\"2024-12-18T19:58:17.869147+08:00\"}",
        "{\"image\":\"SU1BR0UuUENH\",\"birthday\":\"0001-01-01T00:00:00Z\",\"name\":\"kitty\",\"article\":\"Artificial Intelligence\",\"updated_at\":\"2024-12-18T19:40:41.184551+08:00\",\"gender\":2,\"score\":91.5,\"exam_time\":\"15:30:00\",\"id\":227514321192628225,\"created_at\":\"2024-12-18T19:40:41.184549+08:00\",\"identify\":2024080313,\"age\":23}",
        "{\"score\":91.5,\"birthday\":\"0001-01-01T00:00:00Z\",\"name\":\"kitty\",\"article\":\"Artificial Intelligence\",\"exam_time\":\"15:30:00\",\"updated_at\":\"2024-12-17T20:49:17.568859+08:00\",\"id\":227169198692904961,\"age\":23,\"created_at\":\"2024-12-17T20:49:17.568853+08:00\",\"gender\":2,\"identify\":2024080313,\"image\":\"SU1BR0UuUENH\"}"
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
            "updated_at": "2024-12-12T19:30:37+08:00",
            "identify": 2024061211,
            "name": "caohao",
            "score": 89.7,
            "image": "SU1BR0UuUENH",
            "exam_time": "15:30:00",
            "created_at": "2024-11-30T20:53:57+08:00",
            "id": 1,
            "gender": 1,
            "age": 19,
            "article": "Compilation theory, architecture of large systems, and development of Reduced Instruction Set (RISC) computers",
            "birthday": "1995-03-23T00:00:00+08:00"
        },
        {
            "updated_at": "2024-12-12T20:41:00+08:00",
            "identify": 2024070733,
            "gender": 1,
            "age": 17,
            "image": "SU1BR0UuUENH",
            "exam_time": "14:30:00",
            "birthday": "1993-02-22T00:00:00+08:00",
            "created_at": "2024-11-30T20:57:03+08:00",
            "id": 2,
            "name": "jerry",
            "score": 92.3,
            "article": "Design and analysis of algorithms and data structures"
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
            23,
            "{\"id\":227523618735665153,\"gender\":2,\"age\":23,\"name\":\"kitty\",\"identify\":2024080313,\"created_at\":\"2024-12-18T20:17:37.89164+08:00\",\"image\":\"SU1BR0UuUENH\",\"birthday\":\"1987-08-27T00:00:00Z\",\"updated_at\":\"2024-12-18T20:17:37.891649+08:00\",\"article\":\"Artificial Intelligence\",\"exam_time\":\"15:30:00\",\"score\":91.5}"
        ]
    },
    {
        "name": "redis_student(range)",
        "op": "zrangebyscore",
        "key": "student_age_rank",
        "args": [10, 50],
        "params": {
            "with_scores": true
        }
    }
]
```

### 引用（同层级）
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
                "score": 89.7,
                "image": "SU1BR0UuUENH",
                "article": "Compilation theory, architecture of large systems, and development of Reduced Instruction Set (RISC) computers",
                "exam_time": "15:30:00",
                "gender": 1,
                "identify": 2024061211,
                "age": 19,
                "name": "caohao",
                "birthday": "1995-03-23T00:00:00+08:00",
                "created_at": "2024-11-30T20:53:57+08:00",
                "updated_at": "2024-12-12T19:30:37+08:00",
                "id": 1
            },
            {
                "article": "Design and analysis of algorithms and data structures",
                "exam_time": "14:30:00",
                "birthday": "1993-02-22T00:00:00+08:00",
                "id": 2,
                "gender": 1,
                "age": 17,
                "name": "jerry",
                "score": 92.3,
                "created_at": "2024-11-30T20:57:03+08:00",
                "updated_at": "2024-12-12T20:41:00+08:00",
                "identify": 2024070733,
                "image": "SU1BR0UuUENH"
            }
        ]
    },
    "student_course": [
        {
            "id": 1,
            "identify": 2024061211,
            "course": "Math",
            "hours": 54
        },
        {
            "id": 2,
            "identify": 2024061211,
            "course": "Physics",
            "hours": 32
        },
        {
            "course": "English",
            "hours": 68,
            "id": 3,
            "identify": 2024070733
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
        "article": "Compilation theory, architecture of large systems, and development of Reduced Instruction Set (RISC) computers",
        "identify": 2024061211,
        "score": 89.7,
        "image": "SU1BR0UuUENH",
        "name": "caohao",
        "exam_time": "15:30:00",
        "birthday": "1995-03-23T00:00:00+08:00",
        "created_at": "2024-11-30T20:53:57+08:00",
        "updated_at": "2024-12-12T19:30:37+08:00",
        "id": 1,
        "gender": 1,
        "age": 19
    },
    "score_rank": 2
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
        "args": [2024061211]
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
                "id": 1,
                "identify": 2024061211,
                "gender": 1,
                "age": 19,
                "name": "caohao",
                "score": 89.7,
                "image": "SU1BR0UuUENH",
                "article": "Compilation theory, architecture of large systems, and development of Reduced Instruction Set (RISC) computers",
                "exam_time": "15:30:00",
                "birthday": "1995-03-23T00:00:00+08:00",
                "created_at": "2024-11-30T20:53:57+08:00",
                "updated_at": "2024-12-12T19:30:37+08:00",
                "student_course": {
                    "data": [
                        {
                            "id": 1,
                            "identify": 2024061211,
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
                            "identify": 2024061211,
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
                "id": 2,
                "identify": 2024070733,
                "gender": 1,
                "age": 17,
                "name": "jerry",
                "score": 92.3,
                "image": "SU1BR0UuUENH",
                "article": "Design and analysis of algorithms and data structures",
                "exam_time": "14:30:00",
                "birthday": "1993-02-22T00:00:00+08:00",
                "created_at": "2024-11-30T20:57:03+08:00",
                "updated_at": "2024-12-12T20:41:00+08:00",
                "student_course": {
                    "data": [
                        {
                            "id": 3,
                            "identify": 2024070733,
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


# 测试
#### 1.2.1.1 分片、分库、分表
如果有分库、分表、分片等需求，客户端可以通过 shard 字段，告诉服务端我的分片参数，然后在服务端配置分片函数 ShardFunc 来实现复杂的分库、分表、分片逻辑。

```json
[
    {
        "name":"student",
        "op":"find",
        "shard":"student_72",
        "where":{
            "userid":32881772
        }
    }
]
```

### 1.2.2 数据查询
#### 1.2.2.1多表 join（MySQL 特有）
注意，join 和 嵌套查询的区别，join返回结果在同一行，嵌套查询，其他表是在被嵌套的一个字段存在。
```json
[
    {
        "name":"student(s)",
        "op":"find_all",
        "join":[
            {   //JOIN `employee
                "join":"identity" 
            },
            {   //LEFT JOIN `employee` AS `ea` USING (`name`)
                "left":"identity(ea)",  
                "using":["name"]
            },
            {   //RIGHT JOIN `employee` AS `eb` USING (`name`,`age`)
                "right":"identity(eb)",
                "using":["name","age"]
            },
            {   //INNER JOIN `employee` AS `ec` ON `u`.`age`=`ec`.`age` AND `eb`.`sex` =`ec`.`sex` #字段自带别名的情况下，用别名
                "inner":"identity(ec)",
                "on":{
                    "age":"age",
                    "eb.sex":"sex"
                }
            },
            {   //FULL JOIN `employee` ON `u`.`age`=`employee`.`age`
                "full":"identity",
                "on":{
                    "age":"age"
                }
            }
        ],
        "where":{
            "sex":"male"
        }
    }
]
```

#### 1.2.2.2 指定返回字段
通过 column 参数指定需要返回的字段。
- **字段数组**
* 请求
```json
[
    {
        "name":"student",
        "op":"find_all",
		"column":["userid", "name", "age"],
        "where":{
            "status !":0,
            "sex":"male",
            "age >":15
        }
    }
]
```

* 返回：
```json
{
    "student":[
        {
            "userid":32346,
            "age":18,
            "name":"smallhowcao"
        },
        {
            "userid":43216,
            "age":22,
            "name":"jack"
        }
    ]
}
```
- **SELECT 方式（MySQL 特有）**
* 请求
```json
[
    {
        "name":"student",
        "op":"find_all",
        "column":["sex, age, count(id) as cnt"],
        "where":{
            "status !":0,
            "sex":"male",
            "age >":15
        },
        "group":["sex","age"]
    }
]
```

* 返回：
```json
{
    "student":[
        {
            "sex":"male",
            "age":18,
            "cnt":32
        },
        {
            "sex":"male",
            "age":19,
            "cnt":12
        },
        {
            "sex":"female",
            "age":17,
            "cnt":9
        },
        {
            "sex":"female",
            "age":18,
            "cnt":27
        }
    ]
}
```

#### 1.2.2.3 WHERE 条件
##### 1.2.2.3.1 操作符
WHERE 条件一共支持 11 种操作符：
```golang
const (
	OPEqual          = "="  // OPEqual 等于
	OPBetween        = "()" // OPBetween 在某个区间
	OPNotBetween     = "><" // OPNotBetween 不在某个区间
	OPGt             = ">"  // OPGt 大于
	OPGte            = ">=" // OPGte 大于等于
	OPLt             = "<"  // OPLt 小于
	OPLte            = "<=" // OPLte 小于等于
	OPNot            = "!"  // OPNot 去反
	OPLike           = "~"  // OPLike like语句，（或 es 的近似匹配）
	OPNotLike        = "!~" // OPNotLike not like 语句，（或 es 的近似匹配排除）
	OPMatchPhrase    = "?"  // OPMatchPhrase es 短语匹配 match_phrase
	OPNotMatchPhrase = "!?" // OPNotMatchPhrase es 短语匹配排除 must_not match_phrase
	OPMatch          = "*"  // OPMatch es 全文搜索 match 语句
	OPNotMatch       = "!*" // OPNotMatch es 全文搜索排除 must_not match
)
```
##### 1.2.2.3.2 基础用法
```json
[
    {
        "name":"student",
        "op":"find_all",
        "where":{
            "age":29,              //`age` = 29
            "age >":15,            //`age` > 15
            "age >=":15,           //`age` >= 15 
            "age !":30,            //`age` != 30
            "age ()":[20,29],      //`age` BETWEEN 20 AND 29
            "age ><":[35,40],      //NOT (`age` BETWEEN 35 AND 40)
            "score":[60,61,62],    //`score` IN (60, 61, 62)
            "score !":[70,71,72],  //`score` NOT IN (70, 71, 72)
            "name":null,           //`name` is NULL
            "name !":null          //`name` is NOT NULL
        }
    }
]
```
##### 1.2.2.3.3 组合查询
* 示例1： ``` `userid` > 3 OR `sex` = male OR `age` < 30```
```
[
    {
        "name":"student",
        "op":"find_all",
        "where":{
            "OR":{
                "userid >":3,
                "sex":"male",
                "age <":30
            }
        }
    }
]
```
* 示例2：``` ( `id` > 3  OR `sex` = "male" OR `age` < 30 ) AND `height` = 177 ```
```
[
    {
        "name":"student",
        "op":"find_all",
        "where":{
            "AND":{
                "OR":{
                    "userid >":3,
                    "sex":"male",
                    "age <":30
                },
                "height":177
            }
        }
    }
]
```
* 示例3：``` (`age` = 29 OR `sex` = 'female') AND (`uid` != 3 OR `height` >= 170)```
  注意：由于mysql使用map参数，所以在下面的情况下，第一个 OR 会被覆盖，所以需要加注释用于区分
```json
[
    {
        "name":"student",
        "op":"find_all",
        "where":{
            "AND":{
                "OR #注释1":{
                    "id >":3,
                    "sex":"male"
                },
                "OR #注释2":{
                    "uid !":3,
                    "height >=":170
                }
            }
        }
    }
]
```

* 实例4：``` NOT (`id` > 3 AND `sex` = 'male') ```
```json
[
    {
        "name":"student",
        "op":"find_all",
        "where":{
            "NOT":{
                "id >":3,
                "sex":"male"
            }
        }
    }
]
```
##### 1.2.2.3.4 LIKE 模糊匹配
`注意：在 elastic 中 LIKE 有些不同，详细可以看下一个章节`
* 案例1：
```json
[
    {
        "name":"student",
        "op":"find_all",
        "where":{
            "name ~":"%ide%",                       //`name` LIKE '%ide%'
            "addtime ~":["2019-08%","2020-01%"],    //( `addtime` LIKE '2019-08%' OR `addtime` LIKE '2020-01%')
            "name !~":"%ide%",                      //`name` NOT LIKE '%ide%'
            "addtime !~":["2019-08%","2020-01%"]    //( `addtime` NOT LIKE '2019-08%' AND `addtime` NOT LIKE '2020-01%')  ## 注意他和 LIKE 的连接词不一样，NOT LIKE 是 AND，而 LIKE 是 OR
        }
    }
]
```

* 案例 2
```json
[
    {
        "name":"student",
        "op":"find_all",
        "where":{
            "first_name ~":"Londo_",     // London, Londox, Londos...
            "second_name ~":"[BCR]at",   // Bat, Cat, Rat
            "last_name ~":"[!BCR]at"     // Eat, Fat, Hat...
        }
    }
]
```

##### 1.2.2.3.5 部分匹配（prefix、wildcard、regexp）（elastic 特有）
与MySQL的操作符 `~` 用于表示LIKE 不同，  在 es 中，`~` 表示部分匹配。
部分匹配分3中类型，prefix（默认）、wildcard、regexp
* prefix 前缀查询

* wildcard 通配符查询

* regexp 正则表达式查询

##### 1.2.2.3.6 短语匹配 match_phrase（elastic 特有）

##### 1.2.2.3.7 全文搜索 match（elastic 特有）

#### 1.2.2.4 分组/聚合 GROUP（elastic 的聚合暂未支持）
##### 1.2.2.4.1 GROUP BY
```json
[
    {
        "name":"student",
        "op":"find_all",
        "column":"sex, age, count(id) as cnt",
        "where":{
            "status !":0,
            "age >":15
        },
        "group":["sex","age"]
    }
]
```
##### 1.2.2.4.2 HAVING
HAVING 参数与 WHERE 语法一致。
```json
[
    {
        "name":"student",
        "op":"find_all",
        "column":"sex, age,max(class_id) as mc, count(id) as cnt",
        "where":{
            "status !":0,
            "age >":15
        },
        "group":[
            "sex",
            "age"
        ],
        "having":{
            "mc >":2,
            "cnt >":10
        }
    }
]
```
#### 1.2.2.5 排序与分页
##### 1.2.2.5.1 ORDER 排序
asc 表示升序、desc 表示降序

```json
[
    {
        "name":"student",
        "op":"find_all",
        "where":{
            "status !":0,
            "age >":15
        },
        "order":["height asc", "age desc"]
    }
]
```

##### 1.2.2.5.2 LIMIT、OFFSET
如果用户如果不希望设置 limit，请将 limit 设置为 0，否则在不 set limit 的情况下，统一接入系统为了保护数据库会将 limit 设为默认的 50。
```json
[
    {
        "name":"student",
        "op":"find_all",
        "where":{
            "status !":0,
            "age >":15
        },
        "limit":30,
        "offset":0
    }
]
```

##### 1.2.2.5.3 分页
如果存在 page 和 pagesize 参数，则系统会忽略 limit 和 offset 参数，系统会通过 page 和 pagesize 计算对应的 limit 和 offset，页码 page 从 1 开始。
* 请求
```
[
    {
        "name":"student",
        "op":"find_all",
        "where":{
            "status !":0,
            "age >":15
        },
        "page":1,
        "page_size":20
    }
]
```
* 返回
  返回包含数据总数 total、总分页 total_page、当前页数page、页大小page_size，以及以及数据 data。
```
{
    "student":{
        "total":2,
        "total_page":1,
        "page":1,
        "page_size":20,
        "data":[
            {
                "userid":32346,
                "sex":"male",
                "age":18,
                "name":"smallhowcao",
                "status":1
            },
            {
                "userid":43216,
                "sex":"male",
                "age":22,
                "name":"jack",
                "status":1
            }
        ]
    }
}
```


#### 1.2.2.6 相关性评分（elastic 特有）


#### 1.2.2.7 返回结果高亮（elastic 特有）
```
[
    {
        "name":"student",
        "op":"find_all",
        "where":{
            "content *":"北京市长视察冬奥会"
        },
        "page":1,
        "page_size":20,
        "highlight":{
            "fields":["content"],
            "pre_tags":"<span class='highlight-text'>",
            "post_tags":"</span>"
        }
    }
]
```

### 1.2.3 数据维护
#### 1.2.3.1 新增数据

* 请求
```json
[
    {
        "name":"student",
        "op":"insert",
        "data":{
            "sex":"male",
            "age":31,
            "name":"smallhowcao",
            "status":1
        }
    },
    {
        "name":"subject",
        "op":"insert",
        "data":[
            {
                "subject":"语文",
                "teacher":"刘老师"
            },
            {
                "subject":"数学",
                "teacher":"张老师"
            },
            {
                "subject":"英语",
                "teacher":"曹老师"
            }
        ]
    }
]
```

* 返回：
```json
{
    "student":{
        "last_insert_id":32346,
        "rows_affected":1
    },
    "subject":{
        "last_insert_id":1,
        "rows_affected":3
    }
}
```

#### 1.2.3.2 替换数据

* 请求
```json
[
    {
        "name":"student",
        "op":"replace",
        "data":{
            "id":32346,
            "sex":"male",
            "age":31,
            "name":"smallhowcao",
            "status":1
        }
    }
]
```

* 返回：
```json
{
    "student":{
        "rows_affected":1
    }
}
```


#### 1.2.3.3 更新数据

* 请求
```json
[
    {
        "name":"student",
        "op":"update",
        "where":{
            "userid":32346
        },
        "data":{
            "sex":"male",
            "age":31,
            "name":"smallhowcao",
            "status":1
        }
    }
]
```

* 返回：
```json
{
    "student":{
        "rows_affected":1
    }
}
```

#### 1.2.3.4 删除数据

* 请求
```json
[
    {
        "name":"student",
        "op":"delete",
        "where":{
            "userid":32346
        }
    }
]
```

查询 userid = 32346 的学生。
* 返回：
```json
{
    "student":{
        "rows_affected":1
    }
}
```

### 1.2.4 redis 协议
#### 1.2.4.1 基础用法
统一接入协议支持 redis 协议的 NoSQL 数据库操作。下面解释下字段含义：
`name`：指定执行名，服务端以便能获取对应的kv数据库连接信息。
`op`： command 操作命令。
`prefix`：key 的前缀，所有数据存储在库中真实的 real_key = pre + key，这个主要是用于数据的统计，比如说你有上亿用户的资料存储，key=SKYNET_USER_INFO:$userid，如果我们有设置 pre，便可以方便的在统一接入服务统计出用户资料相关的操作并发量，我们建议所有的 redis key 都加上 prefix，如果 key 是一个固定值，例如一个有序集，那么可以只填 prefix 参数，key置空，那么 real_key=prefix。
`key`：键 key
`args`：参数

redis 协议的执行单元不支持嵌套子查询。

* 请求
```json
[
    {
        "name":"skynet(incr_user_age)",
        "op":"HINCRBY",
		"prefix":"SKYNET_USER_INFO:",
		"key":"176294921",
        "args":["score", 15]
    }
]
```

* 返回：
```json
{
    "incr_user_age":85
}
```

#### 1.2.4.2 返回类型
返回类型根据 op 操作决定。然后在客户端，用户可以根据接收类型，做解码。
协议不支持阻塞操作 `BLPOP`、`BRPOP`

##### 1.2.4.2.1 无返回，仅包含 error
这类操作包含`EXPIRE`、`SET`、 `SETEX`、`HSET`、`HMSET`，仅返回 error，如果无 error 则执行成功，参考[执行单元异常](#242-执行单元异常)
示例：
* 请求
```json
[
    {
        "name":"skynet",
        "op":"SET",
		"prefix":"SKYNET_USER_INFO:",
		"key":"176294921",
        "args":["{\"age\":32,\"height\":178,\"version\":1}"]
    }
]
```

* 返回：
```json
{
}
```

##### 1.2.4.2.2 返回 []byte
这类操作包含`GET`、`GETSET`、`HGET`、`LPOP`、`RPOP`

示例：
* 请求
```json
[
    {
        "name":"skynet(user_info)",
        "op":"GET",
		"prefix":"SKYNET_USER_INFO:",
		"key":"176294921",
        "args":[]
    }
]
```

* 返回：
```json
{
    "user_info":"{\"age\":32,\"height\":178,\"version\":1}"
}
```

##### 1.2.4.2.3 返回 bool
这类操作包含`EXISTS`、`SETNX`、`HEXISTS`、`HSETNX`、`SISMEMBER`

示例：
* 请求
```json
[
    {
        "name":"skynet(user_exists)",
        "op":"EXISTS",
		"prefix":"SKYNET_USER_INFO:",
		"key":"176294921",
        "args":[]
    }
]
```

* 返回：
```json
{
    "user_exists":true
}
```

##### 1.2.4.2.4 返回 int64
这类操作包含`INCR`、`DECR`、`INCRBY`、`HINCRBY`、`TTL`、`DEL`、`HDEL`、`HLEN`、`LPUSH`、`RPUSH`、`LLEN`、`SADD`、`SREM`、`SCARD`、`SMOVE`、`ZADD`、`ZREM`、`ZREMRANGEBYSCORE`、`ZREMRANGEBYRANK`、`ZCARD`、`ZRANK`、`ZREVRANK`、`ZCOUNT`

示例：
* 请求
```json
[
    {
        "name":"skynet(incr_user_version)",
        "op":"HINCRBY",
		"prefix":"SKYNET_USER_INFO:",
		"key":"176294921",
        "args":["version", 2]
    }
]
```

* 返回：
```json
{
    "incr_user_version":3
}
```


##### 1.2.4.2.5 返回 float64
这类操作包含`ZSCORE`、`ZINCRBY`

示例：
* 请求
```json
[
    {
        "name":"skynet(user_score)",
        "op":"ZSCORE",
		"prefix":"SKYNET_USER_SCORE",
		"key":"",
        "args":["176294921"]
    }
]
```

* 返回：
```json
{
    "user_score":32.5
}
```

##### 1.2.4.2.6 返回 [][]byte
这类操作包含`HKEYS`、`SMEMBERS`、`SRANDMEMBER`、`SPOP`、`ZRANGE`、`ZRANGEBYSCORE`、`ZREVRANGE`、`ZREVRANGEBYSCORE`

示例：
* 请求
```json
[
    {
        "name":"skynet(version_users)",
        "op":"HKEYS",
		"prefix":"SKYNET_USER_VERSION",
		"key":"",
        "args":[]
    }
]
```

* 返回：
```json
{
    "version_users":[
        "18234234",
        "42343884",
        "82978392",
        "83821124"
    ]
}
```

##### 1.2.4.2.7 返回  map[string]string
这类操作包含`HGETALL`、`HMGET`

示例：
* 请求
```json
[
    {
        "name":"skynet(user_info)",
        "op":"HGETALL",
		"prefix":"SKYNET_USER_INFO:",
		"key":"176294921",
        "args":[]
    }
]
```

* 返回：
```json
{
    "user_info":{
        "age":"32",
        "score":"56.5",
        "height":"173"
    }
}
```


##### 1.2.4.2.9 返回 member 和 score（类型为 [][]byte、[]float64）
这类操作包含`ZPOPMIN`、`ZPOPMAX`、`ZRANGE ... WITHSCORES`、`ZRANGEBYSCORE ... WITHSCORES`、`ZREVRANGE ... WITHSCORES`、`ZREVRANGEBYSCORE ... WITHSCORES`

示例：
* 请求
```json
[
    {
        "name":"skynet(user_score_range)",
        "op":"ZRANGE",
		"with_scores":true,
		"prefix":"SKYNET_USER_SCORE",
		"key":"",
        "args":[1, 4, "WITHSCORES"]
    }
]
```

* 返回：
```json
{
    "user_score_range":{
        "member":[
            "18234234",
            "42343884",
            "82978392",
            "83821124"
        ],
        "score":[
            75,
            68,
            93,
            83
        ]
    }
}
```

### 1.2.5 引用
引用是一个非常有用的概念，通常我们需要查询多个结果，后面的结果需要用到前面的结果返回的时候，我们需要做多次查询，并对数据进行组合，然后在统一接入协议中，我们可以通过引用去解决问题。
对于 myql 、es，如果字段 key 以 `@` 开头或者结尾，则为引用，两者的概念不一样，以 `@`结尾的为`单字段引用`， `@` 开头的为`联合引用`。
对于 redis，暂不支持。

#### 1.2.5.1 引用原则
引用大原则：
1、`引用限制`：由于执行单元的执行顺序问题，引用只能`向上向前`引用，即被引用单元必须是引用单元的`同一层级的前序节点`，或是其`父节点、爷爷节点、一直到 head 节点`，比如下面的例子，student 可以引用 class 的返回字段，因为他们处于同一层级，class不可以引用 student 的返回字段，student 在 class 执行完成之后才会执行，student 也不可以引用 leader 的返回字段。

2、`绝对路径和相对路径`：如果被引用的执行单元与引用的执行单元处于同一个层级，也就是同一个数组下，那么可以用相对路径表示，比如下面的 student 的 class_id 引用自 class 的 id 字段。而 leader 的 class_id 是属于 class 的嵌套子查询，所以需要以 `/` 符号作为开头来表示绝对路径。

3、`引用逻辑`：首先执行被引用的执行单元，并将返回结果的被引用字段提取，然后赋值给引用执行单元。用以执行。

4、`空引用`：如果被引用的返回结果为空，则引用者的结果也为空。

#### 1.2.5.2 单字段引用
单字段引用 `@` 在字段后面。
示例代码：
```json
[
    {
        "name":"class",
        "op":"find_all",
        "where":{
            "student_number >":30
        },
        "sub":[
            {
                "name":"leader",
                "op":"find_all",
                "with":{
                    "class_id":"id"
                },
                "where":{
                    "class_id@":"/class.id"
                }
            }
        ]
    },
    {
        "name":"student",
        "op":"find_all",
        "where":{
            "status !":0,
            "sex":"male",
            "age >":15,
            "class_id@":"class.id"
        },
        "sub":[
            {
                "name":"teachers",
                "op":"find_all",
                "with":{
                    "userid":"userid"
                },
                "where":{
                    "class_id@":"class.id"
                }
            }
        ]
    }
]
```

#### 1.2.5.3 联合引用
联合引用以 `@` 开头，后面跟一个执行单元名称，值为 map，表示被引用单元返回的每条结果的这些字段都需要被满足。如下示例，表示先查出学生数量大于30的班级，假设记录为有两条记录，`班级id=1，班主任 adviser="刘老师"`和`班级id=2，班主任 adviser="刘老师"`，那么第二个查询单元，查询条件则为 WHERE age=19 AND ((class_id=1 and teacher="刘老师") OR (class_id=2 and teacher="张老师"))
```json
[
    {
        "name":"class",
        "op":"find_all",
        "where":{
            "student_number >":30
        }
    },
    {
        "name":"student",
        "op":"find_all",
        "where":{
            "age":19,
            "@/class":{
                "class_id":"id",
                "teacher":"adviser"
            }
        }
    }
]
```

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




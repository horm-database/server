[toc]
# 一 数据统一接入协议
统一接入协议是接入数据统一接入服务而设计的一套包含增删改查等等一系列操作的数据访问协议。接入层采用同一套协议，可以操作统一存储中心的多种存储引擎，如：mysql、clickhouse、postgres等类 sql 协议引擎，还有 redis、BDB、Tendis 等类 redis 协议引擎，另外还支持  elastic search 引擎。<br>
在数据统一接入服务，我们可以将数据的采集、更新加工、展示、监控流向一站式的把控。
协议由一组`执行单元`组成，根据功能他们可以细分为 `变更单元` 或 `查询单元` ，每个执行单元都是对一张表或者一个es 索引的一个操作，包含增删改查等操作。执行单元之间可以是并行执行的关系，还可以存在字段引用的关系、或者嵌套关系，比如一个查询单元的 WHERE 条件中某个字段引用自另一个查询单元的查询结果，那么就需要等被引用的查询单元执行完成之后，再去执行该查询单元。

## 1.1 查询模式
统一接入协议一共包含3种查询模式，单执行单元，并行执行，嵌套查询。
```json
[
    {
        "name":"student",
        "op":"find",
        "where":{
            "userid":32346
        }
    },
    {
        "name":"grade",
        "op":"find_all",
        "where":{
            "score >":90
        }
    }
]
```

### 1.1.1 单执行单元
整个查询仅包含一个执行单元。
#### 1.1.1.1 查询单条记录
* 请求
```json
{
    "name":"student",
    "op":"find",
    "where":{
        "userid":32346
    }
}
```
查询 userid = 32346 的学生。
* 返回：
```json
{
    "userid":32346,
    "sex":"male",
    "age":18,
    "name":"smallhowcao",
    "status":1
}
```

#### 1.1.1.2 查询多条记录
* 请求
```json
{
    "name":"student",
    "op":"find_all",
    "where":{
        "status !":0,
        "sex":"male",
        "age >":15
    }
}
```

查询 status!=0 and sex='male' and age>15 的所有学生。
* 返回：
```json
[
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
```

### 1.1.2 并行执行
并行查询是指的多个执行单元在同一次请求中返回，不同执行单元的结果会放在一个 map[string]interface{} 返回，map 的 key 是执行单元的 name 或别名。
从理论上来说，两个执行单元是应该并行执行的，除了 2 种特殊情况：
- 一种是引用，一个执行单元的执行条件引用自另一个执行单元的结果，那么他必须等到被引用执行单元完成之后才可以执行。
- 另外一种是带有 wait 关键词的，表示我需要等待指定的执行单元完成之后才能执行。

* 请求
```json
[
    {
        "name":"student",
        "op":"find",
        "where":{
            "userid":32346
        }
    },
    {
        "name":"grade",
        "op":"find_all",
        "where":{
            "score >":90
        }
    }
]
```

查询 userid = 32346 的学生和分数大于90分的信息。
* 返回：
```json
{
    "student":{
        "userid":32346,
        "sex":"male",
        "age":31,
        "name":"smallhowcao",
        "status":1
    },
    "grade":[
        {
            "userid":32346,
            "subject_id":1,
            "score":99
        },
        {
            "userid":32346,
            "subject_id":2,
            "score":98
        },
        {
            "userid":32346,
            "subject_id":3,
            "score":93
        }
    ]
}
```

* 引用使用示例如下：grade 的查询条件 userid 引用自 student 的结果。所以，grade 需要等待 student 查询完成之后才可以查询，在使用应用的时候必须注意，被引用执行单元必须插在引用执行单元的前面。
```json
[
    {
        "name":"student",
        "op":"find_all",
        "where":{
            "age >":15
        }
    },
    {
        "name":"grade",
        "op":"find_all",
        "where":{
            "userid@":"student.userid",
            "score >":90
        }
    }
]
```

* wait 关键字使用示例如下：我们需要将新增学生 和 科目的插入完成之后，才执行查询动作，查出最新的学生。
```json
[
    {
        "name":"student(insert_student)",
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
            }
        ]
    },
    {
        "name":"student",
        "wait":["insert_student","subject"],
        "op":"find_all"
    }
]
```

### 1.1.3 嵌套查询
嵌套查询，是指的我们查出父查询的结果之后，会到子查询根据引用标识符 `@` 自动识别将查到的结果嵌套于哪个父查询 ：
* 请求
```json
[
    {
        "name":"student",
        "op":"find_all",
        "where":{
            "sex":"male",
            "age >":15
        },
        "sub":[
            {
                "name":"grade",
                "op":"find_all",
                "where":{
                    "score >":90,
					"@userid":"/student.userid",
					"@status":"/student.status"
                },
                "sub":[
                    {
                        "name":"subject",
                        "op":"find",
                        "where":{
                            "@id":"/student/grade.subject_id"
                        }
                    }
                ]
            }
        ]
    }
]
```

* 返回：
```json
{
    "student":[
        {
            "userid":371293,
            "class_id":1,
            "sex":"male",
            "age":31,
            "name":"smallhowcao",
            "status":1,
            "grade":[
                {
                    "userid":371293,
                    "status":1,
                    "subject_id":1,
                    "subject":{
                        "id":1,
                        "subject":"语文",
                        "teacher":"刘老师"
                    },
                    "score":99
                },
                {
                    "userid":371293,
                    "status":1,
                    "subject_id":2,
                    "subject":{
                        "id":2,
                        "subject":"数学",
                        "teacher":"张老师"
                    },
                    "score":98
                },
                {
                    "userid":371293,
                    "subject_id":3,
                    "subject":{
                        "id":3,
                        "subject":"英语",
                        "teacher":"曹老师"
                    },
                    "score":93
                }
            ]
        },
        {
            "userid":874312,
            "class_id":2,
            "sex":"male",
            "age":31,
            "name":"smallhowcao",
            "status":2,
            "grade":[
                {
                    "userid":874312,
                    "status":2,
                    "subject_id":1,
                    "subject":{
                        "id":1,
                        "subject":"语文",
                        "teacher":"刘老师"
                    },
                    "score":99
                },
                {
                    "userid":874312,
                    "status":2,
                    "subject_id":2,
                    "subject":{
                        "id":2,
                        "subject":"数学",
                        "teacher":"张老师"
                    },
                    "score":98
                },
                {
                    "userid":874312,
                    "status":2,
                    "subject_id":3,
                    "subject":{
                        "id":3,
                        "subject":"英语",
                        "teacher":"曹老师"
                    },
                    "score":93
                }
            ]
        }
    ]
}
```

## 1.2 执行单元
`执行单元`是协议的最小数据执行单位，根据功能他们可以细分为 `变更单元` 或 `查询单元` ，每个执行单元都是对一张表或者一个es 索引的一个操作，包含增删改查等操作。执行单元之间可以是并行执行的关系，还可以存在字段引用的关系、或者嵌套关系，比如一个查询单元的 WHERE 条件中某个字段引用自另一个查询单元的查询结果，那么就需要等被引用的查询单元执行完成之后，再去执行该查询单元。<br>

### 1.2.1 执行单元名与别名
每个`执行单元`都拥有一个 name，同一个数组层级下执行单元名字不能相同，因为返回结果的格式为`map[name]interface`，name 相同会导致后面执行单元的结果覆盖前面执行单元，如果两个执行单元操作了相同的表，name 必须相同，可以使用别名，返回结果的 key 会取自别名，如下：
* 请求
```json
[
    {
        "name":"student",
        "op":"find",
        "where":{
            "userid":32881772
        }
    },
    {
        "name":"student(olderthan13)",
        "op":"find_all",
        "where":{
            "age >":13
        }
    },
    {
        "name":"grade",
        "op":"find_all",
        "where":{
            "userid@":"olderthan13.userid"
        }
    }
]
```

* 返回
```
{
    "student":{
        "userid":32881772,
        "class_id":1,
        "sex":"male",
        "age":31,
        "name":"smallhowcao",
        "status":1
    },
    "olderthan13":[
        {
            "userid":37122393,
            "class_id":1,
            "sex":"male",
            "age":19,
            "name":"mark",
            "status":1
        },
        {
            "userid":87438832,
            "class_id":2,
            "sex":"male",
            "age":23,
            "name":"kidy",
            "status":2
        }
    ],
    "grade":[
        {
            "userid":37122393,
            "status":1,
            "subject_id":1,
            "score":99
        },
        {
            "userid":37122393,
            "status":1,
            "subject_id":2,
            "score":98
        },
        {
            "userid":87438832,
            "status":2,
            "subject_id":1,
            "score":89
        },
        {
            "userid":87438832,
            "status":2,
            "subject_id":2,
            "score":91
        }
    ]
}
```

每个 name 对应服务器后端一个 存储引擎单元，他可以是 mysql 的表、es 的索引等等，我们可以通过 name 去配置里面找到对应的库、表、索引等信息，然后做对应的操作。

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




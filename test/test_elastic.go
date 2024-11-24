package test

import (
	"context"
	"fmt"
	"time"

	"github.com/horm/server/srv/codec"

	"github.com/horm/common/proto"
	"github.com/horm/go-horm/horm"
)

// TestElastic 测试 elastic
func TestElastic() {
	hasChildQueryWithInnerHits()
}

func insert0() {
	ctx, _ := context.WithTimeout(codec.GCtx, 600*time.Second)

	var ret proto.ModResult

	data := horm.Map{
		"_id":        1,
		"article_id": 1,
		"title":      "推进美丽中国建设",
	}

	_, err := horm.NewQuery("es_test").
		Type("_doc").
		Replace(data).
		Exec(ctx, &ret)

	fmt.Println(err)
}

func insert1() {
	ctx, _ := context.WithTimeout(codec.GCtx, 600*time.Second)

	var ret []proto.ModResult

	datas := []horm.Map{
		{
			"_id":        1,
			"article_id": 1,
			"title":      "推进美丽中国建设",
		},
		{
			"_id":        2,
			"article_id": 2,
			"title":      "TFBOYS演唱会门票开售秒空",
		},
	}

	_, err := horm.NewQuery("es_test").
		Type("_doc").
		Insert(datas).
		Exec(ctx, &ret)

	fmt.Println(err)
}

func insert2() {
	ctx, _ := context.WithTimeout(codec.GCtx, 600*time.Second)

	var ret []proto.ModResult

	mapping := `
PUT es_test
{
  "mappings": {
    "properties": {
      "comments": {
        "type": "nested",
        "properties": {
          "votes": {
            "type": "nested"
          }
        }
      }
    }
  }
}
`
	fmt.Println(mapping)

	datas := []horm.Map{
		{
			"_id":        1,
			"article_id": 1,
			"user":       "Sarah",
			"title":      "高效推进美丽中国建设",
			"status":     1,
			"comments": []map[string]interface{}{
				{"user": "Olivia", "say": "共建清洁美丽新世界", "prise": 111, "votes": []map[string]interface{}{{"vote": 1, "level": 1}, {"vote": -1, "level": 3}}},
				{"user": "Jacob", "say": "共同实现全人类的地球梦", "prise": 66, "votes": []map[string]interface{}{{"vote": 1, "level": 5}, {"vote": 1, "level": 1}}},
				{"user": "Jackson", "say": "建地球生命共同体", "prise": 55, "votes": []map[string]interface{}{{"vote": -1, "level": 2}, {"vote": -1, "level": 1}}},
			},
		},
		{
			"_id":        2,
			"article_id": 2,
			"user":       "Emily",
			"title":      "TFBOYS演唱会门票开售秒空",
			"status":     1,
			"comments": []map[string]interface{}{
				{"user": "Olivia", "say": "可能是为了看王源小朋友吧", "prise": 17, "votes": []map[string]interface{}{{"vote": 1, "level": 2}, {"vote": 1, "level": 3}}},
				{"user": "Aiden", "say": "支持他们去美丽国开演唱会赚外汇", "prise": 222, "votes": []map[string]interface{}{{"vote": 1, "level": 1}, {"vote": -1, "level": 2}}},
				{"user": "Ethan", "say": "希望能实名购买演唱会门票", "prise": 111, "votes": []map[string]interface{}{{"vote": 1, "level": 3}, {"vote": 1, "level": 1}}},
				{"user": "Aiden", "say": "赚钱效率可真高啊", "prise": 29, "votes": []map[string]interface{}{{"vote": 1, "level": 4}, {"vote": 1, "level": 2}}},
			},
		},
		{
			"_id":        3,
			"article_id": 3,
			"user":       "Emily",
			"title":      "建设一个更强大的中国，北京在实践",
			"status":     1,
			"comments": []map[string]interface{}{
				{"user": "Ethan", "say": "我觉得非常可以", "prise": 55, "votes": []map[string]interface{}{{"vote": 1, "level": 3}, {"vote": -1, "level": 5}}},
				{"user": "Aiden", "say": "共同实现人类命运共同体", "prise": 0, "votes": []map[string]interface{}{{"vote": 1, "level": 1}, {"vote": -1, "level": 1}}},
				{"user": "Aiden", "say": "先辈们的梦想，在我们这一代实现", "prise": 55, "votes": []map[string]interface{}{{"vote": 1, "level": 5}, {"vote": -1, "level": 2}}},
			},
		},
		{
			"_id":        4,
			"article_id": 4,
			"user":       "Emily",
			"title":      "王一博也要在北京开演唱会了",
			"status":     1,
			"comments": []map[string]interface{}{
				{"user": "Lucas", "say": "演唱会门票不太好买吧", "prise": 133, "votes": []map[string]interface{}{{"vote": -1, "level": 2}, {"vote": -1, "level": 3}}},
				{"user": "Ethan", "say": "我的梦想就是去看一场他的演唱会", "prise": 66, "votes": []map[string]interface{}{{"vote": 1, "level": 4}, {"vote": -1, "level": 3}}},
			},
		},
		{
			"_id":        5,
			"article_id": 5,
			"user":       "Sarah",
			"title":      "台风每小时超10公里冲向广东",
			"status":     0,
			"comments": []map[string]interface{}{
				{"user": "Olivia", "say": "沿海的朋友保护自己", "prise": 66, "votes": []map[string]interface{}{{"vote": 1, "level": 1}, {"vote": -1, "level": 4}}},
				{"user": "Aiden", "say": "出门注意安全，能不出门尽量不出门", "prise": 222, "votes": []map[string]interface{}{{"vote": -1, "level": 1}, {"vote": -1, "level": 1}}},
			},
		},
		{
			"_id":        6,
			"article_id": 6,
			"user":       "Sarah",
			"title":      "ChatGPT能否高效取代人工",
			"status":     0,
			"comments": []map[string]interface{}{
				{"user": "Olivia", "say": "ChatGPT 目前还没有这样的能力", "prise": 55, "votes": []map[string]interface{}{{"vote": -1, "level": 1}, {"vote": -1, "level": 1}}},
				{"user": "Jackson", "say": "短期看不可能，目前效率不一定比人工高", "prise": 39, "votes": []map[string]interface{}{{"vote": -1, "level": 1}, {"vote": -1, "level": 1}}},
			},
		},
	}

	_, err := horm.NewQuery("es_test").
		Type("_doc").
		Replace(datas).
		Exec(ctx, &ret)

	fmt.Println(err)
}

func nested() {
	ctx, _ := context.WithTimeout(codec.GCtx, 600*time.Second)

	where := horm.Where{}
	where["status"] = 1
	where["nested #点赞==55"] = horm.Where{
		"path":  "comments",
		"query": horm.Where{"comments.prise": 55},
	}

	var detail proto.Detail
	var ret []map[string]interface{}

	_, err := horm.NewQuery("es_test").
		Type("_doc").
		FindAll(where).
		Page(1, 10).
		Exec(ctx, &detail, &ret)

	fmt.Println(err)
}

func nestedWithEmptyInnerHits() {
	ctx, _ := context.WithTimeout(codec.GCtx, 600*time.Second)

	var ret []map[string]interface{}

	where := horm.Where{}
	where["status"] = 1
	where["nested"] = horm.Where{
		"path":       "comments",
		"query":      horm.Where{"comments.prise": 55},
		"inner_hits": map[string]interface{}{},
	}

	_, err := horm.NewQuery("es_test").
		Type("_doc").
		FindAll(where).
		Exec(ctx, &ret)

	fmt.Println(err)
}

func nestedWithInnerHits() {
	ctx, _ := context.WithTimeout(codec.GCtx, 600*time.Second)

	var ret []map[string]interface{}

	where := horm.Where{}
	where["status"] = 1
	where["nested"] = horm.Where{
		"path":  "comments",
		"query": horm.Where{"comments.prise": 55},
		"inner_hits": horm.Where{
			"name":   "comments", // 不填默认取 nested/path 值
			"from":   0,
			"size":   10,
			"order":  []string{"comments.prise"},
			"column": []string{"comments.user", "comments.say"},
		},
	}

	_, err := horm.NewQuery("es_test").
		Type("_doc").
		FindAll(where).
		Exec(ctx, &ret)

	fmt.Println(err)
}

func nestedWithInnerHits2() {
	ctx, _ := context.WithTimeout(codec.GCtx, 600*time.Second)

	var ret []map[string]interface{}

	where := horm.Where{}
	where["status"] = 1
	where["nested"] = horm.Where{
		"path":       "comments.votes",
		"query":      horm.Where{"comments.votes.level": 5},
		"inner_hits": horm.Where{},
	}

	_, err := horm.NewQuery("es_test").
		Type("_doc").
		FindAll(where).
		Exec(ctx, &ret)

	fmt.Println(err)
}

func insert3() {
	ctx, _ := context.WithTimeout(codec.GCtx, 600*time.Second)

	var ret []proto.ModResult

	mapping := `
PUT /es_test?pretty
{
    "settings":{
        "number_of_shards":3,
        "number_of_replicas":1
    },
    "mappings":{
        "properties":{
            "creator":{
              "type": "keyword"
            },
            "plan_join":{
                "type":"join",
                "relations":{
                    "plan":[
                        "activity",
                        "book"
                    ],
                    "book":"comments"
                }
            }
        }
    }
}`

	fmt.Println(mapping)

	// 添加父文档数据
	// 1、如果是创建父文档，则需要使用 plan_join 指定父文档的关系的名字(此处为plan)。
	// 2、plan_join为创建索引的 mapping时指定join的字段的名字。
	// 3、指定父文档时,plan_join的这2种写法都可以。
	datas := []horm.Map{
		{
			"_id":         "plan-001",
			"plan_id":     "plan-001",
			"plan_name":   "四月计划",
			"create_time": "2021-04-07 16:27:30",
			"creator":     "Emily",
			"plan_join": horm.Map{
				"name": "plan",
			},
		},
		{
			"_id":         "plan-002",
			"plan_id":     "plan-002",
			"plan_name":   "五月计划",
			"create_time": "2021-05-07 16:27:30",
			"creator":     "Olivia",
			"plan_join":   "plan",
		},
	}

	_, err := horm.NewQuery("es_test").
		Type("_doc").
		Replace(datas).
		Exec(ctx, &ret)

	fmt.Println(err)

	// 添加子文档
	// 1、子文档（子孙文档等）需要和父文档使用相同的路由键。
	// 2、需要指定父文档的id。
	// 3、需要指定join的名字。
	datas = []horm.Map{
		{
			"_id":      "act-001",
			"act_id":   "act-001",
			"act_name": "四月第一个活动",
			"creator":  "Emily",
			"plan_join": horm.Map{
				"name":   "activity",
				"parent": "plan-001",
			},
		},
		{
			"_id":       "book-001",
			"book_id":   "book-001",
			"book_name": "四月读取的第一本书",
			"creator":   "Olivia",
			"plan_join": horm.Map{
				"name":   "book",
				"parent": "plan-001",
			},
		},
		{
			"_id":       "book-002",
			"book_id":   "book-002",
			"book_name": "编程珠玑",
			"creator":   "Jackson",
			"plan_join": horm.Map{
				"name":   "book",
				"parent": "plan-002",
			},
		},
		{
			"_id":       "book-003",
			"book_id":   "book-003",
			"book_name": "java编程思想",
			"creator":   "Olivia",
			"plan_join": horm.Map{
				"name":   "book",
				"parent": "plan-001",
			},
		},
		{
			"_id":          "comment-001",
			"comment_id":   "comment-001",
			"comment_name": "这本书还可以",
			"creator":      "Aiden",
			"plan_join": horm.Map{
				"name":   "comments",
				"parent": "book-001",
			},
		},
		{
			"_id":          "comment-002",
			"comment_id":   "comment-002",
			"comment_name": "值得一读，棒。",
			"creator":      "Aiden",
			"plan_join": horm.Map{
				"name":   "comments",
				"parent": "book-002",
			},
		},
	}

	_, err = horm.NewQuery("es_test").
		Type("_doc").
		Replace(datas).
		Routing("plan-001").
		Exec(ctx, &ret)

	fmt.Println(err)
}

func parentIdQuery() {
	ctx, _ := context.WithTimeout(codec.GCtx, 600*time.Second)

	var ret []map[string]interface{}

	// 返回父文档 id 是 plan-001 下的类型为book的所有子文档
	where := horm.Where{}
	where["parent_id"] = horm.Where{
		"type": "book",
		"id":   "plan-001",
	}

	_, err := horm.NewQuery("es_test").
		Type("_doc").
		FindAll(where).
		Exec(ctx, &ret)

	fmt.Println(err)
}

func hasChildQuery() {
	ctx, _ := context.WithTimeout(codec.GCtx, 600*time.Second)

	ret := []map[string]interface{}{}

	// 返回创建者(creator)是huan.fu,并且子文档最少有2个的父文档。
	where := horm.Where{}
	where["has_child"] = horm.Where{
		"type":         "book",
		"min_children": "2",
		"query":        horm.Where{"creator": "Olivia"},
	}

	isNil, err := horm.NewQuery("es_test").
		Type("_doc").
		FindAll(where).
		Exec(ctx, &ret)

	fmt.Println(isNil, err)
}

func hasChildQueryWithInnerHits() {
	ctx, _ := context.WithTimeout(codec.GCtx, 600*time.Second)

	var ret []map[string]interface{}

	// 返回创建者(creator)是huan.fu,并且子文档最少有2个的父文档。
	where := horm.Where{}
	where["has_child"] = horm.Where{
		"type":         "book",
		"min_children": "2",
		"query":        horm.Where{"creator": "Olivia"},
		"inner_hits":   horm.Where{},
	}

	_, err := horm.NewQuery("es_test").
		Type("_doc").
		FindAll(where).
		WithTraceID("xxxxsfa").WithRequestID(12345).
		Exec(ctx, &ret)

	fmt.Println(err)
}

func hasParentQuery() {
	ctx, _ := context.WithTimeout(codec.GCtx, 600*time.Second)

	var ret []map[string]interface{}

	// 返回创建者(creator)是huan.fu,并且子文档最少有2个的父文档。
	where := horm.Where{}
	where["has_parent"] = horm.Where{
		"parent_type": "book",
		"query":       horm.Where{"creator": "Jackson"},
	}

	_, err := horm.NewQuery("es_test").
		Type("_doc").
		FindAll(where).
		Exec(ctx, &ret)

	fmt.Println(err)
}

func hasParentQueryWithInnerHits() {
	ctx, _ := context.WithTimeout(codec.GCtx, 600*time.Second)

	var ret []map[string]interface{}

	// 返回创建者(creator)是huan.fu,并且子文档最少有2个的父文档。
	where := horm.Where{}
	where["has_parent"] = horm.Where{
		"parent_type": "book",
		"query":       horm.Where{"creator": "Jackson"},
		"inner_hits":  horm.Where{},
	}

	_, err := horm.NewQuery("es_test").
		Type("_doc").
		FindAll(where).
		Exec(ctx, &ret)

	fmt.Println(err)
}

func collapse() {
	ctx, _ := context.WithTimeout(codec.GCtx, 600*time.Second)

	var ret []map[string]interface{}

	where := horm.Where{}
	where["prise >"] = 1

	_, err := horm.NewQuery("es_test").
		Type("_doc").
		Collapse("user.keyword").
		FindAll(where).
		Exec(ctx, &ret)

	fmt.Println(err)
}

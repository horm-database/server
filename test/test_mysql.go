package test

import (
	"context"
	"fmt"
	"time"

	"github.com/horm/common/proto"
	"github.com/horm/go-horm/horm"
	"github.com/horm/server/srv/codec"
)

// TestMySQL 测试 mysql
func TestMySQL() {
	insertMysql()
}

func insertMysql() {
	ctx, _ := context.WithTimeout(codec.GCtx, 600*time.Second)

	var ret proto.ModResult

	data := horm.Map{
		"_id":        1,
		"article_id": 1,
		"title":      "推进美丽中国建设",
	}

	_, err := horm.NewQuery("student").
		Replace(data).
		Exec(ctx, &ret)

	fmt.Println(err)
}

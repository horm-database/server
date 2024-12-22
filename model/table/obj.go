// Copyright (c) 2024 The horm-database Authors (such as CaoHao <18500482693@163.com>). All rights reserved.
//
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
package table

import (
	"time"

	"github.com/horm-database/server/plugin/conf"
)

type TblWorkspace struct {
	Id          int       `orm:"id,int,omitempty" json:"id,omitempty"`
	Workspace   string    `orm:"workspace,string" json:"workspace,omitempty"`       // workspace
	Name        string    `orm:"name,string" json:"name,omitempty"`                 // 空间名
	Intro       string    `orm:"intro,string" json:"intro,omitempty"`               // 简介
	Company     string    `orm:"company,string" json:"company,omitempty"`           // 公司
	Department  string    `orm:"department,string" json:"department,omitempty"`     // 部门
	Token       string    `orm:"token,string" json:"token,omitempty"`               // token
	EnforceSign int8      `orm:"enforce_sign,int8" json:"enforce_sign,omitempty"`   // 是否强制签名 0-否 1-是（请求数据必须得签名或者加密）
	Creator     uint64    `orm:"creator,uint64,omitempty" json:"creator,omitempty"` // Creator
	Manager     string    `orm:"manager,string" json:"manager,omitempty"`           // 管理员，多个逗号分隔
	CreatedAt   time.Time `orm:"created_at,datetime,omitempty" json:"created_at"`   // 记录创建时间
	UpdatedAt   time.Time `orm:"updated_at,datetime,omitempty" json:"updated_at"`   // 记录最后修改时间
}

type TblAppInfo struct {
	Appid     uint64    `orm:"appid,uint64" json:"appid"`                         // 应用appid
	Name      string    `orm:"name,string" json:"name"`                           // 应用名称
	Secret    string    `orm:"secret,string" json:"secret"`                       // 应用秘钥
	Intro     string    `orm:"intro,string" json:"intro"`                         // 简介
	Creator   uint64    `orm:"creator,uint64,omitempty" json:"creator,omitempty"` // Creator
	Manager   string    `orm:"manager,string" json:"manager"`                     // 管理员，多个逗号分隔
	Status    int8      `orm:"status,int8" json:"status"`                         // 1-正常 2-下线
	CreatedAt time.Time `orm:"created_at,datetime,omitempty" json:"created_at"`   // 记录创建时间
	UpdatedAt time.Time `orm:"updated_at,datetime,omitempty" json:"updated_at"`   // 记录最后修改时间
}

type TblAccessDB struct {
	Id        int       `orm:"id,int,omitempty" json:"id"`
	Appid     uint64    `orm:"appid,uint64" json:"appid"`                               // 应用appid
	DB        int       `orm:"db,int" json:"db"`                                        // 数据库id
	Root      int8      `orm:"root,int8" json:"root"`                                   // 超级权限 1-超级权限（所有权限，包含DDL）  2-表数据权限（库下表的所有增删改查权限，不包含 DDL）  3-无
	Op        string    `orm:"op,string" json:"op"`                                     // 支持的操作
	Status    int8      `orm:"status,int8" json:"status"`                               // 状态：1-正常 2-下线 3-审核中 4-审核撤回 5-拒绝
	ApplyUser uint64    `orm:"apply_user,uint64,omitempty" json:"apply_user,omitempty"` // 申请者
	Reason    string    `orm:"reason,string" json:"reason"`                             // 接入原因
	CreatedAt time.Time `orm:"created_at,datetime,omitempty" json:"created_at"`         // 记录创建时间
	UpdatedAt time.Time `orm:"updated_at,datetime,omitempty" json:"updated_at"`         // 记录最后修改时间
}

type TblAccessTable struct {
	Id        int       `orm:"id,int,omitempty" json:"id"`
	Appid     uint64    `orm:"appid,uint64" json:"appid"`                               // 应用appid
	TableId   int       `orm:"table_id,int" json:"table_id"`                            // 表id
	QueryAll  int8      `orm:"query_all,int8" json:"query_all"`                         // 是否支持所有的 query 语句，1-true 2-false
	Op        string    `orm:"op,string" json:"op"`                                     // 支持的表操作
	Status    int8      `orm:"status,int8" json:"status"`                               // 状态：1-正常 2-下线 3-审核中 4-审核撤回 5-拒绝
	ApplyUser uint64    `orm:"apply_user,uint64,omitempty" json:"apply_user,omitempty"` // 申请者
	Reason    string    `orm:"reason,string" json:"reason"`                             // 接入原因
	CreatedAt time.Time `orm:"created_at,datetime,omitempty" json:"created_at"`         // 记录创建时间
	UpdatedAt time.Time `orm:"updated_at,datetime,omitempty" json:"updated_at"`         // 记录最后修改时间
}

type TblPlugin struct {
	Id           int       `orm:"id,int,omitempty" json:"id"`
	Name         string    `orm:"name,string" json:"name"`                           // 插件名称
	Intro        string    `orm:"intro,string" json:"intro"`                         // 中文简介
	Version      string    `orm:"version,string" json:"version"`                     // 所有支持的插件版本，逗号分开
	Func         string    `orm:"func,string" json:"func"`                           // 插件注册函数名
	SupportTypes string    `orm:"support_types,string" json:"support_types"`         // 支持的插件类型 1-前置插件 2-后置插件 3-defer 插件，多个逗号分隔，空串为全部支持
	Online       int8      `orm:"online,int8" json:"online"`                         // 状态 1-上线 2-下线
	Source       int8      `orm:"source,int8" json:"source"`                         // 来源：1-官方插件 2-第三方插件 3-个人插件
	Desc         string    `orm:"desc,string" json:"desc"`                           // 详细介绍
	Creator      uint64    `orm:"creator,uint64,omitempty" json:"creator,omitempty"` // Creator
	Manager      string    `orm:"manager,string" json:"manager,omitempty"`           // 管理员，多个逗号分隔
	CreatedAt    time.Time `orm:"created_at,datetime,omitempty" json:"created_at"`   // 记录创建时间
	UpdatedAt    time.Time `orm:"updated_at,datetime,omitempty" json:"updated_at"`   // 记录最后修改时间
}

type TblPluginConfig struct {
	Id            int       `orm:"id,int,omitempty" json:"id"`
	PluginID      int       `orm:"plugin_id,int" json:"plugin_id"`                  // 插件id
	PluginVersion int       `orm:"plugin_version,int" json:"plugin_version"`        // 插件版本
	Key           string    `orm:"key,string" json:"key"`                           // 插件配置 key
	Name          string    `orm:"name,string" json:"name"`                         // 插件配置名
	Type          int8      `orm:"type,int8" json:"type"`                           // 配置类型 1-bool、2-string、3-int、4-uint、5-float、6-枚举 7-时间、8-array、9-map、10-multi-conf
	NotNull       int8      `orm:"not_null,int8" json:"not_null"`                   // 是否必输 1-是 2-否
	MoreInfo      string    `orm:"more_info,string" json:"more_info"`               // 更多细节
	Default       string    `orm:"default,int8" json:"default"`                     // 默认值，仅用于预填充配置值。
	Desc          string    `orm:"desc,string" json:"desc"`                         // 配置描述
	CreatedAt     time.Time `orm:"created_at,datetime,omitempty" json:"created_at"` // 记录创建时间
	UpdatedAt     time.Time `orm:"updated_at,datetime,omitempty" json:"updated_at"` // 记录最后修改时间
}

type TblTablePlugin struct {
	Id             int       `orm:"id,int,omitempty" json:"id"`
	TableId        int       `orm:"table_id,int" json:"table_id"`                    // 表id
	PluginID       int       `orm:"plugin_id,int" json:"plugin_id"`                  // 插件id
	PluginVersion  int       `orm:"plugin_version,int" json:"plugin_version"`        // 插件版本
	Type           int8      `orm:"type,int8" json:"type"`                           // 插件类型 1-前置插件 2-后置插件 3-defer 插件
	Front          int       `orm:"front,int" json:"front"`                          // plugin execute front of me
	ScheduleConfig string    `orm:"schedule_config,string" json:"schedule_config"`   // 插件调度配置，是一个json，内容是 map[string]interface{}
	Config         string    `orm:"config,string" json:"config"`                     // 插件配置，是一个json，内容是 map[string]interface{}
	Desc           string    `orm:"desc,string" json:"desc"`                         // 描述
	Status         int8      `orm:"status,int8" json:"status"`                       // 状态 1-启用 2-停用
	CreatedAt      time.Time `orm:"created_at,datetime,omitempty" json:"created_at"` // 记录创建时间
	UpdatedAt      time.Time `orm:"updated_at,datetime,omitempty" json:"updated_at"` // 记录最后修改时间

	ScheduleConf *conf.ScheduleConfig // 调度规则
	Conf         conf.PluginConfig    // 解析后的配置
}

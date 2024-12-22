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

package srv

import (
	"io/ioutil"
	"sync/atomic"
	"time"

	"github.com/horm-database/common/log/logger"
	"github.com/horm-database/server/srv/naming"
	"gopkg.in/yaml.v3"
)

const (
	confFile           = "./server.yaml"
	defaultIdleTimeout = 60000 // 单位 ms
	maxCloseWaitTime   = 10 * time.Second
)

// config 配置
type config struct {
	Env       string `yaml:"env"`        // 环境
	Machine   string `yaml:"machine"`    // 机器名（容器名）
	MachineID int    `yaml:"machine_id"` // 机器编号（容器编号）（主要用于 snowflake 生成全局唯一 id）
	LocalIP   string `yaml:"local_ip"`   // 本地 ip

	Server struct {
		Name             string `yaml:"name"`                // 服务名
		CloseWaitTime    int    `yaml:"close_wait_time"`     // 注销名字服务之后的等待时间，让名字服务更新实例列表。 (单位 ms) 默认: 0ms, 最大: 10s.
		MaxCloseWaitTime int    `yaml:"max_close_wait_time"` // 进程结束之前等待请求完成的最大等待时间。(单位 ms)
		RpcPort          uint16 `yaml:"rpc_port"`            // rpc 监听端口
		HttpPort         uint16 `yaml:"http_port"`           // http 监听端口
		WebPort          uint16 `yaml:"web_port"`            // web 监听端口
		Timeout          int    `yaml:"timeout"`             // 服务超时时间(单位 ms)
		IdleTime         int    `yaml:"idle_time"`           // 连接最大空闲时间，默认为 1 分钟。(单位 ms)
		EventLoopNum     int    `yaml:"event_loop_num"`      // gnet loop 大小，默认取 CPU 核数
		TLSKey           string `yaml:"tls_key"`             // tls key
		TLSCert          string `yaml:"tls_cert"`            // tls cert
		CACert           string `yaml:"ca_cert"`             // ca cert
	}

	Log []*logger.Config `yaml:"log"`

	// Register 北极星服务治理
	Register *naming.Config `yaml:"register"`
}

var globalConfig atomic.Value // 服务端配置

// Config returns the common Config.
func Config() *config {
	return globalConfig.Load().(*config)
}

// loadConfig 加载配置文件
func loadConfig(configPath string) (*config, error) {
	cfg, err := parseConfigFile(configPath)
	if err != nil {
		return nil, err
	}

	cfg.Server.IdleTime = defaultIdleTimeout

	globalConfig.Store(cfg)

	return cfg, nil
}

func parseConfigFile(configPath string) (*config, error) {
	buf, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	cfg := &config{}
	if err := yaml.Unmarshal(buf, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

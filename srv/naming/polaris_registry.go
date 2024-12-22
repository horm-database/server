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

package naming

import (
	"fmt"
	"net"
	"strconv"
	"sync"

	"github.com/polarismesh/polaris-go"
	"github.com/polarismesh/polaris-go/pkg/model"
)

const (
	defaultWeight = 100
	defaultTTL    = 5
)

// PolarisRegistry 服务注册
type PolarisRegistry struct {
	Provider polaris.ProviderAPI
	config   *ServiceConfig
	host     string
	port     int
}

// ServiceConfig 配置
type ServiceConfig struct {
	Config      *Config
	Weight      int               // Weight
	InstanceID  string            // InstanceID
	Protocol    string            // 协议
	Namespace   string            // 命名空间
	ServiceName string            // 服务名
	BindAddress string            // 指定上报地址
	Metadata    map[string]string // 用户自定义 metadata 信息
}

// Config 配置
type Config struct {
	Enable   bool      `yaml:"enable"`
	Version  string    `yaml:"version"`
	TTL      int       `yaml:"ttl"`   // 服务端检查周期实例是否健康的周期，单位s
	Token    string    `yaml:"token"` // token
	Debug    bool      `yaml:"debug"`
	Location *Location `yaml:"location"`
}

type Location struct {
	Region string `yaml:"region"`
	Zone   string `yaml:"zone"`
	Campus string `yaml:"compus"`
}

func Add(protocol, serviceName, address string, cfg *Config) (Registry, error) {
	provider, err := polaris.NewProviderAPI()
	if err != nil {
		return nil, err
	}

	config := &ServiceConfig{
		Protocol:    protocol,
		Namespace:   "workspace",
		ServiceName: serviceName,
		BindAddress: address,
		Config:      cfg,
	}

	if config.Config.TTL == 0 {
		config.Config.TTL = defaultTTL
	}

	if config.Weight == 0 {
		config.Weight = defaultWeight
	}

	reg := &PolarisRegistry{
		Provider: provider,
		config:   config,
	}

	Register(serviceName, reg)

	return reg, nil
}

// Register 注册服务
func (r *PolarisRegistry) Register(_ string, address string) error {
	if !r.config.Config.Enable {
		return nil
	}

	if r.config.BindAddress != "" {
		address = parseHostPort(r.config.BindAddress)
	}

	host, port, _ := net.SplitHostPort(address)
	r.host = host
	r.port, _ = strconv.Atoi(port)

	if err := r.register(); err != nil {
		return err
	}

	return nil
}

// Deregister 反注册
func (r *PolarisRegistry) Deregister(_ string) error {
	if !r.config.Config.Enable {
		return nil
	}

	req := &polaris.InstanceDeRegisterRequest{
		InstanceDeRegisterRequest: model.InstanceDeRegisterRequest{
			Service:      r.config.ServiceName,
			ServiceToken: r.config.Config.Token,
			Namespace:    r.config.Namespace,
			InstanceID:   r.config.InstanceID,
			Host:         r.host,
			Port:         r.port,
		},
	}

	if err := r.Provider.Deregister(req); err != nil {
		return fmt.Errorf("deregister error: %s", err.Error())
	}
	return nil
}

func (r *PolarisRegistry) register() error {
	registerRequest := &polaris.InstanceRegisterRequest{}
	registerRequest.Service = r.config.ServiceName
	registerRequest.ServiceToken = r.config.Config.Token
	registerRequest.Namespace = r.config.Namespace
	registerRequest.Host = r.host
	registerRequest.Port = r.port
	registerRequest.Protocol = &r.config.Protocol
	registerRequest.Weight = &r.config.Weight
	registerRequest.Version = &r.config.Config.Version
	registerRequest.TTL = &r.config.Config.TTL

	if r.config.Config.Location != nil {
		registerRequest.Location = &model.Location{
			Region: r.config.Config.Location.Region,
			Zone:   r.config.Config.Location.Zone,
			Campus: r.config.Config.Location.Campus,
		}
	}

	resp, err := r.Provider.RegisterInstance(registerRequest)
	if err != nil {
		return fmt.Errorf("fail to Register instance, err is %v", err)
	}

	//plog.GetBaseLogger().Debugf("success to register instance1, id is %s\n", resp.InstanceID)
	r.config.InstanceID = resp.InstanceID
	return nil
}

// parseHostPort 解析地址
func parseHostPort(address string) string {
	host, port, err := net.SplitHostPort(address)
	if err != nil {
		return address
	}

	ip := net.ParseIP(host)
	if ip != nil {
		return address
	}

	// host 不是 ip
	parsedIP := getIP(host)
	return net.JoinHostPort(parsedIP, port)
}

func getIP(nic string) string {
	ip := localIP.getIPByNic(nic)
	return ip
}

type netInterfaceIP struct {
	once sync.Once
	ips  map[string]*nicIP
}

func (p *netInterfaceIP) enumAllIP() map[string]*nicIP {
	p.once.Do(func() {
		p.ips = make(map[string]*nicIP)
		interfaces, err := net.Interfaces()
		if err != nil {
			return
		}
		for _, i := range interfaces {
			p.addInterface(i)
		}
	})
	return p.ips
}

func (p *netInterfaceIP) addInterface(i net.Interface) {
	addrs, err := i.Addrs()
	if err != nil {
		return
	}
	for _, v := range addrs {
		ipNet, ok := v.(*net.IPNet)
		if !ok {
			continue
		}
		if ipNet.IP.To4() != nil {
			p.addIPv4(i.Name, ipNet.IP.String())
		} else if ipNet.IP.To16() != nil {
			p.addIPv6(i.Name, ipNet.IP.String())
		}
	}
}

func (p *netInterfaceIP) addIPv4(nic string, ip4 string) {
	ips := p.getNicIP(nic)
	ips.ipv4 = append(ips.ipv4, ip4)
}

func (p *netInterfaceIP) addIPv6(nic string, ip6 string) {
	ips := p.getNicIP(nic)
	ips.ipv6 = append(ips.ipv6, ip6)
}

func (p *netInterfaceIP) getNicIP(nic string) *nicIP {
	if _, ok := p.ips[nic]; !ok {
		p.ips[nic] = &nicIP{nic: nic}
	}
	return p.ips[nic]
}

func (p *netInterfaceIP) getIPByNic(nic string) string {
	p.enumAllIP()
	if len(p.ips) <= 0 {
		return ""
	}
	if _, ok := p.ips[nic]; !ok {
		return ""
	}
	ip := p.ips[nic]
	if len(ip.ipv4) > 0 {
		return ip.ipv4[0]
	}
	if len(ip.ipv6) > 0 {
		return ip.ipv6[0]
	}
	return ""
}

// localIP records the local nic name->nicIP mapping.
var localIP = &netInterfaceIP{}

// nicIP defines the parameters used to record the ip address (ipv4 & ipv6) of the nic.
type nicIP struct {
	nic  string
	ipv4 []string
	ipv6 []string
}

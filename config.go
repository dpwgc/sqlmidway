package main

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
	"time"
)

type ConfigOptions struct {
	Server ServerOptions `yaml:"server"`
	Log    LogOptions    `yaml:"log"`
	DB     DBOptions     `yaml:"db"`
	APIs   []APIOptions  `yaml:"apis"`
}

type ServerOptions struct {
	Addr     string `yaml:"addr"`
	Port     int    `yaml:"port"`
	TLS      bool   `yaml:"tls"`
	CertFile string `yaml:"cert-file"`
	KeyFile  string `yaml:"key-file"`
}

type LogOptions struct {
	Path       string `yaml:"path"`
	MaxSize    int    `yaml:"max-size"`
	MaxAge     int    `yaml:"max-age"`
	MaxBackups int    `yaml:"max-backups"`
}

type DBOptions struct {
	Type string `yaml:"type"`
	DSN  string `yaml:"dsn"`
}

type APIOptions struct {
	Service    string `yaml:"service"`
	Method     string `yaml:"method"`
	Sql        string `yaml:"sql"`
	Params     string `yaml:"params"`
	LowerCamel bool   `yaml:"lower-camel"`
	UpperCamel bool   `yaml:"upper-camel"`
	Underscore bool   `yaml:"underscore"`
	Debug      bool   `yaml:"debug"`
}

var config ConfigOptions

func Config() ConfigOptions {
	return config
}

func InitConfig() {
	//加载客户端配置
	configBytes, err := os.ReadFile("./config.yaml")
	if err != nil {
		fmt.Println("config error:", err)
		time.Sleep(3 * time.Second)
		panic(err)
	}
	err = yaml.Unmarshal(configBytes, &config)
	if err != nil {
		fmt.Println("config error:", err)
		time.Sleep(3 * time.Second)
		panic(err)
	}
}

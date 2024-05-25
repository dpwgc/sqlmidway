package main

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
	"regexp"
	"strings"
	"time"
)

type ConfigOptions struct {
	Server ServerOptions `yaml:"server"`
	Log    LogOptions    `yaml:"log"`
	DBs    []DBOptions   `yaml:"dbs"`
}

type ServerOptions struct {
	Addr     string `yaml:"addr"`
	Port     int    `yaml:"port"`
	TLS      bool   `yaml:"tls"`
	CertFile string `yaml:"cert-file"`
	KeyFile  string `yaml:"key-file"`
}

type LogOptions struct {
	Path    string `yaml:"path"`
	Size    int    `yaml:"size"`
	Age     int    `yaml:"age"`
	Backups int    `yaml:"backups"`
}

type DBOptions struct {
	Name string       `yaml:"name"`
	Type string       `yaml:"type"`
	DSN  string       `yaml:"dsn"`
	APIs []APIOptions `yaml:"apis"`
}

type APIOptions struct {
	Service    string `yaml:"service"`
	Method     string `yaml:"method"`
	Sql        string `yaml:"sql"`
	LowerCamel bool   `yaml:"lower-camel"`
	UpperCamel bool   `yaml:"upper-camel"`
	Underscore bool   `yaml:"underscore"`
	Debug      bool   `yaml:"debug"`
	Params     []string
	ORM        *ORM
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
	for _, db := range Config().DBs {
		for i := 0; i < len(db.APIs); i++ {
			db.APIs[i].Params = matchParams(db.APIs[i].Sql)
		}
	}
}

func matchParams(src string) []string {

	reg := regexp.MustCompile("\\{(.*?)}")
	arr := reg.FindAllString(src, -1)

	var r []string
	m := make(map[string]bool)
	for _, v := range arr {
		if strings.Contains(v, "{#") || strings.Contains(v, "{/") {
			continue
		}
		v = strings.ReplaceAll(v, "{", "")
		v = strings.ReplaceAll(v, "}", "")
		if m[v] {
			continue
		}
		r = append(r, v)
		m[v] = true
	}
	for _, v := range arr {
		if !strings.Contains(v, "{#") && !strings.Contains(v, "{/") {
			continue
		}
		v = strings.ReplaceAll(v, "{#", "")
		v = strings.ReplaceAll(v, "{/", "")
		v = strings.ReplaceAll(v, "}", "")
		if m[v] {
			continue
		}
		r = append(r, v)
		m[v] = true
	}
	return r
}

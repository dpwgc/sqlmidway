package main

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
	"regexp"
	"strings"
	"time"
)

const (
	LowerCamel = "lowerCamel"
	UpperCamel = "upperCamel"
	Underscore = "underscore"
)

type ConfigOptions struct {
	Server ServerOptions `yaml:"server"`
	Log    LogOptions    `yaml:"log"`
	DBs    []DBOptions   `yaml:"dbs"`
}

type ServerOptions struct {
	Addr     string           `yaml:"addr"`
	Port     int              `yaml:"port"`
	Auth     bool             `yaml:"auth"`
	Accounts []AccountOptions `yaml:"accounts"`
	TLS      bool             `yaml:"tls"`
	CertFile string           `yaml:"cert-file"`
	KeyFile  string           `yaml:"key-file"`
	Debug    bool             `yaml:"debug"`
}

type AccountOptions struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

type LogOptions struct {
	Path    string `yaml:"path"`
	Size    int    `yaml:"size"`
	Age     int    `yaml:"age"`
	Backups int    `yaml:"backups"`
}

type DBOptions struct {
	Name   string         `yaml:"name"`
	Type   string         `yaml:"type"`
	DSN    string         `yaml:"dsn"`
	Format string         `yaml:"format"`
	Groups []GroupOptions `yaml:"groups"`
}

type GroupOptions struct {
	Name   string       `yaml:"name"`
	Format string       `yaml:"format"`
	APIs   []APIOptions `yaml:"apis"`
}

type APIOptions struct {
	Name   string `yaml:"name"`
	Sql    string `yaml:"sql"`
	Format string `yaml:"format"`
	Hide   []string
	Show   []string
	Params []string
	Store  *Store
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
		for i := 0; i < len(db.Groups); i++ {
			for j := 0; j < len(db.Groups[i].APIs); j++ {
				db.Groups[i].APIs[j].Params = matchParams(db.Groups[i].APIs[j].Sql)
				if len(db.Groups[i].APIs[j].Format) > 0 {
					continue
				}
				if len(db.Groups[i].Format) > 0 {
					db.Groups[i].APIs[j].Format = db.Groups[i].Format
					continue
				}
				db.Groups[i].APIs[j].Format = db.Format
			}
		}
	}
	if Config().Server.Debug {
		for _, db := range Config().DBs {
			fmt.Println("[DEBUG]", "> DB:", db.Name)
			for _, group := range db.Groups {
				fmt.Println("[DEBUG]", ">>> GROUP:", group.Name)
				for _, api := range group.APIs {
					fmt.Println("[DEBUG]", ">>>>> API:", api.Name)
					fmt.Println("[DEBUG]", ">>>>>>> URI:", fmt.Sprintf("/%s/%s/%s", db.Name, group.Name, api.Name))
					fmt.Println("[DEBUG]", ">>>>>>> SQL:", api.Sql)
					fmt.Println("[DEBUG]", ">>>>>>> PARAMS:", api.Params)
				}
			}
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

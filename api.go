package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/dpwgc/easierweb"
	"reflect"
	"regexp"
	"strings"
)

type API struct {
	route map[string]APIOptions
}

func NewAPI() (*API, error) {
	route := make(map[string]APIOptions)
	for _, db := range Config().DBs {
		orm, err := NewORM(db)
		if err != nil {
			return nil, err
		}
		for i := 0; i < len(db.APIs); i++ {
			db.APIs[i].ORM = orm
			route[fmt.Sprintf("%s.%s.%s", db.Name, db.APIs[i].Service, db.APIs[i].Method)] = db.APIs[i]
		}
	}
	return &API{
		route: route,
	}, nil
}

type CommandReply struct {
	Sql    string           `json:"sql,omitempty"`
	Args   []any            `json:"args,omitempty"`
	Result map[string]int64 `json:"result,omitempty"`
}

type QueryReply struct {
	Sql    string           `json:"sql,omitempty"`
	Args   []any            `json:"args,omitempty"`
	Result []map[string]any `json:"result,omitempty"`
}

type InfoReply struct {
	Result []map[string]any `json:"result,omitempty"`
}

type ErrorReply struct {
	Error string `json:"error,omitempty"`
}

func (a *API) Query(ctx *easierweb.Context, request map[string]any) (*QueryReply, error) {
	api := a.route[fmt.Sprintf("%s.%s.%s", ctx.Path.Get("db"), ctx.Path.Get("service"), ctx.Path.Get("method"))]
	if len(api.Service) == 0 || len(api.Method) == 0 {
		return nil, errors.New("not found")
	}
	sql := api.Sql
	sql, args := a.handleSql(sql, request, api.Params)
	rows, err := api.ORM.Query(context.Background(), api, sql, args...)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("%v / sql: %s / args: %v", err, sql, args))
	}
	reply := &QueryReply{
		Result: rows,
	}
	if api.Debug {
		reply.Sql = sql
		reply.Args = args
	}
	return reply, nil
}

func (a *API) Command(ctx *easierweb.Context, request map[string]any) (*CommandReply, error) {
	api := a.route[fmt.Sprintf("%s.%s.%s", ctx.Path.Get("db"), ctx.Path.Get("service"), ctx.Path.Get("method"))]
	if len(api.Service) == 0 || len(api.Method) == 0 {
		return nil, errors.New("not found")
	}
	sql := api.Sql
	sql, args := a.handleSql(sql, request, api.Params)
	res, err := api.ORM.Command(context.Background(), sql, args...)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("%v / sql: %s / args: %v", err, sql, args))
	}
	reply := &CommandReply{
		Result: res,
	}
	if api.Debug {
		reply.Sql = sql
		reply.Args = args
	}
	return reply, nil
}

func (a *API) Info(ctx *easierweb.Context) *InfoReply {
	reply := &InfoReply{}
	for _, db := range Config().DBs {
		for _, api := range db.APIs {
			reply.Result = append(reply.Result, map[string]any{
				"db":      db.Name,
				"service": api.Service,
				"method":  api.Method,
				"params":  api.Params,
			})
		}
	}
	return reply
}

func (a *API) Health(ctx *easierweb.Context) {

}

func (a *API) handleSql(sql string, request map[string]any, params []string) (string, []any) {
	var args []any
	for _, field := range params {
		value := request[field]
		if value == nil {
			sql = a.delStr(sql, fmt.Sprintf("{#%s}", field), fmt.Sprintf("{/%s}", field))
		}
	}
	for _, field := range params {
		value := request[field]
		if value == nil {
			continue
		}
		f := fmt.Sprintf("{%s}", field)
		if strings.Contains(sql, f) {
			if reflect.TypeOf(value).Kind() == reflect.Slice {
				var in []string
				arr := value.([]any)
				for _, item := range arr {
					in = append(in, "?")
					args = append(args, item)
				}
				sql = strings.ReplaceAll(sql, f, fmt.Sprintf("(%s)", strings.Join(in, ",")))
			} else {
				sql = strings.ReplaceAll(sql, f, "?")
				args = append(args, value)
			}
		}
		sql = strings.ReplaceAll(sql, fmt.Sprintf("{#%s}", field), "")
		sql = strings.ReplaceAll(sql, fmt.Sprintf("{/%s}", field), "")
	}
	return sql, args
}

func (a *API) delStr(s string, start string, end string) string {
	// 查找需要删除的部分
	startIndex := strings.Index(s, start)
	endIndex := strings.Index(s, end) + len(end)
	if startIndex == -1 || endIndex == -1 {
		return s
	}

	// 删除指定部分（包括起始和结束字符串）
	re := regexp.MustCompile(start + "(.*?)" + end)
	return re.ReplaceAllLiteralString(s, "")
}

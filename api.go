package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/dpwgc/easierweb"
	"regexp"
	"strings"
)

type API struct {
	orm   *ORM
	route map[string]APIOptions
}

func NewAPI() (*API, error) {
	orm, err := NewORM()
	if err != nil {
		return nil, err
	}
	route := make(map[string]APIOptions)
	for _, api := range Config().APIs {
		route[fmt.Sprintf("%s.%s", api.Service, api.Method)] = api
	}
	return &API{
		orm:   orm,
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

type ErrorReply struct {
	Error string `json:"error,omitempty"`
}

func (a *API) Query(ctx *easierweb.Context, request map[string]any) (*QueryReply, error) {
	api := a.route[fmt.Sprintf("%s.%s", ctx.Path.Get("service"), ctx.Path.Get("method"))]
	if len(api.Service) == 0 || len(api.Method) == 0 {
		return nil, errors.New("not found")
	}
	sql := api.Sql
	sql, args := a.handleSql(sql, request, strings.Split(strings.ReplaceAll(api.Params, " ", ""), ","))
	rows, err := a.orm.Query(context.Background(), api, sql, args...)
	if err != nil {
		return nil, err
	}
	if api.Debug {
		return &QueryReply{
			Sql:    sql,
			Args:   args,
			Result: rows,
		}, nil
	}
	return &QueryReply{
		Result: rows,
	}, nil
}

func (a *API) Command(ctx *easierweb.Context, request map[string]any) (*CommandReply, error) {
	api := a.route[fmt.Sprintf("%s.%s", ctx.Path.Get("service"), ctx.Path.Get("method"))]
	if len(api.Service) == 0 || len(api.Method) == 0 {
		return nil, errors.New("not found")
	}
	sql := api.Sql
	sql, args := a.handleSql(sql, request, strings.Split(strings.ReplaceAll(api.Params, " ", ""), ","))
	res, err := a.orm.Command(context.Background(), sql, args...)
	if err != nil {
		return nil, err
	}
	if api.Debug {
		return &CommandReply{
			Sql:    sql,
			Args:   args,
			Result: res,
		}, nil
	}
	return &CommandReply{
		Result: res,
	}, nil
}

func (a *API) Health(ctx *easierweb.Context) {

}

func (a *API) handleSql(sql string, request map[string]any, params []string) (string, []any) {
	var args []any
	for _, field := range params {
		value := request[field]
		if value == nil {
			sql = a.delStr(sql, fmt.Sprintf("{#%s}", field), fmt.Sprintf("{/%s}", field))
		} else {
			sql = strings.ReplaceAll(sql, fmt.Sprintf("{%s}", field), "?")
			args = append(args, value)
			sql = strings.ReplaceAll(sql, fmt.Sprintf("{#%s}", field), "")
			sql = strings.ReplaceAll(sql, fmt.Sprintf("{/%s}", field), "")
		}
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

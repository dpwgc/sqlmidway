package main

import (
	"context"
	"database/sql"
	"errors"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	_ "gorm.io/driver/clickhouse"
	_ "gorm.io/driver/mysql"
	_ "gorm.io/driver/postgres"
	_ "gorm.io/driver/sqlserver"
	"regexp"
	"strings"
)

type ORM struct {
	db *sql.DB
}

func NewORM() (*ORM, error) {
	db, err := sql.Open(Config().DB.Type, Config().DB.DSN)
	if err != nil {
		return nil, err
	}
	return &ORM{
		db: db,
	}, nil
}

func (c *ORM) Command(ctx context.Context, sql string, args ...any) (map[string]int64, error) {
	result, err := c.db.Exec(sql, args...)
	if err != nil {
		return nil, err
	}
	rowsAffected, _ := result.RowsAffected()
	lastInsertId, _ := result.LastInsertId()
	return map[string]int64{
		"rowsAffected": rowsAffected,
		"lastInsertId": lastInsertId,
	}, nil
}

func (c *ORM) Query(ctx context.Context, api APIOptions, sql string, args ...any) ([]map[string]any, error) {
	rows, err := c.db.Query(sql, args...)
	if err != nil {
		return nil, err
	}
	defer func() {
		err = rows.Close()
		if err != nil {
			Logger.Error(err.Error())
		}
	}()

	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	var newColumns []string

	for _, col := range columns {
		if api.LowerCamel {
			newColumns = append(newColumns, toCamelCase(col, false))
		} else if api.UpperCamel {
			newColumns = append(newColumns, toCamelCase(col, true))
		} else if api.Underscore {
			newColumns = append(newColumns, toUnderscore(col))
		} else {
			newColumns = append(newColumns, col)
		}
	}

	count := len(newColumns)

	result := make([]map[string]any, 0)
	values := make([]any, count)
	valPointers := make([]any, count)
	for rows.Next() {
		select {
		case <-ctx.Done():
			return nil, errors.New("context done")
		default:
		}
		for i := 0; i < count; i++ {
			valPointers[i] = &values[i]
		}
		err = rows.Scan(valPointers...)
		if err != nil {
			return nil, err
		}
		item := make(map[string]any)
		for i, col := range newColumns {
			var v any
			val := values[i]
			b, ok := val.([]byte)
			if ok {
				v = string(b)
			} else {
				v = val
			}
			item[col] = v
		}
		result = append(result, item)
	}
	return result, nil
}

// 将下划线分割的字符串转换为驼峰式
func toCamelCase(str string, upperFirst bool) string {
	// 首先将下划线替换为' '，然后使用Title函数将每个单词的首字母转为大写
	camel := cases.Title(language.Und, cases.NoLower).String(strings.ReplaceAll(str, "_", " "))
	// 如果需要小驼峰，则将第一个单词的首字母转为小写
	if !upperFirst {
		camel = strings.ToLower(camel[:1]) + camel[1:]
	}
	// 最后再将-替换为空字符串，得到最终的驼峰式字符串
	return strings.ReplaceAll(camel, " ", "")
}

// 将驼峰式转为下划线分隔的字符串
func toUnderscore(str string) string {
	// 首先将大写字母前面加上下划线，然后转为小写字母
	re := regexp.MustCompile(`([A-Z])`)
	underscore := strings.ToLower(re.ReplaceAllString(str, "_$1"))
	// 最后去掉第一个下划线，得到最终的下划线分隔字符串
	return strings.TrimPrefix(underscore, "_")
}

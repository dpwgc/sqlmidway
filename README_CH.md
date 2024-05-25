# SqlMidway

## 一个可根据自定义SQL模板配置自动实现HTTP增删改查接口的数据库网关服务，类似于Elasticsearch的查询模板功能。

***

### 一个简单的例子

* 修改config.yaml中的dbs参数，新建一个名为test_db的数据库连接，并为其配置一个查询接口

```yaml
# 数据库信息（可配置多个）
dbs:
  - name: test_db
    type: mysql
    dsn: root:123456@tcp(127.0.0.1:3306)/test_db?charset=utf8mb4&parseTime=True&loc=Local
    # 接口信息（可配置多个）
    apis:
      - service: test
        method: list
        # SQL模板（类似ES查询模板）
        sql: select * from test where 0=0 {#name} and name like {name} {/name} {#id} and id = {id} {/id} {#size} limit {size} {/size}
        # 返回的字段名转为小驼峰
        lower-camel: true
        # debug模式，响应结果时返回生成的SQL语句
        debug: true
```

* 访问这个HTTP接口（接口路径：/query/{db.name}/{api.service}/{api.method}）

> http://127.0.0.1:8899/query/test_db/test/list

* 请求

```json
{
    "name": "%测试%",
    "size": 10
}
```

* 响应

```json
{
	"sql": "select * from test where 0=0  and name like ?    limit ? ",
	"args": [
		"%测试%",
		10
	],
	"result": [
		{
			"createdAt": "2023-12-09T16:12:31+08:00",
			"id": 1,
			"name": "【测试】数据1",
			"status": 2,
			"tag": "test",
			"updatedAt": "2023-12-09T17:19:15+08:00"
		},
		{
			"createdAt": "2023-12-09T17:14:08+08:00",
			"id": 2,
			"name": "【测试】数据2",
			"status": 1,
			"tag": "test",
			"updatedAt": "2024-01-28T02:08:41+08:00"
		}
	]
}
```
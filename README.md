# SqlMidway

## A database gateway service that automatically implements HTTP CRUD API based on custom SQL template configurations. Similar to Elasticsearch search template.

***

### A simple example

* modify the dbs parameters in config.yaml to create a database connection named 'test_db' and configure a query API for the database

```yaml
# database information (multiple data sources can be configured)
dbs:
  - name: test_db
    type: mysql
    dsn: root:123456@tcp(127.0.0.1:3306)/test_db?charset=utf8mb4&parseTime=True&loc=Local
    # API information (multiple API can be configured)
    apis:
      - service: test
        method: list
        # sql template (similar to elasticsearch search template)
        sql: select * from test where 0=0 {#name} and name like {name} {/name} {#id} and id = {id} {/id} {#size} limit {size} {/size}
        # returned field name is changed to lower camel
        lower-camel: true
        # debug mode, which returns the generated SQL statement when responding to the result
        debug: true
```

* access this API (URI: /query/{db.name}/{api.service}/{api.method})

> http://127.0.0.1:8899/query/test_db/test/list

* request

```json
{
    "name": "%test%",
    "size": 10
}
```

* response

```json
{
	"sql": "select * from test where 0=0  and name like ?    limit ? ",
	"args": [
		"%test%",
		10
	],
	"result": [
		{
			"createdAt": "2023-12-09T16:12:31+08:00",
			"id": 1,
			"name": "[test]data1",
			"status": 2,
			"tag": "test",
			"updatedAt": "2023-12-09T17:19:15+08:00"
		},
		{
			"createdAt": "2023-12-09T17:14:08+08:00",
			"id": 2,
			"name": "[test]data2",
			"status": 1,
			"tag": "test",
			"updatedAt": "2024-01-28T02:08:41+08:00"
		}
	]
}
```

#### support 'in' query

* template

```yaml
sql: select * from test where 0=0 {#ids} and id in {ids} {/ids} limit 100
```

* request

```json
{
    "ids": [
      1,
      2
    ]
}
```
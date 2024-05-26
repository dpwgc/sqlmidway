# SqlMidway

### A database gateway that automatically implements HTTP CRUD API based on SQL template configurations. Similar to Elasticsearch search template.

***

## How to use

### Quick start

#### modify the dbs parameters in 'config.yaml' to create a database connection named 'testDB' and configure two APIs for the database

```yaml
# database information (multiple database can be configured)
dbs:

    # DB Name (make sure it is unique)
  - name: testDB
    type: mysql
    dsn: root:123456@tcp(127.0.0.1:3306)/test_db?charset=utf8mb4&parseTime=True&loc=Local

    # API Group information (multiple group can be configured)
    groups:
      
        # Group Name (make sure it is unique in the DB)
      - name: testGroup
        # returned field name is changed to lower camel (support: lowerCamel,upperCamel,underscore)
        format: lowerCamel
        # debug mode, which returns the generated SQL statement when responding to the result
        debug: true

        # API information (multiple API can be configured)
        apis:

            # API(1): /query/testDB/testGroup/listByIdOrName
            # API Name (make sure it is unique in the group)
          - name: listByIdOrName
            # sql template (similar to elasticsearch search template)
            sql: select * from test where 0=0 {#name} and name like {name} {/name} {#id} and id = {id} {/id} {#size} limit {size} {/size}

            # API(2): /query/testDB/testGroup/listByIds
          - name: listByIds
            sql: select * from test where id in {ids}
```

#### Explanation of SQL template rules

Similar to the Elasticsearch search template

* For example, template: 'select * from test where 0=0 {#name} and name like {name} {/name} {#id} and id = {id} {/id}'
* When the 'id' parameter is not specified, the statement '{#id} and id = {id} {/id}' will be eliminated at the time of execution
* When 'id' is passed in but 'name' is not passed, the SQL statement is generated as follows: 'select * from test where 0=0 and id = ?'

#### Start the project

* launch main.go

#### access this API(1) (URI: /query/{db.name}/{group.name}/{api.name})

> http://127.0.0.1:8899/query/testDB/testGroup/listByIdOrName

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

***

### Control field returns (like graphql)

#### Hide the 'id' and 'name' fields (URI with 'hide' parameter)

> http://127.0.0.1:8899/query/testDB/testGroup/listByIdOrName?hide=id,name

* request

```json
{
  "id": 1
}
```

* response

```json
{
  "result": [
    {
      "createdAt": "2023-12-09T16:12:31+08:00",
      "status": 2,
      "tag": "test",
      "updatedAt": "2023-12-09T17:19:15+08:00"
    }
  ]
}
```

#### Format the return field as upper camel, only the 'Id' and 'Name' fields are returned (URI with 'format' and 'shaw' parameters)

> http://127.0.0.1:8899/query/testDB/testGroup/listByIdOrName?format=upperCamel

* request

```json
{
  "id": 1
}
```

* response

```json
{
  "result": [
    {
      "Id": 1,
      "Name": "[test]data1"
    }
  ]
}
```

***

### Support 'in' query

#### pass in the array parameters

* sql template

```yaml
sql: select * from test where id in {ids}
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
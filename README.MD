# a extension sql library. enhance sql template/sql expression

## Features

- database factory: multiple database management support
- sql template: sql template support
- sql expression: sql expression support
- base mapper: basic sql mission support

## Installation

```bash
go get github.com/gnodux/sqlmx
```

## DBManager: management your database connections
always initialize DBManager instance named 'Manager'

```go
import "github.com/gnodux/sqlmx"
```
    
Open a database connection named 'db1' with mysql driver.
initialize sql template and sql expression with 'db1' connection.
```go   
db1,err:=sqlmx.Open("db1","mysql", "root:root@tcp(...)")
if err!=nil{
    log.Fatal(err)
}
db1.ParseTemplateFS(os.DirFS("./testdata"), "examples/*.sql", "initialize/*.sql", "my_mapper/*.sql")
```
## sql template

## sql expression

## base mapper

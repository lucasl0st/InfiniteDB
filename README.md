# InfiniteDB

A scalable database

## Server

### Installation

#### Docker

`docker run -v ./data:/var/lib/infinitedb -p 8080:8080 -d ghcr.io/lucasl0st/infinitedb:latest-unstable`

#### Docker Compose

A [docker-compose.yml](docker-compose.yml) file is in this repository, just run `docker-compose up -d`

#### Compiling from source

```
git clone https://github.com/lucasl0st/InfiniteDB.git

go build -o infinitedb-server

./infinitedb-server
```

### Environment variables

| Variable             | Description                                 | Default              |
|----------------------|---------------------------------------------|----------------------|
| DATABASE_PATH        | Path to database files                      | /var/lib/infinitedb/ |
| AUTHENTICATION       | Enables authentication                      | true                 |
| PORT                 | Database listen port                        | 8080                 |
| REQUEST_LOGGING      | Prints request logs to console              | false                |
| CACHE_SIZE           | Size of in-memory object cache              | 1000                 |
| TLS                  | Enables TLS                                 | false                |
| TLS_CERT             | Path to TLS Cert                            |                      |
| TLS_KEY              | Path to TLS Key                             |                      |
| WEBSOCKET_READ_LIMIT | Read limit of websocket connection in bytes | 10000000             |

## Client

### Installation

To use the go client install it using 
`go get github.com/lucasl0st/InfiniteDB/client`

### Usage

```go
package main

import (
	"fmt"
	InfiniteDB "github.com/lucasl0st/InfiniteDB/client"
	"log"
)

func main() {
	db := InfiniteDB.New(InfiniteDB.Options{
		Hostname: "localhost",
		Port:     8080,
	})

	err := db.Connect()

	if err != nil {
		log.Fatal(err)
	}

	r, err := db.GetDatabases()

	if err != nil {
		log.Fatal(err)
	}

	for _, databaseName := range r.Databases {
		fmt.Println(databaseName)
	}
}
```

### Queries

#### Request

```json
{
  "query": {},
  "sort": {},
  "implement": [],
  "skip": 50,
  "limit": 50
}
```

Query: [Query](#query)   
Sort: [Sort](#sort)   
Implement: array of [Implement](#implement)   
Skip: number   
Limit: number   


#### Query

```json
{
  "where": {},
  "functions": [],
  "and": {},
  "or": {}
}
```

Where: [Where](#where)   
Functions: array of [Function](#function)   
And: [Query](#query)   
Or: [Query](#query)   

#### Where

```json
{
  "field": "",
  "operator": "",
  "value": "",
  "all": [],
  "any": []
}
```

Field: string   
Operator: one of =, !=, >, < , ><   
Value: string, number or boolean   
All: array of string, number or boolean   
Any: array of string, number or boolean   

#### Function

```json
{
  "function": "",
  "parameters": {}
}
```

Function: string   
Parameters: object

#### Sort

```json
{
  "field": "",
  "direction": ""
}
```

Field: string   
Direction: asc or desc   

#### Implement

```json
{
  "from": {
    "table": "",
    "field": ""
  },
  "field": "",
  "as": "",
  "forceArray": false
}
```

From Table: string   
From Field: string   
Field: string   
As: string   
ForceArray: boolean   
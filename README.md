![Build Status](https://github.com/cloudacademy/voteapp-api-go/actions/workflows/go.yml/badge.svg) 
![GitHub release (latest SemVer)](https://img.shields.io/github/v/release/cloudacademy/voteapp-api-go)

# Background
Provides a CRUD based API written in Go. The API is designed to read and write into a MongoDB backend database.

# API endpoints
The API provides the following endpoints:
```
GET /ok
GET /cpu
GET /version
GET /languages
GET /languages/{name}
GET /languages/{name}/vote
POST /languages/{name}
DELETE /languages/{name}
```

# API GETs with curl and jq
The API can be used to perform GETs with **curl** and **jq** like so:
```
curl -s http://localhost:8080/languages | jq .
curl -s http://localhost:8080/languages/{name} | jq .
curl -s http://localhost:8080/languages/{name}/vote | jq .
```

# API POSTs with curl
The API can be used to perform POSTs with **curl** like so:
```
curl http://localhost:8080/languages/{name} \
--header "Content-Type: application/json" \
--request POST \
--data-binary @- <<BODY
{
    "Usecase": "system, web, server-side",
    "Rank": 5,
    "Compiled": false,
    "Homepage": "https://www.php.net/",
    "Download": "https://www.php.net/downloads.php",
    "Votes": 0
}
BODY
```

# API DELETEs with curl and jq
The API can be used to perform DELETEs with **curl** like so:
```
curl -s -X DELETE http://localhost:8080/languages/{name}
```

# Linux Compiling
The API can be compiled using the following commands:
```
#ensure to be in the same dir as the **main.go** file
go get -v -t -d ./...
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o api
```

# API MongoDB Database
The API is designed to read/write into a MongoDB 4.2.x database. The MongoDB database should be setup and populated using the following commands performed within the mongo shell:
```
use langdb;

db.createUser({user: "admin",
pwd: "password",
roles:[{role: "userAdmin" , db:"langdb"}]
});

db.languages.insert({"name" : "csharp", "codedetail" : { "usecase" : "system, web, server-side", "rank" : 5, "compiled" : false, "homepage" : "https://dotnet.microsoft.com/learn/csharp", "download" : "https://dotnet.microsoft.com/download/", "votes" : 0}});
db.languages.insert({"name" : "python", "codedetail" : { "usecase" : "system, web, server-side", "rank" : 3, "script" : false, "homepage" : "https://www.python.org/", "download" : "https://www.python.org/downloads/", "votes" : 0}});
db.languages.insert({"name" : "javascript", "codedetail" : { "usecase" : "web, client-side", "rank" : 7, "script" : false, "homepage" : "https://en.wikipedia.org/wiki/JavaScript", "download" : "n/a", "votes" : 0}});
db.languages.insert({"name" : "go", "codedetail" : { "usecase" : "system, web, server-side", "rank" : 12, "compiled" : true, "homepage" : "https://golang.org", "download" : "https://golang.org/dl/", "votes" : 0}});
db.languages.insert({"name" : "java", "codedetail" : { "usecase" : "system, web, server-side", "rank" : 1, "compiled" : true, "homepage" : "https://www.java.com/en/", "download" : "https://www.java.com/en/download/", "votes" : 0}});
db.languages.insert({"name" : "nodejs", "codedetail" : { "usecase" : "system, web, server-side", "rank" : 20, "script" : false, "homepage" : "https://nodejs.org/en/", "download" : "https://nodejs.org/en/download/", "votes" : 0}});

show collections;

db.languages.find();
```

# API Environment Vars
The API looks for the following defined environment variables:
```
MONGO_CONN_STR=mongodb://localhost:27017/langdb
MONGO_USERNAME=admin
MONGO_PASSWORD=password
```
**Note**: The environment variables `MONGO_USERNAME` and `MONGO_USERNAME` don't need to be specified if authentication is disabled or not configured on the MongoDB service.

# API Startup
The API can be started directly using the **main.go** file like so
```
MONGO_CONN_STR=mongodb://localhost:27017/langdb MONGO_USERNAME=admin MONGO_PASSWORD=password go run main.go
```
or by using the binary:
```
MONGO_CONN_STR=mongodb://localhost:27017/langdb MONGO_USERNAME=admin MONGO_PASSWORD=password ./api
```

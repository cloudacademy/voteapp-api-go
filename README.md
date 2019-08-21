# voteapp-api-go
Provides a CRUD based API written in Go. The API is designed to read and write into a MongoDB backend database. The API is utilised by the [Language Voting](https://github.com/cloudacademy/voteapp-frontend-react) frontend web app. The frontend is developed using React and makes AJAX requests to this API.

# API endpoints
The API provides the following endpoints:
```
GET /lanugages
GET /languages/{name}
GET /languages/{name}/vote
GET /ok
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
curl http://localhost:8080/languages/php \
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

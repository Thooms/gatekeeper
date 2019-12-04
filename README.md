# gatekeeper

[![GoDoc](https://godoc.org/github.com/Thooms/gatekeeper?status.svg)](https://godoc.org/github.com/Thooms/gatekeeper)


## Build

```
# go test ./...
$ go build ./...
```

## Usage

```go
db, _ := ...                             // from github.com/jmoiron/sqlx
backend := sql.FromxDB(db, "api_usages")
middleware := gatekeeper.FromKeeper(backend)
http.Handle("/hello", middleware.Wrap(someHandler))
```

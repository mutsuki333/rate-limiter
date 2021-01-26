package main

import (
	"database/sql"
	"net/http"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
	limiter "github.com/mutsuki333/rate-limiter"
)

var l *limiter.Limiter

func main() {
	if len(os.Args) > 1 {
		var db *sql.DB
		var err error
		if os.Args[1] == "mysql" {
			db, err = sql.Open("mysql", "root:admin@/limiter?charset=utf8")
		} else {
			db, err = sql.Open("sqlite3", os.Args[1]+".db")
		}
		if err != nil {
			panic(err)
		}
		l = &limiter.Limiter{
			Interval: time.Minute,
			Limit:    5,
			Store:    db,
		}
		l.Init()
	} else {
		l = limiter.Default()
	}

	http.HandleFunc("/hit", l.Handler)

	http.ListenAndServe(":8080", nil)
}

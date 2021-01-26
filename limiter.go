package limiter

import (
	"database/sql"
	"errors"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

//Default setting of the limiter, which limits 60hits/minute and uses in-memory sqlite db
func Default() *Limiter {
	var err error
	db, err := sql.Open("sqlite3", "file::memory:?cache=shared")
	// db, err := sql.Open("sqlite3", "test.db")
	if err != nil {
		panic(err)
	}
	l := &Limiter{
		Interval: time.Minute,
		Limit:    60,
		Mux:      &sync.Mutex{},
		Store:    db,
	}
	l.Init()
	return l
}

//Limiter the limiter instance
type Limiter struct {

	//Interval the base of counting rate
	Interval time.Duration

	//Limit of counts within given interval
	Limit int

	Mux   *sync.Mutex
	Store *sql.DB
}

//Init the db table, and start cleanup goroutine
func (l *Limiter) Init() {
	_, err := l.Store.Exec(`drop table if exists hit;`)
	if err != nil {
		panic(err)
	}
	_, err = l.Store.Exec(`create table hit (ip varchar(20), hit_time datetime);`)
	if err != nil {
		panic(err)
	}
	l.Mux = &sync.Mutex{}
	go func() {
		for {
			time.Sleep(l.Interval)
			l.clear()
		}
	}()
}

//Rate within interval
func (l *Limiter) Rate(ip string) (rate int, err error) {
	l.Mux.Lock()
	defer l.Mux.Unlock()
	row := l.Store.QueryRow(
		"select count(*) from hit where ip = ? and hit_time > ?",
		ip,
		time.Now().Add(-l.Interval),
	)
	err = row.Scan(&rate)
	return
}

//Hit record the hit from an ip
func (l *Limiter) Hit(ip string) error {
	l.Mux.Lock()
	defer l.Mux.Unlock()
	_, err := l.Store.Exec(
		"insert into hit (ip, hit_time) values (?, ?)",
		ip,
		time.Now(),
	)
	return err
}

//HitOrError hit and return the rate, and error if exceeds limit.
func (l *Limiter) HitOrError(ip string) (rate int, err error) {
	l.Hit(ip)
	rate, _ = l.Rate(ip)
	if rate > l.Limit {
		err = errors.New("Rate limit exceeded")
	}
	return
}

//Handler response the rate, and error if exceeds limit.
func (l *Limiter) Handler(w http.ResponseWriter, r *http.Request) {
	ip := r.Header.Get("X-Real-Ip")
	if ip == "" {
		ip = r.Header.Get("X-Forwarded-For")
	}
	if ip == "" {
		ip, _, _ = net.SplitHostPort(r.RemoteAddr)
	}
	rate, err := l.HitOrError(ip)
	if err != nil {
		fmt.Fprintln(w, "Error")
	} else {
		fmt.Fprintln(w, rate)
	}
	return
}

func (l *Limiter) clear() {
	l.Mux.Lock()
	defer l.Mux.Unlock()
	l.Store.Exec("delete from hit where hit_time < ?", time.Now().Add(-l.Interval*2))
}

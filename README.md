# Rate Limiter

A go lib for limiting hit rate of a given ip within a given interval. 
The counting of the rate is based on the hits within the past **interval**

## Start

Run `go run example/main.go` and `curl http://localhost:8080/hit`. (Default 60 hits per minute, and uses in-memory sqlite as database)


## Limiter Usage

### As An Endpoint
```go
var l *limiter.Limiter

func main() {
    l = limiter.Default()

    // Limiter settings can be changed
    // for example:
    // l.Interval = time.Minute * 10
    // l.Limit = 5
    
    http.HandleFunc("/hit", l.Handler)

    http.ListenAndServe(":8080", nil)
}
```
### Use Other Database

This library should comply with most databases, but is only tested with mysql.

```go
var l *limiter.Limiter

func main() {
    db, err = sql.Open("mysql", "root:admin@/limiter?charset=utf8")
    if err != nil {
        panic(err)
    }

    l = &limiter.Limiter{
        Interval: time.Minute,
        Limit:    60,
        Store:    db,
    }
    l.Init()
    
    http.HandleFunc("/hit", l.Handler)

    http.ListenAndServe(":8080", nil)
}
```

### As Custom Gin Middleware

```go
func RateLimit() gin.HandlerFunc {
    l := limiter.Default()
    return func(c *gin.Context) {
        ip := c.Header.Get("X-Real-Ip")
        if ip == "" {
            ip = c.Header.Get("X-Forwarded-For")
        }
        if ip == "" {
            ip, _, _ = net.SplitHostPort(c.RemoteAddr)
        }
        rate, err := l.HitOrError(ip)
        if err != nil {
            c.String(500, fmt.Sprintf("error: %s", err))
        } else {
            c.Next()
        }
    }
}
```
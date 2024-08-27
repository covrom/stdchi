# stdchi
Go 1.22+ standard http router wrapper with API like [chi](https://github.com/go-chi/chi) router.
It uses a new syntax for path values ​​within groups and subroutes.
All of 'chi' routing syntax is supported. The middleware stack and path values providing is more efficient than chi.
It supports lazy mounting. You can create an independent API and then mount it to another router. 

Example:

```go
import "github.com/covrom/stdchi"

// ...

r := stdchi.NewRouter()
r.Use(func(h http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        fmt.Println("Route /sharing")
        r.Write(os.Stdout)
        h.ServeHTTP(w, r)
    })
})

r.Get("/{hash}", func(w http.ResponseWriter, r *http.Request) {
    v := r.PathValue("hash")
    w.Write([]byte(fmt.Sprintf("/%s", v)))
    fmt.Println("Done GET /{hash}")
})

r.Route("/{hash}/share", func(r Router) {
    r.Use(func(h http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            fmt.Println("Route /{hash}/share/")
            r.Write(os.Stdout)
            h.ServeHTTP(w, r)
        })
    })

    r.Get("/{$}", func(w http.ResponseWriter, r *http.Request) {
        v := r.PathValue("hash")
        w.Write([]byte(fmt.Sprintf("/%s/share", v)))
        fmt.Println("Done GET /{hash}/share/")
    })
    r.Get("/{network}", func(w http.ResponseWriter, r *http.Request) {
        v := r.PathValue("hash")
        n := r.PathValue("network")
        w.Write([]byte(fmt.Sprintf("/%s/share/%s", v, n)))
        fmt.Println("Done GET /{hash}/share/{network}")
    })
})

m := stdchi.NewRouter()
m.Mount("/sharing", r)
m.Use(func(h http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        fmt.Println("Route /")
        r.Write(os.Stdout)
        h.ServeHTTP(w, r)
    })
})

// ...

http.ListenAndServe(":8080", m)
```

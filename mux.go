package stdchi

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

var _ Router = &Mux{}

type Mux struct {
	stdmux      *http.ServeMux
	middlewares []func(http.Handler) http.Handler
}

func NewMux() *Mux {
	mux := &Mux{stdmux: http.NewServeMux()}
	return mux
}

func (mx *Mux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	mx.stdmux.ServeHTTP(w, r)
}

func (mx *Mux) mwsHandler(pattern string, h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		chain(mx.middlewares, mwWildcards(pattern, h)).ServeHTTP(w, r)
	})
}

func mwWildcards(pattern string, next http.Handler) http.Handler {
	wilds := uniWildcards(pattern)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		wcs := wildcardsFromContext(ctx)
		for _, ws := range wilds {
			wcs[ws] = r.PathValue(ws)
		}
		ctx = withWildcards(ctx, wcs)
		r = r.WithContext(ctx)
		for k, v := range wcs {
			r.SetPathValue(k, v)
		}
		next.ServeHTTP(w, r)
	})
}

// Use appends a middleware handler to the Mux middleware stack.
//
// The middleware stack for any Mux will execute before searching for a matching
// route to a specific handler, which provides opportunity to respond early,
// change the course of the request execution, or set request-scoped values for
// the next http.Handler.
func (mx *Mux) Use(middlewares ...func(http.Handler) http.Handler) {
	mx.middlewares = append(mx.middlewares, middlewares...)
}

// Handle adds the route `pattern` that matches any http method to
// execute the `handler` http.Handler.
func (mx *Mux) Handle(pattern string, handler http.Handler) {
	parts := strings.SplitN(pattern, " ", 2)
	if len(parts) == 2 {
		mx.Method(parts[0], parts[1], handler)
		return
	}

	mx.handle(mALL, pattern, handler)
}

// HandleFunc adds the route `pattern` that matches any http method to
// execute the `handlerFn` http.HandlerFunc.
func (mx *Mux) HandleFunc(pattern string, handlerFn http.HandlerFunc) {
	parts := strings.SplitN(pattern, " ", 2)
	if len(parts) == 2 {
		mx.Method(parts[0], parts[1], handlerFn)
		return
	}

	mx.handle(mALL, pattern, handlerFn)
}

// Method adds the route `pattern` that matches `method` http method to
// execute the `handler` http.Handler.
func (mx *Mux) Method(method, pattern string, handler http.Handler) {
	m, ok := methodMap[strings.ToUpper(method)]
	if !ok {
		panic(fmt.Sprintf("stdchi: '%s' http method is not supported.", method))
	}
	mx.handle(m, pattern, handler)
}

// MethodFunc adds the route `pattern` that matches `method` http method to
// execute the `handlerFn` http.HandlerFunc.
func (mx *Mux) MethodFunc(method, pattern string, handlerFn http.HandlerFunc) {
	mx.Method(method, pattern, handlerFn)
}

// Connect adds the route `pattern` that matches a CONNECT http method to
// execute the `handlerFn` http.HandlerFunc.
func (mx *Mux) Connect(pattern string, handlerFn http.HandlerFunc) {
	mx.handle(mCONNECT, pattern, handlerFn)
}

// Delete adds the route `pattern` that matches a DELETE http method to
// execute the `handlerFn` http.HandlerFunc.
func (mx *Mux) Delete(pattern string, handlerFn http.HandlerFunc) {
	mx.handle(mDELETE, pattern, handlerFn)
}

// Get adds the route `pattern` that matches a GET http method to
// execute the `handlerFn` http.HandlerFunc.
func (mx *Mux) Get(pattern string, handlerFn http.HandlerFunc) {
	mx.handle(mGET, pattern, handlerFn)
}

// Head adds the route `pattern` that matches a HEAD http method to
// execute the `handlerFn` http.HandlerFunc.
func (mx *Mux) Head(pattern string, handlerFn http.HandlerFunc) {
	mx.handle(mHEAD, pattern, handlerFn)
}

// Options adds the route `pattern` that matches an OPTIONS http method to
// execute the `handlerFn` http.HandlerFunc.
func (mx *Mux) Options(pattern string, handlerFn http.HandlerFunc) {
	mx.handle(mOPTIONS, pattern, handlerFn)
}

// Patch adds the route `pattern` that matches a PATCH http method to
// execute the `handlerFn` http.HandlerFunc.
func (mx *Mux) Patch(pattern string, handlerFn http.HandlerFunc) {
	mx.handle(mPATCH, pattern, handlerFn)
}

// Post adds the route `pattern` that matches a POST http method to
// execute the `handlerFn` http.HandlerFunc.
func (mx *Mux) Post(pattern string, handlerFn http.HandlerFunc) {
	mx.handle(mPOST, pattern, handlerFn)
}

// Put adds the route `pattern` that matches a PUT http method to
// execute the `handlerFn` http.HandlerFunc.
func (mx *Mux) Put(pattern string, handlerFn http.HandlerFunc) {
	mx.handle(mPUT, pattern, handlerFn)
}

// Trace adds the route `pattern` that matches a TRACE http method to
// execute the `handlerFn` http.HandlerFunc.
func (mx *Mux) Trace(pattern string, handlerFn http.HandlerFunc) {
	mx.handle(mTRACE, pattern, handlerFn)
}

// With adds inline middlewares for an endpoint handler.
func (mx *Mux) With(middlewares ...func(http.Handler) http.Handler) Router {
	mws := append(mx.middlewares, middlewares...)

	im := &Mux{
		stdmux:      mx.stdmux,
		middlewares: mws,
	}

	return im
}

// Group creates a new inline-Mux with a copy of middleware stack. It's useful
// for a group of handlers along the same routing path that use an additional
// set of middlewares. See _examples/.
func (mx *Mux) Group(fn func(r Router)) Router {
	im := mx.With()
	if fn != nil {
		fn(im)
	}
	return im
}

// Route creates a new Mux and mounts it along the `pattern` as a subrouter.
// Effectively, this is a short-hand call to Mount. See _examples/.
func (mx *Mux) Route(pattern string, fn func(r Router)) Router {
	if fn == nil {
		panic(fmt.Sprintf("stdchi: attempting to Route() a nil subrouter on '%s'", pattern))
	}
	subRouter := NewRouter()
	fn(subRouter)
	mx.Mount(pattern, subRouter)
	return subRouter
}

// Mount attaches another http.Handler as a subrouter along a routing
// path. It's very useful to split up a large API as many independent routers and
// compose them as a single service using Mount.
//
// Note that Mount() simply sets a wildcard along the `pattern` that will continue
// routing at the `handler`, which in most cases is another stdchi.Router. As a result,
// if you define two Mount() routes on the exact same pattern the mount will panic.
func (mx *Mux) Mount(pattern string, handler http.Handler) {
	if handler == nil {
		panic(fmt.Sprintf("stdchi: attempting to Mount() a nil handler on '%s'", pattern))
	}

	if pattern == "" || (pattern[len(pattern)-1] != '/' && !strings.HasSuffix(pattern, "...}")) {
		pattern += "/"
	}

	mx.handle(mALL, pattern, StripSegments(pattern, handler))
}

// StripSegments works like http.StripPrefix, but skips entire segments (including wildcards) and provides path values ​​to subrouters.
func StripSegments(pat string, h http.Handler) http.Handler {
	wilds := wildcards(pat)
	if len(wilds) == 0 {
		return h
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p := stripToLastSlash(r.URL.Path, len(wilds))
		rp := stripToLastSlash(r.URL.RawPath, len(wilds))
		if len(p) < len(r.URL.Path) && (r.URL.RawPath == "" || len(rp) < len(r.URL.RawPath)) {
			r2 := new(http.Request)
			*r2 = *r
			r2.URL = new(url.URL)
			*r2.URL = *r.URL
			r2.URL.Path = p
			r2.URL.RawPath = rp

			ctx := r.Context()
			wcs := wildcardsFromContext(ctx)
			for _, ws := range wilds {
				if ws == "" {
					continue
				}
				wcs[ws] = r.PathValue(ws)
			}
			ctx = withWildcards(ctx, wcs)
			r2 = r2.WithContext(ctx)
			h.ServeHTTP(w, r2)
		} else {
			http.NotFound(w, r)
		}
	})
}

// Middlewares returns a slice of middleware handler functions.
func (mx *Mux) Middlewares() Middlewares {
	return mx.middlewares
}

// handle registers a http.Handler in the routing tree for a particular http method
// and routing pattern.
func (mx *Mux) handle(method methodTyp, pattern string, handler http.Handler) {
	if len(pattern) == 0 || pattern[0] != '/' {
		panic(fmt.Sprintf("stdchi: routing pattern must begin with '/' in '%s'", pattern))
	}

	if method&mALL == mALL {
		mx.stdmux.Handle(pattern, mx.mwsHandler(pattern, handler))
	} else {
		for k, v := range methodMap {
			if method&v == v {
				mx.stdmux.Handle(fmt.Sprintf("%s %s", k, pattern), mx.mwsHandler(pattern, handler))
			}
		}
	}
}

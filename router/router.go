package router

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/rest/httpx"
	"github.com/zeromicro/go-zero/rest/router"
)

// Router returns a customized httpx Router
func New(path string, fs http.FileSystem) httpx.Router {
	rt := router.NewRouter()
	return newFileServingRouter(rt, path, fs)
}

type fileServingRouter struct {
	httpx.Router
	middleware rest.Middleware
}

func newFileServingRouter(router httpx.Router, path string, fs http.FileSystem) httpx.Router {
	return &fileServingRouter{
		Router:     router,
		middleware: fileHandler(path, fs),
	}
}

func (f *fileServingRouter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	f.middleware(f.Router.ServeHTTP)(w, r)
}

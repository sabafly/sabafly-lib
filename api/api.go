package api

import (
	"sync"

	"github.com/gin-gonic/gin"
)

// APIサーバーの構造体
type Server[T any] struct {
	sync.Mutex

	gin      *gin.Engine
	PageTree *Page[T]
	NoRoute  []gin.HandlerFunc
	NoMethod []gin.HandlerFunc

	Client T
}

// ページの構造を表す構造体
type Page[T any] struct {
	Path     string
	Handlers []*Handler[T]

	Child []*Page[T]
}

type HandlerFunc[T any] func(*Server[T], *gin.Context)

type HandlerCheck[T any] func(*Server[T], *gin.Context) bool

// メソッドとハンダラの構造体
type Handler[T any] struct {
	Method  string
	Check   HandlerCheck[T]
	Handler HandlerFunc[T]
}

// 新たなサーバー構造を生成する
func New[T any](client T) *Server[T] {
	return &Server[T]{
		gin:     gin.New(),
		NoRoute: []gin.HandlerFunc{},
		Client:  client,
	}
}

// ページを解析してサーバーを起動する
func (s *Server[T]) Serve(addr ...string) (err error) {
	s.gin.Use(gin.Logger())
	s.gin.HandleMethodNotAllowed = true
	s.gin.NoRoute(s.NoRoute...)
	s.gin.NoMethod(s.NoMethod...)
	s.PageTree.Parse(s, s.gin)
	return s.gin.Run(addr...)
}

// ページ構造を解析してginエンジンに登録する
//
// ハンダラがnilだった場合無視される
func (p *Page[T]) Parse(s *Server[T], g *gin.Engine) {
	for _, h := range p.Handlers {
		if h.Handler != nil {
			g.Handle(h.Method, p.Path, s.checkHandler(h.Check, h.Handler))
		}
	}
	for _, p2 := range p.Child {
		p2.parse(s, p2, p.Path+p2.Path, g)
	}
}

// ginエンジンにページハンダラを登録する
func (p *Page[T]) parse(s *Server[T], page *Page[T], path string, g *gin.Engine) {
	for _, h := range page.Handlers {
		if h.Handler != nil {
			g.Handle(h.Method, path, s.checkHandler(h.Check, h.Handler))
		}
	}
	for _, p2 := range p.Child {
		p2.parse(s, p2, path+p2.Path, g)
	}
}

func (s *Server[T]) checkHandler(check func(*Server[T], *gin.Context) bool, handler func(*Server[T], *gin.Context)) func(*gin.Context) {
	if check == nil {
		return s.handler(handler)
	}
	return func(ctx *gin.Context) {
		if check(s, ctx) {
			handler(s, ctx)
		}
	}
}

func (s *Server[T]) handler(handler func(*Server[T], *gin.Context)) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		handler(s, ctx)
	}
}

package gin

import "github.com/gin-gonic/gin"

type HandlerFunc = gin.HandlerFunc
type Engine = gin.Engine

func New(middlewares ...HandlerFunc) *Engine {
	e := gin.New()
	e.Use(middlewares...)
	return e
}

func Logger() HandlerFunc {
	return gin.Logger()
}

func Recovery() HandlerFunc {
	return gin.Recovery()
}

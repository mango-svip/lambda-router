package router

import (
    "github.com/aws/aws-lambda-go/events"
)

// MiddlewareFunc 是中间件函数的类型
type MiddlewareFunc func(*Context)

type Lambda struct {
    *Router
    middlewares []MiddlewareFunc
}

type Context struct {
    index int8
    *events.LambdaFunctionURLRequest
}

func Default() *Lambda {
    l := &Lambda{
        New(),
        make([]MiddlewareFunc, 0),
    }

    return l
}

// Use 方法用于添加中间件到中间件链
func (l *Lambda) Use(middleware MiddlewareFunc) {
    l.middlewares = append(l.middlewares, middleware)
}

func (l *Lambda) Next(c *Context) {
    if c.index < int8(len(l.middlewares)) {
        c.index++
        l.middlewares[c.index-1](c)
    }
}

func (l *Lambda) ServeHTTP(req events.LambdaFunctionURLRequest) (events.LambdaFunctionURLResponse, error) {
    if len(l.middlewares) > 0 {
        l.Next(&Context{
            0,
            &req,
        })
    }
    return l.Router.ServeHTTP(req)
}

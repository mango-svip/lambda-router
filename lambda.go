package router

import (
    "github.com/aws/aws-lambda-go/events"
)

// MiddlewareFunc 是中间件函数的类型
type MiddlewareFunc func(*events.LambdaFunctionURLRequest, func())

type Lambda struct {
    *Router
    middlewares []MiddlewareFunc
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

func (l *Lambda) Next(req *events.LambdaFunctionURLRequest) {
    var index int

    var next func()
    next = func() {
        if index < len(l.middlewares) {
            index++
            l.middlewares[index-1](req, next)
        }
    }
    next()
}

func (l *Lambda) ServeHTTP(req events.LambdaFunctionURLRequest) (events.LambdaFunctionURLResponse, error) {
    if len(l.middlewares) > 0 {
        l.Next(&req)
    }
    return l.Router.ServeHTTP(req)
}

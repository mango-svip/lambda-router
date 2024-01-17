package router

import (
    "fmt"
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
    middlewares []MiddlewareFunc
    resp        events.LambdaFunctionURLResponse
    err         error
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

func (c *Context) Next() {
    if c.index < int8(len(c.middlewares)) {
        c.index++
        c.middlewares[c.index-1](c)
    }
}

func (l *Lambda) ServeHTTP(req events.LambdaFunctionURLRequest) (events.LambdaFunctionURLResponse, error) {
    c := &Context{
        index:                    0,
        LambdaFunctionURLRequest: &req,
        middlewares:              l.middlewares,
    }

    c.middlewares = append(c.middlewares, func(c *Context) {
        fmt.Println("处理请求")
        response, err := l.Router.ServeHTTP(*c.LambdaFunctionURLRequest)
        c.resp = response
        c.err = err
        c.Next()
    })

    c.Next()
    return c.resp, c.err

}

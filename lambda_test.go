package router

import (
    "github.com/aws/aws-lambda-go/events"
    "testing"
)

func TestLambda_next(t *testing.T) {

    l := Default()
    l.Use(func(c *Context) {
        t.Log("执行第一个中间件")
        c.Next()
    })
    l.Use(func(c *Context) {
        c.Next()
        t.Log("执行第二个中间件")
    })
    // l.Use(func(c *Context) {
    //     t.Log("执行第三个中间件")
    // })
    l.ServeHTTP(events.LambdaFunctionURLRequest{})

}

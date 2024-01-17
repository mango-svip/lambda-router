package router

import (
    "github.com/aws/aws-lambda-go/events"
    "testing"
)

func TestLambda_next(t *testing.T) {

    l := Default()
    l.Use(func(c *Context) {
        t.Log("执行第一个中间件")
        l.Next(c)
    })
    l.Use(func(c *Context) {
        t.Log("执行第二个中间件")
        l.Next(c)
    })
    l.Use(func(c *Context) {
        t.Log("执行第三个中间件")
        l.Next(c)
    })
    l.ServeHTTP(events.LambdaFunctionURLRequest{})

}

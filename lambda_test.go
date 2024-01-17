package router

import (
    "github.com/aws/aws-lambda-go/events"
    "testing"
)

func TestLambda_next(t *testing.T) {

    l := Default()
    l.Use(func(request *events.LambdaFunctionURLRequest, next func()) {
        t.Log("执行第一个中间件")
        next()
    })
    l.Use(func(request *events.LambdaFunctionURLRequest, next func()) {
        t.Log("执行第二个中间件")
        next()
    })
    l.ServeHTTP(events.LambdaFunctionURLRequest{})

}

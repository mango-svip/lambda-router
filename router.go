// Copyright 2013 Julien Schmidt. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

// Package httprouter is a trie based high performance HTTP request router.
//
// A trivial example is:
//
//  package main
//
//  import (
//      "fmt"
//      "github.com/julienschmidt/httprouter"
//      "net/http"
//      "log"
//  )
//
//  func Index(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
//      fmt.Fprint(w, "Welcome!\n")
//  }
//
//  func Hello(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
//      fmt.Fprintf(w, "hello, %s!\n", ps.ByName("name"))
//  }
//
//  func main() {
//      router := httprouter.New()
//      router.GET("/", Index)
//      router.GET("/hello/:name", Hello)
//
//      log.Fatal(http.ListenAndServe(":8080", router))
//  }
//
// The router matches incoming requests by the request method and the path.
// If a handle is registered for this path and method, the router delegates the
// request to that function.
// For the methods GET, POST, PUT, PATCH and DELETE shortcut functions exist to
// register handles, for all other methods router.Handle can be used.
//
// The registered path, against which the router matches incoming requests, can
// contain two types of parameters:
//  Syntax    Type
//  :name     named parameter
//  *name     catch-all parameter
//
// Named parameters are dynamic path segments. They match anything until the
// next '/' or the path end:
//  Path: /blog/:category/:post
//
//  Requests:
//   /blog/go/request-routers            match: category="go", post="request-routers"
//   /blog/go/request-routers/           no match, but the router would redirect
//   /blog/go/                           no match
//   /blog/go/request-routers/comments   no match
//
// Catch-all parameters match anything until the path end, including the
// directory index (the '/' before the catch-all). Since they match anything
// until the end, catch-all paramerters must always be the final path element.
//  Path: /files/*filepath
//
//  Requests:
//   /files/                             match: filepath="/"
//   /files/LICENSE                      match: filepath="/LICENSE"
//   /files/templates/article.html       match: filepath="/templates/article.html"
//   /files                              no match, but the router would redirect
//
// The value of parameters is saved as a slice of the Param struct, consisting
// each of a key and a value. The slice is passed to the Handle func as a third
// parameter.
// There are two ways to retrieve the value of a parameter:
//  // by the name of the parameter
//  user := ps.ByName("user") // defined by :user or *user
//
//  // by the index of the parameter. This way you can also get the name (key)
//  thirdKey   := ps[2].Key   // the name of the 3rd parameter
//  thirdValue := ps[2].Value // the value of the 3rd parameter
package router

import (
    "github.com/aws/aws-lambda-go/events"
    "net/http"
)

// Handle is a function that can be registered to a route to handle HTTP
// requests. Like http.HandlerFunc, but has a third parameter for the values of
// wildcards (variables).
type Handle func(request *events.LambdaFunctionURLRequest, param Params) (*events.LambdaFunctionURLResponse, error)

// Param is a single URL parameter, consisting of a key and a value.
type Param struct {
    Key   string
    Value string
}

// Params is a Param-slice, as returned by the router.
// The slice is ordered, the first URL parameter is also the first slice value.
// It is therefore safe to read values by the index.
type Params []Param

// ByName returns the value of the first Param which key matches the given name.
// If no matching Param is found, an empty string is returned.
func (ps Params) ByName(name string) string {
    for i := range ps {
        if ps[i].Key == name {
            return ps[i].Value
        }
    }
    return ""
}

// Router is a http.Handler which can be used to dispatch requests to different
// handler functions via configurable routes
type Router struct {
    trees map[string]*node

    // Enables automatic redirection if the current route can't be matched but a
    // handler for the path with (without) the trailing slash exists.
    // For example if /foo/ is requested but a route only exists for /foo, the
    // client is redirected to /foo with http status code 301 for GET requests
    // and 307 for all other request methods.
    RedirectTrailingSlash bool

    // If enabled, the router tries to fix the current request path, if no
    // handle is registered for it.
    // First superfluous path elements like ../ or // are removed.
    // Afterwards the router does a case-insensitive lookup of the cleaned path.
    // If a handle can be found for this route, the router makes a redirection
    // to the corrected path with status code 301 for GET requests and 307 for
    // all other request methods.
    // For example /FOO and /..//Foo could be redirected to /foo.
    // RedirectTrailingSlash is independent of this option.
    RedirectFixedPath bool

    // If enabled, the router checks if another method is allowed for the
    // current route, if the current request can not be routed.
    // If this is the case, the request is answered with 'Method Not Allowed'
    // and HTTP status code 405.
    // If no other Method is allowed, the request is delegated to the NotFound
    // handler.
    HandleMethodNotAllowed bool

    // Configurable http.HandlerFunc which is called when no matching route is
    // found. If it is not set, http.NotFound is used.
    NotFound http.HandlerFunc

    // Configurable http.HandlerFunc which is called when a request
    // cannot be routed and HandleMethodNotAllowed is true.
    // If it is not set, http.Error with http.StatusMethodNotAllowed is used.
    MethodNotAllowed http.HandlerFunc

    // Function to handle panics recovered from http handlers.
    // It should be used to generate a error page and return the http error code
    // 500 (Internal Server Error).
    // The handler can be used to keep your server from crashing because of
    // unrecovered panics.
    PanicHandler func(request events.LambdaFunctionURLRequest, rcv interface{}) (events.LambdaFunctionURLResponse, error)
}

// Make sure the Router conforms with the http.Handler interface

// New returns a new initialized Router.
// Path auto-correction, including trailing slashes, is enabled by default.
func New() *Router {
    return &Router{
        RedirectTrailingSlash:  true,
        RedirectFixedPath:      true,
        HandleMethodNotAllowed: true,
    }
}

// GET is a shortcut for router.Handle("GET", path, handle)
func (r *Router) GET(path string, handle Handle) {
    r.Handle("GET", path, handle)
}

// HEAD is a shortcut for router.Handle("HEAD", path, handle)
func (r *Router) HEAD(path string, handle Handle) {
    r.Handle("HEAD", path, handle)
}

// OPTIONS is a shortcut for router.Handle("OPTIONS", path, handle)
func (r *Router) OPTIONS(path string, handle Handle) {
    r.Handle("OPTIONS", path, handle)
}

// POST is a shortcut for router.Handle("POST", path, handle)
func (r *Router) POST(path string, handle Handle) {
    r.Handle("POST", path, handle)
}

// PUT is a shortcut for router.Handle("PUT", path, handle)
func (r *Router) PUT(path string, handle Handle) {
    r.Handle("PUT", path, handle)
}

// PATCH is a shortcut for router.Handle("PATCH", path, handle)
func (r *Router) PATCH(path string, handle Handle) {
    r.Handle("PATCH", path, handle)
}

// DELETE is a shortcut for router.Handle("DELETE", path, handle)
func (r *Router) DELETE(path string, handle Handle) {
    r.Handle("DELETE", path, handle)
}

// Handle registers a new request handle with the given path and method.
//
// For GET, POST, PUT, PATCH and DELETE requests the respective shortcut
// functions can be used.
//
// This function is intended for bulk loading and to allow the usage of less
// frequently used, non-standardized or custom methods (e.g. for internal
// communication with a proxy).
func (r *Router) Handle(method, path string, handle Handle) {
    if path[0] != '/' {
        panic("path must begin with '/'")
    }

    if r.trees == nil {
        r.trees = make(map[string]*node)
    }

    root := r.trees[method]
    if root == nil {
        root = new(node)
        r.trees[method] = root
    }

    root.addRoute(path, handle)
}

// HandlerFunc is an adapter which allows the usage of an http.HandlerFunc as a
// request handle.

// ServeFiles serves files from the given file system root.
// The path must end with "/*filepath", files are then served from the local
// path /defined/root/dir/*filepath.
// For example if root is "/etc" and *filepath is "passwd", the local file
// "/etc/passwd" would be served.
// Internally a http.FileServer is used, therefore http.NotFound is used instead
// of the Router's NotFound handler.
// To use the operating system's file system implementation,
// use http.Dir:
//     router.ServeFiles("/src/*filepath", http.Dir("/var/www"))

func (r *Router) recv(request events.LambdaFunctionURLRequest) (events.LambdaFunctionURLResponse, error) {
    if rcv := recover(); rcv != nil {
        return r.PanicHandler(request, rcv)
    }
    return events.LambdaFunctionURLResponse{Body: "recover"}, nil
}

// Lookup allows the manual lookup of a method + path combo.
// This is e.g. useful to build a framework around this router.
// If the path was found, it returns the handle function and the path parameter
// values. Otherwise the third return value indicates whether a redirection to
// the same path with an extra / without the trailing slash should be performed.
func (r *Router) Lookup(method, path string) (Handle, Params, bool) {
    if root := r.trees[method]; root != nil {
        return root.getValue(path)
    }
    return nil, nil, false
}

// ServeHTTP makes the router implement the http.Handler interface.
func (r *Router) ServeHTTP(req events.LambdaFunctionURLRequest) (events.LambdaFunctionURLResponse, error) {
    if r.PanicHandler != nil {
        defer r.recv(req)
    }
    context := req.RequestContext
    method := context.HTTP.Method
    path := context.HTTP.Path
    if root := r.trees[method]; root != nil {
        if handle, ps, _ := root.getValue(path); handle != nil {
            response, err := handle(&req, ps)
            return *response, err
        }
    }

    // Handle 404
    return events.LambdaFunctionURLResponse{StatusCode: 404, Body: "NOT FOUND"}, nil
}

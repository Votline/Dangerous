// Package middlewares repo.go provides Middleware interface
package middlewares

import "net/http"

type Middleware interface {
	Handle(next http.HandlerFunc) http.HandlerFunc
}

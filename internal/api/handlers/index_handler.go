package handlers

import (
	"fmt"
	"net/http"
)

func IndexHandler() http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		fmt.Fprint(writer, "Hello world")
	})

}

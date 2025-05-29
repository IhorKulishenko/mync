package testutils

import (
	"fmt"
	"net/http"
	"net/http/httptest"
)

func StartTestHttpServer(response string) *httptest.Server {
	return httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, response)
		}),
	)
}

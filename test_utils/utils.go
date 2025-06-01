package testutils

import (
	"fmt"
	"io"
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

func StartExtendedTestHttpServer() *httptest.Server {
	return httptest.NewServer(
		http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			writer.Header().Add("Content-Type", "plain/text")

			fmt.Fprintf(writer, "method: %s\n", request.Method)
			buffer, error := io.ReadAll(request.Body)
			if error != nil {
				fmt.Fprintf(writer, "error: %s\n", error.Error())
			}
			defer request.Body.Close()

			fmt.Fprintf(writer, "body: %#v\n", string(buffer))

		}),
	)
}

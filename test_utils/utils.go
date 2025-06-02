package testutils

import (
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
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
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			contentType := r.Header.Get("Content-Type")
			if strings.HasPrefix(contentType, "multipart/form-data") {
				multipartServerHandler(w, r)
			} else {
				extendServerHandler(w, r)
			}
		}),
	)
}

func extendServerHandler(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Add("Content-Type", "plain/text")

	fmt.Fprintf(writer, "method: %s\n", request.Method)
	buffer, error := io.ReadAll(request.Body)
	if error != nil {
		fmt.Fprintf(writer, "error: %s\n", error.Error())
	}
	defer request.Body.Close()

	fmt.Fprintf(writer, "body: %#v\n", string(buffer))
}

func multipartServerHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(5000)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	if len(r.MultipartForm.Value) > 0 {
		fmt.Fprint(w, "value: ")
		fmt.Fprintln(w, strings.Join(mfValueToStringSlices(r.MultipartForm.Value), ";"))
	}

	if len(r.MultipartForm.File) > 0 {
		fmt.Fprint(w, "files: ")
		for k, v := range r.MultipartForm.File {
			fmt.Fprintf(w, "%s=%s;", k, strings.Join(toFilenameSlice(v), ","))
		}
		fmt.Fprintln(w)
	}

}

func mfValueToStringSlices(m map[string][]string) []string {
	result := make([]string, len(m))

	var index = 0
	for k, v := range m {
		result[index] = fmt.Sprintf("%s=%s", k, strings.Join(v, ","))
		index++
	}

	slices.Sort(result)

	return result
}

func toFilenameSlice(slice []*multipart.FileHeader) []string {
	result := make([]string, len(slice))
	for index, s := range slice {
		result[index] = s.Filename
	}

	return result
}

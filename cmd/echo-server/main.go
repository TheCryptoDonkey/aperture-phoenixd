package main

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
)

func main() {
	addr := ":8080"
	if port := os.Getenv("PORT"); port != "" {
		addr = ":" + port
	}

	handler := func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, 1<<20) // 1 MiB
		body, err := io.ReadAll(r.Body)
		if err != nil {
			var maxErr *http.MaxBytesError
			if errors.As(err, &maxErr) {
				http.Error(w, "request body too large", http.StatusRequestEntityTooLarge)
				return
			}
			fmt.Fprintln(w, "hello")
			return
		}
		if len(body) == 0 {
			fmt.Fprintln(w, "hello")
			return
		}
		_, _ = w.Write(body)
	}

	http.HandleFunc("/v1/echo", handler)
	http.HandleFunc("/", handler)

	fmt.Fprintf(os.Stderr, "echo-server listening on %s\n", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		fmt.Fprintf(os.Stderr, "echo-server: %v\n", err)
		os.Exit(1)
	}
}

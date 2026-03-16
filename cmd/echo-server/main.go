package main

import (
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

	http.HandleFunc("/v1/echo", func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil || len(body) == 0 {
			fmt.Fprintln(w, "hello")
			return
		}
		w.Write(body)
	})

	fmt.Fprintf(os.Stderr, "echo-server listening on %s\n", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		fmt.Fprintf(os.Stderr, "echo-server: %v\n", err)
		os.Exit(1)
	}
}

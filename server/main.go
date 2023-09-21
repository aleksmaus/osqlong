package main

import (
	"fmt"
	"net/http"
	"time"
)

// Simple http server that allows to emulate long running queries with osquery delaying response by the provided sleep param
func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		var (
			sp  time.Duration
			err error
		)

		s := r.URL.Query().Get("sleep")
		if s != "" {
			sp, err = time.ParseDuration(s)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				fmt.Fprintf(w, "Invalid sleep parameter value: %v, err: %v\n", s, err)
				return
			}
		}
		if sp > 0 {
			time.Sleep(sp)
			fmt.Fprintf(w, "Done after wait of %v\n", sp)
			return
		}

		fmt.Fprintln(w, "Done")
	})

	addr := "localhost:8080"

	fmt.Printf("Server is listening on %s...\n", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		fmt.Println("Error:", err)
	}
}

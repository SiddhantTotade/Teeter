package main

import (
	"fmt"
	"net/http"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Backend 1 received request")
		w.Write([]byte("Hello from Backend 1"))
	})

	fmt.Println("Backend 1 running on :8081")
	http.ListenAndServe(":8081", nil)
}

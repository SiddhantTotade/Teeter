package main

import (
	"fmt"
	"net/http"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Backend 2 received request")
		w.Write([]byte("Hello from Backend 2"))
	})

	fmt.Println("Backend 2 running on :8082")
	http.ListenAndServe(":8082", nil)
}

package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

const port = "8080"

func main() {
	fs := http.FileServer(http.Dir("."))

	http.HandleFunc("/*", func(w http.ResponseWriter, r *http.Request) {
		log.Println("Serving", r.URL.Path)

		if _, err := os.Stat("." + r.URL.Path); os.IsNotExist(err) {
			http.ServeFile(w, r, "./index.html")
		} else {
			fs.ServeHTTP(w, r)
		}
	})

	log.Println("Server running on port", port)
	http.ListenAndServe(fmt.Sprintf(":%s", port), nil)
}

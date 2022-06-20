package t

import (
	"fmt"
	"html"
	"log"
	"net/http"
)

func t() {

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello, %q", html.EscapeString(r.URL.Path))
	})

	http.HandleFunc("/hi", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hi")
	})

	log.Fatal(http.ListenAndServeTLS(":8081", "../YOURPUBLIC.pem", "../YOURPRIVATE.key", nil))

}

package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

var DefaultPort = os.Getenv("PORT")

func main() {

	if DefaultPort == "" {
		DefaultPort = "8080"
	}

	log.Println("starting server, listening on port " + DefaultPort)

	http.HandleFunc("/", DebugRequest)
	http.ListenAndServe(":"+DefaultPort, nil)
}

func DebugRequest(w http.ResponseWriter, r *http.Request) {
	fmt.Println(">", r.Method, r.RequestURI)
	fmt.Println("Headers:")
	for key := range r.Header {
		fmt.Println("\t", key, "=", r.Header.Get(key))
	}
	bodyBytes, _ := ioutil.ReadAll(r.Body)
	fmt.Printf("Body (len = %d bytes):\n", len(bodyBytes))
	if len(bodyBytes) > 0 {
		fmt.Println("BODY START>>")
		fmt.Println(string(bodyBytes))
		fmt.Println("<<BODY END")
	}
	fmt.Println()

	defer r.Body.Close()

	w.WriteHeader(http.StatusOK)
}

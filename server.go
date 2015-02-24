package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

var (
	port       = flag.Int("p", 8081, "Port to listen on")
	size       = flag.Int("s", 100, "Size of cluster abstracted as number of offers")
	frameworks = newFrameworks()
)

func main() {
	flag.Parse()
	rand.Seed(time.Now().UnixNano())

	r := mux.NewRouter()
	call := newCall()
	r.Path("/call").Methods("POST").HandlerFunc(call.handle)
	r.Path("/events").Methods("POST").HandlerFunc(events)

	addr := fmt.Sprintf("0.0.0.0:%d", *port)

	log.Printf("Example app listening at http://%s\n", addr)
	log.Fatal(http.ListenAndServe(addr, r))
}

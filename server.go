package main

import (
	"flag"
	"fmt"

	"math/rand"
	"net/http"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	"github.com/jimenez/mesos-APIproto/mesosproto"
)

var (
	credentials = flag.String("c", "", "Credentials for framework authentication")
	port        = flag.Int("p", 8081, "Port to listen on")
	rateLimits  = flag.String("r", "", "Rate limits")

	size       = flag.Int("s", 100, "Size of cluster abstracted as number of offers")
	timeout    = flag.Float64("t", mesosproto.Default_FrameworkInfo_FailoverTimeout, "Failover timeout")
	frameworks = newFrameworks()
)

func main() {
	flag.Parse()
	rand.Seed(time.Now().UnixNano())

	r := mux.NewRouter()
	call := newCall()
	r.Path("/call").Methods("POST").HandlerFunc(call.handle)

	addr := fmt.Sprintf("0.0.0.0:%d", *port)

	log.Infof("Example scheduler listening at http://%s", addr)
	log.Fatal(http.ListenAndServe(addr, r))
}

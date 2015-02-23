package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/VoltFramework/volt/mesosproto"
	"github.com/gorilla/mux"
)

type FrameworksChans struct {
	sync.RWMutex
	queues map[string]map[string]chan *mesosproto.Event
}

func newFrameworksChans() *FrameworksChans {
	return &FrameworksChans{
		queues: make(map[string]map[string]chan *mesosproto.Event),
	}
}

func (fc *FrameworksChans) create(ID, remoteAddr string) chan *mesosproto.Event {
	fc.Lock()
	defer fc.Unlock()

	if _, ok := fc.queues[ID]; !ok {
		fc.queues[ID] = make(map[string]chan *mesosproto.Event)
	}
	fc.queues[ID][remoteAddr] = make(chan *mesosproto.Event)
	return fc.queues[ID][remoteAddr]
}

func (fc *FrameworksChans) delete(ID, remoteAddr string) {
	fc.Lock()
	delete(fc.queues[ID], remoteAddr)
	fc.Unlock()
}

func (fc *FrameworksChans) getChans(ID string) map[string]chan *mesosproto.Event {
	fc.RLock()
	defer fc.RUnlock()

	return fc.queues[ID]
}

func (fc *FrameworksChans) hasID(ID string) bool {
	fc.RLock()
	_, ok := fc.queues[ID]
	fc.RUnlock()

	return ok
}

func (fc *FrameworksChans) send(ID string, event *mesosproto.Event) {
	fc.RLock()
	for _, ch := range fc.queues[ID] {
		ch <- event
	}
	fc.RUnlock()
}

var frameworksChans = newFrameworksChans()

func main() {
	rand.Seed(time.Now().UnixNano())

	r := mux.NewRouter()
	call := newCall()
	r.Path("/call").Methods("POST").HandlerFunc(call.handle)
	r.Path("/events").Methods("POST").HandlerFunc(events)

	addr := fmt.Sprintf("0.0.0.0:%d", 8081)

	log.Printf("Example app listening at http://%s\n", addr)
	log.Fatal(http.ListenAndServe(addr, r))
}

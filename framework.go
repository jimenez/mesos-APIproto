package main

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/VoltFramework/volt/mesosproto"
)

type Frameworks struct {
	sync.RWMutex
	fws map[string]*Framework
}

func newFrameworks() *Frameworks {
	return &Frameworks{
		fws: make(map[string]*Framework),
	}
}

func (fws *Frameworks) Get(ID string) *Framework {
	fws.RLock()
	defer fws.RUnlock()
	return fws.fws[ID]
}

func (fws *Frameworks) New(ID string) *Framework {
	fws.Lock()
	defer fws.Unlock()
	fws.fws[ID] = newFramework(ID)
	return fws.fws[ID]
}

func (fws *Frameworks) OffersSize() int {
	fws.RLock()
	defer fws.RUnlock()
	size := 0
	for _, f := range fws.fws {
		size += len(f.offers)
	}
	return size
}

func (fws *Frameworks) deleteFramework(ID string) {
	fws.Lock()
	delete(fws.fws, ID)
	fws.Unlock()
}

type Framework struct {
	sync.RWMutex
	notify     <-chan bool
	connection chan bool
	ID         string
	ch         chan *mesosproto.Event
	offers     map[string]struct{}
	tasks      map[string]struct{}
}

func newFramework(ID string) *Framework {
	return &Framework{
		offers: make(map[string]struct{}),
		tasks:  make(map[string]struct{}),
		ID:     ID,
	}
}

func (fw *Framework) listenOnNotify(timeout float64) {
	<-fw.notify
	fw.deleteChan()
	log.Infof("Connection closed for framework %q", fw.ID)

	select {
	case <-fw.connection:

		log.Infof("Reregistering framework %q", fw.ID)
		return
	case <-time.After(time.Duration(timeout) * time.Second):

		log.Warnf("Deleting framework %q", fw.ID)
		frameworks.deleteFramework(fw.ID)
	}
}

func (fw *Framework) hasChan() bool {
	return fw.ch != nil
}

func (fw *Framework) newChan(res http.ResponseWriter, timeout float64) chan *mesosproto.Event {
	fw.Lock()
	defer fw.Unlock()

	if fw.connection != nil {
		close(fw.connection)
	}
	fw.ch = make(chan *mesosproto.Event)
	fw.connection = make(chan bool)
	fw.notify = res.(http.CloseNotifier).CloseNotify()
	go fw.listenOnNotify(timeout)
	return fw.ch
}

func (fw *Framework) deleteChan() {
	fw.Lock()
	close(fw.ch)

	fw.ch = nil
	fw.Unlock()
}

func (fw *Framework) AddOffer(offerID string) {
	fw.Lock()
	fw.offers[offerID] = struct{}{}
	fw.Unlock()
}

func (fw *Framework) deleteOffer(offerID string) error {
	fw.Lock()
	defer fw.Unlock()
	if _, ok := fw.offers[offerID]; !ok {
		return fmt.Errorf("Unknown offer %q", offerID)

	}

	delete(fw.offers, offerID)
	return nil
}

func (fw *Framework) AddTask(taskID string) {
	fw.Lock()
	fw.tasks[taskID] = struct{}{}
	fw.Unlock()
}

func (fw *Framework) GetTask(taskID string) bool {
	fw.Lock()
	defer fw.Unlock()
	_, ok := fw.tasks[taskID]
	return ok
}

func (fw *Framework) GetTasks() map[string]struct{} {
	fw.Lock()
	defer fw.Unlock()
	return fw.tasks
}

func (fw *Framework) deleteTask(taskID string) error {
	fw.Lock()
	defer fw.Unlock()
	if _, ok := fw.tasks[taskID]; !ok {
		return fmt.Errorf("Unknown task %q", taskID)

	}

	delete(fw.tasks, taskID)
	return nil
}

func (fw *Framework) send(event *mesosproto.Event) {
	fw.ch <- event
}

package main

import (
	"fmt"
	"sync"

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
	fws.fws[ID] = newFramework()
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

// func (fc *Frameworks) deleteFramework(ID string) {
// 	fc.Lock()
// 	delete(fc.queues, ID)
// 	fc.Unlock()
// }

type Framework struct {
	sync.RWMutex
	chans  map[string]chan *mesosproto.Event
	offers map[string]struct{}
	tasks  map[string]struct{}
}

func newFramework() *Framework {
	return &Framework{
		chans:  make(map[string]chan *mesosproto.Event),
		offers: make(map[string]struct{}),
		tasks:  make(map[string]struct{}),
	}
}

func (fw *Framework) newChan(remoteAddr string) chan *mesosproto.Event {
	fw.Lock()
	defer fw.Unlock()

	fw.chans[remoteAddr] = make(chan *mesosproto.Event)
	return fw.chans[remoteAddr]
}

func (fw *Framework) deleteChan(remoteAddr string) {
	fw.Lock()
	delete(fw.chans, remoteAddr)
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

func (fw *Framework) send(event *mesosproto.Event) {
	fw.RLock()
	for _, ch := range fw.chans {
		ch <- event
	}
	fw.RUnlock()
}

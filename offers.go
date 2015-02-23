package main

import "sync"

type FrameworksOffers struct {
	sync.RWMutex
	offers map[string][]string //frameworkId -> [] offerIds
}

func newFrameworksOffers() *FrameworksOffers {
	return &FrameworksOffers{
		offers: make(map[string][]string),
	}
}

func (fwo *FrameworksOffers) writeOffer(ID, info string) {
	fwo.Lock()
	fwo.offers[ID] = append(fwo.offers[ID], info)
	fwo.Unlock()
}

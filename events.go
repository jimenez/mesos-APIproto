package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"time"

	"code.google.com/p/goprotobuf/proto"
	log "github.com/Sirupsen/logrus"
	"github.com/VoltFramework/volt/mesosproto"
)

func sendEvent(encoder *json.Encoder, eventType mesosproto.Event_Type, event *mesosproto.Event) error {
	event.Type = &eventType
	log.WithFields(log.Fields{"Type": event.Type.String()}).Info("Sending event")
	return encoder.Encode(event)
}

func createOffer(frameworkID *mesosproto.FrameworkID) *mesosproto.Offer {
	slaveId, _ := generateID()
	offerId, _ := generateID()
	hostname := "slave.host" + slaveId
	return &mesosproto.Offer{
		Id:          &mesosproto.OfferID{Value: &offerId},
		FrameworkId: frameworkID,
		SlaveId:     &mesosproto.SlaveID{Value: &slaveId},
		Hostname:    &hostname,
		Resources:   resources(),
	}

}

func sendOffers(f *Framework, frameworkID *mesosproto.FrameworkID, remoteAddr string, notifier <-chan bool) {
	for {
		select {
		case <-notifier:
			f.deleteChan(remoteAddr)
			fmt.Println("HTTP connection just closed.")
			return
		case <-time.After(time.Duration(rand.Intn(10)) * time.Second):

			if frameworks.OffersSize() >= *size {
				continue
			}
			offer := createOffer(frameworkID)
			event_type := mesosproto.Event_OFFERS
			event := &mesosproto.Event{
				Type: &event_type,
				Offers: &mesosproto.Event_Offers{
					Offers: []*mesosproto.Offer{offer},
				},
			}
			f.send(event)
			f.AddOffer(offer.Id.GetValue())
		}
	}
}

func events(res http.ResponseWriter, req *http.Request) {
	frameworkInfo := mesosproto.FrameworkInfo{}
	if req.Header.Get("Content-Type") == "application/json" {
		if err := json.NewDecoder(req.Body).Decode(&frameworkInfo); err != nil {
			http.Error(res, err.Error(), 501)
			return
		}
	} else {
		buf, err := ioutil.ReadAll(req.Body)
		if err != nil {
			http.Error(res, err.Error(), 502)
			return
		}
		if err := proto.Unmarshal(buf, &frameworkInfo); err != nil {
			http.Error(res, err.Error(), 503)
			return
		}
	}

	encoder := json.NewEncoder(NewWriteFlusher(res))
	var (
		mchan chan *mesosproto.Event
		ID    string
		f     *Framework
	)
	res.Header().Set("Connection", "keep-alive")
	res.Header().Set("Accept", "application/json")
	res.Header().Set("Content-Type", "application/json")

	if frameworkInfo.GetId() != nil {
		ID = frameworkInfo.Id.GetValue()
		//Check that this framework hasn't registerd yet
		f = frameworks.Get(ID)
		if f == nil {
			http.Error(res, "Unknown framework", 403)
			return
		}
		mchan = f.newChan(req.RemoteAddr)
		// Reregistering framework
		err := sendEvent(encoder, mesosproto.Event_REREGISTERED, &mesosproto.Event{
			Reregistered: &mesosproto.Event_Reregistered{
				FrameworkId: frameworkInfo.Id,
			},
		})
		if err != nil {
			http.Error(res, err.Error(), 500)
			return
		}

	} else {
		//Create and register framework to chan
		ID, err := generateID()
		if err != nil {
			http.Error(res, err.Error(), 500)
			return
		}
		f = frameworks.New(ID)
		mchan = f.newChan(req.RemoteAddr)
		frameworkInfo.Id = &mesosproto.FrameworkID{
			Value: &ID,
		}
		err = sendEvent(encoder, mesosproto.Event_REGISTERED, &mesosproto.Event{
			Registered: &mesosproto.Event_Registered{
				FrameworkId: frameworkInfo.Id,
			},
		})
		if err != nil {
			http.Error(res, err.Error(), 500)
			return
		}
	}

	notify := res.(http.CloseNotifier).CloseNotify()
	go sendOffers(f, frameworkInfo.Id, req.RemoteAddr, notify)

	for {
		mess := <-mchan
		if err := sendEvent(encoder, *mess.Type, mess); err != nil {
			f.deleteChan(req.RemoteAddr)
		}
	}
}

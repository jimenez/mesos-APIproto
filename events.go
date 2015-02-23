package main

import (
	"encoding/json"
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
	log.Infof(" ======= SENDING EVENT ======= Type: %s", event.Type.String())
	return encoder.Encode(event)
}

func sendOffers(encoder *json.Encoder, frameworkInfo *mesosproto.FrameworkInfo) {
	for {
		time.Sleep(time.Duration(rand.Intn(10)) * time.Second)
		slaveId, _ := generateID()
		offerId, _ := generateID()
		hostname := "slave.host" + slaveId
		offer := mesosproto.Offer{
			Id:          &mesosproto.OfferID{Value: &offerId},
			FrameworkId: frameworkInfo.Id,
			SlaveId:     &mesosproto.SlaveID{Value: &slaveId},
			Hostname:    &hostname,
			Resources:   resources(),
		}
		event_type := mesosproto.Event_OFFERS

		event := &mesosproto.Event{
			Type: &event_type,
			Offers: &mesosproto.Event_Offers{
				Offers: []*mesosproto.Offer{&offer},
			},
		}
		frameworksChans.send(frameworkInfo.Id.GetValue(), event)
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
	)

	if frameworkInfo.GetId() != nil {
		ID = frameworkInfo.Id.GetValue()
		//Check that this framework hasn't registerd yet
		if !frameworksChans.hasID(ID) {
			http.Error(res, "Unknown framework", 403)
			return
		}
		mchan = frameworksChans.create(ID, req.RemoteAddr)
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
		mchan = frameworksChans.create(ID, req.RemoteAddr)
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
	go sendOffers(encoder, &frameworkInfo)
	for {
		mess := <-mchan
		if err := sendEvent(encoder, *mess.Type, mess); err != nil {
			frameworksChans.delete(ID, req.RemoteAddr)
		}
	}
}

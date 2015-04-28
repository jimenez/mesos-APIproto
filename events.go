package main

import (
	"io"
	"math/rand"
	"net/http"
	"time"

	"code.google.com/p/goprotobuf/proto"
	log "github.com/Sirupsen/logrus"
	"github.com/jimenez/mesos-APIproto/mesosproto"
)

func sendEvent(encoder icoder, eventType mesosproto.Event_Type, event *mesosproto.Event) error {
	event.Type = &eventType
	if event.Type.String() == "UPDATE" {
		log.WithFields(log.Fields{"Type": event.Type.String(), "TaskState": event.Update.GetStatus().GetState().String()}).Info("Sending event")
	} else {
		log.WithFields(log.Fields{"Type": event.Type.String()}).Info("Sending event")
	}
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

func sendTasksUpdates(f *Framework, frameworkId *mesosproto.FrameworkID) {
	for {
		select {
		case <-f.connection:
			return
		case <-time.After(time.Duration(rand.Intn(10)) * time.Second):
			for ID, _ := range f.GetTasks() {
				taskID := &mesosproto.TaskID{Value: &ID}
				event := generateEventUpdate((mesosproto.TaskState)(mesosproto.TaskState_value[mesosproto.TaskState_name[rand.Int31n(3)]]), taskID)
				f.send(event)
			}
		}
	}
}

func sendOffers(f *Framework, frameworkID *mesosproto.FrameworkID) {
	for {
		select {
		case <-f.connection:
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

type icoder interface {
	Encode(v interface{}) error
}

type protobufEncoder struct {
	w io.Writer
}

func (pe *protobufEncoder) Encode(v interface{}) error {
	data, err := proto.Marshal(v.(proto.Message))
	if err != nil {
		return err
	}
	_, err = pe.w.Write(data)
	return err
}

func events(frameworkInfo *mesosproto.FrameworkInfo, encoder icoder, res http.ResponseWriter) {

	var (
		mchan    chan *mesosproto.Event
		ID       string
		f        *Framework
		failover = *timeout
	)
	res.Header().Set("Connection", "keep-alive")
	res.Header().Set("Accept", "application/json")
	res.Header().Set("Content-Type", "application/json")

	if frameworkInfo.GetFailoverTimeout() != 0 {
		failover = frameworkInfo.GetFailoverTimeout()
	}

	if frameworkInfo.GetId() != nil {
		ID = frameworkInfo.Id.GetValue()
		//Check that this framework hasn't subsribed yet
		f = frameworks.Get(ID)
		if f == nil {
			http.Error(res, "Unknown framework", 403)
			return
		}

		if f.hasChan() {
			http.Error(res, "Framework already subscribed", 403)
			return
		}

		// Resubscribing framework
		mchan = f.newChan(res, failover)
		// Sending subcribed for resubscribtion
		err := sendEvent(encoder, mesosproto.Event_SUBSCRIBED, &mesosproto.Event{
			Subscribed: &mesosproto.Event_Subscribed{
				FrameworkId: frameworkInfo.Id,
			},
		})
		if err != nil {
			http.Error(res, err.Error(), 500)
			return
		}

	} else {
		//Create and subscribe framework to chan
		ID, err := generateID()
		if err != nil {
			http.Error(res, err.Error(), 500)
			return
		}
		f = frameworks.New(ID)
		mchan = f.newChan(res, failover)
		frameworkInfo.Id = &mesosproto.FrameworkID{
			Value: &ID,
		}

		err = sendEvent(encoder, mesosproto.Event_SUBSCRIBED, &mesosproto.Event{
			Subscribed: &mesosproto.Event_Subscribed{
				FrameworkId: frameworkInfo.Id,
			},
		})
		if err != nil {
			http.Error(res, err.Error(), 500)
			return
		}
	}

	go sendOffers(f, frameworkInfo.Id)
	go sendTasksUpdates(f, frameworkInfo.Id)
	for {
		mess := <-mchan
		if mess == nil {
			return
		}
		if err := sendEvent(encoder, *mess.Type, mess); err != nil {
			log.Error(err)
		}
	}
}

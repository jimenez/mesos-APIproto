package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"code.google.com/p/goprotobuf/proto"
	log "github.com/Sirupsen/logrus"
	"github.com/VoltFramework/volt/mesosproto"
)

func sendEvent(encoder *json.Encoder, eventType mesosproto.Event_Type, event *mesosproto.Event) error {
	event.Type = &eventType
	log.Infof(" ======= SENDING EVENT ======= Type: %s", event.Type.String())
	return encoder.Encode(event)
}

func events(res http.ResponseWriter, req *http.Request) {
	registerMessage := mesosproto.RegisterFrameworkMessage{}
	if req.Header.Get("Content-Type") == "application/json" {
		if err := json.NewDecoder(req.Body).Decode(&registerMessage); err != nil {
			http.Error(res, err.Error(), 501)
			return
		}
	} else {
		buf, err := ioutil.ReadAll(req.Body)
		if err != nil {
			http.Error(res, err.Error(), 502)
			return
		}
		if err := proto.Unmarshal(buf, &registerMessage); err != nil {
			http.Error(res, err.Error(), 503)
			return
		}
	}

	encoder := json.NewEncoder(NewWriteFlusher(res))
	var (
		mchan chan *mesosproto.Event
		ID    string
	)

	if registerMessage.GetFramework() != nil && registerMessage.GetFramework().GetId() != nil {
		ID = registerMessage.Framework.Id.GetValue()
		//Check that this framework hasn't registerd yet
		if !frameworksChans.hasID(ID) {
			http.Error(res, "Unknown framework", 406)
			return
		}
		mchan = frameworksChans.create(ID, req.RemoteAddr)
		// Reregistering framework
		err := sendEvent(encoder, mesosproto.Event_REREGISTERED, &mesosproto.Event{
			Reregistered: &mesosproto.Event_Reregistered{
				FrameworkId: registerMessage.Framework.Id,
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
		err = sendEvent(encoder, mesosproto.Event_REGISTERED, &mesosproto.Event{
			Registered: &mesosproto.Event_Registered{
				FrameworkId: &mesosproto.FrameworkID{
					Value: &ID,
				},
			},
		})
		if err != nil {
			http.Error(res, err.Error(), 500)
			return
		}
	}

	for {
		mess := <-mchan
		if err := sendEvent(encoder, *mess.Type, mess); err != nil {
			frameworksChans.delete(ID, req.RemoteAddr)
		}
	}
}

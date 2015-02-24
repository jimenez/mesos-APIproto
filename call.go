package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/VoltFramework/volt/mesosproto"
)

type Call map[string]func(call *mesosproto.Call, ID string) error

func generateEventUpdate(taskState mesosproto.TaskState, taskId *mesosproto.TaskID) *mesosproto.Event {
	event_type := mesosproto.Event_UPDATE
	uuid, _ := generateID()
	return &mesosproto.Event{
		Type: &event_type,
		Update: &mesosproto.Event_Update{
			Uuid: bytes.NewBufferString(uuid).Bytes(),
			Status: &mesosproto.TaskStatus{
				TaskId: taskId,
				State:  &taskState,
			},
		},
	}
}

func acceptLaunch(operation *mesosproto.Offer_Operation, ID string) error {
	if operation.Launch != nil && operation.Launch.TaskInfos != nil {
		for _, taskinfo := range operation.Launch.TaskInfos {
			if taskinfo.TaskId != nil {
				f := frameworks.Get(ID)
				event := generateEventUpdate(mesosproto.TaskState_TASK_STARTING, taskinfo.TaskId)
				f.send(event)
				time.Sleep(1)

				event = generateEventUpdate(mesosproto.TaskState_TASK_RUNNING, taskinfo.TaskId)
				f.send(event)
				time.Sleep(1)

				event = generateEventUpdate(mesosproto.TaskState_TASK_FINISHED, taskinfo.TaskId)
				f.send(event)
			}
		}
		return nil
	}
	return fmt.Errorf("malformed launch operation")
}

func acceptReserve(call *mesosproto.Call, ID string) error {
	return nil
}

func acceptUnReserve(call *mesosproto.Call, ID string) error {
	return nil
}

func acceptCreate(call *mesosproto.Call, ID string) error {
	return nil
}

func acceptDestroy(call *mesosproto.Call, ID string) error {
	return nil
}

func Accept(call *mesosproto.Call, ID string) error {

	if call.Accept != nil && call.Accept.Operations != nil {
		for _, operation := range call.Accept.Operations {
			switch operation.Type.String() {
			case "LAUNCH":
				return acceptLaunch(operation, ID)
			case "RESERVE":
				acceptReserve(call, ID)
			case "UNRESERVE":
				acceptUnReserve(call, ID)
			case "CREATE":
				acceptCreate(call, ID)
			case "DESTROY":
				acceptDestroy(call, ID)
			}
		}
	}
	return fmt.Errorf("malformed accept call")
}

// TODO offer declined must not be sent again
func Decline(call *mesosproto.Call, ID string) error {
	if call.Decline != nil && call.Decline.GetOfferIds() != nil {
		for _, offerId := range call.Decline.GetOfferIds() {
			f := frameworks.Get(ID)
			if err := f.deleteOffer(offerId.GetValue()); err != nil {
				return err
			}
		}
		return nil
	}
	return fmt.Errorf("Bad format for Decline Call")
}

func Unregister(call *mesosproto.Call, ID string) error {
	event_type := mesosproto.Event_ERROR
	event := &mesosproto.Event{
		Type: &event_type,
	}
	f := frameworks.Get(ID)
	f.send(event)
	return nil // faire methode sur frmwrkchans
}

func Kill(call *mesosproto.Call, Id string) error {
	if call.Kill != nil && call.Kill.GetTaskId() != nil {
		return nil
	}
	return nil
}

func Acknowledge(call *mesosproto.Call, Id string) error {
	if call.Acknowledge != nil && call.Acknowledge.GetTaskId() != nil {
		// remove task for status map
		return nil
	}
	return nil
}

func Reconcile(call *mesosproto.Call, Id string) error {
	if call.Reconcile != nil && call.Reconcile.GetStatuses() != nil {
		return nil
	}
	return nil
}

func Message(call *mesosproto.Call, Id string) error {
	if call.Message != nil {
		return nil
	}
	return nil
}

func (c *Call) handle(res http.ResponseWriter, req *http.Request) {
	call := mesosproto.Call{}

	if req.Header.Get("Content-Type") == "application/json" {
		if err := json.NewDecoder(req.Body).Decode(&call); err != nil {
			http.Error(res, err.Error(), 500)
			return
		}
	} //else protobuf

	ID := call.FrameworkInfo.Id.GetValue()
	if frameworks.Get(ID) != nil {
		if _, ok := (*c)[call.Type.String()]; !ok {
			// unsuported call
			res.WriteHeader(500)
			return
		}
		if err := (*c)[call.Type.String()](&call, ID); err != nil {
			//check type
			http.Error(res, err.Error(), 500)
			return
		}
		res.WriteHeader(http.StatusAccepted)
		return
	}
	http.Error(res, fmt.Sprintf("Unknown framework %q", ID), 404)
}

func newCall() *Call {
	return &Call{
		"UNREGISTER":  Unregister,
		"DECLINE":     Decline,
		"ACCEPT":      Accept,
		"KILL":        Kill,
		"ACKNOWLEDGE": Acknowledge,
		"RECONCILE":   Reconcile,
		"MESSAGE":     Message,
	}
}

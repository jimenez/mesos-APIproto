package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/VoltFramework/volt/mesosproto"
)

type Calls map[string]func(call *mesosproto.Call, f *Framework) error
type Accepts map[string]func(operation *mesosproto.Offer_Operation, f *Framework) error

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

func acceptLaunch(operation *mesosproto.Offer_Operation, f *Framework) error {
	if operation.Launch != nil && operation.Launch.TaskInfos != nil {
		for _, taskinfo := range operation.Launch.TaskInfos {
			if taskinfo.TaskId != nil {
				event := generateEventUpdate(mesosproto.TaskState_TASK_STARTING, taskinfo.TaskId)
				f.send(event)
				time.Sleep(1)

				event = generateEventUpdate(mesosproto.TaskState_TASK_RUNNING, taskinfo.TaskId)
				f.send(event)
				time.Sleep(1)

				event = generateEventUpdate(mesosproto.TaskState_TASK_FINISHED, taskinfo.TaskId)
				f.send(event)
				f.AddTask(taskinfo.TaskId.GetValue())
			}
		}
		return nil
	}
	return fmt.Errorf("malformed launch operation")
}

func acceptReserve(operation *mesosproto.Offer_Operation, f *Framework) error {
	return nil
}

func acceptUnReserve(operation *mesosproto.Offer_Operation, f *Framework) error {
	return nil
}

func acceptCreate(operation *mesosproto.Offer_Operation, f *Framework) error {
	return nil
}

func acceptDestroy(operation *mesosproto.Offer_Operation, f *Framework) error {
	return nil
}

func Accept(call *mesosproto.Call, f *Framework) error {

	if call.Accept != nil && call.Accept.Operations != nil {
		accepts := newAccepts()
		for _, operation := range call.Accept.Operations {
			if err := (*accepts)[operation.Type.String()](operation, f); err != nil {
				return fmt.Errorf("malformed accept call")
			}
			return nil
		}

	}
	return fmt.Errorf("malformed accept call")

}

func Decline(call *mesosproto.Call, f *Framework) error {
	if call.Decline != nil && call.Decline.GetOfferIds() != nil {
		for _, offerId := range call.Decline.GetOfferIds() {
			if err := f.deleteOffer(offerId.GetValue()); err != nil {
				return err
			}
		}
		return nil
	}
	return fmt.Errorf("Bad format for Decline Call")
}

func Unregister(call *mesosproto.Call, f *Framework) error {
	event_type := mesosproto.Event_ERROR
	event := &mesosproto.Event{
		Type: &event_type,
	}
	f.send(event)
	return nil
}

func Kill(call *mesosproto.Call, f *Framework) error {
	if call.Kill != nil && call.Kill.GetTaskId() != nil {
		return nil
	}
	return nil
}

func Acknowledge(call *mesosproto.Call, f *Framework) error {
	if call.Acknowledge != nil && call.Acknowledge.GetTaskId() != nil {
		// remove task for status map
		return nil
	}
	return nil
}

func Reconcile(call *mesosproto.Call, f *Framework) error {
	if call.Reconcile != nil && call.Reconcile.GetStatuses() != nil {
		return nil
	}
	return nil
}

func Message(call *mesosproto.Call, f *Framework) error {
	if call.Message != nil {
		return nil
	}
	return nil
}

func (c *Calls) handle(res http.ResponseWriter, req *http.Request) {
	call := mesosproto.Call{}
	if req.Header.Get("Content-Type") == "application/json" {
		if err := json.NewDecoder(req.Body).Decode(&call); err != nil {
			http.Error(res, err.Error(), 500)
			return
		}
	} //else protobuf

	ID := call.FrameworkInfo.Id.GetValue()
	if f := frameworks.Get(ID); f != nil {
		if _, ok := (*c)[call.Type.String()]; !ok {
			// unsuported call
			res.WriteHeader(500)
			return
		}
		if err := (*c)[call.Type.String()](&call, f); err != nil {
			//check type
			http.Error(res, err.Error(), 500)
			return
		}
		res.WriteHeader(http.StatusAccepted)
		return
	}
	http.Error(res, fmt.Sprintf("Unknown framework %q", ID), 404)
}

func newCall() *Calls {
	return &Calls{
		"UNREGISTER":  Unregister,
		"DECLINE":     Decline,
		"ACCEPT":      Accept,
		"KILL":        Kill,
		"ACKNOWLEDGE": Acknowledge,
		"RECONCILE":   Reconcile,
		"MESSAGE":     Message,
	}
}

func newAccepts() *Accepts {
	return &Accepts{
		"LAUNCH":    acceptLaunch,
		"CREATE":    acceptCreate,
		"DESTROY":   acceptDestroy,
		"RESERVE":   acceptReserve,
		"UNRESERVE": acceptUnReserve,
	}
}

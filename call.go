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

func handleLaunch(operation *mesosproto.Offer_Operation, ID string) error {
	if operation.Launch != nil && operation.Launch.TaskInfos != nil {
		for _, taskinfo := range operation.Launch.TaskInfos {
			if taskinfo.TaskId != nil {
				event := generateEventUpdate(mesosproto.TaskState_TASK_STARTING, taskinfo.TaskId)
				frameworksChans.send(ID, event)
				time.Sleep(1)

				event = generateEventUpdate(mesosproto.TaskState_TASK_RUNNING, taskinfo.TaskId)
				frameworksChans.send(ID, event)
				time.Sleep(1)

				event = generateEventUpdate(mesosproto.TaskState_TASK_FINISHED, taskinfo.TaskId)
				frameworksChans.send(ID, event)
			}
		}
		return nil
	}
	return fmt.Errorf("malformed launch operation")
}

func handleAcceptReserve(call *mesosproto.Call, ID string) error {
	return nil
}

func handleAcceptUnReserve(call *mesosproto.Call, ID string) error {
	return nil
}

func handleAcceptCreate(call *mesosproto.Call, ID string) error {
	return nil
}

func handleAcceptDestroy(call *mesosproto.Call, ID string) error {
	return nil
}

func handleAccept(call *mesosproto.Call, ID string) error {

	if call.Accept != nil && call.Accept.Operations != nil {
		for _, operation := range call.Accept.Operations {
			switch operation.Type.String() {
			case "LAUNCH":
				return handleLaunch(operation, ID)
			case "RESERVE":
				handleAcceptReserve(call, ID)
			case "UNRESERVE":
				handleAcceptUnReserve(call, ID)
			case "CREATE":
				handleAcceptCreate(call, ID)
			case "DESTROY":
				handleAcceptDestroy(call, ID)
			}
		}
	}
	return fmt.Errorf("malformed accept call")
}

// TODO(jimenez) return proper errors
func handleDecline(call *mesosproto.Call, ID string) error {
	if call.Decline != nil && call.Decline.GetOfferIds() != nil {
		return nil
	}
	return nil
}

func handleUnresgister(call *mesosproto.Call, ID string) error {
	event_type := mesosproto.Event_ERROR
	evnt := &mesosproto.Event{
		Type: &event_type,
	}
	frameworksChans.send(ID, evnt)
	return nil // faire methode sur frmwrkchans
}

func handleReconcile(call *mesosproto.Call, Id string) error {
	if call.Reconcile != nil && call.Reconcile.GetStatuses() != nil {
		return nil
	}
	return nil
}

func handleMessage(call *mesosproto.Call, Id string) error {
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
	if frameworksChans.hasID(ID) {
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
		//		"UNREGISTER":  handleUnregister,
		//		"DECLINE":     handleDecline,
		"ACCEPT": handleAccept,
		//		"KILL":        handleKill,
		//		"ACKNOWLEDGE": handleAknowledge,
		//		"RECONCILE":   handleRencocile,
		//		"MESSAGE":     handleMessage,
	}
}

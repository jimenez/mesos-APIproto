package main

import (
	"bytes"
	"fmt"

	"net"
	"net/http"
	"os"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/VoltFramework/volt/mesosproto"
)

type Calls map[string]func(call *mesosproto.Call, f *Framework) error
type Operations map[string]func(operation *mesosproto.Offer_Operation, f *Framework) error

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

// NOT YET IMPLEMENTED
func acceptReserve(operation *mesosproto.Offer_Operation, f *Framework) error {
	return nil
}

// NOT YET IMPLEMENTED
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
		operations := newOperations()
		for _, operation := range call.Accept.Operations {
			if err := (*operations)[operation.Type.String()](operation, f); err != nil {
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
		var taskState mesosproto.TaskState
		if f.GetTask(call.Kill.GetTaskId().GetValue()) {
			taskState = mesosproto.TaskState_TASK_KILLED
		} else {
			taskState = mesosproto.TaskState_TASK_LOST
		}
		event := generateEventUpdate(taskState, call.Kill.GetTaskId())
		f.send(event)
		return nil
	}
	return fmt.Errorf("Bad format for Decline Call")

}

func Acknowledge(call *mesosproto.Call, f *Framework) error {
	if call.Acknowledge != nil && call.Acknowledge.GetTaskId() != nil {
		return f.deleteTask(call.Acknowledge.GetTaskId().GetValue())
	}
	return fmt.Errorf("Unknown task %q", call.Acknowledge.GetTaskId())
}

func Reconcile(call *mesosproto.Call, f *Framework) error {
	if call.Reconcile != nil && call.Reconcile.GetStatuses() != nil {
		for _, task := range call.Reconcile.GetStatuses() {
			var taskState mesosproto.TaskState
			if f.GetTask(task.TaskId.GetValue()) {
				taskState = mesosproto.TaskState_TASK_RUNNING
			} else {
				taskState = mesosproto.TaskState_TASK_LOST
			}
			event := generateEventUpdate(taskState, task.TaskId)
			f.send(event)
		}
	}
	return nil
}

func Message(call *mesosproto.Call, f *Framework) error {
	if call.Message != nil {
		return nil
	}
	return fmt.Errorf("Missing message call to executor")
}

//TODO(ijimenez): As soon as this get added to the protobuf add this
// func Shutdown(call *mesosproto.Call, f *Framework) error {
// 	if call.Shutdown != nil {
// 		return nil
// 	}
// 	return nil
// }

func (c *Calls) handle(res http.ResponseWriter, req *http.Request) {
	call, encoder, err := decodeCall(res, req)
	if err != nil {
		http.Error(res, err.Error(), 400)
		return
	}
	if call.GetFrameworkInfo().GetId() != nil {
		events(call.GetFrameworkInfo(), encoder, res)
		return
	}

	res.Header().Set("Content-Type", "application/json")
	name, err := os.Hostname()
	if err != nil {
		log.Fatalf("Failed to get hostname: %+v", err)
	}

	addrs, err := net.LookupHost(name)
	if err != nil {
		log.Fatalf("Failed to get address for hostname %q: %+v", name, err)
	}
	var localhost string
	for _, addr := range addrs {
		if !strings.HasPrefix(addr, "127") {
			localhost = addr
		}
	}
	res.Header().Set("Host", fmt.Sprintf("%s:%#v", localhost, *port))

	ID := call.FrameworkInfo.Id.GetValue()
	if f := frameworks.Get(ID); f != nil {
		if _, ok := (*c)[(*call).Type.String()]; !ok {
			// unsuported call
			res.WriteHeader(500)
			return
		}
		if err := (*c)[(*call).Type.String()](call, f); err != nil {
			//check type
			log.Error(err)
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
		//		"SHUTDOWN":    Shutdown, see TODO for shutdown
	}
}

func newOperations() *Operations {
	return &Operations{
		"LAUNCH":    acceptLaunch,
		"CREATE":    acceptCreate,
		"DESTROY":   acceptDestroy,
		"RESERVE":   acceptReserve,
		"UNRESERVE": acceptUnReserve,
	}
}

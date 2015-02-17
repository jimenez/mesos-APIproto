package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/VoltFramework/volt/mesosproto"
)

func call(res http.ResponseWriter, req *http.Request) {
	call := mesosproto.Call{}
	if req.Header.Get("Content-Type") == "application/json" {
		if err := json.NewDecoder(req.Body).Decode(&call); err != nil {
			http.Error(res, err.Error(), 500)
			return
		}
	} //else

	ID := call.FrameworkInfo.Id.GetValue()
	if frameworksChans.hasID(ID) {
		var event *mesosproto.Event

		switch call.Type.String() {
		case "LAUNCH":
			if call.Launch != nil && call.Launch.TaskInfos != nil {
				for _, taskinfo := range call.Launch.TaskInfos {
					if taskinfo.TaskId != nil {
						event_type := mesosproto.Event_UPDATE
						task_state := mesosproto.TaskState_TASK_STARTING
						event = &mesosproto.Event{
							Type: &event_type,
							Update: &mesosproto.Event_Update{
								Uuid: bytes.NewBufferString("0120032100120452").Bytes(),
								Status: &mesosproto.TaskStatus{
									TaskId: taskinfo.TaskId,
									State:  &task_state,
								},
							},
						}
						frameworksChans.send(ID, event)
						time.Sleep(1)
						//send mesosproto.TaskState_TASK_RUNNING
						event_type = mesosproto.Event_UPDATE
						task_state = mesosproto.TaskState_TASK_RUNNING
						event = &mesosproto.Event{
							Type: &event_type,
							Update: &mesosproto.Event_Update{
								Uuid: bytes.NewBufferString("0120032100120452").Bytes(),
								Status: &mesosproto.TaskStatus{
									TaskId: taskinfo.TaskId,
									State:  &task_state,
								},
							},
						}
						frameworksChans.send(ID, event)
						time.Sleep(1)
						//mesosproto.TaskState_TASK_FINNISHED
						event_type = mesosproto.Event_UPDATE
						task_state = mesosproto.TaskState_TASK_FINISHED
						event = &mesosproto.Event{
							Type: &event_type,
							Update: &mesosproto.Event_Update{
								Uuid: bytes.NewBufferString("0120032100120452").Bytes(),
								Status: &mesosproto.TaskStatus{
									TaskId: taskinfo.TaskId,
									State:  &task_state,
								},
							},
						}
						frameworksChans.send(ID, event)

					}
				}

			}

		}

		res.WriteHeader(http.StatusAccepted)
		return
	}

	http.Error(res, fmt.Sprintf("Unknown framework %q", ID), 404)
}

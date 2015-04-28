package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/gogo/protobuf/proto"
	"github.com/jimenez/mesos-APIproto/mesosproto"
)

func decodeCall(res http.ResponseWriter, req *http.Request) (*mesosproto.Call, icoder, error) {

	buf, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return nil, nil, err
	}
	req.Body.Close()
	call := mesosproto.Call{}
	if req.Header.Get("Content-Type") == "application/json" {
		encoder := json.NewEncoder(NewWriteFlusher(res))
		if err = json.Unmarshal(buf, &call); err != nil {
			return nil, nil, err
		}
		return &call, encoder, nil
	}
	encoder := &protobufEncoder{w: res}
	if err = proto.Unmarshal(buf, &call); err != nil {
		return nil, nil, err
	}
	return &call, encoder, nil
}

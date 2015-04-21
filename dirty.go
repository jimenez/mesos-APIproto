package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/VoltFramework/volt/mesosproto"
	"github.com/gogo/protobuf/proto"
)

func decodeCallorFrameworkInfo(res http.ResponseWriter, req *http.Request) (*mesosproto.FrameworkInfo, *mesosproto.Call, icoder, error) {
	buf, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return nil, nil, nil, err
	}
	req.Body.Close()
	frameworkInfo := mesosproto.FrameworkInfo{}
	call := mesosproto.Call{}
	if req.Header.Get("Content-Type") == "application/json" {
		encoder := json.NewEncoder(NewWriteFlusher(res))
		if err := json.Unmarshal(buf, &frameworkInfo); err != nil || frameworkInfo.GetName() == "" {
			if err = json.Unmarshal(buf, &call); err != nil {
				return nil, nil, nil, err
			}
			return nil, &call, encoder, nil
		}
		return &frameworkInfo, nil, encoder, nil
	}
	encoder := &protobufEncoder{w: res}
	if err := proto.Unmarshal(buf, &frameworkInfo); err != nil {
		if err = proto.Unmarshal(buf, &call); err != nil {
			return nil, nil, nil, err
		}
		return nil, &call, encoder, nil
	}
	return &frameworkInfo, nil, encoder, nil
}

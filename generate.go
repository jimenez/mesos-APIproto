package main

import (
	crand "crypto/rand"
	"encoding/hex"
	"math/rand"

	"github.com/VoltFramework/volt/mesosproto"
)

func generateID() (string, error) {
	id := make([]byte, 6)
	n, err := crand.Read(id)
	if n != len(id) || err != nil {
		return "", err
	}
	return hex.EncodeToString(id), nil
}

func createRangeResource(name string, begin uint64) *mesosproto.Resource {
	rType := mesosproto.Resource_STATIC
	disk := mesosproto.Resource_DiskInfo{}
	end := begin + 10000
	return &mesosproto.Resource{
		Name: &name,
		Type: mesosproto.Value_RANGES.Enum(),
		Ranges: &mesosproto.Value_Ranges{
			Range: []*mesosproto.Value_Range{
				&mesosproto.Value_Range{
					Begin: &begin,
					End:   &end,
				},
			},
		},
		ReservationType: &rType,
		Disk:            &disk,
	}
}

func createScalarResource(name string, value float64) *mesosproto.Resource {
	rType := mesosproto.Resource_DYNAMIC
	disk := mesosproto.Resource_DiskInfo{}
	return &mesosproto.Resource{
		Name:            &name,
		Type:            mesosproto.Value_SCALAR.Enum(),
		Scalar:          &mesosproto.Value_Scalar{Value: &value},
		ReservationType: &rType,
		Disk:            &disk,
	}
}

func resources() []*mesosproto.Resource {
	return []*mesosproto.Resource{
		createScalarResource("cpus", float64(rand.Intn(32))),
		createScalarResource("mem", float64(rand.Intn(32)*1024)),
		createScalarResource("disk", float64(rand.Intn(100000))),
		createRangeResource("ports", uint64(1024+rand.Intn(10000))),
	}
}

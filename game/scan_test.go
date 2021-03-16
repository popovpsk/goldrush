package game

import (
	"fmt"
	"goldrush/api"
	"goldrush/metrics"
	"testing"
)

func TestScan(t *testing.T) {

	q := make(api.PostLicenseRequest, 0, 10)
	for i := 0; i < 10; i++ {
		q = append(q, uint32(i))
	}
	w := &q
	q.MarshalJSON()
	ss, _ := w.MarshalJSON()
	fmt.Println(string(ss))

	d := NewDigger(nil, metrics.NewMetricsSvc())

	r := &api.ExploreResponse{
		Area: api.Area{
			PosX:  15,
			PosY:  15,
			SizeX: 13,
			SizeY: 11,
		},
		Amount: 8,
	}
	d.clearSector(&r.Area, 8)

}

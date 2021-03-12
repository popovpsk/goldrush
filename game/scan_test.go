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
	return
	d := NewDigger(nil, metrics.NewMetricsSvc())
	d.bSearch(&api.ExploreResponse{
		Area: api.Area{
			PosX:  100,
			PosY:  100,
			SizeX: 24,
			SizeY: 25,
		},
		Amount: 500,
	})
}

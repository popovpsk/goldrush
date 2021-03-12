package game

import (
	"fmt"
	"goldrush/api"
	"goldrush/utils"
	"sync/atomic"
	"time"
)

var allScan int32
var allSearch int32

func (d *Digger) scan(area *api.Area) {
	res := &api.ExploreResponse{}
	t := time.Now()
	d.apiClient.Explore(area, res)
	d.metrics.Add(fmt.Sprintf("explore %v", area.SizeX*area.SizeY), time.Since(t))
	atomic.AddInt32(&allScan, int32(res.Amount))
	//d.bSearch(res)
	d.areaQueue.Push(res)
}

func (d *Digger) bSearch(zone *api.ExploreResponse) {
	size := zone.Area.SizeX * zone.Area.SizeY
	if throttle && size >= 1250 {
		return
	}

	if size <= 100 {
		d.clearSector(&zone.Area, zone.Amount)
		return
	}

	res2 := &api.ExploreResponse{
		Area: zone.Area,
	}
	req := zone.Area

	divX := req.SizeX >= req.SizeY
	if divX {
		req.SizeX = req.SizeX / 2
		res2.Area.PosX += zone.Area.SizeX / 2
		res2.Area.SizeX = res2.Area.SizeX - req.SizeX

	} else {
		req.SizeY = req.SizeY / 2
		res2.Area.PosY += zone.Area.SizeY / 2
		res2.Area.SizeY = res2.Area.SizeY - req.SizeY
	}

	res := &api.ExploreResponse{}
	t := time.Now()
	d.apiClient.Explore(&req, res)
	d.metrics.Add(fmt.Sprintf("explore %v", req.SizeX*req.SizeY), time.Since(t))
	d.areaQueue.Push(res)

	res2.Amount = zone.Amount - res.Amount

	if zone.Area.SizeX*zone.Area.SizeY == sizeX*sizeY {
		atomic.AddInt32(&allSearch, int32(res.Amount+res2.Amount))
	}
	d.areaQueue.Push(res2)
}

func (d *Digger) clearSector(area *api.Area, amount int) {
	for x := area.PosX; x < area.PosX+area.SizeX; x++ {
		for y := area.PosY; y < area.PosY+area.SizeY; y++ {
			req := &api.Area{
				PosX:  x,
				PosY:  y,
				SizeX: 1,
				SizeY: 1,
			}
			res := &api.ExploreResponse{}
			t := time.Now()
			d.apiClient.Explore(req, res)
			d.metrics.Add("explore 1", time.Since(t))
			if res.Amount > 0 {
				d.pointQueue.Push(utils.DigPoint{X: int32(x), Y: int32(y), Amount: int32(res.Amount)})
				amount -= res.Amount
				if amount <= 0 {
					return
				}
			}
		}
	}
}

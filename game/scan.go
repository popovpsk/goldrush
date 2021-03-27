package game

import (
	"fmt"
	"goldrush/types"
	"time"
)

func (d *Digger) scan(area *types.Area) {
	res := &types.ExploreResponse{}
	t := time.Now()
	d.apiClient.Explore(area, res)
	d.metrics.Add(fmt.Sprintf("explore %v", area.SizeX*area.SizeY), time.Since(t))
	d.areaQueue.Push(res)
}

func (d *Digger) bSearch(zone *types.ExploreResponse) {
	res := &types.ExploreResponse{}

	square := zone.Area.Size()
	if square <= 16 {
		d.clearSector(&zone.Area, zone.Amount)
		return
	}

	res2 := types.ExploreResponse{
		Area: *d.divideArea(&zone.Area),
	}

	t := time.Now()
	d.apiClient.Explore(&zone.Area, res)
	d.metrics.Add(fmt.Sprintf("explore %v", zone.Area.Size()), time.Since(t))
	d.areaQueue.Push(res)

	res2.Amount = zone.Amount - res.Amount
	d.areaQueue.Push(&res2)
}

func (d *Digger) clearSector(area *types.Area, amount int32) {
	req := &types.Area{
		SizeX: 1,
		SizeY: 1,
	}
	res := &types.ExploreResponse{}

	if area.SizeX < area.SizeY {
		for x := area.PosX; x < area.PosX+area.SizeX; {
			p := float32(amount) / float32(area.Size())
			for y := area.PosY; y < area.PosY+area.SizeY; y++ {
				req.PosX = x
				req.PosY = y
				t := time.Now()
				d.apiClient.Explore(req, res)
				d.metrics.Add("explore 1", time.Since(t))
				if res.Amount > 0 {
					d.dig(x, y, res.Amount)
					amount -= res.Amount
					if amount <= 0 {
						return
					}
				}
			}
			x++

			tmp := types.Area{
				PosX:  area.PosX + (x - area.PosX),
				SizeX: area.SizeX - (x - area.PosX),
				PosY:  area.PosY,
				SizeY: area.SizeY,
			}
			if tmp.SizeX == 0 || tmp.Size() <= 4 || p <= float32(tmp.Size())/float32(amount) {
				continue
			} else {
				d.areaQueue.Push(&types.ExploreResponse{Area: tmp, Amount: amount})
				return
			}
		}
	} else {
		for y := area.PosY; y < area.PosY+area.SizeY; {
			p := float32(amount) / float32(area.Size())
			for x := area.PosX; x < area.PosX+area.SizeX; x++ {
				req.PosX = x
				req.PosY = y
				d.apiClient.Explore(req, res)
				if res.Amount > 0 {
					d.dig(x, y, res.Amount)
					amount -= res.Amount
					if amount <= 0 {
						return
					}
				}
			}
			y++
			tmp := types.Area{
				PosX:  area.PosX,
				SizeX: area.SizeX,
				PosY:  area.PosY + (y - area.PosY),
				SizeY: area.SizeY - (y - area.PosY),
			}
			if tmp.SizeY == 0 || tmp.Size() <= 4 || p <= float32(amount)/float32(area.Size()) {
				continue
			} else {
				d.areaQueue.Push(&types.ExploreResponse{Area: tmp, Amount: amount})
				return
			}
		}
	}
}

func (d *Digger) divideArea(a *types.Area) *types.Area {
	b := *a
	if a.SizeX >= a.SizeY {
		a.SizeX = a.SizeX / 2
		b.PosX += a.SizeX
		b.SizeX -= a.SizeX
	} else {
		a.SizeY = a.SizeY / 2
		b.PosY += a.SizeY
		b.SizeY -= a.SizeY
	}
	return &b
}

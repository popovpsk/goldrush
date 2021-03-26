package game

import (
	"fmt"
	"goldrush/api"
	"goldrush/datastruct/pointqueue"
	"time"
)

func (d *Digger) digWork(state *int) {
	p := d.pointQueue.Peek()
	d.dig(p.X, p.Y, p.Amount)
}

func (d *Digger) check(p pointqueue.DigPoint) {
	res := &api.ExploreResponse{}
	d.apiClient.Explore(&api.Area{
		PosX:  int(p.X),
		PosY:  int(p.Y),
		SizeX: 1,
		SizeY: 1,
	}, res)
	if res.Amount == 0 {
		fmt.Printf("%v:%v|", p.X, p.Y)
	}
	return
}

func (d *Digger) dig(x, y, amount int32) {
	depth := 1
	var license *api.License

	for depth <= 10 && amount > 0 {
		t1 := time.Now()
		if license == nil {
			license = d.getLicense()
		} else if license.DigUsed >= license.DigAllowed {
			d.returnLicense(license)
			license = d.getLicense()
		}
		d.metrics.Add("getLicense", time.Since(t1))

		req := &api.DigRequest{LicenseID: license.ID, PosX: x, PosY: y, Depth: depth}
		res := &api.Treasures{}
		t := time.Now()
		ok := d.apiClient.Dig(req, res)
		el := time.Since(t)
		d.metrics.Add("dig", el)
		license.DigUsed++
		depth++

		if ok {
			for _, r := range *res {
				s := &api.Payment{}
				t2 := time.Now()
				d.apiClient.Cash(r, s)
				d.metrics.Add("cash", time.Since(t2))
				d.bank.Store(*s)
			}
			amount--
		}
	}
	d.returnLicense(license)
}

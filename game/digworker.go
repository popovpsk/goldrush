package game

import (
	"fmt"
	"goldrush/api"
	"sync/atomic"
	"time"
)

func (d *Digger) digWork(state *int) {
	p := d.pointQueue.Peek()
	d.dig(p.X, p.Y, p.Amount)
}

var qqq = make([]int32, 10)
var qqq2 = make([]int32, 10)

func EndLog() {
	for i, v := range qqq {
		fmt.Printf("%v:%v; ", i+1, qqq2[i]/v)
	}
}

func (d *Digger) dig(x, y, amount int32) {
	depth := 1
	var license *api.License

	for depth <= 10 && amount > 0 {
		if license == nil {
			license = d.getLicense()
		} else if license.DigUsed >= license.DigAllowed {
			d.returnLicense(license)
			license = d.getLicense()
		}
		req := &api.DigRequest{LicenseID: license.ID, PosX: x, PosY: y, Depth: depth}
		res := &api.Treasures{}
		t := time.Now()
		ok := d.apiClient.Dig(req, res)
		el := time.Since(t)
		d.metrics.Add("dig", el)
		license.DigUsed++
		depth++

		atomic.AddInt32(&qqq[depth-2], 1)
		atomic.AddInt32(&qqq2[depth-2], int32(el.Milliseconds()))

		if ok {
			d.metrics.AddInt("array gold len", len(*res))

			money := 0
			for _, r := range *res {
				s := &api.Payment{}
				d.apiClient.Cash(r, s)
				money += len(*s)
				d.bank.Store(*s)
			}

			d.metrics.AddInt(fmt.Sprintf("DP:%v", depth), money)
			d.goldPot.Store(*res)
			amount--

		}
	}
	d.returnLicense(license)
}

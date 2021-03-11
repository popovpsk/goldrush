package game

import (
	"goldrush/api"
	"time"
)

func (d *Digger) startDigWorker() {
	go func() {
		for {
			p := d.pointQueue.Peek()
			d.dig(p.X, p.Y, p.Amount)
		}
	}()
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
		d.metrics.Add("dig", time.Since(t))
		license.DigUsed++
		depth++

		if ok {
			d.exchangeCash(*res)
			amount--
		}
	}
	d.returnLicense(license)
}

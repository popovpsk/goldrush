package game

import (
	"goldrush/types"
	"time"
)

func (d *Digger) dig(x, y, amount int32) {
	var depth int32 = 1
	var license *types.License

	for depth <= 10 && amount > 0 {
		t1 := time.Now()
		if license == nil {
			license = d.getLicense()
		} else if license.DigUsed >= license.DigAllowed {
			d.returnLicense(license)
			license = d.getLicense()
		}
		d.metrics.Add("getLicense", time.Since(t1))

		req := &types.DigRequest{LicenseID: license.ID, PosX: x, PosY: y, Depth: depth}
		res := &types.Treasures{}
		t := time.Now()
		ok := d.apiClient.Dig(req, res)
		el := time.Since(t)
		d.metrics.Add("dig", el)
		license.DigUsed++
		depth++

		if ok {
			for _, r := range *res {
				s := &types.Payment{}
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

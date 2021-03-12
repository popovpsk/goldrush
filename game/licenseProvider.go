package game

import (
	"goldrush/api"
	"sync/atomic"
	"time"
)

func (d *Digger) getLicense() *api.License {
	return <-d.licenses
}

func (d *Digger) returnLicense(license *api.License) {
	if license.DigUsed >= license.DigAllowed {
		atomic.AddInt32(d.activeLicenses, -1)
		select {
		case d.licWakeCh <- struct{}{}:
		default:
		}
	} else {
		d.licenses <- license
	}
}

func (d *Digger) startGetLicensesWorker() {
	go func() {
		for {
			if atomic.LoadInt32(d.activeLicenses) >= 10 {
				<-d.licWakeCh
			}
			license := &api.License{}
			t := time.Now()
			if d.bank.Count() > 2000 && atomic.LoadInt32(d.activeLicenses) <= 4 /*&& atomic.LoadInt32(&price) < 10000 */ {
				d.getPaidLicenses(license)
			} else {
				d.apiClient.PostLicenses(nil, license)
			}
			d.metrics.Add("license", time.Since(t))
			atomic.AddInt32(d.activeLicenses, 1)
			d.licenses <- license
		}
	}()
}

func (d *Digger) getPaidLicenses(license *api.License) {
	var p int32 = 22
	if coins, ok := d.bank.Get(p); ok {
		d.apiClient.PostLicenses(coins, license)
		d.metrics.AddInt("allowed", license.DigAllowed)
	} else {
		d.apiClient.PostLicenses(nil, license)
	}
}

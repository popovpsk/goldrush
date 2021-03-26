package game

import (
	"goldrush/api"
	"goldrush/datastruct/foreman"
	"sync/atomic"
	"time"
)

func (d *Digger) getLicense() *api.License {
	return <-d.licenses
}

func (d *Digger) returnLicense(license *api.License) {
	if license.DigUsed >= license.DigAllowed {
		atomic.AddInt32(&d.activeLicenses, -1)
		select {
		case d.licWakeCh <- struct{}{}:
		default:
		}
	} else {
		d.licenses <- license
	}
}

func (d *Digger) licensesWork(state *int) {
	if atomic.LoadInt32(&d.activeLicenses) >= 10 {
		<-d.licWakeCh
	}
	if *state == foreman.Stopped {
		return
	}
	license := &api.License{}
	t := time.Now()
	var p int32 = 21
	if coins, ok := d.bank.Get(p); ok {
		d.apiClient.PostLicenses(coins, license)
		d.metrics.AddInt("allowed", license.DigAllowed)
	} else {
		d.metrics.AddInt("free", 1)
		d.apiClient.PostLicenses(nil, license)
	}
	d.metrics.Add("license", time.Since(t))
	atomic.AddInt32(&d.activeLicenses, 1)
	d.licenses <- license
}

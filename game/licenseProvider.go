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
			d.apiClient.PostLicenses(nil, license)
			d.metrics.Add("license", time.Since(t))
			atomic.AddInt32(d.activeLicenses, 1)
			d.licenses <- license
		}
	}()
}

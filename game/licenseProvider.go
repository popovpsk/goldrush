package game

import (
	"goldrush/api"
	"goldrush/datastruct/bank"
	"goldrush/types"
	"sync/atomic"
)

const licenseCost = 21

type LicenseProvider struct {
	bank           *bank.Bank
	licenses       chan *types.License
	activeLicenses int32
	licWakeCh      chan struct{}
	apiClient      *api.Client
}

func NewLicenseProvider(bank *bank.Bank, client *api.Client) *LicenseProvider {
	return &LicenseProvider{
		apiClient:      client,
		bank:           bank,
		activeLicenses: 0,
		licenses:       make(chan *types.License, 10),
		licWakeCh:      make(chan struct{}),
	}
}

func (p *LicenseProvider) GetLicense() *types.License {
	return <-p.licenses
}

func (p *LicenseProvider) ReturnLicense(license *types.License) {
	if license.DigUsed >= license.DigAllowed {
		atomic.AddInt32(&p.activeLicenses, -1)
		select {
		case p.licWakeCh <- struct{}{}:
		default:
		}
	} else {
		p.licenses <- license
	}
}

func (p *LicenseProvider) LicensesWork(state *int) {
	if atomic.LoadInt32(&p.activeLicenses) >= 10 {
		<-p.licWakeCh
	}
	if *state == Stopped {
		return
	}
	license := &types.License{}
	if coins, ok := p.bank.Get(licenseCost); ok {
		p.apiClient.PostLicenses(coins, license)
	} else {
		p.apiClient.PostLicenses(nil, license)
	}
	atomic.AddInt32(&p.activeLicenses, 1)
	p.licenses <- license
}

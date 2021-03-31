package game

import (
	"goldrush/api"
	"goldrush/datastruct/areaqueue"
	"goldrush/datastruct/bank"
	"goldrush/metrics"
	"goldrush/types"
	"goldrush/utils"
	"time"
)

type Digger struct {
	apiClient *api.Client
	bank      *bank.Bank
	areaQueue *areaqueue.AreaQueue
	metrics   *metrics.Svc
	foreman   *Foreman
	licenses  *LicenseProvider

	areaCh chan *types.Area
}

func NewDigger(licenseProvider *LicenseProvider, client *api.Client, metrics *metrics.Svc, bank *bank.Bank) *Digger {
	d := &Digger{
		apiClient: client,
		licenses:  licenseProvider,
		bank:      bank,
		metrics:   metrics,
		areaCh:    make(chan *types.Area, 50),
		areaQueue: areaqueue.NewAreaQueue(),
	}
	f := New()
	f.AddWorker(firstScanWorker, d.firstScanWork)
	f.AddWorker(zoneScanWorker, d.scanWork)
	f.AddWorker(licensesWorker, d.licenses.LicensesWork)
	d.foreman = f
	return d
}

const (
	sizeX = 100
	sizeY = 250
)

const (
	firstScanWorker = iota
	zoneScanWorker
	licensesWorker
)

func (d *Digger) Start() {
	go d.divideGameArea()
	d.foreman.Start(licensesWorker, 1)

	d.foreman.Start(firstScanWorker, 2)
	d.foreman.Start(zoneScanWorker, 6)

	<-utils.WaitGameTime(time.Minute * 8)
	d.foreman.ChangeState(zoneScanWorker, Slow, 6)
}

func (d *Digger) divideGameArea() {
	for x := 0; x < 3500; x += sizeX {
		for y := 0; y < 3500; y += sizeY {
			d.areaCh <- &types.Area{PosX: int32(x), PosY: int32(y), SizeX: sizeX, SizeY: sizeY}
		}
	}
	close(d.areaCh)
}

func (d *Digger) firstScanWork(state *int) {
	a, ok := <-d.areaCh
	if ok {
		res := &types.ExploreResponse{}
		d.apiClient.Explore(a, res)
		d.areaQueue.Push(res)
	} else {
		*state = Stopped
	}
}

func (d *Digger) scanWork(state *int) {
	req := d.areaQueue.Peek()
	if *state == Slow && req.Area.Size() > 150 {
		return
	} else {
		d.search(req)
	}
}

func (d *Digger) dig(x, y, amount int32) {
	var depth int32 = 1
	var license *types.License

	request := &types.DigRequest{}
	response := make(types.Treasures, 0, 1)

	for depth <= 10 && amount > 0 {
		if license == nil {
			license = d.licenses.GetLicense()
		} else if license.DigUsed >= license.DigAllowed {
			d.licenses.ReturnLicense(license)
			license = d.licenses.GetLicense()
		}

		request.LicenseID = license.ID
		request.PosX = x
		request.PosY = y
		request.Depth = depth

		t := time.Now()
		ok := d.apiClient.Dig(request, &response)
		el := time.Since(t)
		d.metrics.Add("dig", el)
		license.DigUsed++
		depth++

		if ok {
			for _, r := range response {
				s := &types.Payment{}
				t2 := time.Now()
				d.apiClient.Cash(r, s)
				d.metrics.Add("cash", time.Since(t2))
				d.bank.Store(*s)
			}
			amount--
			response = response[:0]
		}
	}
	d.licenses.ReturnLicense(license)
}

func (d *Digger) search(zone *types.ExploreResponse) {
	res := &types.ExploreResponse{}

	if zone.Area.Size() <= 16 {
		d.exploreSector(&zone.Area, zone.Amount)
		return
	}

	res2 := types.ExploreResponse{Area: *d.divideArea(&zone.Area)}

	d.apiClient.Explore(&zone.Area, res)
	d.areaQueue.Push(res)

	res2.Amount = zone.Amount - res.Amount
	d.areaQueue.Push(&res2)
}

func (d *Digger) exploreSector(area *types.Area, amount int32) {
	req := &types.Area{
		SizeX: 1,
		SizeY: 1,
	}
	res := &types.ExploreResponse{}

	if area.SizeX < area.SizeY {
		for x := area.PosX; x < area.PosX+area.SizeX; {
			p := float32(amount) / float32(area.Size())
			//y line
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
			if tmp.SizeX == 0 || p <= float32(tmp.Size())/float32(amount) {
				continue
			} else {
				d.areaQueue.Push(&types.ExploreResponse{Area: tmp, Amount: amount})
				return
			}
		}
	} else {
		for y := area.PosY; y < area.PosY+area.SizeY; {
			p := float32(amount) / float32(area.Size())
			//x line
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
			if tmp.SizeY == 0 || p <= float32(amount)/float32(area.Size()) {
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

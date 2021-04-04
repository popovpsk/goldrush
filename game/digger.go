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
	sizeX   = 100
	sizeY   = 100
	workers = 10
)

const (
	firstScanWorker = iota
	zoneScanWorker
	licensesWorker
)

func (d *Digger) Start() {
	go d.divideGameArea()
	d.foreman.Start(firstScanWorker, 3)
	d.foreman.Start(licensesWorker, 1)

	d.foreman.Start(zoneScanWorker, 10)
	<-utils.WaitGameTime(time.Minute*9 + time.Second*55)
	d.foreman.ChangeState(zoneScanWorker, Slow, workers)
}

func (d *Digger) divideGameArea() {
	for x := 0; x < 3500; x += sizeX {
		for y := 0; y < 3500; y += sizeY {
			area := AcquireArea()
			area.PosX = int32(x)
			area.PosY = int32(y)
			area.SizeX = sizeX
			area.SizeY = sizeY
			d.areaCh <- area
		}
	}
	close(d.areaCh)
}

func (d *Digger) firstScanWork(state *int) {
	a, ok := <-d.areaCh
	if ok {
		res := AcquireExploredArea()
		d.apiClient.Explore(a, res)
		d.areaQueue.Push(res)
		ReleaseArea(a)
	} else {
		*state = Stopped
	}
}

func (d *Digger) scanWork(state *int) {
	req := d.areaQueue.Peek()
	if *state == Slow && req.Area.Size() > 50 {
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

		ok := d.apiClient.Dig(request, &response)
		license.DigUsed++
		depth++

		if ok {
			for _, r := range response {
				s := &types.Payment{}
				d.apiClient.Cash(r, s)
				d.bank.Store(*s)
			}
			amount--
			response = response[:0]
		}
	}
	d.licenses.ReturnLicense(license)
}

func (d *Digger) search(zone *types.ExploredArea) {
	if zone.Area.Size() <= 4 {
		d.exploreSector(zone)
		return
	}
	a1 := AcquireExploredArea()
	a2 := AcquireExploredArea()
	a2.Area = d.divideArea(&zone.Area)

	d.apiClient.Explore(&zone.Area, a1)
	d.pushIntoQueue(a1)

	a2.Amount = zone.Amount - a1.Amount
	d.pushIntoQueue(a2)
	ReleaseExploredArea(zone)
}

func (d *Digger) pushIntoQueue(area *types.ExploredArea) {
	if area.Amount > 0 {
		d.areaQueue.Push(area)
	} else {
		ReleaseExploredArea(area)
	}
}

func (d *Digger) exploreSector(explored *types.ExploredArea) {
	area := explored.Area
	amount := explored.Amount
	ReleaseExploredArea(explored)
	req := AcquireArea()
	defer ReleaseArea(req)
	res := AcquireExploredArea()
	defer ReleaseExploredArea(res)

	req.SizeX = 1
	req.SizeY = 1
	if area.SizeX < area.SizeY {
		for x := area.PosX; x < area.PosX+area.SizeX; {
			p := float32(amount) / float32(area.Size())
			//y line
			for y := area.PosY; y < area.PosY+area.SizeY; y++ {
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
				a := AcquireExploredArea()
				a.Area = tmp
				a.Amount = amount
				d.areaQueue.Push(a)
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
				d.areaQueue.Push(&types.ExploredArea{Area: tmp, Amount: amount})
				return
			}
		}
	}
}

func (d *Digger) divideArea(a *types.Area) types.Area {
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
	return b
}

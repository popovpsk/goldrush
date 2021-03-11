package game

import (
	"fmt"
	"goldrush/api"
	"goldrush/metrics"
	"goldrush/utils"
	"sync/atomic"
	"time"
)

type Digger struct {
	apiClient      *api.Client
	licenses       chan *api.License
	activeLicenses *int32
	cashCh         chan string
	bank           *Bank
	licWakeCh      chan struct{}
	digWorkerCnt   *int32
	pointQueue     *utils.PointQueue
	metrics        *metrics.Svc
	areaCh         chan *api.Area
	areaQueue      *utils.AreaQueue
}

func NewDigger(client *api.Client, metrics *metrics.Svc) *Digger {
	var l int32
	return &Digger{
		apiClient:      client,
		licenses:       make(chan *api.License, 10),
		cashCh:         make(chan string, 1000),
		licWakeCh:      make(chan struct{}),
		bank:           NewBank(),
		activeLicenses: &l,
		pointQueue:     utils.NewPointQueue(),
		metrics:        metrics,
		areaCh:         make(chan *api.Area, 4),
		areaQueue:      utils.NewAreaQueue(),
	}
}

var all int32 = 0

func (d *Digger) Start() {
	d.startGetLicensesWorker()
	d.startGetLicensesWorker()
	d.startGetLicensesWorker()

	d.startDigWorker()
	d.startDigWorker()

	d.startExchangeCashWorker()
	d.startExchangeCashWorker()

	d.startExploreWorkers()

	d.search()
}

const sizeX = 100
const sizeY = 100

func (d *Digger) search() {
	for x := 0; x < 3500; x += sizeX {
		for y := 0; y < 3500; y += sizeY {
			d.areaCh <- &api.Area{PosX: x, PosY: y, SizeX: sizeX, SizeY: sizeY}
		}
	}
	close(d.areaCh)
	d.startGetLicensesWorker()
	d.startGetLicensesWorker()
	d.startGetLicensesWorker()
	d.startGetLicensesWorker()
	d.startGetLicensesWorker()
	d.startGetLicensesWorker()
}

func (d *Digger) startExploreWorkers() {
	for i := 0; i < 2; i++ {
		d.startScan()
	}
	for i := 0; i < 4; i++ {
		d.startScanB()
	}
}

func (d *Digger) startScan() {
	go func() {
		for {
			a, ok := <-d.areaCh
			if ok {
				d.scan(a)
			} else {
				d.startScanB()
				return
			}
		}
	}()
}

func (d *Digger) startScanB() {
	go func() {
		for {
			d.bSearch(d.areaQueue.Peek())
		}
	}()
}

func (d *Digger) scan(area *api.Area) {
	res := &api.ExploreResponse{}

	t := time.Now()
	d.apiClient.Explore(area, res)
	d.metrics.Add(fmt.Sprintf("explore %v", area.SizeX*area.SizeY), time.Since(t))
	atomic.AddInt32(&all, int32(res.Amount))
	d.areaQueue.Push(res)
}

func (d *Digger) bSearch(zone *api.ExploreResponse) {
	if zone.Area.SizeX*zone.Area.SizeY <= 40 {
		d.clearSector(&zone.Area, zone.Amount)
		return
	}

	req := &api.Area{
		PosX:  zone.Area.PosX,
		PosY:  zone.Area.PosY,
		SizeX: zone.Area.SizeX / 2,
		SizeY: zone.Area.SizeY / 2,
	}
	res := &api.ExploreResponse{}

	t := time.Now()
	d.apiClient.Explore(req, res)
	d.metrics.Add(fmt.Sprintf("explore %v", req.SizeX*req.SizeY), time.Since(t))
	d.areaQueue.Push(res)
	res2 := &api.ExploreResponse{
		Area: api.Area{
			PosX:  zone.Area.PosX + zone.Area.SizeX/2,
			PosY:  zone.Area.PosY + zone.Area.SizeY/2,
			SizeX: zone.Area.SizeX / 2,
			SizeY: zone.Area.SizeX / 2,
		},
		Amount: zone.Amount - res.Amount,
	}
	d.areaQueue.Push(res2)
}

func (d *Digger) clearSector(area *api.Area, amount int) {
	for x := area.PosX; x < area.PosX+area.SizeX; x++ {
		for y := area.PosY; y < area.PosY+area.SizeY; y++ {
			req := &api.Area{
				PosX:  x,
				PosY:  y,
				SizeX: 1,
				SizeY: 1,
			}
			res := &api.ExploreResponse{}
			t := time.Now()
			d.apiClient.Explore(req, res)
			d.metrics.Add("explore 1", time.Since(t))
			if res.Amount > 0 {
				d.pointQueue.Push(utils.DigPoint{X: int32(x), Y: int32(y), Amount: int32(res.Amount)})
				amount -= res.Amount
				if amount <= 0 {
					return
				}
			}
		}
	}
}

func (d *Digger) exchangeCash(cash []string) {
	needAsync := false
	for i, c := range cash {
		select {
		case d.cashCh <- c:
		default:
			cash = cash[i:]
			needAsync = true
			break
		}
	}
	if needAsync {
		fmt.Println("cash async")
		go func() {
			for _, c := range cash {
				d.cashCh <- c
			}
		}()
	}
}

func (d *Digger) startExchangeCashWorker() {
	go func() {
		for {
			c := <-d.cashCh
			p := make([]uint32, 0, 64)
			var res api.Payment = p
			t := time.Now()
			d.apiClient.Cash(c, &res)
			d.metrics.Add("cash", time.Since(t))
			d.bank.Store(res)
		}
	}()
}

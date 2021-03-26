package game

import (
	"goldrush/api"
	"goldrush/datastruct/areaqueue"
	"goldrush/datastruct/bank"
	"goldrush/datastruct/foreman"
	"goldrush/datastruct/goldpot"
	"goldrush/datastruct/pointqueue"
	"goldrush/metrics"
	"goldrush/utils"
	"time"
)

type Digger struct {
	apiClient  *api.Client
	bank       *bank.Bank
	pointQueue *pointqueue.PointQueue
	areaQueue  *areaqueue.AreaQueue
	metrics    *metrics.Svc
	foreman    *foreman.Foreman
	goldPot    *goldpot.GoldPot

	licenses       chan *api.License
	activeLicenses int32
	licWakeCh      chan struct{}

	areaCh chan *api.Area

	pointsCh chan pointqueue.DigPoint
}

func NewDigger(client *api.Client, metrics *metrics.Svc) *Digger {
	d := &Digger{
		apiClient:  client,
		licenses:   make(chan *api.License, 10),
		licWakeCh:  make(chan struct{}),
		bank:       bank.NewBank(),
		pointQueue: pointqueue.NewPointQueue(),
		metrics:    metrics,
		areaCh:     make(chan *api.Area, 50),
		areaQueue:  areaqueue.NewAreaQueue(),
		goldPot:    goldpot.New(),
	}
	f := foreman.New()
	f.AddWorker(firstScanWorker, d.fstScanWork)
	f.AddWorker(zoneScanWorker, d.zoneScanWork)
	f.AddWorker(licensesWorker, d.licensesWork)
	f.AddWorker(digWorker, d.digWork)
	f.AddWorker(exchangeCashWorker, d.exchangeCashWork)
	d.foreman = f
	return d
}

const (
	sizeX = 100
	sizeY = 100
)

const (
	firstScanWorker = iota
	zoneScanWorker
	licensesWorker
	digWorker
	exchangeCashWorker
)

func (d *Digger) Start() {
	go d.divideGameArea()
	d.foreman.Start(licensesWorker, 1)

	d.foreman.Start(firstScanWorker, 1)
	d.foreman.Start(zoneScanWorker, 4)
	d.foreman.Start(digWorker, 4)

	<-utils.WaitGameTime(time.Minute*9 + time.Second*30)
	d.foreman.ChangeState(zoneScanWorker, foreman.Slow, 4)
	<-time.After(utils.GetEndDelay() - time.Second*1)
	d.foreman.StopAll(licensesWorker)
}

func (d *Digger) divideGameArea() {
	for x := 0; x < 3500; x += sizeX {
		for y := 0; y < 3500; y += sizeY {
			d.areaCh <- &api.Area{PosX: x, PosY: y, SizeX: sizeX, SizeY: sizeY}
		}
	}
	close(d.areaCh)
}

func (d *Digger) fstScanWork(state *int) {
	a, ok := <-d.areaCh
	if ok {
		d.scan(a)
	} else {
		*state = foreman.Stopped
	}
}

func (d *Digger) zoneScanWork(state *int) {
	req := d.areaQueue.Peek()
	if *state == foreman.Slow && req.Area.Size() > 150 {
		return
	} else {
		d.bSearch(req)
	}
}

func (d *Digger) exchangeCashWork(state *int) {
	c := d.goldPot.Get()
	p := make([]uint32, 0, 64)
	var res api.Payment = p
	t := time.Now()
	d.apiClient.Cash(c, &res)
	d.metrics.Add("cash", time.Since(t))
	d.bank.Store(res)
}

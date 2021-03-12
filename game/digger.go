package game

import (
	"goldrush/api"
	"goldrush/bank"
	"goldrush/metrics"
	"goldrush/utils"
	"time"
)

type Digger struct {
	apiClient      *api.Client
	licenses       chan *api.License
	activeLicenses *int32
	cashCh         chan string
	bank           *bank.Bank
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
		cashCh:         make(chan string, 5000),
		licWakeCh:      make(chan struct{}),
		bank:           bank.NewBank(),
		activeLicenses: &l,
		pointQueue:     utils.NewPointQueue(),
		metrics:        metrics,
		areaCh:         make(chan *api.Area, 4),
		areaQueue:      utils.NewAreaQueue(),
	}
}

var throttle = false

func (d *Digger) Start() {
	go d.search()
	d.startWorkers(2, d.startScan)
	d.startWorkers(4, d.startScanB)

	<-time.After(time.Second)
	d.startWorkers(2, d.startGetLicensesWorker)
	d.startWorkers(3, d.startDigWorker)

	d.startWorkers(5, d.startExchangeCashWorker)

	go func() {
		<-time.After(time.Minute*6 + time.Second*40)
		throttle = true
		d.startWorkers(1, d.startDigWorker)
		d.startWorkers(3, d.startExchangeCashWorker)
		<-time.After(time.Minute * 2)
		d.startWorkers(1, d.startExchangeCashWorker)
	}()
}

const sizeX = 100
const sizeY = 500

func (d *Digger) search() {
	for x := 0; x < 3500; x += sizeX {
		for y := 0; y < 3500; y += sizeY {
			d.areaCh <- &api.Area{PosX: x, PosY: y, SizeX: sizeX, SizeY: sizeY}
		}
	}
	close(d.areaCh)
	d.startWorkers(6, d.startGetLicensesWorker)
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
		d.metrics.Add("CASH ASYNC", 1)
		go func() {
			for _, c := range cash {
				d.cashCh <- c
			}
		}()
	}
}

func (d *Digger) startWorkers(cnt int, f func()) {
	for i := 0; i < cnt; i++ {
		f()
	}
}

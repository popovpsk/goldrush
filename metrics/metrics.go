package metrics

import (
	"fmt"
	"runtime"
	"sync"
	"time"

	"github.com/montanaflynn/stats"
)

type Svc struct {
	data map[string][]int64
	m    sync.Mutex
}

func NewMetricsSvc() *Svc {
	return &Svc{
		data: make(map[string][]int64, 10),
	}
}

const metricsEnabled = true

func (s *Svc) Start() {
	go func() {
		if !metricsEnabled {
			return
		}

		for range time.Tick(time.Minute * 2) {
			cp := s.data
			s.m.Lock()
			s.data = make(map[string][]int64, 10)
			s.m.Unlock()
			fmt.Println("=======================")
			for name, v := range cp {
				if len(v) == 0 {
					return
				}
				var sum int64
				for _, t := range v {
					sum += t
				}
				d := stats.LoadRawData(v)
				avg := sum / int64(len(v))
				med, _ := d.Median()
				p90, _ := d.Percentile(90)
				p75, _ := d.Percentile(75)
				p95, _ := d.Percentile(99.5)
				str := fmt.Sprintf("%s: cnt:%v, avg:%v, med:%v, 75p:%v, 90p:%v, 99.5p:%v", name, len(v), avg, med, p75, p90, p95)
				fmt.Println(str)
				runtime.Gosched()
			}
		}
	}()
}

func (s *Svc) Add(name string, t time.Duration) {
	if !metricsEnabled {
		return
	}
	s.m.Lock()
	defer s.m.Unlock()
	_, ok := s.data[name]
	if ok {
		s.data[name] = append(s.data[name], t.Milliseconds())
	} else {
		s.data[name] = []int64{t.Milliseconds()}
	}
}

func (s *Svc) AddInt(name string, v int) {
	if !metricsEnabled {
		return
	}
	s.m.Lock()
	defer s.m.Unlock()
	_, ok := s.data[name]
	if ok {
		s.data[name] = append(s.data[name], int64(v))
	} else {
		s.data[name] = []int64{int64(v)}
	}
}

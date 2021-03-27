package main

import (
	"fmt"
	"github.com/montanaflynn/stats"
	"goldrush/api"
	"goldrush/types"
	"goldrush/utils"
	"os"
	"sync"
	"time"
)

var (
	rps     = 0
	workers = 16
	step    = 15
	cv      *sync.Cond

	output chan []time.Duration
)

func main() {
	scan()
	return

	utils.SetStartTime(time.Now())
	addr := os.Getenv("ADDRESS")
	port := 8000
	url := fmt.Sprintf("http://%s:%v", addr, port)
	cv = sync.NewCond(&sync.Mutex{})

	client := api.NewClient(url, api.NewGateWay())

	fmt.Println("Explore")
	fmt.Printf("rps, sum, avg, median, p75, p90, p99, max, err\n")

	for i := 0; i < workers; i++ {
		startWorker(client, i*200)
	}
	output = make(chan []time.Duration)
	for i := 800; i < 1400; i += 16 {
		rps = i
		<-time.After(time.Second)
		cv.Broadcast()

		var sum, max int64

		data := make([]int64, 0, 5000)
		for w := 0; w < workers; w++ {
			res := <-output
			for _, v := range res {
				t := v.Microseconds()
				sum += t
				if max < t {
					max = t
				}
				data = append(data, t)
			}
		}
		st := stats.LoadRawData(data)

		reqCnt := rps * step
		avg := int(sum) / reqCnt
		p75, _ := st.Percentile(75)
		p90, _ := st.Percentile(90)
		p99, _ := st.Percentile(99)
		median, _ := st.Median()
		fmt.Printf("%v, %v, %v, %v, %v, %v, %v, %v, %v\n", rps, sum, avg, median, p75, p90, p99, max, api.ErcntExp)
		api.ErcntExp = 0
		data = data[:0]
	}
}

func scan() {
	utils.SetStartTime(time.Now())
	addr := os.Getenv("ADDRESS")
	port := 8000
	url := fmt.Sprintf("http://%s:%v", addr, port)
	client := api.NewClient(url, api.NewGateWay())
	i := 0
	for x := 0; x < 3500; x++ {
		for y := 0; y < 3500; y++ {
			res := &types.ExploreResponse{}
			client.Explore(&types.Area{PosX: x, PosY: y, SizeX: 1, SizeY: 1}, res)
			if res.Amount > 0 {
				fmt.Println(x, y)
				i++
				if i == 101 {
					return
				}
			}
		}
	}
}

func startWorker(client *api.Client, x int) {
	go func() {
		area := &types.Area{
			SizeX: 1,
			SizeY: 1,
		}
		result := make([]time.Duration, 0, 1000)
		for y := 0; y < 3500; y++ {
			cv.L.Lock()
			cv.Wait()
			cv.L.Unlock()
			reqCnt := rps / workers * step
			delay := time.Duration(1000*workers/rps) * time.Millisecond
			startTime := time.Now()
			for i := 0; i < reqCnt; i++ {
				area.PosX = x
				area.PosY = y
				t := time.Now()
				client.Explore(area, &types.ExploreResponse{})
				d := time.Now().Sub(t)
				result = append(result, d)
				x++
				if x == 3500 {
					x = 0
				}

				zzz := delay - time.Since(t)
				if zzz > 0 {
					<-time.After(zzz)
				}
				if time.Since(startTime) >= time.Duration(step)*time.Second {
					break
				}
			}
			output <- result
			result = result[:0]
		}
	}()
}

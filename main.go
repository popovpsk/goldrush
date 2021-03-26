package main

import (
	"fmt"
	"goldrush/api"
	"goldrush/datastruct/pointqueue"
	"goldrush/game"
	"goldrush/metrics"
	"goldrush/utils"
	"os"
	"time"
)

func main() {

	utils.SetStartTime(time.Now())
	addr := os.Getenv("ADDRESS")
	port := 8000
	url := fmt.Sprintf("http://%s:%v", addr, port)

	m := metrics.NewMetricsSvc()

	cl := api.NewClient(url, api.NewGateWay())

	digger := game.NewDigger(cl, m)
	go digger.Start()
	go func() {
		<-time.After(utils.GetEndDelay() - time.Second)
		api.EndLog()
		pointqueue.EndLog()
	}()
	m.Start()
	<-time.After(time.Minute * 12)
}

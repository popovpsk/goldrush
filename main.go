package main

import (
	"fmt"
	"goldrush/api"
	"goldrush/datastruct/goldpot"
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

	go func() {
		<-time.After(utils.GetEndDelay())
		pointqueue.EndLog()
		api.EndLog()
		goldpot.EndLog()
		game.EndLog()
	}()

	m := metrics.NewMetricsSvc()
	m.Start()

	cl := api.NewClient(url, api.NewGateWay())

	digger := game.NewDigger(cl, m)
	digger.Start()
	<-time.After(time.Minute * 12)
}

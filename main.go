package main

import (
	"fmt"
	"goldrush/api"
	"goldrush/datastruct/bank"
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

	b := bank.NewBank()
	cl := api.NewClient(url, api.NewGateWay())

	licP := game.NewLicenseProvider(b, cl)
	digger := game.NewDigger(licP, cl, m, b)
	go digger.Start()
	go func() {
		<-time.After(utils.GetEndDelay() - time.Second)
		api.EndLog()
		pointqueue.EndLog()
	}()
	m.Start()
	<-time.After(time.Minute * 12)
}

package main

import (
	"fmt"
	"goldrush/api"
	"goldrush/game"
	"goldrush/metrics"
	"os"
	"time"
)

func main() {
	addr := os.Getenv("ADDRESS")
	port := 8000
	url := fmt.Sprintf("http://%s:%v", addr, port)

	m := metrics.NewMetricsSvc()
	m.Start()
	cl := api.NewClient(url)

	digger := game.NewDigger(cl, m)
	digger.Start()
	<-time.After(time.Minute * 12)
}

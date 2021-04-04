package main

import (
	"encoding/json"
	"fmt"
	"goldrush/api"
	"goldrush/datastruct/bank"
	"goldrush/game"
	"goldrush/metrics"
	"goldrush/utils"
	"os"
	"runtime"
	"time"
)

func main() {
	utils.SetStartTime(time.Now())
	addr := os.Getenv("ADDRESS")
	port := 8000
	url := fmt.Sprintf("http://%s:%v", addr, port)

	m := metrics.NewMetricsSvc()

	b := bank.NewBank()
	cl := api.NewClient(url, api.NewGateway(m), m)

	licP := game.NewLicenseProvider(b, cl)
	digger := game.NewDigger(licP, cl, m, b)
	go digger.Start()
	go func() {
		<-time.After(utils.GetEndDelay() - time.Second)
		api.EndLog()
		return
		mem := runtime.MemStats{}
		runtime.ReadMemStats(&mem)
		memJs, _ := json.Marshal(&mem)
		fmt.Println(string(memJs))
	}()
	m.Start()
	<-time.After(time.Minute * 12)
}

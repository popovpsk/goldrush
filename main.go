package main

import (
	"fmt"
	"os"
	"time"

	"goldrush/api"
	"goldrush/datastruct/bank"
	"goldrush/game"
	"goldrush/utils"
)

func main() {
	utils.SetStartTime(time.Now())
	addr := os.Getenv("ADDRESS")
	url := fmt.Sprintf("http://%s:8000", addr)
	b := bank.NewBank()
	cl := api.NewClient(url, api.NewGateway())
	licP := game.NewLicenseProvider(b, cl)
	digger := game.NewDigger(licP, cl, b)
	go digger.Start()
	<-time.After(time.Minute * 12)
}

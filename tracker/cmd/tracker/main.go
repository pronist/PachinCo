package main

import "github.com/pronist/upbit/tracker"

func main() {
	t := tracker.New()
	t.GetMarketByVolumePower(0.50)
}
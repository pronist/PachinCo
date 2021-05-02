package main

import (
	"flag"
	"github.com/pronist/upbit/observer"
	"log"
)

func main() {
	flag.Parse()
	log.SetFlags(0)

	o := observer.New()
	o.Run()
}
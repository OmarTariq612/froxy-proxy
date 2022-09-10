package main

import (
	"flag"
	"log"
	"strings"
)

func main() {
	addr := flag.String("addr", ":5555", "address to serve on")
	allowedPortsStr := flag.String("allow", "443", "CONNECT allowed ports (comma separated list)")
	flag.Parse()

	proxy := NewFroxyProxy(*addr, strings.Split((*allowedPortsStr), ","))
	err := proxy.ListenAndServe()
	if err != nil {
		log.Println(err)
	}
}

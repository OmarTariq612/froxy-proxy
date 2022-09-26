package main

import (
	"flag"
	"log"
	"strings"
)

func main() {
	addr := flag.String("addr", ":5555", "address to serve on")
	allowedPortsStr := flag.String("allow", "443", "CONNECT allowed ports (comma separated list)")
	cred := flag.String("cred", "", "username:password that will be used to authenticate http clients")
	flag.Parse()

	if *cred != "" && !strings.Contains(*cred, ":") {
		log.Println("cred must take the username:password form")
		return
	}

	proxy := NewFroxyProxy(*addr, strings.Split((*allowedPortsStr), ","), *cred)
	err := proxy.ListenAndServe()
	if err != nil {
		log.Println(err)
	}
}

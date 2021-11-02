package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"strings"
	"time"
)

type result struct {
	test   string
	result string
}

func main() {
	args := os.Args
	if len(args) != 2 {
		log.Fatalln("usage: domainchecker <url>")
	}
	url := args[1]
	if strings.Contains(url, "://") {
		url = strings.Split(url, "://")[1]
	}
	if strings.Contains(url, "/") {
		url = strings.Split(url, "/")[0]
	}
	c := make(chan result, 2)
	go whois(url, c)
	go dns(url, c)
	var r result
	for i := 0; i < 2; i++ {
		r = <-c
		fmt.Println(r.test + ": " + r.result)
	}
}

func whois(url string, res chan result) {
	var r result = result{test: "whois", result: "failed"}
	c := exec.Command("whois", url)
	b, e := c.Output()
	if e != nil {
		r.result = "failed"
	} else {
		s := string(b)
		if strings.Contains(s, "No match for domain") {
			r.result = "available"
		} else if strings.Contains(s, "Requests of this client are not permitted") {
			r.result = "failed"
		} else {
			r.result = "taken"
		}
	}
	res <- r
}

func dns(url string, res chan result) {
	var r result = result{test: "dns", result: "failed"}
	resolver := &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{
				Timeout: time.Millisecond * time.Duration(10000),
			}
			return d.DialContext(ctx, network, "8.8.8.8:53")
		},
	}
	ip, e := resolver.LookupHost(context.Background(), "www.google.com")
	if e != nil {
		r.result = "failed"
	} else {
		if len(ip) == 0 {
			r.result = "available"
		} else {
			r.result = "taken"
		}
	}
	res <- r
}

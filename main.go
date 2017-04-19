package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"time"
)

func main() {
	if len(os.Args) != 3 {
		log.Fatal("Usage:  valuedomain DOMAIN.NAME PASSWORD")
	}
	// ドメインが有効かチェック
	_, err := net.LookupIP(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	for {
		time.Sleep(time.Minute * 15)
		// Global IP Addr
		res, err := http.Get("https://dyn.value-domain.com/cgi-bin/dyn.fcg?ip")
		if err != nil {
			fmt.Println(err)
			continue
		}
		defer res.Body.Close()
		b := new(bytes.Buffer)
		io.Copy(b, res.Body)
		globalIP := b.String()
		fmt.Println("Router Global IP Addr:", globalIP)

		// LookUP Addr
		iplist, err := net.LookupIP(os.Args[1])
		if err != nil {
			fmt.Println(err)
			continue
		}
		// 比較して、一致するものを確認
		var diff = true
		for n, lookupIP := range iplist {
			fmt.Println("DNS Lookup IP Addr", n+1, ":", lookupIP)
			if globalIP == lookupIP.String() {
				diff = false
				break
			}
		}
		// 一致するものがなければDDNS更新処理
		if !diff {
			fmt.Println("No problem.")
			continue
		}
		// 変更をリクエストする
		res, err = http.Get("https://dyn.value-domain.com/cgi-bin/dyn.fcg?d=" + os.Args[1] + "&p=" + os.Args[2] + "&h=*&i=" + globalIP)
		defer res.Body.Close()
		io.Copy(os.Stdout, res.Body)
		fmt.Println()
		switch res.StatusCode {
		case 0:
			fmt.Println("Successful.")
		case 1:
			fmt.Println("Bad Request.")
		case 2:
			fmt.Println("Bad domain and password.")
		case 3:
			fmt.Println("Bad IP Addr.")
		case 4:
			fmt.Println("Bad password.")
		case 5:
			fmt.Println("Busy DBServer.")
		case 9:
			fmt.Println("Unexpected Error.")
		default:
			fmt.Println("Unexpected Status Code.")
		}
	}
}

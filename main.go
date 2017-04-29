package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"regexp"
	"strings"
	"syscall"
	"time"

	"golang.org/x/crypto/ssh/terminal"
)

func main() {
	if len(os.Args) != 2 {
		log.Fatal("Usage:  valuedomain DOMAIN.NAME")
	}
	// ドメインが有効かチェック
	_, err := net.LookupIP(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	// パスワード入力
	fmt.Print("Password:")
	password, err := terminal.ReadPassword(syscall.SYS_READ)
	fmt.Println()
	if err != nil {
		log.Fatal(err)
	}
	if len(password) == 0 {
		log.Fatal("パスワードを入力して!")
	}
	// パスワードが有効かチェック
	globalIP, err := getGlobalIP()
	if err != nil {
		log.Fatal(err)
	}
	res, err := setDDNS(os.Args[1], globalIP, string(password))
	if err != nil {
		log.Fatal(err)
	}
	if !strings.Contains(res, "status=0") {
		fmt.Println(res)
		log.Fatal("パスワードかドメインが正しくありません")
	}
	fmt.Println("パスワードのチェックに成功しました")
	fmt.Println("Start:", globalIP)
	for {
		time.Sleep(time.Minute * 15)

		globalIP, lookupIP, err := getIP()
		if err != nil {
			fmt.Println(err)
			continue
		}

		// 変更をリクエストする
		res, err := setDDNS(os.Args[1], globalIP, string(password))
		if err != nil {
			fmt.Println(err)
			continue
		}
		// 返されたステータスコードによって表示を切り替える
		switch parseStatus(res) {
		case 0:
			fmt.Println("Successful.")
			fmt.Println("Updated A recode in DNS", lookupIP, "->", globalIP)
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

var statusreg = regexp.MustCompile("status=\\d")

func parseStatus(s string) int {
	ss := statusreg.FindAllString(s, -1)
	if len(ss) != 1 {
		return 9
	}
	status := ss[0]
	return int(status[len(status)-1] - '0')
}

// グローバルIPとDNSに登録されたIPを取得するか、エラー(変更なし)
func getIP() (string, []net.IP, error) {
	// ルータのグローバルIPの取得
	globalIP, err := getGlobalIP()
	if err != nil {
		return "", nil, err
	}

	// DNSに登録されたIPアドレスを取得
	iplist, err := net.LookupIP(os.Args[1])
	if err != nil {
		return "", nil, err
	}
	// アドレス比較して、一致するものを確認
	for _, ip := range iplist {
		if globalIP == ip.String() {
			return "", nil, errors.New("アドレスの変更は必要ありません。")
		}
	}
	return globalIP, iplist, nil
}

func getGlobalIP() (string, error) {
	res, err := http.Get("https://dyn.value-domain.com/cgi-bin/dyn.fcg?ip")
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	b := new(bytes.Buffer)
	io.Copy(b, res.Body)
	return b.String(), nil
}

func setDDNS(domain, newIP, passwd string) (string, error) {
	request := "https://dyn.value-domain.com/cgi-bin/dyn.fcg?d=" + domain + "&p=" + passwd + "&h=*&i=" + newIP
	res, err := http.Get(request)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	b := new(bytes.Buffer)
	io.Copy(b, res.Body)
	return b.String(), nil
}

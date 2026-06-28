package main

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"
)

type IPInfo struct {
	IP       string `json:"ip"`
	City     string `json:"city"`
	Region   string `json:"region"`
	Country  string `json:"country"`
	Org      string `json:"org"`
	Location string `json:"loc"`
}

func main() {
	printBanner()

	fmt.Println("[*] Fetching Local IP Details...")
	getLocalIPs()
	fmt.Println("\n--------------------------------------------------")

	fmt.Println("[*] Fetching Public IP & Geolocation...")
	getPublicIPDetails()
	fmt.Println("\n--------------------------------------------------")

	fmt.Println("[*] Scanning Local Network for Connected Devices...")
	scanLocalNetwork()
}

func printBanner() {
	// Dynamically fetches the real current date
	currentTime := time.Now().Format("02/01/2006")

	banner := fmt.Sprintf(`
+-------------------------------------------------------------+
|                       SOVEREIGN RECON                       |
|             "Absolute Self-Sovereignty in Tech"             |
+-------------------------------------------------------------+
| A COMPLETE OSINT TOOL :)        | CODED BY: ARJUN RAJ       |
+-------------------------------------------------------------+
| [+] VERSION: 1.0.0              | CURRENT-DATE: %s  |
| Email: arjunraj.cyber@gmail.com | CLI-LANGUAGE: ENGLISH     |
+-------------------------------------------------------------+
`, currentTime)
	fmt.Println(banner)
}

func getLocalIPs() {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		fmt.Printf("[-] Error getting local IPs: %v\n", err)
		return
	}
	for _, address := range addrs {
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				fmt.Printf("[+] Private IP Address: %s\n", ipnet.IP.String())
			}
		}
	}
}

func getPublicIPDetails() {
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get("https://ipinfo.io/json")
	if err != nil {
		fmt.Printf("[-] Error fetching public IP info: %v\n", err)
		return
	}
	defer resp.Body.Close()

	var info IPInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		fmt.Printf("[-] Error parsing JSON data: %v\n", err)
		return
	}

	fmt.Printf("[+] Public IP:       %s\n", info.IP)
	fmt.Printf("[+] Location:        %s, %s, %s\n", info.City, info.Region, info.Country)
	fmt.Printf("[+] ISP/Organization: %s\n", info.Org)
	fmt.Printf("[+] Coordinates:     %s\n", info.Location)
}

func scanLocalNetwork() {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		fmt.Printf("[-] Error determining subnet: %v\n", err)
		return
	}

	var activeIP string
	for _, address := range addrs {
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				ipStr := ipnet.IP.String()
				if ipStr != "192.168.56.1" && ipStr != "169.254.218.0" {
					activeIP = ipStr
				}
			}
		}
	}

	if activeIP == "" {
		fmt.Println("[-] Could not find an active local network interface to scan.")
		return
	}

	ipObj := net.ParseIP(activeIP).To4()
	subnet := fmt.Sprintf("%d.%d.%d.", ipObj[0], ipObj[1], ipObj[2])

	fmt.Printf("[*] Dynamically targeting active subnet: %s0/24\n", subnet)
	fmt.Println("    (Scanning common ports, please wait...)\n")

	var wg sync.WaitGroup
	ports := []int{80, 443, 22, 445, 53}
	foundDevices := 0
	var mu sync.Mutex

	for i := 1; i <= 254; i++ {
		wg.Add(1)
		go func(hostID int) {
			defer wg.Done()
			ip := fmt.Sprintf("%s%d", subnet, hostID)

			for _, port := range ports {
				target := fmt.Sprintf("%s:%d", ip, port)
				conn, err := net.DialTimeout("tcp", target, 250*time.Millisecond)
				if err == nil {
					conn.Close()
					mu.Lock()
					fmt.Printf("[+] Active Host Found: %-15s (Port %d Open)\n", ip, port)
					foundDevices++
					mu.Unlock()
					break
				}
			}
		}(i)
	}

	wg.Wait()
	if foundDevices == 0 {
		fmt.Printf("[-] Scan complete. No other active devices responded to ports on %s0/24.\n", subnet)
	} else {
		fmt.Printf("\n[*] Scan complete. Found %d active network points.\n", foundDevices)
	}
}

package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// setupNetwork creates a veth pair, connects it, and configures NAT for internet access
func setupNetwork(pid int) {
	fmt.Println("🌐 Configuring network and NAT routing...")

	// 1. Create veth pair
	must(exec.Command("ip", "link", "add", "veth-host", "type", "veth", "peer", "name", "veth-child").Run())

	// 2. Move veth-child into the container
	must(exec.Command("ip", "link", "set", "veth-child", "netns", fmt.Sprintf("%d", pid)).Run())

	// 3. Configure the Host end
	must(exec.Command("ip", "link", "set", "veth-host", "up").Run())
	must(exec.Command("ip", "addr", "add", "10.0.0.1/24", "dev", "veth-host").Run())

	// 4. Configure the Container end
	must(exec.Command("nsenter", "-t", fmt.Sprintf("%d", pid), "-n", "ip", "link", "set", "veth-child", "name", "eth0").Run())
	must(exec.Command("nsenter", "-t", fmt.Sprintf("%d", pid), "-n", "ip", "addr", "add", "10.0.0.2/24", "dev", "eth0").Run())
	must(exec.Command("nsenter", "-t", fmt.Sprintf("%d", pid), "-n", "ip", "link", "set", "eth0", "up").Run())
	must(exec.Command("nsenter", "-t", fmt.Sprintf("%d", pid), "-n", "ip", "link", "set", "lo", "up").Run())
	must(exec.Command("nsenter", "-t", fmt.Sprintf("%d", pid), "-n", "ip", "route", "add", "default", "via", "10.0.0.1").Run())

	// 5. Enable IP forwarding on the host
	must(os.WriteFile("/proc/sys/net/ipv4/ip_forward", []byte("1\n"), 0644))

	// 6. Automatically detect the primary interface (e.g., eth0, wlan0)
	// 'ip route get 1' returns the route to 1.1.1.1, which includes the interface name
	out, err := exec.Command("ip", "route", "get", "1").Output()
	if err != nil {
		fmt.Println("Warning: Could not detect default interface for NAT.")
		return
	}
	
	// Parse the output: "1.1.1.1 dev eth0 src 192.168.1.5 ..."
	parts := strings.Split(string(out), " ")
	dev := ""
	for i, part := range parts {
		if part == "dev" && i+1 < len(parts) {
			dev = parts[i+1]
			break
		}
	}

	// 7. Configure NAT
	if dev != "" {
		fmt.Printf("✅ NATting via interface: %s\n", dev)
		exec.Command("iptables", "-t", "nat", "-A", "POSTROUTING", "-s", "10.0.0.0/24", "-o", dev, "-j", "MASQUERADE").Run()
	} else {
		fmt.Println("Warning: Could not apply NAT rule, check your host routes.")
	}
}
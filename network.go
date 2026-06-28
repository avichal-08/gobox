package main

import (
	"fmt"
	"os"
	"os/exec"
)

func setupNetwork(pid int) {
	fmt.Println("Configuring network and NAT routing...")

	must(exec.Command("ip", "link", "add", "veth-host", "type", "veth", "peer", "name", "veth-child").Run())

	must(exec.Command("ip", "link", "set", "veth-child", "netns", fmt.Sprintf("%d", pid)).Run())

	must(exec.Command("ip", "link", "set", "veth-host", "up").Run())
	must(exec.Command("ip", "addr", "add", "10.0.0.1/24", "dev", "veth-host").Run())

	must(exec.Command("nsenter", "-t", fmt.Sprintf("%d", pid), "-n", "ip", "link", "set", "veth-child", "name", "eth0").Run())
	must(exec.Command("nsenter", "-t", fmt.Sprintf("%d", pid), "-n", "ip", "addr", "add", "10.0.0.2/24", "dev", "eth0").Run())
	must(exec.Command("nsenter", "-t", fmt.Sprintf("%d", pid), "-n", "ip", "link", "set", "eth0", "up").Run())
	must(exec.Command("nsenter", "-t", fmt.Sprintf("%d", pid), "-n", "ip", "link", "set", "lo", "up").Run())

	must(exec.Command("nsenter", "-t", fmt.Sprintf("%d", pid), "-n", "ip", "route", "add", "default", "via", "10.0.0.1").Run())

	must(os.WriteFile("/proc/sys/net/ipv4/ip_forward", []byte("1\n"), 0644))

	exec.Command("iptables", "-t", "nat", "-A", "POSTROUTING", "-s", "10.0.0.0/24", "-j", "MASQUERADE").Run()
}
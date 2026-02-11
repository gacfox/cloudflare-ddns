package network

import (
	"net"
	"strings"
)

func GetNetworkAddresses(interfaceName string) (ipv4 string, ipv6 string, err error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return "", "", err
	}

	var ipv4s, ipv6s []string

	for _, iface := range interfaces {
		if iface.Name != interfaceName {
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			ipNet, ok := addr.(*net.IPNet)
			if !ok {
				continue
			}
			ip := ipNet.IP

			if ip.To4() != nil {
				ipv4s = append(ipv4s, ip.String())
				continue
			}

			ipv6Str := ip.String()
			if idx := strings.Index(ipv6Str, "%"); idx != -1 {
				ipv6Str = ipv6Str[:idx]
			}
			ipv6s = append(ipv6s, ipv6Str)
		}
		break
	}

	ipv4 = GetPublicIPv4(ipv4s)
	ipv6 = GetPublicIPv6(ipv6s)

	return ipv4, ipv6, nil
}

func GetPublicIPv4(addrs []string) string {
	for _, addr := range addrs {
		ip := net.ParseIP(addr)
		if ip == nil {
			continue
		}
		if isGlobalV4(ip) {
			return addr
		}
	}
	return ""
}

func GetPublicIPv6(addrs []string) string {
	for _, addr := range addrs {
		ip := net.ParseIP(addr)
		if ip == nil {
			continue
		}
		if isGlobalV6(ip) {
			return addr
		}
	}
	return ""
}

func isGlobalV4(ip net.IP) bool {
	if ip.IsLoopback() || ip.IsLinkLocalUnicast() || ip.IsPrivate() {
		return false
	}
	return ip.To4() != nil && !ip.IsUnspecified()
}

func isGlobalV6(ip net.IP) bool {
	if ip.IsLoopback() || ip.IsLinkLocalUnicast() || ip.IsInterfaceLocalMulticast() {
		return false
	}
	return ip.To4() == nil && !ip.IsUnspecified()
}

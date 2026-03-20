package netif

import (
	"log/slog"
	"net"
	"os"
)

type NetworkInterfaces []net.Interface

func (ni NetworkInterfaces) WithIPs(ips []net.IP) NetworkInterfaces {
	filtered := make([]net.Interface, 0)

	for _, iface := range ni {
		if addrs, err := iface.Addrs(); err == nil {
			for _, addr := range addrs {
				for _, ip := range ips {
					if _, n, err := net.ParseCIDR(addr.String()); err == nil {
						if n.Contains(ip) {
							filtered = append(filtered, iface)
						}
					}
				}
			}
		}
	}

	return filtered
}

func Get(list []string) NetworkInterfaces {
	if len(list) == 0 {
		ret, err := net.Interfaces()
		if err != nil {
			slog.Error("Unable to get interfaces", "error", err)
			os.Exit(1)
		}

		return ret
	}

	ret := make(NetworkInterfaces, 0)

	for _, name := range list {
		iface, err := net.InterfaceByName(name)
		if err != nil {
			slog.Warn("Failed to get interface, skipping", "error", err, "name", name)
		} else {
			ret = append(ret, *iface)
		}
	}

	return ret
}

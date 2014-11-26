// +build linux darwin

package network

import (
	"bufio"
	"bytes"
	"errors"
	"net"
	"os/exec"
	"strings"
)

type Network struct{}

const name = "network"

func (self *Network) Name() string {
	return name
}

func (self *Network) Collect() (result interface{}, err error) {
	result, err = getNetworkInfo()
	return
}

type TopLevel struct{}

func (t *TopLevel) Name() string {
	return "top_level"
}

func (t *TopLevel) Collect() (interface{}, error) {
	result, err := getTopLevel()
	return result, err
}

type Ipv6Address struct{}

func externalIpv6Address() (string, error) {
	ifaces, err := net.Interfaces()

	if err != nil {
		return "", err
	}

	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			// interface down or loopback interface
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil {
			return "", err
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip == nil || ip.IsLoopback() {
				continue
			}
			if ip.To4() != nil {
				// ipv4 address
				continue
			}
			return ip.String(), nil
		}
	}
	return "", errors.New("not connected to the network")
}

type IpAddress struct{}

func externalIpAddress() (string, error) {
	ifaces, err := net.Interfaces()

	if err != nil {
		return "", err
	}

	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			// interface down or loopback interface
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil {
			return "", err
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip == nil || ip.IsLoopback() {
				continue
			}
			ip = ip.To4()
			if ip == nil {
				// not an ipv4 address
				continue
			}
			return ip.String(), nil
		}
	}
	return "", errors.New("not connected to the network")
}

type MacAddress struct{}

func macAddress() (string, error) {
	ifaces, err := net.Interfaces()

	if err != nil {
		return "", err
	}

	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			// interface down or loopback interface
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil {
			return "", err
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip == nil || ip.IsLoopback() || ip.To4() == nil {
				continue
			}
			return iface.HardwareAddr.String(), nil
		}
	}
	return "", errors.New("not connected to the network")
}

func networkInterfaces() (map[string]interface{}, error) {
	ifaces := make(map[string]interface{})
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	for _, i := range interfaces {
		iInfo := make(map[string]interface{})
		iInfo["mtu"] = i.MTU
		iInfo["flags"] = i.Flags.String()
		iInfo["mac_addr"] = i.HardwareAddr.String()

		addrs, err := i.Addrs()
		if err != nil {
			return nil, err
		}
		iAddrs := make(map[string]interface{})
		for _, a := range addrs {
			//iAddrs[a.String()] = map[string]interface{}{ "network": a.Network() }
			ip, ipnet, err := net.ParseCIDR(a.String())
			if err != nil {
				return nil, err
			}
			var family string
			var mask string
			var broadcast string
			if ip.To4() == nil {
				family = "inet6"
			} else {
				family = "inet"
				maskip := net.IPv4(ipnet.Mask[0], ipnet.Mask[1], ipnet.Mask[2], ipnet.Mask[3])
				mask = maskip.String()
				if !ip.IsLoopback() {
					broadcast = net.IPv4(ipnet.IP[0]|(^ipnet.Mask[0]), ipnet.IP[1]|(^ipnet.Mask[1]), ipnet.IP[2]|(^ipnet.Mask[2]), ipnet.IP[3]|(^ipnet.Mask[3])).String()
				}
			}
			iAddrs[ip.String()] = map[string]interface{}{"family": family}
			if mask != "" {
				iAddrs[ip.String()].(map[string]interface{})["netmask"] = mask
			}
			if broadcast != "" {
				iAddrs[ip.String()].(map[string]interface{})["broadcast"] = broadcast
			}
		}

		iInfo["addresses"] = iAddrs

		ifaces[i.Name] = iInfo
	}
	return ifaces, nil
}

func settings() (map[string]interface{}, error) {
	s, err := exec.Command("sysctl", "-a", "net").Output()
	if err != nil {
		return nil, err
	}
	sets := make(map[string]interface{})
	sread := bufio.NewScanner(bytes.NewBuffer(s))
	for sread.Scan() {
		st := strings.Split(sread.Text(), ":")
		if len(st) < 2 {
			continue
		}
		sets[st[0]] = strings.TrimSpace(st[1])
	}
	return sets, nil
}
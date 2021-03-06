package manet

import (
	"fmt"
	"net"

	ma "mbfs/go-mbfs/gx/QmRKLtwMw131aK7ugC3G7ybpumMz78YrJe5dzneyindvG1/go-multiaddr"
)

var errIncorrectNetAddr = fmt.Errorf("incorrect network addr conversion")

// FromNetAddr converts a net.Addr type to a Multiaddr.
func FromNetAddr(a net.Addr) (ma.Multiaddr, error) {
	return defaultCodecs.FromNetAddr(a)
}

// FromNetAddr converts a net.Addr to Multiaddress.
func (cm *CodecMap) FromNetAddr(a net.Addr) (ma.Multiaddr, error) {
	if a == nil {
		return nil, fmt.Errorf("nil multiaddr")
	}
	p, err := cm.getAddrParser(a.Network())
	if err != nil {
		return nil, err
	}

	return p(a)
}

// ToNetAddr converts a Multiaddr to a net.Addr
// Must be ThinWaist. acceptable protocol stacks are:
// /ip{4,6}/{tcp, udp}
func ToNetAddr(maddr ma.Multiaddr) (net.Addr, error) {
	return defaultCodecs.ToNetAddr(maddr)
}

// ToNetAddr converts a Multiaddress to a standard net.Addr.
func (cm *CodecMap) ToNetAddr(maddr ma.Multiaddr) (net.Addr, error) {
	protos := maddr.Protocols()
	final := protos[len(protos)-1]

	p, err := cm.getMaddrParser(final.Name)
	if err != nil {
		return nil, err
	}

	return p(maddr)
}

func parseBasicNetMaddr(maddr ma.Multiaddr) (net.Addr, error) {
	network, host, err := DialArgs(maddr)
	if err != nil {
		return nil, err
	}

	switch network {
	case "tcp", "tcp4", "tcp6":
		return net.ResolveTCPAddr(network, host)
	case "udp", "udp4", "udp6":
		return net.ResolveUDPAddr(network, host)
	case "ip", "ip4", "ip6":
		return net.ResolveIPAddr(network, host)
	}

	return nil, fmt.Errorf("network not supported: %s", network)
}

func FromIPAndZone(ip net.IP, zone string) (ma.Multiaddr, error) {
	switch {
	case ip.To4() != nil:
		return ma.NewComponent("ip4", ip.String())
	case ip.To16() != nil:
		ip6, err := ma.NewComponent("ip6", ip.String())
		if err != nil {
			return nil, err
		}
		if zone == "" {
			return ip6, nil
		} else {
			zone, err := ma.NewComponent("ip6zone", zone)
			if err != nil {
				return nil, err
			}
			return zone.Encapsulate(ip6), nil
		}
	default:
		return nil, errIncorrectNetAddr
	}
}

// FromIP converts a net.IP type to a Multiaddr.
func FromIP(ip net.IP) (ma.Multiaddr, error) {
	return FromIPAndZone(ip, "")
}

// DialArgs is a convenience function returning arguments for use in net.Dial
func DialArgs(m ma.Multiaddr) (string, string, error) {
	var (
		zone, network, ip, port string
		err                     error
	)

	ma.ForEach(m, func(c ma.Component) bool {
		switch network {
		case "":
			switch c.Protocol().Code {
			case ma.P_IP6ZONE:
				if zone != "" {
					err = fmt.Errorf("%s has multiple zones", m)
					return false
				}
				zone = c.Value()
				return true
			case ma.P_IP6:
				network = "ip6"
				ip = c.Value()
				return true
			case ma.P_IP4:
				if zone != "" {
					err = fmt.Errorf("%s has ip4 with zone", m)
					return false
				}
				network = "ip4"
				ip = c.Value()
				return true
			}
		case "ip4":
			switch c.Protocol().Code {
			case ma.P_UDP:
				network = "udp4"
			case ma.P_TCP:
				network = "tcp4"
			default:
				return false
			}
			port = c.Value()
		case "ip6":
			switch c.Protocol().Code {
			case ma.P_UDP:
				network = "udp6"
			case ma.P_TCP:
				network = "tcp6"
			default:
				return false
			}
			port = c.Value()
		}
		// Done.
		return false
	})
	if err != nil {
		return "", "", err
	}
	switch network {
	case "ip6":
		if zone != "" {
			ip += "%" + zone
		}
		fallthrough
	case "ip4":
		return network, ip, nil
	case "tcp4", "udp4":
		return network, ip + ":" + port, nil
	case "tcp6", "udp6":
		if zone != "" {
			ip += "%" + zone
		}
		return network, "[" + ip + "]" + ":" + port, nil
	default:
		return "", "", fmt.Errorf("%s is not a 'thin waist' address", m)
	}
}

func parseTCPNetAddr(a net.Addr) (ma.Multiaddr, error) {
	ac, ok := a.(*net.TCPAddr)
	if !ok {
		return nil, errIncorrectNetAddr
	}

	// Get IP Addr
	ipm, err := FromIPAndZone(ac.IP, ac.Zone)
	if err != nil {
		return nil, errIncorrectNetAddr
	}

	// Get TCP Addr
	tcpm, err := ma.NewMultiaddr(fmt.Sprintf("/tcp/%d", ac.Port))
	if err != nil {
		return nil, errIncorrectNetAddr
	}

	// Encapsulate
	return ipm.Encapsulate(tcpm), nil
}

func parseUDPNetAddr(a net.Addr) (ma.Multiaddr, error) {
	ac, ok := a.(*net.UDPAddr)
	if !ok {
		return nil, errIncorrectNetAddr
	}

	// Get IP Addr
	ipm, err := FromIPAndZone(ac.IP, ac.Zone)
	if err != nil {
		return nil, errIncorrectNetAddr
	}

	// Get UDP Addr
	udpm, err := ma.NewMultiaddr(fmt.Sprintf("/udp/%d", ac.Port))
	if err != nil {
		return nil, errIncorrectNetAddr
	}

	// Encapsulate
	return ipm.Encapsulate(udpm), nil
}

func parseIPNetAddr(a net.Addr) (ma.Multiaddr, error) {
	ac, ok := a.(*net.IPAddr)
	if !ok {
		return nil, errIncorrectNetAddr
	}
	return FromIPAndZone(ac.IP, ac.Zone)
}

func parseIPPlusNetAddr(a net.Addr) (ma.Multiaddr, error) {
	ac, ok := a.(*net.IPNet)
	if !ok {
		return nil, errIncorrectNetAddr
	}
	return FromIP(ac.IP)
}

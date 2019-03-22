package discovery

import "net"

const (
	mdnsPort      = "5353"
	dnsPacketSize = 512 // in bytes
)

var (
	// Multicast groups used by mDNS
	mdnsGroupIPv4 = net.ParseIP("224.0.0.251")
	mdnsGroupIPv6 = net.ParseIP("ff02::fb")

	// mDNS wildcard addresses
	mdnsWildcardAddrIPv4 = net.ParseIP("224.0.0.0")
	mdnsWildcardAddrIPv6 = net.ParseIP("ff02::")
)

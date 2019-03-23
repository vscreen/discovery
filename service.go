package discovery

import (
	"context"
	"fmt"
	"log"
	"net"

	"github.com/miekg/dns"

	"golang.org/x/net/ipv4"
)

type Service struct {
	Name string
	Type string
	Port uint
	Data map[string]string
}

func Publish(ctx context.Context, s *Service) error {
	conn, err := net.ListenPacket("udp", net.JoinHostPort(mdnsWildcardAddrIPv4.String(), mdnsPort))
	if err != nil {
		return err
	}
	defer conn.Close()

	ifis, err := s.listInterfaces()
	if err != nil {
		return err
	}

	ipv4Conn, err := s.joinIPv4Group(conn, ifis)
	if err != nil {
		return err
	}
	defer s.leaveIPv4Group(ipv4Conn, ifis)
	defer ipv4Conn.Close()

	ipv4Ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	go s.ipv4Loop(ipv4Ctx, ipv4Conn)

	<-ctx.Done()
	return nil
}

// listInterfaces returns a set of interfaces that support multicast and are up.
func (s *Service) listInterfaces() ([]net.Interface, error) {
	filtered := make([]net.Interface, 0)

	interfaces, err := net.Interfaces()
	if err != nil {
		return filtered, err
	}

	for _, ifi := range interfaces {
		if (ifi.Flags & net.FlagUp) == 0 {
			continue
		}
		if (ifi.Flags & net.FlagMulticast) > 0 {
			filtered = append(filtered, ifi)
		}
	}

	return filtered, nil
}

func (s *Service) joinIPv4Group(conn net.PacketConn, ifis []net.Interface) (*ipv4.PacketConn, error) {
	var err error
	p := ipv4.NewPacketConn(conn)
	group := net.UDPAddr{IP: mdnsGroupIPv4}

	for _, ifi := range ifis {
		if err = p.JoinGroup(&ifi, &group); err != nil {
			return nil, err
		}
	}

	if err = p.SetControlMessage(ipv4.FlagSrc|ipv4.FlagDst, true); err != nil {
		return nil, err
	}
	return p, nil
}

func (s *Service) leaveIPv4Group(p *ipv4.PacketConn, ifis []net.Interface) error {
	group := net.UDPAddr{IP: mdnsGroupIPv4}
	for _, ifi := range ifis {
		if err := p.LeaveGroup(&ifi, &group); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) ipv4Loop(ctx context.Context, conn *ipv4.PacketConn) {
	var msg dns.Msg
	buff := make([]byte, dnsPacketSize)
	for {
		select {
		case <-ctx.Done():
		default: // continue looping
		}

		_, cm, _, err := conn.ReadFrom(buff)
		if err != nil {
			log.Println("[warning]", err)
			continue
		}

		if !cm.Dst.IsMulticast() || !cm.Dst.Equal(mdnsGroupIPv4) {
			continue
		}

		if err = msg.Unpack(buff); err != nil {
			log.Println("[warning]", err)
			continue
		}

		if len(msg.Question) == 0 {
			continue
		}

		switch msg.Question[0].Qtype {
		case dns.TypePTR:
			fmt.Println("PTR")
			fmt.Println(msg)
		case dns.TypeSRV:
			fmt.Println("SRV")
			fmt.Println(msg)
		}
	}
}

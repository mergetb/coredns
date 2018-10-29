package main

import (
	"net"

	dhcp "github.com/krolaw/dhcp4"
	log "github.com/sirupsen/logrus"

	"gitlab.com/mergetb/nex"
)

type Handler struct{}

func main() {

	log.SetLevel(log.DebugLevel)
	log.Infof("nex-dhcpd: %s", nex.Version)

	handler := &Handler{}
	log.Fatal(dhcp.ListenAndServeIf("eth1", handler))

}

func (h *Handler) ServeDHCP(
	pkt dhcp.Packet,
	msgType dhcp.MessageType,
	options dhcp.Options,
) dhcp.Packet {

	switch msgType {

	case dhcp.Discover:
		log.Debugf("discover: %s", pkt.CHAddr())

		response := func(server, addr net.IP, options []dhcp.Option) dhcp.Packet {
			log.Debugf("discover: OK for %s %s", addr, pkt.CHAddr())
			return dhcp.ReplyPacket(pkt, dhcp.Offer, server, addr, nex.LEASE_DURATION,
				options)
		}

		// Collect network information
		network, err := nex.GetNet(pkt.CHAddr())
		if err != nil {
			log.Errorf("discover: %v", err)
			return nil
		}
		if network == "" {
			log.Warnf("discover: %s has no net", pkt.CHAddr())
			return nil
		}

		server, err := nex.GetDhcp4ServerIp(network)
		if err != nil {
			log.Errorf("discover: %v", err)
			return nil
		}

		options, err := nex.GetIp4Options(network)
		if err != nil {
			log.Errorf("discover: %v", err)
			return nil
		}

		// If there is already an address use that
		addr, err := nex.GetIp4(pkt.CHAddr())
		if err != nil {
			log.Errorf("discover: %v", err)
			return nil
		}
		if addr != nil {
			return response(server, addr, options)
		}

		// If no address was found allocate a new one
		addr, err = nex.NewLease4(pkt.CHAddr(), network)
		if err != nil {
			log.Errorf("discover: %v", err)
			return nil
		}
		if addr != nil {
			return response(server, addr, options)
		}

	case dhcp.Request:
		log.Debugf("request: %s", pkt.CHAddr())

		rqAddr := net.IP(pkt.CIAddr())

		network, err := nex.GetNet(pkt.CHAddr())
		if err != nil {
			log.Errorf("request: %v", err)
			return nil
		}
		if network == "" {
			log.Warnf("request: %s has no net", pkt.CHAddr())
			return nil
		}

		server, err := nex.GetDhcp4ServerIp(network)
		if err != nil {
			log.Errorf("request: %v", err)
			return nil
		}

		_, subnet_mask, err := nex.GetSubnet4Mask(network)
		if err != nil {
			log.Errorf("request: %v", err)
			return nil
		}

		gws, err := nex.GetGateways(network)
		if err != nil {
			log.Errorf("request: %v", err)
			return nil
		}

		ns, err := nex.GetNameservers(network)
		if err != nil {
			log.Errorf("request: %v", err)
			return nil
		}

		addr, err := nex.GetIp4(pkt.CHAddr())
		if err != nil {
			log.Errorf("request: %v", err)
			return nil
		}
		if addr == nil {
			log.Warnf("request: no addr for %s", pkt.CHAddr())
			return dhcp.ReplyPacket(pkt, dhcp.NAK, server, nil, 0, nil)
		}

		var opts []dhcp.Option
		opts = append(opts,
			dhcp.Option{Code: dhcp.OptionSubnetMask, Value: subnet_mask.Mask})

		for _, x := range gws {
			opts = append(opts,
				dhcp.Option{Code: dhcp.OptionRouter, Value: x.To4()})
		}

		for _, x := range ns {
			opts = append(opts,
				dhcp.Option{Code: dhcp.OptionDomainNameServer, Value: x.To4()})
		}

		if addr.Equal(rqAddr) || rqAddr.Equal(net.IPv4(0, 0, 0, 0)) {

			log.Debugf("reqeust: OK for %s %s ~ %s", addr, pkt.CHAddr(), server)
			nex.RenewLease(pkt.CHAddr())
			return dhcp.ReplyPacket(pkt, dhcp.ACK, server, addr, nex.LEASE_DURATION,
				opts)

		} else {

			log.Warnf("request: unsolicited IP for  %s %s != %s", pkt.CHAddr(), addr,
				rqAddr)
			return dhcp.ReplyPacket(pkt, dhcp.NAK, server, nil, 0, nil)

		}

	case dhcp.Release:
		log.Debug("release: %s", pkt.CHAddr())

	case dhcp.Decline:
		log.Debug("decline: %s", pkt.CHAddr())

	}

	return nil

}

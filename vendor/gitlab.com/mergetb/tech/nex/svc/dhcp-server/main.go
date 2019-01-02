package main

import (
	"flag"
	"net"
	"strings"

	dhcp "github.com/krolaw/dhcp4"
	conn "github.com/krolaw/dhcp4/conn"
	log "github.com/mergetb/logrus"

	"gitlab.com/mergetb/tech/nex/pkg"
)

type Handler struct{}

var configpath = flag.String("config", "", "nex config file")

func main() {

	log.SetLevel(log.DebugLevel)
	log.Infof("nex-dhcpd: %s", nex.Version)

	flag.Parse()
	if *configpath != "" {
		nex.ConfigPath = *configpath
	}

	err := nex.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}

	handler := &Handler{}
	log.Infof("listening on %s", nex.Current.Dhcpd.Interface)
	cnx, err := conn.NewUDP4BoundListener(nex.Current.Dhcpd.Interface, ":67")
	if err != nil {
		log.Fatal(err)
	}
	log.Fatal(dhcp.Serve(cnx, handler))
}

func (h *Handler) ServeDHCP(
	pkt dhcp.Packet,
	msgType dhcp.MessageType,
	options dhcp.Options,
) dhcp.Packet {

	fields := log.Fields{}

	switch msgType {

	case dhcp.Discover:
		fields["mac"] = pkt.CHAddr()
		log.WithFields(fields).Debug("discover: start")

		response := func(server, addr net.IP, options []dhcp.Option) dhcp.Packet {

			fields["addr"] = addr
			log.WithFields(fields).Debug("discover: OK")

			return dhcp.ReplyPacket(
				pkt, dhcp.Offer, server, addr, nex.DHCP_LEASE_DURATION, options)
		}

		// Collect network information
		network, err := nex.FindMacNetwork(pkt.CHAddr())
		if err != nil {
			log.WithError(err).Error("discover: error")
			return nil
		}
		fields["network"] = network.Name
		log.WithFields(fields).Debug("found network")
		if network == nil {
			log.WithFields(fields).Warn("discover: has no net")
			return nil
		}

		// If there is already an address use that
		addr, err := nex.FindMacIpv4(pkt.CHAddr())
		if err != nil {
			log.WithError(err).Errorf("discover: error")
			return nil
		}
		if addr != nil {
			fields["static"] = addr.String()
			return response(net.ParseIP(network.Dhcp4Server), addr, ToOpt(network.Options))
		}

		// If no address was found allocate a new one
		addr, err = nex.NewLease4(pkt.CHAddr(), network.Name)
		if err != nil {
			log.WithError(err).Errorf("discover: error")
			return nil
		}
		if addr != nil {
			return response(net.ParseIP(network.Dhcp4Server), addr, ToOpt(network.Options))
		} else {
			log.WithFields(fields).Error("address pool depleted")
			//Address is nil, so no discover response will go out
		}

	case dhcp.Request:
		fields["mac"] = pkt.CHAddr()
		log.WithFields(fields).Debug("request: start")

		rqAddr := net.IP(pkt.CIAddr())

		network, err := nex.FindMacNetwork(pkt.CHAddr())
		if err != nil {
			log.WithError(err).Error("request: error")
			return nil
		}
		if network == nil {
			log.WithFields(fields).Warn("request: has no net")
			return nil
		}
		server := net.ParseIP(network.Dhcp4Server)

		addr, err := nex.FindMacIpv4(pkt.CHAddr())
		if err != nil {
			log.WithError(err).Error("request: error")
			return nil
		}
		if addr == nil {

			log.WithFields(fields).Warn("request: no address found")
			return dhcp.ReplyPacket(pkt, dhcp.NAK, server, nil, 0, nil)

		}

		var opts []dhcp.Option

		_, subnetCIDR, err := net.ParseCIDR(network.Subnet4)
		if err != nil {
			log.Errorf("bad subnet: %v", err)
			return nil
		}

		opts = append(opts,
			dhcp.Option{
				Code: dhcp.OptionSubnetMask,
				//Value: subnet_mask.Mask,
				Value: subnetCIDR.Mask,
			})

		for _, x := range network.Gateways {
			ip := net.ParseIP(x)
			opts = append(opts,
				dhcp.Option{Code: dhcp.OptionRouter, Value: ip.To4()})
		}

		for _, x := range network.Nameservers {
			ip := net.ParseIP(x)
			opts = append(opts,
				dhcp.Option{Code: dhcp.OptionDomainNameServer, Value: ip.To4()})
		}

		if network.Domain != "" {
			cn := compressedDnsName(network.Domain)

			opts = append(opts,
				dhcp.Option{Code: dhcp.OptionDomainSearch, Value: cn})
		}

		if addr.Equal(rqAddr) || rqAddr.Equal(net.IPv4(0, 0, 0, 0)) {

			log.WithFields(fields).Debug("request: OK")

			nex.RenewLease(pkt.CHAddr())
			return dhcp.ReplyPacket(pkt, dhcp.ACK, server, addr, nex.DHCP_LEASE_DURATION,
				opts)

		} else {

			fields["rqAddr"] = rqAddr
			fields["addr"] = addr
			log.WithFields(fields).Warn("request: unsolicited IP")
			return dhcp.ReplyPacket(pkt, dhcp.NAK, server, nil, 0, nil)

		}

	case dhcp.Release:
		log.Debugf("release: %s", pkt.CHAddr())

	case dhcp.Decline:
		log.Debugf("decline: %s", pkt.CHAddr())

	}

	return nil

}

func compressedDnsName(name string) []byte {
	parts := strings.Split(name, ".")

	var payload []byte

	for _, p := range parts {
		payload = append(payload, byte(len(p)))
		payload = append(payload, []byte(p)...)
	}

	payload = append(payload, byte(0))

	return payload
}

func ToOpt(opts []*nex.Option) []dhcp.Option {
	var result []dhcp.Option
	for _, x := range opts {
		result = append(result, dhcp.Option{
			Code:  dhcp.OptionCode(x.Number),
			Value: []byte(x.Value),
		})
	}
	return result
}

package nex

import (
	"net"
	"time"

	"github.com/golang/protobuf/ptypes/timestamp"
)

type Pool struct {
	CountSet
	Net string
}

func NewLease4(mac net.HardwareAddr, network string) (net.IP, error) {

	net := NewNetworkObj(&Network{Name: network})
	err := Read(net)
	if err != nil {
		return nil, err
	}

	pool := NewPoolObj(&Pool{Net: network})
	err = Read(pool)
	if err != nil {
		return nil, err
	}

	index, cs, err := pool.CountSet.Add()
	if err != nil {
		return nil, nil
	}
	pool.CountSet = cs
	ip := net.Range4.Select(index)

	m := NewMacIndex(&Member{Mac: mac.String()})
	err = Read(m)
	if err != nil {
		return nil, err
	}

	expires := time.Now().Add(DHCP_LEASE_DURATION)
	m.Ip4 = &Lease{
		Address: ip.String(),
		Expires: &timestamp.Timestamp{
			Seconds: expires.Unix(),
			Nanos:   0,
		},
	}

	err = WriteObjects([]Object{pool, m, NewIp4Index(m.Member)})
	if err != nil {
		return nil, err
	}

	return ip, err

}

func RenewLease(mac net.HardwareAddr) error {

	m := &Member{Mac: mac.String()}
	obj := NewMacIndex(m)
	err := Read(obj)
	if err != nil {
		return nil
	}

	// Nothing to do
	if m.Ip4 == nil {
		return nil
	}

	expires := time.Now().Add(DHCP_LEASE_DURATION)
	m.Ip4.Expires.Seconds = expires.Unix()
	m.Ip4.Expires.Nanos = 0

	return Write(obj)

}

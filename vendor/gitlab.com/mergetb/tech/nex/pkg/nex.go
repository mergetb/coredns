package nex

import (
	"encoding/binary"
	"flag"
	"fmt"
	"io/ioutil"
	"net"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

var Version string = "v0.4.3"
var ConfigPath = flag.String("config", "/etc/nex/nex.yml", "config file location")
var Current *Config

//TODO configurable
var DHCP_LEASE_DURATION time.Duration = 4 * time.Hour

type Addrs struct {
	Ip4 net.IP
	Ip6 net.IP
}

/* Primary API functions ++++++++++++++++++++++++++++++++++++++++++++++++++++++
+
+ All of the primary API functions exist to modify the nex database. However,
+ the do not modify the database directly. They return transaction operations
+ that can be composed into transactions by higher level calling functions.
+ This is necessary to support database API operations with non-trivial data
+ dependencies.
+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++*/

func FindMacNetwork(mac net.HardwareAddr) (*Network, error) {

	// First look up static member
	member := NewMacIndex(&Member{Mac: strings.ToLower(mac.String())})
	err := Read(member)
	if err != nil && !IsNotFound(err) {
		return nil, err
	}
	if err == nil {
		netobj := NewNetworkObj(&Network{Name: member.Net})
		err = Read(netobj)
		if err != nil {
			return nil, err
		}
		return netobj.Network, nil
	}

	// If no static member found, search for dynamic members
	nets, err := GetNetworks()
	for _, x := range nets {

		log.Debugf("%s in %s?", mac, x)
		hwbegin, err := net.ParseMAC(x.MacRange.Begin)
		if err != nil {
			log.Warnf("network '%s' has invalid mac_range begin", x)
			continue
		}
		hwend, err := net.ParseMAC(x.MacRange.End)
		if err != nil {
			log.Warnf("network '%s' has invalid mac_range end", x)
			continue
		}

		begin := binary.BigEndian.Uint64(append([]byte{0, 0}, []byte(hwbegin)...))
		end := binary.BigEndian.Uint64(append([]byte{0, 0}, []byte(hwend)...))
		here := binary.BigEndian.Uint64(append([]byte{0, 0}, []byte(mac)...))

		log.Debugf("lower=%d (%s)", begin, hwbegin)
		log.Debugf("upper=%d (%s)", end, hwend)
		log.Debugf("here =%d (%s)", here, mac)

		if begin < here && here < end {
			return x, nil
		}
	}

	return nil, nil

}

func FindMacIpv4(mac net.HardwareAddr) (net.IP, error) {

	member := NewMacIndex(&Member{Mac: mac.String()})
	err := Read(member)
	if err != nil {
		return nil, err
	}
	if member.Ip4 == nil {
		return nil, nil
	}

	return net.ParseIP(member.Ip4.Address), nil

}

func ResolveName(name string) (*Addrs, error) {

	log.WithFields(log.Fields{"name": name}).Info("resolving name")

	ni := NewNameIndex(&Member{Name: name})
	err := Read(ni)
	if err != nil {
		return nil, err
	}

	log.Printf("%v", ni)

	mi := NewMacIndex(ni.Member)
	err = Read(mi)
	if err != nil {
		return nil, err
	}
	if mi.Ip4 == nil {
		return nil, nil
	}

	log.Printf("%v", mi)

	log.WithFields(log.Fields{"ip": mi.Ip4.Address}).Info("resolved")
	return &Addrs{
		Ip4: net.ParseIP(mi.Ip4.Address),
	}, nil

}

/* types ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~*/

type Opt4 struct {
	Number int
	Value  string
}

type Opt6 struct {
	Number int
	Value  string
}

type EtcdConfig struct {
	Host   string `yaml:"host"`
	Port   int    `yaml:"port"`
	Cert   string `yaml:"cert"`
	Key    string `yaml:"key"`
	CAcert string `yaml:"cacert"`
}

type Config struct {
	Dhcpd DhcpdConfig `yaml:"dhcpd"`
	Etcd  EtcdConfig  `yaml:"etcd"`
	Nexd  NexdConfig  `yaml:"nexd"`
}

type DhcpdConfig struct {
	Interface string `yaml:"interface"`
}

type NexdConfig struct {
	Listen string `yaml:"listen"`
}

func (c EtcdConfig) HasTls() bool {
	return c.CAcert != "" && c.Cert != "" && c.Key != ""
}

/* helper functions ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~*/

func init() {
	flag.Parse()
}

func Errorf(message string, err error) error {
	err = fmt.Errorf("%s : %s", message, err)
	log.Error(err)
	return err
}

func LoadConfig() error {

	data, err := ioutil.ReadFile(*ConfigPath)
	if err != nil {
		log.Error(err)
		return fmt.Errorf("could not read configuration file")
	}

	err = yaml.Unmarshal(data, &Current)
	if err != nil {
		log.Error(err)
		return fmt.Errorf("could not parse configuration file")
	}

	return nil

}

// helpers ====================================================================

func poolIndex(key string) (int, error) {
	parts := strings.Split(key, "/")
	index := parts[len(parts)-1]
	return strconv.Atoi(index)
}

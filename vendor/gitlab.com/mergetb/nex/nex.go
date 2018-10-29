package nex

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/mvcc/mvccpb"
	mapset "github.com/deckarep/golang-set"
	dhcp "github.com/krolaw/dhcp4"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"

	proto "gitlab.com/mergetb/nex/proto"
)

var Version string = "v0.1.1"
var ConfigPath string = "/etc/merge/nex.yml"
var Current *Config

var LEASE_DURATION time.Duration = 1 * time.Hour

/* nex database structure  ====================================================
XXX /mac/:mac: -> { ip4, ip6, name, net }
XXX /name/:name: -> [ mac ]

/mac/:mac:/net 	-> net
/mac/:mac:/ip4 	-> ip4
/mac/:mac:/ip6 	-> ip6
/mac/:mac:/name -> name

# A name can be assigned to multiple macs
/name/:name:/:mac:
/name/:name:/:mac:

# An IP can be assigned to multiple macs
/ip4/:ip4:/name/:mac:
/ip6/:ip6:/name/:mac:



Each mac is associated to one ip address set, name and network. Names can map
onto many mac addresses, and subsequently ip address sets. Addresses map onto
a single mac.

addr: { ip4, ip6 }
name: string
mac: string
net: string

/:net: -> { subnet, ip4_range, ip6_range, pool4, pool6, gateways, domain, opts }
/net/:name:
	/subnet
	/dhcp4server
	/dhcp6server
  /ip4_range
  /ip6_range
  /mac_range
  /pool6
  /pool4
  /gateways -> [ip]
  /nameservers -> [ip]
  /domain
	/opts4/3   -> value
	      /47  -> value
	      /92  -> value
	/opts6/3   -> value
	      /88  -> value
	      /135 -> value
	/pool4/1 -> <mac>
        /2 -> <mac>
        /5 -> <mac>
        When the etcd lease expires the key disappears and may be reallocated
         ....
	/members/:mac: -> :mac:

* ============================================================================ */

type Member struct {
	Mac  string `json:"mac" yaml:"mac" omitempty`
	Name string `json:"name" yaml:"name" omitempty`
	Ip4  string `json:"ip4" yaml:"ip4" omitempty`
	Ip6  string `json:"ip6" yaml:"ip6" omitempty`
	Net  string `json:"net" yaml:"net" omitempty`
}

type Addrs struct {
	Ip4 net.IP
	Ip6 net.IP
}

type AddressRange struct {
	Begin string `json:"begin" yaml:"begin"`
	End   string `json:"end" yaml:"end"`
}

type Option struct {
	Number int    `json:"number" yaml:"number"`
	Value  string `json:"value" yaml:"value"`
}

type Network struct {
	Name        string        `json:"name" yaml:"name"`
	Subnet4     string        `json:"subnet4" yaml:"subnet4"`
	Subnet6     string        `json:"subnet6" yaml:"subnet6"`
	Dhcp4Server string        `json:"dhcp4server" yaml:"dhcp4server"`
	Dhcp6Server string        `json:"dhcp6server" yaml:"dhcp6server"`
	Ip4Range    *AddressRange `json:"ip4_range" yaml:"ip4_range"`
	Ip6Range    *AddressRange `json:"ip6_range" yaml:"ip6_range"`
	Gateways    []string      `json:"gateways" yaml:"gateways"`
	Nameservers []string      `json:"nameservers" yaml:"nameservers"`
	Options     []Option      `json:"options" yaml:"options"`
	Domain      string        `json:"domain" yaml:"domain"`
	MacRange    *AddressRange `json:"mac_range" yaml:"mac_range"`
	Members     []Member      `json:"members" yaml:"members"`
}

/* Primary API functions ++++++++++++++++++++++++++++++++++++++++++++++++++++++
+
+ All of the primary API functions exist to modify the nex database. However,
+ the do not modify the database directly. They return transaction operations
+ that can be composed into transactions by higher level calling functions.
+ This is necessary to support database API operations with non-trivial data
+ dependencies.
+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++*/

func GetIp4(mac net.HardwareAddr) (net.IP, error) {

	value, err := fetchOneValue("/mac/%s/ip4", mac)
	return net.ParseIP(value), err

}

func GetNet(mac net.HardwareAddr) (string, error) {

	network, err := fetchOneValue("/mac/%s/net", mac)
	if err != nil {
		return "", err
	}
	if network != "" {
		return network, nil
	}

	nets, _, err := getNets(nil)
	for _, x := range nets {
		log.Debugf("%s in %s?", mac, x)
		mac_range, err := fetchOneValue("/net/%s/mac_range", x)
		if err != nil || mac_range == "" {
			log.Debugf("%s has no mac_range (%v)", x, err)
			continue
		}
		var mr AddressRange
		err = json.Unmarshal([]byte(mac_range), &mr)
		if err != nil {
			log.Warnf("network '%s' has invalid mac_range", x)
			continue
		}
		hwbegin, err := net.ParseMAC(mr.Begin)
		if err != nil {
			log.Warnf("network '%s' has invalid mac_range begin", x)
			continue
		}
		hwend, err := net.ParseMAC(mr.End)
		if err != nil {
			log.Warnf("network '%s' has invalid mac_range end", x)
			continue
		}

		begin := binary.BigEndian.Uint64(append([]byte{0, 0}, []byte(hwbegin)...))
		end := binary.BigEndian.Uint64(append([]byte{0, 0}, []byte(hwend)...))
		here := binary.BigEndian.Uint64(append([]byte{0, 0}, []byte(mac)...))

		log.Debug("lower=%d (%s)", begin, hwbegin)
		log.Debug("upper=%d (%s)", end, hwend)
		log.Debug("here =%d (%s)", here, mac)

		if begin < here && here < end {
			return x, nil
		}
	}

	return "", nil

}

func GetDhcp4ServerIp(network string) (net.IP, error) {

	value, err := fetchOneValue("/net/%s/dhcp4server", network)
	return net.ParseIP(value).To4(), err

}

func GetGateways(network string) ([]net.IP, error) {
	value, err := fetchOneValue("/net/%s/gateways", network)
	if err != nil {
		return nil, err
	}

	var gws []string
	err = json.Unmarshal([]byte(value), &gws)
	if err != nil {
		return nil, err
	}

	var result []net.IP
	for _, x := range gws {
		result = append(result, net.ParseIP(x))
	}

	return result, nil
}

func GetNameservers(network string) ([]net.IP, error) {
	value, err := fetchOneValue("/net/%s/nameservers", network)
	if err != nil {
		return nil, err
	}

	var ns []string
	err = json.Unmarshal([]byte(value), &ns)
	if err != nil {
		return nil, err
	}

	var result []net.IP
	for _, x := range ns {
		result = append(result, net.ParseIP(x))
	}

	return result, nil
}

func GetSubnet4Mask(network string) (net.IP, *net.IPNet, error) {

	value, err := fetchOneValue("/net/%s/subnet4", network)
	if err != nil {
		return nil, nil, err
	}
	return net.ParseCIDR(value)

}

func GetIp4Options(network string) ([]dhcp.Option, error) {

	kvs, err := fetchKvs("/net/%s/opts4", network)
	if err != nil {
		return nil, err
	}

	var result []dhcp.Option
	for _, x := range kvs {
		i, err := strconv.Atoi(string(x.Key))
		if err != nil {
			log.Warning("%s: invalid option index '%s'", network, x.Key)
			continue
		}
		result = append(result, dhcp.Option{Code: dhcp.OptionCode(i), Value: x.Value})
	}

	return result, nil
}

//TODO
func NewLease4(mac net.HardwareAddr, network string) (net.IP, error) {

	kvs, err := fetchKvs("/net/%s/pool4/", network)
	if err != nil {
		return nil, err
	}

	buf, err := fetchOneValue("/net/%s/ip4_range", network)
	if err != nil {
		return nil, err
	}
	var rng AddressRange
	err = json.Unmarshal([]byte(buf), &rng)
	if err != nil {
		log.Errorf("failed to parse ip4_range @%s - '%s'", network, buf)
		return nil, err
	}
	size := rng.Size()

	space := unusedKeyspace(kvs, size)

	choice := rand.Intn(space.Cardinality())
	offset := space.ToSlice()[choice].(int)
	result := rng.Select(offset)

	err = SaveLease4(mac, network, offset, result)

	return result, err

}

func SaveLease4(mac net.HardwareAddr, network string, offset int,
	ip net.IP) error {

	c, err := EtcdClient()
	if err != nil {
		return err
	}
	defer c.Close()

	for i := 0; i < 100; i++ {
		lease, err := c.Grant(context.TODO(), int64(LEASE_DURATION.Seconds()))
		if err != nil {
			return err
		}

		poolkey := fmt.Sprintf("/net/%s/pool4/%d", network, offset)
		poolvalue := mac.String()

		lookupkey := fmt.Sprintf("/mac/%s/ip4", mac)
		lookupvalue := ip.String()

		kvc := clientv3.NewKV(c)
		_, err = kvc.Txn(context.TODO()).
			If().
			Then(
				clientv3.OpPut(poolkey, poolvalue, clientv3.WithLease(lease.ID)),
				clientv3.OpPut(lookupkey, lookupvalue, clientv3.WithLease(lease.ID)),
			).
			Commit()

		if err == nil {
			return nil
		} else {
			log.Warnf("lease commit failed: %v - trying again", err)
		}
	}

	return err

}

func RenewLease(mac net.HardwareAddr) error {

	kv, err := fetchOneKV("/mac/%s/ip4", mac)
	if err != nil {
		return err
	}
	if kv == nil {
		log.Warnf("/mac/%s/ip4 - does not exist", mac)
		return nil
	}

	c, err := EtcdClient()
	if err != nil {
		return err
	}
	defer c.Close()

	_, err = c.KeepAlive(context.TODO(), clientv3.LeaseID(kv.Lease))
	return err

}

func GetMember(mac string) (*Member, error) {

	m := &Member{Mac: mac}

	var err error

	m.Net, err = fetchOneValue("/mac/%s/net")
	if err != nil {
		return nil, err
	}

	//optional fields
	m.Name, err = fetchOneValue("/mac/%s/name")
	m.Ip4, err = fetchOneValue("/mac/%s/ip4")
	m.Ip6, err = fetchOneValue("/mac/%s/ip6")

	return m, nil

}

func IsStaticMember(network, mac string) (bool, error) {

	value, err := fetchOneValue("/net/%s/members/%s", network, mac)
	if err == nil && value != "" {
		return true, nil
	}

	return false, err

}

func AddNetwork(p Network) ([]clientv3.Op, error) {
	ops := []clientv3.Op{}
	ifs := []clientv3.Cmp{}

	nets := []string{p.Name}

	nets = append(nets, p.Name)

	/* dhcp4server */
	op, err := SetNetworkDhcp4Server(p.Name, p.Dhcp4Server)
	if err == nil {
		ops = append(ops, op...)
	} else {
		log.Printf("warning: %v", err)
	}

	/* dhcp6server */
	op, err = SetNetworkDhcp6Server(p.Name, p.Dhcp6Server)
	if err == nil {
		ops = append(ops, op...)
	} else {
		log.Printf("warning: %v", err)
	}

	/* subnet4 */
	op, err = SetNetworkSubnet4(p.Name, p.Subnet4)
	if err == nil {
		ops = append(ops, op...)
	} else {
		log.Printf("warning: %v", err)
	}

	/* subnet6 */
	op, err = SetNetworkSubnet6(p.Name, p.Subnet6)
	if err == nil {
		ops = append(ops, op...)
	} else {
		log.Printf("warning: %v", err)
	}

	/* ip4_range */
	if p.Ip4Range != nil {
		op, err = SetNetworkIp4Range(p.Name, *p.Ip4Range)
		if err == nil {
			ops = append(ops, op...)
		} else {
			log.Printf("warning: %v", err)
		}
	}

	/* ip6_range */
	if p.Ip6Range != nil {
		op, err = SetNetworkIp6Range(p.Name, *p.Ip6Range)
		if err == nil {
			ops = append(ops, op...)
		} else {
			log.Printf("warning: %v", err)
		}
	}

	/* gateways */
	op, err = SetNetworkGateway(p.Name, p.Gateways)
	if err == nil {
		ops = append(ops, op...)
	} else {
		log.Printf("warning: %v", err)
	}

	/* nameservers */
	op, err = SetNetworkNameservers(p.Name, p.Nameservers)
	if err == nil {
		ops = append(ops, op...)
	} else {
		log.Printf("warning: %v", err)
	}

	/* options */
	op, err = SetNetworkOptions(p.Name, p.Options)
	if err == nil {
		ops = append(ops, op...)
	} else {
		log.Printf("warning: %v", err)
	}

	/* domain */
	op, err = SetNetworkDomain(p.Name, p.Domain)
	if err == nil {
		ops = append(ops, op...)
	} else {
		log.Printf("warning: %v", err)
	}

	/* mac_range */
	if p.MacRange != nil {
		op, err = SetNetworkMacRange(p.Name, *p.MacRange)
		if err == nil {
			ops = append(ops, op...)
		} else {
			log.Printf("warning: %v", err)
		}
	}

	/* members */
	for _, m := range p.Members {
		m.Net = p.Name
		if_, op, err := SetNetworkMember(p.Name, m)
		if err == nil {
			ops = append(ops, op...)
			ifs = append(ifs, if_...)
		} else {
			log.Printf("warning: %v", err)
		}
	}

	return ops, nil

}

func SetNetworkDhcp4Server(name, server string) ([]clientv3.Op, error) {

	//TODO validate input

	key := fmt.Sprintf("/net/%s/dhcp4server", name)
	return []clientv3.Op{clientv3.OpPut(key, server)}, nil

}

func SetNetworkDhcp6Server(name, server string) ([]clientv3.Op, error) {

	//TODO validate input

	key := fmt.Sprintf("/net/%s/dhcp6server", name)
	return []clientv3.Op{clientv3.OpPut(key, server)}, nil

}

func SetNetworkSubnet4(name, subnet string) ([]clientv3.Op, error) {

	//TODO validate input

	key := fmt.Sprintf("/net/%s/subnet4", name)
	return []clientv3.Op{clientv3.OpPut(key, subnet)}, nil

}

func SetNetworkSubnet6(name, subnet string) ([]clientv3.Op, error) {

	//TODO validate input

	key := fmt.Sprintf("/net/%s/subnet6", name)
	return []clientv3.Op{clientv3.OpPut(key, subnet)}, nil

}

func SetNetworkIp4Range(name string, ip4_range AddressRange) (
	[]clientv3.Op, error) {

	//TODO validate input
	value, err := json.Marshal(ip4_range)
	if err != nil {
		return nil, err
	}

	ops := []clientv3.Op{}

	key := fmt.Sprintf("/net/%s/ip4_range", name)
	ops = append(ops, clientv3.OpPut(key, string(value)))

	if ip4_range.Begin == "" {
		return ops, nil
	}

	begin := net.ParseIP(ip4_range.Begin)
	if begin == nil {
		return nil, fmt.Errorf("invalid ip address for beginning of range")
	}

	end := net.ParseIP(ip4_range.End)
	if end == nil {
		return nil, fmt.Errorf("invalid ip address for end of range")
	}

	key = fmt.Sprintf("/net/%s/pool4/0", name)
	ops = append(ops, clientv3.OpPut(key, "00:00:00:00:00:00"))

	return ops, nil

}

func SetNetworkIp6Range(name string, ip6_range AddressRange) ([]clientv3.Op, error) {

	//TODO validate input
	value, err := json.Marshal(ip6_range)
	if err != nil {
		return nil, err
	}

	ops := []clientv3.Op{}

	key := fmt.Sprintf("/net/%s/ip6_range", name)
	ops = append(ops, clientv3.OpPut(key, string(value)))

	return ops, nil

}

func SetNetworkGateway(name string, gateway []string) ([]clientv3.Op, error) {

	//TODO validate input

	value, err := json.Marshal(gateway)
	if err != nil {
		return nil, err
	}

	key := fmt.Sprintf("/net/%s/gateways", name)
	return []clientv3.Op{clientv3.OpPut(key, string(value))}, nil

}

func SetNetworkNameservers(name string, nameservers []string) (
	[]clientv3.Op, error) {

	//TODO validate input

	value, err := json.Marshal(nameservers)
	if err != nil {
		return nil, err
	}

	key := fmt.Sprintf("/net/%s/nameservers", name)
	return []clientv3.Op{clientv3.OpPut(key, string(value))}, nil

}

func SetNetworkList(names []string) ([]clientv3.Op, error) {

	value, err := json.Marshal(names)
	if err != nil {
		return nil, err
	}

	key := "/nets"
	return []clientv3.Op{clientv3.OpPut(key, string(value))}, nil

}

func AddNetworkToList(name string, c *clientv3.Client) ([]clientv3.Op, []clientv3.Cmp, error) {

	nets, version, err := getNets(c)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get current networks: %v", err)
	}

	if netExists(nets, name) {
		return []clientv3.Op{}, []clientv3.Cmp{}, nil
	}

	nets = append(nets, name)

	buf, err := json.MarshalIndent(nets, "", "  ")
	if err != nil {
		log.Errorf("[AddNetwork] failed to marshall networks: %v", err)
		return nil, nil, fmt.Errorf("corrupt database")
	}
	ops := []clientv3.Op{clientv3.OpPut("/nets", string(buf))}
	ifs := []clientv3.Cmp{clientv3.Compare(clientv3.Version("/nets"), "=", version)}

	return ops, ifs, nil

}

func netExists(nets []string, net string) bool {
	for _, x := range nets {
		if x == net {
			return true
		}
	}
	return false
}

func getNets(c *clientv3.Client) ([]string, int64, error) {

	var err error
	if c == nil {
		c, err = EtcdClient()
		if err != nil {
			return nil, -1, err
		}
	}

	resp, err := c.Get(context.TODO(), "/nets")
	if err != nil {
		return nil, -1, fmt.Errorf("get nets query failed: %v", err)
	}
	if len(resp.Kvs) == 0 {
		return []string{}, 0, nil
	}

	var nets []string
	err = json.Unmarshal(resp.Kvs[0].Value, &nets)
	if err != nil {
		return nil, -1, fmt.Errorf("failed to parse networks: %v", err)
	}

	return nets, resp.Kvs[0].Version, nil

}

func SetNetworkDomain(name, domain string) ([]clientv3.Op, error) {

	//TODO validate input

	key := fmt.Sprintf("/net/%s/domain", name)
	return []clientv3.Op{clientv3.OpPut(key, domain)}, nil

}

func SetNetworkOptions(name string, opts []Option) ([]clientv3.Op, error) {

	//TODO validate input
	if opts == nil {
		return []clientv3.Op{}, nil
	}

	value, err := json.Marshal(opts)
	if err != nil {
		return nil, err
	}

	key := fmt.Sprintf("/net/%s/opts", name)
	return []clientv3.Op{clientv3.OpPut(key, string(value))}, nil

}

func SetNetworkMacRange(name string, mr AddressRange) ([]clientv3.Op, error) {

	mr.Begin = strings.ToLower(mr.Begin)
	mr.End = strings.ToLower(mr.End)

	//TODO validate input
	value, err := json.Marshal(mr)
	if err != nil {
		return nil, err
	}

	key := fmt.Sprintf("/net/%s/mac_range", name)
	return []clientv3.Op{clientv3.OpPut(key, string(value))}, nil

}

// SetNetworkMember returns two sets of operations, the first is a set of
// preconditions, the second is a set of operations to be executed
// if the preconditions pass
func SetNetworkMember(name string, member Member) (
	[]clientv3.Cmp, []clientv3.Op, error) {

	if member.Mac == "" || name == "" {
		return nil, nil, fmt.Errorf("must specify network name and member MAC")
	}

	member.Mac = strings.ToLower(member.Mac)

	//TODO validate input more

	ops := []clientv3.Op{}
	ifs := []clientv3.Cmp{}

	// net entry
	key := fmt.Sprintf("/net/%s/members/%s", name, member.Mac)
	ops = append(ops, clientv3.OpPut(key, member.Mac))

	// mac:net entry
	key = fmt.Sprintf("/mac/%s/net", member.Mac)
	ops = append(ops, clientv3.OpPut(key, name))

	// mac:ip4 & ip4:mac entry
	if member.Ip4 != "" {
		key = fmt.Sprintf("/mac/%s/ip4", member.Mac)
		ops = append(ops, clientv3.OpPut(key, member.Ip4))

		key = fmt.Sprintf("/ip4/%s/%s", member.Ip4, member.Mac)
		ops = append(ops, clientv3.OpPut(key, member.Mac))
	}

	// mac:ip6 & ip6:mac entry
	if member.Ip6 != "" {
		key = fmt.Sprintf("/mac/%s/ip6", member.Mac)
		ops = append(ops, clientv3.OpPut(key, member.Ip6))

		key = fmt.Sprintf("/ip6/%s", member.Ip4)
		ops = append(ops, clientv3.OpPut(key, member.Mac))
	}

	// mac:name & name:mac entry
	if member.Name != "" {
		key = fmt.Sprintf("/mac/%s/name", member.Mac)
		ops = append(ops, clientv3.OpPut(key, member.Name))

		key = fmt.Sprintf("/name/%s/%s", member.Name, member.Mac)
		ops = append(ops, clientv3.OpPut(key, member.Mac))
	}

	return ifs, ops, nil
}

func ResolveName(name string) (*Addrs, error) {

	kvs, err := fetchKvs("/name/%s", name)
	if err != nil {
		return nil, err
	}

	var ip4s []net.IP
	var ip6s []net.IP

	for _, x := range kvs {

		ip4, err := fetchOneValue("/mac/%s/ip4", string(x.Value))
		if ip4 != "" && err == nil {
			ip4s = append(ip4s, net.ParseIP(ip4))
		}

		ip6, err := fetchOneValue("/mac/%s/ip6", string(x.Value))
		if ip6 != "" && err == nil {
			ip6s = append(ip6s, net.ParseIP(ip6))
		}

	}

	result := &Addrs{}

	if len(ip4s) > 0 {
		result.Ip4 = ip4s[rand.Intn(len(ip4s))]
	}
	if len(ip6s) > 0 {
		result.Ip6 = ip6s[rand.Intn(len(ip6s))]
	}

	return result, nil

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
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

type Config struct {
	Interface string     `yaml:"interface"`
	Etcd      EtcdConfig `yaml:"etcd"`
}

/* helper functions ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~*/

func EtcdClient() (*clientv3.Client, error) {
	if Current == nil {
		LoadConfig()
	}
	connstr := fmt.Sprintf("%s:%d", Current.Etcd.Host, Current.Etcd.Port)
	return clientv3.New(clientv3.Config{Endpoints: []string{connstr}})
}

func LoadConfig() error {

	data, err := ioutil.ReadFile(ConfigPath)
	if err != nil {
		return fmt.Errorf("could not read configuration file")
	}

	err = yaml.Unmarshal(data, &Current)
	if err != nil {
		return fmt.Errorf("could not parse configuration file")
	}

	return nil

}

func rangeSize4(begin, end net.IP) (int, error) {

	b := begin.To4()
	if b == nil {
		return -1, fmt.Errorf("begin address is not ipv4")
	}

	e := end.To4()
	if e == nil {
		return -1, fmt.Errorf("end address is not ipv4")
	}

	size := (int(e[0]) << 24) - (int(b[0]) << 24)
	size += (int(e[1]) << 16) - (int(b[1]) << 16)
	size += (int(e[2]) << 8) - (int(b[2]) << 8)
	size += int(e[3]) - int(b[3])

	if size < 0 {
		return size, fmt.Errorf("invalid range: begin > end")
	}

	return size, nil

}

func (net *Network) ToProto() *proto.Network {

	_net := &proto.Network{
		Name:        net.Name,
		Subnet4:     net.Subnet4,
		Subnet6:     net.Subnet6,
		Gateways:    net.Gateways,
		Nameservers: net.Nameservers,
		Domain:      net.Domain,
		Dhcp4Server: net.Dhcp4Server,
		Dhcp6Server: net.Dhcp6Server,
	}
	if net.Ip4Range != nil {
		_net.Ip4Range = &proto.AddressRange{
			Begin: net.Ip4Range.Begin,
			End:   net.Ip4Range.End,
		}
	}
	if net.Ip6Range != nil {
		_net.Ip6Range = &proto.AddressRange{
			Begin: net.Ip6Range.Begin,
			End:   net.Ip6Range.End,
		}
	}
	if net.MacRange != nil {
		_net.MacRange = &proto.AddressRange{
			Begin: net.MacRange.Begin,
			End:   net.MacRange.End,
		}
	}
	//options
	for _, x := range net.Options {
		_net.Options = append(_net.Options, &proto.Option{Number: int32(x.Number), Value: x.Value})
	}
	//members
	for _, x := range net.Members {
		m := &proto.Member{
			Mac:  x.Mac,
			Name: x.Name,
			Ip4:  x.Ip4,
			Ip6:  x.Ip6,
			Net:  x.Net,
		}
		_net.Members = append(_net.Members, m)
	}

	return _net

}

func (n *Network) FromProto(net *proto.Network) {

	n.Name = net.Name
	n.Subnet4 = net.Subnet4
	n.Subnet6 = net.Subnet6
	n.Dhcp4Server = net.Dhcp4Server
	n.Dhcp6Server = net.Dhcp6Server
	if net.Ip4Range != nil && net.Ip4Range.Begin != "" && net.Ip4Range.End != "" {
		n.Ip4Range = &AddressRange{
			Begin: net.Ip4Range.Begin,
			End:   net.Ip4Range.End,
		}
	}
	if net.Ip6Range != nil && net.Ip6Range.Begin != "" && net.Ip6Range.End != "" {
		n.Ip6Range = &AddressRange{
			Begin: net.Ip6Range.Begin,
			End:   net.Ip6Range.End,
		}
	}
	n.Gateways = net.Gateways
	n.Nameservers = net.Nameservers
	n.Domain = net.Domain
	if net.MacRange != nil && net.MacRange.Begin != "" && net.MacRange.End != "" {
		n.MacRange = &AddressRange{
			Begin: net.MacRange.Begin,
			End:   net.MacRange.End,
		}
	}
	//options
	for _, x := range net.Options {
		n.Options = append(n.Options, Option{Number: int(x.Number), Value: x.Value})
	}
	//members
	for _, x := range net.Members {
		m := Member{
			Mac:  x.Mac,
			Name: x.Name,
			Ip4:  x.Ip4,
			Ip6:  x.Ip6,
			Net:  x.Net,
		}
		n.Members = append(n.Members, m)
	}

}

// helpers ====================================================================

func fetchOneValue(format string, args ...interface{}) (string, error) {
	kv, err := fetchOneKV(format, args...)
	if err != nil {
		return "", err
	}
	if kv == nil {
		return "", err
	}
	return string(kv.Value), nil
}

func fetchOneKV(format string, args ...interface{}) (*mvccpb.KeyValue, error) {
	c, err := EtcdClient()
	if err != nil {
		return nil, err
	}
	defer c.Close()

	key := fmt.Sprintf(format, args...)
	resp, err := c.Get(context.TODO(), key)
	if err != nil {
		return nil, err
	}

	if len(resp.Kvs) == 0 {
		return nil, nil
	}

	return resp.Kvs[0], nil
}

func fetchKvs(format string, args ...interface{}) ([]*mvccpb.KeyValue, error) {
	c, err := EtcdClient()
	if err != nil {
		return nil, err
	}
	defer c.Close()

	key := fmt.Sprintf(format, args...)
	resp, err := c.Get(context.TODO(), key, clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}

	return resp.Kvs, nil
}

// TODO: there is certianly a much faster way to do this by simultaneous
//       iteration of the key space and current space that does not
//       require actually constructing the total space.
func unusedKeyspace(kvs []*mvccpb.KeyValue, size int) mapset.Set {

	current := mapset.NewSet()
	space := mapset.NewSet()

	for _, x := range kvs {
		i, err := poolIndex(string(x.Key))
		if err != nil {
			log.Errorf("bad pool index '%s'", string(x.Key))
			continue
		}
		current.Add(i)
	}

	for i := 1; i < size; i++ {
		space.Add(i)
	}

	return space.Difference(current)

}

func poolIndex(key string) (int, error) {
	parts := strings.Split(key, "/")
	index := parts[len(parts)-1]
	return strconv.Atoi(index)
}

func (a *AddressRange) Size() int {
	begin := binary.BigEndian.Uint32([]byte(net.ParseIP(a.Begin).To4()))
	end := binary.BigEndian.Uint32([]byte(net.ParseIP(a.End).To4()))
	return int(end - begin)
}

func (a *AddressRange) Select(offset int) net.IP {
	begin := binary.BigEndian.Uint32([]byte(net.ParseIP(a.Begin).To4()))
	chosen := begin + uint32(offset)

	buf := make([]byte, 4)
	binary.BigEndian.PutUint32(buf, chosen)
	return net.IP(buf)
}

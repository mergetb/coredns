package nex

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net"

	"github.com/coreos/etcd/clientv3"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"

	proto "gitlab.com/mergetb/nex/proto"
)

/* nex database structure  ====================================================
/mac/:mac: -> { ip4, ip6, name, net }
/name/:name: -> [ mac ]
/ip4/:ip4: -> mac
/ip6/:ip6: -> mac

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
  /ip4_range
  /ip6_range
  /mac_range
  /pool6
  /pool4
  /gateways
  /domain
	/opts4
	/opts6
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
	Ip4 string
	Ip6 string
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
	Name        string       `json:"name" yaml:"name"`
	Subnet4     string       `json:"subnet4" yaml:"subnet4"`
	Subnet6     string       `json:"subnet6" yaml:"subnet6"`
	Ip4Range    AddressRange `json:"ip4_range" yaml:"ip4_range"`
	Ip6Range    AddressRange `json:"ip6_range" yaml:"ip6_range"`
	Gateways    []string     `json:"gateways" yaml:"gateways"`
	Nameservers []string     `json:"nameservers" yaml:"nameservers"`
	Options     []Option     `json:"options" yaml:"options"`
	Domain      string       `json:"domain" yaml:"domain"`
	MacRange    AddressRange `json:"mac_range" yaml:"mac_range"`
	Members     []Member     `json:"members" yaml:"members"`
}

/* Primary API functions ++++++++++++++++++++++++++++++++++++++++++++++++++++++
+
+ All of the primary API functions exist to modify the nex database. However,
+ the do not modify the database directly. They return transaction operations
+ that can be composed into transactions by higher level calling functions.
+ This is necessary to support database API operations with non-trivial data
+ dependencies.
+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++*/

func GetMember(mac string) (*Member, error) {

	c, err := EtcdClient()
	if err != nil {
		log.Fatal(err)
	}
	defer c.Close()

	key := fmt.Sprintf("/mac/%s", mac)
	resp, err := c.Get(context.TODO(), key)
	if err != nil {
		return nil, err
	}

	if len(resp.Kvs) == 0 {
		return nil, nil
	}

	m := &Member{}
	err = json.Unmarshal(resp.Kvs[0].Value, m)
	if err != nil {
		return nil, err
	}

	return m, nil

}

func IsStaticMember(network, mac string) (bool, error) {

	c, err := EtcdClient()
	if err != nil {
		log.Fatal(err)
	}
	defer c.Close()

	key := fmt.Sprintf("/net/%s/members/%s", network, mac)
	resp, err := c.Get(context.TODO(), key)
	if err != nil {
		return false, err
	}

	return len(resp.Kvs) > 0, nil

}

func AddNetwork(p Network) ([]clientv3.Op, error) {
	ops := []clientv3.Op{}
	ifs := []clientv3.Cmp{}

	nets := []string{p.Name}

	nets = append(nets, p.Name)

	/* subnet4 */
	op, err := SetNetworkSubnet4(p.Name, p.Subnet4)
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
	op, err = SetNetworkIp4Range(p.Name, p.Ip4Range)
	if err == nil {
		ops = append(ops, op...)
	} else {
		log.Printf("warning: %v", err)
	}

	/* ip6_range */
	op, err = SetNetworkIp6Range(p.Name, p.Ip6Range)
	if err == nil {
		ops = append(ops, op...)
	} else {
		log.Printf("warning: %v", err)
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
	op, err = SetNetworkMacRange(p.Name, p.MacRange)
	if err == nil {
		ops = append(ops, op...)
	} else {
		log.Printf("warning: %v", err)
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

	//TODO validate input more

	ops := []clientv3.Op{}
	ifs := []clientv3.Cmp{}

	// net entry
	key := fmt.Sprintf("/net/%s/members/%s", name, member.Mac)
	ops = append(ops, clientv3.OpPut(key, member.Mac))

	// mac entry
	key = fmt.Sprintf("/mac/%s", member.Mac)
	value, err := json.Marshal(member)
	if err != nil {
		return nil, nil, err
	}
	ops = append(ops, clientv3.OpPut(key, string(value)))

	// name entry
	if member.Name != "" {
		key = fmt.Sprintf("/name/%s", member.Name)

		cli, err := EtcdClient()
		if err != nil {
			return nil, nil, err
		}
		defer cli.Close()

		resp, err := cli.Get(context.TODO(), key)
		if err != nil {
			return nil, nil, err
		}

		var macs []string
		if len(resp.Kvs) > 0 {
			json.Unmarshal(resp.Kvs[0].Value, &macs)
		}
		current_value, err := json.Marshal(macs)
		macs = append(macs, member.Mac)
		if err != nil {
			return nil, nil, err
		}
		new_value, err := json.Marshal(macs)

		if len(resp.Kvs) > 0 {
			ifs = append(ifs,
				clientv3.Compare(clientv3.Value(key), "=", string(current_value)))
		}
		ops = append(ops, clientv3.OpPut(key, string(new_value)))
	}

	// ip4 entry
	if member.Ip4 != "" {
		key = fmt.Sprintf("/ip4/%s", member.Ip4)
		ops = append(ops, clientv3.OpPut(key, member.Mac))
	}

	// ip6 entry
	if member.Ip6 != "" {
		key = fmt.Sprintf("/ip6/%s", member.Ip4)
		ops = append(ops, clientv3.OpPut(key, member.Mac))
	}

	return ifs, ops, nil
}

func ResolveName(name string) (*Addrs, error) {

	c, err := EtcdClient()
	if err != nil {
		return nil, err
	}
	defer c.Close()

	resp, err := c.Get(context.TODO(), fmt.Sprintf("/name/%s", name))
	if err != nil {
		return nil, err
	}

	if len(resp.Kvs) == 0 {
		return nil, nil
	}

	macs_json := resp.Kvs[0].Value
	var macs []string
	err = json.Unmarshal(macs_json, &macs)
	if err != nil {
		return nil, err
	}
	if len(macs) == 0 {
		log.Warnf("empty maclist for %s", name)
	}

	//collect all the ip4s and ip6s associated with this mac
	var ip4s, ip6s []string
	for _, mac := range macs {

		resp, err = c.Get(context.TODO(), fmt.Sprintf("/mac/%s", mac))
		if err != nil {
			return nil, err
		}

		if len(resp.Kvs) == 0 {
			return nil, nil
		}
		member_json := resp.Kvs[0].Value

		var member Member
		err = json.Unmarshal(member_json, &member)
		if err != nil {
			continue
		}

		if member.Ip4 != "" {
			ip4s = append(ip4s, member.Ip4)
		}
		if member.Ip6 != "" {
			ip6s = append(ip6s, member.Ip6)
		}

	}

	//randomly choose and ip4 and ip6 from the pool
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

type Config struct {
	Host string
	Port int
}

/* helper functions ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~*/

func EtcdClient() (*clientv3.Client, error) {
	c, err := loadConfig()
	if err != nil || c == nil {
		return nil, err
	}
	log.Infof("connecting to datastore %s:%d", c.Host, c.Port)
	connstr := fmt.Sprintf("%s:%d", c.Host, c.Port)
	return clientv3.New(clientv3.Config{Endpoints: []string{connstr}})
}

func loadConfig() (*Config, error) {

	data, err := ioutil.ReadFile("/etc/merge/nex.yml")
	if err != nil {
		return nil, fmt.Errorf("cnuld not read configuration file")
	}

	c := &Config{}
	err = yaml.Unmarshal(data, c)
	if err != nil {
		return nil, fmt.Errorf("could not parse configuration file")
	}

	log.Info("read configuration file")
	return c, nil

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
		Name:    net.Name,
		Subnet4: net.Subnet4,
		Subnet6: net.Subnet6,
		Ip4Range: &proto.AddressRange{
			Begin: net.Ip4Range.Begin,
			End:   net.Ip4Range.End,
		},
		Ip6Range: &proto.AddressRange{
			Begin: net.Ip6Range.Begin,
			End:   net.Ip6Range.End,
		},
		Gateways:    net.Gateways,
		Nameservers: net.Nameservers,
		Domain:      net.Domain,
		MacRange: &proto.AddressRange{
			Begin: net.MacRange.Begin,
			End:   net.MacRange.End,
		},
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
	n.Ip4Range = AddressRange{
		Begin: net.Ip4Range.Begin,
		End:   net.Ip4Range.End,
	}
	n.Ip6Range = AddressRange{
		Begin: net.Ip6Range.Begin,
		End:   net.Ip6Range.End,
	}
	n.Gateways = net.Gateways
	n.Nameservers = net.Nameservers
	n.Domain = net.Domain
	n.MacRange = AddressRange{
		Begin: net.MacRange.Begin,
		End:   net.MacRange.End,
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

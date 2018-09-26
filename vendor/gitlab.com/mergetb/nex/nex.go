package nex

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"

	"github.com/coreos/etcd/clientv3"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

// TODO =======================================================================
/*

At the end of the day we need the following fast lookup mappings. All of which
are tied to temporal leases in they dynamic case.

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
         ....
	/members/:mac: -> :mac:
 When the etcd lease expires the key disappears and may be reallocated

*/
// \TODO ^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^

/* nex database structure +++++++++++++++++++++++++++++++++++++++++++++++++++++
+
+   # A list of network names
+		/nets
+
+   /net/<name>/subnet
+              /ip4_range
+              /ip6_range
+              /mac_range
+              /pool6
+              /pool4
+              /gateways
+              /domain
+              /dns
+              /tftp
+   # Pools hold dynamic address range usage via leases. Each lease is an
+   # integer offset from the base address e.g.
+		#		/pool4/1 -> <mac>
+   #         /2 -> <mac>
+   #         /5 -> <mac>
+   #         ....
+   # When the etcd lease expires the key disappears and may be reallocated
+
+   # Address and options information about a MAC. Interface level options
+   # override named group options.
+   /ifx/<mac>/net
+             /ip4
+             /ip6
+							/fqdn
+             /opts4
+             /opts6
+
+   # Map IP4 addresses to MACs and FQDNs
+   /addr/<ip4>/mac
+              /fqdn
+
+   # Map IP6 addresses to MACs and FQDNs
+   /addr/<ip6>/mac
+               /fqdn
+
+   # Map FQDNs to IP addresses and MACs
+   /name/<fqdn>/ip4
+               /ip6
+
+
+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++*/

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

func SetIfxMacIp4(mac, ip4 string) ([]clientv3.Op, error) {

	//TODO validate mac and ip4

	key := fmt.Sprintf("/ifx/%s/ip4", mac)
	return []clientv3.Op{clientv3.OpPut(key, ip4)}, nil

}

func SetIfxMacIp6(mac, ip6 string) ([]clientv3.Op, error) {

	//TODO validate mac and ip6

	key := fmt.Sprintf("/ifx/%s/ip6", mac)
	return []clientv3.Op{clientv3.OpPut(key, ip6)}, nil

}

func SetIfxMacOpts4(mac string, opts []Opt4) ([]clientv3.Op, error) {

	//TODO validate mac and opts

	value, err := json.Marshal(opts)
	if err != nil {
		return nil, err
	}

	key := fmt.Sprintf("/ifx/%s/opts4", mac)
	return []clientv3.Op{clientv3.OpPut(key, string(value))}, nil

}

func SetIfxMacOpts6(mac string, opts []Opt6) ([]clientv3.Op, error) {

	//TODO validate mac and opts

	value, err := json.Marshal(opts)
	if err != nil {
		return nil, err
	}

	key := fmt.Sprintf("/ifx/%s/opts6", mac)
	return []clientv3.Op{clientv3.OpPut(key, string(value))}, nil

}

func SetAddr4Mac(ip4, mac string) ([]clientv3.Op, error) {

	//TODO validate mac and ip4

	key := fmt.Sprintf("/addr/%s/mac", ip4)
	return []clientv3.Op{clientv3.OpPut(key, mac)}, nil

}

func SetAddr6Mac(ip6, mac string) ([]clientv3.Op, error) {

	//TODO validate mac and ip6

	key := fmt.Sprintf("/addr/%s/mac", ip6)
	return []clientv3.Op{clientv3.OpPut(key, mac)}, nil

}

func SetAddr4Fqdn(ip4, fqdn string) ([]clientv3.Op, error) {

	//TODO validate name and ip4

	key := fmt.Sprintf("/addr/%s/fqdn", ip4)
	return []clientv3.Op{clientv3.OpPut(key, fqdn)}, nil

}

func SetAddr6Fqdn(ip6, fqdn string) ([]clientv3.Op, error) {

	//TODO validate name and ip6

	key := fmt.Sprintf("/addr/%s/fqdn", ip6)
	return []clientv3.Op{clientv3.OpPut(key, fqdn)}, nil

}

func SetFqdnAddr4(fqdn, ip4 string) ([]clientv3.Op, error) {

	//TODO validate name and ip4

	key := fmt.Sprintf("/name/%s/ip4", fqdn)
	return []clientv3.Op{clientv3.OpPut(key, ip4)}, nil

}

func SetFqdnAddr6(fqdn, ip6 string) ([]clientv3.Op, error) {

	//TODO validate name and ip6

	key := fmt.Sprintf("/name/%s/ip6", fqdn)
	return []clientv3.Op{clientv3.OpPut(key, ip6)}, nil

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

	/*
		key1 := fmt.Sprintf("/ifx/%s/net", member.Mac)
		key2 := fmt.Sprintf("/net/%s/members/%s", name, member.Mac)
		ops = append(ops, clientv3.OpPut(key1, name))
		ops = append(ops, clientv3.OpPut(key2, member.Mac))

		if member.Name != "" {
			key := fmt.Sprintf("/ifx/%s/name", member.Mac)
			ops = append(ops, clientv3.OpPut(key, member.Name))
		}

		if member.Ip4 != "" {
			key := fmt.Sprintf("/ifx/%s/ip4", member.Mac)
			ops = append(ops, clientv3.OpPut(key, member.Ip4))
		}

		if member.Ip6 != "" {
			key := fmt.Sprintf("/ifx/%s/ip6", member.Mac)
			ops = append(ops, clientv3.OpPut(key, member.Ip6))
		}
	*/

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

		resp, err := cli.Get(context.TODO(), key)
		if err != nil {
			return nil, nil, err
		}

		var macs []string
		if len(resp.Kvs) > 0 {
			json.Unmarshal(resp.Kvs[0].Value, &macs)
		}
		macs = append(macs, member.Mac)
		value, err = json.Marshal(macs)
		if err != nil {
			return nil, nil, err
		}

		if len(resp.Kvs) > 0 {
			ifs = append(ifs,
				clientv3.Compare(clientv3.Value(key), "=", string(value)))
		}
		ops = append(ops, clientv3.OpPut(key, string(value)))
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
	//XXX
	//  arbitrarily selecting the first mac in the list
	//TODO
	//	actually find all IPs for all macs and randomly choose one as
	//  is the default behavior for round-robin dns
	if len(macs) == 0 {
		log.Warnf("empty maclist for %s", name)
	}
	mac := macs[0]

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
		return nil, err
	}

	return &Addrs{Ip4: member.Ip4, Ip6: member.Ip6}, nil

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
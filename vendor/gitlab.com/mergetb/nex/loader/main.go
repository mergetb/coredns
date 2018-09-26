package main

import (
	"context"
	"io/ioutil"
	"log"
	"os"

	"github.com/coreos/etcd/clientv3"
	"gopkg.in/yaml.v2"

	"gitlab.com/mergetb/nex"
)

func main() {
	log.SetFlags(0)
	if len(os.Args) < 2 {
		log.Fatal("usage: loader <spec>")
	}

	spec, err := loadSpec(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("spec: %#v", spec)

	setSpec(spec)
}

type Spec struct {
	Networks []nex.Network
}

func setSpec(spec *Spec) error {

	ops := []clientv3.Op{}
	ifs := []clientv3.Cmp{}

	nets := []string{}

	for _, p := range spec.Networks {

		nets = append(nets, p.Name)

		/* subnet4 */
		op, err := nex.SetNetworkSubnet4(p.Name, p.Subnet4)
		if err == nil {
			ops = append(ops, op...)
		} else {
			log.Printf("warning: %v", err)
		}

		/* subnet6 */
		op, err = nex.SetNetworkSubnet6(p.Name, p.Subnet6)
		if err == nil {
			ops = append(ops, op...)
		} else {
			log.Printf("warning: %v", err)
		}

		/* ip4_range */
		op, err = nex.SetNetworkIp4Range(p.Name, p.Ip4Range)
		if err == nil {
			ops = append(ops, op...)
		} else {
			log.Printf("warning: %v", err)
		}

		/* ip6_range */
		op, err = nex.SetNetworkIp6Range(p.Name, p.Ip6Range)
		if err == nil {
			ops = append(ops, op...)
		} else {
			log.Printf("warning: %v", err)
		}

		/* gateways */
		op, err = nex.SetNetworkGateway(p.Name, p.Gateways)
		if err == nil {
			ops = append(ops, op...)
		} else {
			log.Printf("warning: %v", err)
		}

		/* nameservers */
		op, err = nex.SetNetworkNameservers(p.Name, p.Nameservers)
		if err == nil {
			ops = append(ops, op...)
		} else {
			log.Printf("warning: %v", err)
		}

		/* options */
		op, err = nex.SetNetworkOptions(p.Name, p.Options)
		if err == nil {
			ops = append(ops, op...)
		} else {
			log.Printf("warning: %v", err)
		}

		/* domain */
		op, err = nex.SetNetworkDomain(p.Name, p.Domain)
		if err == nil {
			ops = append(ops, op...)
		} else {
			log.Printf("warning: %v", err)
		}

		/* mac_range */
		op, err = nex.SetNetworkMacRange(p.Name, p.MacRange)
		if err == nil {
			ops = append(ops, op...)
		} else {
			log.Printf("warning: %v", err)
		}

		/* members */
		for _, m := range p.Members {
			if_, op, err := nex.SetNetworkMember(p.Name, m)
			if err == nil {
				ops = append(ops, op...)
				ifs = append(ifs, if_...)
			} else {
				log.Printf("warning: %v", err)
			}
		}

	}

	op, err := nex.SetNetworkList(nets)
	if err == nil {
		ops = append(ops, op...)
	} else {
		log.Fatal("error: could not set network name list %v", err)
	}

	c, err := nex.EtcdClient()
	if err != nil {
		log.Fatal(err)
	}

	var txn clientv3.Txn
	txn = c.Txn(context.TODO()).If(ifs...).Then(ops...)
	_, err = txn.Commit()
	// TODO retry logic on concurrency collision
	if err != nil {
		log.Fatal(err)
	}

	return nil

}

func loadSpec(file string) (*Spec, error) {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}

	s := &Spec{}
	err = yaml.Unmarshal(data, s)
	if err != nil {
		return nil, err
	}

	return s, nil
}

func littleTest() {
	ops, err := nex.SetIfxMacIp4("00:11:44:77:00:22", "10.47.0.2/24")
	if err != nil {
		log.Fatal(err)
	}

	c, err := nex.EtcdClient()
	if err != nil {
		log.Fatal(err)
	}

	txn := c.Txn(context.TODO()).Then(ops...)
	_, err = txn.Commit()
	if err != nil {
		log.Fatal(err)
	}
}

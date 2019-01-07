package nex

import (
	"context"
	"encoding/json"
	"time"

	etcd "github.com/coreos/etcd/clientv3"
)

func AddNetwork(n *Network) error {

	var objs []Object
	if n.Range4 != nil {
		p := &Pool{Net: n.Name}
		p.Size = n.Range4.Size()
		objs = append(objs, NewPoolObj(p))
	}

	objs = append(objs, NewNetworkObj(n))

	err := WriteObjects(objs)
	if err != nil {
		return err
	}

	return nil

}

func GetNetworks() ([]*Network, error) {

	var nets []*Network

	err := withEtcd(func(c *etcd.Client) error {

		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		resp, err := c.Get(ctx, "/net", etcd.WithPrefix())
		cancel()
		if err != nil {
			return err
		}

		for _, kv := range resp.Kvs {
			net := &Network{}
			err := json.Unmarshal(kv.Value, &net)
			if err != nil {
				return err
			}
			nets = append(nets, net)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return nets, nil
}

func DeleteNetwork(name string) error {

	// Get the networks members
	members, err := GetMembers(name)
	if err != nil {
		return err
	}

	// Gather all the member index objects
	var objects []Object
	for _, m := range members {
		objects = append(objects, NewMacIndex(m))
	}
	objects = DeleteMemberObjects(objects)

	// Get the pool if there is one
	objects = append(objects, NewPoolObj(&Pool{Net: name}))

	// Add the network object itself to the list
	objects = append(objects, NewNetworkObj(&Network{Name: name}))

	// Clear out everything in one txn
	return DeleteObjects(objects)

}

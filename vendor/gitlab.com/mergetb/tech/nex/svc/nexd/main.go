package main

import (
	"context"
	"fmt"
	"net"
	"strings"

	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"

	"gitlab.com/mergetb/tech/nex/pkg"
)

type NexD struct{}

func main() {

	log.Printf("nexd %s\n", nex.Version)
	log.SetLevel(log.DebugLevel)

	go nex.RunLeaseManager()

	grpcServer := grpc.NewServer()
	nex.RegisterNexServer(grpcServer, &NexD{})

	err := nex.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}

	l, err := net.Listen("tcp", nex.Current.Nexd.Listen)
	if err != nil {
		log.Fatal("failed to listen: %#v", err)
	}

	log.Infof("Listening on tcp://%s", nex.Current.Nexd.Listen)
	grpcServer.Serve(l)

}

/***~~~~ Networks ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~*/

func (s *NexD) GetNetwork(
	ctx context.Context, e *nex.GetNetworkRequest,
) (*nex.GetNetworkResponse, error) {

	net := nex.NewNetworkObj(&nex.Network{Name: e.Name})
	err := nex.Read(net)
	if err != nil {
		return nil, err
	}
	return &nex.GetNetworkResponse{Net: net.Network}, nil

}

func (s *NexD) GetNetworks(
	ctx context.Context, e *nex.GetNetworksRequest,
) (*nex.GetNetworksResponse, error) {

	list, err := nex.GetNetworks()
	if err != nil {
		return nil, err
	}
	var result []string
	for _, net := range list {
		result = append(result, net.Name)
	}

	return &nex.GetNetworksResponse{Nets: result}, nil

}

func (s *NexD) AddNetwork(
	ctx context.Context, e *nex.AddNetworkRequest,
) (*nex.AddNetworkResponse, error) {

	err := nex.AddNetwork(e.Network)
	if err != nil {
		return nil, err
	}

	return &nex.AddNetworkResponse{}, nil

}

func (s *NexD) DeleteNetwork(
	ctx context.Context, e *nex.DeleteNetworkRequest,
) (*nex.DeleteNetworkResponse, error) {

	err := nex.DeleteNetwork(e.Name)
	if err != nil {
		return nil, err
	}

	return &nex.DeleteNetworkResponse{}, nil

}

/***~~~~~~~ Members ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~*/

func (s *NexD) GetMembers(
	ctx context.Context, e *nex.GetMembersRequest,
) (*nex.GetMembersResponse, error) {

	members, err := nex.GetMembers(e.Network)
	if err != nil {
		return nil, err
	}

	return &nex.GetMembersResponse{Members: members}, nil

}

func (s *NexD) AddMembers(
	ctx context.Context, e *nex.MemberList,
) (*nex.AddMembersResponse, error) {

	net := nex.NewNetworkObj(&nex.Network{Name: e.Net})
	err := nex.Read(net)
	if err != nil {
		return nil, err
	}

	var objects []nex.Object
	for _, m := range e.List {

		if err := nex.ValidateMac(m.Mac); err != nil {
			return nil, err
		}
		m.Mac = strings.ToLower(m.Mac)

		m.Net = e.Net

		objects = append(objects, []nex.Object{
			nex.NewMacIndex(m),
			nex.NewNetIndex(m),
		}...)

		if m.Ip4 != nil {
			if net.Range4 != nil {
				return nil, fmt.Errorf("cannot assign static IP to pool member")
			}
			objects = append(objects, nex.NewIp4Index(m))
		}
		if m.Name != "" {
			m.Name = m.Name + "." + net.Domain
			objects = append(objects, nex.NewNameIndex(m))
		}

	}

	err = nex.WriteObjects(objects)
	if err != nil {
		if nex.IsTxnFailed(err) {
			return nil, fmt.Errorf("some or all members already exist")
		}
		return nil, err
	}

	return &nex.AddMembersResponse{}, nil

}

func (s *NexD) UpdateMembers(
	ctx context.Context, e *nex.UpdateList,
) (*nex.UpdateMembersResponse, error) {

	net := nex.NewNetworkObj(&nex.Network{Name: e.Net})
	err := nex.Read(net)
	if err != nil {
		return nil, err
	}

	// Read the current state of the objects being updated in a single shot txn.
	var otx nex.ObjectTx
	for _, u := range e.List {

		otx.Put = append(otx.Put, nex.NewMacIndex(&nex.Member{Mac: u.Mac}))

	}
	err = nex.ReadObjects(otx.Put)
	if err != nil {
		return nil, err
	}

	// Update the objects in a single shot txn. The txn will fail if any of the
	// objects have been modified since reading.
	for i, object := range otx.Put {

		m := object.(*nex.MacIndex).Member
		update := e.List[i]
		if update.Name != nil {
			if m.Name != update.Name.GetValue() {
				otx.Delete = append(otx.Delete, nex.NewNameIndex(&nex.Member{Name: m.Name}))
				otx.Put = append(otx.Put, nex.NewNameIndex(m))
			}
			m.Name = update.Name.GetValue() + "." + net.Domain
		}
		if update.Ip4 != nil {
			if net.Range4 != nil {
				return nil, fmt.Errorf("cannot assign static IP to pool member")
			}
			if m.Ip4 == nil {
				otx.Delete = append(otx.Delete, nex.NewIp4Index(&nex.Member{Ip4: m.Ip4}))
				otx.Put = append(otx.Put, nex.NewIp4Index(m))
			}
			m.Ip4 = update.Ip4
		}

	}
	err = nex.RunObjectTx(otx)
	if err != nil {
		return nil, err
	}

	return &nex.UpdateMembersResponse{}, nil

}

func (s *NexD) DeleteMembers(
	ctx context.Context, e *nex.DeleteMembersRequest,
) (*nex.DeleteMembersResponse, error) {

	err := nex.DeleteMembers(e.List)
	if err != nil {
		return nil, err
	}

	return &nex.DeleteMembersResponse{}, nil
}

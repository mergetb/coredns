package main

import (
	"context"
	"fmt"
	"net"

	"github.com/coreos/etcd/clientv3"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"

	"gitlab.com/mergetb/nex"
	proto "gitlab.com/mergetb/nex/proto"
)

type NexD struct{}

/* ip4 ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~*/

func (s *NexD) GetIp4(
	ctx context.Context, e *proto.GetIp4Request,
) (*proto.GetIp4Response, error) {

	lease := &proto.Lease{Mac: e.Mac}
	result := &proto.GetIp4Response{Lease: lease}

	member, err := nex.GetMember(e.Mac)
	if err != nil {
		log.Errorf("[GetIp4] GetMember: (%s) - %v", e.Mac, err)
		return nil, fmt.Errorf("error retrieving member info")
	}
	if member == nil {
		return result, nil
	}
	result.Lease.Addr = member.Ip4

	isStatic, err := nex.IsStaticMember(member.Net, e.Mac)
	if err != nil {
		log.Errorf("[GetIp4] IsStaticMember: (%s) - %v", e.Mac, err)
		return nil, fmt.Errorf("error retrieving network info")
	}
	if isStatic {
		result.Lease.Type = proto.Lease_Static
	} else {
		result.Lease.Type = proto.Lease_Dynamic
	}

	return result, nil

}

func (s *NexD) GetIp4S(
	ctx context.Context, e *proto.GetIp4SRequest,
) (*proto.GetIp4SResponse, error) {

	return nil, nil

}

func (s *NexD) SetIp4(
	ctx context.Context, e *proto.SetIp4Request,
) (*proto.SetIp4Response, error) {

	return nil, nil

}

func (s *NexD) DelIp4(
	ctx context.Context, e *proto.DelIp4Request,
) (*proto.DelIp4Response, error) {

	return nil, nil

}

/* ip6 ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~*/

func (s *NexD) GetIp6(
	ctx context.Context, e *proto.GetIp6Request,
) (*proto.GetIp6Response, error) {

	return nil, nil

}

func (s *NexD) GetIp6S(
	ctx context.Context, e *proto.GetIp6SRequest,
) (*proto.GetIp6SResponse, error) {

	return nil, nil

}

func (s *NexD) SetIp6(
	ctx context.Context, e *proto.SetIp6Request,
) (*proto.SetIp6Response, error) {

	return nil, nil

}

func (s *NexD) DelIp6(
	ctx context.Context, e *proto.DelIp6Request,
) (*proto.DelIp6Response, error) {

	return nil, nil

}

/* name ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~*/
func (s *NexD) GetName(
	ctx context.Context, e *proto.GetNameRequest,
) (*proto.GetNameResponse, error) {

	result := &proto.GetNameResponse{}

	member, err := nex.GetMember(e.Mac)
	if err != nil {
		log.Errorf("[GetIp4] GetMember: (%s) - %v", e.Mac, err)
		return nil, fmt.Errorf("error retrieving member info")
	}
	if member == nil {
		return result, nil
	}
	result.Name = member.Name

	return result, nil

}

func (s *NexD) GetNames(
	ctx context.Context, e *proto.GetNamesRequest,
) (*proto.GetNamesResponse, error) {

	return nil, nil

}

func (s *NexD) SetName(
	ctx context.Context, e *proto.SetNameRequest,
) (*proto.SetNameResponse, error) {

	return nil, nil

}

func (s *NexD) DelName(
	ctx context.Context, e *proto.DelNameRequest,
) (*proto.DelNameResponse, error) {

	return nil, nil

}

/* subnet4 ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~*/
func (s *NexD) GetSubnet4(
	ctx context.Context, e *proto.GetSubnet4Request,
) (*proto.GetSubnet4Response, error) {

	return nil, nil

}

func (s *NexD) SetSubnet4(
	ctx context.Context, e *proto.SetSubnet4Request,
) (*proto.SetSubnet4Response, error) {

	return nil, nil

}

func (s *NexD) DelSubnet4(
	ctx context.Context, e *proto.DelSubnet4Request,
) (*proto.DelSubnet4Response, error) {

	return nil, nil

}

/* subnet6 ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~*/
func (s *NexD) GetSubnet6(
	ctx context.Context, e *proto.GetSubnet6Request,
) (*proto.GetSubnet6Response, error) {

	return nil, nil

}

func (s *NexD) SetSubnet6(
	ctx context.Context, e *proto.SetSubnet6Request,
) (*proto.SetSubnet6Response, error) {

	return nil, nil

}

func (s *NexD) DelSubnet6(
	ctx context.Context, e *proto.DelSubnet6Request,
) (*proto.DelSubnet6Response, error) {

	return nil, nil

}

/* range4 ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~*/
func (s *NexD) GetRange4(
	ctx context.Context, e *proto.GetRange4Request,
) (*proto.GetRange4Response, error) {

	return nil, nil

}

func (s *NexD) SetRange4(
	ctx context.Context, e *proto.SetRange4Request,
) (*proto.SetRange4Response, error) {

	return nil, nil

}

func (s *NexD) DelRange4(
	ctx context.Context, e *proto.DelRange4Request,
) (*proto.DelRange4Response, error) {

	return nil, nil

}

/* range6 ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~*/
func (s *NexD) GetRange6(
	ctx context.Context, e *proto.GetRange6Request,
) (*proto.GetRange6Response, error) {

	return nil, nil

}

func (s *NexD) SetRange6(
	ctx context.Context, e *proto.SetRange6Request,
) (*proto.SetRange6Response, error) {

	return nil, nil

}

func (s *NexD) DelRange6(
	ctx context.Context, e *proto.DelRange6Request,
) (*proto.DelRange6Response, error) {

	return nil, nil

}

/* gateways ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~*/
func (s *NexD) GetGateways(
	ctx context.Context, e *proto.GetGatewaysRequest,
) (*proto.GetGatewaysResponse, error) {

	return nil, nil

}

func (s *NexD) SetGateways(
	ctx context.Context, e *proto.SetGatewaysRequest,
) (*proto.SetGatewaysResponse, error) {

	return nil, nil

}

func (s *NexD) DelGateways(
	ctx context.Context, e *proto.DelGatewaysRequest,
) (*proto.DelGatewaysResponse, error) {

	return nil, nil

}

/* domain ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~*/
func (s *NexD) GetDomain(
	ctx context.Context, e *proto.GetDomainRequest,
) (*proto.GetDomainResponse, error) {

	return nil, nil

}

func (s *NexD) SetDomain(
	ctx context.Context, e *proto.SetDomainRequest,
) (*proto.SetDomainResponse, error) {

	return nil, nil

}

func (s *NexD) DelDomain(
	ctx context.Context, e *proto.DelDomainRequest,
) (*proto.DelDomainResponse, error) {

	return nil, nil

}

/* option ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~*/
func (s *NexD) GetOption(
	ctx context.Context, e *proto.GetOptionRequest,
) (*proto.GetOptionResponse, error) {

	return nil, nil

}

func (s *NexD) GetOptions(
	ctx context.Context, e *proto.GetOptionsRequest,
) (*proto.GetOptionsResponse, error) {

	return nil, nil

}

func (s *NexD) SetOption(
	ctx context.Context, e *proto.SetOptionRequest,
) (*proto.SetOptionResponse, error) {

	return nil, nil

}

func (s *NexD) DelOption(
	ctx context.Context, e *proto.DelOptionRequest,
) (*proto.DelOptionResponse, error) {

	return nil, nil

}

/* membership ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~*/
func (s *NexD) AddMembers(
	ctx context.Context, e *proto.AddMemberRequest,
) (*proto.AddMemberResponse, error) {

	return nil, nil

}

func (s *NexD) DelMembers(
	ctx context.Context, e *proto.DelMemberRequest,
) (*proto.DelMemberResponse, error) {

	return nil, nil

}

func (s *NexD) GetMembers(
	ctx context.Context, e *proto.GetMemberRequest,
) (*proto.GetMemberResponse, error) {

	return nil, nil

}

/* network ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~*/
func (s *NexD) AddNetwork(
	ctx context.Context, e *proto.AddNetworkRequest,
) (*proto.AddNetworkResponse, error) {

	_net := &nex.Network{}
	_net.FromProto(e.Network)

	ops, err := nex.AddNetwork(*_net)
	if err != nil {
		log.Errorf("[AddNetwork] add-error: %v", err)
		return nil, fmt.Errorf("failed to add network")
	}

	c, err := nex.EtcdClient()
	if err != nil {
		log.Errorf("[AddNetwork] failed to connect to db: %v", err)
		return nil, fmt.Errorf("failed to connect to db")
	}
	defer c.Close()

	_ops, ifs, err := nex.AddNetworkToList(_net.Name, c)
	if err != nil {
		log.Errorf("[AddNetwork] failed to add network to list: %v", err)
		return nil, fmt.Errorf("failed to add network to global list")
	}
	ops = append(ops, _ops...)

	txn := c.Txn(context.TODO()).If(ifs...).Then(ops...)
	_, err = txn.Commit()
	//TODO handle concurrency collision with retry logic
	if err != nil {
		log.Errorf("[AddNetwork] commit-error: %v", err)
		return nil, fmt.Errorf("failed to commit new network")
	}

	return &proto.AddNetworkResponse{}, nil

}

func (s *NexD) DelNetwork(
	ctx context.Context, e *proto.DelNetworkRequest,
) (*proto.DelNetworkResponse, error) {

	c, err := nex.EtcdClient()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to db")
	}
	defer c.Close()

	key := fmt.Sprintf("/net/%s", e.Name)
	_, err = c.Delete(context.TODO(), key, clientv3.WithPrefix())
	if err != nil {
		log.Errorf("[DelNetwork] %v", err)
		return nil, fmt.Errorf("failed to delete network")
	}

	return &proto.DelNetworkResponse{}, nil

}

func (s *NexD) GetNetwork(
	ctx context.Context, e *proto.GetNetworkRequest,
) (*proto.GetNetworkResponse, error) {

	return nil, nil

}

func (s *NexD) GetNetworks(
	ctx context.Context, e *proto.GetNetworksRequest,
) (*proto.GetNetworksResponse, error) {

	return nil, nil

}

func main() {

	fmt.Println("nexd 0.1.0")

	grpcServer := grpc.NewServer()
	proto.RegisterNexServer(grpcServer, &NexD{})

	l, err := net.Listen("tcp", "0.0.0.0:6000")
	if err != nil {
		log.Fatal("failed to listen: %#v", err)
	}

	log.Info("Listening on tcp://0.0.0.0:6000")
	grpcServer.Serve(l)

}

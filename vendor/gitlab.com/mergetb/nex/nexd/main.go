package main

import (
	"context"
	"fmt"
	"net"

	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"

	proto "gitlab.com/mergetb/nex/proto"
)

type NexD struct{}

/* ip4 ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~*/

func (s *NexD) GetIp4(
	ctx context.Context, e *proto.GetIp4Request,
) (*proto.GetIp4Response, error) {

	return nil, nil

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

	return nil, nil

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

func main() {

	fmt.Println("nexd")

	grpcServer := grpc.NewServer()
	proto.RegisterNexServer(grpcServer, &NexD{})

	l, err := net.Listen("tcp", "0.0.0.0:6000")
	if err != nil {
		log.Fatal("failed to listen: %#v", err)
	}

	log.Info("Listening on tcp://0.0.0.0:6000")
	grpcServer.Serve(l)

}

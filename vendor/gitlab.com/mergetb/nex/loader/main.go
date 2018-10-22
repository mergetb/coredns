package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"google.golang.org/grpc"
	"gopkg.in/yaml.v2"

	"gitlab.com/mergetb/nex"
	proto "gitlab.com/mergetb/nex/proto"
)

var host string = "localhost"

func main() {
	log.SetFlags(0)
	if len(os.Args) < 2 {
		log.Fatal("usage: loader <spec> [server]")
	}

	spec, err := loadSpec(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}

	if len(os.Args) > 2 {
		host = os.Args[2]
	}

	setSpec(spec)
}

type Spec struct {
	Networks []nex.Network
}

func addNetwork(net nex.Network) {

	conn, err := grpc.Dial(fmt.Sprintf("%s:6000", host), grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}
	cli := proto.NewNexClient(conn)

	_, err = cli.AddNetwork(context.TODO(), &proto.AddNetworkRequest{
		Network: net.ToProto(),
	})
	if err != nil {
		log.Fatal(err)
	}

}

func setSpec(spec *Spec) error {

	for _, n := range spec.Networks {
		addNetwork(n)
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

package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"strings"

	"github.com/spf13/cobra"
	"gitlab.com/mergetb/tech/nex/pkg"
	"gopkg.in/yaml.v2"
)

func networkCmds(get, set, add, delete *cobra.Command) {

	getNetworks := &cobra.Command{
		Use:   "networks",
		Short: "Get network list",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			getNetworks()
		},
	}
	get.AddCommand(getNetworks)

	getNetwork := &cobra.Command{
		Use:   "network",
		Short: "Get network list",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			getNetwork(args[0])
		},
	}
	get.AddCommand(getNetwork)

	setNetwork := &cobra.Command{
		Use:   "network [yaml spec]",
		Short: "Set network properties from yaml spec",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			setNetwork(args[0])
		},
	}
	set.AddCommand(setNetwork)

	deleteNetwork := &cobra.Command{
		Use:   "network [name]",
		Short: "Delete a network and all it's members",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			deleteNetwork(args[0])
		},
	}
	delete.AddCommand(deleteNetwork)

}

func getNetworks() {
	withClient(func(cli nex.NexClient) error {

		resp, err := cli.GetNetworks(ctx, &nex.GetNetworksRequest{})
		if err != nil {
			grpcFatal(err)
		}

		for _, n := range resp.Nets {
			log.Println(n)
		}

		return nil

	})
}

func getNetwork(name string) {
	withClient(func(cli nex.NexClient) error {

		resp, err := cli.GetNetwork(ctx, &nex.GetNetworkRequest{
			Name: name,
		})
		if err != nil {
			grpcFatal(err)
		}

		fmt.Fprintf(tw, "name:\t%s\n", resp.Net.Name)
		fmt.Fprintf(tw, "subnet4:\t%s\n", resp.Net.Subnet4)
		fmt.Fprintf(tw, "gateways:\t%s\n", strings.Join(resp.Net.Gateways, " "))
		fmt.Fprintf(tw, "nameservers:\t%s\n", strings.Join(resp.Net.Nameservers, " "))
		fmt.Fprintf(tw, "dhcp4server:\t%s\n", resp.Net.Dhcp4Server)
		fmt.Fprintf(tw, "domain:\t%s\n", resp.Net.Domain)
		if resp.Net.Range4 != nil {
			fmt.Fprintf(tw, "range4:\t%s-%s\n", resp.Net.Range4.Begin, resp.Net.Range4.End)
		}
		tw.Flush()

		return nil

	})
}

func setNetwork(file string) {

	net, err := loadSpec(file)
	if err != nil {
		log.Fatal(err)
	}

	withClient(func(cli nex.NexClient) error {

		_, err = cli.AddNetwork(ctx, &nex.AddNetworkRequest{Network: net})
		if err != nil {
			grpcFatal(err)
		}

		return nil

	})

}
func deleteNetwork(name string) {
	withClient(func(cli nex.NexClient) error {

		_, err := cli.DeleteNetwork(ctx, &nex.DeleteNetworkRequest{Name: name})
		if err != nil {
			grpcFatal(err)
		}

		return nil

	})
}

func loadSpec(file string) (*nex.Network, error) {

	data, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}

	net := &nex.Network{}
	err = yaml.Unmarshal(data, net)
	if err != nil {
		return nil, err
	}

	return net, nil
}

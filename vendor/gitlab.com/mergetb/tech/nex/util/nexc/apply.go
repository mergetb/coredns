package main

import (
	//"context"
	"log"

	//log "github.com/mergetb/logrus"
	"github.com/spf13/cobra"
	"gitlab.com/mergetb/tech/nex/pkg"
	//"gopkg.in/yaml.v2"
)

var applyHelp string = `
Apply an object specification. The specification is a file that may
contain multiple YAML documents. Each document may include exacly 1

  - Network
  - MemberList

There may be an arbirary number of documents in the file.

For example:

  kind:         Network
  name:         mini
  subnet4:      10.0.0.0/24
  gateways:     [10.0.0.1, 10.0.0.2]
  nameservers:  [10.0.0.1]
  dhcp4server:  10.0.0.1
  domain:       mini.net
  range4:
    begin: 10.0.0.0
    end:   10.0.0.254

  ---
  kind:   MemberList
  net:    mini
  list:
  - mac:  00:00:11:11:00:01
    name: whiskey

  - mac:  00:00:22:22:00:01
    name: tango

  - mac:  00:00:33:33:00:01
    name: foxtrot
`

func applyCmd(root *cobra.Command) {

	apply := &cobra.Command{
		Use:   "apply [yaml spec]",
		Short: "Apply an object specification",
		Long:  applyHelp,
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			doApply(args[0])
		},
	}
	root.AddCommand(apply)

}

func doApply(file string) {

	metaObjects, err := nex.ReadSpec(file)
	if err != nil {
		log.Fatal(err)
	}

	for _, mo := range metaObjects {

		switch obj := mo.Object.(type) {

		case *nex.Network:
			applyNetwork(obj)

		case *nex.MemberList:
			applyMemberList(obj)

		}

	}

}

func applyNetwork(net *nex.Network) {
	withClient(func(cli nex.NexClient) error {

		_, err := cli.AddNetwork(ctx, &nex.AddNetworkRequest{
			Network: net,
		})
		if err != nil {
			grpcFatal(err)
		}

		return nil

	})

}

func applyMemberList(list *nex.MemberList) {
	withClient(func(cli nex.NexClient) error {

		_, err := cli.AddMembers(ctx, list)
		if err != nil {
			grpcFatal(err)
		}

		return nil

	})
}

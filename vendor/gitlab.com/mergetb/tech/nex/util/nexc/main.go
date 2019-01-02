package main

import (
	"context"
	"log"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"gitlab.com/mergetb/tech/nex/pkg"
)

var server string = "localhost"
var ctx = context.TODO()
var tw *tabwriter.Writer = tabwriter.NewWriter(os.Stdout, 0, 0, 4, ' ', 0)

func main() {
	log.SetFlags(0)

	root := &cobra.Command{
		Use:   "nex",
		Short: "Nex dhcp/dns client",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
		},
	}
	root.PersistentFlags().StringVarP(
		&server, "server", "s", "localhost", "nexd server to connect to")

	rootCmds(root)

	root.Execute()
}

func rootCmds(root *cobra.Command) {

	version := &cobra.Command{
		Use:   "version",
		Short: "Show version",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			log.Print(nex.Version)
		},
	}
	root.AddCommand(version)

	get := &cobra.Command{
		Use:   "get",
		Short: "Get something",
	}
	root.AddCommand(get)

	set := &cobra.Command{
		Use:   "set",
		Short: "Set something",
	}
	root.AddCommand(set)

	add := &cobra.Command{
		Use:   "add",
		Short: "Add something",
	}
	root.AddCommand(add)

	delete := &cobra.Command{
		Use:   "delete",
		Short: "Delete something",
	}
	root.AddCommand(delete)

	applyCmd(root)
	memberCmds(get, set, add, delete)
	networkCmds(get, set, add, delete)

}

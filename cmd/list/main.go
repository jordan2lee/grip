package list

import (
	"fmt"

	"github.com/bmeg/grip/gripql"
	"github.com/bmeg/grip/util/rpc"
	"github.com/golang/protobuf/jsonpb"
	"github.com/spf13/cobra"
)

var host = "localhost:8202"

// Cmd is the declaration of the command line
var Cmd = &cobra.Command{
	Use:   "list",
	Short: "List operations",
}

var listGraphsCmd = &cobra.Command{
	Use:   "graphs",
	Short: "List graphs",
	Long:  ``,
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		conn, err := gripql.Connect(rpc.ConfigWithDefaults(host), true)
		if err != nil {
			return err
		}

		resp, err := conn.ListGraphs()
		if err != nil {
			return err
		}

		for _, g := range resp.Graphs {
			fmt.Printf("%s\n", g)
		}
		return nil
	},
}

var listLabelsCmd = &cobra.Command{
	Use:   "labels <graph>}",
	Short: "List the vertex and edge labels in a graph",
	Long:  ``,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		graph := args[0]

		conn, err := gripql.Connect(rpc.ConfigWithDefaults(host), true)
		if err != nil {
			return err
		}

		resp, err := conn.ListLabels(graph)
		if err != nil {
			return err
		}
		m := jsonpb.Marshaler{
			EnumsAsInts:  false,
			EmitDefaults: false,
			Indent:       "  ",
			OrigName:     false,
		}
		txt, err := m.MarshalToString(resp)
		if err != nil {
			return fmt.Errorf("failed to marshal ListLabels response: %v", err)
		}
		fmt.Printf("%s\n", txt)
		return nil
	},
}

func init() {
	gflags := listGraphsCmd.Flags()
	gflags.StringVar(&host, "host", host, "grip server url")

	lflags := listLabelsCmd.Flags()
	lflags.StringVar(&host, "host", host, "grip server url")

	Cmd.AddCommand(listGraphsCmd)
	Cmd.AddCommand(listLabelsCmd)
}

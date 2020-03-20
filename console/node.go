package main

import (
	"errors"
	"fmt"
	"github.com/glvd/accipfs/config"
	"github.com/glvd/accipfs/core"
	"github.com/glvd/accipfs/general"
	"github.com/glvd/accipfs/service"
	"github.com/spf13/cobra"
)

func nodeCmd() *cobra.Command {
	nodeCmd := &cobra.Command{
		Use:   "node",
		Short: "node run",
		Long:  "node can operate to change the parameters of some nodes",
	}
	nodeCmd.AddCommand(nodeConnectCmd())
	return nodeCmd
}

func nodeConnectCmd() *cobra.Command {
	var addr string
	connect := &cobra.Command{
		Use:   "connect",
		Short: "connect run",
		Long:  "connect a remote node",
		Run: func(cmd *cobra.Command, args []string) {
			config.Initialize()
			cfg := config.Global()
			url := fmt.Sprintf("http://localhost:%d/rpc", cfg.Port)
			reply := new(core.NodeInfo)
			if err := general.RPCPost(url, "Accelerate.ID", &service.Empty{}, reply); err != nil {
				panic(err)
			}
			remoteURL := fmt.Sprintf("http://%s/rpc", addr)
			status := new(bool)
			if err := general.RPCPost(remoteURL, "Accelerate.Connect", reply, status); err != nil {
				panic(err)
			}

			if !(*status) {
				panic(errors.New("failed connect to remote"))
			}
			return
		},
	}
	connect.Flags().StringVar(&addr, "addr", "localhost", "set a remote address to connect")
	return connect
}

package main

import (
	"github.com/spf13/cobra"
)

func addCmd() *cobra.Command {
	var path string
	var info string
	cmd := &cobra.Command{
		Use:   "add",
		Short: "add a source to this node",
		Long:  "add a source to this node for service some users by your self",
		Run: func(cmd *cobra.Command, args []string) {
			//add a file with rule to accipfs

		},
	}
	cmd.Flags().StringVar(&path, "path", "", "set the file dirctory path to add")
	cmd.Flags().StringVar(&info, "info", "", "set the file info to load")
	return cmd
}

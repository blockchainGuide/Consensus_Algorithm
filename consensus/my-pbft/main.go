package main

import (
	"github.com/spf13/cobra"
	"my-pbft/client"
	"my-pbft/p2p"
)

func main()  {
	var rootCmd = &cobra.Command{
		Use: "my-pbft",
	}

	rootCmd.AddCommand(client.ClientCmd())
	rootCmd.AddCommand(p2p.ServerCmd())
	rootCmd.Execute()
}

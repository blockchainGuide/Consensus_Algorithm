package main

import (
	"github.com/spf13/cobra"
	"my-dpos/blockchain"
	"my-dpos/consensus"
	"my-dpos/p2p"
)

func main()  {
	var rootCmd = &cobra.Command{
		Use: "my-dpos",
	}

	rootCmd.AddCommand(blockchain.BlockchainCmd())
	rootCmd.AddCommand(p2p.P2PCmd())
	rootCmd.AddCommand(consensus.DPoSCmd())

	rootCmd.Execute()

}

package main

import (
	"github.com/spf13/cobra"
	"my-pos/blockchain"
	"my-pos/consensus"
)

func main()  {

	var rootCmd = cobra.Command{
		Use: "my-pos",
	}

	rootCmd.AddCommand(blockchain.BlockchianCmd())
	rootCmd.AddCommand(consensus.ConsensusCmd())
	rootCmd.Execute()
}



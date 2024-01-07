package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

var rootCmd = &cobra.Command{
	Use:   "pareto",
	Short: "pareto command line interface",
	Long:  `pareto is a CLI tool for managing pareto applications`,
	Run: func(cmd *cobra.Command, args []string) {
		//
	},
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(-1)
	}
}

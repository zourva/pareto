package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

var appCmd = &cobra.Command{
	Use:   "app",
	Short: "app management",
	Long:  `app management commands`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 2 {
			fmt.Println("insufficient args")
			return
		}
		//subCmd := args[0]
	},
}

func appCmdCreate() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "create",
		Short: "create an app",
		Long:  `create an app with the given name`,
		Run: func(cmd *cobra.Command, args []string) {
			if cmd.Flag("name") == nil {
				fmt.Println("name is necessary")
				return
			}

			name := cmd.Flag("name").Value.String()

			//make app dir
			err := os.MkdirAll(name, 0755)
			if err != nil {
				fmt.Println(err)
				return
			}

			////create working dirs
			//wd := env.NewWorkingDir(true,
			//	[]*env.DirInfo{
			//		{Name: "bin", Mode: 0755},
			//		{Name: "data", Mode: 0755},
			//		{Name: "etc", Mode: 0755},
			//		{Name: "log", Mode: 0755},
			//		{Name: "tool", Mode: 0755},
			//	})
			//
			////generate config file
			//config.Create("etc/config.json")

			//generate framework code

			fmt.Println(name)
		},
	}

	cmd.Flags().StringP("name", "n", "", "name of the app")
	_ = cmd.MarkFlagRequired("name")

	return cmd
}

func init() {
	appCmd.AddCommand(appCmdCreate())
	rootCmd.AddCommand(appCmd)
}

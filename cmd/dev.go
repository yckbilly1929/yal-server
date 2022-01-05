package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/yckbilly1929/yalive-server/app/server"
)

var (
	devConfig string
)

// devCmd represents the dev command
var devCmd = &cobra.Command{
	Use:   "dev",
	Short: "Start dev server",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		var config server.ServeConfig
		err := json.Unmarshal([]byte(devConfig), &config)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			os.Exit(1)
		}

		server.Run(config)
	},
}

func init() {
	rootCmd.AddCommand(devCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// devCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// devCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	devCmd.Flags().StringVarP(&devConfig, "config", "c", "", "json configuration")
}

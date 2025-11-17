package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Display version information",
	Long: `Display detailed version information including version number,
git commit hash, build time, Go version, and target platform.`,
	Run: func(cmd *cobra.Command, args []string) {
		v := GetVersion()
		fmt.Println(v.String())
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}

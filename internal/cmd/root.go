/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"log"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "mrtdump <input-file>",
	Short: "mrtdump - A tool to export MRT binary files to human-readable format",
	Args:  cobra.ExactArgs(1), // Ensure that exactly one argument is provided
	Run: func(cmd *cobra.Command, args []string) {
		verboseFlag, _ := cmd.Flags().GetBool("verbose")
		jsonFlag, _ := cmd.Flags().GetBool("json")
		if len(args) == 1 {
			configFS := os.DirFS(filepath.Dir(args[0])) // Use the directory of the provided file as the filesystem
			// test if the file exists
			rf := NewReadFileOptions(configFS, filepath.Base(args[0]), verboseFlag, jsonFlag)
			rf.ReadFile() // Call ReadFile with the provided filepath
		} else {
			cmd.Help() // Display help if no subcommand is provided
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		log.Fatal(err)
	} else {
		os.Exit(0) // Exit with success if no error occurred
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.mrtdump.yaml)")

	// verboseFlag, _ := rootCmd.Flags().GetBool("verbose")
	rootCmd.Flags().BoolP("verbose", "v", false, "enable verbose output printing the PeerIndex and RIB entries")
	rootCmd.Flags().BoolP("json", "j", false, "enable JSON output format")
}

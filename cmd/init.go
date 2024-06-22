/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"log"
	"os"

	"github.com/kodova/got/repo"
	"github.com/spf13/cobra"
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "create a empty got repository",
	Run: func(cmd *cobra.Command, args []string) {
		pwd, err := os.Getwd()
		if err != nil {
			log.Fatal("could not get current working directory")
		}
		err = repo.Init(pwd)
		if err != nil {
			log.Fatalf("failed to create got repository, %v", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}

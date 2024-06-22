package cmd

import (
	"fmt"

	"github.com/kodova/got/repo"
	"github.com/spf13/cobra"
)

var writeTreeCmd = &cobra.Command{
	Use: "write-tree",
	RunE: func(cmd *cobra.Command, args []string) error {
		rep, err := repo.NewRepository(".")
		if err != nil {
			return err
		}

		obj, err := rep.WriteTree(".")
		if err != nil {
			return err
		}

		fmt.Println("root OID", obj.Hash)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(writeTreeCmd)
}

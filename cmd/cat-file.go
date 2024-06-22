package cmd

import (
	"fmt"

	"github.com/kodova/got/repo"
	"github.com/spf13/cobra"
)

var catFileCmd = &cobra.Command{
	Use:   "cat-file [type] [object ID]",
	Short: "provides content of reposiotry objects",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {

		oType := repo.ObjectType(args[0])
		hash := args[1]

		r, err := repo.NewRepository(".")
		if err != nil {
			return err
		}

		obj, err := r.ReadObject(hash)
		if err != nil {
			return err
		}

		if oType != obj.Type {
			return fmt.Errorf("object %v is not of type %v", args[1], oType)
		}

		fmt.Print(string(obj.Data))

		return nil
	},
}

func init() {
	rootCmd.AddCommand(catFileCmd)
}

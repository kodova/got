package cmd

import (
	"fmt"
	"os"

	"github.com/kodova/got/repo"
	"github.com/spf13/cobra"
)

var hashObjectFlags = struct {
	Type       string
	Stdin      bool
	StdinPaths string
	Path       string
	Write      bool
}{}

var hashObjectCmd = &cobra.Command{
	Use:   "hash-object",
	Short: "Compute object ID and optionally create and object from a file",
	Long: `Computes the object ID value for a object with specified type witha the content of 
the named file (which cand be out of the work treed), and optionally writes the
resulting object to the object database . Reports its object ID to standard 
output. When type is notsified, it defaults to blob`,
	RunE: func(cmd *cobra.Command, args []string) error {
		rep, err := repo.NewRepository(".")
		if err != nil && hashObjectFlags.Write {
			return err
		}

		for _, f := range args {
			file, err := os.Open(f)
			if err != nil {
				return fmt.Errorf("could not open %v for reading: %v\n", f, err)
			}
			obj, err := repo.NewObject(repo.ObjTypBlob, file)
			if err != nil {
				return fmt.Errorf("could not has object: %v\n", err)
			}

			fmt.Println(obj.Hash)
			if hashObjectFlags.Write {
				err = rep.WriteObject(obj)
				if err != nil {
					return fmt.Errorf("failed to write object: %w", err)
				}
			}
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(hashObjectCmd)
	hashObjectCmd.Flags().StringVarP(&hashObjectFlags.Type, "type", "t", "blob", "Specify the type of object to be created, possible valuea are commit, tree, blob, and tag")
	hashObjectCmd.Flags().BoolVarP(&hashObjectFlags.Write, "write", "w", false, "Write the objec to the object DB")
}

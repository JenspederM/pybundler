package cmd

import (
	"fmt"
	"os"

	"github.com/cloudflare/cfssl/log"
	"github.com/jenspederm/pybundler/internal/bundle"
	"github.com/spf13/cobra"
)

func BundleCmd() *cobra.Command {
	cmd := &cobra.Command{}

	cmd.Use = "bundle"
	cmd.Short = "Bundle a Python project"
	cmd.Long = `Bundle a Python project into a single executable file.`

	cmd.Flags().StringP("path", "p", ".", "Path to the Python project")
	cmd.Flags().StringP("output", "o", "dist", "Output directory for the bundle")
	cmd.Flags().BoolP("overwrite", "w", false, "Overwrite existing files")

	cmd.Run = func(cmd *cobra.Command, args []string) {
		// Implementation here
		path := cmd.Flag("path").Value.String()
		output := cmd.Flag("output").Value.String()

		bo, err := bundle.NewBundleOptions(path, output)
		cobra.CheckErr(err)
		log.Info("Bundle options: ", bo)

		if _, err := os.Stat(bo.Output); err == nil {
			if cmd.Flag("overwrite").Value.String() == "false" {
				fmt.Printf("File %s already exists. Use --overwrite to overwrite.\n", fmt.Sprintf("%s/%s", bo.Output, "main.go"))
				return
			}
			err := os.RemoveAll(bo.Output)
			cobra.CheckErr(err)
			err = os.MkdirAll(bo.Output, os.ModePerm)
			cobra.CheckErr(err)
		}

		err = bundle.Run(bo)
		if err != nil {
			cobra.CheckErr(err)
		}

	}

	return cmd
}

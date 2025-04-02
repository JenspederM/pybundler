package cmd

import (
	"github.com/jenspederm/pybundler/internal/bundle"
	"github.com/spf13/cobra"
)

func BundleCmd() *cobra.Command {
	cmd := &cobra.Command{}

	cmd.Use = "bundle"
	cmd.Short = "Bundle a Python project"
	cmd.Long = `Bundle a Python project into a single executable file.`

	cmd.Flags().StringP("path", "p", ".", "Path to the Python project")
	cmd.Flags().StringP("output", "o", "", "Output directory for the bundle")
	cmd.Flags().BoolP("overwrite", "w", false, "Overwrite existing files")
	cmd.Flags().BoolP("verbose", "v", false, "Enable verbose output")

	cmd.Run = func(cmd *cobra.Command, args []string) {
		// Implementation here
		path := cmd.Flag("path").Value.String()
		output := cmd.Flag("output").Value.String()
		overwrite := cmd.Flag("overwrite").Value.String()
		verbose := cmd.Flag("verbose").Value.String()

		b, err := bundle.New(path, output, overwrite == "true")
		cobra.CheckErr(err)
		err = b.Run(verbose == "true")
		cobra.CheckErr(err)
	}

	return cmd
}

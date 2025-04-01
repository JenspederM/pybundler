package cmd

import (
	"io"
	"os"
	"path/filepath"

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
	cmd.Flags().StringP("output", "o", "", "Output directory for the bundle")
	cmd.Flags().BoolP("overwrite", "w", false, "Overwrite existing files")
	cmd.Flags().BoolP("verbose", "v", false, "Enable verbose output")

	cmd.Run = func(cmd *cobra.Command, args []string) {
		// Implementation here
		path := cmd.Flag("path").Value.String()
		output := cmd.Flag("output").Value.String()
		overwrite := cmd.Flag("overwrite").Value.String()
		verbose := cmd.Flag("verbose").Value.String()

		bo, err := bundle.New(path, output)
		cobra.CheckErr(err)
		log.Infof("Creating bundle for %s at %s", bo.Path, bo.Output)
		if _, err := os.Stat(bo.Output); err == nil {
			isEmpty, err := IsEmpty(bo.Output)
			cobra.CheckErr(err)
			if !isEmpty && overwrite == "false" {
				fp := filepath.Join(bo.Output, "main.go")
				log.Fatalf("File %s already exists. Use --overwrite to overwrite.\n", fp)
				return
			}
			err = os.RemoveAll(bo.Output)
			cobra.CheckErr(err)
			err = os.MkdirAll(bo.Output, os.ModePerm)
			cobra.CheckErr(err)
		}
		err = bundle.Run(bo, verbose == "true")
		if err != nil {
			cobra.CheckErr(err)
		}

	}

	return cmd
}

func IsEmpty(name string) (bool, error) {
	f, err := os.Open(name)
	if err != nil {
		return false, err
	}
	defer f.Close()

	_, err = f.Readdirnames(1) // Or f.Readdir(1)
	if err == io.EOF {
		return true, nil
	}
	return false, err // Either not empty or error, suits both cases
}

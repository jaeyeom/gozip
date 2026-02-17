// Command gounzip extracts zip archives.
package main

import (
	"fmt"
	"os"

	"github.com/jaeyeom/gozip/ziplib"
	"github.com/spf13/cobra"
)

func main() {
	var (
		list      bool
		overwrite bool
		outputDir string
		junkPaths bool
	)

	rootCmd := &cobra.Command{
		Use:   "gounzip [flags] zipfile [file ...]",
		Short: "Extract zip archives",
		Long:  "gounzip extracts zip archives, compatible with standard unzip.",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			zipPath := args[0]
			filePatterns := args[1:]

			if list {
				return listArchive(zipPath)
			}

			opts := ziplib.UnzipOptions{
				OutputDir:    outputDir,
				Overwrite:    overwrite,
				JunkPaths:    junkPaths,
				FilePatterns: filePatterns,
				Output:       os.Stdout,
			}

			return ziplib.Unzip(zipPath, opts)
		},
		SilenceUsage: true,
	}

	rootCmd.Flags().BoolVarP(&list, "list", "l", false, "List archive contents")
	rootCmd.Flags().BoolVarP(&overwrite, "overwrite", "o", false, "Overwrite existing files")
	rootCmd.Flags().StringVarP(&outputDir, "directory", "d", ".", "Extract files into directory")
	rootCmd.Flags().BoolVarP(&junkPaths, "junk-paths", "j", false, "Junk (ignore) directory paths")

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func listArchive(zipPath string) error {
	entries, err := ziplib.List(zipPath)
	if err != nil {
		return err
	}

	fmt.Printf("  Length      Date    Time    Name\n")
	fmt.Printf("---------  ---------- -----   ----\n")

	var totalSize uint64
	for _, e := range entries {
		mod := e.Modified
		fmt.Printf("%9d  %04d-%02d-%02d %02d:%02d   %s\n",
			e.UncompressedSize,
			mod.Year(), mod.Month(), mod.Day(),
			mod.Hour(), mod.Minute(),
			e.Name,
		)
		totalSize += e.UncompressedSize
	}

	fmt.Printf("---------                     -------\n")
	fmt.Printf("%9d                     %d files\n", totalSize, len(entries))

	return nil
}

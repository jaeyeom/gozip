// Command gozip creates zip archives.
package main

import (
	"fmt"
	"os"

	"github.com/jaeyeom/gozip/ziplib"
	"github.com/spf13/cobra"
)

func main() {
	var (
		recursive       bool
		excludePatterns []string
		levels          [10]bool // -0 through -9
	)

	rootCmd := &cobra.Command{
		Use:   "gozip [flags] zipfile file1 [file2 ...]",
		Short: "Create zip archives",
		Long:  "gozip creates zip archives, compatible with standard zip.",
		Args:  cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			zipPath := args[0]
			files := args[1:]

			level := -1 // default
			for i, set := range levels {
				if set {
					level = i
				}
			}

			opts := ziplib.ZipOptions{
				Recursive:        recursive,
				CompressionLevel: level,
				ExcludePatterns:  excludePatterns,
				Output:           os.Stdout,
			}

			return ziplib.Zip(zipPath, files, opts)
		},
		SilenceUsage: true,
	}

	rootCmd.Flags().BoolVarP(&recursive, "recurse-paths", "r", false, "Travel the directory structure recursively")
	rootCmd.Flags().StringArrayVarP(&excludePatterns, "exclude", "x", nil, "Exclude files matching pattern")

	for i := 0; i <= 9; i++ {
		rootCmd.Flags().BoolVarP(&levels[i], fmt.Sprintf("%d", i), fmt.Sprintf("%d", i), false, fmt.Sprintf("Compression level %d", i))
	}

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

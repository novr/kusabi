package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/novr/kusabi/internal/version"
)

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(version.Version)
		},
	}
}

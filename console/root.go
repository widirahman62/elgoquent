package console

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:     appname,
	Short:   fmt.Sprintf("%s\nYour %s on the Go.\n",asciiArt,appname),
	Version: fmt.Sprintf(":\nelgoquent %s\n", version),
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	completion := &cobra.Command{
		Use:   "completion",
		Short: "Generate the autocompletion script for the specified shell",
	}
	completion.Hidden = true
	rootCmd.AddCommand(completion)
}



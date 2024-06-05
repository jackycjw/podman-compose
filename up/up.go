package up

import (
	"fmt"
	"github.com/spf13/cobra"
	"podman-compose/registry"
)

var upCmd = &cobra.Command{
	Use:   "up",
	Short: "Create and start containers",
	Run:   up,
}

var daemon bool = false

func init() {
	upCmd.Flags().BoolVarP(&daemon, "daemon", "d", false, "daemon mode")
	registry.Commands = append(registry.Commands, upCmd)
}

func up(cmd *cobra.Command, args []string) {
	fmt.Println("up", daemon)
}

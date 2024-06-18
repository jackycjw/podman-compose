package ps

import (
	"fmt"
	"github.com/spf13/cobra"
	"podman-compose/cli"
	"podman-compose/compose"
	"podman-compose/registry"
	"podman-compose/util"
	"strings"
)

var psCmd = &cobra.Command{
	Use:   "ps",
	Short: "List containers.",
	Run:   ps,
}

var detach = false

// 删除孤立项
var all = false

func init() {
	psCmd.Flags().BoolVarP(&all, "all", "a", false, "Show all stopped containers (including those created by the run command)")
	registry.Commands = append(registry.Commands, psCmd)
}

func ps(cmd *cobra.Command, args []string) {
	compose.InitContainerList()
	fmt.Println("   Name                   Command                State                         Ports                     ")
	fmt.Println("---------------------------------------------------------------------------------------------------------")
	for _, container := range compose.ContainerList {
		if container.State == "running" {
			fmt.Println(toString(container))
		} else if all {
			fmt.Println(toString(container))
		}
	}
}

func toString(container cli.ListContainer) string {
	return util.FixSizeString(container.Names[0], 16, false) +
		util.FixSizeString(" "+strings.Join(container.Command, " "), 30, true) +
		util.FixSizeString(" "+container.State, 9, false) +
		util.FixSizeString(" "+formatPortString(container.Ports), 46, true)
}

func formatPortString(ports []cli.PortMapping) string {
	portsArr := make([]string, 0)
	for _, port := range ports {
		portsArr = append(portsArr, port.String())
	}
	return strings.Join(portsArr, ",")
}

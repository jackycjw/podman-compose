package ps

import (
	"fmt"
	"github.com/spf13/cobra"
	"podman-compose/cli"
	"podman-compose/compose"
	"podman-compose/registry"
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
	return formatString(container.Names[0], 16, false) +
		formatString(" "+strings.Join(container.Command, " "), 30, true) +
		formatString(" "+container.State, 9, false) +
		formatString(" "+formatPortString(container.Ports), 46, true)
}

func formatString(str string, length int, middle bool) string {
	if len(str) == length {
		return str
	} else if len(str) > length {
		return str[0:length]
	} else {
		if middle {
			left := length - len(str)
			before := left / 2
			after := left - before
			return strings.Repeat(" ", before) + str + strings.Repeat(" ", after)
		} else {
			return str + strings.Repeat(" ", length-len(str))
		}

	}
}
func formatPortString(ports []cli.PortMapping) string {
	portsArr := make([]string, 0)
	for _, port := range ports {
		portsArr = append(portsArr, port.String())
	}
	return strings.Join(portsArr, ",")
}

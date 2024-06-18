package down

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"podman-compose/cli"
	"podman-compose/compose"
	"podman-compose/constant"
	"podman-compose/registry"
)

var downCmd = &cobra.Command{
	Use:   "down",
	Short: "Stops containers and removes containers, networks, volumes, and images created by `up`",
	Run:   down,
}

// 删除孤立项
var removeOrphans = false

func init() {
	downCmd.Flags().BoolVarP(&removeOrphans, "remove-orphans", "", false, "Remove containers for services not defined in the Compose file")
	registry.Commands = append(registry.Commands, downCmd)
}

func down(cmd *cobra.Command, args []string) {
	var err error

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	services := map[string]string{}
	for _, arg := range args {
		services[arg] = "1"
	}

	dockerCompose := compose.GetDockerCompose()
	//指定服务停止
	if len(services) == 0 {
		for serviceName := range dockerCompose.Services {
			serviceDown(serviceName)
		}
	} else {
		//全部停止
		for service, _ := range services {
			_, exist := dockerCompose.Services[service]
			if !exist {
				fmt.Printf("Service %s does not exist\n", service)
			} else {
				serviceDown(service)
			}
		}
	}

	RemoveOrphans(removeOrphans)

}

// 删除孤立项
func RemoveOrphans(removeOrphans bool) {
	dockerCompose := compose.GetDockerCompose()
	if removeOrphans {
		for _, container := range compose.ContainerList {
			expectServiceName := container.Labels[constant.LabelComposeServiceName]
			_, exist := dockerCompose.Services[expectServiceName]
			if !exist {
				fmt.Println("orphans {" + expectServiceName + "} removing...")
				force := true
				cli.Remove(container.ID, &force, nil)
			}
		}
	} else {
		existOrphans := false
		for _, container := range compose.ContainerList {
			expectServiceName := container.Labels[constant.LabelComposeServiceName]
			_, exist := dockerCompose.Services[expectServiceName]
			if !exist {
				existOrphans = true
				break
			}
		}
		if existOrphans {
			fmt.Println("exist orphans, you can clean orphan containers with `--remove-orphans`")
		}
	}
}

func serviceDown(serviceName string) {
	container, exist := compose.GetContainer(serviceName)

	if exist {
		force := true
		fmt.Println(serviceName + " removing...")
		cli.Remove(container.ID, &force, nil)
	}
}

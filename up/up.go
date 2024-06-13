package up

import (
	"context"
	"errors"
	"fmt"
	"github.com/containers/libpod/pkg/bindings"
	"github.com/spf13/cobra"
	"os"
	"podman-compose/cli"
	"podman-compose/compose"
	"podman-compose/constant"
	"podman-compose/registry"
	"strconv"
	"strings"
	"sync"
)

var upCmd = &cobra.Command{
	Use:   "up",
	Short: "Create and start containers",
	Run:   up,
}

var detach bool = false

var connection context.Context

func init() {
	upCmd.Flags().BoolVarP(&detach, "detach", "d", false, "daemon mode")
	registry.Commands = append(registry.Commands, upCmd)
}

func up(cmd *cobra.Command, args []string) {
	var err error
	connection, err = bindings.NewConnection(context.Background(), "unix:///run/podman/podman.sock")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	dockerCompose := compose.GetDockerCompose()
	for serviceName := range dockerCompose.Services {
		serviceUp(serviceName, dockerCompose.Services[serviceName])
		//fmt.Println("服务名称: ", serviceName)
		//command, err := getCommand(serviceName, dockerCompose.Services[serviceName])
		//if err != nil {
		//	fmt.Println(err)
		//	os.Exit(1)
		//	return
		//}
		//podmanCmd, err := exec.LookPath("podman")
		//cmd := exec.Command(podmanCmd, command...)
		//
		//cmd.Stdout = os.Stdout
		//cmd.Stderr = os.Stdout
		//err = cmd.Run()
		//if err != nil {
		//	fmt.Println(err)
		//}
	}
}

func serviceUp(serviceName string, service compose.ServiceConfig) {
	container, exist := getContainer(serviceName)
	fmt.Println(container.Image, container)
	upToDate := false
	if exist {

	}

	if exist {

	} else {
		upToDate = false
	}
	if !upToDate {

	} else {
		fmt.Println(serviceName, "is up to date")
	}
}

// 是否是最新
func isUpToDate(listContainer cli.ListContainer, service compose.ServiceConfig) bool {

	key1V := listContainer.Labels[constant.LabelConfigKey1]
	key2V := listContainer.Labels[constant.LabelConfigKey2]

	expectKey1V, expectKey2V := service.GetUnique()
	if key1V == expectKey1V || key2V == expectKey2V {
		return true
	} else {
		return false
	}
}

var containerList []cli.ListContainer
var lock sync.Mutex

func getContainer(serviceName string) (cli.ListContainer, bool) {
	initContainerList()

	for _, container := range containerList {
		v, ok := container.Labels[constant.LabelComposeServiceName]
		if ok && v == serviceName {
			return container, true
		}
	}
	return cli.ListContainer{}, false
}

func initContainerList() {
	workDir, _ := os.Getwd()

	if containerList == nil {
		lock.Lock()
		if containerList == nil {
			containerListTmp := make([]cli.ListContainer, 0)
			cs, err := cli.List(connection, nil, nil, nil, nil, nil, nil)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			for _, c := range cs {
				fmt.Println(c.Image, c)
				v, ok := c.Labels[constant.LabelComposeDir]
				if ok {
					if v == workDir {
						containerListTmp = append(containerListTmp, c)
					}
				}
			}
			containerList = containerListTmp
		}
		lock.Unlock()
	}
}

/*
*
获取命令
*/
func getCommand(name string, service compose.ServiceConfig) ([]string, error) {
	command := []string{"run"}
	if detach {
		command = append(command, "--detach")
	}
	//端口
	command, err := formatPort(command, &service)
	if err != nil {
		return command, err
	}

	//挂载卷
	command, err = formatVolumes(command, &service)
	if err != nil {
		return command, err
	}

	//容器名称
	fmt.Println("service.ContainerName ", service.ContainerName)
	if strings.TrimSpace(service.ContainerName) != "" {
		command = append(command, "--name", service.ContainerName)
	}

	//workdir
	if strings.TrimSpace(service.WorkingDir) != "" {
		command = append(command, "--workdir", service.WorkingDir)
	}

	//Entrypoint
	if strings.TrimSpace(service.Entrypoint) != "" {
		command = append(command, "--entrypoint", service.Entrypoint)
	}

	// restart
	if strings.TrimSpace(service.Restart) != "" {
		command = append(command, "--restart", service.Restart)
	}

	//镜像
	command, err = formatImage(command, &service)
	if err != nil {
		return command, err
	}

	//容器命令
	command, err = formatContainerCommand(command, &service)
	if err != nil {
		return command, err
	}
	return command, nil
}

// 镜像
func formatContainerCommand(command []string, service *compose.ServiceConfig) ([]string, error) {
	if len(service.Command) > 0 {
		for _, cmd := range service.Command {
			command = append(command, cmd)
		}
	}
	return command, nil
}

// 镜像
func formatImage(command []string, service *compose.ServiceConfig) ([]string, error) {
	image := strings.TrimSpace(service.Image)
	if image == "" {
		return command, errors.New("image is required")
	}
	command = append(command, service.Image)
	return command, nil
}

// 挂载卷
func formatVolumes(command []string, service *compose.ServiceConfig) ([]string, error) {
	for _, volume := range service.Volumes {
		command = append(command, "-v", volume)
	}
	return command, nil
}

// 组装端口
func formatPort(command []string, service *compose.ServiceConfig) ([]string, error) {
	for _, port := range service.Ports {
		pair := strings.Split(port, ":")
		if len(pair) == 1 {
			_, err := strconv.Atoi(pair[0])
			if err != nil {
				return nil, errors.New("port [" + port + "] is Invalid")
			}
		}
		if len(pair) == 2 {
			_, err := strconv.Atoi(pair[0])
			if err != nil {
				return nil, errors.New("port [" + port + "] is Invalid")
			}
			_, err = strconv.Atoi(pair[1])
			if err != nil {
				return nil, errors.New("port [" + port + "] is Invalid")
			}
		}
		command = append(command, "-p", port)
	}

	return command, nil
}

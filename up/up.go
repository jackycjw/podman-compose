package up

import (
	"context"
	"errors"
	"fmt"
	"github.com/containers/libpod/pkg/bindings"
	"github.com/containers/libpod/pkg/bindings/containers"
	"github.com/spf13/cobra"
	"os"
	"os/exec"
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

	services := map[string]string{}
	for _, arg := range args {
		services[arg] = "1"
	}
	if len(services) == 0 {
		dockerCompose := compose.GetDockerCompose()
		for serviceName := range dockerCompose.Services {
			serviceUp(serviceName, dockerCompose.Services[serviceName])
		}
	} else {
		for service, _ := range services {
			dockerCompose := compose.GetDockerCompose()
			serviceConfig, exist := dockerCompose.Services[service]
			if !exist {
				fmt.Printf("Service %s does not exist\n", service)
			} else {
				serviceUp(service, serviceConfig)
			}
		}

	}

}

func serviceUp(serviceName string, service compose.ServiceConfig) {
	container, exist := getContainer(serviceName)
	upToDate := exist && isUpToDate(container, service)

	if upToDate {
		fmt.Println(serviceName, "is up to date")
	} else {
		if exist {
			force := true
			fmt.Println(serviceName + " recreating...")
			containers.Remove(connection, container.ID, &force, nil)
		}

		command, err := getCommand(serviceName, service)
		if err != nil {
			fmt.Println(serviceName, ":", err)
			os.Exit(1)
		}
		podmanCmd, err := exec.LookPath("podman")
		cmd := exec.Command(podmanCmd, command...)

		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stdout
		err = cmd.Run()
		if err != nil {
			fmt.Println(serviceName, ":", err)
			os.Exit(1)
		}
		fmt.Println(serviceName + " created")
	}
}

// 是否是最新
func isUpToDate(listContainer cli.ListContainer, service compose.ServiceConfig) bool {

	// 运行中 && 配置未修改 && 镜像也未修改  则是 up to date
	if listContainer.State == "running" {
		key := listContainer.Labels[constant.LabelConfigKey]
		expectKey := service.GetUnique()
		if key == expectKey {
			detail, err := cli.Inspect(connection, listContainer.ID, nil)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			image, err := cli.GetImage(connection, service.Image, nil)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			if detail.Image == image.ID {
				return true
			}
		}
	}

	return false
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

/*
*
初始化容器列表
*/
func initContainerList() {
	workDir, _ := os.Getwd()

	if containerList == nil {
		lock.Lock()
		if containerList == nil {
			containerListTmp := make([]cli.ListContainer, 0)
			all := true
			cs, err := cli.List(connection, nil, &all, nil, nil, nil, nil)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			for _, c := range cs {
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

	//标签
	command = append(command, "--label", constant.LabelComposeDir+"="+compose.GetComposeDir())
	command = append(command, "--label", constant.LabelComposeServiceName+"="+name)
	command = append(command, "--label", constant.LabelConfigKey+"="+service.GetUnique())

	//环境
	for k, v := range service.Environment {
		command = append(command, "--env", k+"="+v)
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
	_, err := cli.GetImage(connection, image, nil)
	if err != nil {
		return command, err
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

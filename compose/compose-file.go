package compose

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
	"podman-compose/util"
)

// ServiceResources 定义了服务资源的限制
type ServiceResources struct {
	Limits struct {
		CPUs   float64 `yaml:"cpus,omitempty"`
		Memory string  `yaml:"memory,omitempty"`
	} `yaml:"limits,omitempty"`
}

// ServiceConfig 定义了服务的配置
type ServiceConfig struct {
	Image         string            `yaml:"image"`
	Restart       string            `yaml:"restart,omitempty"`
	Entrypoint    string            `yaml:"entrypoint,omitempty"`
	WorkingDir    string            `yaml:"working_dir,omitempty"`
	Deploy        ServiceResources  `yaml:"resources,omitempty"`
	ContainerName string            `yaml:"container_name,omitempty"`
	Command       []string          `yaml:"command,omitempty"`
	Ports         []string          `yaml:"ports,omitempty"`
	Environment   map[string]string `yaml:"environment,omitempty"`
	Volumes       []string          `yaml:"volumes,omitempty"`
}

// DockerCompose 定义了整个docker-compose的配置
type DockerCompose struct {
	Version  string                   `yaml:"version"`
	Services map[string]ServiceConfig `yaml:"services"`
	Workdir  string
}

var dockerCompose DockerCompose

// GetDockerCompose /*
func GetDockerCompose() DockerCompose {
	return dockerCompose
}

func GetComposeDir() string {
	dir, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
	}
	return dir
}

var fixServiceNameSize = 10

func FormatServiceName(name string) string {
	return util.FixSizeString(name, fixServiceNameSize, false)
}
func InitCompose() error {
	file, err := getComposeFile()
	if err != nil {
		return err
	}
	yaml.NewDecoder(file).Decode(&dockerCompose)

	for key := range dockerCompose.Services {
		if len(key) >= fixServiceNameSize {
			fixServiceNameSize = len(key) + 1
		}
	}
	return nil
}

var fileNames = []string{"docker-compose.yml", "docker-compose.yaml", "compose.yml", "compose.yaml"}

/*
*
获取compose配置文件
*/
func getComposeFile() (*os.File, error) {
	for _, fileName := range fileNames {
		_, err := os.Stat(fileName)
		if err == nil {
			return os.Open(fileName)
		}
	}
	return nil, fmt.Errorf(`ERROR: 
        Can't find a suitable configuration file in this directory or any
        parent. Are you in the right directory?

        Supported filenames: docker-compose.yml, docker-compose.yaml, compose.yml, compose.yaml`)
}

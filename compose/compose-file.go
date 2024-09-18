package compose

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
	"podman-compose/util"
	"strings"
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
	Image         string           `yaml:"image"`
	Restart       string           `yaml:"restart,omitempty"`
	Entrypoint    string           `yaml:"entrypoint,omitempty"`
	WorkingDir    string           `yaml:"working_dir,omitempty"`
	Deploy        ServiceResources `yaml:"resources,omitempty"`
	ContainerName string           `yaml:"container_name,omitempty"`
	Command       []string         `yaml:"command,omitempty"`
	Ports         []string         `yaml:"ports,omitempty"`
	Environment   any              `yaml:"environment,omitempty"`
	Volumes       []string         `yaml:"volumes,omitempty"`
}

func (c *ServiceConfig) GetEnvironment() (map[string]string, error) {
	if c.Environment == nil {
		return nil, nil
	}
	result := make(map[string]string)
	envMap, ok := c.Environment.(map[string]any)
	if ok {
		for key, val := range envMap {
			result[fmt.Sprintf("%v", key)] = fmt.Sprintf("%v", val)
		}
		return result, nil
	}

	list, ok := c.Environment.([]interface{})
	if ok {
		for _, item := range list {
			kvString, ok := item.(string)
			if !ok {
				return nil, fmt.Errorf("environment \"", item, "\" format error")
			}
			idx := strings.IndexByte(kvString, '=')
			if idx == -1 {
				return nil, fmt.Errorf("environment \"", kvString, "\" format error")
			}
			result[kvString[:idx]] = kvString[idx+1:]
		}
		return result, nil
	}
	return nil, fmt.Errorf("environment format error")
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

		svr := dockerCompose.Services[key]
		_, err = svr.GetEnvironment()
		if err != nil {
			return err
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

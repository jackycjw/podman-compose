package cli

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"
)

// ContainerSize holds the size of the container's root filesystem and top
// read-write layer.
type ContainerSize struct {
	RootFsSize int64 `json:"rootFsSize"`
	RwSize     int64 `json:"rwSize"`
}

type ListContainer struct {
	// Container command
	Command []string
	// If container has exited/stopped
	Exited bool
	// Time container exited
	ExitedAt int64
	// If container has exited, the return code from the command
	ExitCode int32
	// The unique identifier for the container
	ID string `json:"Id"`
	// Container image
	Image string
	// If this container is a Pod infra container
	IsInfra bool
	// Labels for container
	Labels map[string]string
	// User volume mounts
	Mounts []string
	// The names assigned to the container
	Names []string
	// Namespaces the container belongs to.  Requires the
	// namespace boolean to be true
	// The process id of the container
	Pid int
	// If the container is part of Pod, the Pod ID. Requires the pod
	// boolean to be set
	Pod string
	// If the container is part of Pod, the Pod name. Requires the pod
	// boolean to be set
	PodName string
	// Port mappings
	Ports []PortMapping
	// Size of the container rootfs.  Requires the size boolean to be true
	Size *ContainerSize
	// Time when container started
	StartedAt int64
	// State of container
	State string
}

type PortMapping struct {
	// HostPort is the port number on the host.
	HostPort  int32 `json:"hostPort"`
	HostPort2 int32 `json:"host_port"`
	// ContainerPort is the port number inside the sandbox.
	ContainerPort  int32 `json:"containerPort"`
	ContainerPort2 int32 `json:"container_port"`
	// Protocol is the protocol of the port mapping.
	Protocol string `json:"protocol"`
	// HostIP is the host ip to use.
	HostIP  string `json:"hostIP"`
	HostIP2 string `json:"host_ip"`
}

func (r *PortMapping) String() string {
	hostIp := r.HostIP + r.HostIP2
	hostPort := r.HostPort + r.HostPort2
	containerPort := r.ContainerPort + r.ContainerPort2
	if hostIp == "" {
		hostIp = "0.0.0.0"
	}
	return hostIp + ":" + strconv.Itoa(int(hostPort)) + "->" + strconv.Itoa(int(containerPort)) + "/" + r.Protocol
}

type ContainerDetail struct {
	Image      string                      `json:"Image"`
	ImageName  string                      `json:"ImageName"`
	HostConfig *InspectContainerHostConfig `json:"HostConfig"`
}

type InspectContainerHostConfig struct {
	// RestartPolicy contains the container's restart policy.
	RestartPolicy *InspectRestartPolicy `json:"RestartPolicy"`
}

// InspectRestartPolicy holds information about the container's restart policy.
type InspectRestartPolicy struct {
	// Name contains the container's restart policy.
	Name string `json:"Name"`
}

type ImageData struct {
	ID string `json:"Id"`
}

var connection context.Context

func init() {
	var err error
	connection, err = NewConnection(context.Background(), "unix:///run/podman/podman.sock")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
func List(filters map[string][]string, all *bool, last *int, pod, size, sync *bool) ([]ListContainer, error) { // nolint:typecheck
	conn, err := GetClient(connection)
	if err != nil {
		return nil, err
	}
	var containers []ListContainer
	params := url.Values{}
	if all != nil {
		params.Set("all", strconv.FormatBool(*all))
	}
	if last != nil {
		params.Set("last", strconv.Itoa(*last))
	}
	if pod != nil {
		params.Set("pod", strconv.FormatBool(*pod))
	}
	if size != nil {
		params.Set("size", strconv.FormatBool(*size))
	}
	if sync != nil {
		params.Set("sync", strconv.FormatBool(*sync))
	}
	if filters != nil {
		filterString, err := FiltersToString(filters)
		if err != nil {
			return nil, err
		}
		params.Set("filters", filterString)
	}
	response, err := conn.DoRequest(nil, http.MethodGet, "/containers/json", params)
	if err != nil {
		return containers, err
	}
	return containers, response.Process(&containers)
}

func Inspect(nameOrID string, size *bool) (*ContainerDetail, error) {
	conn, err := GetClient(connection)
	if err != nil {
		return nil, err
	}
	params := url.Values{}
	if size != nil {
		params.Set("size", strconv.FormatBool(*size))
	}
	response, err := conn.DoRequest(nil, http.MethodGet, "/containers/%s/json", params, nameOrID)
	if err != nil {
		return nil, err
	}
	inspect := ContainerDetail{}
	return &inspect, response.Process(&inspect)
}

func GetImage(nameOrID string, size *bool) (*ImageData, error) {
	conn, err := GetClient(connection)
	if err != nil {
		return nil, err
	}
	params := url.Values{}
	if size != nil {
		params.Set("size", strconv.FormatBool(*size))
	}
	inspectedData := ImageData{}
	response, err := conn.DoRequest(nil, http.MethodGet, "/images/%s/json", params, nameOrID)
	if err != nil {
		return &inspectedData, err
	}
	return &inspectedData, response.Process(&inspectedData)
}

// Remove removes a container from local storage.  The force bool designates
// that the container should be removed forcibly (example, even it is running).  The volumes
// bool dictates that a container's volumes should also be removed.
func Remove(nameOrID string, force, volumes *bool) error {
	conn, err := GetClient(connection)
	if err != nil {
		return err
	}
	params := url.Values{}
	if force != nil {
		params.Set("force", strconv.FormatBool(*force))
	}
	if volumes != nil {
		params.Set("vols", strconv.FormatBool(*volumes))
	}
	response, err := conn.DoRequest(nil, http.MethodDelete, "/containers/%s", params, nameOrID)
	if err != nil {
		return err
	}
	return response.Process(nil)
}

func Start(nameOrID string, detachKeys *string) error {
	conn, err := GetClient(connection)
	if err != nil {
		return err
	}
	params := url.Values{}
	if detachKeys != nil {
		params.Set("detachKeys", *detachKeys)
	}
	response, err := conn.DoRequest(nil, http.MethodPost, "/containers/%s/start", params, nameOrID)
	if err != nil {
		return err
	}
	return response.Process(nil)
}

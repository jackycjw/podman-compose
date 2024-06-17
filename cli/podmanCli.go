package cli

import (
	"context"
	"github.com/containers/libpod/cmd/podman/shared"
	"github.com/containers/libpod/pkg/bindings"
	"github.com/cri-o/ocicni/pkg/ocicni"
	"net/http"
	"net/url"
	"strconv"
)

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
	Ports []ocicni.PortMapping
	// Size of the container rootfs.  Requires the size boolean to be true
	Size *shared.ContainerSize
	// Time when container started
	StartedAt int64
	// State of container
	State string
}

type ContainerDetail struct {
	Image     string `json:"Image"`
	ImageName string `json:"ImageName"`
}

type ImageData struct {
	ID string `json:"Id"`
}

func List(ctx context.Context, filters map[string][]string, all *bool, last *int, pod, size, sync *bool) ([]ListContainer, error) { // nolint:typecheck
	conn, err := bindings.GetClient(ctx)
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
		filterString, err := bindings.FiltersToString(filters)
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

func Inspect(ctx context.Context, nameOrID string, size *bool) (*ContainerDetail, error) {
	conn, err := bindings.GetClient(ctx)
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

func GetImage(ctx context.Context, nameOrID string, size *bool) (*ImageData, error) {
	conn, err := bindings.GetClient(ctx)
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

package docker

import (
	"context"
	"fmt"
	"strings"

	"github.com/docker/go-connections/nat"
	"google3/third_party/golang/github_com/moby/moby/v/v24/api/types/container/container"
	"google3/third_party/golang/github_com/moby/moby/v/v24/api/types/mount/mount"
	"google3/third_party/golang/github_com/moby/moby/v/v24/api/types/network/network"
	"google3/third_party/golang/github_com/moby/moby/v/v24/api/types/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"github.com/openconfig/containerz/containers"

	cpb "github.com/openconfig/gnoi/containerz"
)

// ContainerStart starts a container provided the image exists and that the ports requested are not
// currently in use.
func (m *Manager) ContainerStart(ctx context.Context, image, tag, cmd string, opts ...options.Option) (string, error) {
	optionz := options.ApplyOptions(opts...)

	images, err := m.client.ImageList(ctx, types.ImageListOptions{
		// TODO(alshabib): consider filtering for the image we care about
	})
	if err != nil {
		return "", err
	}

	ref := fmt.Sprintf("%s:%s", image, tag)
	if err := findImage(ref, images); err != nil {
		return "", err
	}

	cnts, err := m.client.ContainerList(ctx, types.ContainerListOptions{
		// TODO(alshabib): consider filtering for the image we care about
	})
	if err != nil {
		return "", err
	}

	if err := checkExistingInstanceAndPorts(optionz.InstanceName, optionz.PortMapping, cnts); err != nil {
		return "", err
	}

	mounts := make([]mount.Mount, 0, len(optionz.Volumes))
	for _, vol := range optionz.Volumes {
		mounts = append(mounts, mount.Mount{
			Type:     "volume",
			Source:   vol.GetName(),
			Target:   vol.GetMountPoint(),
			ReadOnly: vol.GetReadOnly(),
		})
	}

	// TODO(alshabib): add resource support (i.e. CPU and memory quotas.)
	hostConfig := &container.HostConfig{
		Mounts:      mounts,
		NetworkMode: "host",
		AutoRemove:  true,
	}
	config := &container.Config{
		Cmd:          strings.Split(cmd, " "),
		Image:        ref,
		AttachStdin:  false,
		AttachStdout: false,
		AttachStderr: false,
		StdinOnce:    false,
		Tty:          true,
	}
	if len(optionz.PortMapping) > 0 {
		portMap := nat.PortMap{}
		portSet := nat.PortSet{}
		for in, out := range optionz.PortMapping {
			internal := fmt.Sprintf("%d", in)
			external := fmt.Sprintf("%d", out)
			in, err := nat.NewPort("tcp", internal)
			if err != nil {
				return "", err
			}

			portSet[in] = struct{}{}
			bindingV4 := nat.PortBinding{
				HostIP:   "0.0.0.0", // TODO(alshabib): do we want this to be configurable?
				HostPort: external,
			}
			bindingV6 := nat.PortBinding{
				HostIP:   "::",
				HostPort: external,
			}

			portMap[in] = []nat.PortBinding{bindingV4, bindingV6}
		}

		hostConfig.PortBindings = portMap
		config.ExposedPorts = portSet
	}

	if len(optionz.EnvMapping) > 0 {
		for envName, envVal := range optionz.EnvMapping {
			config.Env = append(config.Env, fmt.Sprintf("%s=%s", envName, envVal))
		}
	}

	// Handle Network
	if optionz.Network != "" {
		hostConfig.NetworkMode = container.NetworkMode(optionz.Network)
	}

	// Handle Capabilities
	if optionz.Capabilities != nil {
		caps := optionz.Capabilities.(*cpb.StartContainerRequest_Capabilities)
		hostConfig.CapAdd = caps.GetAdd()
		hostConfig.CapDrop = caps.GetRemove()
	}

	// Handle RestartPolicy
	if optionz.RestartPolicy != nil {
		restartPolicy := optionz.RestartPolicy.(*cpb.StartContainerRequest_Restart)

		var policy string
		switch restartPolicy.GetPolicy() {
		case cpb.StartContainerRequest_Restart_ALWAYS:
			policy = "always"
		case cpb.StartContainerRequest_Restart_ON_FAILURE:
			policy = "on-failure"
		case cpb.StartContainerRequest_Restart_NONE:
			policy = "no"
		case cpb.StartContainerRequest_Restart_UNLESS_STOPPED:
			policy = "unless-stopped"
		default:
			return "", status.Errorf(codes.FailedPrecondition, "unkown restart policy '%v'", restartPolicy.GetPolicy())
		}

		hostConfig.RestartPolicy = container.RestartPolicy{
			Name:              policy,
			MaximumRetryCount: int(restartPolicy.GetAttempts()),
		}
	}

	// Handle RunAs
	if optionz.RunAs != nil {
		runAs := optionz.RunAs.(*cpb.StartContainerRequest_RunAs)
		user := runAs.GetUser()
		if user == "" {
			return "", status.Errorf(codes.FailedPrecondition, "user can not be empty in RunAs option")
		}
		if runAs.GetGroup() != "" {
			user = fmt.Sprintf("%s:%s", user, runAs.GetGroup())
		}
		config.User = user
	}

	resp, err := m.client.ContainerCreate(ctx, config, hostConfig, &network.NetworkingConfig{}, nil, optionz.InstanceName)
	if err != nil {
		return "", status.Errorf(codes.Internal, "unable to create container: %v", err)
	}

	if err := m.client.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		return "", status.Errorf(codes.Internal, "unable to start container: %v", err)
	}

	name := resp.ID
	if optionz.InstanceName != "" {
		name = optionz.InstanceName
	}

	return name, nil
}

func checkExistingInstanceAndPorts(instance string, ports map[uint32]uint32, cnts []types.Container) error {
	if instance == "" && len(ports) == 0 {
		return nil
	}

	for _, cnt := range cnts {
		for _, name := range cnt.Names {
			strippedname := strings.Replace(name, "/", "", 1)
			if strippedname == instance {
				return status.Errorf(codes.AlreadyExists, "instance name %s already in use", instance)
			}
		}
		for _, port := range cnt.Ports {
			for _, ext := range ports {
				if ext == uint32(port.PublicPort) {
					return status.Errorf(codes.Unavailable, "port %d already in use", ext)
				}
			}
		}
	}
	return nil
}

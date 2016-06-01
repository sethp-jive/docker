// +build !windows,!solaris

package daemon

import (
	"github.com/docker/docker/api/types/backend"
	"github.com/docker/docker/container"
	"github.com/docker/docker/daemon/exec"
	"github.com/docker/engine-api/types"
	"github.com/docker/engine-api/types/versions/v1p19"
)

// This sets platform-specific fields
func setPlatformSpecificContainerFields(container *container.Container, contJSONBase *types.ContainerJSONBase) *types.ContainerJSONBase {
	contJSONBase.AppArmorProfile = container.AppArmorProfile
	contJSONBase.ResolvConfPath = container.ResolvConfPath
	contJSONBase.HostnamePath = container.HostnamePath
	contJSONBase.HostsPath = container.HostsPath

	return contJSONBase
}

// containerInspectPre120 gets containers for pre 1.20 APIs.
func (daemon *Daemon) containerInspectPre120(name string) (*v1p19.ContainerJSON, error) {
	container, err := daemon.GetContainer(name)
	if err != nil {
		return nil, err
	}

	container.DataLock.Lock()
	defer container.DataLock.Unlock()

	base, err := daemon.getInspectData(container, false)
	if err != nil {
		return nil, err
	}

	volumes := make(map[string]string)
	volumesRW := make(map[string]bool)
	for _, m := range container.MountPoints {
		volumes[m.Destination] = m.Path()
		volumesRW[m.Destination] = m.RW
	}

	config := &v1p19.ContainerConfig{
		Config:          container.Config,
		MacAddress:      container.Config.MacAddress,
		NetworkDisabled: container.Config.NetworkDisabled,
		ExposedPorts:    container.Config.ExposedPorts,
		VolumeDriver:    container.HostConfig.VolumeDriver,
		Memory:          container.HostConfig.Memory,
		MemorySwap:      container.HostConfig.MemorySwap,
		CPUShares:       container.HostConfig.CPUShares,
		CPUSet:          container.HostConfig.CpusetCpus,
	}
	networkSettings := daemon.getBackwardsCompatibleNetworkSettings(container.NetworkSettings)

	return &v1p19.ContainerJSON{
		ContainerJSONBase: base,
		Volumes:           volumes,
		VolumesRW:         volumesRW,
		Config:            config,
		NetworkSettings:   networkSettings,
	}, nil
}

func addMountPoints(container *container.Container) []types.MountPoint {
	mountPoints := make([]types.MountPoint, 0, len(container.MountPoints))
	for _, m := range container.MountPoints {
		mountPoints = append(mountPoints, types.MountPoint{
			Name:        m.Name,
			Source:      m.Path(),
			Destination: m.Destination,
			Driver:      m.Driver,
			Mode:        m.Mode,
			RW:          m.RW,
			Propagation: m.Propagation,
		})
	}
	return mountPoints
}

func inspectExecProcessConfig(e *exec.Config) *backend.ExecProcessConfig {
	return &backend.ExecProcessConfig{
		Tty:        e.Tty,
		Entrypoint: e.Entrypoint,
		Arguments:  e.Args,
		Privileged: &e.Privileged,
		User:       e.User,
	}
}

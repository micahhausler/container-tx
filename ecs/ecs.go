package ecs

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"sort"
	"strings"

	"github.com/micahhausler/container-tx/transform"
)

func (c Container) ingestEnvironment() map[string]string {
	if c.Environment != nil {
		env := map[string]string{}
		for _, envVar := range *c.Environment {
			env[envVar.Name] = envVar.Value
		}
		return env
	}
	return nil
}

func (c *Container) emitEnvironment(env map[string]string) {
	if len(env) > 0 {
		envs := Environments{}
		for n, v := range env {
			envs = append(envs, Environment{Name: n, Value: v})
		}
		sort.Sort(envs)
		c.Environment = &envs
	}
}

// Environment is a type for storing ECS environment objects
type Environment struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// Environments is a type for slices of ECS environment objects
type Environments []Environment

func (env Environments) Len() int      { return len(env) }
func (env Environments) Swap(i, j int) { env[i], env[j] = env[j], env[i] }
func (env Environments) Less(i, j int) bool {
	return strings.Compare(env[i].Name, env[j].Name) < 0
}

func (c Container) ingestLogging() *transform.Logging {
	if c.Logging != nil {
		return &transform.Logging{
			Driver:  c.Logging.Driver,
			Options: c.Logging.Options,
		}
	}
	return nil
}

func (c *Container) emitLogging(l *transform.Logging) {
	if l != nil {
		c.Logging = &Logging{
			Driver:  l.Driver,
			Options: l.Options,
		}
	}
}

// Logging is a type for storing ECS Logging information
type Logging struct {
	Driver  string            `json:"logDriver"`
	Options map[string]string `json:"options"`
}

func (c Container) ingestMemory() int {
	var memoryIn = c.Memory << 20
	if memoryIn == 0 {
		memoryIn = 4 << 20
	}
	return memoryIn
}

func (c *Container) emitMemory(mem int) {
	memInMb := mem >> 20
	if 4 > memInMb {
		c.Memory = 4
	} else {
		c.Memory = memInMb
	}
}

func (c Container) ingestPortMappings() *transform.PortMappings {
	if c.PortMappings != nil && len(*c.PortMappings) > 0 {
		response := transform.PortMappings{}
		for _, pm := range *c.PortMappings {
			response = append(response, transform.PortMapping{
				HostPort:      pm.HostPort,
				ContainerPort: pm.ContainerPort,
				Protocol:      strings.ToLower(pm.Protocol),
			})
		}
		return &response
	}
	return nil
}

func (c *Container) emitPortMappings(in *transform.PortMappings) {
	if in != nil && len(*in) > 0 {
		output := PortMappings{}
		for _, pm := range *in {
			output = append(output, PortMapping{
				HostPort:      pm.HostPort,
				ContainerPort: pm.ContainerPort,
				Protocol:      strings.ToLower(pm.Protocol),
			})
		}
		sort.Sort(output)
		c.PortMappings = &output
	}
}

// PortMapping is a type for storing ECS port information
type PortMapping struct {
	HostPort      int    `json:"hostPort,omitempty"`
	ContainerPort int    `json:"containerPort"`
	Protocol      string `json:"protocol,omitempty"`
}

// PortMappings is a composite type for slices of EcsPortMapping
type PortMappings []PortMapping

func (pm PortMappings) Len() int           { return len(pm) }
func (pm PortMappings) Swap(i, j int)      { pm[i], pm[j] = pm[j], pm[i] }
func (pm PortMappings) Less(i, j int) bool { return pm[i].ContainerPort < pm[j].ContainerPort }

func (c Container) ingestVolumes(volumeMap map[string]string) *transform.IntermediateVolumes {
	if c.Volumes != nil && len(*c.Volumes) > 0 {
		response := transform.IntermediateVolumes{}
		for _, vol := range *c.Volumes {
			iv := transform.IntermediateVolume{
				Container: vol.ContainerPath,
				ReadOnly:  vol.ReadOnly,
				Host:      volumeMap[vol.SourceVolume],
			}
			response = append(response, iv)
		}
		return &response
	}
	return nil
}

func (c *Container) emitVolumes(vols *transform.IntermediateVolumes) map[string]string {
	response := map[string]string{}
	if vols != nil && len(*vols) > 0 {
		mountPoints := MountPoints{}
		for _, volume := range *vols {
			sourceVolume := strings.Trim(strings.Replace(volume.Host, "/", "-", -1), "-")
			response[sourceVolume] = volume.Host
			mountPoints = append(mountPoints, MountPoint{
				SourceVolume:  sourceVolume,
				ContainerPath: volume.Container,
				ReadOnly:      volume.ReadOnly,
			})
		}
		sort.Sort(mountPoints)
		c.Volumes = &mountPoints
	}
	return response
}

// MountPoint is a type for storing ECS mount information
type MountPoint struct {
	SourceVolume  string `json:"sourceVolume"`
	ContainerPath string `json:"containerPath"`
	ReadOnly      bool   `json:"readOnly,omitempty"`
}

// MountPoints is a composite type for slices of MountPoint
type MountPoints []MountPoint

func (mp MountPoints) Len() int      { return len(mp) }
func (mp MountPoints) Swap(i, j int) { mp[i], mp[j] = mp[j], mp[i] }
func (mp MountPoints) Less(i, j int) bool {
	return strings.Compare(mp[i].ContainerPath, mp[j].ContainerPath) < 0
}

func (c Container) ingestVolumesFrom() []string {
	if c.VolumesFrom != nil {
		response := []string{}
		format := func(vf VolumeFrom) string {
			out := vf.SourceContainer
			if vf.ReadOnly {
				out += ":ro"
			}
			return out
		}
		for _, v := range *c.VolumesFrom {
			response = append(response, format(v))
		}
		return response
	}
	return nil
}

func (c *Container) emitVolumesFrom(vsf []string) {
	if len(vsf) > 0 {
		evf := VolumesFrom{}
		for _, vf := range vsf {
			evf = append(evf, VolumeFrom{
				SourceContainer: strings.SplitN(vf, ":", 2)[0],
				ReadOnly:        strings.HasSuffix(vf, ":ro"),
			})
		}
		sort.Sort(evf)
		c.VolumesFrom = &evf
	}
}

// VolumeFrom is a type for storing VolumeFrom information
type VolumeFrom struct {
	SourceContainer string `json:"sourceContainer"`
	ReadOnly        bool   `json:"readOnly,omitempty"`
}

// VolumesFrom is a composite type for slices of VolumeFrom
type VolumesFrom []VolumeFrom

func (evf VolumesFrom) Len() int      { return len(evf) }
func (evf VolumesFrom) Swap(i, j int) { evf[i], evf[j] = evf[j], evf[i] }
func (evf VolumesFrom) Less(i, j int) bool {
	return strings.Compare(evf[i].SourceContainer, evf[j].SourceContainer) < 0
}

// Container represents the ECS container information
type Container struct {
	Command      []string          `json:"command,omitempty"`
	CPU          int               `json:"cpu,omitempty"`
	DNS          []string          `json:"dnsServers,omitempty"`
	Domain       []string          `json:"dnsSearchDomains,omitempty"`
	Entrypoint   []string          `json:"entryPoint,omitempty"`
	Environment  *Environments     `json:"environment,omitempty"`
	Essential    bool              `json:"essential,omitempty"`
	Hostname     string            `json:"hostname,omitempty"`
	Image        string            `json:"image" ctx:"required"`
	Labels       map[string]string `json:"dockerLabels"`
	Links        []string          `json:"links,omitempty"`
	Logging      *Logging          `json:"logConfiguration,omitempty"`
	Memory       int               `json:"memory" ctx:"required"`
	Name         string            `json:"name" ctx:"required"`
	NetworkMode  string            `json:"networkMode,omitempty"`
	PortMappings *PortMappings     `json:"portMappings,omitempty"`
	Privileged   bool              `json:"privileged,omitempty"`
	User         string            `json:"user,omitempty"`
	Volumes      *MountPoints      `json:"mountPoints,omitempty"`
	VolumesFrom  *VolumesFrom      `json:"volumesFrom,omitempty"`
	WorkDir      string            `json:"workingDirectory,omitempty"`
}

// Containers is a composite type for a slice of ECS Containers
type Containers []Container

func (ec Containers) Len() int      { return len(ec) }
func (ec Containers) Swap(i, j int) { ec[i], ec[j] = ec[j], ec[i] }
func (ec Containers) Less(i, j int) bool {
	return strings.Compare(ec[i].Name, ec[j].Name) < 0
}

// Volume is a type for storing a task-level volume
type Volume struct {
	Name string     `json:"name"`
	Host VolumeHost `json:"host"`
}

// VolumeHost is a type for storing task-level volume's host path
type VolumeHost struct {
	SourcePath string `json:"sourcePath"`
}

// Volumes is a composite type for slices of ECS Volume
type Volumes []Volume

func (evs Volumes) Len() int      { return len(evs) }
func (evs Volumes) Swap(i, j int) { evs[i], evs[j] = evs[j], evs[i] }
func (evs Volumes) Less(i, j int) bool {
	return strings.Compare(evs[i].Name, evs[j].Name) < 0
}

// Task represents an ECS Task. It implements InputFormat and OutputFormat
type Task struct {
	Family               string      `json:"family"`
	NetworkMode          string      `json:"networkMode,omitempty"`
	ContainerDefinitions *Containers `json:"containerDefinitions"`
	Volumes              *Volumes    `json:"volumes"`
}

func volumesToMap(vols *Volumes) map[string]string {
	response := map[string]string{}
	for _, vol := range *vols {
		response[vol.Name] = vol.Host.SourcePath
	}
	return response
}

func mapToVolumes(names map[string]string) *Volumes {
	response := Volumes{}
	for name, path := range names {
		response = append(response, Volume{Name: name, Host: VolumeHost{SourcePath: path}})
	}
	return &response
}

// IngestContainers satisfies InputFormat so ECS tasks can be ingested
func (t Task) IngestContainers(input io.ReadCloser) (*transform.PodData, error) {

	body, err := ioutil.ReadAll(input)
	defer input.Close()
	if err != nil && err != io.EOF {
		return nil, err
	}
	err = json.Unmarshal(body, &t)
	if err != nil {
		return nil, err
	}

	outputPod := transform.PodData{Name: t.Family}
	containers := transform.Containers{}

	volMap := volumesToMap(t.Volumes)

	for _, container := range *t.ContainerDefinitions {
		ir := transform.Container{}
		if len(container.Command) > 0 {
			ir.Command = strings.Join(container.Command, " ")
		}
		ir.CPU = container.CPU
		ir.DNS = container.DNS
		ir.Domain = container.Domain
		if len(container.Entrypoint) > 0 {
			ir.Entrypoint = strings.Join(container.Entrypoint, " ")
		}
		ir.Environment = container.ingestEnvironment()
		ir.Essential = container.Essential
		ir.Hostname = container.Hostname
		ir.Image = container.Image
		ir.Labels = container.Labels
		ir.Links = container.Links
		ir.Logging = container.ingestLogging()
		ir.Memory = container.ingestMemory()
		ir.Name = container.Name
		ir.NetworkMode = container.NetworkMode
		ir.PortMappings = container.ingestPortMappings()
		ir.Privileged = container.Privileged
		ir.User = container.User
		ir.Volumes = container.ingestVolumes(volMap)
		ir.VolumesFrom = container.ingestVolumesFrom()
		ir.WorkDir = container.WorkDir
		containers = append(containers, ir)
	}
	sort.Sort(containers)
	outputPod.Containers = &containers

	return &outputPod, nil
}

// EmitContainers satisfies OutputFormat so ECS tasks can be emitted
func (t Task) EmitContainers(input *transform.PodData) ([]byte, error) {
	output := &Task{Family: input.Name}
	containers := Containers{}

	volumesMap := map[string]string{}

	for _, container := range *input.Containers {
		EcsContainer := Container{}
		if len(container.Command) > 0 {
			EcsContainer.Command = strings.Split(container.Command, " ")
		}
		EcsContainer.CPU = container.CPU
		EcsContainer.DNS = container.DNS
		EcsContainer.Domain = container.Domain
		if len(container.Entrypoint) > 0 {
			EcsContainer.Entrypoint = strings.Split(container.Entrypoint, " ")
		}
		EcsContainer.emitEnvironment(container.Environment)
		EcsContainer.Hostname = container.Hostname
		EcsContainer.Image = container.Image
		EcsContainer.Labels = container.Labels
		EcsContainer.Links = container.Links
		EcsContainer.emitLogging(container.Logging)
		EcsContainer.emitMemory(container.Memory)
		EcsContainer.Name = container.Name
		EcsContainer.NetworkMode = container.NetworkMode
		EcsContainer.emitPortMappings(container.PortMappings)
		EcsContainer.Privileged = container.Privileged
		EcsContainer.User = container.User
		for k, v := range EcsContainer.emitVolumes(container.Volumes) {
			volumesMap[k] = v
		}
		EcsContainer.emitVolumesFrom(container.VolumesFrom)
		EcsContainer.WorkDir = container.WorkDir
		containers = append(containers, EcsContainer)
	}
	output.Volumes = mapToVolumes(volumesMap)
	sort.Sort(output.Volumes)

	sort.Sort(containers)
	output.ContainerDefinitions = &containers

	return json.MarshalIndent(output, "", "    ")
}

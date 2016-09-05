package ecs

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"sort"
	"strings"

	"github.com/micahhausler/container-tx/transform"
)

func (c EcsContainer) ingestEnvironment() map[string]string {
	if c.Environment != nil {
		env := map[string]string{}
		for _, envVar := range *c.Environment {
			env[envVar.Name] = envVar.Value
		}
		return env
	}
	return nil
}

func (c *EcsContainer) emitEnvironment(env map[string]string) {
	if len(env) > 0 {
		envs := EcsEnvironments{}
		for n, v := range env {
			envs = append(envs, EcsEnvironment{Name: n, Value: v})
		}
		sort.Sort(envs)
		c.Environment = &envs
	}
}

// EcsEnvironment is a type for storing ECS environment objects
type EcsEnvironment struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// EcsEnvironments is a type for collecting ECS environment objects
type EcsEnvironments []EcsEnvironment

func (env EcsEnvironments) Len() int      { return len(env) }
func (env EcsEnvironments) Swap(i, j int) { env[i], env[j] = env[j], env[i] }
func (env EcsEnvironments) Less(i, j int) bool {
	return strings.Compare(env[i].Name, env[j].Name) < 0
}

func (c EcsContainer) ingestLogging() *transform.Logging {
	if c.Logging != nil {
		return &transform.Logging{
			Driver:  c.Logging.Driver,
			Options: c.Logging.Options,
		}
	}
	return nil
}

func (c *EcsContainer) emitLogging(l *transform.Logging) {
	if l != nil {
		c.Logging = &EcsLogging{
			Driver:  l.Driver,
			Options: l.Options,
		}
	}
}

// EcsLogging is a type for storing ECS Logging information
type EcsLogging struct {
	Driver  string            `json:"logDriver"`
	Options map[string]string `json:"options"`
}

func (c EcsContainer) ingestMemory() int {
	var memoryIn = c.Memory << 20
	if memoryIn == 0 {
		memoryIn = 4 << 20
	}
	return memoryIn
}

func (c *EcsContainer) emitMemory(mem int) {
	memInMb := mem >> 20
	if 4 > memInMb {
		c.Memory = 4
	} else {
		c.Memory = memInMb
	}
}

func (c EcsContainer) ingestPortMappings() *transform.PortMappings {
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

func (c *EcsContainer) emitPortMappings(in *transform.PortMappings) {
	if in != nil && len(*in) > 0 {
		output := EcsPortMappings{}
		for _, pm := range *in {
			output = append(output, EcsPortMapping{
				HostPort:      pm.HostPort,
				ContainerPort: pm.ContainerPort,
				Protocol:      strings.ToLower(pm.Protocol),
			})
		}
		sort.Sort(output)
		c.PortMappings = &output
	}
}

// EcsPortMapping is a type for storing ECS port information
type EcsPortMapping struct {
	HostPort      int    `json:"hostPort,omitempty"`
	ContainerPort int    `json:"containerPort"`
	Protocol      string `json:"protocol,omitempty"`
}

// EcsPortMappings is a composite type for slices of EcsPortMapping
type EcsPortMappings []EcsPortMapping

func (pm EcsPortMappings) Len() int           { return len(pm) }
func (pm EcsPortMappings) Swap(i, j int)      { pm[i], pm[j] = pm[j], pm[i] }
func (pm EcsPortMappings) Less(i, j int) bool { return pm[i].ContainerPort < pm[j].ContainerPort }

func (c EcsContainer) ingestVolumes(volumeMap map[string]string) *transform.IntermediateVolumes {
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

func (c *EcsContainer) emitVolumes(vols *transform.IntermediateVolumes) map[string]string {
	response := map[string]string{}
	if vols != nil && len(*vols) > 0 {
		mountPoints := EcsMountPoints{}
		for _, volume := range *vols {
			sourceVolume := strings.Trim(strings.Replace(volume.Host, "/", "-", -1), "-")
			response[sourceVolume] = volume.Host
			mountPoints = append(mountPoints, EcsMountPoint{
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

// EcsMountPoint is a type for storing ECS mount information
type EcsMountPoint struct {
	SourceVolume  string `json:"sourceVolume"`
	ContainerPath string `json:"containerPath"`
	ReadOnly      bool   `json:"readOnly,omitempty"`
}

// EcsMountPoints is a composite type for slices of EcsMountPoint
type EcsMountPoints []EcsMountPoint

func (mp EcsMountPoints) Len() int      { return len(mp) }
func (mp EcsMountPoints) Swap(i, j int) { mp[i], mp[j] = mp[j], mp[i] }
func (mp EcsMountPoints) Less(i, j int) bool {
	return strings.Compare(mp[i].ContainerPath, mp[j].ContainerPath) < 0
}

func (c EcsContainer) ingestVolumesFrom() []string {
	if c.VolumesFrom != nil {
		response := []string{}
		format := func(vf EcsVolumeFrom) string {
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

func (c *EcsContainer) emitVolumesFrom(vsf []string) {
	if len(vsf) > 0 {
		evf := EcsVolumesFrom{}
		for _, vf := range vsf {
			evf = append(evf, EcsVolumeFrom{
				SourceContainer: strings.SplitN(vf, ":", 2)[0],
				ReadOnly:        strings.HasSuffix(vf, ":ro"),
			})
		}
		sort.Sort(evf)
		c.VolumesFrom = &evf
	}
}

// EcsVolumeFrom is a type for storing VolumeFrom information
type EcsVolumeFrom struct {
	SourceContainer string `json:"sourceContainer"`
	ReadOnly        bool   `json:"readOnly,omitempty"`
}

// EcsVolumesFrom is a composite type for slices of EcsVolumeFrom
type EcsVolumesFrom []EcsVolumeFrom

func (evf EcsVolumesFrom) Len() int      { return len(evf) }
func (evf EcsVolumesFrom) Swap(i, j int) { evf[i], evf[j] = evf[j], evf[i] }
func (evf EcsVolumesFrom) Less(i, j int) bool {
	return strings.Compare(evf[i].SourceContainer, evf[j].SourceContainer) < 0
}

// EcsContainer represents the ECS container information
type EcsContainer struct {
	Command      []string          `json:"command,omitempty"`
	CPU          int               `json:"cpu,omitempty"`
	DNS          []string          `json:"dnsServers,omitempty"`
	Domain       []string          `json:"dnsSearchDomains,omitempty"`
	Entrypoint   []string          `json:"entryPoint,omitempty"`
	Environment  *EcsEnvironments  `json:"environment,omitempty"`
	Essential    bool              `json:"essential,omitempty"`
	Hostname     string            `json:"hostname,omitempty"`
	Image        string            `json:"image" ctx:"required"`
	Labels       map[string]string `json:"dockerLabels"`
	Links        []string          `json:"links,omitempty"`
	Logging      *EcsLogging       `json:"logConfiguration,omitempty"`
	Memory       int               `json:"memory" ctx:"required"`
	Name         string            `json:"name" ctx:"required"`
	NetworkMode  string            `json:"networkMode,omitempty"`
	PortMappings *EcsPortMappings  `json:"portMappings,omitempty"`
	Privileged   bool              `json:"privileged,omitempty"`
	User         string            `json:"user,omitempty"`
	Volumes      *EcsMountPoints   `json:"mountPoints,omitempty"`
	VolumesFrom  *EcsVolumesFrom   `json:"volumesFrom,omitempty"`
	WorkDir      string            `json:"workingDirectory,omitempty"`
}

// EcsContainers is a composite type for a slice of EcsContainers
type EcsContainers []EcsContainer

func (ec EcsContainers) Len() int      { return len(ec) }
func (ec EcsContainers) Swap(i, j int) { ec[i], ec[j] = ec[j], ec[i] }
func (ec EcsContainers) Less(i, j int) bool {
	return strings.Compare(ec[i].Name, ec[j].Name) < 0
}

// EcsVolume is a type for storing a task-level volume
type EcsVolume struct {
	Name string        `json:"name"`
	Host EcsVolumeHost `json:"host"`
}

// EcsVolumeHost is a type for storing task-level volume's host path
type EcsVolumeHost struct {
	SourcePath string `json:"sourcePath"`
}

// EcsVolumes is a composite type for slices of EcsVolume
type EcsVolumes []EcsVolume

func (evs EcsVolumes) Len() int      { return len(evs) }
func (evs EcsVolumes) Swap(i, j int) { evs[i], evs[j] = evs[j], evs[i] }
func (evs EcsVolumes) Less(i, j int) bool {
	return strings.Compare(evs[i].Name, evs[j].Name) < 0
}

// EcsFormat represents an ECS Task. It implements InputFormat and OutputFormat
type EcsFormat struct {
	Family               string         `json:"family"`
	NetworkMode          string         `json:"networkMode,omitempty"`
	ContainerDefinitions *EcsContainers `json:"containerDefinitions"`
	Volumes              *EcsVolumes    `json:"volumes"`
}

func volumesToMap(vols *EcsVolumes) map[string]string {
	response := map[string]string{}
	for _, vol := range *vols {
		response[vol.Name] = vol.Host.SourcePath
	}
	return response
}

func mapToVolumes(names map[string]string) *EcsVolumes {
	response := EcsVolumes{}
	for name, path := range names {
		response = append(response, EcsVolume{Name: name, Host: EcsVolumeHost{SourcePath: path}})
	}
	return &response
}

// IngestContainers satisfies InputFormat so ECS tasks can be ingested
func (f EcsFormat) IngestContainers(input io.ReadCloser) (*transform.BasePodData, error) {

	body, err := ioutil.ReadAll(input)
	defer input.Close()
	if err != nil && err != io.EOF {
		return nil, err
	}
	ef := &EcsFormat{}
	err = json.Unmarshal(body, ef)
	if err != nil {
		return nil, err
	}

	outputPod := transform.BasePodData{Name: ef.Family}
	outputPod.Containers = []*transform.BaseContainerFormat{}

	volMap := volumesToMap(ef.Volumes)

	for _, container := range *ef.ContainerDefinitions {
		ir := transform.BaseContainerFormat{}
		outputPod.Containers = append(outputPod.Containers, &ir)
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
	}

	return &outputPod, nil
}

// EmitContainers satisfies OutputFormat so ECS tasks can be emitted
func (f EcsFormat) EmitContainers(input *transform.BasePodData) ([]byte, error) {
	output := &EcsFormat{Family: input.Name}
	containers := EcsContainers{}

	volumesMap := map[string]string{}

	for _, container := range input.Containers {
		EcsContainer := EcsContainer{}
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

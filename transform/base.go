package transform

import (
	"io"
	"strings"
)

// Logging is an intermediate representation for logging information
type Logging struct {
	Driver  string
	Options map[string]string
}

// PortMapping is an intermediate representation for port mapping information
type PortMapping struct {
	HostIP        string
	HostPort      int
	ContainerIP   string
	ContainerPort int
	Protocol      string
	Name          string
}

// PortMappings is a composite type for slices of PortMapping
type PortMappings []PortMapping

func (pm PortMappings) Len() int           { return len(pm) }
func (pm PortMappings) Swap(i, j int)      { pm[i], pm[j] = pm[j], pm[i] }
func (pm PortMappings) Less(i, j int) bool { return pm[i].ContainerPort < pm[j].ContainerPort }

// IntermediateVolume is an intermediate representation for volume information
type IntermediateVolume struct {
	Host         string
	Container    string
	SourceVolume string
	ReadOnly     bool
}

// IntermediateVolumes is a composite type for slices of IntermediateVolume
type IntermediateVolumes []IntermediateVolume

func (iv IntermediateVolumes) Len() int      { return len(iv) }
func (iv IntermediateVolumes) Swap(i, j int) { iv[i], iv[j] = iv[j], iv[i] }
func (iv IntermediateVolumes) Less(i, j int) bool {
	return strings.Compare(iv[i].Container, iv[j].Container) < 0
}

// Fetch is an intermediate representation for fetching information
type Fetch struct {
	URI string
}

// HealthCheck is an intermediate representation for health check information
type HealthCheck struct {
	Exec string

	HTTPPath string
	Port     int
	Host     string
	Scheme   string
	Headers  map[string]string

	Interval         int
	Timeout          int
	FailureThreshold int
}

// BuildContext is an intermediary representation for build information
type BuildContext struct {
	Context    string
	Dockerfile string
	Args       map[string]string
}

// BaseContainerFormat represents the intermediate format in between input and output formats
type BaseContainerFormat struct {
	Build           *BuildContext
	Command         string
	CPU             int // out of 1024
	DNS             []string
	Domain          []string
	Entrypoint      string
	EnvFile         []string
	Environment     map[string]string
	Essential       bool
	Expose          []int
	Fetch           []*Fetch       // TODO make a struct
	HealthChecks    []*HealthCheck // TODO make a struct
	Hostname        string
	Image           string
	Labels          map[string]string
	Links           []string
	Logging         *Logging
	Memory          int // in bytes
	Name            string
	Network         []string
	NetworkMode     string
	Pid             string
	PortMappings    *PortMappings
	Privileged      bool
	PullImagePolicy string
	Replicas        int
	StopSignal      string
	User            string
	Volumes         *IntermediateVolumes
	VolumesFrom     []string // todo make a struct
	WorkDir         string
}

// BasePodData is the intermediary between each container format
type BasePodData struct {
	Name         string
	Containers   []*BaseContainerFormat
	GlobalLabels map[string]string
	HostNetwork  bool
	HostPID      bool
	Replicas     int
}

// InputFormat is an interface for other container formats to ingest containers
type InputFormat interface {
	IngestContainers(input io.ReadCloser) (*BasePodData, error)
}

// OutputFormat is an interface for other container formats to emit containers
type OutputFormat interface {
	EmitContainers(input *BasePodData) ([]byte, error)
}

package transform

import (
	"io"
	"strings"
)

type Logging struct {
	Driver  string
	Options map[string]string
}

type PortMapping struct {
	HostIP        string
	HostPort      int
	ContainerIP   string
	ContainerPort int
	Protocol      string
	Name          string
}

type PortMappings []PortMapping

func (pm PortMappings) Len() int           { return len(pm) }
func (pm PortMappings) Swap(i, j int)      { pm[i], pm[j] = pm[j], pm[i] }
func (pm PortMappings) Less(i, j int) bool { return pm[i].ContainerPort < pm[j].ContainerPort }

type IntermediateVolume struct {
	Host         string
	Container    string
	SourceVolume string
	ReadOnly     bool
}

type IntermediateVolumes []IntermediateVolume

func (iv IntermediateVolumes) Len() int      { return len(iv) }
func (iv IntermediateVolumes) Swap(i, j int) { iv[i], iv[j] = iv[j], iv[i] }
func (iv IntermediateVolumes) Less(i, j int) bool {
	return strings.Compare(iv[i].Container, iv[j].Container) < 0
}

type Fetch struct {
	URI string
}

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

func (bc *BuildContext) UnmarshalYAML(unmarshal func(interface{}) error) error {
	err := unmarshal(*bc)
	if err != nil {
		var ctx string
		err = unmarshal(&ctx)
		if err != nil {
			return err
		}
		bc.Context = ctx
	}
	return nil
}

type BuildContext struct {
	Context    string            `yaml:"context"`
	Dockerfile string            `yaml:"dockerfile"`
	Args       map[string]string `yaml:"args"`
}

// BaseFormat represents the intermediate format in between input and output formats
type BaseContainerFormat struct {
	Build           *BuildContext
	Command         string
	CPU             int         // out of 1024
	DNS             []string
	Domain          []string
	Entrypoint      string
	EnvFile         []string
	Environment     map[string]string
	Essential       bool
	Expose          []int
	Fetch           []*Fetch
	HealthChecks    []*HealthCheck
	Hostname        string
	Image           string
	Labels          map[string]string
	Links           []string
	Logging         *Logging
	Memory          int         // in bytes
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

type HostVolume struct {
}

type BasePodData struct {
	Name         string
	Containers   []*BaseContainerFormat
	GlobalLabels map[string]string
	HostNetwork  bool
	HostPID      bool
	Replicas     int
	Volumes      []HostVolume
	//Networks
}

type InputFormat interface {
	IngestContainers(input io.ReadCloser) (*BasePodData, error)
}

type OutputFormat interface {
	EmitContainers(input *BasePodData) ([]byte, error)
}

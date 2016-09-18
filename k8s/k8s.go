package k8s

import (
	"io"
	"io/ioutil"
	"path"
	"sort"
	"strings"

	"github.com/ghodss/yaml"
	"github.com/micahhausler/container-tx/transform"
	"k8s.io/kubernetes/pkg/api"
)

func ingestEnvironment(k8sEnv []api.EnvVar) map[string]string {
	response := map[string]string{}
	for _, envVar := range k8sEnv {
		response[envVar.Name] = envVar.Value
	}
	return response
}

func ingestHealthCheck(probe api.Probe) []*transform.HealthCheck {
	hc := transform.HealthCheck{}
	if probe.Exec != nil {
		hc.Exec = strings.Join(probe.Exec.Command, " ")
	}
	if probe.HTTPGet != nil {
		hc.HTTPPath = probe.HTTPGet.Path
		hc.Port = int(probe.HTTPGet.Port.IntVal)
		hc.Host = probe.HTTPGet.Host
		hc.Scheme = string(probe.HTTPGet.Scheme)
		headerMap := map[string]string{}
		for _, header := range probe.HTTPGet.HTTPHeaders {
			headerMap[header.Name] = header.Value
		}
		hc.Headers = headerMap
		hc.Interval = int(probe.PeriodSeconds)
		hc.Timeout = int(probe.TimeoutSeconds)
		hc.FailureThreshold = int(probe.FailureThreshold)
	}
	hcs := []*transform.HealthCheck{&hc}
	return hcs
}

func ingestPortMappings(pms []api.ContainerPort) *transform.PortMappings {
	response := transform.PortMappings{}
	for _, port := range pms {
		response = append(response, transform.PortMapping{
			HostIP:        port.HostIP,
			HostPort:      int(port.HostPort),
			ContainerPort: int(port.ContainerPort),
			Protocol:      string(port.Protocol),
			Name:          port.Name,
		})
	}
	return &response
}

func ingestVolumes(vols []api.VolumeMount, podVolumes []api.Volume) *transform.IntermediateVolumes {
	response := transform.IntermediateVolumes{}

	// key = name, value = path
	hostVolMap := map[string]string{}
	for _, vol := range podVolumes {
		hostVolMap[vol.Name] = vol.HostPath.Path
	}

	for _, vol := range vols {
		hostPath := hostVolMap[vol.Name]
		if len(vol.SubPath) > 0 {
			hostPath = path.Join(hostPath, vol.SubPath)
		}
		response = append(response, transform.IntermediateVolume{
			Container: vol.MountPath,
			Host:      hostPath,
			ReadOnly:  vol.ReadOnly,
		})
	}
	return &response
}

// Pod implements InputFormat and OutputFormat
type Pod struct {
	api.Pod
}

// IngestContainers satisfies InputFormat so k8s pods can be ingested
func (p Pod) IngestContainers(input io.ReadCloser) (*transform.PodData, error) {
	body, err := ioutil.ReadAll(input)
	defer input.Close()
	if err != nil && err != io.EOF {
		return nil, err
	}
	err = yaml.Unmarshal(body, &p)
	if err != nil {
		return nil, err
	}

	outputPod := transform.PodData{Name: p.Name}

	containers := transform.Containers{}

	for _, c := range p.Spec.Containers {
		ir := transform.Container{}
		ir.Command = strings.Join(c.Args, " ")
		// TODO resources/CPU
		// ir.CPU
		ir.Entrypoint = strings.Join(c.Command, " ")
		ir.Environment = ingestEnvironment(c.Env)
		if c.LivenessProbe != nil {
			ir.HealthChecks = ingestHealthCheck(*c.LivenessProbe)
		}
		ir.Hostname = p.Spec.Hostname
		ir.Image = c.Image
		// TODO resources/Memory
		ir.Name = c.Name
		// TODO Network
		// TODO Pid
		ir.PortMappings = ingestPortMappings(c.Ports)
		if c.SecurityContext != nil {
			ir.Privileged = *c.SecurityContext.Privileged
		}
		ir.PullImagePolicy = string(c.ImagePullPolicy)
		ir.Volumes = ingestVolumes(c.VolumeMounts, p.Spec.Volumes)
		ir.WorkDir = c.WorkingDir
		containers = append(containers, ir)
	}

	sort.Sort(containers)
	outputPod.Containers = &containers

	return &outputPod, nil
}

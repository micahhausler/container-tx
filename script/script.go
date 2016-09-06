package script

import (
	"bytes"
	"log"
	"strconv"
	"strings"
	"text/template"

	"github.com/micahhausler/container-tx/transform"
)

func stringifyPortMapping(mapping transform.PortMapping) string {
	response := ""
	if len(mapping.HostIP) > 0 {
		response += mapping.HostIP
		response += ":"
	}
	if mapping.HostPort > 0 {
		response += strconv.Itoa(mapping.HostPort)
	}
	if mapping.HostPort > 0 && mapping.ContainerPort > 0 {
		response += ":"
	}
	if mapping.ContainerPort > 0 {
		response += strconv.Itoa(mapping.ContainerPort)
	}
	if strings.Compare(strings.ToLower(mapping.Protocol), "udp") == 0 {
		response += "/udp"
	}
	return response
}

func stringifyVolume(volume transform.IntermediateVolume) string {
	readOnly := ""
	if volume.ReadOnly {
		readOnly = "ro"
	}
	volStr := []string{volume.Host, volume.Container, readOnly}
	return strings.Trim(strings.Join(volStr, ":"), ":")
}

// Script represents a list of docker container run commands.
// It implements OutputFormat
type Script struct{}

// EmitContainers satisfies OutputFormat so ECS tasks can be emitted
func (s Script) EmitContainers(input *transform.PodData) ([]byte, error) {

	funcMap := template.FuncMap{
		"stringifyPort":   stringifyPortMapping,
		"stringifyVolume": stringifyVolume,
	}

	t := template.Must(template.New("container").Funcs(funcMap).Parse(dockerRunTemplate))

	var buffer bytes.Buffer
	for _, c := range *input.Containers {
		err := t.Execute(&buffer, c)
		if err != nil {
			log.Println("Error executing template:", err)
		}
	}

	return buffer.Bytes(), nil
}

package script

const dockerRunTemplate = `######## {{ .Name }} ########
docker run \
    {{ if .CPU }}--cpu-shares={{.CPU}} \
    {{end -}}
    {{ range .DNS -}}
    --dns {{.}} \
    {{end -}}
    {{ range .Domain -}}
    --dns-search {{.}} \
    {{end -}}
    {{ if .Entrypoint }}--entrypoint={{.Entrypoint}} \
    {{end -}}
    {{ range .EnvFile -}}
    --env-file {{.}} \
    {{end -}}
    {{ range $key, $value := .Environment -}}
    --env {{$key}}={{$value}} \
    {{end -}}
    {{ range .Expose -}}
    --expose {{.}} \
    {{end -}}
    {{ if .Hostname }}--hostname={{.Hostname}} \
    {{end -}}
    {{ range $key, $value := .Labels -}}
    --label {{$key}}={{$value}} \
    {{end -}}
    {{ range .Links -}}
    --link {{.}} \
    {{end -}}
    {{ if .Logging  }}--log-driver {{.Logging.Driver}} \
    {{ range $key, $value := .Logging.Options -}}
    --log-opt {{$key}}={{$value}} \
    {{end -}}
    {{end -}}
    {{ if .Memory  }}--memory={{.Memory}}b \
    {{end -}}
    {{ if .Name }}--name {{.Name}} \
    {{end -}}
    {{ range .Network -}}
    --net-alias {{.}} \
    {{end -}}
    {{ if .NetworkMode }}--net {{.NetworkMode}} \
    {{end -}}
    {{ if .Pid }}--pid {{.Pid}} \
    {{end -}}
    {{if .PortMappings}}{{ range .PortMappings -}}
    --publish {{ stringifyPort . }} \
    {{end}}{{end -}}
    {{ if .Privileged }}--privileged \
    {{end -}}
    {{ if .StopSignal }}--stop-signal={{.StopSignal}} \
    {{end -}}
    {{ if .User }}--user={{.User}} \
    {{end -}}
    {{ if .Volumes }}{{ range .Volumes -}}
    --volume {{ stringifyVolume . }} \
    {{end}}{{end -}}
    {{ range .VolumesFrom -}}
    --volumes-from {{ . }} \
    {{end -}}
    {{ if .WorkDir }}--workdir={{.WorkDir}} \
    {{end -}}
    {{.Image }} {{- with .Command }} \
        {{.}}
{{- end }}
`

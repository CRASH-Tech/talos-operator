{{- define "machinesecrets" -}}
ca: {{ .cluster.secrets.talosconfig.ca }}
crt: {{ .cluster.secrets.talosconfig.crt }}
key: {{ .cluster.secrets.talosconfig.key }}
{{- end -}}

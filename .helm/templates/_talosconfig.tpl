{{- define "talosconfig" -}}
context: {{ .cluster.name }}
contexts:
  {{ .cluster.name }}:
    endpoints:
      - 127.0.0.1
    ca: {{ .cluster.secrets.talosconfig.ca }}
    crt: {{ .cluster.secrets.talosconfig.crt }}
    key: {{ .cluster.secrets.talosconfig.key }}
{{- end -}}

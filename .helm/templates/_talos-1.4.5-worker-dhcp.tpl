{{- define "talos-1.4.5-worker-dhcp" -}}
version: v1alpha1
debug: false
persist: true
machine:
  type: worker
  token: {{ .cluster.secrets.machine.token }}
  ca:
    crt: {{ .cluster.secrets.machine.ca.crt }}
    key: {{ .cluster.secrets.machine.ca.key }}
  certSANs: []
  kubelet:
    image: {{ index .cluster.images "kubelet" }}
    defaultRuntimeSeccompProfileEnabled: true
    disableManifestsDirectory: true
  network: {}
  install:
    disk: /dev/sda
    image: {{ index .cluster.images "installer" }}
    bootloader: true
    wipe: false
  registries: {}
  features:
    rbac: true
    stableHostname: true
    apidCheckExtKeyUsage: true
cluster:
  id: {{ .cluster.secrets.cluster.id }}
  secret: {{ .cluster.secrets.cluster.secret }}
  controlPlane:
    endpoint: {{ .cluster.endpoint }}
  network:
    dnsDomain: {{ .cluster.network.dnsDomain }}
    podSubnets: {{ .cluster.network.podSubnets | toYaml | nindent 6 }}
    serviceSubnets: {{ .cluster.network.serviceSubnets | toYaml | nindent 6 }}
  token: {{ .cluster.secrets.cluster.token }}
  ca:
    crt: {{ .cluster.secrets.cluster.ca.crt }}
    key: {{ .cluster.secrets.cluster.ca.key }}
  discovery:
    enabled: true
    registries:
      kubernetes:
        disabled: true
      service: {}
{{- end }}

{{- define "talos-1.4.5-cp-static" -}}
version: v1alpha1
debug: false
persist: true
machine:
  type: controlplane
  token: {{ .cluster.secrets.machine.token }}
  ca:
    crt: {{ .cluster.secrets.machine.ca.crt }}
    key: {{ .cluster.secrets.machine.ca.key }}
  certSANs: []
  kubelet:
    image: {{ index .cluster.images "kubelet" }}
    defaultRuntimeSeccompProfileEnabled: true
    disableManifestsDirectory: true
  network:
    hostname: {{ .host }}
    interfaces:
      - interface: eth0
        addresses:
          - {{ .hostData.ip }}
        routes:
          - network: 0.0.0.0/0
            gateway: {{ .hostData.gateway }}
        dhcp: false
        vip:
          ip: {{ .hostData.vip }}
    nameservers: {{ .cluster.network.nameservers | toYaml | nindent 6 }}
  install:
    disk: /dev/sda
    image: {{ index .cluster.images "installer" }}
    extensions:
      - image: ghcr.io/siderolabs/iscsi-tools:v0.1.4
    bootloader: true
    wipe: false
  features:
    rbac: true
    stableHostname: true
    apidCheckExtKeyUsage: true
cluster:
  id: {{ .cluster.secrets.cluster.id }}
  secret: {{ .cluster.secrets.cluster.secret }}
  controlPlane:
    endpoint: {{ .cluster.endpoint }}
  clusterName: {{ .cluster.name }}
  network:
    cni:
      name: {{ .cluster.network.cni }}
    dnsDomain: {{ .cluster.network.dnsDomain }}
    podSubnets: {{ .cluster.network.podSubnets | toYaml | nindent 6 }}
    serviceSubnets: {{ .cluster.network.serviceSubnets | toYaml | nindent 6 }}
  token: {{ .cluster.secrets.cluster.token }}
  secretboxEncryptionSecret: {{ .cluster.secrets.cluster.secretboxEncryptionSecret }}
  ca:
    crt: {{ .cluster.secrets.cluster.ca.crt }}
    key: {{ .cluster.secrets.cluster.ca.key }}
  aggregatorCA:
    crt: {{ .cluster.secrets.cluster.aggregatorCA.crt }}
    key: {{ .cluster.secrets.cluster.aggregatorCA.key }}
  serviceAccount:
    key: {{ .cluster.secrets.cluster.serviceAccount.key }}
  apiServer:
    image: {{ index .cluster.images "api-server" }}
    certSANs:
      - 10.171.120.200
    disablePodSecurityPolicy: {{ .cluster.disablePodSecurityPolicy }}
    admissionControl:
      - name: PodSecurity
        configuration:
          apiVersion: pod-security.admission.config.k8s.io/v1alpha1
          defaults:
            audit: privileged
            audit-version: latest
            enforce: privileged
            enforce-version: latest
            warn: privileged
            warn-version: latest
          exemptions:
            namespaces:
              - kube-system
            runtimeClasses: []
            usernames: []
          kind: PodSecurityConfiguration
    auditPolicy:
      apiVersion: audit.k8s.io/v1
      kind: Policy
      rules:
        - level: Metadata
  controllerManager:
    image: {{ index .cluster.images "kube-controller-manager" }}
  proxy:
    disabled: true
    image: {{ index .cluster.images "kube-proxy" }}
    extraArgs:
      metrics-bind-address: 0.0.0.0:10249
  scheduler:
    image: {{ index .cluster.images "kube-scheduler" }}
  discovery:
    enabled: true
    registries:
      kubernetes:
        disabled: true
      service: {}
  etcd:
    ca:
      crt: {{ .cluster.secrets.cluster.etcd.ca.crt }}
      key: {{ .cluster.secrets.cluster.etcd.ca.key }}
    extraArgs:
      listen-metrics-urls: http://0.0.0.0:2381
  inlineManifests:
    - name: admin-sa
      contents: |-
        apiVersion: v1
        kind: ServiceAccount
        metadata:
          name: admin-user
          namespace: kube-system
    - name: admin-sa-secret
      contents: |-
        apiVersion: v1
        kind: Secret
        metadata:
          name: admin-user
          namespace: kube-system
          annotations:
          kubernetes.io/service-account.name: admin-user
        type: kubernetes.io/service-account-token
    - name: admin-sa-crb
      contents: |-
        apiVersion: rbac.authorization.k8s.io/v1
        kind: ClusterRoleBinding
        metadata:
          name: admin-user
        roleRef:
          apiGroup: rbac.authorization.k8s.io
          kind: ClusterRole
          name: cluster-admin
        subjects:
          - kind: ServiceAccount
          name: admin-user
          namespace: kube-system
  adminKubeconfig:
    certLifetime: 87600h0m0s
  allowSchedulingOnControlPlanes: {{ .cluster.allowSchedulingOnControlPlanes }}
{{- end }}
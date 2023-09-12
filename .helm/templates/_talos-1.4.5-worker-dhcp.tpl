{{- define "talos-1.4.5-worker-dhcp" -}}
version: v1alpha1 # Indicates the schema used to decode the contents.
debug: false # Enable verbose logging to the console.
persist: true # Indicates whether to pull the machine config upon every boot.
# Provides machine specific configuration options.
machine:
    type: worker # Defines the role of the machine within the cluster.
    token: kubvxk.3o87hrn0nwq4nnp2 # The `token` is used by a machine to join the PKI of the cluster.
    # The root certificate authority of the PKI.
    ca:
        crt: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUJQekNCOHFBREFnRUNBaEVBL1ZrOGZyOXhidlByREFpS0pZazVnakFGQmdNclpYQXdFREVPTUF3R0ExVUUKQ2hNRmRHRnNiM013SGhjTk1qTXdOakF6TVRNeE9ESXpXaGNOTXpNd05UTXhNVE14T0RJeldqQVFNUTR3REFZRApWUVFLRXdWMFlXeHZjekFxTUFVR0F5dGxjQU1oQUZrUHpJdHFtRS9jdElkbURNVjRUNDI5REszVEdTeGRhTGRjCkhHajBES1RnbzJFd1h6QU9CZ05WSFE4QkFmOEVCQU1DQW9Rd0hRWURWUjBsQkJZd0ZBWUlLd1lCQlFVSEF3RUcKQ0NzR0FRVUZCd01DTUE4R0ExVWRFd0VCL3dRRk1BTUJBZjh3SFFZRFZSME9CQllFRk5QcEd3RTlTWG5wSWZOSQpEdW5FM0J0ZmU4Q3BNQVVHQXl0bGNBTkJBTEhXUVk2UDkzejQ3YlFKYStHejZpMDUwTDRCVGNNcjRSTGgrY3B0CmYwZGdDVmJONmtMM0pEWUJBQmdwL0ppcmQ1Rm9oQ2hLNFZQZnlPVEE3bWI3eGd3PQotLS0tLUVORCBDRVJUSUZJQ0FURS0tLS0tCg==
        key: ""
    # Extra certificate subject alternative names for the machine's certificate.
    certSANs: []
    # Used to provide additional options to the kubelet.
    kubelet:
        image: ghcr.io/siderolabs/kubelet:v1.27.2 # The `image` field is an optional reference to an alternative kubelet image.
        defaultRuntimeSeccompProfileEnabled: true # Enable container runtime default Seccomp profile.
        disableManifestsDirectory: true # The `disableManifestsDirectory` field configures the kubelet to get static pod manifests from the /etc/kubernetes/manifests directory.
    # Provides machine specific network configuration options.
    network: {}
    # Used to provide instructions for installations.
    install:
        disk: /dev/sda # The disk used for installations.
        image: ghcr.io/siderolabs/installer:v1.4.5 # Allows for supplying the image used to perform the installation.
        bootloader: true # Indicates if a bootloader should be installed.
        wipe: false # Indicates if the installation disk should be wiped at installation time.
    # Used to configure the machine's container image registry mirrors.
    registries: {}
    # Features describe individual Talos features that can be switched on or off.
    features:
        rbac: true # Enable role-based access control (RBAC).
        stableHostname: true # Enable stable default hostname.
        apidCheckExtKeyUsage: true # Enable checks for extended key usage of client certificates in apid.
# Provides cluster specific configuration options.
cluster:
    id: cHUoVX6jKYU0c7Dn_uHJJyqsgYlvclR0qJ7oNZcuXKE= # Globally unique identifier for this cluster (base64 encoded random 32 bytes).
    secret: CY890tDI6Jk7eLkrwa6qbMm7GLsHeBbcVZgbBirAC6Y= # Shared secret of cluster (base64 encoded random 32 bytes).
    # Provides control plane specific configuration options.
    controlPlane:
        endpoint: https://10.171.120.200:6443 # Endpoint is the canonical controlplane endpoint, which can be an IP address or a DNS hostname.
    # Provides cluster specific network configuration options.
    network:
        dnsDomain: cluster.local # The domain used by Kubernetes DNS.
        # The pod subnet CIDR.
        podSubnets:
            - 10.244.0.0/16
        # The service subnet CIDR.
        serviceSubnets:
            - 10.96.0.0/12
    token: 2p0i1d.5rzx16axk69iqvtv # The [bootstrap token](https://kubernetes.io/docs/reference/access-authn-authz/bootstrap-tokens/) used to join the cluster.
    # The base64 encoded root certificate authority used by Kubernetes.
    ca:
        crt: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUJpekNDQVRDZ0F3SUJBZ0lSQU91a0ZZUzg4eWdjZ1hmSFRNNldGb0F3Q2dZSUtvWkl6ajBFQXdJd0ZURVQKTUJFR0ExVUVDaE1LYTNWaVpYSnVaWFJsY3pBZUZ3MHlNekEyTURNeE16RTRNak5hRncwek16QTFNekV4TXpFNApNak5hTUJVeEV6QVJCZ05WQkFvVENtdDFZbVZ5Ym1WMFpYTXdXVEFUQmdjcWhrak9QUUlCQmdncWhrak9QUU1CCkJ3TkNBQVNWNnlTTkZpVWR2M0VBMU4yNE9uVkx2YmNMVStBV2RRTkM4a1FVVUVRbkgzSzVhdmtBNFJwc3VzSSsKVUwrYlJ4RlNGYmVWaEZPRERDNkFVK3FyV3gwWm8yRXdYekFPQmdOVkhROEJBZjhFQkFNQ0FvUXdIUVlEVlIwbApCQll3RkFZSUt3WUJCUVVIQXdFR0NDc0dBUVVGQndNQ01BOEdBMVVkRXdFQi93UUZNQU1CQWY4d0hRWURWUjBPCkJCWUVGSzcvNnZDc1o5eXYvaWFRNUNEKytlTWpUWVFHTUFvR0NDcUdTTTQ5QkFNQ0Ewa0FNRVlDSVFDVGpGeGwKVGRUbGZ2TXFpUlYxMGtjdEhFVVVnQkJSSkF6ZGRvbXorMit6dlFJaEFJNk1YQ1FNYTFLYk8rVFl0WWIwMXJTRwpZaTU2bmtRajVoNnZzQitwaXNqTAotLS0tLUVORCBDRVJUSUZJQ0FURS0tLS0tCg==
        key: ""
    # Configures cluster member discovery.
    discovery:
        enabled: true # Enable the cluster membership discovery feature.
        # Configure registries used for cluster member discovery.
        registries:
            # Kubernetes registry uses Kubernetes API server to discover cluster members and stores additional information
            kubernetes:
                disabled: true # Disable Kubernetes discovery registry.
            # Service registry is using an external service to push and pull information about cluster members.
            service: {}
{{- end }}

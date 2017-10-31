package api

type Cluster struct {
	Name   string `json:"name"` // This will be in ObjectMeta
	//metav1.ObjectMeta
	Spec   ClusterSpec `json:"spec"`
	//Status ClusterStatus
}


type ClusterSpec struct {
	Cloud        string `json:"cloud"` // e.g. aws, azure, gcp
	Project      string `json:"project"`
	SSH          *SSHConfig `json:"ssh"`
	KubernetesVersion KubernetesVersionInfo `json:"kubernetesVersion"`
}

type SSHConfig struct {
	//Name          string // e.g. id_rsa
	PublicKeyPath string `json:"publicKeyPath"`
	User      string `json:"user"`
}

type ControlPlaneConfig struct {
	APIServer         APIServerConfig
	ControllerManager ControllerManagerConfig
}

type APIServerConfig struct {
	AdvertiseAddress  string // e.g. "0.0.0.0"
	Port              uint32
	CertExtraSANs     string
	AuthorizationMode string
}

type ControllerManagerConfig struct {
	LeaderElection    bool
	Port              uint32
}

type EtcdConfig struct {
	Version string // full semver
	API     string // e.g. v2, v3
}

type KubernetesVersionInfo struct {
	// Semantic version of Kubernetes to run.
	Version string
}
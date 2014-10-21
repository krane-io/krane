package runconfig

import (
	dockerEngine "github.com/docker/docker/engine"
	dockerNat "github.com/docker/docker/nat"
	dockerUtils "github.com/docker/docker/utils"
)

type NetworkMode string

type DeviceMapping struct {
	PathOnHost        string
	PathInContainer   string
	CgroupPermissions string
}

type RestartPolicy struct {
	Name              string
	MaximumRetryCount int
}

type Config struct {
	Hostname        string
	Domainname      string
	User            string
	Memory          int64  // Memory limit (in bytes)
	MemorySwap      int64  // Total memory usage (memory + swap); set `-1' to disable swap
	CpuShares       int64  // CPU shares (relative weight vs. other containers)
	Cpuset          string // Cpuset 0-2, 0,1
	AttachStdin     bool
	AttachStdout    bool
	AttachStderr    bool
	PortSpecs       []string // Deprecated - Can be in the format of 8080/tcp
	ExposedPorts    map[dockerNat.Port]struct{}
	Tty             bool // Attach standard streams to a tty, including stdin if it is not closed.
	OpenStdin       bool // Open stdin
	StdinOnce       bool // If true, close stdin after the 1 attached client disconnects.
	Env             []string
	Cmd             []string
	Image           string // Name of the image as it was passed by the operator (eg. could be symbolic)
	Volumes         map[string]struct{}
	WorkingDir      string
	Entrypoint      []string
	NetworkDisabled bool
	OnBuild         []string
	SecurityOpt     []string
	Ship            string
}

type HostConfig struct {
	Binds           []string
	ContainerIDFile string
	LxcConf         []dockerUtils.KeyValuePair
	Privileged      bool
	PortBindings    dockerNat.PortMap
	Links           []string
	PublishAllPorts bool
	Dns             []string
	DnsSearch       []string
	ExtraHosts      []string
	VolumesFrom     []string
	Devices         []DeviceMapping
	NetworkMode     NetworkMode
	CapAdd          []string
	CapDrop         []string
	RestartPolicy   RestartPolicy
}

func ContainerHostConfigFromJob(job *dockerEngine.Job) *HostConfig {
	if job.EnvExists("HostConfig") {
		hostConfig := HostConfig{}
		job.GetenvJson("HostConfig", &hostConfig)
		return &hostConfig
	}

	hostConfig := &HostConfig{
		ContainerIDFile: job.Getenv("ContainerIDFile"),
		Privileged:      job.GetenvBool("Privileged"),
		PublishAllPorts: job.GetenvBool("PublishAllPorts"),
		NetworkMode:     NetworkMode(job.Getenv("NetworkMode")),
	}

	job.GetenvJson("LxcConf", &hostConfig.LxcConf)
	job.GetenvJson("PortBindings", &hostConfig.PortBindings)
	job.GetenvJson("Devices", &hostConfig.Devices)
	job.GetenvJson("RestartPolicy", &hostConfig.RestartPolicy)
	if Binds := job.GetenvList("Binds"); Binds != nil {
		hostConfig.Binds = Binds
	}
	if Links := job.GetenvList("Links"); Links != nil {
		hostConfig.Links = Links
	}
	if Dns := job.GetenvList("Dns"); Dns != nil {
		hostConfig.Dns = Dns
	}
	if DnsSearch := job.GetenvList("DnsSearch"); DnsSearch != nil {
		hostConfig.DnsSearch = DnsSearch
	}
	if ExtraHosts := job.GetenvList("ExtraHosts"); ExtraHosts != nil {
		hostConfig.ExtraHosts = ExtraHosts
	}
	if VolumesFrom := job.GetenvList("VolumesFrom"); VolumesFrom != nil {
		hostConfig.VolumesFrom = VolumesFrom
	}
	if CapAdd := job.GetenvList("CapAdd"); CapAdd != nil {
		hostConfig.CapAdd = CapAdd
	}
	if CapDrop := job.GetenvList("CapDrop"); CapDrop != nil {
		hostConfig.CapDrop = CapDrop
	}

	return hostConfig
}

func ContainerConfigFromJob(job *dockerEngine.Job) *Config {
	config := &Config{
		Hostname:        job.Getenv("Hostname"),
		Domainname:      job.Getenv("Domainname"),
		User:            job.Getenv("User"),
		Memory:          job.GetenvInt64("Memory"),
		MemorySwap:      job.GetenvInt64("MemorySwap"),
		CpuShares:       job.GetenvInt64("CpuShares"),
		Cpuset:          job.Getenv("Cpuset"),
		AttachStdin:     job.GetenvBool("AttachStdin"),
		AttachStdout:    job.GetenvBool("AttachStdout"),
		AttachStderr:    job.GetenvBool("AttachStderr"),
		Tty:             job.GetenvBool("Tty"),
		OpenStdin:       job.GetenvBool("OpenStdin"),
		StdinOnce:       job.GetenvBool("StdinOnce"),
		Image:           job.Getenv("Image"),
		WorkingDir:      job.Getenv("WorkingDir"),
		NetworkDisabled: job.GetenvBool("NetworkDisabled"),
		Ship:            job.Getenv("Ship"),
	}
	job.GetenvJson("ExposedPorts", &config.ExposedPorts)
	job.GetenvJson("Volumes", &config.Volumes)
	config.SecurityOpt = job.GetenvList("SecurityOpt")
	if PortSpecs := job.GetenvList("PortSpecs"); PortSpecs != nil {
		config.PortSpecs = PortSpecs
	}
	if Env := job.GetenvList("Env"); Env != nil {
		config.Env = Env
	}
	if Cmd := job.GetenvList("Cmd"); Cmd != nil {
		config.Cmd = Cmd
	}
	if Entrypoint := job.GetenvList("Entrypoint"); Entrypoint != nil {
		config.Entrypoint = Entrypoint
	}
	return config
}

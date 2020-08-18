package config

// JenkinsJobs is the list of all (whitelisted) Jenkins jobs
type JenkinsJobs map[string]JobConfig

// Jenkins is the main Jenkins config, including credentials and the whitelisted jobs
type Jenkins struct {
	Host     string
	Username string
	Password string
	Jobs     JenkinsJobs
}

// IsEnabled checks if a host was defined...by default it's not set
func (c Jenkins) IsEnabled() bool {
	return c.Host != ""
}

// JobConfig concrete job configuration -> only defined jobs are (re)startable
type JobConfig struct {
	Parameters []JobParameter
	Trigger    string
	OnStart    []string
	OnSuccess  []string
	OnFailure  []string
}

// JobParameter are defined build parameters per job
type JobParameter struct {
	Name    string
	Default string
	Type    string
}

package config

import (
	"regexp"

	"gopkg.in/yaml.v2"
)

const (
	DefaultWorkingDir       = "tmp"
	defaultGithubDomain     = "github.com"
	defaultGithubEnvvarName = "GITHUB_TOKEN"
	defaultRegex            = ".*"
	TargetRepo              = "repo"
	TargetSrc               = "src"
)

type Config struct {
	WorkingDir    string         `yaml:"working_dir"`
	GithubConfigs []GithubConfig `yaml:"github_configs"`
}

type GithubConfig struct {
	domain          string       `yaml:"domain,omitempty"`
	ApiEndpoint     string       `yaml:"api_endpoint,omitempty"`
	UploadEndpoint  string       `yaml:"upload_endpoint,omitempty"`
	tokenEnvvarName string       `yaml:"token_envvar_name"`
	UserRepoConfigs []RepoConfig `yaml:"user_repo_configs"`
	OrgRepoConfigs  []RepoConfig `yaml:"org_repo_configs"`
}

func (c GithubConfig) Domain() string {
	if len(c.domain) == 0 {
		return defaultGithubDomain
	} else {
		return c.domain
	}
}

func (c GithubConfig) EnvvarName() string {
	if len(c.tokenEnvvarName) == 0 {
		return defaultGithubEnvvarName
	} else {
		return c.tokenEnvvarName
	}
}

type RepoConfig struct {
	Name            string           `yaml:"name"`
	Regex           *Regexp          `yaml:"regex,omitempty"`
	ConftestConfigs []ConftestConfig `yaml:"conftest_configs"`
}

func (c *RepoConfig) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type raw RepoConfig
	d := raw{
		Regex: &Regexp{regexp.MustCompile(defaultRegex), false, false},
	}
	if err := unmarshal(&d); err != nil {
		return err
	}
	*c = RepoConfig(d)
	return nil
}

type Regexp struct {
	*regexp.Regexp
	UsedWithRepo bool
	UsedWithSrc  bool
}

func (r *Regexp) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var s string
	if err := unmarshal(&s); err != nil {
		return err
	}
	regex, err := regexp.Compile(s)
	if err != nil {
		return err
	}
	r.Regexp = regex
	return nil
}

type ConftestConfig struct {
	Target   string   `yaml:"target"`
	Input    string   `yaml:"input"`
	Combine  bool     `yaml:"combine"`
	Policies []string `yaml:"policies"`
}

func Parse(buf []byte) (*Config, error) {
	cfg := Config{
		WorkingDir: DefaultWorkingDir,
	}
	err := yaml.Unmarshal(buf, &cfg)
	// initialize flags in regex
	for _, ghConfig := range cfg.GithubConfigs {
		for _, userRepoConfig := range ghConfig.UserRepoConfigs {
			for _, ConftestConfig := range userRepoConfig.ConftestConfigs {
				if ConftestConfig.Target == TargetRepo {
					userRepoConfig.Regex.UsedWithRepo = true
				}
				if ConftestConfig.Target == TargetSrc {
					userRepoConfig.Regex.UsedWithSrc = true
				}
			}
		}
		for _, orgRepoConfig := range ghConfig.OrgRepoConfigs {
			for _, ConftestConfig := range orgRepoConfig.ConftestConfigs {
				if ConftestConfig.Target == TargetRepo {
					orgRepoConfig.Regex.UsedWithRepo = true
				}
				if ConftestConfig.Target == TargetSrc {
					orgRepoConfig.Regex.UsedWithSrc = true
				}
			}
		}
	}

	if err != nil {
		return nil, err
	}
	return &cfg, nil
}

package config

import (
	"errors"
	"regexp"

	"gopkg.in/yaml.v2"
)

const (
	DefaultWorkingDir = "tmp"
	defaultRegex      = ".*"
)

type Regexp struct {
	*regexp.Regexp
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

type Config struct {
	WorkingDir string `yaml:"working_dir"`
	Users      []User `yaml:"users"`
}

type User struct {
	Name  string  `yaml:"name"`
	Type  string  `yaml:"type,omitempty"`
	Regex *Regexp `yaml:"regex,omitempty"`
}

func (s *User) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type raw User
	d := raw{
		Regex: &Regexp{regexp.MustCompile(defaultRegex)},
	}
	if err := unmarshal(&d); err != nil {
		return err
	}
	*s = User(d)
	return nil
}

func Validate(c *Config) []error {
	var res []error
	for _, users := range c.Users {
		if len(users.Name) == 0 {
			res = append(res, errors.New("config error: require `name` at users"))
		}
	}
	return res
}

func Parse(buf []byte) (*Config, error) {
	cfg := Config{
		WorkingDir: DefaultWorkingDir,
	}
	err := yaml.Unmarshal(buf, &cfg)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}

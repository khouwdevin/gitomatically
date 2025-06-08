package config

type CronSettings struct {
	Cron bool   `yaml:"cron"`
	Spec string `yaml:"spec"`
}

type RepositoryConfig struct {
	Url      string   `yaml:"url"`
	Clone    string   `yaml:"clone"`
	Branch   string   `yaml:"branch"`
	Path     string   `yaml:"path"`
	Commands []string `yaml:"commands"`
}

type Config struct {
	Preference   CronSettings                `yaml:"preference"`
	Repositories map[string]RepositoryConfig `yaml:"repositories"`
}

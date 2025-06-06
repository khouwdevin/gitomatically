package config

type RepositoryConfig struct {
	Url      string   `yaml:"url"`
	Clone    string   `yaml:"clone"`
	Branch   string   `yaml:"branch"`
	Path     string   `yaml:"path"`
	Commands []string `yaml:"commands"`
}

type Config struct {
	Repositories map[string]RepositoryConfig `yaml:"repositories"`
}

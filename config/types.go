package config

type PreferenceSettings struct {
	PrivateKey string `yaml:"private_key"`
	Paraphrase string `yaml:"paraphrase"`
	Cron       bool   `yaml:"cron"`
	Spec       string `yaml:"spec"`
}

type RepositoryConfig struct {
	Url      string   `yaml:"url"`
	Clone    string   `yaml:"clone"`
	Branch   string   `yaml:"branch"`
	Path     string   `yaml:"path"`
	Commands []string `yaml:"commands"`
}

type Config struct {
	Preference   PreferenceSettings          `yaml:"preference"`
	Repositories map[string]RepositoryConfig `yaml:"repositories"`
}

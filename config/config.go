package config

// DefaultConfig provides the default config values that are used when no config file is specified
// Please keep defaults.ini in sync with this so there isn't any confusion
var DefaultConfig = []byte(`
app_mode = production
listen_addr = 0.0.0.0:8080

[chef]
server_url = http://localhost/organizations/example/
username = example
key_file = /path/to/example.pem
ssl_verify = true

[logging]
level = info
output = stdout
format = json
`)

type chefConfig struct {
	ServerURL string `mapstructure:"server_url"`
	Username  string `mapstructure:"username"`
	KeyFile   string `mapstructure:"key_file"`
	SSLVerify bool   `mapstructure:"ssl_verify"`
}

type appConfig struct {
	AppMode    string `mapstructure:"app_mode"`
	ListenAddr string `mapstructure:"listen_addr"`
}

type loggingConfig struct {
	Level  string `mapstructure:"level"`
	Output string `mapstructure:"output"`
	Format string `mapstructure:"format"`
}

type Config struct {
	App     appConfig     `mapstructure:"default"` // default is kind of misleading; it means top-level keys
	Chef    chefConfig    `mapstructure:"chef"`
	Logging loggingConfig `mapstructure:"logging"`
}

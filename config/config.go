package config

// DefaultConfig provides the default config values that are used when no config file is specified
// Please keep defaults.ini in sync with this so there isn't any confusion
var DefaultConfig = []byte(`
app_mode = production
listen_addr = 0.0.0.0:8080
use_mock_data = false

[chef]
server_url = http://localhost/organizations/example/
username = example
key_file = /path/to/example.pem
ssl_verify = true

[logging]
level = info
output = stdout
format = json
request_logging = true
log_health_checks = true

[server]
base_path = /
trusted_proxies =
`)

type chefConfig struct {
	ServerURL string `mapstructure:"server_url"`
	Username  string `mapstructure:"username"`
	KeyFile   string `mapstructure:"key_file"`
	SSLVerify bool   `mapstructure:"ssl_verify"`
}

type appConfig struct {
	AppMode     string `mapstructure:"app_mode"`
	ListenAddr  string `mapstructure:"listen_addr"`
	UseMockData bool   `mapstructure:"use_mock_data"`
}

type loggingConfig struct {
	Level           string `mapstructure:"level"`
	Output          string `mapstructure:"output"`
	Format          string `mapstructure:"format"`
	RequestLogging  bool   `mapstructure:"request_logging"`
	LogHealthChecks bool   `mapstructure:"log_health_checks"`
}

type serverConfig struct {
	BasePath       string `mapstructure:"base_path"`
	EnableGzip     bool   `mapstructure:"enable_gzip"`
	TrustedProxies string `mapstructure:"trusted_proxies"`
}

type customLinksConfig struct {
	Nodes        map[int]customLink `mapstructure:"nodes"`
	Environments map[int]customLink `mapstructure:"environments"` // Unused, but maybe in the future
	Roles        map[int]customLink `mapstructure:"roles"`        // Unused, but maybe in the future
	DataBags     map[int]customLink `mapstructure:"data_bags"`    // Unused, but maybe in the future
}

type customLink struct {
	Title  string `mapstructure:"title"`
	Href   string `mapstructure:"href"`
	NewTab bool   `mapstructure:"new_tab"`
}

type Config struct {
	App         appConfig         `mapstructure:"default"`
	Chef        chefConfig        `mapstructure:"chef"`
	Logging     loggingConfig     `mapstructure:"logging"`
	Server      serverConfig      `mapstructure:"server"`
	CustomLinks customLinksConfig `mapstructure:"custom_links"`
}

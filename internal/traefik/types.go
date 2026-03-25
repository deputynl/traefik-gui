package traefik

// StaticConfig mirrors the relevant fields of traefik's static configuration
// file (traefik.yml / traefik.toml). Only the fields we need to read or write
// are included.
type StaticConfig struct {
	API          *APIConfig                     `yaml:"api,omitempty"          json:"api,omitempty"`
	EntryPoints  map[string]EntryPoint          `yaml:"entryPoints,omitempty"  json:"entryPoints,omitempty"`
	Providers    *Providers                     `yaml:"providers,omitempty"    json:"providers,omitempty"`
	CertResolvers map[string]CertResolver       `yaml:"certificatesResolvers,omitempty" json:"certificatesResolvers,omitempty"`
	Log          *LogConfig                     `yaml:"log,omitempty"          json:"log,omitempty"`
	AccessLog    *AccessLogConfig               `yaml:"accessLog,omitempty"    json:"accessLog,omitempty"`
	Global       *GlobalConfig                  `yaml:"global,omitempty"       json:"global,omitempty"`
}

type APIConfig struct {
	Dashboard bool   `yaml:"dashboard,omitempty" json:"dashboard,omitempty"`
	Insecure  bool   `yaml:"insecure,omitempty"  json:"insecure,omitempty"`
	Debug     bool   `yaml:"debug,omitempty"     json:"debug,omitempty"`
}

type EntryPoint struct {
	Address string          `yaml:"address,omitempty" json:"address,omitempty"`
	HTTP    *EntryPointHTTP `yaml:"http,omitempty"    json:"http,omitempty"`
}

type EntryPointHTTP struct {
	Redirections *Redirections `yaml:"redirections,omitempty" json:"redirections,omitempty"`
	TLS          *EntryPointTLS `yaml:"tls,omitempty"         json:"tls,omitempty"`
}

type Redirections struct {
	EntryPoint *RedirectEntryPoint `yaml:"entryPoint,omitempty" json:"entryPoint,omitempty"`
}

type RedirectEntryPoint struct {
	To     string `yaml:"to,omitempty"     json:"to,omitempty"`
	Scheme string `yaml:"scheme,omitempty" json:"scheme,omitempty"`
}

type EntryPointTLS struct {
	CertResolver string   `yaml:"certResolver,omitempty" json:"certResolver,omitempty"`
	Domains      []Domain `yaml:"domains,omitempty"      json:"domains,omitempty"`
}

type Domain struct {
	Main string   `yaml:"main,omitempty" json:"main,omitempty"`
	SANs []string `yaml:"sans,omitempty" json:"sans,omitempty"`
}

type Providers struct {
	Docker        *DockerProvider `yaml:"docker,omitempty"        json:"docker,omitempty"`
	File          *FileProvider   `yaml:"file,omitempty"          json:"file,omitempty"`
	KubernetesIngress *struct{}   `yaml:"kubernetesIngress,omitempty" json:"kubernetesIngress,omitempty"`
}

type DockerProvider struct {
	Endpoint         string `yaml:"endpoint,omitempty"         json:"endpoint,omitempty"`
	ExposedByDefault bool   `yaml:"exposedByDefault,omitempty" json:"exposedByDefault,omitempty"`
	Network          string `yaml:"network,omitempty"          json:"network,omitempty"`
	Watch            bool   `yaml:"watch,omitempty"            json:"watch,omitempty"`
}

type FileProvider struct {
	Directory string `yaml:"directory,omitempty" json:"directory,omitempty"`
	Filename  string `yaml:"filename,omitempty"  json:"filename,omitempty"`
	Watch     bool   `yaml:"watch,omitempty"     json:"watch,omitempty"`
}

type CertResolver struct {
	ACME *ACMEConfig `yaml:"acme,omitempty" json:"acme,omitempty"`
}

type ACMEConfig struct {
	Email         string        `yaml:"email,omitempty"         json:"email,omitempty"`
	Storage       string        `yaml:"storage,omitempty"       json:"storage,omitempty"`
	CAServer      string        `yaml:"caServer,omitempty"      json:"caServer,omitempty"`
	HTTPChallenge *HTTPChallenge `yaml:"httpChallenge,omitempty" json:"httpChallenge,omitempty"`
	TLSChallenge  *TLSChallenge  `yaml:"tlsChallenge,omitempty"  json:"tlsChallenge,omitempty"`
	DNSChallenge  *DNSChallenge  `yaml:"dnsChallenge,omitempty"  json:"dnsChallenge,omitempty"`
}

type HTTPChallenge struct {
	EntryPoint string `yaml:"entryPoint,omitempty" json:"entryPoint,omitempty"`
}

type TLSChallenge struct{}

type DNSChallenge struct {
	Provider                  string   `yaml:"provider,omitempty"                  json:"provider,omitempty"`
	DelayBeforeCheck          int      `yaml:"delayBeforeCheck,omitempty"          json:"delayBeforeCheck,omitempty"`
	Resolvers                 []string `yaml:"resolvers,omitempty"                 json:"resolvers,omitempty"`
	DisablePropagationCheck   bool     `yaml:"disablePropagationCheck,omitempty"   json:"disablePropagationCheck,omitempty"`
}

type LogConfig struct {
	Level    string `yaml:"level,omitempty"    json:"level,omitempty"`
	FilePath string `yaml:"filePath,omitempty" json:"filePath,omitempty"`
	Format   string `yaml:"format,omitempty"   json:"format,omitempty"`
}

type AccessLogConfig struct {
	FilePath string `yaml:"filePath,omitempty" json:"filePath,omitempty"`
	Format   string `yaml:"format,omitempty"   json:"format,omitempty"`
}

type GlobalConfig struct {
	CheckNewVersion    bool `yaml:"checkNewVersion,omitempty"    json:"checkNewVersion,omitempty"`
	SendAnonymousUsage bool `yaml:"sendAnonymousUsage,omitempty" json:"sendAnonymousUsage,omitempty"`
}

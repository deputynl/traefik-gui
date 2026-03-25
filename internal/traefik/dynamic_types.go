package traefik

// DynamicConfig represents the content of a single Traefik dynamic config file.
type DynamicConfig struct {
	HTTP *HTTPDynamicConfig `yaml:"http,omitempty" json:"http,omitempty"`
}

type HTTPDynamicConfig struct {
	Routers           map[string]DynRouter          `yaml:"routers,omitempty"           json:"routers,omitempty"`
	Services          map[string]DynService         `yaml:"services,omitempty"          json:"services,omitempty"`
	Middlewares       map[string]DynMiddleware       `yaml:"middlewares,omitempty"       json:"middlewares,omitempty"`
	ServersTransports map[string]ServersTransport    `yaml:"serversTransports,omitempty" json:"serversTransports,omitempty"`
}

type DynRouter struct {
	Rule        string     `yaml:"rule,omitempty"        json:"rule,omitempty"`
	EntryPoints []string   `yaml:"entryPoints,omitempty" json:"entryPoints,omitempty"`
	Service     string     `yaml:"service,omitempty"     json:"service,omitempty"`
	TLS         *DynTLS    `yaml:"tls,omitempty"         json:"tls,omitempty"`
	Priority    int        `yaml:"priority,omitempty"    json:"priority,omitempty"`
	Middlewares []string   `yaml:"middlewares,omitempty" json:"middlewares,omitempty"`
}

type DynTLS struct {
	CertResolver string   `yaml:"certResolver,omitempty" json:"certResolver,omitempty"`
	Domains      []Domain `yaml:"domains,omitempty"      json:"domains,omitempty"`
}

type DynService struct {
	LoadBalancer *LoadBalancer `yaml:"loadBalancer,omitempty" json:"loadBalancer,omitempty"`
}

type LoadBalancer struct {
	Servers          []BackendServer `yaml:"servers,omitempty"          json:"servers,omitempty"`
	PassHostHeader   *bool           `yaml:"passHostHeader,omitempty"   json:"passHostHeader,omitempty"`
	ServersTransport string          `yaml:"serversTransport,omitempty" json:"serversTransport,omitempty"`
}

type BackendServer struct {
	URL string `yaml:"url,omitempty" json:"url,omitempty"`
}

type ServersTransport struct {
	InsecureSkipVerify bool `yaml:"insecureSkipVerify,omitempty" json:"insecureSkipVerify,omitempty"`
}

type DynMiddleware struct {
	BasicAuth      *MiddlewareBasicAuth      `yaml:"basicAuth,omitempty"      json:"basicAuth,omitempty"`
	RedirectScheme *MiddlewareRedirectScheme `yaml:"redirectScheme,omitempty" json:"redirectScheme,omitempty"`
	Headers        *MiddlewareHeaders        `yaml:"headers,omitempty"        json:"headers,omitempty"`
	StripPrefix    *MiddlewareStripPrefix    `yaml:"stripPrefix,omitempty"    json:"stripPrefix,omitempty"`
	IPAllowList    *MiddlewareIPAllowList    `yaml:"ipAllowList,omitempty"    json:"ipAllowList,omitempty"`
	ForwardAuth    *MiddlewareForwardAuth    `yaml:"forwardAuth,omitempty"    json:"forwardAuth,omitempty"`
	RateLimit      *MiddlewareRateLimit      `yaml:"rateLimit,omitempty"      json:"rateLimit,omitempty"`
}

type MiddlewareBasicAuth struct {
	Users []string `yaml:"users,omitempty" json:"users,omitempty"`
	Realm string   `yaml:"realm,omitempty" json:"realm,omitempty"`
}

type MiddlewareRedirectScheme struct {
	Scheme    string `yaml:"scheme,omitempty"    json:"scheme,omitempty"`
	Permanent bool   `yaml:"permanent,omitempty" json:"permanent,omitempty"`
}

type MiddlewareHeaders struct {
	SSLRedirect           bool              `yaml:"sslRedirect,omitempty"           json:"sslRedirect,omitempty"`
	CustomRequestHeaders  map[string]string `yaml:"customRequestHeaders,omitempty"  json:"customRequestHeaders,omitempty"`
	CustomResponseHeaders map[string]string `yaml:"customResponseHeaders,omitempty" json:"customResponseHeaders,omitempty"`
}

type MiddlewareStripPrefix struct {
	Prefixes []string `yaml:"prefixes,omitempty" json:"prefixes,omitempty"`
}

type MiddlewareIPAllowList struct {
	SourceRange []string `yaml:"sourceRange,omitempty" json:"sourceRange,omitempty"`
}

type MiddlewareForwardAuth struct {
	Address string `yaml:"address,omitempty" json:"address,omitempty"`
}

type MiddlewareRateLimit struct {
	Average int `yaml:"average,omitempty" json:"average,omitempty"`
	Burst   int `yaml:"burst,omitempty"   json:"burst,omitempty"`
}

// ServiceSpec is the input for the "new service" wizard.
type ServiceSpec struct {
	Name            string   `json:"name"`
	Hostname        string   `json:"hostname"`
	BackendURL      string   `json:"backendUrl"`
	InsecureBackend bool     `json:"insecureBackend"`
	CertResolver    string   `json:"certResolver"`
	EntryPoints     []string `json:"entryPoints"`
}

// FileSummary is a condensed view of a dynamic config file for the list card.
type FileSummary struct {
	Name               string   `json:"name"`
	Active             bool     `json:"active"`
	Hostnames          []string `json:"hostnames"`
	Backends           []string `json:"backends"`
	CertResolver       string   `json:"certResolver"`
	InsecureSkipVerify bool     `json:"insecureSkipVerify"`
	RouterCount        int      `json:"routerCount"`
	ServiceCount       int      `json:"serviceCount"`
	MiddlewareCount    int      `json:"middlewareCount"`
}

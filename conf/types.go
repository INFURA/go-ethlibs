package conf

type ContextKeyMethod int

// HealthCheckEnvSpec represents the health check env variables struct.
type HealthCheckEnvSpec struct {
	ClientBlockLeeway     string `required:"true" envconfig:"CLIENT_BLOCK_LEEWAY"`
	ClientChainID         string `required:"true" envconfig:"CLIENT_CHAIN_ID"`
	ClientEndpoint        string `required:"true" envconfig:"CLIENT_ENDPOINT"`
	ExternalEndpoint      string `required:"true" envconfig:"HEALTHCHECK_EXTERNAL_ENDPOINT"`
	ListenerPort          string `required:"false" envconfig:"LISTENER_PORT" default:"8080"`
	LogLevel              string `required:"false" envconfig:"LOG_LEVEL" default:"info"`
	MetricsEnabled        string `required:"false" envconfig:"METRICS_ENABLED" default:"false"`
	MetricsRecordInterval string `required:"false" envconfig:"METRICS_RECORD_INTERVAL" default:"10s"`
	MinPeerCount		  string `required:"true" envconfig:"CLIENT_MIN_PEERCOUNT" default:"30"`
}

// MonitoringEnvSpec represents the monitoring env variables struct.
type MonitoringEnvSpec struct {
	Env                  string `required:"true"`
	Endpoint             string `required:"true"`
	ExternalEndpoint     string `required:"true" split_words:"true"`
	Network              string `required:"true"`
	InfuraProjectID      string `required:"true" split_words:"true"`
	InfuraProjectSecret  string `required:"true" split_words:"true"`
	ClientBlockLeeway    string `required:"true" envconfig:"CLIENT_BLOCK_LEEWAY"`
	ClientChainID        string `required:"true" envconfig:"CLIENT_CHAIN_ID"`
	DatadogAgentHost     string `required:"true" envconfig:"DD_AGENT_HOST"`
	CheckInterval        string `required:"false" envconfig:"CHECK_INTERVAL" default:"10s"`
	Internal             string `required:"false" default:"false"`
	LogLevel             string `required:"false" envconfig:"LOG_LEVEL" default:"info"`
	SendTxBlockedAddress string `required:"true" envconfig:"SEND_TX_BLOCKED_ADDRESS"`
	SendTxGoodAddress    string `required:"true" envconfig:"SEND_TX_GOOD_ADDRESS"`
}

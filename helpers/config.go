package helpers

type ServerConfig struct {
	Debug               bool
	Host                string
	Port                string
	HostURL             string
	HTTPSEnabled        bool
	DBDriver            string
	DBConfig            string
	AV_URL              string
	SkipSameDomainCheck bool
}

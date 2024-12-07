package config

type Jwt struct {
	ExpiredHours int    `yaml:"expired-hours"`
	Issuer       string `yaml:"issuer"`
	SecretKey    string `yaml:"secret-key"`
}

type Config struct {
	Jwt Jwt `yaml:"jwt"`
}

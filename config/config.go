package config

type Jwt struct {
	SecretKey string `json:"secretKey"`
}

type Config struct {
	Jwt Jwt `json:"jwt"`
}

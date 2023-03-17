package config

type AppConfig struct {
	RestAPIURL string `json:"restapi_url"`
	Secret     string `json:"secret"`
}

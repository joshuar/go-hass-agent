package agent

type AppConfig struct {
	RestAPIURL string `json:"restapi_url"`
	secret     string
	token      string
}

package schema

// HealthResponse is the health check response.
type HealthResponse struct {
	Status  string `json:"status"`
	Service string `json:"service"`
}

// RootResponse is the root endpoint response.
type RootResponse struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Status      string `json:"status"`
}

// ConfigResponse is the configuration endpoint response.
type ConfigResponse struct {
	ProviderBaseURL string `json:"provider_base_url"`
	StudioOrigin    string `json:"studio_origin"`
	CORS            string `json:"cors"`
}

// ErrorResponse is a generic error response.
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

package dto

// HealthMonitor matches the iac-proxy-v2 HealthMonitor proto message.
type HealthMonitor struct {
	ID                 string `json:"id"`
	Name               string `json:"name"`
	PoolID             string `json:"poolId"`
	Type               string `json:"type"`
	Delay              int    `json:"delay"`
	Timeout            int    `json:"timeout"`
	MaxRetries         int    `json:"maxRetries"`
	MaxRetriesDown     int    `json:"maxRetriesDown"`
	HTTPMethod         string `json:"httpMethod"`
	URLPath            string `json:"urlPath"`
	ExpectedCodes      string `json:"expectedCodes"`
	Status             string `json:"status"`
	ProvisioningStatus string `json:"provisioningStatus"`
	OperatingStatus    string `json:"operatingStatus"`
}

// CreateHealthMonitorRequest matches the iac-proxy-v2 CreateHealthMonitorRequest proto message.
// project_id is passed via URL path.
type CreateHealthMonitorRequest struct {
	Name           string `json:"name,omitempty"`
	PoolID         string `json:"poolId"`
	Type           string `json:"type"`
	Delay          int    `json:"delay"`
	Timeout        int    `json:"timeout"`
	MaxRetries     int    `json:"maxRetries"`
	MaxRetriesDown int    `json:"maxRetriesDown,omitempty"`
	HTTPMethod     string `json:"httpMethod,omitempty"`
	URLPath        string `json:"urlPath,omitempty"`
	ExpectedCodes  string `json:"expectedCodes,omitempty"`
}

// UpdateHealthMonitorRequest matches the iac-proxy-v2 UpdateHealthMonitorRequest proto message.
// project_id and id are passed via URL path.
type UpdateHealthMonitorRequest struct {
	Name           string `json:"name,omitempty"`
	Delay          int    `json:"delay,omitempty"`
	Timeout        int    `json:"timeout,omitempty"`
	MaxRetries     int    `json:"maxRetries,omitempty"`
	MaxRetriesDown int    `json:"maxRetriesDown,omitempty"`
	HTTPMethod     string `json:"httpMethod,omitempty"`
	URLPath        string `json:"urlPath,omitempty"`
	ExpectedCodes  string `json:"expectedCodes,omitempty"`
}

// HealthMonitorResponse matches the iac-proxy-v2 HealthMonitorResponse proto message.
type HealthMonitorResponse struct {
	HealthMonitor HealthMonitor `json:"healthMonitor"`
}

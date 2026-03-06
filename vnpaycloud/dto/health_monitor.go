package dto

// HealthMonitor matches the iac-proxy-v2 HealthMonitor proto message.
type HealthMonitor struct {
	ID            string `json:"id"`
	PoolID        string `json:"poolId"`
	Type          string `json:"type"`
	Delay         int    `json:"delay"`
	Timeout       int    `json:"timeout"`
	MaxRetries    int    `json:"maxRetries"`
	HTTPMethod    string `json:"httpMethod"`
	URLPath       string `json:"urlPath"`
	ExpectedCodes string `json:"expectedCodes"`
	Status        string `json:"status"`
}

// CreateHealthMonitorRequest matches the iac-proxy-v2 CreateHealthMonitorRequest proto message.
// project_id is passed via URL path.
type CreateHealthMonitorRequest struct {
	PoolID        string `json:"poolId"`
	Type          string `json:"type"`
	Delay         int    `json:"delay"`
	Timeout       int    `json:"timeout"`
	MaxRetries    int    `json:"maxRetries"`
	HTTPMethod    string `json:"httpMethod,omitempty"`
	URLPath       string `json:"urlPath,omitempty"`
	ExpectedCodes string `json:"expectedCodes,omitempty"`
}

// HealthMonitorResponse matches the iac-proxy-v2 HealthMonitorResponse proto message.
type HealthMonitorResponse struct {
	HealthMonitor HealthMonitor `json:"healthMonitor"`
}

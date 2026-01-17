package dto

type Monitor struct {
	ID                 string   `json:"id"`
	Name               string   `json:"name"`
	ProjectID          string   `json:"project_id"`
	Type               string   `json:"type"`
	Delay              int      `json:"delay"`
	Timeout            int      `json:"timeout"`
	MaxRetries         int      `json:"max_retries"`
	MaxRetriesDown     int      `json:"max_retries_down"`
	HTTPMethod         string   `json:"http_method"`
	HTTPVersion        string   `json:"http_version"`
	URLPath            string   `json:"url_path" `
	ExpectedCodes      string   `json:"expected_codes"`
	DomainName         string   `json:"domain_name"`
	AdminStateUp       bool     `json:"admin_state_up"`
	Status             string   `json:"status"`
	Pools              []PoolID `json:"pools"`
	ProvisioningStatus string   `json:"provisioning_status"`
	OperatingStatus    string   `json:"operating_status"`
	Tags               []string `json:"tags"`
}

type GetMonitorResponse struct {
	HealthMonitor Monitor `json:"healthmonitor"`
}

type CreateMonitorRequest struct {
	HealthMonitor CreateMonitorOpts `json:"healthmonitor"`
}

type CreateMonitorOpts struct {
	Name           string `json:"name,omitempty"`
	Type           string `json:"type"`                  // "PING", "TCP", "HTTP", "HTTPS"
	Delay          int    `json:"delay"`                 // seconds
	Timeout        int    `json:"timeout"`               // seconds
	MaxRetries     int    `json:"max_retries"`           // số lần success liên tục thì ONLINE
	MaxRetriesDown int    `json:"max_retries_down"`      // số lần fail liên tục thì down
	HTTPMethod     string `json:"http_method,omitempty"` // nếu là HTTP/HTTPS
	URLPath        string `json:"url_path,omitempty"`
	ExpectedCodes  string `json:"expected_codes,omitempty"`
	PoolID         string `json:"pool_id"` // pool gắn monitor
	AdminStateUp   bool   `json:"admin_state_up"`
	//Tags           []string `json:"tags,omitempty"`
}

type CreateMonitorResponse struct {
	HealthMonitor Monitor `json:"healthmonitor"`
}

type UpdateMonitorRequest struct {
	HealthMonitor UpdateMonitorOpts `json:"healthmonitor"`
}

type UpdateMonitorOpts struct {
	Name          string   `json:"name,omitempty"`
	Delay         int      `json:"delay,omitempty"`
	Timeout       int      `json:"timeout,omitempty"`
	MaxRetries    int      `json:"max_retries,omitempty"`
	HTTPMethod    string   `json:"http_method,omitempty"`
	URLPath       string   `json:"url_path,omitempty"`
	ExpectedCodes string   `json:"expected_codes,omitempty"`
	AdminStateUp  bool     `json:"admin_state_up,omitempty"`
	Tags          []string `json:"tags,omitempty"`
}

type UpdateMonitorResponse struct {
	HealthMonitor Monitor `json:"healthmonitor"`
}

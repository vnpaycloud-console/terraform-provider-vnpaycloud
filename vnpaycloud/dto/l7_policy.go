package dto

// L7Policy matches the backend L7Policy proto message.
type L7Policy struct {
	ID                 string `json:"id"`
	Name               string `json:"name"`
	ListenerID         string `json:"listenerId"`
	Action             string `json:"action"`
	Position           int    `json:"position"`
	Description        string `json:"description"`
	RedirectPoolID     string `json:"redirectPoolId"`
	RedirectURL        string `json:"redirectUrl"`
	Status             string `json:"status"`
	ProvisioningStatus string `json:"provisioningStatus"`
	OperatingStatus    string `json:"operatingStatus"`
}

type CreateL7PolicyRequest struct {
	Name           string `json:"name,omitempty"`
	ListenerID     string `json:"listenerId"`
	Action         string `json:"action"`
	Position       int    `json:"position,omitempty"`
	Description    string `json:"description,omitempty"`
	RedirectPoolID string `json:"redirectPoolId,omitempty"`
	RedirectURL    string `json:"redirectUrl,omitempty"`
}

type UpdateL7PolicyRequest struct {
	Name           string `json:"name,omitempty"`
	Action         string `json:"action,omitempty"`
	Position       int    `json:"position,omitempty"`
	Description    string `json:"description"`
	RedirectPoolID string `json:"redirectPoolId"`
	RedirectURL    string `json:"redirectUrl"`
}

type L7PolicyResponse struct {
	L7Policy L7Policy `json:"l7policy"`
}

type ListL7PoliciesResponse struct {
	L7Policies []L7Policy `json:"l7policies"`
}

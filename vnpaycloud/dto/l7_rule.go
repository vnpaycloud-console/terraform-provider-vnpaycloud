package dto

// L7Rule matches the iac-proxy-v2 L7Rule proto message.
type L7Rule struct {
	ID                 string `json:"id"`
	L7PolicyID         string `json:"l7policyId"`
	RuleType           string `json:"ruleType"`
	CompareType        string `json:"compareType"`
	Value              string `json:"value"`
	Key                string `json:"key"`
	Invert             bool   `json:"invert"`
	Status             string `json:"status"`
	ProvisioningStatus string `json:"provisioningStatus"`
	OperatingStatus    string `json:"operatingStatus"`
}

type CreateL7RuleRequest struct {
	RuleType    string `json:"ruleType"`
	CompareType string `json:"compareType"`
	Value       string `json:"value"`
	Key         string `json:"key,omitempty"`
	Invert      bool   `json:"invert"`
}

type UpdateL7RuleRequest struct {
	RuleType     string `json:"ruleType,omitempty"`
	CompareType  string `json:"compareType,omitempty"`
	Value        string `json:"value,omitempty"`
	Key          string `json:"key"`
	Invert       bool   `json:"invert"`
}

type L7RuleResponse struct {
	L7Rule L7Rule `json:"l7rule"`
}

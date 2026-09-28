package dto

// Monitoring as a Service (MaaS): pipelines, pipeline executors, access keys and role bindings.
// `order` is backed by a proto int64 on iac-proxy-v2, which grpc-gateway encodes as a quoted string.

type MaasPipeline struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Type        string `json:"type"`
	Status      string `json:"status"`
	CreatedAt   string `json:"createdAt"`
	ProjectID   string `json:"projectId"`
}

type MaasPipelineResponse struct {
	Pipeline MaasPipeline `json:"pipeline"`
}

type ListMaasPipelinesResponse struct {
	Pipelines []MaasPipeline `json:"pipelines"`
}

type CreateMaasPipelineRequest struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Description string `json:"description,omitempty"`
}

type UpdateMaasPipelineRequest struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

type MaasPipelineExecutor struct {
	ID           string            `json:"id"`
	PipelineID   string            `json:"pipelineId"`
	Order        int               `json:"order,string"`
	Description  string            `json:"description"`
	Type         string            `json:"type"`
	ExecutorType string            `json:"executorType"`
	StaticLabels map[string]string `json:"staticLabels"`
	RenameFields map[string]string `json:"renameFields"`
	RenameLabels map[string]string `json:"renameLabels"`
	Status       string            `json:"status"`
	CreatedAt    string            `json:"createdAt"`
	ProjectID    string            `json:"projectId"`
}

type MaasPipelineExecutorResponse struct {
	PipelineExecutor MaasPipelineExecutor `json:"pipelineExecutor"`
}

type ListMaasPipelineExecutorsResponse struct {
	PipelineExecutors []MaasPipelineExecutor `json:"pipelineExecutors"`
}

type CreateMaasPipelineExecutorRequest struct {
	PipelineID   string            `json:"pipelineId"`
	Order        int               `json:"order,string"`
	Description  string            `json:"description,omitempty"`
	ExecutorType string            `json:"executorType"`
	StaticLabels map[string]string `json:"staticLabels,omitempty"`
	RenameFields map[string]string `json:"renameFields,omitempty"`
	RenameLabels map[string]string `json:"renameLabels,omitempty"`
}

type UpdateMaasPipelineExecutorRequest struct {
	Order        int               `json:"order,string"`
	Description  string            `json:"description,omitempty"`
	ExecutorType string            `json:"executorType"`
	StaticLabels map[string]string `json:"staticLabels,omitempty"`
	RenameFields map[string]string `json:"renameFields,omitempty"`
	RenameLabels map[string]string `json:"renameLabels,omitempty"`
}

type MaasLabelMatcher struct {
	Name  string `json:"name"`
	Op    string `json:"op"`
	Value string `json:"value"`
}

type MaasEndpoints struct {
	LogOtlpPush          string `json:"logOtlpPush"`
	MetricOtlpPush       string `json:"metricOtlpPush"`
	LogLokiPush          string `json:"logLokiPush"`
	MetricPrometheusPush string `json:"metricPrometheusPush"`
}

type MaasAccessKey struct {
	ID                  string             `json:"id"`
	Name                string             `json:"name"`
	Description         string             `json:"description"`
	LogPermission       string             `json:"logPermission"`
	MetricPermission    string             `json:"metricPermission"`
	LogPipelineID       string             `json:"logPipelineId"`
	MetricPipelineID    string             `json:"metricPipelineId"`
	LogLabelMatchers    []MaasLabelMatcher `json:"logLabelMatchers"`
	MetricLabelMatchers []MaasLabelMatcher `json:"metricLabelMatchers"`
	Endpoints           *MaasEndpoints     `json:"endpoints"`
	APIKey              string             `json:"apiKey"`
	Username            string             `json:"username"`
	Password            string             `json:"password"`
	Status              string             `json:"status"`
	CreatedAt           string             `json:"createdAt"`
	ProjectID           string             `json:"projectId"`
}

type MaasAccessKeyResponse struct {
	AccessKey MaasAccessKey `json:"accessKey"`
}

type ListMaasAccessKeysResponse struct {
	AccessKeys []MaasAccessKey `json:"accessKeys"`
}

type CreateMaasAccessKeyRequest struct {
	Name                string             `json:"name"`
	Description         string             `json:"description,omitempty"`
	LogPermission       string             `json:"logPermission,omitempty"`
	MetricPermission    string             `json:"metricPermission,omitempty"`
	LogPipelineID       string             `json:"logPipelineId,omitempty"`
	MetricPipelineID    string             `json:"metricPipelineId,omitempty"`
	LogLabelMatchers    []MaasLabelMatcher `json:"logLabelMatchers,omitempty"`
	MetricLabelMatchers []MaasLabelMatcher `json:"metricLabelMatchers,omitempty"`
}

type UpdateMaasAccessKeyRequest struct {
	Name                string             `json:"name"`
	Description         string             `json:"description,omitempty"`
	LogPermission       string             `json:"logPermission,omitempty"`
	MetricPermission    string             `json:"metricPermission,omitempty"`
	LogPipelineID       string             `json:"logPipelineId,omitempty"`
	MetricPipelineID    string             `json:"metricPipelineId,omitempty"`
	LogLabelMatchers    []MaasLabelMatcher `json:"logLabelMatchers,omitempty"`
	MetricLabelMatchers []MaasLabelMatcher `json:"metricLabelMatchers,omitempty"`
}

type UpdateMaasAccessKeyStatusRequest struct {
	Status string `json:"status"`
}

type MaasRoleBinding struct {
	ID           string `json:"id"`
	PortalUserID string `json:"portalUserId"`
	Role         string `json:"role"`
	Status       string `json:"status"`
	CreatedAt    string `json:"createdAt"`
	ProjectID    string `json:"projectId"`
}

type MaasRoleBindingResponse struct {
	RoleBinding MaasRoleBinding `json:"roleBinding"`
}

type ListMaasRoleBindingsResponse struct {
	RoleBindings []MaasRoleBinding `json:"roleBindings"`
}

type CreateMaasRoleBindingRequest struct {
	PortalUserID string `json:"portalUserId"`
	Role         string `json:"role"`
}

type UpdateMaasRoleBindingRequest struct {
	Role string `json:"role"`
}

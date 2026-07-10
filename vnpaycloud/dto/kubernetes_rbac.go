package dto

// KubernetesRole matches the backend KubernetesRole proto message. A role available in a
// cluster that can be bound to a user (cluster-admin/admin/edit/view are seeded per cluster).
type KubernetesRole struct {
	ID            string `json:"id"`
	ClusterID     string `json:"clusterId"`
	Name          string `json:"name"`
	Namespace     string `json:"namespace"`
	IsClusterRole bool   `json:"isClusterRole"`
	Status        string `json:"status"`
	CreatedAt     string `json:"createdAt"`
}

// ListKubernetesRolesResponse matches the backend ListKubernetesRolesResponse proto message.
type ListKubernetesRolesResponse struct {
	KubernetesRoles []KubernetesRole `json:"kubernetesRoles"`
}

// KubernetesRbac matches the backend KubernetesRbac proto message: a role binding that binds
// a user to a (cluster) role in a cluster.
type KubernetesRbac struct {
	ID                   string `json:"id"`
	ClusterID            string `json:"clusterId"`
	UserID               string `json:"userId"`
	Email                string `json:"email"`
	KubernetesRoleID     string `json:"kubernetesRoleId"`
	Role                 string `json:"role"`
	BindingType          string `json:"bindingType"`
	IsClusterRoleBinding bool   `json:"isClusterRoleBinding"`
	Namespace            string `json:"namespace"`
	Name                 string `json:"name"`
	Status               string `json:"status"`
	CreatedAt            string `json:"createdAt"`
}

// CreateKubernetesRbacRequest matches the backend CreateKubernetesRbacRequest proto message.
// project_id and cluster_id are passed via the URL path.
type CreateKubernetesRbacRequest struct {
	UserID      string `json:"userId"`
	Role        string `json:"role"`
	BindingType string `json:"bindingType,omitempty"`
	Namespace   string `json:"namespace,omitempty"`
}

// KubernetesRbacResponse matches the backend KubernetesRbacResponse proto message.
type KubernetesRbacResponse struct {
	Rbac KubernetesRbac `json:"rbac"`
}

// ListKubernetesRbacsResponse matches the backend ListKubernetesRbacsResponse proto message.
type ListKubernetesRbacsResponse struct {
	Rbacs []KubernetesRbac `json:"rbacs"`
}

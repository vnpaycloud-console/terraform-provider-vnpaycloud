package dto

// --- Postgres Instance ---

type PostgresInstance struct {
	ID                    string `json:"id"`
	Name                  string `json:"name"`
	Description           string `json:"description"`
	DatabaseClusterID     string `json:"databaseClusterId"`
	FlavorDatabaseID      string `json:"flavorDatabaseId"`
	Version               string `json:"version"`
	VolumeType            string `json:"volumeType"`
	VolumeSize            int64  `json:"volumeSize,string"`
	Mode                  string `json:"mode"`
	PrimaryIP             string `json:"primaryIp"`
	PrimaryPort           int    `json:"primaryPort"`
	StandbyIP             string `json:"standbyIp"`
	StandbyPort           int    `json:"standbyPort"`
	ProjectName           string `json:"projectName"`
	Replica               int    `json:"replica"`
	Namespace             string `json:"namespace"`
	Purpose               string `json:"purpose"`
	IsAutoExpandVolume    bool   `json:"isAutoExpandVolume"`
	UsageThreshold        int    `json:"usageThreshold"`
	ScalePercent          int    `json:"scalePercent"`
	EnableTls             bool   `json:"enableTls"`
	CertificateID         string `json:"certificateId"`
	TlsMode               string `json:"tlsMode"`
	IsAttachedGateway     bool   `json:"isAttachedGateway"`
	DataMsg               string `json:"dataMsg"`
	ZoneID                string `json:"zoneId"`
	Status                string `json:"status"`
	CreatedAt             string `json:"createdAt"`
	CustomerAdminUsername  string `json:"customerAdminUsername"`
}

type CreatePostgresInstanceRequest struct {
	Name            string `json:"name"`
	Description     string `json:"description,omitempty"`
	FlavorDatabaseID string `json:"flavorDatabaseId"`
	Version         string `json:"version"`
	VolumeType      string `json:"volumeType"`
	VolumeSize      int64  `json:"volumeSize"`
	Mode            string `json:"mode"`
	Replica         int    `json:"replica"`
	Purpose         string `json:"purpose,omitempty"`
	EnableTls       bool   `json:"enableTls,omitempty"`
	CertificateID   string `json:"certificateId,omitempty"`
	TlsMode         string `json:"tlsMode,omitempty"`
	ZoneID          string `json:"zoneId"`
}

type PostgresInstanceResponse struct {
	PostgresInstance PostgresInstance `json:"postgresInstance"`
}

type ListPostgresInstancesResponse struct {
	PostgresInstances []PostgresInstance `json:"postgresInstances"`
	Total             int               `json:"total"`
}

type ScalePostgresInstanceRequest struct {
	Replica int `json:"replica"`
}

type ChangeFlavorPostgresInstanceRequest struct {
	FlavorDatabaseID string `json:"flavorDatabaseId"`
}

type ExpandVolumePostgresInstanceRequest struct {
	VolumeSize int64 `json:"volumeSize"`
}

type EnableAutoExpandVolumePostgresInstanceRequest struct {
	UsageThreshold int `json:"usageThreshold"`
	ScalePercent   int `json:"scalePercent"`
}

type EnableTlsPostgresInstanceRequest struct {
	CertificateID string `json:"certificateId,omitempty"`
	TlsMode       string `json:"tlsMode,omitempty"`
}

// --- Redis Instance ---

type RedisInstance struct {
	ID                    string `json:"id"`
	Name                  string `json:"name"`
	Description           string `json:"description"`
	DatabaseClusterID     string `json:"databaseClusterId"`
	FlavorDatabaseID      string `json:"flavorDatabaseId"`
	Version               string `json:"version"`
	VolumeType            string `json:"volumeType"`
	VolumeSize            int64  `json:"volumeSize,string"`
	Mode                  string `json:"mode"`
	PrimaryIP             string `json:"primaryIp"`
	PrimaryPort           int    `json:"primaryPort"`
	ProjectName           string `json:"projectName"`
	Replica               int    `json:"replica"`
	Namespace             string `json:"namespace"`
	Purpose               string `json:"purpose"`
	IsAutoExpandVolume    bool   `json:"isAutoExpandVolume"`
	UsageThreshold        int    `json:"usageThreshold"`
	ScalePercent          int    `json:"scalePercent"`
	EnableTls             bool   `json:"enableTls"`
	CertificateID         string `json:"certificateId"`
	IsAttachedGateway     bool   `json:"isAttachedGateway"`
	DataMsg               string `json:"dataMsg"`
	ZoneID                string `json:"zoneId"`
	Status                string `json:"status"`
	CreatedAt             string `json:"createdAt"`
	CustomerAdminUsername  string `json:"customerAdminUsername"`
}

type CreateRedisInstanceRequest struct {
	Name            string `json:"name"`
	Description     string `json:"description,omitempty"`
	FlavorDatabaseID string `json:"flavorDatabaseId"`
	Version         string `json:"version"`
	VolumeType      string `json:"volumeType"`
	VolumeSize      int64  `json:"volumeSize"`
	Replica         int    `json:"replica"`
	Purpose         string `json:"purpose,omitempty"`
	EnableTls       bool   `json:"enableTls,omitempty"`
	CertificateID   string `json:"certificateId,omitempty"`
	ZoneID          string `json:"zoneId"`
}

type RedisInstanceResponse struct {
	RedisInstance RedisInstance `json:"redisInstance"`
}

type ListRedisInstancesResponse struct {
	RedisInstances []RedisInstance `json:"redisInstances"`
	Total          int            `json:"total"`
}

type ChangeFlavorRedisInstanceRequest struct {
	FlavorDatabaseID string `json:"flavorDatabaseId"`
}

type ExpandVolumeRedisInstanceRequest struct {
	VolumeSize int64 `json:"volumeSize"`
}

type EnableAutoExpandVolumeRedisInstanceRequest struct {
	UsageThreshold int `json:"usageThreshold"`
	ScalePercent   int `json:"scalePercent"`
}

type EnableTlsRedisInstanceRequest struct {
	CertificateID string `json:"certificateId,omitempty"`
}

// --- Redis Sentinel Instance ---

type RedisSentinelInstance struct {
	ID                       string `json:"id"`
	Name                     string `json:"name"`
	Description              string `json:"description"`
	DatabaseClusterID        string `json:"databaseClusterId"`
	FlavorDatabaseID         string `json:"flavorDatabaseId"`
	Version                  string `json:"version"`
	VolumeType               string `json:"volumeType"`
	VolumeSize               int64  `json:"volumeSize,string"`
	PrimaryIP                string `json:"primaryIp"`
	PrimaryPort              int    `json:"primaryPort"`
	StandbyIP                string `json:"standbyIp"`
	StandbyPort              int    `json:"standbyPort"`
	ProjectName              string `json:"projectName"`
	Replica                  int    `json:"replica"`
	Namespace                string `json:"namespace"`
	Purpose                  string `json:"purpose"`
	IsAutoExpandVolume       bool   `json:"isAutoExpandVolume"`
	UsageThreshold           int    `json:"usageThreshold"`
	ScalePercent             int    `json:"scalePercent"`
	SentinelName             string `json:"sentinelName"`
	SentinelReplica          int    `json:"sentinelReplica"`
	SentinelFlavorDatabaseID string `json:"sentinelFlavorDatabaseId"`
	SentinelVolumeSize       int64  `json:"sentinelVolumeSize,string"`
	EnableTls                bool   `json:"enableTls"`
	CertificateID            string `json:"certificateId"`
	IsAttachedGateway        bool   `json:"isAttachedGateway"`
	DataMsg                  string `json:"dataMsg"`
	ZoneID                   string `json:"zoneId"`
	Status                   string `json:"status"`
	CreatedAt                string `json:"createdAt"`
	CustomerAdminUsername     string `json:"customerAdminUsername"`
}

type CreateRedisSentinelInstanceRequest struct {
	Name                     string `json:"name"`
	Description              string `json:"description,omitempty"`
	FlavorDatabaseID         string `json:"flavorDatabaseId"`
	Version                  string `json:"version"`
	VolumeType               string `json:"volumeType"`
	VolumeSize               int64  `json:"volumeSize"`
	Replica                  int    `json:"replica"`
	Purpose                  string `json:"purpose,omitempty"`
	SentinelName             string `json:"sentinelName"`
	SentinelReplica          int    `json:"sentinelReplica"`
	SentinelFlavorDatabaseID string `json:"sentinelFlavorDatabaseId"`
	SentinelVolumeSize       int64  `json:"sentinelVolumeSize"`
	EnableTls                bool   `json:"enableTls,omitempty"`
	CertificateID            string `json:"certificateId,omitempty"`
	ZoneID                   string `json:"zoneId"`
}

type RedisSentinelInstanceResponse struct {
	RedisSentinelInstance RedisSentinelInstance `json:"redisSentinelInstance"`
}

type ListRedisSentinelInstancesResponse struct {
	RedisSentinelInstances []RedisSentinelInstance `json:"redisSentinelInstances"`
	Total                  int                     `json:"total"`
}

type ScaleRedisSentinelInstanceRequest struct {
	Replica int `json:"replica"`
}

type ChangeFlavorRedisSentinelInstanceRequest struct {
	FlavorDatabaseID string `json:"flavorDatabaseId"`
}

type ExpandVolumeRedisSentinelInstanceRequest struct {
	VolumeSize int64 `json:"volumeSize"`
}

type EnableAutoExpandVolumeRedisSentinelInstanceRequest struct {
	UsageThreshold int `json:"usageThreshold"`
	ScalePercent   int `json:"scalePercent"`
}

type ScaleRedisSentinelRequest struct {
	SentinelReplica int `json:"sentinelReplica"`
}

type ChangeFlavorRedisSentinelRequest struct {
	SentinelFlavorDatabaseID string `json:"sentinelFlavorDatabaseId"`
}

type EnableTlsRedisSentinelInstanceRequest struct {
	CertificateID string `json:"certificateId,omitempty"`
}

// --- Flavor Database ---

type FlavorDatabase struct {
	ID       string `json:"id"`
	Class    string `json:"class"`
	Ratio    string `json:"ratio"`
	Name     string `json:"name"`
	CpuReq   int    `json:"cpuReq"`
	MemReq   int    `json:"memReq"`
	CpuLimit int    `json:"cpuLimit"`
	MemLimit int    `json:"memLimit"`
}

type FlavorDatabaseResponse struct {
	FlavorDatabase FlavorDatabase `json:"flavorDatabase"`
}

type ListFlavorDatabasesResponse struct {
	FlavorDatabases []FlavorDatabase `json:"flavorDatabases"`
	Total           int              `json:"total"`
}

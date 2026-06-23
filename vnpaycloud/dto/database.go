package dto

// --- Postgres Instance ---

type PostgresInstance struct {
	EnableReadOnlyEndpoint bool   `json:"enableReadOnlyEndpoint"`
	ID                     string `json:"id"`
	Name                   string `json:"name"`
	Description            string `json:"description"`
	DatabaseClusterID      string `json:"databaseClusterId"`
	FlavorDatabaseID       string `json:"flavorDatabaseId"`
	Version                string `json:"version"`
	VolumeType             string `json:"volumeType"`
	VolumeSize             int64  `json:"volumeSize,string"`
	Mode                   string `json:"mode"`
	PrimaryIP              string `json:"primaryIp"`
	PrimaryPort            int    `json:"primaryPort"`
	StandbyIP              string `json:"standbyIp"`
	StandbyPort            int    `json:"standbyPort"`
	ProjectName            string `json:"projectName"`
	Replica                int    `json:"replica"`
	Namespace              string `json:"namespace"`
	Purpose                string `json:"purpose"`
	IsAutoExpandVolume     bool   `json:"isAutoExpandVolume"`
	UsageThreshold         int    `json:"usageThreshold"`
	ScalePercent           int    `json:"scalePercent"`
	EnableTls              bool   `json:"enableTls"`
	CertificateID          string `json:"certificateId"`
	TlsMode                string `json:"tlsMode"`
	IsAttachedGateway      bool   `json:"isAttachedGateway"`
	DataMsg                string `json:"dataMsg"`
	ZoneID                 string `json:"zoneId"`
	Status                 string `json:"status"`
	CreatedAt              string `json:"createdAt"`
	CustomerAdminUsername  string `json:"customerAdminUsername"`
}

type CreatePostgresInstanceRequest struct {
	Name             string `json:"name"`
	Description      string `json:"description,omitempty"`
	FlavorDatabaseID string `json:"flavorDatabaseId"`
	Version          string `json:"version"`
	VolumeType       string `json:"volumeType"`
	VolumeSize       int64  `json:"volumeSize"`
	Mode             string `json:"mode"`
	Replica          int    `json:"replica"`
	Purpose          string `json:"purpose,omitempty"`
	EnableTls        bool   `json:"enableTls,omitempty"`
	CertificateID    string `json:"certificateId,omitempty"`
	TlsMode          string `json:"tlsMode,omitempty"`
	ZoneID           string `json:"zoneId"`
}

type PostgresInstanceResponse struct {
	PostgresInstance PostgresInstance `json:"postgresInstance"`
}

type ListPostgresInstancesResponse struct {
	PostgresInstances []PostgresInstance `json:"postgresInstances"`
	Total             int                `json:"total"`
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

// --- Postgres Account ---

type PostgresAccountGrant struct {
	DbName            string `json:"dbName"`
	DbSchema          string `json:"dbSchema"`
	PrivilegeTemplate string `json:"privilegeTemplate"`
}

type PostgresAccount struct {
	ID                 string                 `json:"id"`
	Name               string                 `json:"name"`
	PostgresInstanceID string                 `json:"postgresInstanceId"`
	Grants             []PostgresAccountGrant `json:"grants"`
	ZoneID             string                 `json:"zoneId"`
	Status             string                 `json:"status"`
	CreatedAt          string                 `json:"createdAt"`
}

type CreatePostgresAccountRequest struct {
	Name               string `json:"name"`
	PostgresInstanceID string `json:"postgresInstanceId"`
	Password           string `json:"password"`
}

type PostgresAccountResponse struct {
	PostgresAccount PostgresAccount `json:"postgresAccount"`
}

type ListPostgresAccountsResponse struct {
	PostgresAccounts []PostgresAccount `json:"postgresAccounts"`
	Total            int               `json:"total"`
}

type ChangePasswordPostgresAccountRequest struct {
	NewPassword string `json:"newPassword"`
}

type GrantPrivilegesPostgresAccountRequest struct {
	PrivilegeTemplate string `json:"privilegeTemplate"`
	DbName            string `json:"dbName"`
	DbSchema          string `json:"dbSchema"`
}

type RevokePrivilegesPostgresAccountRequest struct {
	DbName   string `json:"dbName"`
	DbSchema string `json:"dbSchema"`
}

// --- Postgres Database ---

type PostgresDatabase struct {
	ID                 string `json:"id"`
	Name               string `json:"name"`
	PostgresInstanceID string `json:"postgresInstanceId"`
	Owner              string `json:"owner"`
	ZoneID             string `json:"zoneId"`
	Status             string `json:"status"`
	CreatedAt          string `json:"createdAt"`
}

type CreatePostgresDatabaseRequest struct {
	Name               string `json:"name"`
	PostgresInstanceID string `json:"postgresInstanceId"`
	Owner              string `json:"owner"`
}

type PostgresDatabaseResponse struct {
	PostgresDatabase PostgresDatabase `json:"postgresDatabase"`
}

type ListPostgresDatabasesResponse struct {
	PostgresDatabases []PostgresDatabase `json:"postgresDatabases"`
	Total             int                `json:"total"`
}

type ChangeOwnershipPostgresDatabaseRequest struct {
	NewOwner string `json:"newOwner"`
}

// --- Redis Account ---

type RedisAccount struct {
	ID                string `json:"id"`
	Name              string `json:"name"`
	RedisInstanceID   string `json:"redisInstanceId"`
	PrivilegeTemplate string `json:"privilegeTemplate"`
	ZoneID            string `json:"zoneId"`
	Status            string `json:"status"`
	CreatedAt         string `json:"createdAt"`
}

type CreateRedisAccountRequest struct {
	Name              string `json:"name"`
	RedisInstanceID   string `json:"redisInstanceId"`
	Password          string `json:"password"`
	PrivilegeTemplate string `json:"privilegeTemplate"`
}

type RedisAccountResponse struct {
	RedisAccount RedisAccount `json:"redisAccount"`
}

type ListRedisAccountsResponse struct {
	RedisAccounts []RedisAccount `json:"redisAccounts"`
	Total         int            `json:"total"`
}

type ChangePasswordRedisAccountRequest struct {
	NewPassword       string `json:"newPassword"`
	PrivilegeTemplate string `json:"privilegeTemplate"`
}

type GrantPrivilegeRedisAccountRequest struct {
	PrivilegeTemplate string `json:"privilegeTemplate"`
}

// --- Redis Sentinel Account ---

type RedisSentinelAccount struct {
	ID                      string `json:"id"`
	Name                    string `json:"name"`
	RedisSentinelInstanceID string `json:"redisSentinelInstanceId"`
	PrivilegeTemplate       string `json:"privilegeTemplate"`
	ZoneID                  string `json:"zoneId"`
	Status                  string `json:"status"`
	CreatedAt               string `json:"createdAt"`
}

type CreateRedisSentinelAccountRequest struct {
	Name                    string `json:"name"`
	RedisSentinelInstanceID string `json:"redisSentinelInstanceId"`
	Password                string `json:"password"`
	PrivilegeTemplate       string `json:"privilegeTemplate"`
}

type RedisSentinelAccountResponse struct {
	RedisSentinelAccount RedisSentinelAccount `json:"redisSentinelAccount"`
}

type ListRedisSentinelAccountsResponse struct {
	RedisSentinelAccounts []RedisSentinelAccount `json:"redisSentinelAccounts"`
	Total                 int                    `json:"total"`
}

type ChangePasswordRedisSentinelAccountRequest struct {
	NewPassword       string `json:"newPassword"`
	PrivilegeTemplate string `json:"privilegeTemplate"`
}

type GrantPrivilegeRedisSentinelAccountRequest struct {
	PrivilegeTemplate string `json:"privilegeTemplate"`
}

// --- Database Versions ---

type ListPostgresVersionsResponse struct {
	Versions []string `json:"versions"`
}

type ListRedisVersionsResponse struct {
	Versions []string `json:"versions"`
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
	CustomerAdminUsername string `json:"customerAdminUsername"`
}

type CreateRedisInstanceRequest struct {
	Name             string `json:"name"`
	Description      string `json:"description,omitempty"`
	FlavorDatabaseID string `json:"flavorDatabaseId"`
	Version          string `json:"version"`
	VolumeType       string `json:"volumeType"`
	VolumeSize       int64  `json:"volumeSize"`
	Replica          int    `json:"replica"`
	Purpose          string `json:"purpose,omitempty"`
	EnableTls        bool   `json:"enableTls,omitempty"`
	CertificateID    string `json:"certificateId,omitempty"`
	ZoneID           string `json:"zoneId"`
}

type RedisInstanceResponse struct {
	RedisInstance RedisInstance `json:"redisInstance"`
}

type ListRedisInstancesResponse struct {
	RedisInstances []RedisInstance `json:"redisInstances"`
	Total          int             `json:"total"`
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
	EnableReadOnlyEndpoint   bool   `json:"enableReadOnlyEndpoint"`
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
	CustomerAdminUsername    string `json:"customerAdminUsername"`
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

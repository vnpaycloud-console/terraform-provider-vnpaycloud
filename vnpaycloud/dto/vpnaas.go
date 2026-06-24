package dto

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// --- VPNaaS VPNGateway ---

// VPNGateway matches the backend VPNGateway proto message.
type VPNGateway struct {
	ID             string   `json:"id"`
	Name           string   `json:"name"`
	Description    string   `json:"description"`
	VPNType        string   `json:"vpnType"`
	Status         string   `json:"status"`
	AttachedVPCIDs []string `json:"attachedVpcIds"`
	CreatedAt      string   `json:"createdAt"`
	ProjectID      string   `json:"projectId"`
	ZoneID         string   `json:"zoneId"`
}

// CreateVPNGatewayRequest matches the backend CreateVPNGatewayRequest proto message.
// project_id is passed via URL path.
type CreateVPNGatewayRequest struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	VPNType     string `json:"vpnType"`
}

// UpdateVPNGatewayRequest matches the backend UpdateVPNGatewayRequest proto message.
// project_id and id are passed via URL path.
type UpdateVPNGatewayRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// AttachVPCToVPNGatewayRequest matches the backend AttachVPCToVPNGatewayRequest proto message.
// project_id and vpn_gateway_id are passed via URL path.
type AttachVPCToVPNGatewayRequest struct {
	VPCID string `json:"vpcId"`
}

// DetachVPCFromVPNGatewayRequest matches the backend DetachVPCFromVPNGatewayRequest proto message.
// project_id and vpn_gateway_id are passed via URL path.
type DetachVPCFromVPNGatewayRequest struct {
	VPCID string `json:"vpcId"`
}

// VPNGatewayResponse matches the backend VPNGatewayResponse proto message.
type VPNGatewayResponse struct {
	VPNGateway VPNGateway `json:"vpnGateway"`
}

// ListVPNGatewaysResponse matches the backend ListVPNGatewaysResponse proto message.
type ListVPNGatewaysResponse struct {
	VPNGateways []VPNGateway `json:"vpnGateways"`
}

// --- VPNaaS VPNConnection ---

type IKEProfileConfig struct {
	IKEVersion     string `json:"ikeVersion,omitempty"`
	IKELifetime    int    `json:"ikeLifetime,omitempty"`
	IKECloseAction string `json:"ikeCloseAction,omitempty"`
	IKEDH          string `json:"ikeDh,omitempty"`
	IKEEncryption  string `json:"ikeEncryption,omitempty"`
	IKEHash        string `json:"ikeHash,omitempty"`
	IKEPRF         string `json:"ikePrf,omitempty"`
	IKEDPDAction   string `json:"ikeDpdAction,omitempty"`
	IKEDPDInterval int    `json:"ikeDpdInterval,omitempty"`
	IKEDPDTimeout  int    `json:"ikeDpdTimeout,omitempty"`
	IKEV2Reauth    bool   `json:"ikev2Reauth,omitempty"`
}

type IPSecProfileConfig struct {
	IPSecLifetime        int    `json:"ipsecLifetime,omitempty"`
	IPSecPFS             string `json:"ipsecPfs,omitempty"`
	IPSecEncryption      string `json:"ipsecEncryption,omitempty"`
	IPSecHash            string `json:"ipsecHash,omitempty"`
	IPSecDisableRekey    bool   `json:"ipsecDisableRekey,omitempty"`
	IPSecLifetimeBytes   int64  `json:"ipsecLifetimeBytes,omitempty"`
	IPSecLifetimePackets int64  `json:"ipsecLifetimePackets,omitempty"`
}

func (c *IPSecProfileConfig) UnmarshalJSON(data []byte) error {
	var raw struct {
		IPSecLifetime        int             `json:"ipsecLifetime"`
		IPSecPFS             string          `json:"ipsecPfs"`
		IPSecEncryption      string          `json:"ipsecEncryption"`
		IPSecHash            string          `json:"ipsecHash"`
		IPSecDisableRekey    bool            `json:"ipsecDisableRekey"`
		IPSecLifetimeBytes   json.RawMessage `json:"ipsecLifetimeBytes"`
		IPSecLifetimePackets json.RawMessage `json:"ipsecLifetimePackets"`
	}

	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	lifetimeBytes, err := parseFlexibleInt64(raw.IPSecLifetimeBytes)
	if err != nil {
		return fmt.Errorf("ipsecLifetimeBytes: %w", err)
	}

	lifetimePackets, err := parseFlexibleInt64(raw.IPSecLifetimePackets)
	if err != nil {
		return fmt.Errorf("ipsecLifetimePackets: %w", err)
	}

	c.IPSecLifetime = raw.IPSecLifetime
	c.IPSecPFS = raw.IPSecPFS
	c.IPSecEncryption = raw.IPSecEncryption
	c.IPSecHash = raw.IPSecHash
	c.IPSecDisableRekey = raw.IPSecDisableRekey
	c.IPSecLifetimeBytes = lifetimeBytes
	c.IPSecLifetimePackets = lifetimePackets

	return nil
}

type IPSecAuthenticationConfig struct {
	PSK string `json:"psk,omitempty"`
}

type RouteBaseConfig struct {
	VTIMSS int `json:"vtiMss,omitempty"`
}

type ConnectionBGPConfig struct {
	BGPKeepalive int `json:"bgpKeepalive,omitempty"`
	BGPHoldtime  int `json:"bgpHoldtime,omitempty"`
}

// VPNConnection matches the backend VPNConnection proto message.
type VPNConnection struct {
	ID                  string               `json:"id"`
	Name                string               `json:"name"`
	Description         string               `json:"description"`
	VPNGatewayID        string               `json:"vpnGatewayId"`
	CustomerGatewayID   string               `json:"customerGatewayId"`
	VPNType             string               `json:"vpnType"`
	Status              string               `json:"status"`
	VPNPublicIPID       string               `json:"vpnPublicIpId"`
	IKEProfileConfig    *IKEProfileConfig    `json:"ikeProfileConfig,omitempty"`
	IPSecProfileConfig  *IPSecProfileConfig  `json:"ipsecProfileConfig,omitempty"`
	RouteBaseConfig     *RouteBaseConfig     `json:"routeBaseConfig,omitempty"`
	ConnectionBGPConfig *ConnectionBGPConfig `json:"connectionBgpConfig,omitempty"`
	CreatedAt           string               `json:"createdAt"`
	ProjectID           string               `json:"projectId"`
	ZoneID              string               `json:"zoneId"`
}

// CreateVPNConnectionRequest matches the backend CreateVPNConnectionRequest proto message.
// project_id is passed via URL path.
type CreateVPNConnectionRequest struct {
	Name                string                     `json:"name"`
	Description         string                     `json:"description,omitempty"`
	VPNGatewayID        string                     `json:"vpnGatewayId"`
	CustomerGatewayID   string                     `json:"customerGatewayId"`
	VPNType             string                     `json:"vpnType"`
	VPNPublicIPID       string                     `json:"vpnPublicIpId"`
	IKEProfileConfig    *IKEProfileConfig          `json:"ikeProfileConfig,omitempty"`
	IPSecProfileConfig  *IPSecProfileConfig        `json:"ipsecProfileConfig,omitempty"`
	IPSecAuthConfig     *IPSecAuthenticationConfig `json:"ipsecAuthConfig,omitempty"`
	RouteBaseConfig     *RouteBaseConfig           `json:"routeBaseConfig,omitempty"`
	ConnectionBGPConfig *ConnectionBGPConfig       `json:"connectionBgpConfig,omitempty"`
}

// VPNConnectionResponse matches the backend VPNConnectionResponse proto message.
type VPNConnectionResponse struct {
	VPNConnection VPNConnection `json:"vpnConnection"`
}

// ListVPNConnectionsResponse matches the backend ListVPNConnectionsResponse proto message.
type ListVPNConnectionsResponse struct {
	VPNConnections []VPNConnection `json:"vpnConnections"`
}

// --- VPNaaS CustomerGateway ---

// BGPConfig matches the backend BGPConfig proto message.
type BGPConfig struct {
	LocalAs int64  `json:"localAs"`
	PeerAs  int64  `json:"peerAs"`
	AsPath  string `json:"asPath"`
}

func (c *BGPConfig) UnmarshalJSON(data []byte) error {
	var raw struct {
		LocalAs json.RawMessage `json:"localAs"`
		PeerAs  json.RawMessage `json:"peerAs"`
		AsPath  string          `json:"asPath"`
	}

	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	localAs, err := parseBGPASN(raw.LocalAs)
	if err != nil {
		return fmt.Errorf("localAs: %w", err)
	}

	peerAs, err := parseBGPASN(raw.PeerAs)
	if err != nil {
		return fmt.Errorf("peerAs: %w", err)
	}

	c.LocalAs = localAs
	c.PeerAs = peerAs
	c.AsPath = raw.AsPath

	return nil
}

func parseBGPASN(raw json.RawMessage) (int64, error) {
	return parseFlexibleInt64(raw)
}

func parseFlexibleInt64(raw json.RawMessage) (int64, error) {
	if len(raw) == 0 || string(raw) == "null" {
		return 0, nil
	}

	var num int64
	if err := json.Unmarshal(raw, &num); err == nil {
		return num, nil
	}

	var rawString string
	if err := json.Unmarshal(raw, &rawString); err != nil {
		return 0, err
	}

	rawString = strings.TrimSpace(rawString)
	if rawString == "" {
		return 0, nil
	}

	return strconv.ParseInt(rawString, 10, 64)
}

// CustomerGateway matches the backend CustomerGateway proto message.
type CustomerGateway struct {
	ID             string     `json:"id"`
	Name           string     `json:"name"`
	Description    string     `json:"description"`
	PublicIP       string     `json:"publicIp"`
	VPNType        string     `json:"vpnType"`
	Status         string     `json:"status"`
	RemotePrefixes []string   `json:"remotePrefixes"`
	RemoteTunnelIP string     `json:"remoteTunnelIp"`
	LocalTunnelIP  string     `json:"localTunnelIp"`
	RoutingMode    string     `json:"routingMode"`
	BGPConfig      *BGPConfig `json:"bgpConfig,omitempty"`
	CreatedAt      string     `json:"createdAt"`
	ProjectID      string     `json:"projectId"`
	ZoneID         string     `json:"zoneId"`
}

// CreateCustomerGatewayRequest matches the backend CreateCustomerGatewayRequest proto message.
// project_id is passed via URL path.
type CreateCustomerGatewayRequest struct {
	Name           string     `json:"name"`
	Description    string     `json:"description,omitempty"`
	PublicIP       string     `json:"publicIp"`
	VPNType        string     `json:"vpnType"`
	RemotePrefixes []string   `json:"remotePrefixes"`
	RemoteTunnelIP string     `json:"remoteTunnelIp,omitempty"`
	LocalTunnelIP  string     `json:"localTunnelIp,omitempty"`
	RoutingMode    string     `json:"routingMode,omitempty"`
	BGPConfig      *BGPConfig `json:"bgpConfig,omitempty"`
}

// UpdateCustomerGatewayRequest matches the backend UpdateCustomerGatewayRequest proto message.
// project_id and id are passed via URL path.
type UpdateCustomerGatewayRequest struct {
	Name           string     `json:"name"`
	Description    string     `json:"description"`
	PublicIP       string     `json:"publicIp"`
	VPNType        string     `json:"vpnType"`
	RemotePrefixes []string   `json:"remotePrefixes"`
	RemoteTunnelIP string     `json:"remoteTunnelIp"`
	LocalTunnelIP  string     `json:"localTunnelIp"`
	RoutingMode    string     `json:"routingMode"`
	BGPConfig      *BGPConfig `json:"bgpConfig"`
}

// CustomerGatewayResponse matches the backend CustomerGatewayResponse proto message.
type CustomerGatewayResponse struct {
	CustomerGateway CustomerGateway `json:"customerGateway"`
}

// ListCustomerGatewaysResponse matches the backend ListCustomerGatewaysResponse proto message.
type ListCustomerGatewaysResponse struct {
	CustomerGateways []CustomerGateway `json:"customerGateways"`
}

// --- VPNaaS VPNPublicIP ---

// VPNPublicIP matches the backend VPNPublicIP proto message.
type VPNPublicIP struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	FloatingIP  string `json:"floatingIp"`
	Status      string `json:"status"`
	CreatedAt   string `json:"createdAt"`
	ProjectID   string `json:"projectId"`
	ZoneID      string `json:"zoneId"`
}

// CreateVPNPublicIPRequest matches the backend CreateVPNPublicIPRequest proto message.
// project_id is passed via URL path.
type CreateVPNPublicIPRequest struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

// UpdateVPNPublicIPRequest matches the backend UpdateVPNPublicIPRequest proto message.
// project_id and id are passed via URL path.
type UpdateVPNPublicIPRequest struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
}

// VPNPublicIPResponse matches the backend VPNPublicIPResponse proto message.
type VPNPublicIPResponse struct {
	VPNPublicIP VPNPublicIP `json:"vpnPublicIp"`
}

// ListVPNPublicIPsResponse matches the backend ListVPNPublicIPsResponse proto message.
type ListVPNPublicIPsResponse struct {
	VPNPublicIPs []VPNPublicIP `json:"vpnPublicIps"`
}

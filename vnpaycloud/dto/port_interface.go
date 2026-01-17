package dto

type Interface struct {
	PortState string    `json:"port_state"`
	FixedIPs  []FixedIP `json:"fixed_ips"`
	PortID    string    `json:"port_id"`
	NetID     string    `json:"net_id"`
	MACAddr   string    `json:"mac_addr"`
}

type GetInterfaceResponse struct {
	Interface Interface `json:"interfaceAttachments"`
}

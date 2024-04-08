package proxmox

import (
	"fmt"
	"net"
	"strings"
)

type APIReponse[T any] struct {
	Data T `json:"data"`
}

type TokenInfo struct {
	Expire int64 `json:"expire,omitempty"`
}

type TokenResponse struct {
	FullTokenID string    `json:"full-tokenid"`
	Info        TokenInfo `json:"info"`
	Value       string    `json:"value"`
}

func (t TokenResponse) String() string {
	return fmt.Sprintf("%s=%s", t.FullTokenID, t.Value)
}

type APITicket struct {
	Ticket              string `json:"ticket"`
	CSRFPreventionToken string
}

type UserResponse struct {
	Enabled bool                 `json:"enable"`
	Expire  int64                `json:"expire"`
	Tokens  map[string]TokenInfo `json:"tokens"`
}

type NetworkV4Config struct {
	GatewayIP string
	StaticIP  string
	Netmask   string
}

func (n NetworkV4Config) String() string {
	if strings.Contains(n.StaticIP, "/") {
		return fmt.Sprintf("gw=%s,ip=%s", n.GatewayIP, n.StaticIP)
	}

	ip := net.ParseIP(n.Netmask)
	if ip != nil {
		to4 := ip.To4()
		cidrSize, _ := net.IPv4Mask(to4[0], to4[1], to4[2], to4[3]).Size()
		return fmt.Sprintf("gw=%s,ip=%s/%d", n.GatewayIP, n.StaticIP, cidrSize)
	}

	return ""
}

type BoolAsInteger bool

type IntOrString int

type MachineConfig struct {
	ID          uint            `url:"-"`
	Node        string          `url:"-"`
	Name        string          `url:"name"`
	CPUs        uint            `url:"sockets"`
	CoresPerCPU uint            `url:"cores"`
	MemoryMiB   uint            `url:"memory"`
	Network     NetworkV4Config `url:"ipconfig[0]"`
	OnBoot      BoolAsInteger   `url:"onboot" urlformat:"int"`
}

type TaskSummary struct {
	UPID       string `json:"upid"`
	ExitStatus string `json:"exitstatus"`
	Status     string `json:"status"`
}

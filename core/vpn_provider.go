package core

// VPNProvider define a interface comum para todos os provedores de VPN
// VPNProvider defines the common interface for all VPN providers
// VPNProvider define la interfaz común para todos los proveedores de VPN
type VPNProvider interface {
	// Start inicia o serviço de VPN
	Start() error
	
	// Stop para o serviço de VPN
	Stop() error
	
	// IsRunning retorna se o serviço está em execução
	IsRunning() bool
	
	// AddPeer adiciona um peer à VPN
	AddPeer(peer TrustedPeer) error
	
	// RemovePeer remove um peer da VPN
	RemovePeer(nodeID string) error
	
	// GetConfig retorna a configuração atual da VPN
	GetConfig() *Config
	
	// SaveConfig salva a configuração em disco
	SaveConfig(path string) error
	
	// GetNodeInfo retorna as informações do nó local (nodeID, publicKey, virtualIP)
	GetNodeInfo() (string, string, string)
}

// Garantir que as implementações satisfazem a interface
var _ VPNProvider = (*VPNCore)(nil)
var _ VPNProvider = (*VPNCoreMulti)(nil)

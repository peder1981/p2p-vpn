package core

import (
	"fmt"
	"sync"
	"time"

	"github.com/p2p-vpn/p2p-vpn/platform"
)

// VPNCoreMulti representa o núcleo da VPN com suporte multiplataforma
// VPNCoreMulti represents the VPN core with multi-platform support
// VPNCoreMulti representa el núcleo de la VPN con soporte multiplataforma
type VPNCoreMulti struct {
	config     *Config
	listenPort int
	
	// Interface com o WireGuard
	interfaceName string
	platform      platform.VPNPlatform
	
	// Controle de status e sincronização
	running  bool
	mutex    sync.Mutex
	stopChan chan struct{}
}

// NewVPNCoreMulti cria uma nova instância do núcleo da VPN multiplataforma
// NewVPNCoreMulti creates a new instance of the multi-platform VPN core
// NewVPNCoreMulti crea una nueva instancia del núcleo de la VPN multiplataforma
func NewVPNCoreMulti(config *Config, listenPort int) (*VPNCoreMulti, error) {
	// Determinar o nome da interface
	interfaceName := "wg0"
	if config.InterfaceName != "" {
		interfaceName = config.InterfaceName
	}
	
	// Obter implementação da plataforma atual
	plat, err := platform.GetPlatform()
	if err != nil {
		return nil, fmt.Errorf("plataforma não suportada: %w", err)
	}
	
	fmt.Printf("Usando plataforma: %s\n", plat.Name())
	
	core := &VPNCoreMulti{
		config:        config,
		listenPort:    listenPort,
		interfaceName: interfaceName,
		platform:      plat,
		running:       false,
		stopChan:      make(chan struct{}),
	}
	
	// Verificar se a interface já existe
	exists, err := plat.GetInterfaceStatus(interfaceName)
	if err != nil {
		return nil, fmt.Errorf("erro ao verificar interface: %w", err)
	}
	
	if exists {
		// Interface já existe, remover para garantir configuração limpa
		fmt.Printf("Interface %s já existe, removendo...\n", interfaceName)
		if err := plat.RemoveWireGuardInterface(interfaceName); err != nil {
			return nil, fmt.Errorf("erro ao remover interface existente: %w", err)
		}
	}
	
	return core, nil
}

// Start inicia o serviço de VPN
// Start starts the VPN service
// Start inicia el servicio de VPN
func (v *VPNCoreMulti) Start() error {
	v.mutex.Lock()
	defer v.mutex.Unlock()
	
	if v.running {
		return fmt.Errorf("o serviço de VPN já está em execução")
	}
	
	fmt.Printf("Criando interface WireGuard: %s\n", v.interfaceName)
	
	// 1. Criar a interface WireGuard
	if err := v.platform.CreateWireGuardInterface(v.interfaceName, v.listenPort, v.config.PrivateKey); err != nil {
		return fmt.Errorf("erro ao criar interface WireGuard: %w", err)
	}
	
	// 2. Configurar endereço IP
	parts := ParseCIDR(v.config.VirtualIP, v.config.VirtualCIDR)
	if err := v.platform.ConfigureInterfaceAddress(v.interfaceName, v.config.VirtualIP, parts.MaskSize); err != nil {
		// Rollback em caso de erro
		v.platform.RemoveWireGuardInterface(v.interfaceName)
		return fmt.Errorf("erro ao configurar endereço IP: %w", err)
	}
	
	// 3. Configurar roteamento
	if err := v.platform.ConfigureRouting(v.interfaceName, v.config.VirtualCIDR); err != nil {
		// Rollback em caso de erro
		v.platform.RemoveWireGuardInterface(v.interfaceName)
		return fmt.Errorf("erro ao configurar roteamento: %w", err)
	}
	
	// 4. Adicionar peers configurados
	if len(v.config.TrustedPeers) > 0 {
		for _, peer := range v.config.TrustedPeers {
			if err := v.addPeer(peer); err != nil {
				fmt.Printf("Aviso: erro ao adicionar peer %s: %v\n", peer.NodeID, err)
			}
		}
	}
	
	fmt.Printf("Interface WireGuard %s configurada e ativada com sucesso\n", v.interfaceName)
	
	v.running = true
	
	// Iniciar goroutine para monitoramento
	go v.monitorRoutine()
	
	return nil
}

// Stop para o serviço de VPN
// Stop stops the VPN service
// Stop detiene el servicio de VPN
func (v *VPNCoreMulti) Stop() error {
	v.mutex.Lock()
	defer v.mutex.Unlock()
	
	if !v.running {
		return nil // Já está parado
	}
	
	// Sinalizar para as goroutines pararem
	close(v.stopChan)
	
	fmt.Printf("Desativando interface WireGuard %s...\n", v.interfaceName)
	
	// Remover a interface
	if err := v.platform.RemoveWireGuardInterface(v.interfaceName); err != nil {
		fmt.Printf("Aviso: erro ao remover interface: %v\n", err)
	}
	
	v.running = false
	fmt.Println("Serviço de VPN encerrado com sucesso")
	
	return nil
}

// IsRunning retorna se o serviço está em execução
// IsRunning returns if the service is running
// IsRunning devuelve si el servicio está en ejecución
func (v *VPNCoreMulti) IsRunning() bool {
	v.mutex.Lock()
	defer v.mutex.Unlock()
	
	return v.running
}

// AddPeer adiciona um peer à VPN
// AddPeer adds a peer to the VPN
// AddPeer añade un peer a la VPN
func (v *VPNCoreMulti) AddPeer(peer TrustedPeer) error {
	v.mutex.Lock()
	defer v.mutex.Unlock()
	
	// Adicionar ao registro de peers
	v.config.AddTrustedPeer(peer)
	
	// Se o serviço não estiver em execução, apenas adicionar à configuração
	if !v.running {
		return nil
	}
	
	// Adicionar o peer à interface
	return v.addPeer(peer)
}

// RemovePeer remove um peer da VPN
// RemovePeer removes a peer from the VPN
// RemovePeer elimina un peer de la VPN
func (v *VPNCoreMulti) RemovePeer(nodeID string) error {
	v.mutex.Lock()
	defer v.mutex.Unlock()
	
	// Buscar peer pelo nodeID
	var peerToRemove *TrustedPeer
	for _, peer := range v.config.TrustedPeers {
		if peer.NodeID == nodeID {
			peerCopy := peer
			peerToRemove = &peerCopy
			break
		}
	}
	
	if peerToRemove == nil {
		return fmt.Errorf("peer %s não encontrado", nodeID)
	}
	
	// Remover do registro de peers
	v.config.RemoveTrustedPeer(nodeID)
	
	// Se o serviço não estiver em execução, apenas remover da configuração
	if !v.running {
		return nil
	}
	
	// Remover o peer da interface
	return v.platform.RemovePeer(v.interfaceName, peerToRemove.PublicKey)
}

// GetConfig retorna a configuração atual da VPN
// GetConfig returns the current VPN configuration
// GetConfig devuelve la configuración actual de la VPN
func (v *VPNCoreMulti) GetConfig() *Config {
	return v.config
}

// SaveConfig salva a configuração em disco
// SaveConfig saves the configuration to disk
// SaveConfig guarda la configuración en disco
func (v *VPNCoreMulti) SaveConfig(path string) error {
	return v.config.SaveConfig(path)
}

// addPeer adiciona um peer à interface WireGuard
// helper interno, já assume que o mutex está bloqueado
func (v *VPNCoreMulti) addPeer(peer TrustedPeer) error {
	if !v.running {
		return fmt.Errorf("o serviço de VPN não está em execução")
	}
	
	// Construir AllowedIPs
	allowedIPs := peer.VirtualIP + "/32"
	if len(peer.AllowedIPs) > 0 {
		allowedIPs = peer.AllowedIPs[0]
	}
	
	// Usar o primeiro endpoint, se disponível
	endpoint := ""
	if len(peer.Endpoints) > 0 {
		endpoint = peer.Endpoints[0]
		// Verificar se a porta está especificada
		if endpoint != "" && !ContainsPort(endpoint) {
			endpoint = endpoint + ":51820"
		}
	}
	
	// Adicionar peer usando a implementação de plataforma
	err := v.platform.AddPeer(v.interfaceName, peer.PublicKey, allowedIPs, endpoint, peer.KeepAlive)
	if err != nil {
		return fmt.Errorf("erro ao adicionar peer: %w", err)
	}
	
	fmt.Printf("Peer %s (%s) adicionado com sucesso\n", peer.NodeID, peer.VirtualIP)
	return nil
}

// Rotina de monitoramento
func (v *VPNCoreMulti) monitorRoutine() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			// Verificar status da interface
			v.mutex.Lock()
			if v.running {
				isActive, err := v.platform.GetInterfaceStatus(v.interfaceName)
				if err != nil {
					fmt.Printf("Erro ao verificar status da interface: %v\n", err)
				} else if !isActive {
					fmt.Printf("Interface %s não está ativa! Tentando reiniciar...\n", v.interfaceName)
					// Aqui poderíamos implementar tentativas de recuperação automática
				}
			}
			v.mutex.Unlock()
			
		case <-v.stopChan:
			return
		}
	}
}

// ParseCIDR é um helper para extrair partes de um CIDR
type CIDRParts struct {
	Network  string
	IP       string
	Mask     string
	MaskSize string
}

// ParseCIDR extrai partes de um endereço IP e máscara
func ParseCIDR(ip, cidr string) CIDRParts {
	// Implementação simplificada - em uma versão real faríamos um parse completo
	// Assumindo que a máscara é do tipo /24
	return CIDRParts{
		Network:  cidr,
		IP:       ip,
		Mask:     "255.255.255.0",  // Assumindo /24
		MaskSize: "24",            // Assumindo /24
	}
}

// ContainsPort verifica se um endereço contém especificação de porta
func ContainsPort(addr string) bool {
	return len(addr) > 0 && addr[len(addr)-1] >= '0' && addr[len(addr)-1] <= '9'
}

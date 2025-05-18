package core

import (
	"encoding/base64"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/vishvananda/netlink"
	"golang.zx2c4.com/wireguard/wgctrl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

// VPNCore é o núcleo da VPN, responsável por gerenciar as interfaces WireGuard
// VPNCore is the VPN core responsible for managing WireGuard interfaces
// VPNCore es el núcleo de la VPN responsable de gestionar las interfaces WireGuard
type VPNCore struct {
	config     *Config
	listenPort int
	
	// Interface com o WireGuard
	interfaceName string         // Nome da interface WireGuard (ex: wg0)
	linkAttrs     *netlink.LinkAttrs // Atributos da interface
	wgClient      *wgctrl.Client    // Cliente para controlar a interface WireGuard
	
	// Controle de status e sincronização
	running  bool
	mutex    sync.Mutex
	stopChan chan struct{}
}

// NewVPNCore cria uma nova instância do núcleo da VPN
// NewVPNCore creates a new instance of the VPN core
// NewVPNCore crea una nueva instancia del núcleo de la VPN
func NewVPNCore(config *Config, listenPort int) (*VPNCore, error) {
	// Determinar o nome da interface
	interfaceName := "wg0"
	if config.InterfaceName != "" {
		interfaceName = config.InterfaceName
	}
	
	core := &VPNCore{
		config:        config,
		listenPort:    listenPort,
		interfaceName: interfaceName,
		running:       false,
		stopChan:      make(chan struct{}),
	}
	
	// Criar cliente WireGuard para configurar a interface
	wgClient, err := wgctrl.New()
	if err != nil {
		return nil, fmt.Errorf("erro ao criar cliente WireGuard: %w", err)
	}
	core.wgClient = wgClient
	
	// Verificar se a interface já existe
	links, err := netlink.LinkList()
	for _, link := range links {
		if link.Attrs().Name == interfaceName {
			// Interface já existe, vamos removê-la para garantir uma configuração limpa
			fmt.Printf("Interface %s já existe, removendo...\n", interfaceName)
			if err := netlink.LinkDel(link); err != nil {
				return nil, fmt.Errorf("erro ao remover interface existente: %w", err)
			}
			break
		}
	}
	
	return core, nil
}

// Start inicia o serviço de VPN
// Start starts the VPN service
// Start inicia el servicio de VPN
func (v *VPNCore) Start() error {
	v.mutex.Lock()
	defer v.mutex.Unlock()
	
	if v.running {
		return fmt.Errorf("o serviço de VPN já está em execução")
	}
	
	fmt.Printf("Criando interface WireGuard: %s\n", v.interfaceName)
	
	// 1. Criar a interface WireGuard usando netlink
	linkAttrs := netlink.NewLinkAttrs()
	linkAttrs.Name = v.interfaceName
	linkAttrs.MTU = 1420 // MTU padrão para WireGuard
	if v.config.MTU > 0 {
		linkAttrs.MTU = v.config.MTU
	}
	
	// Criar link WireGuard
	wgLink := &netlink.GenericLink{
		LinkAttrs: linkAttrs,
		LinkType:  "wireguard",
	}
	
	// Adicionar a interface ao sistema
	if err := netlink.LinkAdd(wgLink); err != nil {
		return fmt.Errorf("erro ao criar interface WireGuard: %w", err)
	}
	v.linkAttrs = &linkAttrs
	
	// 2. Configurar a interface WireGuard com as chaves
	// Converter a chave privada de base64 para bytes
	privateKeyBytes, err := base64.StdEncoding.DecodeString(v.config.PrivateKey)
	if err != nil {
		return fmt.Errorf("erro ao decodificar chave privada: %w", err)
	}
	
	// Criar configuração WireGuard
	var privateKey wgtypes.Key
	copy(privateKey[:], privateKeyBytes)
	
	wgConfig := wgtypes.Config{
		PrivateKey:   &privateKey,
		ListenPort:   &v.listenPort,
		ReplacePeers: true,
	}
	
	// Aplicar configuração à interface
	if err := v.wgClient.ConfigureDevice(v.interfaceName, wgConfig); err != nil {
		// Remover a interface em caso de erro
		netlink.LinkDel(wgLink)
		return fmt.Errorf("erro ao configurar interface WireGuard: %w", err)
	}
	
	// 3. Configurar o endereço IP da interface
	ipAddr, ipNet, err := net.ParseCIDR(fmt.Sprintf("%s/32", v.config.VirtualIP))
	if err != nil {
		// Remover a interface em caso de erro
		netlink.LinkDel(wgLink)
		return fmt.Errorf("erro ao analisar endereço IP virtual: %w", err)
	}
	
	// Adicionar endereço à interface
	ipConfig := &netlink.Addr{
		IPNet: &net.IPNet{
			IP:   ipAddr,
			Mask: ipNet.Mask,
		},
	}
	
	if err := netlink.AddrAdd(wgLink, ipConfig); err != nil {
		// Remover a interface em caso de erro
		netlink.LinkDel(wgLink)
		return fmt.Errorf("erro ao configurar endereço IP: %w", err)
	}
	
	// 4. Ativar a interface
	if err := netlink.LinkSetUp(wgLink); err != nil {
		// Remover a interface em caso de erro
		netlink.LinkDel(wgLink)
		return fmt.Errorf("erro ao ativar interface: %w", err)
	}
	
	// 5. Configurar rotas para a rede virtual
	_, virtualNet, err := net.ParseCIDR(v.config.VirtualCIDR)
	if err != nil {
		// Remover a interface em caso de erro
		netlink.LinkDel(wgLink)
		return fmt.Errorf("erro ao analisar CIDR virtual: %w", err)
	}
	
	// Adicionar rota para a rede virtual
	virtualRoute := netlink.Route{
		Dst:       virtualNet,
		LinkIndex: wgLink.Attrs().Index,
	}
	
	if err := netlink.RouteAdd(&virtualRoute); err != nil {
		// Ignorar erro se a rota já existir
		if !strings.Contains(err.Error(), "file exists") {
			// Remover a interface em caso de erro
			netlink.LinkDel(wgLink)
			return fmt.Errorf("erro ao adicionar rota: %w", err)
		}
	}
	
	// 6. Adicionar peers configurados
	if len(v.config.TrustedPeers) > 0 {
		for _, peer := range v.config.TrustedPeers {
			if err := v.addWireGuardPeer(peer); err != nil {
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
func (v *VPNCore) Stop() error {
	v.mutex.Lock()
	defer v.mutex.Unlock()
	
	if !v.running {
		return nil // Já está parado
	}
	
	// Sinalizar para as goroutines pararem
	close(v.stopChan)
	
	fmt.Printf("Desativando interface WireGuard %s...\n", v.interfaceName)
	
	// Obter link para a interface
	link, err := netlink.LinkByName(v.interfaceName)
	if err != nil {
		fmt.Printf("Aviso: interface %s não encontrada para remoção: %v\n", v.interfaceName, err)
	} else {
		// Desativar a interface
		if err := netlink.LinkSetDown(link); err != nil {
			fmt.Printf("Aviso: erro ao desativar interface: %v\n", err)
		}
		
		// Remover a interface
		if err := netlink.LinkDel(link); err != nil {
			fmt.Printf("Aviso: erro ao remover interface: %v\n", err)
		}
	}
	
	// Fechar o cliente WireGuard
	if v.wgClient != nil {
		v.wgClient.Close()
		v.wgClient = nil
	}
	
	v.running = false
	fmt.Println("Serviço de VPN encerrado com sucesso")
	
	return nil
}

// IsRunning verifica se o serviço está em execução
func (v *VPNCore) IsRunning() bool {
	v.mutex.Lock()
	defer v.mutex.Unlock()
	return v.running
}

// AddPeer adiciona um novo peer à configuração do WireGuard
func (v *VPNCore) AddPeer(peer TrustedPeer) error {
	v.mutex.Lock()
	defer v.mutex.Unlock()
	
	// Adicionar à configuração
	v.config.AddTrustedPeer(peer)
	
	// Se estiver em execução, atualizar a configuração WireGuard
	if v.running {
		// Aqui atualizaríamos a configuração do WireGuard em tempo real
		fmt.Printf("Simulando a adição do peer %s à interface WireGuard...\n", peer.NodeID)
	}
	
	return nil
}

// RemovePeer remove um peer da configuração do WireGuard
func (v *VPNCore) RemovePeer(nodeID string) error {
	v.mutex.Lock()
	defer v.mutex.Unlock()
	
	// Remover da configuração
	if !v.config.RemoveTrustedPeer(nodeID) {
		return fmt.Errorf("peer %s não encontrado", nodeID)
	}
	
	// Se estiver em execução, atualizar a configuração WireGuard
	if v.running {
		// Aqui atualizaríamos a configuração do WireGuard em tempo real
		fmt.Printf("Simulando a remoção do peer %s da interface WireGuard...\n", nodeID)
	}
	
	return nil
}

// UpdatePeerEndpoint atualiza os endpoints de um peer
func (v *VPNCore) UpdatePeerEndpoint(nodeID string, endpoint string) error {
	v.mutex.Lock()
	defer v.mutex.Unlock()
	
	// Encontrar o peer
	var targetPeer *TrustedPeer
	for i := range v.config.TrustedPeers {
		if v.config.TrustedPeers[i].NodeID == nodeID {
			targetPeer = &v.config.TrustedPeers[i]
			break
		}
	}
	
	if targetPeer == nil {
		return fmt.Errorf("peer %s não encontrado", nodeID)
	}
	
	// Verificar se o endpoint já existe
	endpointExists := false
	for _, ep := range targetPeer.Endpoints {
		if ep == endpoint {
			endpointExists = true
			break
		}
	}
	
	// Adicionar novo endpoint se necessário
	if !endpointExists {
		targetPeer.Endpoints = append(targetPeer.Endpoints, endpoint)
	}
	
	// Atualizar lastSeen
	targetPeer.LastSeen = time.Now().Unix()
	
	// Se estiver em execução, atualizar a configuração WireGuard
	if v.running {
		// Aqui atualizaríamos a configuração do WireGuard em tempo real
		fmt.Printf("Simulando a atualização do endpoint %s para o peer %s...\n", endpoint, nodeID)
	}
	
	return nil
}

// GetPeers retorna a lista de peers configurados
func (v *VPNCore) GetPeers() []TrustedPeer {
	v.mutex.Lock()
	defer v.mutex.Unlock()
	
	// Fazer uma cópia para evitar problemas de concorrência
	peers := make([]TrustedPeer, len(v.config.TrustedPeers))
	copy(peers, v.config.TrustedPeers)
	
	return peers
}

// GetNodeInfo retorna as informações do nó local
func (v *VPNCore) GetNodeInfo() (string, string, string) {
	v.mutex.Lock()
	defer v.mutex.Unlock()
	
	return v.config.NodeID, v.config.PublicKey, v.config.VirtualIP
}

// GetPeerByID retorna um peer específico com base no seu ID
// GetPeerByID returns a specific peer based on its ID
// GetPeerByID devuelve un peer específico basado en su ID
func (v *VPNCore) GetPeerByID(peerID string) *TrustedPeer {
	v.mutex.Lock()
	defer v.mutex.Unlock()
	
	for i := range v.config.TrustedPeers {
		if v.config.TrustedPeers[i].NodeID == peerID {
			// Retornar uma cópia do peer para evitar problemas de concorrência
			peerCopy := v.config.TrustedPeers[i]
			return &peerCopy
		}
	}
	
	return nil
}

// ProcessIncomingData processa dados recebidos de um peer específico
// ProcessIncomingData processes incoming data from a specific peer
// ProcessIncomingData procesa datos entrantes de un peer específico
func (v *VPNCore) ProcessIncomingData(peerID string, data []byte) error {
	v.mutex.Lock()
	defer v.mutex.Unlock()
	
	// Verificar se o peer existe
	var foundPeer bool
	for i := range v.config.TrustedPeers {
		if v.config.TrustedPeers[i].NodeID == peerID {
			foundPeer = true
			break
		}
	}
	
	if !foundPeer {
		return fmt.Errorf("peer %s não encontrado", peerID)
	}
	
	// Aqui seria implementado o processamento real dos dados
	// Este é apenas um stub para o método
	fmt.Printf("Processando %d bytes de dados do peer %s\n", len(data), peerID)
	
	return nil
}

// Rotina de monitoramento
func (v *VPNCore) monitorRoutine() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			// Aqui implementaríamos a verificação de status dos peers
			// e outras tarefas de manutenção
			v.mutex.Lock()
			if v.running {
				// Simulação de monitoramento
				fmt.Println("Simulando monitoramento da interface WireGuard...")
			}
			v.mutex.Unlock()
			
		case <-v.stopChan:
			return
		}
	}
}

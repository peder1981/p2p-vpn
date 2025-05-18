package discovery

import (
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/p2p-vpn/p2p-vpn/core"
)

// PeerDiscovery gerencia a descoberta de peers na rede
type PeerDiscovery struct {
	config      *core.Config
	vpnCore     core.VPNProvider
	listenPort  int
	
	// Informações do nó local
	nodeID      string
	publicKey   string
	virtualIP   string
	
	// Para comunicação via UDP
	udpConn     *net.UDPConn
	
	// Controle de estado
	running     bool
	mutex       sync.Mutex
	stopChan    chan struct{}
	
	// Cache de nós conhecidos
	knownNodes  map[string]*PeerInfo
	nodesMutex  sync.RWMutex
}

// PeerInfo armazena informações sobre um peer descoberto
type PeerInfo struct {
	NodeID      string
	PublicKey   string
	VirtualIP   string
	Endpoints   []string
	LastSeen    time.Time
}

// NewPeerDiscovery cria uma nova instância do sistema de descoberta
func NewPeerDiscovery(config *core.Config, listenPort int, vpnCore core.VPNProvider) (*PeerDiscovery, error) {
	// Obter informações do nó local do VPNCore
	nodeID, publicKey, virtualIP := vpnCore.GetNodeInfo()
	
	discovery := &PeerDiscovery{
		config:     config,
		vpnCore:    vpnCore,
		listenPort: listenPort,
		nodeID:     nodeID,
		publicKey:  publicKey,
		virtualIP:  virtualIP,
		running:    false,
		stopChan:   make(chan struct{}),
		knownNodes: make(map[string]*PeerInfo),
	}
	
	return discovery, nil
}

// Start inicia o serviço de descoberta
func (p *PeerDiscovery) Start() error {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	
	if p.running {
		return fmt.Errorf("o serviço de descoberta já está em execução")
	}
	
	// Iniciar o listener UDP
	addr := &net.UDPAddr{
		IP:   net.IPv4zero,
		Port: p.listenPort,
	}
	
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return fmt.Errorf("erro ao abrir porta UDP para descoberta: %w", err)
	}
	
	p.udpConn = conn
	p.running = true
	
	// Iniciar goroutines para recebimento de mensagens e anúncios periódicos
	go p.receiveMessages()
	go p.announceRoutine()
	go p.maintenanceRoutine()
	
	return nil
}

// Stop para o serviço de descoberta
func (p *PeerDiscovery) Stop() error {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	
	if !p.running {
		return nil // Já está parado
	}
	
	// Sinalizar para as goroutines pararem
	close(p.stopChan)
	
	// Fechar a conexão UDP
	if p.udpConn != nil {
		p.udpConn.Close()
		p.udpConn = nil
	}
	
	p.running = false
	
	return nil
}

// IsRunning verifica se o serviço está em execução
func (p *PeerDiscovery) IsRunning() bool {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	return p.running
}

// receiveMessages processa mensagens recebidas via UDP
func (p *PeerDiscovery) receiveMessages() {
	buffer := make([]byte, 2048)
	
	for {
		select {
		case <-p.stopChan:
			return
		default:
			// Configurar timeout para não bloquear indefinidamente
			p.udpConn.SetReadDeadline(time.Now().Add(1 * time.Second))
			
			n, addr, err := p.udpConn.ReadFromUDP(buffer)
			if err != nil {
				// Ignorar erros de timeout
				if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					continue
				}
				
				fmt.Printf("Erro ao ler mensagem UDP: %v\n", err)
				continue
			}
			
			// Processar a mensagem recebida
			p.handleMessage(buffer[:n], addr)
		}
	}
}

// handleMessage processa uma mensagem recebida do serviço de descoberta
func (p *PeerDiscovery) handleMessage(data []byte, addr *net.UDPAddr) {
	// Aqui implementaríamos a análise da mensagem e o processamento
	// Por enquanto, apenas simulamos o processamento
	
	// Simular extração de informações da mensagem
	receivedNodeID := "node-simulated"
	receivedPublicKey := "key-simulated"
	receivedVirtualIP := "10.0.0.123"
	
	fmt.Printf("Recebida mensagem de descoberta do nó %s (%s)\n", receivedNodeID, addr.String())
	
	// Atualizar a informação do peer
	p.updatePeerInfo(receivedNodeID, receivedPublicKey, receivedVirtualIP, addr.String())
}

// updatePeerInfo atualiza as informações de um peer conhecido
func (p *PeerDiscovery) updatePeerInfo(nodeID, publicKey, virtualIP, endpoint string) {
	p.nodesMutex.Lock()
	defer p.nodesMutex.Unlock()
	
	// Verificar se o nó já é conhecido
	peer, exists := p.knownNodes[nodeID]
	if !exists {
		// Novo nó descoberto
		peer = &PeerInfo{
			NodeID:    nodeID,
			PublicKey: publicKey,
			VirtualIP: virtualIP,
			Endpoints: []string{endpoint},
			LastSeen:  time.Now(),
		}
		p.knownNodes[nodeID] = peer
		
		fmt.Printf("Novo peer descoberto: %s (%s)\n", nodeID, endpoint)
	} else {
		// Atualizar informações do nó existente
		peer.LastSeen = time.Now()
		
		// Verificar se o endpoint já existe
		endpointExists := false
		for _, ep := range peer.Endpoints {
			if ep == endpoint {
				endpointExists = true
				break
			}
		}
		
		// Adicionar novo endpoint se necessário
		if !endpointExists {
			peer.Endpoints = append(peer.Endpoints, endpoint)
		}
	}
	
	// Atualizar o endpoint no VPNCore para configuração do WireGuard
	trustedPeer := core.TrustedPeer{
		NodeID:    nodeID,
		PublicKey: publicKey,
		VirtualIP: virtualIP,
		Endpoints: []string{endpoint},
		LastSeen:  time.Now().Unix(),
	}
	
	p.vpnCore.AddPeer(trustedPeer)
}

// announceRoutine envia anúncios periódicos para descoberta de peers
func (p *PeerDiscovery) announceRoutine() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	
	// Fazer um anúncio inicial
	p.sendAnnouncement()
	
	for {
		select {
		case <-ticker.C:
			p.sendAnnouncement()
		case <-p.stopChan:
			return
		}
	}
}

// sendAnnouncement envia um anúncio para os peers conhecidos e para endereços de rendezvous
func (p *PeerDiscovery) sendAnnouncement() {
	p.mutex.Lock()
	running := p.running
	p.mutex.Unlock()
	
	if !running {
		return
	}
	
	// Aqui construiríamos e enviaríamos a mensagem de anúncio
	// Por enquanto, apenas simulamos o envio
	fmt.Println("Simulando envio de anúncio para descoberta de peers...")
	
	// Enviar para os rendezvous servers (seriam servidores STUN ou outro mecanismo)
	// Isso permitiria a descoberta mesmo através de NATs
	
	// Enviar para os peers conhecidos para manter as conexões ativas
	p.nodesMutex.RLock()
	defer p.nodesMutex.RUnlock()
	
	for nodeID, peer := range p.knownNodes {
		// Ignorar nós que não foram vistos recentemente
		if time.Since(peer.LastSeen) > 1*time.Hour {
			continue
		}
		
		// Enviar para todos os endpoints conhecidos do peer
		for _, endpoint := range peer.Endpoints {
			// Simular envio
			fmt.Printf("Simulando envio de anúncio para o peer %s em %s\n", nodeID, endpoint)
		}
	}
}

// maintenanceRoutine executa tarefas de manutenção periódicas
func (p *PeerDiscovery) maintenanceRoutine() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			p.cleanupStaleNodes()
		case <-p.stopChan:
			return
		}
	}
}

// cleanupStaleNodes remove nós que não foram vistos recentemente
func (p *PeerDiscovery) cleanupStaleNodes() {
	p.nodesMutex.Lock()
	defer p.nodesMutex.Unlock()
	
	for nodeID, peer := range p.knownNodes {
		// Remover nós que não foram vistos há mais de 24 horas
		if time.Since(peer.LastSeen) > 24*time.Hour {
			delete(p.knownNodes, nodeID)
			fmt.Printf("Removendo peer inativo: %s (último contato: %v)\n", 
				nodeID, peer.LastSeen)
			
			// Remover também do VPNCore
			p.vpnCore.RemovePeer(nodeID)
		}
	}
}

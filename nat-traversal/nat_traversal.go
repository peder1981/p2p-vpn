package nattraversal

import (
	"fmt"
	"net"
	"sync"
	"time"
)

// Técnicas de NAT traversal suportadas
const (
	TechniqueHolePunching = "hole-punching"
	TechniqueUPnP         = "upnp"
	TechniqueSTUN         = "stun"
	TechniqueTURN         = "turn" // Fallback quando métodos diretos falham
)

// STUNServer representa um servidor STUN para descoberta de endereço público
type STUNServer struct {
	Address string
	Port    int
}

// DefaultSTUNServers é uma lista de servidores STUN públicos
var DefaultSTUNServers = []STUNServer{
	{Address: "stun.l.google.com", Port: 19302},
	{Address: "stun1.l.google.com", Port: 19302},
	{Address: "stun2.l.google.com", Port: 19302},
	{Address: "stun.ekiga.net", Port: 3478},
}

// NATInfo armazena informações sobre o tipo de NAT detectado
type NATInfo struct {
	Type            string    // "open", "full-cone", "restricted-cone", "port-restricted", "symmetric"
	PublicIP        string    // Endereço IP público
	PublicPort      int       // Porta pública mapeada
	LastUpdate      time.Time // Última vez que a informação foi atualizada
	MappingLifetime int       // Tempo de vida estimado do mapeamento em segundos
}

// NATTraversal gerencia técnicas de NAT traversal
type NATTraversal struct {
	stunServers     []STUNServer
	localPort       int
	
	// Informações sobre o NAT local
	natInfo         NATInfo
	natInfoMutex    sync.RWMutex
	
	// UPnP e outras técnicas
	useUPnP         bool
	upnpMappingID   string
	
	// Controle de estado
	running         bool
	mutex           sync.Mutex
	stopChan        chan struct{}
}

// NewNATTraversal cria uma nova instância do sistema de NAT traversal
func NewNATTraversal(localPort int) *NATTraversal {
	return &NATTraversal{
		stunServers: DefaultSTUNServers,
		localPort:   localPort,
		useUPnP:     true,
		running:     false,
		stopChan:    make(chan struct{}),
	}
}

// Start inicia o serviço de NAT traversal
func (n *NATTraversal) Start() error {
	n.mutex.Lock()
	defer n.mutex.Unlock()
	
	if n.running {
		return fmt.Errorf("o serviço de NAT traversal já está em execução")
	}
	
	// Iniciar a detecção de NAT
	go n.detectNATType()
	
	// Tentar configurar UPnP se disponível
	if n.useUPnP {
		go n.setupUPnP()
	}
	
	n.running = true
	
	// Iniciar rotina de manutenção
	go n.maintenanceRoutine()
	
	return nil
}

// Stop para o serviço de NAT traversal
func (n *NATTraversal) Stop() error {
	n.mutex.Lock()
	defer n.mutex.Unlock()
	
	if !n.running {
		return nil // Já está parado
	}
	
	// Sinalizar para as goroutines pararem
	close(n.stopChan)
	
	// Remover mapeamentos UPnP
	if n.useUPnP && n.upnpMappingID != "" {
		n.removeUPnPMapping()
	}
	
	n.running = false
	
	return nil
}

// detectNATType detecta o tipo de NAT usando servidores STUN
func (n *NATTraversal) detectNATType() {
	// Aqui implementaríamos a lógica real de detecção de NAT com STUN
	// Por enquanto, apenas simulamos a detecção
	
	fmt.Println("Iniciando detecção de NAT...")
	
	// Simular um atraso para a detecção
	time.Sleep(1 * time.Second)
	
	// Atualizar com valores simulados
	n.natInfoMutex.Lock()
	n.natInfo = NATInfo{
		Type:            "restricted-cone", // Valor simulado
		PublicIP:        "203.0.113.45",    // IP simulado
		PublicPort:      12345,             // Porta simulada
		LastUpdate:      time.Now(),
		MappingLifetime: 300, // 5 minutos (simulado)
	}
	n.natInfoMutex.Unlock()
	
	fmt.Printf("NAT detectado: tipo=%s, IP público=%s:%d\n", 
		n.natInfo.Type, n.natInfo.PublicIP, n.natInfo.PublicPort)
}

// GetPublicEndpoint retorna o endpoint público detectado
func (n *NATTraversal) GetPublicEndpoint() (string, int, error) {
	n.natInfoMutex.RLock()
	defer n.natInfoMutex.RUnlock()
	
	if n.natInfo.PublicIP == "" {
		return "", 0, fmt.Errorf("ainda não foi possível detectar o endereço público")
	}
	
	// Verificar se a informação é atual
	if time.Since(n.natInfo.LastUpdate) > time.Duration(n.natInfo.MappingLifetime)*time.Second {
		return n.natInfo.PublicIP, n.natInfo.PublicPort, 
			fmt.Errorf("informação de NAT pode estar desatualizada")
	}
	
	return n.natInfo.PublicIP, n.natInfo.PublicPort, nil
}

// setupUPnP tenta configurar mapeamento de portas UPnP no roteador
func (n *NATTraversal) setupUPnP() {
	// Aqui implementaríamos a configuração real de UPnP
	// Usando bibliotecas como github.com/huin/goupnp
	// Por enquanto, apenas simulamos a configuração
	
	fmt.Println("Tentando configurar mapeamento UPnP...")
	
	// Simular um atraso para a configuração
	time.Sleep(2 * time.Second)
	
	// Simulação de sucesso
	n.mutex.Lock()
	n.upnpMappingID = "simulado-12345"
	n.mutex.Unlock()
	
	fmt.Println("Mapeamento UPnP configurado com sucesso!")
}

// removeUPnPMapping remove o mapeamento UPnP do roteador
func (n *NATTraversal) removeUPnPMapping() {
	// Aqui implementaríamos a remoção real do mapeamento UPnP
	// Por enquanto, apenas simulamos a remoção
	
	fmt.Println("Removendo mapeamento UPnP...")
	
	// Simular um atraso para a remoção
	time.Sleep(500 * time.Millisecond)
	
	n.mutex.Lock()
	n.upnpMappingID = ""
	n.mutex.Unlock()
	
	fmt.Println("Mapeamento UPnP removido com sucesso!")
}

// FacilitateConnection tenta facilitar uma conexão com um peer remoto
func (n *NATTraversal) FacilitateConnection(remoteIP string, remotePort int) error {
	fmt.Printf("Tentando facilitar conexão com %s:%d...\n", remoteIP, remotePort)
	
	// Obter informações do NAT local
	n.natInfoMutex.RLock()
	localNATType := n.natInfo.Type
	n.natInfoMutex.RUnlock()
	
	// Aqui implementaríamos diferentes estratégias com base no tipo de NAT
	switch localNATType {
	case "open":
		fmt.Println("NAT aberto, conexão direta possível")
		return nil
		
	case "full-cone":
		fmt.Println("NAT full-cone, tentando conexão direta")
		return n.directConnection(remoteIP, remotePort)
		
	case "restricted-cone", "port-restricted":
		fmt.Println("NAT restrito, tentando hole punching")
		return n.holePunching(remoteIP, remotePort)
		
	case "symmetric":
		fmt.Println("NAT simétrico, conexão direta pode não ser possível")
		return n.tryRelayIfNeeded(remoteIP, remotePort)
		
	default:
		return fmt.Errorf("tipo de NAT desconhecido: %s", localNATType)
	}
}

// directConnection implementa conexão direta com o peer remoto
func (n *NATTraversal) directConnection(remoteIP string, remotePort int) error {
	// Simulação de conexão direta
	fmt.Printf("Simulando conexão direta com %s:%d\n", remoteIP, remotePort)
	return nil
}

// holePunching implementa a técnica de hole punching para atravessar NAT
func (n *NATTraversal) holePunching(remoteIP string, remotePort int) error {
	// Simulação de hole punching
	fmt.Printf("Simulando UDP hole punching com %s:%d\n", remoteIP, remotePort)
	
	// Na implementação real, enviaríamos pacotes UDP para criar o "buraco" no NAT
	// e permitir a comunicação bidirecional
	
	return nil
}

// tryRelayIfNeeded tenta usar um servidor TURN como relay se necessário
func (n *NATTraversal) tryRelayIfNeeded(remoteIP string, remotePort int) error {
	// Primeiro tentar hole punching
	err := n.holePunching(remoteIP, remotePort)
	if err == nil {
		return nil
	}
	
	// Se falhar, usar relay TURN
	fmt.Println("Hole punching falhou, recorrendo a servidor relay TURN")
	
	// Simulação de conexão via relay
	fmt.Printf("Simulando conexão via relay TURN com %s:%d\n", remoteIP, remotePort)
	
	return nil
}

// maintenanceRoutine executa tarefas de manutenção periódicas
func (n *NATTraversal) maintenanceRoutine() {
	ticker := time.NewTicker(2 * time.Minute)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			// Atualizar informações de NAT periodicamente
			go n.detectNATType()
			
			// Renovar mapeamentos UPnP se necessário
			if n.useUPnP && n.upnpMappingID != "" {
				go n.setupUPnP()
			}
			
		case <-n.stopChan:
			return
		}
	}
}

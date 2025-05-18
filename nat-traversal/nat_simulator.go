package nattraversal

import (
	"fmt"
	"net"
	"sync"
	"time"
)

// NATSimulatorType define o tipo de NAT que será simulado
// NATSimulatorType defines the type of NAT that will be simulated
// NATSimulatorType define el tipo de NAT que será simulado
type NATSimulatorType int

const (
	SimulateFullCone NATSimulatorType = iota     // Permite qualquer endereço externo acessar uma porta mapeada
	SimulateRestrictedCone                        // Restringe acesso ao IP externo que já recebeu pacotes
	SimulatePortRestrictedCone                    // Restringe acesso ao par IP:porta externo específico
	SimulateSymmetric                             // Cria mapeamento diferente para cada destino
)

// NATMapping representa um mapeamento de porta NAT
type NATMapping struct {
	InternalAddr *net.UDPAddr        // Endereço interno (cliente)
	ExternalPort int                 // Porta externa atribuída
	Destinations map[string]struct{} // Conjunto de destinos permitidos (IP:porta)
	LastActivity time.Time           // Última atividade neste mapeamento
}

// NATSimulator simula diferentes tipos de NAT para testes
// NATSimulator simulates different types of NAT for testing
// NATSimulator simula diferentes tipos de NAT para pruebas
type NATSimulator struct {
	natType       NATSimulatorType       // Tipo de NAT simulado
	externalIP    net.IP                 // IP externo simulado
	internalNet   *net.IPNet             // Rede interna simulada
	mappings      map[string]*NATMapping // Mapeamentos internos -> externos
	mappingsMutex sync.RWMutex           // Mutex para acesso seguro aos mapeamentos
	
	// Sockets UDP
	externalConn  *net.UDPConn           // Socket para tráfego externo
	internalConn  *net.UDPConn           // Socket para tráfego interno
	
	nextPort      int                    // Próxima porta externa a ser atribuída
	portMutex     sync.Mutex             // Mutex para alocação de porta
	
	running       bool                   // Estado do simulador
	stopChan      chan struct{}          // Canal para sinalizar parada
}

// NewNATSimulator cria um novo simulador de NAT
// NewNATSimulator creates a new NAT simulator
// NewNATSimulator crea un nuevo simulador de NAT
func NewNATSimulator(natType NATSimulatorType, externalIP string, internalNet string) (*NATSimulator, error) {
	// Analisar o IP externo
	ip := net.ParseIP(externalIP)
	if ip == nil {
		return nil, fmt.Errorf("IP externo inválido: %s", externalIP)
	}
	
	// Analisar a rede interna
	_, ipNet, err := net.ParseCIDR(internalNet)
	if err != nil {
		return nil, fmt.Errorf("rede interna inválida: %w", err)
	}
	
	simulator := &NATSimulator{
		natType:     natType,
		externalIP:  ip,
		internalNet: ipNet,
		mappings:    make(map[string]*NATMapping),
		nextPort:    10000, // Iniciar portas externas a partir de 10000
		stopChan:    make(chan struct{}),
	}
	
	return simulator, nil
}

// Start inicia o simulador de NAT
// Start starts the NAT simulator
// Start inicia el simulador de NAT
func (s *NATSimulator) Start(externalPort, internalPort int) error {
	// Criar socket para o lado externo
	externalAddr := &net.UDPAddr{
		IP:   s.externalIP,
		Port: externalPort,
	}
	externalConn, err := net.ListenUDP("udp", externalAddr)
	if err != nil {
		return fmt.Errorf("erro ao criar socket externo: %w", err)
	}
	
	// Criar socket para o lado interno
	internalAddr := &net.UDPAddr{
		IP:   net.ParseIP("0.0.0.0"),
		Port: internalPort,
	}
	internalConn, err := net.ListenUDP("udp", internalAddr)
	if err != nil {
		externalConn.Close()
		return fmt.Errorf("erro ao criar socket interno: %w", err)
	}
	
	s.externalConn = externalConn
	s.internalConn = internalConn
	s.running = true
	
	// Iniciar rotinas de tratamento
	go s.handleInternalTraffic()
	go s.handleExternalTraffic()
	go s.cleanupMappings()
	
	fmt.Printf("Simulador de NAT iniciado. Tipo: %s\n", s.getNATTypeString())
	fmt.Printf("Endereço externo: %s:%d\n", s.externalIP, externalPort)
	fmt.Printf("Rede interna: %s\n", s.internalNet)
	
	return nil
}

// Stop para o simulador de NAT
// Stop stops the NAT simulator
// Stop detiene el simulador de NAT
func (s *NATSimulator) Stop() {
	if !s.running {
		return
	}
	
	// Sinalizar para as goroutines pararem
	close(s.stopChan)
	
	// Fechar os sockets
	if s.externalConn != nil {
		s.externalConn.Close()
		s.externalConn = nil
	}
	
	if s.internalConn != nil {
		s.internalConn.Close()
		s.internalConn = nil
	}
	
	s.running = false
	fmt.Println("Simulador de NAT encerrado.")
}

// handleInternalTraffic processa o tráfego vindo da rede interna
func (s *NATSimulator) handleInternalTraffic() {
	buffer := make([]byte, 4096)
	
	for {
		select {
		case <-s.stopChan:
			return
		default:
			// Configurar timeout para não bloquear indefinidamente
			s.internalConn.SetReadDeadline(time.Now().Add(1 * time.Second))
			
			n, addr, err := s.internalConn.ReadFromUDP(buffer)
			if err != nil {
				// Ignorar erros de timeout
				if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					continue
				}
				
				fmt.Printf("Erro ao ler tráfego interno: %v\n", err)
				continue
			}
			
			// Processar pacote interno e enviá-lo para o destino externo
			s.processInternalPacket(buffer[:n], addr)
		}
	}
}

// handleExternalTraffic processa o tráfego vindo da rede externa
func (s *NATSimulator) handleExternalTraffic() {
	buffer := make([]byte, 4096)
	
	for {
		select {
		case <-s.stopChan:
			return
		default:
			// Configurar timeout para não bloquear indefinidamente
			s.externalConn.SetReadDeadline(time.Now().Add(1 * time.Second))
			
			n, addr, err := s.externalConn.ReadFromUDP(buffer)
			if err != nil {
				// Ignorar erros de timeout
				if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					continue
				}
				
				fmt.Printf("Erro ao ler tráfego externo: %v\n", err)
				continue
			}
			
			// Processar pacote externo e decidir se deve ser encaminhado
			s.processExternalPacket(buffer[:n], addr)
		}
	}
}

// processInternalPacket processa um pacote da rede interna
func (s *NATSimulator) processInternalPacket(data []byte, srcAddr *net.UDPAddr) {
	// Extrair endereço de destino do cabeçalho do pacote (simplificado para simulação)
	// Em um cenário real, o endereço estaria nos dados do pacote
	dstAddrStr := string(data[:20])
	dstAddr, err := net.ResolveUDPAddr("udp", dstAddrStr)
	if err != nil {
		// Se não for um endereço válido, considerar como pacote normal
		// Na prática, usaríamos um protocolo adequado
		dstAddr = &net.UDPAddr{IP: net.IPv4(8, 8, 8, 8), Port: 53} // Exemplo: DNS
	}
	
	// Obter ou criar mapeamento para esta conexão
	mapping := s.getOrCreateMapping(srcAddr, dstAddr)
	
	// Criar endereço de origem "traduzido"
	srcNATAddr := &net.UDPAddr{
		IP:   s.externalIP,
		Port: mapping.ExternalPort,
	}
	
	// Registrar o destino no mapeamento para NATs restritos e simétricos
	destKey := dstAddr.String()
	
	s.mappingsMutex.Lock()
	mapping.Destinations[destKey] = struct{}{}
	mapping.LastActivity = time.Now()
	s.mappingsMutex.Unlock()
	
	// Encaminhar o pacote para o destino
	_, err = s.externalConn.WriteToUDP(data, dstAddr)
	if err != nil {
		fmt.Printf("Erro ao encaminhar pacote interno: %v\n", err)
		return
	}
	
	fmt.Printf("Pacote encaminhado: %s -> %s via %s\n",
		srcAddr.String(), dstAddr.String(), srcNATAddr.String())
}

// processExternalPacket processa um pacote da rede externa
func (s *NATSimulator) processExternalPacket(data []byte, srcAddr *net.UDPAddr) {
	// Encontrar o mapeamento com base na porta de destino
	dstPort := s.externalConn.LocalAddr().(*net.UDPAddr).Port
	
	s.mappingsMutex.RLock()
	
	// Verificar todos os mapeamentos para encontrar o destino correto
	var internalAddr *net.UDPAddr
	var mapping *NATMapping
	
	for _, m := range s.mappings {
		if m.ExternalPort == dstPort {
			// Verificar restrições com base no tipo de NAT
			switch s.natType {
			case SimulateFullCone:
				// Full Cone: aceita qualquer pacote do exterior para a porta mapeada
				internalAddr = m.InternalAddr
				mapping = m
				break
				
			case SimulateRestrictedCone:
				// Restricted Cone: verifica se o IP de origem já foi contatado
				found := false
				
				for destKey := range m.Destinations {
					destAddr, _ := net.ResolveUDPAddr("udp", destKey)
					if destAddr.IP.Equal(srcAddr.IP) {
						found = true
						break
					}
				}
				
				if found {
					internalAddr = m.InternalAddr
					mapping = m
				}
				break
				
			case SimulatePortRestrictedCone:
				// Port Restricted Cone: verifica IP:porta específica
				if _, exists := m.Destinations[srcAddr.String()]; exists {
					internalAddr = m.InternalAddr
					mapping = m
				}
				break
				
			case SimulateSymmetric:
				// Symmetric: verifica mapeamento específico para IP:porta
				if _, exists := m.Destinations[srcAddr.String()]; exists {
					internalAddr = m.InternalAddr
					mapping = m
				}
				break
			}
			
			// Se encontrou um mapeamento válido, não precisa continuar procurando
			if internalAddr != nil {
				break
			}
		}
	}
	
	// Se não encontrou mapeamento válido, descartar o pacote
	if internalAddr == nil {
		s.mappingsMutex.RUnlock()
		fmt.Printf("Pacote descartado de %s: nenhum mapeamento válido encontrado\n", srcAddr.String())
		return
	}
	
	// Atualizar timestamp de atividade
	s.mappingsMutex.RUnlock()
	
	s.mappingsMutex.Lock()
	mapping.LastActivity = time.Now()
	s.mappingsMutex.Unlock()
	
	// Encaminhar o pacote para o cliente interno
	_, err := s.internalConn.WriteToUDP(data, internalAddr)
	if err != nil {
		fmt.Printf("Erro ao encaminhar pacote externo: %v\n", err)
		return
	}
	
	fmt.Printf("Pacote encaminhado: %s -> %s\n", srcAddr.String(), internalAddr.String())
}

// getOrCreateMapping obtém um mapeamento existente ou cria um novo
func (s *NATSimulator) getOrCreateMapping(internalAddr *net.UDPAddr, dstAddr *net.UDPAddr) *NATMapping {
	s.mappingsMutex.Lock()
	defer s.mappingsMutex.Unlock()
	
	var key string
	
	// Para NAT simétrico, usar combinação de origem+destino como chave
	if s.natType == SimulateSymmetric {
		key = fmt.Sprintf("%s->%s", internalAddr.String(), dstAddr.String())
	} else {
		// Para outros tipos, usar apenas o endereço interno
		key = internalAddr.String()
	}
	
	// Verificar se já existe um mapeamento
	if mapping, exists := s.mappings[key]; exists {
		return mapping
	}
	
	// Criar novo mapeamento
	s.portMutex.Lock()
	externalPort := s.nextPort
	s.nextPort++
	s.portMutex.Unlock()
	
	mapping := &NATMapping{
		InternalAddr: internalAddr,
		ExternalPort: externalPort,
		Destinations: make(map[string]struct{}),
		LastActivity: time.Now(),
	}
	
	s.mappings[key] = mapping
	
	fmt.Printf("Novo mapeamento criado: %s -> %s:%d\n",
		internalAddr.String(), s.externalIP, externalPort)
	
	return mapping
}

// cleanupMappings remove mapeamentos inativos após um tempo
func (s *NATSimulator) cleanupMappings() {
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-s.stopChan:
			return
		case <-ticker.C:
			s.mappingsMutex.Lock()
			
			now := time.Now()
			timeout := 5 * time.Minute
			
			for key, mapping := range s.mappings {
				if now.Sub(mapping.LastActivity) > timeout {
					fmt.Printf("Removendo mapeamento inativo: %s -> %s:%d\n",
						mapping.InternalAddr.String(), s.externalIP, mapping.ExternalPort)
					delete(s.mappings, key)
				}
			}
			
			s.mappingsMutex.Unlock()
		}
	}
}

// GetNATType retorna o tipo de NAT simulado
// GetNATType returns the simulated NAT type
// GetNATType devuelve el tipo de NAT simulado
func (s *NATSimulator) GetNATType() NATSimulatorType {
	return s.natType
}

// getNATTypeString retorna o tipo de NAT como string
func (s *NATSimulator) getNATTypeString() string {
	switch s.natType {
	case SimulateFullCone:
		return "Full Cone"
	case SimulateRestrictedCone:
		return "Restricted Cone"
	case SimulatePortRestrictedCone:
		return "Port Restricted Cone"
	case SimulateSymmetric:
		return "Symmetric"
	default:
		return "Unknown"
	}
}

// ListMappings lista todos os mapeamentos ativos
// ListMappings lists all active mappings
// ListMappings lista todos los mapeos activos
func (s *NATSimulator) ListMappings() []map[string]interface{} {
	s.mappingsMutex.RLock()
	defer s.mappingsMutex.RUnlock()
	
	result := make([]map[string]interface{}, 0, len(s.mappings))
	
	for key, mapping := range s.mappings {
		destinations := make([]string, 0)
		for dest := range mapping.Destinations {
			destinations = append(destinations, dest)
		}
		
		mappingInfo := map[string]interface{}{
			"key":           key,
			"internal_addr": mapping.InternalAddr.String(),
			"external_addr": fmt.Sprintf("%s:%d", s.externalIP, mapping.ExternalPort),
			"destinations":  destinations,
			"last_activity": mapping.LastActivity.Format(time.RFC3339),
			"idle_time":     time.Since(mapping.LastActivity).String(),
		}
		
		result = append(result, mappingInfo)
	}
	
	return result
}

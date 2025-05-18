package nattraversal

import (
	"fmt"
	"net"
	"sync"
	"time"
)

// PunchingResults contém os resultados do teste de hole punching
// PunchingResults contains the results of hole punching test
// PunchingResults contiene los resultados de la prueba de hole punching
type PunchingResults struct {
	ConnectionsAttempted   int           // Número de tentativas de conexão
	ConnectionsEstablished int           // Número de conexões estabelecidas com sucesso
	AverageConnectionTime  time.Duration // Tempo médio para estabelecer uma conexão
	SuccessRate            float64       // Taxa de sucesso (0.0 a 1.0)
	FailureReasons         map[string]int // Motivos de falha e contagem
}

// HolePunchingTest implementa um teste de UDP hole punching
// HolePunchingTest implements a UDP hole punching test
// HolePunchingTest implementa una prueba de UDP hole punching
type HolePunchingTest struct {
	rendezvousServer string
	id               string
	localPort        int
	
	conn             *net.UDPConn
	serverAddr       *net.UDPAddr
	peerAddrs        map[string]*net.UDPAddr
	
	results          PunchingResults
	connectionTimes  map[string]time.Time
	established      map[string]bool
	
	running          bool
	mutex            sync.Mutex
	peersMutex       sync.RWMutex
	resultsMutex     sync.RWMutex
	
	stopChan         chan struct{}
}

// NewHolePunchingTest cria um novo teste de hole punching
// NewHolePunchingTest creates a new hole punching test
// NewHolePunchingTest crea una nueva prueba de hole punching
func NewHolePunchingTest(rendezvousServer, id string, localPort int) (*HolePunchingTest, error) {
	test := &HolePunchingTest{
		rendezvousServer: rendezvousServer,
		id:               id,
		localPort:        localPort,
		peerAddrs:        make(map[string]*net.UDPAddr),
		connectionTimes:  make(map[string]time.Time),
		established:      make(map[string]bool),
		results: PunchingResults{
			FailureReasons: make(map[string]int),
		},
		stopChan:         make(chan struct{}),
	}
	
	return test, nil
}

// Start inicia o teste de hole punching
// Start starts the hole punching test
// Start inicia la prueba de hole punching
func (t *HolePunchingTest) Start() error {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	
	if t.running {
		return fmt.Errorf("teste já está em execução")
	}
	
	// Criar socket UDP local
	addr := &net.UDPAddr{
		IP:   net.IPv4zero,
		Port: t.localPort,
	}
	
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return fmt.Errorf("erro ao criar socket UDP: %w", err)
	}
	
	t.conn = conn
	
	// Resolver o endereço do servidor de rendezvous
	serverAddr, err := net.ResolveUDPAddr("udp", t.rendezvousServer)
	if err != nil {
		t.conn.Close()
		return fmt.Errorf("erro ao resolver endereço do servidor: %w", err)
	}
	
	t.serverAddr = serverAddr
	t.running = true
	
	// Iniciar goroutines para o teste
	go t.receiveLoop()
	go t.announceLoop()
	
	// Registrar com o servidor de rendezvous
	t.register()
	
	return nil
}

// Stop para o teste de hole punching
// Stop stops the hole punching test
// Stop detiene la prueba de hole punching
func (t *HolePunchingTest) Stop() {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	
	if !t.running {
		return
	}
	
	// Sinalizar para as goroutines pararem
	close(t.stopChan)
	
	// Fechar o socket
	if t.conn != nil {
		t.conn.Close()
		t.conn = nil
	}
	
	t.running = false
	
	// Calcular resultados finais
	t.calculateResults()
}

// GetResults retorna os resultados do teste
// GetResults returns the test results
// GetResults devuelve los resultados de la prueba
func (t *HolePunchingTest) GetResults() PunchingResults {
	t.resultsMutex.RLock()
	defer t.resultsMutex.RUnlock()
	
	// Criar uma cópia dos resultados
	results := PunchingResults{
		ConnectionsAttempted:   t.results.ConnectionsAttempted,
		ConnectionsEstablished: t.results.ConnectionsEstablished,
		AverageConnectionTime:  t.results.AverageConnectionTime,
		SuccessRate:            t.results.SuccessRate,
		FailureReasons:         make(map[string]int),
	}
	
	// Copiar os motivos de falha
	for reason, count := range t.results.FailureReasons {
		results.FailureReasons[reason] = count
	}
	
	return results
}

// receiveLoop processa mensagens recebidas
func (t *HolePunchingTest) receiveLoop() {
	buffer := make([]byte, 2048)
	
	for {
		select {
		case <-t.stopChan:
			return
		default:
			// Configurar timeout para não bloquear indefinidamente
			t.conn.SetReadDeadline(time.Now().Add(1 * time.Second))
			
			n, addr, err := t.conn.ReadFromUDP(buffer)
			if err != nil {
				// Ignorar erros de timeout
				if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					continue
				}
				
				fmt.Printf("Erro ao ler mensagem UDP: %v\n", err)
				continue
			}
			
			// Processar a mensagem recebida
			t.handleMessage(buffer[:n], addr)
		}
	}
}

// announceLoop envia anúncios periódicos para o servidor
func (t *HolePunchingTest) announceLoop() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			t.announce()
		case <-t.stopChan:
			return
		}
	}
}

// register registra este nó com o servidor de rendezvous
func (t *HolePunchingTest) register() {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	
	if !t.running || t.conn == nil || t.serverAddr == nil {
		return
	}
	
	// Enviar mensagem de registro com ID
	message := fmt.Sprintf("REGISTER %s", t.id)
	_, err := t.conn.WriteToUDP([]byte(message), t.serverAddr)
	if err != nil {
		fmt.Printf("Erro ao enviar registro ao servidor: %v\n", err)
	} else {
		fmt.Printf("Registro enviado ao servidor de rendezvous: %s\n", t.serverAddr)
	}
}

// announce envia um anúncio para o servidor de rendezvous
func (t *HolePunchingTest) announce() {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	
	if !t.running || t.conn == nil || t.serverAddr == nil {
		return
	}
	
	// Enviar mensagem de anúncio com ID
	message := fmt.Sprintf("ANNOUNCE %s", t.id)
	_, err := t.conn.WriteToUDP([]byte(message), t.serverAddr)
	if err != nil {
		fmt.Printf("Erro ao enviar anúncio ao servidor: %v\n", err)
	}
}

// handleMessage processa uma mensagem recebida
func (t *HolePunchingTest) handleMessage(data []byte, addr *net.UDPAddr) {
	message := string(data)
	
	// Verificar o tipo de mensagem
	switch {
	case len(message) >= 5 && message[:5] == "PEER ":
		// Formato: PEER <id> <ip:port>
		var peerId, peerAddr string
		fmt.Sscanf(message[5:], "%s %s", &peerId, &peerAddr)
		
		if peerId != "" && peerAddr != "" {
			t.handlePeerInfo(peerId, peerAddr)
		}
		
	case len(message) >= 6 && message[:6] == "PUNCH ":
		// Formato: PUNCH <id>
		peerId := message[6:]
		
		// Tentativa de hole punching de um peer
		t.handlePunchRequest(peerId, addr)
		
	case message == "PONG":
		// Resposta a um PING, confirma conexão estabelecida
		t.handlePongResponse(addr)
		
	case len(message) >= 5 && message[:5] == "PING ":
		// Solicitação de PING, responder com PONG
		peerId := message[5:]
		t.handlePingRequest(peerId, addr)
	}
}

// handlePeerInfo processa informações sobre um peer
func (t *HolePunchingTest) handlePeerInfo(peerId, peerAddrStr string) {
	// Resolver o endereço do peer
	peerAddr, err := net.ResolveUDPAddr("udp", peerAddrStr)
	if err != nil {
		fmt.Printf("Erro ao resolver endereço do peer %s: %v\n", peerId, err)
		return
	}
	
	// Adicionar à lista de peers
	t.peersMutex.Lock()
	t.peerAddrs[peerId] = peerAddr
	t.peersMutex.Unlock()
	
	fmt.Printf("Novo peer descoberto: %s (%s)\n", peerId, peerAddrStr)
	
	// Iniciar tentativa de hole punching
	t.initiatePunching(peerId, peerAddr)
}

// initiatePunching inicia o processo de hole punching com um peer
func (t *HolePunchingTest) initiatePunching(peerId string, peerAddr *net.UDPAddr) {
	t.resultsMutex.Lock()
	t.results.ConnectionsAttempted++
	t.resultsMutex.Unlock()
	
	// Registrar o tempo de início da tentativa
	t.connectionTimes[peerId] = time.Now()
	
	// Enviar várias mensagens de PUNCH para criar o "buraco" no NAT
	message := fmt.Sprintf("PUNCH %s", t.id)
	
	fmt.Printf("Iniciando hole punching com %s (%s)...\n", peerId, peerAddr)
	
	// Enviar múltiplas mensagens com pequenos intervalos
	for i := 0; i < 5; i++ {
		t.mutex.Lock()
		if !t.running || t.conn == nil {
			t.mutex.Unlock()
			return
		}
		
		_, err := t.conn.WriteToUDP([]byte(message), peerAddr)
		t.mutex.Unlock()
		
		if err != nil {
			fmt.Printf("Erro ao enviar mensagem de hole punching: %v\n", err)
			
			t.resultsMutex.Lock()
			t.results.FailureReasons["envio_falhou"]++
			t.resultsMutex.Unlock()
			
			return
		}
		
		time.Sleep(200 * time.Millisecond)
	}
	
	// Após as mensagens iniciais, enviar um PING para verificar se a conexão foi estabelecida
	pingMessage := fmt.Sprintf("PING %s", t.id)
	
	t.mutex.Lock()
	if !t.running || t.conn == nil {
		t.mutex.Unlock()
		return
	}
	
	_, err := t.conn.WriteToUDP([]byte(pingMessage), peerAddr)
	t.mutex.Unlock()
	
	if err != nil {
		fmt.Printf("Erro ao enviar PING após hole punching: %v\n", err)
	}
	
	// Agendar uma verificação de timeout
	go t.checkConnectionTimeout(peerId, peerAddr)
}

// handlePunchRequest processa uma solicitação de hole punching
func (t *HolePunchingTest) handlePunchRequest(peerId string, addr *net.UDPAddr) {
	fmt.Printf("Recebida solicitação de hole punching de %s (%s)\n", peerId, addr)
	
	// Adicionar à lista de peers se for novo
	t.peersMutex.Lock()
	if _, exists := t.peerAddrs[peerId]; !exists {
		t.peerAddrs[peerId] = addr
		
		t.resultsMutex.Lock()
		t.results.ConnectionsAttempted++
		t.resultsMutex.Unlock()
		
		// Registrar o tempo de início da tentativa
		t.connectionTimes[peerId] = time.Now()
	}
	t.peersMutex.Unlock()
	
	// Responder com outra mensagem de PUNCH para abrir o caminho de volta
	message := fmt.Sprintf("PUNCH %s", t.id)
	
	t.mutex.Lock()
	if !t.running || t.conn == nil {
		t.mutex.Unlock()
		return
	}
	
	_, err := t.conn.WriteToUDP([]byte(message), addr)
	t.mutex.Unlock()
	
	if err != nil {
		fmt.Printf("Erro ao responder à solicitação de hole punching: %v\n", err)
		return
	}
}

// handlePingRequest processa uma solicitação de PING
func (t *HolePunchingTest) handlePingRequest(peerId string, addr *net.UDPAddr) {
	// Responder com PONG para confirmar a conexão
	t.mutex.Lock()
	if !t.running || t.conn == nil {
		t.mutex.Unlock()
		return
	}
	
	_, err := t.conn.WriteToUDP([]byte("PONG"), addr)
	t.mutex.Unlock()
	
	if err != nil {
		fmt.Printf("Erro ao enviar PONG para %s: %v\n", peerId, err)
		return
	}
	
	// Marcar a conexão como estabelecida
	t.peersMutex.Lock()
	
	// Adicionar o peer se necessário
	if _, exists := t.peerAddrs[peerId]; !exists {
		t.peerAddrs[peerId] = addr
		t.connectionTimes[peerId] = time.Now()
	}
	
	alreadyEstablished := t.established[peerId]
	t.established[peerId] = true
	
	t.peersMutex.Unlock()
	
	// Se for a primeira vez que a conexão é estabelecida, atualizar as estatísticas
	if !alreadyEstablished {
		t.recordSuccessfulConnection(peerId)
	}
}

// handlePongResponse processa uma resposta PONG
func (t *HolePunchingTest) handlePongResponse(addr *net.UDPAddr) {
	// Identificar o peer pelo endereço
	var peerId string
	
	t.peersMutex.Lock()
	for id, peerAddr := range t.peerAddrs {
		if peerAddr.IP.Equal(addr.IP) && peerAddr.Port == addr.Port {
			peerId = id
			break
		}
	}
	
	if peerId == "" {
		// Peer desconhecido, ignorar
		t.peersMutex.Unlock()
		return
	}
	
	alreadyEstablished := t.established[peerId]
	t.established[peerId] = true
	
	t.peersMutex.Unlock()
	
	// Se for a primeira vez que a conexão é estabelecida, atualizar as estatísticas
	if !alreadyEstablished {
		t.recordSuccessfulConnection(peerId)
	}
}

// recordSuccessfulConnection registra uma conexão bem-sucedida
func (t *HolePunchingTest) recordSuccessfulConnection(peerId string) {
	// Calcular o tempo que levou para estabelecer a conexão
	startTime, exists := t.connectionTimes[peerId]
	if !exists {
		startTime = time.Now() // Fallback, não deveria acontecer
	}
	
	elapsed := time.Since(startTime)
	
	t.resultsMutex.Lock()
	defer t.resultsMutex.Unlock()
	
	t.results.ConnectionsEstablished++
	
	fmt.Printf("Conexão estabelecida com %s após %v\n", peerId, elapsed)
}

// checkConnectionTimeout verifica se uma conexão expirou
func (t *HolePunchingTest) checkConnectionTimeout(peerId string, peerAddr *net.UDPAddr) {
	// Esperar pelo timeout
	time.Sleep(30 * time.Second)
	
	t.peersMutex.RLock()
	established := t.established[peerId]
	t.peersMutex.RUnlock()
	
	// Se ainda não estiver estabelecida, registrar falha
	if !established {
		t.resultsMutex.Lock()
		t.results.FailureReasons["timeout"]++
		t.resultsMutex.Unlock()
		
		fmt.Printf("Timeout na conexão com %s\n", peerId)
	}
}

// calculateResults calcula os resultados finais do teste
func (t *HolePunchingTest) calculateResults() {
	t.resultsMutex.Lock()
	defer t.resultsMutex.Unlock()
	
	// Calcular taxa de sucesso
	if t.results.ConnectionsAttempted > 0 {
		t.results.SuccessRate = float64(t.results.ConnectionsEstablished) / float64(t.results.ConnectionsAttempted)
	}
	
	// Calcular tempo médio para conexões bem-sucedidas
	var totalTime time.Duration
	count := 0
	
	t.peersMutex.RLock()
	for peerId, established := range t.established {
		if established {
			startTime, exists := t.connectionTimes[peerId]
			if exists {
				totalTime += time.Since(startTime)
				count++
			}
		}
	}
	t.peersMutex.RUnlock()
	
	if count > 0 {
		t.results.AverageConnectionTime = totalTime / time.Duration(count)
	}
}

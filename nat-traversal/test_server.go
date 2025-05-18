package nattraversal

import (
	"fmt"
	"net"
	"sync"
	"time"
)

// TestServer implementa um servidor de teste para NAT traversal
// TestServer implements a test server for NAT traversal
// TestServer implementa un servidor de prueba para NAT traversal
type TestServer struct {
	port      int
	conn      *net.UDPConn
	running   bool
	mutex     sync.Mutex
	clients   map[string]*net.UDPAddr
	clientsMx sync.RWMutex
	stopChan  chan struct{}
}

// NewTestServer cria um novo servidor de teste
// NewTestServer creates a new test server
// NewTestServer crea un nuevo servidor de prueba
func NewTestServer(port int) (*TestServer, error) {
	server := &TestServer{
		port:     port,
		clients:  make(map[string]*net.UDPAddr),
		stopChan: make(chan struct{}),
	}
	
	return server, nil
}

// Start inicia o servidor para escutar conexões
// Start starts the server to listen for connections
// Start inicia el servidor para escuchar conexiones
func (s *TestServer) Start() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	if s.running {
		return fmt.Errorf("servidor já está em execução")
	}
	
	// Criar socket UDP para escutar
	addr := &net.UDPAddr{
		IP:   net.IPv4zero,
		Port: s.port,
	}
	
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return fmt.Errorf("erro ao criar socket UDP: %w", err)
	}
	
	s.conn = conn
	s.running = true
	
	// Iniciar goroutine para processar mensagens
	go s.processMessages()
	
	return nil
}

// Stop para o servidor
// Stop stops the server
// Stop detiene el servidor
func (s *TestServer) Stop() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	if !s.running {
		return nil
	}
	
	// Sinalizar para a goroutine parar
	close(s.stopChan)
	
	// Fechar o socket
	if s.conn != nil {
		s.conn.Close()
		s.conn = nil
	}
	
	s.running = false
	return nil
}

// processMessages processa mensagens recebidas dos clientes
func (s *TestServer) processMessages() {
	buffer := make([]byte, 2048)
	
	for {
		select {
		case <-s.stopChan:
			return
		default:
			// Configurar timeout para não bloquear indefinidamente
			s.conn.SetReadDeadline(time.Now().Add(1 * time.Second))
			
			n, addr, err := s.conn.ReadFromUDP(buffer)
			if err != nil {
				// Ignorar erros de timeout
				if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					continue
				}
				
				fmt.Printf("Erro ao ler mensagem UDP: %v\n", err)
				continue
			}
			
			// Processar a mensagem recebida
			s.handleMessage(buffer[:n], addr)
		}
	}
}

// handleMessage processa uma mensagem recebida de um cliente
func (s *TestServer) handleMessage(data []byte, addr *net.UDPAddr) {
	message := string(data)
	
	// Verificar se é um comando ou uma mensagem normal
	if len(message) > 0 && message[0] == '/' {
		s.handleCommand(message[1:], addr)
	} else {
		s.broadcastMessage(fmt.Sprintf("[%s:%d] %s", addr.IP, addr.Port, message), addr)
	}
}

// handleCommand processa um comando do cliente
func (s *TestServer) handleCommand(cmd string, addr *net.UDPAddr) {
	// Comandos possíveis:
	// /register <name> - Registrar como cliente com um nome
	// /list - Listar clientes conectados
	// /connect <name> - Solicitar conexão com outro cliente
	
	switch {
	case len(cmd) >= 9 && cmd[:8] == "register ":
		name := cmd[8:]
		s.registerClient(name, addr)
		
	case cmd == "list":
		s.sendClientList(addr)
		
	case len(cmd) >= 9 && cmd[:8] == "connect ":
		targetName := cmd[8:]
		s.facilitateConnection(targetName, addr)
		
	default:
		s.sendToClient("Comando desconhecido. Comandos disponíveis: /register <name>, /list, /connect <name>", addr)
	}
}

// registerClient registra um cliente com um nome
func (s *TestServer) registerClient(name string, addr *net.UDPAddr) {
	s.clientsMx.Lock()
	defer s.clientsMx.Unlock()
	
	// Verificar se o nome já está em uso
	for existingName, existingAddr := range s.clients {
		if existingName == name {
			// Se o mesmo endereço, apenas atualizar
			if existingAddr.IP.Equal(addr.IP) && existingAddr.Port == addr.Port {
				s.clients[name] = addr
				s.sendToClient(fmt.Sprintf("Registro atualizado como '%s'", name), addr)
				return
			}
			
			// Outro cliente com mesmo nome
			s.sendToClient(fmt.Sprintf("Nome '%s' já está em uso", name), addr)
			return
		}
	}
	
	// Registrar o novo cliente
	s.clients[name] = addr
	s.sendToClient(fmt.Sprintf("Registrado com sucesso como '%s'", name), addr)
	
	// Informar outros clientes
	s.broadcastMessage(fmt.Sprintf("Novo cliente conectado: %s (%s:%d)", 
		name, addr.IP, addr.Port), addr)
}

// sendClientList envia a lista de clientes para um cliente
func (s *TestServer) sendClientList(addr *net.UDPAddr) {
	s.clientsMx.RLock()
	defer s.clientsMx.RUnlock()
	
	if len(s.clients) == 0 {
		s.sendToClient("Nenhum cliente registrado", addr)
		return
	}
	
	message := "Clientes conectados:\n"
	for name, clientAddr := range s.clients {
		message += fmt.Sprintf("  - %s (%s:%d)\n", name, clientAddr.IP, clientAddr.Port)
	}
	
	s.sendToClient(message, addr)
}

// facilitateConnection facilita uma conexão entre dois clientes
func (s *TestServer) facilitateConnection(targetName string, requesterAddr *net.UDPAddr) {
	s.clientsMx.RLock()
	defer s.clientsMx.RUnlock()
	
	// Encontrar o cliente alvo pelo nome
	targetAddr, exists := s.clients[targetName]
	if !exists {
		s.sendToClient(fmt.Sprintf("Cliente '%s' não encontrado", targetName), requesterAddr)
		return
	}
	
	// Encontrar o nome do solicitante
	requesterName := ""
	for name, addr := range s.clients {
		if addr.IP.Equal(requesterAddr.IP) && addr.Port == requesterAddr.Port {
			requesterName = name
			break
		}
	}
	
	if requesterName == "" {
		s.sendToClient("Você precisa se registrar primeiro com /register <name>", requesterAddr)
		return
	}
	
	// Enviar informações de conexão para ambos os clientes
	s.sendToClient(fmt.Sprintf("Solicitação de conexão enviada para '%s' (%s:%d)", 
		targetName, targetAddr.IP, targetAddr.Port), requesterAddr)
	
	s.sendToClient(fmt.Sprintf("Solicitação de conexão recebida de '%s' (%s:%d)", 
		requesterName, requesterAddr.IP, requesterAddr.Port), targetAddr)
	
	// Facilitar o hole punching enviando endereços para ambos
	connRequest := fmt.Sprintf("/punch %s:%d", requesterAddr.IP, requesterAddr.Port)
	s.sendToClient(connRequest, targetAddr)
	
	connRequest = fmt.Sprintf("/punch %s:%d", targetAddr.IP, targetAddr.Port)
	s.sendToClient(connRequest, requesterAddr)
}

// sendToClient envia uma mensagem para um cliente específico
func (s *TestServer) sendToClient(message string, addr *net.UDPAddr) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	if !s.running || s.conn == nil {
		return
	}
	
	_, err := s.conn.WriteToUDP([]byte(message), addr)
	if err != nil {
		fmt.Printf("Erro ao enviar mensagem para %s:%d: %v\n", 
			addr.IP, addr.Port, err)
	}
}

// broadcastMessage envia uma mensagem para todos os clientes exceto o remetente
func (s *TestServer) broadcastMessage(message string, sender *net.UDPAddr) {
	s.clientsMx.RLock()
	defer s.clientsMx.RUnlock()
	
	for _, addr := range s.clients {
		// Não enviar para o remetente
		if addr.IP.Equal(sender.IP) && addr.Port == sender.Port {
			continue
		}
		
		s.sendToClient(message, addr)
	}
}

package nattraversal

import (
	"fmt"
	"net"
	"sync"
	"time"
)

// TestClient implementa um cliente de teste para NAT traversal
// TestClient implements a test client for NAT traversal
// TestClient implementa un cliente de prueba para NAT traversal
type TestClient struct {
	localPort  int
	conn       *net.UDPConn
	serverAddr *net.UDPAddr
	remoteAddr *net.UDPAddr
	connected  bool
	mutex      sync.Mutex
	stopChan   chan struct{}
}

// NewTestClient cria um novo cliente de teste
// NewTestClient creates a new test client
// NewTestClient crea un nuevo cliente de prueba
func NewTestClient(localPort int) (*TestClient, error) {
	client := &TestClient{
		localPort: localPort,
		stopChan:  make(chan struct{}),
	}
	
	// Criar socket UDP local
	var addr *net.UDPAddr
	if localPort > 0 {
		addr = &net.UDPAddr{
			IP:   net.IPv4zero,
			Port: localPort,
		}
	} else {
		addr = &net.UDPAddr{
			IP:   net.IPv4zero,
			Port: 0, // Porta aleatória
		}
	}
	
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return nil, fmt.Errorf("erro ao criar socket UDP: %w", err)
	}
	
	client.conn = conn
	return client, nil
}

// Connect conecta ao servidor de teste
// Connect connects to the test server
// Connect conecta al servidor de prueba
func (c *TestClient) Connect(serverAddr string) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	
	// Resolver o endereço do servidor
	addr, err := net.ResolveUDPAddr("udp", serverAddr)
	if err != nil {
		return fmt.Errorf("erro ao resolver endereço do servidor: %w", err)
	}
	
	c.serverAddr = addr
	c.connected = true
	
	// Enviar uma mensagem inicial para verificar a conexão
	_, err = c.conn.WriteToUDP([]byte("hello"), c.serverAddr)
	if err != nil {
		return fmt.Errorf("erro ao enviar mensagem inicial: %w", err)
	}
	
	return nil
}

// ConnectToPeer conecta diretamente a outro peer
// ConnectToPeer connects directly to another peer
// ConnectToPeer conecta directamente a otro par
func (c *TestClient) ConnectToPeer(peerAddr string) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	
	// Resolver o endereço do peer
	addr, err := net.ResolveUDPAddr("udp", peerAddr)
	if err != nil {
		return fmt.Errorf("erro ao resolver endereço do peer: %w", err)
	}
	
	c.remoteAddr = addr
	
	// Enviar uma mensagem inicial para iniciar o hole punching
	for i := 0; i < 5; i++ {
		_, err = c.conn.WriteToUDP([]byte("punch"), c.remoteAddr)
		if err != nil {
			return fmt.Errorf("erro ao enviar mensagem de hole punch: %w", err)
		}
		time.Sleep(100 * time.Millisecond)
	}
	
	fmt.Printf("Enviadas mensagens de hole punch para %s\n", peerAddr)
	return nil
}

// SendMessage envia uma mensagem para o servidor ou peer conectado
// SendMessage sends a message to the connected server or peer
// SendMessage envía un mensaje al servidor o par conectado
func (c *TestClient) SendMessage(message string) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	
	if !c.connected {
		return fmt.Errorf("cliente não está conectado")
	}
	
	// Determinar para onde enviar a mensagem
	targetAddr := c.serverAddr
	if c.remoteAddr != nil {
		targetAddr = c.remoteAddr
	}
	
	_, err := c.conn.WriteToUDP([]byte(message), targetAddr)
	if err != nil {
		return fmt.Errorf("erro ao enviar mensagem: %w", err)
	}
	
	return nil
}

// ReceiveLoop inicia um loop para receber mensagens
// ReceiveLoop starts a loop to receive messages
// ReceiveLoop inicia un bucle para recibir mensajes
func (c *TestClient) ReceiveLoop() {
	buffer := make([]byte, 2048)
	
	for {
		select {
		case <-c.stopChan:
			return
		default:
			// Configurar timeout para não bloquear indefinidamente
			c.conn.SetReadDeadline(time.Now().Add(1 * time.Second))
			
			n, addr, err := c.conn.ReadFromUDP(buffer)
			if err != nil {
				// Ignorar erros de timeout
				if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					continue
				}
				
				fmt.Printf("Erro ao ler mensagem: %v\n", err)
				continue
			}
			
			message := string(buffer[:n])
			
			// Verificar se é um comando
			if len(message) > 0 && message[0] == '/' {
				c.handleCommand(message[1:], addr)
			} else {
				// Exibir a mensagem recebida
				fmt.Printf("\rRecebido de %s:%d: %s\n> ", 
					addr.IP, addr.Port, message)
			}
		}
	}
}

// handleCommand processa comandos recebidos
func (c *TestClient) handleCommand(cmd string, addr *net.UDPAddr) {
	// Comandos possíveis:
	// /punch <ip:porta> - Solicita hole punching com outro peer
	
	if len(cmd) >= 7 && cmd[:6] == "punch " {
		peerAddr := cmd[6:]
		fmt.Printf("\rComando de hole punching recebido. Tentando conectar com %s\n> ", peerAddr)
		
		// Iniciar conexão com o peer
		go func() {
			if err := c.ConnectToPeer(peerAddr); err != nil {
				fmt.Printf("\rErro ao conectar com peer: %v\n> ", err)
			}
		}()
	}
}

// LocalAddr retorna o endereço local do cliente
// LocalAddr returns the client's local address
// LocalAddr devuelve la dirección local del cliente
func (c *TestClient) LocalAddr() string {
	return c.conn.LocalAddr().String()
}

// RemoteAddr retorna o endereço remoto conectado
// RemoteAddr returns the connected remote address
// RemoteAddr devuelve la dirección remota conectada
func (c *TestClient) RemoteAddr() string {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	
	if c.remoteAddr != nil {
		return c.remoteAddr.String()
	}
	if c.serverAddr != nil {
		return c.serverAddr.String()
	}
	return "não conectado"
}

// Close fecha o cliente e libera recursos
// Close closes the client and frees resources
// Close cierra el cliente y libera recursos
func (c *TestClient) Close() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	
	close(c.stopChan)
	
	if c.conn != nil {
		c.conn.Close()
		c.conn = nil
	}
	
	c.connected = false
}

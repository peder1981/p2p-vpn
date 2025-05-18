package core

import (
	"fmt"
	"log"
	"net"
	"sync"

	"github.com/p2p-vpn/p2p-vpn/security/packets"
)

// SecureTransport é um wrapper para transporte de dados que implementa assinatura digital
// SecureTransport is a wrapper for data transport that implements digital signatures
// SecureTransport es un envoltorio para transporte de datos que implementa firmas digitales
type SecureTransport struct {
	vpnCore          *VPNCore
	secureHandler    *SecurePacketHandler
	peerConnections  map[string]*peerConnection
	connectionsMutex sync.RWMutex
}

// peerConnection representa uma conexão segura com um peer
// peerConnection represents a secure connection with a peer
// peerConnection representa una conexión segura con un peer
type peerConnection struct {
	peerID     string
	conn       net.Conn
	publicKey  []byte
	authorized bool
}

// NewSecureTransport cria um novo wrapper de transporte seguro
// NewSecureTransport creates a new secure transport wrapper
// NewSecureTransport crea un nuevo envoltorio de transporte seguro
func NewSecureTransport(vpnCore *VPNCore, configDir string) (*SecureTransport, error) {
	// Criar gerenciador de pacotes seguros
	secureHandler, err := NewSecurePacketHandler(configDir, vpnCore.config.NodeID)
	if err != nil {
		return nil, fmt.Errorf("erro ao criar gerenciador de pacotes seguros: %w", err)
	}

	return &SecureTransport{
		vpnCore:         vpnCore,
		secureHandler:   secureHandler,
		peerConnections: make(map[string]*peerConnection),
	}, nil
}

// SendSecurePacket envia um pacote com assinatura digital para um peer
// SendSecurePacket sends a digitally signed packet to a peer
// SendSecurePacket envía un paquete con firma digital a un peer
func (st *SecureTransport) SendSecurePacket(peerID string, data []byte, packetType packets.PacketType) error {
	// Verificar se a conexão existe
	st.connectionsMutex.RLock()
	peerConn, exists := st.peerConnections[peerID]
	st.connectionsMutex.RUnlock()

	if !exists {
		return fmt.Errorf("conexão com peer %s não encontrada", peerID)
	}

	// Verificar se o peer está autorizado
	if !peerConn.authorized {
		return fmt.Errorf("peer %s não está autorizado", peerID)
	}

	// Assinar o pacote
	signedData, err := st.secureHandler.SignPacket(peerID, packetType, data)
	if err != nil {
		return fmt.Errorf("erro ao assinar pacote: %w", err)
	}

	// Enviar o pacote assinado
	_, err = peerConn.conn.Write(signedData)
	if err != nil {
		return fmt.Errorf("erro ao enviar pacote: %w", err)
	}

	return nil
}

// HandleIncomingPacket processa um pacote recebido, verificando sua assinatura
// HandleIncomingPacket processes a received packet, verifying its signature
// HandleIncomingPacket procesa un paquete recibido, verificando su firma
func (st *SecureTransport) HandleIncomingPacket(data []byte) error {
	// Verificar e extrair dados do pacote
	senderID, payload, packetType, err := st.secureHandler.VerifyAndExtractPacket(data)
	if err != nil {
		return fmt.Errorf("erro ao verificar pacote: %w", err)
	}

	// Processar o pacote com base no tipo
	switch packetType {
	case packets.PacketTypeHandshake:
		return st.handleHandshakePacket(senderID, payload)
	case packets.PacketTypeData:
		return st.handleDataPacket(senderID, payload)
	case packets.PacketTypeControl:
		return st.handleControlPacket(senderID, payload)
	case packets.PacketTypeKeepAlive:
		return st.handleKeepAlivePacket(senderID)
	default:
		return fmt.Errorf("tipo de pacote desconhecido: %d", packetType)
	}
}

// handleHandshakePacket processa um pacote de handshake para estabelecer uma conexão segura
// handleHandshakePacket processes a handshake packet to establish a secure connection
// handleHandshakePacket procesa un paquete de handshake para establecer una conexión segura
func (st *SecureTransport) handleHandshakePacket(peerID string, data []byte) error {
	// Verifica se é um peer válido/esperado
	trustedPeer := st.vpnCore.GetPeerByID(peerID)
	if trustedPeer == nil {
		return fmt.Errorf("handshake recebido de peer desconhecido: %s", peerID)
	}

	// Adiciona a chave pública do peer
	err := st.secureHandler.AddPeerPublicKey(peerID, data)
	if err != nil {
		return fmt.Errorf("erro ao adicionar chave pública do peer: %w", err)
	}

	// Marca o peer como autorizado
	st.connectionsMutex.Lock()
	if conn, exists := st.peerConnections[peerID]; exists {
		conn.authorized = true
	}
	st.connectionsMutex.Unlock()

	log.Printf("Handshake seguro concluído com peer %s", peerID)
	return nil
}

// handleDataPacket processa um pacote de dados regular
// handleDataPacket processes a regular data packet
// handleDataPacket procesa un paquete de datos regular
func (st *SecureTransport) handleDataPacket(peerID string, data []byte) error {
	// Encaminhar para o core da VPN para processamento
	return st.vpnCore.ProcessIncomingData(peerID, data)
}

// handleControlPacket processa um pacote de controle
// handleControlPacket processes a control packet
// handleControlPacket procesa un paquete de control
func (st *SecureTransport) handleControlPacket(peerID string, data []byte) error {
	// Implementação de processamento de pacotes de controle
	// Poderia incluir comandos como alteração de configuração, reconexão, etc.
	log.Printf("Pacote de controle recebido de %s", peerID)
	return nil
}

// handleKeepAlivePacket processa um pacote de manutenção de conexão
// handleKeepAlivePacket processes a connection maintenance packet
// handleKeepAlivePacket procesa un paquete de mantenimiento de conexión
func (st *SecureTransport) handleKeepAlivePacket(peerID string) error {
	// Atualizar timestamp da última atividade do peer
	st.connectionsMutex.Lock()
	if conn, exists := st.peerConnections[peerID]; exists {
		// Atualizaria timestamp aqui se necessário
		_ = conn
	}
	st.connectionsMutex.Unlock()
	
	return nil
}

// RegisterPeerConnection registra uma conexão com um peer
// RegisterPeerConnection registers a connection with a peer
// RegisterPeerConnection registra una conexión con un peer
func (st *SecureTransport) RegisterPeerConnection(peerID string, conn net.Conn) {
	st.connectionsMutex.Lock()
	defer st.connectionsMutex.Unlock()
	
	st.peerConnections[peerID] = &peerConnection{
		peerID:     peerID,
		conn:       conn,
		authorized: false,
	}
}

// InitiateHandshake inicia o processo de handshake com um peer
// InitiateHandshake initiates the handshake process with a peer
// InitiateHandshake inicia el proceso de handshake con un peer
func (st *SecureTransport) InitiateHandshake(peerID string) error {
	// Verificar se a conexão existe
	st.connectionsMutex.RLock()
	peerConn, exists := st.peerConnections[peerID]
	st.connectionsMutex.RUnlock()

	if !exists {
		return fmt.Errorf("conexão com peer %s não encontrada", peerID)
	}

	// Obter chave pública em formato PEM
	publicKeyPEM, err := st.secureHandler.GetPublicKeyPEM()
	if err != nil {
		return fmt.Errorf("erro ao obter chave pública: %w", err)
	}

	// Enviar pacote de handshake com a chave pública
	err = st.SendSecurePacket(peerID, publicKeyPEM, packets.PacketTypeHandshake)
	if err != nil {
		return fmt.Errorf("erro ao enviar handshake: %w", err)
	}

	log.Printf("Handshake iniciado com peer %s", peerID)
	
	// A autorização completa ocorre quando receber o handshake de resposta
	peerConn.publicKey = publicKeyPEM
	
	return nil
}

// ClosePeerConnection fecha e remove uma conexão com um peer
// ClosePeerConnection closes and removes a connection with a peer
// ClosePeerConnection cierra y elimina una conexión con un peer
func (st *SecureTransport) ClosePeerConnection(peerID string) {
	st.connectionsMutex.Lock()
	defer st.connectionsMutex.Unlock()
	
	if conn, exists := st.peerConnections[peerID]; exists {
		// Fechar conexão
		if conn.conn != nil {
			conn.conn.Close()
		}
		
		// Remover da lista
		delete(st.peerConnections, peerID)
	}
}

// SendKeepAlive envia um pacote de manutenção de conexão para um peer
// SendKeepAlive sends a connection maintenance packet to a peer
// SendKeepAlive envía un paquete de mantenimiento de conexión a un peer
func (st *SecureTransport) SendKeepAlive(peerID string) error {
	// Enviar pacote vazio de tipo KeepAlive
	return st.SendSecurePacket(peerID, []byte{}, packets.PacketTypeKeepAlive)
}

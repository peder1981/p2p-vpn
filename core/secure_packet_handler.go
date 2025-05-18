package core

import (
	"fmt"
	"log"

	"github.com/p2p-vpn/p2p-vpn/security/packets"
)

// SecurePacketHandler gerencia o processamento seguro de pacotes com assinatura digital
// SecurePacketHandler manages secure packet processing with digital signatures
// SecurePacketHandler gestiona el procesamiento seguro de paquetes con firma digital
type SecurePacketHandler struct {
	keyManager *packets.KeyManager
	nodeID     string
}

// NewSecurePacketHandler cria um novo gerenciador de pacotes seguros
// NewSecurePacketHandler creates a new secure packet handler
// NewSecurePacketHandler crea un nuevo gestor de paquetes seguros
func NewSecurePacketHandler(configDir string, nodeID string) (*SecurePacketHandler, error) {
	// Criar gerenciador de chaves
	keyManager, err := packets.NewKeyManager(fmt.Sprintf("%s/keys", configDir))
	if err != nil {
		return nil, fmt.Errorf("erro ao inicializar gerenciador de chaves: %w", err)
	}

	// Carregar chaves de peers armazenadas
	if err := keyManager.LoadPeerKeys(); err != nil {
		log.Printf("Aviso: erro ao carregar chaves de peers: %v", err)
	}

	return &SecurePacketHandler{
		keyManager: keyManager,
		nodeID:     nodeID,
	}, nil
}

// SignPacket assina um pacote de dados para transmissão segura
// SignPacket signs a data packet for secure transmission
// SignPacket firma un paquete de datos para transmisión segura
func (h *SecurePacketHandler) SignPacket(recipientID string, packetType packets.PacketType, data []byte) ([]byte, error) {
	// Obter chave privada para assinatura
	privateKey := h.keyManager.GetPrivateKey()

	// Criar pacote assinado
	signedPacket, err := packets.NewSignedPacket(h.nodeID, recipientID, packetType, data, privateKey)
	if err != nil {
		return nil, fmt.Errorf("erro ao criar pacote assinado: %w", err)
	}

	// Codificar pacote para transmissão
	encodedPacket, err := signedPacket.Encode()
	if err != nil {
		return nil, fmt.Errorf("erro ao codificar pacote: %w", err)
	}

	return encodedPacket, nil
}

// VerifyAndExtractPacket verifica e extrai dados de um pacote assinado
// VerifyAndExtractPacket verifies and extracts data from a signed packet
// VerifyAndExtractPacket verifica y extrae datos de un paquete firmado
func (h *SecurePacketHandler) VerifyAndExtractPacket(data []byte) (string, []byte, packets.PacketType, error) {
	// Decodificar pacote
	packet, err := packets.Decode(data)
	if err != nil {
		return "", nil, 0, fmt.Errorf("erro ao decodificar pacote: %w", err)
	}

	// Verificar se o pacote é para este nó
	if packet.Header.RecipientID != h.nodeID && packet.Header.RecipientID != "broadcast" {
		return "", nil, 0, fmt.Errorf("pacote destinado a outro nó: %s", packet.Header.RecipientID)
	}

	// Obter chave pública do remetente
	senderPublicKey, err := h.keyManager.GetPeerKey(packet.Header.SenderID)
	if err != nil {
		return "", nil, 0, fmt.Errorf("chave pública do remetente não encontrada: %w", err)
	}

	// Verificar assinatura
	if err := packet.Verify(senderPublicKey); err != nil {
		return "", nil, 0, fmt.Errorf("verificação de assinatura falhou: %w", err)
	}

	return packet.Header.SenderID, packet.Payload, packet.Header.PacketType, nil
}

// GetPublicKeyPEM obtém a chave pública do nó em formato PEM
// GetPublicKeyPEM gets the node's public key in PEM format
// GetPublicKeyPEM obtiene la clave pública del nodo en formato PEM
func (h *SecurePacketHandler) GetPublicKeyPEM() ([]byte, error) {
	return h.keyManager.GetPublicKeyPEM()
}

// AddPeerPublicKey adiciona a chave pública de um peer
// AddPeerPublicKey adds a peer's public key
// AddPeerPublicKey añade la clave pública de un peer
func (h *SecurePacketHandler) AddPeerPublicKey(peerID string, publicKeyPEM []byte) error {
	// Adicionar ao gerenciador de chaves
	if err := h.keyManager.AddPeerKeyFromPEM(peerID, publicKeyPEM); err != nil {
		return fmt.Errorf("erro ao adicionar chave do peer: %w", err)
	}

	// Armazenar para uso futuro
	// Obter a chave pública do peer
	peerKey, err := h.keyManager.GetPeerKey(peerID)
	if err != nil {
		return fmt.Errorf("erro ao obter chave do peer para armazenamento: %w", err)
	}
	
	// Armazenar a chave do peer
	return h.keyManager.StorePeerKey(peerID, peerKey)
}

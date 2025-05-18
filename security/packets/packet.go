package packets

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

// SignedPacket representa um pacote com assinatura digital
// SignedPacket represents a packet with digital signature
// SignedPacket representa un paquete con firma digital
type SignedPacket struct {
	// Header contém metadados sobre o pacote
	// Header contains packet metadata
	// Header contiene metadatos sobre el paquete
	Header PacketHeader `json:"header"`

	// Payload contém os dados do pacote
	// Payload contains the packet data
	// Payload contiene los datos del paquete
	Payload []byte `json:"payload"`

	// Signature é a assinatura digital do pacote
	// Signature is the digital signature of the packet
	// Signature es la firma digital del paquete
	Signature []byte `json:"signature"`
}

// PacketHeader contém informações sobre o pacote
// PacketHeader contains information about the packet
// PacketHeader contiene información sobre el paquete
type PacketHeader struct {
	// SenderID é o identificador do remetente
	// SenderID is the sender identifier
	// SenderID es el identificador del remitente
	SenderID string `json:"sender_id"`

	// RecipientID é o identificador do destinatário
	// RecipientID is the recipient identifier
	// RecipientID es el identificador del destinatario
	RecipientID string `json:"recipient_id"`

	// Timestamp é o momento de criação do pacote
	// Timestamp is the moment the packet was created
	// Timestamp es el momento en que se creó el paquete
	Timestamp int64 `json:"timestamp"`

	// PacketType indica o tipo de pacote
	// PacketType indicates the type of packet
	// PacketType indica el tipo de paquete
	PacketType PacketType `json:"packet_type"`

	// Version é a versão do formato do pacote
	// Version is the packet format version
	// Version es la versión del formato del paquete
	Version uint8 `json:"version"`

	// Nonce é um valor único para evitar ataques de repetição
	// Nonce is a unique value to prevent replay attacks
	// Nonce es un valor único para evitar ataques de repetición
	Nonce []byte `json:"nonce"`
}

// PacketType define os tipos de pacotes possíveis
// PacketType defines the possible packet types
// PacketType define los tipos de paquetes posibles
type PacketType uint8

const (
	// PacketTypeData indica um pacote regular de dados
	// PacketTypeData indicates a regular data packet
	// PacketTypeData indica un paquete de datos regular
	PacketTypeData PacketType = iota

	// PacketTypeControl indica um pacote de controle
	// PacketTypeControl indicates a control packet
	// PacketTypeControl indica un paquete de control
	PacketTypeControl

	// PacketTypeHandshake indica um pacote de negociação
	// PacketTypeHandshake indicates a handshake packet
	// PacketTypeHandshake indica un paquete de negociación
	PacketTypeHandshake

	// PacketTypeKeepAlive indica um pacote de manutenção de conexão
	// PacketTypeKeepAlive indicates a connection maintenance packet
	// PacketTypeKeepAlive indica un paquete de mantenimiento de conexión
	PacketTypeKeepAlive
)

// Versão atual do formato de pacote
// Current packet format version
// Versión actual del formato de paquete
const (
	CurrentPacketVersion = uint8(1)
)

// ErrInvalidSignature é retornado quando a assinatura não é válida
// ErrInvalidSignature is returned when the signature is not valid
// ErrInvalidSignature se devuelve cuando la firma no es válida
var ErrInvalidSignature = errors.New("assinatura digital inválida")

// ErrExpiredPacket é retornado quando o pacote está expirado
// ErrExpiredPacket is returned when the packet is expired
// ErrExpiredPacket se devuelve cuando el paquete ha expirado
var ErrExpiredPacket = errors.New("pacote expirado")

// MaxPacketAge define o tempo máximo de validade de um pacote em segundos
// MaxPacketAge defines the maximum validity time of a packet in seconds
// MaxPacketAge define el tiempo máximo de validez de un paquete en segundos
const MaxPacketAge = 300 // 5 minutos / 5 minutes / 5 minutos

// NewSignedPacket cria um novo pacote assinado
// NewSignedPacket creates a new signed packet
// NewSignedPacket crea un nuevo paquete firmado
func NewSignedPacket(senderID, recipientID string, packetType PacketType, payload []byte, privateKey ed25519.PrivateKey) (*SignedPacket, error) {
	// Criar nonce aleatório
	nonce := make([]byte, 16)
	if _, err := rand.Read(nonce); err != nil {
		return nil, fmt.Errorf("erro ao gerar nonce: %w", err)
	}

	// Criar cabeçalho do pacote
	header := PacketHeader{
		SenderID:    senderID,
		RecipientID: recipientID,
		Timestamp:   time.Now().Unix(),
		PacketType:  packetType,
		Version:     CurrentPacketVersion,
		Nonce:       nonce,
	}

	// Criar o pacote sem assinatura
	packet := &SignedPacket{
		Header:  header,
		Payload: payload,
	}

	// Calcular o hash do pacote para assinar
	dataToSign, err := packet.dataToSign()
	if err != nil {
		return nil, fmt.Errorf("erro ao preparar dados para assinatura: %w", err)
	}

	// Assinar o pacote
	signature := ed25519.Sign(privateKey, dataToSign)
	packet.Signature = signature

	return packet, nil
}

// dataToSign prepara os dados para assinatura
// dataToSign prepares the data for signing
// dataToSign prepara los datos para la firma
func (p *SignedPacket) dataToSign() ([]byte, error) {
	// Concatenar os campos relevantes para assinar
	// Marshalling para JSON garante consistência
	headerJSON, err := json.Marshal(p.Header)
	if err != nil {
		return nil, err
	}

	// Calcular o hash da combinação do cabeçalho e payload
	h := sha256.New()
	h.Write(headerJSON)
	h.Write(p.Payload)
	
	return h.Sum(nil), nil
}

// Verify verifica a assinatura digital do pacote
// Verify checks the digital signature of the packet
// Verify verifica la firma digital del paquete
func (p *SignedPacket) Verify(publicKey ed25519.PublicKey) error {
	// Verificar se o pacote não está expirado
	currentTime := time.Now().Unix()
	if currentTime - p.Header.Timestamp > MaxPacketAge {
		return ErrExpiredPacket
	}

	// Obter dados para verificação da assinatura
	dataToVerify, err := p.dataToSign()
	if err != nil {
		return fmt.Errorf("erro ao preparar dados para verificação: %w", err)
	}

	// Verificar assinatura
	if !ed25519.Verify(publicKey, dataToVerify, p.Signature) {
		return ErrInvalidSignature
	}

	return nil
}

// Encode serializa o pacote para transmissão
// Encode serializes the packet for transmission
// Encode serializa el paquete para transmisión
func (p *SignedPacket) Encode() ([]byte, error) {
	return json.Marshal(p)
}

// Decode deserializa um pacote recebido
// Decode deserializes a received packet
// Decode deserializa un paquete recibido
func Decode(data []byte) (*SignedPacket, error) {
	var packet SignedPacket
	if err := json.Unmarshal(data, &packet); err != nil {
		return nil, fmt.Errorf("erro ao decodificar pacote: %w", err)
	}
	return &packet, nil
}

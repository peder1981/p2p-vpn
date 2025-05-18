package packets

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// KeyManager gerencia as chaves criptográficas para assinatura de pacotes
// KeyManager manages cryptographic keys for packet signing
// KeyManager gestiona las claves criptográficas para firmar paquetes
type KeyManager struct {
	// privateKey é a chave privada usada para assinatura
	// privateKey is the private key used for signing
	// privateKey es la clave privada utilizada para firmar
	privateKey ed25519.PrivateKey

	// publicKey é a chave pública usada para verificação
	// publicKey is the public key used for verification
	// publicKey es la clave pública utilizada para verificación
	publicKey ed25519.PublicKey

	// peerKeys armazena as chaves públicas dos peers
	// peerKeys stores the public keys of peers
	// peerKeys almacena las claves públicas de los peers
	peerKeys     map[string]ed25519.PublicKey
	peerKeyMutex sync.RWMutex

	// keysDir é o diretório onde as chaves são armazenadas
	// keysDir is the directory where keys are stored
	// keysDir es el directorio donde se almacenan las claves
	keysDir string
}

// NewKeyManager cria uma nova instância do gerenciador de chaves
// NewKeyManager creates a new instance of the key manager
// NewKeyManager crea una nueva instancia del gestor de claves
func NewKeyManager(keysDir string) (*KeyManager, error) {
	// Criar diretório para chaves se não existir
	if err := os.MkdirAll(keysDir, 0700); err != nil {
		return nil, fmt.Errorf("erro ao criar diretório de chaves: %w", err)
	}

	km := &KeyManager{
		peerKeys: make(map[string]ed25519.PublicKey),
		keysDir:  keysDir,
	}

	// Verificar se já existem chaves
	if err := km.loadOrGenerateKeys(); err != nil {
		return nil, err
	}

	return km, nil
}

// loadOrGenerateKeys carrega as chaves existentes ou gera novas
// loadOrGenerateKeys loads existing keys or generates new ones
// loadOrGenerateKeys carga claves existentes o genera nuevas
func (km *KeyManager) loadOrGenerateKeys() error {
	privKeyPath := filepath.Join(km.keysDir, "private_key.pem")
	pubKeyPath := filepath.Join(km.keysDir, "public_key.pem")

	// Verificar se as chaves já existem
	if _, err := os.Stat(privKeyPath); err == nil {
		// Carregar chaves existentes
		var err error
		km.privateKey, err = loadPrivateKey(privKeyPath)
		if err != nil {
			return fmt.Errorf("erro ao carregar chave privada: %w", err)
		}

		km.publicKey, err = loadPublicKey(pubKeyPath)
		if err != nil {
			return fmt.Errorf("erro ao carregar chave pública: %w", err)
		}
	} else {
		// Gerar novo par de chaves
		var err error
		km.publicKey, km.privateKey, err = ed25519.GenerateKey(rand.Reader)
		if err != nil {
			return fmt.Errorf("erro ao gerar par de chaves: %w", err)
		}

		// Salvar as novas chaves
		if err := savePrivateKey(km.privateKey, privKeyPath); err != nil {
			return fmt.Errorf("erro ao salvar chave privada: %w", err)
		}

		if err := savePublicKey(km.publicKey, pubKeyPath); err != nil {
			return fmt.Errorf("erro ao salvar chave pública: %w", err)
		}
	}

	return nil
}

// GetPrivateKey retorna a chave privada para assinatura
// GetPrivateKey returns the private key for signing
// GetPrivateKey devuelve la clave privada para firmar
func (km *KeyManager) GetPrivateKey() ed25519.PrivateKey {
	return km.privateKey
}

// GetPublicKey retorna a chave pública para verificação
// GetPublicKey returns the public key for verification
// GetPublicKey devuelve la clave pública para verificación
func (km *KeyManager) GetPublicKey() ed25519.PublicKey {
	return km.publicKey
}

// GetPublicKeyPEM retorna a chave pública em formato PEM
// GetPublicKeyPEM returns the public key in PEM format
// GetPublicKeyPEM devuelve la clave pública en formato PEM
func (km *KeyManager) GetPublicKeyPEM() ([]byte, error) {
	pubKeyPath := filepath.Join(km.keysDir, "public_key.pem")
	return os.ReadFile(pubKeyPath)
}

// AddPeerKey adiciona uma chave pública de um peer
// AddPeerKey adds a public key of a peer
// AddPeerKey añade una clave pública de un peer
func (km *KeyManager) AddPeerKey(peerID string, publicKey ed25519.PublicKey) {
	km.peerKeyMutex.Lock()
	defer km.peerKeyMutex.Unlock()
	km.peerKeys[peerID] = publicKey
}

// AddPeerKeyFromPEM adiciona uma chave pública de um peer a partir do formato PEM
// AddPeerKeyFromPEM adds a peer public key from PEM format
// AddPeerKeyFromPEM añade una clave pública de un peer desde formato PEM
func (km *KeyManager) AddPeerKeyFromPEM(peerID string, pemData []byte) error {
	publicKey, err := parsePEMPublicKey(pemData)
	if err != nil {
		return err
	}
	
	km.AddPeerKey(peerID, publicKey)
	return nil
}

// GetPeerKey obtém a chave pública de um peer específico
// GetPeerKey gets the public key of a specific peer
// GetPeerKey obtiene la clave pública de un peer específico
func (km *KeyManager) GetPeerKey(peerID string) (ed25519.PublicKey, error) {
	km.peerKeyMutex.RLock()
	defer km.peerKeyMutex.RUnlock()
	
	publicKey, exists := km.peerKeys[peerID]
	if !exists {
		return nil, errors.New("chave pública do peer não encontrada")
	}
	
	return publicKey, nil
}

// StorePeerKey armazena a chave pública de um peer em arquivo
// StorePeerKey stores a peer's public key in a file
// StorePeerKey almacena la clave pública de un peer en un archivo
func (km *KeyManager) StorePeerKey(peerID string, publicKey ed25519.PublicKey) error {
	peerKeyPath := filepath.Join(km.keysDir, fmt.Sprintf("peer_%s.pem", peerID))
	return savePublicKey(publicKey, peerKeyPath)
}

// LoadPeerKeys carrega as chaves públicas de peers armazenadas
// LoadPeerKeys loads stored peer public keys
// LoadPeerKeys carga las claves públicas de peers almacenadas
func (km *KeyManager) LoadPeerKeys() error {
	// Localizar todos os arquivos de chaves de peers
	files, err := filepath.Glob(filepath.Join(km.keysDir, "peer_*.pem"))
	if err != nil {
		return fmt.Errorf("erro ao buscar chaves de peers: %w", err)
	}

	for _, file := range files {
		// Extrair ID do peer do nome do arquivo
		base := filepath.Base(file)
		peerID := base[5 : len(base)-4] // Remove "peer_" e ".pem"

		// Carregar a chave
		publicKey, err := loadPublicKey(file)
		if err != nil {
			return fmt.Errorf("erro ao carregar chave do peer %s: %w", peerID, err)
		}

		// Adicionar à coleção
		km.AddPeerKey(peerID, publicKey)
	}

	return nil
}

// loadPrivateKey carrega uma chave privada de um arquivo PEM
// loadPrivateKey loads a private key from a PEM file
// loadPrivateKey carga una clave privada desde un archivo PEM
func loadPrivateKey(filePath string) (ed25519.PrivateKey, error) {
	pemData, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	return parsePEMPrivateKey(pemData)
}

// loadPublicKey carrega uma chave pública de um arquivo PEM
// loadPublicKey loads a public key from a PEM file
// loadPublicKey carga una clave pública desde un archivo PEM
func loadPublicKey(filePath string) (ed25519.PublicKey, error) {
	pemData, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	return parsePEMPublicKey(pemData)
}

// savePrivateKey salva uma chave privada em formato PEM
// savePrivateKey saves a private key in PEM format
// savePrivateKey guarda una clave privada en formato PEM
func savePrivateKey(privateKey ed25519.PrivateKey, filePath string) error {
	// Codificar a chave privada para PKCS8
	pkcs8Bytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		return err
	}

	// Criar bloco PEM
	pemBlock := &pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: pkcs8Bytes,
	}

	// Escrever para arquivo com permissões restritas
	file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer file.Close()

	return pem.Encode(file, pemBlock)
}

// savePublicKey salva uma chave pública em formato PEM
// savePublicKey saves a public key in PEM format
// savePublicKey guarda una clave pública en formato PEM
func savePublicKey(publicKey ed25519.PublicKey, filePath string) error {
	// Codificar a chave pública para PKIX
	pkixBytes, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		return err
	}

	// Criar bloco PEM
	pemBlock := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pkixBytes,
	}

	// Escrever para arquivo
	file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	return pem.Encode(file, pemBlock)
}

// parsePEMPrivateKey analisa um bloco PEM para extrair uma chave privada
// parsePEMPrivateKey parses a PEM block to extract a private key
// parsePEMPrivateKey analiza un bloque PEM para extraer una clave privada
func parsePEMPrivateKey(pemData []byte) (ed25519.PrivateKey, error) {
	block, _ := pem.Decode(pemData)
	if block == nil || block.Type != "PRIVATE KEY" {
		return nil, errors.New("bloco PEM inválido ou não é uma chave privada")
	}

	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	edKey, ok := key.(ed25519.PrivateKey)
	if !ok {
		return nil, errors.New("a chave não é uma chave privada Ed25519")
	}

	return edKey, nil
}

// parsePEMPublicKey analisa um bloco PEM para extrair uma chave pública
// parsePEMPublicKey parses a PEM block to extract a public key
// parsePEMPublicKey analiza un bloque PEM para extraer una clave pública
func parsePEMPublicKey(pemData []byte) (ed25519.PublicKey, error) {
	block, _ := pem.Decode(pemData)
	if block == nil || block.Type != "PUBLIC KEY" {
		return nil, errors.New("bloco PEM inválido ou não é uma chave pública")
	}

	key, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	edKey, ok := key.(ed25519.PublicKey)
	if !ok {
		return nil, errors.New("a chave não é uma chave pública Ed25519")
	}

	return edKey, nil
}

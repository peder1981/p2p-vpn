package core

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"time"

	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
	"gopkg.in/yaml.v3"
)

// Config contém todas as configurações da VPN
// Config contains all VPN configuration
// Config contiene toda la configuración de la VPN
type Config struct {
	// Identificação do nó
	NodeID       string `yaml:"nodeId"`
	PrivateKey   string `yaml:"privateKey"`
	PublicKey    string `yaml:"publicKey"`
	
	// Configuração de rede
	VirtualIP    string `yaml:"virtualIp"`
	VirtualCIDR  string `yaml:"virtualCidr"`
	MTU          int    `yaml:"mtu,omitempty"`       // MTU da interface (padrão: 1420)
	DNS          []string `yaml:"dns,omitempty"`    // Servidores DNS
	
	// Lista de peers confiáveis
	TrustedPeers []TrustedPeer `yaml:"trustedPeers"`
	
	// Configuração da interface
	InterfaceName string `yaml:"interfaceName,omitempty"` // Nome da interface (padrão: wg0)
}

// TrustedPeer representa um peer remoto confiável
// TrustedPeer represents a trusted remote peer
// TrustedPeer representa un par remoto confiable
type TrustedPeer struct {
	NodeID      string `yaml:"nodeId"`
	PublicKey   string `yaml:"publicKey"`
	VirtualIP   string `yaml:"virtualIp"`
	Endpoints   []string `yaml:"endpoints,omitempty"`
	LastSeen    int64    `yaml:"lastSeen,omitempty"`
	
	// Campos adicionais para WireGuard
	AllowedIPs  []string `yaml:"allowedIps,omitempty"`  // IPs permitidos através deste peer
	KeepAlive   int      `yaml:"keepAlive,omitempty"`   // Intervalo de keepalive em segundos
}

// LoadConfig carrega a configuração a partir de um arquivo YAML
func LoadConfig(path string) (*Config, error) {
	// Verificar se o arquivo existe
	_, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	
	config := &Config{}
	if err := yaml.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("erro ao analisar arquivo de configuração: %w", err)
	}
	
	return config, nil
}

// SaveConfig salva a configuração atual em um arquivo YAML
func (c *Config) SaveConfig(path string) error {
	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("erro ao serializar configuração: %w", err)
	}
	
	if err := ioutil.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("erro ao salvar arquivo de configuração: %w", err)
	}
	
	return nil
}

// GenerateDefaultConfig cria uma nova configuração com valores padrão
// GenerateDefaultConfig creates a new configuration with default values
// GenerateDefaultConfig crea una nueva configuración con valores predeterminados
func GenerateDefaultConfig(path string) *Config {
	// Inicializar o gerador de números aleatórios
	rand.Seed(time.Now().UnixNano())
	// Gerar chave privada WireGuard
	privateKey, err := wgtypes.GeneratePrivateKey()
	if err != nil {
		// Em caso de erro, gerar uma chave manualmente
		fmt.Printf("Erro ao gerar chave WireGuard: %v. Usando método alternativo.\n", err)
		rawKey := make([]byte, 32)
		rand.Read(rawKey)
		// Usar a função correta para criar a chave privada
		var key wgtypes.Key
		copy(key[:], rawKey)
		privateKey = key
	}
	
	// Obter a chave pública correspondente
	publicKey := privateKey.PublicKey()
	
	// Converter para string em base64
	encodedPrivateKey := base64.StdEncoding.EncodeToString(privateKey[:])
	encodedPublicKey := base64.StdEncoding.EncodeToString(publicKey[:])
	
	// Gerar um IP virtual aleatório na faixa 10.0.0.0/8
	ip1 := 10
	ip2 := byte(rand.Intn(255))
	ip3 := byte(rand.Intn(255))
	ip4 := byte(1 + rand.Intn(254)) // Evitar .0 e .255
	
	config := &Config{
		NodeID:      fmt.Sprintf("node-%x%x%x", ip2, ip3, ip4),
		PrivateKey:  encodedPrivateKey,
		PublicKey:   encodedPublicKey,
		VirtualIP:   fmt.Sprintf("%d.%d.%d.%d", ip1, ip2, ip3, ip4),
		VirtualCIDR: fmt.Sprintf("%d.%d.%d.0/24", ip1, ip2, ip3),
		TrustedPeers: []TrustedPeer{},
	}
	
	// Salvar a configuração
	if err := config.SaveConfig(path); err != nil {
		fmt.Printf("Aviso: não foi possível salvar a configuração: %v\n", err)
	}
	
	return config
}

// AddTrustedPeer adiciona um peer confiável à configuração
func (c *Config) AddTrustedPeer(peer TrustedPeer) {
	// Verificar se o peer já existe
	for i, existingPeer := range c.TrustedPeers {
		if existingPeer.NodeID == peer.NodeID || existingPeer.PublicKey == peer.PublicKey {
			// Atualizar o peer existente
			c.TrustedPeers[i] = peer
			return
		}
	}
	
	// Adicionar novo peer
	c.TrustedPeers = append(c.TrustedPeers, peer)
}

// RemoveTrustedPeer remove um peer confiável da configuração
func (c *Config) RemoveTrustedPeer(nodeID string) bool {
	for i, peer := range c.TrustedPeers {
		if peer.NodeID == nodeID {
			// Remover o peer
			c.TrustedPeers = append(c.TrustedPeers[:i], c.TrustedPeers[i+1:]...)
			return true
		}
	}
	return false
}

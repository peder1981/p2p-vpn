package platform

import (
	"fmt"
	"sync"
)

// VPNPlatform representa uma interface para operações específicas de plataforma
// VPNPlatform represents an interface for platform-specific operations
// VPNPlatform representa una interfaz para operaciones específicas de plataforma
type VPNPlatform interface {
	// Nome da plataforma
	// Platform name
	// Nombre de la plataforma
	Name() string
	
	// Verifica se a plataforma atual é suportada
	// Checks if the current platform is supported
	// Verifica si la plataforma actual es soportada
	IsSupported() bool
	
	// Cria e configura uma interface WireGuard
	// Creates and configures a WireGuard interface
	// Crea y configura una interfaz WireGuard
	CreateWireGuardInterface(interfaceName string, listenPort int, privateKey string) error
	
	// Remove uma interface WireGuard existente
	RemoveWireGuardInterface(interfaceName string) error
	
	// Configura o endereço IP na interface
	ConfigureInterfaceAddress(interfaceName, address, subnet string) error
	
	// Adiciona um peer à interface WireGuard
	AddPeer(interfaceName, publicKeyStr, allowedIPs, endpointStr string, keepAlive int) error
	
	// Remove um peer da interface WireGuard
	RemovePeer(interfaceName, publicKeyStr string) error
	
	// Configura rotas para o tráfego VPN
	ConfigureRouting(interfaceName, vpnCIDR string) error
	
	// Obtém o status da interface WireGuard
	GetInterfaceStatus(interfaceName string) (bool, error)
}

// PlatformFactory é um tipo de função que tenta criar uma implementação VPNPlatform
type PlatformFactory func() (VPNPlatform, error)

// Registro global de fábricas de plataformas
var (
	platformFactories []PlatformFactory
	platformMutex     sync.Mutex
)

// RegisterPlatform registra uma nova fábrica de plataforma
// As fábricas são tentadas na ordem em que são registradas
func RegisterPlatform(factory PlatformFactory) {
	platformMutex.Lock()
	defer platformMutex.Unlock()
	
	platformFactories = append(platformFactories, factory)
}

// GetPlatform retorna a implementação da plataforma atual
// GetPlatform returns the current platform implementation
// GetPlatform devuelve la implementación de la plataforma actual
func GetPlatform() (VPNPlatform, error) {
	platformMutex.Lock()
	defer platformMutex.Unlock()
	
	// Tentar cada fábrica registrada
	for _, factory := range platformFactories {
		platform, err := factory()
		if err == nil && platform != nil {
			return platform, nil
		}
		// Se a fábrica retornou um erro, continuar com a próxima
	}
	
	return nil, ErrPlatformNotSupported
}

// ErrPlatformNotSupported é retornado quando a plataforma atual não é suportada
var ErrPlatformNotSupported = fmt.Errorf("plataforma não suportada")

// +build linux

package platform

import (
	"encoding/base64"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"time"

	"github.com/vishvananda/netlink"
	"golang.zx2c4.com/wireguard/wgctrl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

// LinuxPlatform implementa a interface VPNPlatform para Linux
// LinuxPlatform implements the VPNPlatform interface for Linux
// LinuxPlatform implementa la interfaz VPNPlatform para Linux
type LinuxPlatform struct{}

// Nome da plataforma
func (p *LinuxPlatform) Name() string {
	return "Linux"
}

// Verifica se a plataforma é suportada
func (p *LinuxPlatform) IsSupported() bool {
	// Verificar se o módulo do kernel wireguard está carregado
	_, err := os.Stat("/sys/module/wireguard")
	if err == nil {
		return true
	}

	// Verificar se o utilitário wireguard-go está instalado (userspace)
	_, err = os.Stat("/usr/bin/wireguard-go")
	if err == nil {
		return true
	}

	return false
}

// Cria e configura uma interface WireGuard
func (p *LinuxPlatform) CreateWireGuardInterface(interfaceName string, listenPort int, privateKeyStr string) error {
	// Criar a interface wireguard usando netlink
	linkAttrs := netlink.NewLinkAttrs()
	linkAttrs.Name = interfaceName
	linkAttrs.MTU = 1420
	
	wgLink := &netlink.GenericLink{
		LinkAttrs: linkAttrs,
		LinkType:  "wireguard",
	}
	
	// Adicionar link
	if err := netlink.LinkAdd(wgLink); err != nil {
		return fmt.Errorf("erro ao criar interface wireguard: %w", err)
	}
	
	// Ativar a interface
	if err := netlink.LinkSetUp(wgLink); err != nil {
		netlink.LinkDel(wgLink)
		return fmt.Errorf("erro ao ativar interface wireguard: %w", err)
	}
	
	// Decodificar a chave privada
	privateKeyBytes, err := base64.StdEncoding.DecodeString(privateKeyStr)
	if err != nil {
		return fmt.Errorf("erro ao decodificar chave privada: %w", err)
	}
	
	var privateKey wgtypes.Key
	copy(privateKey[:], privateKeyBytes)
	
	// Criar cliente WireGuard
	wgClient, err := wgctrl.New()
	if err != nil {
		return fmt.Errorf("erro ao criar cliente WireGuard: %w", err)
	}
	defer wgClient.Close()
	
	// Configurar a interface com chave privada e porta
	deviceConfig := wgtypes.Config{
		PrivateKey: &privateKey,
		ListenPort: &listenPort,
	}
	
	if err := wgClient.ConfigureDevice(interfaceName, deviceConfig); err != nil {
		netlink.LinkDel(wgLink)
		return fmt.Errorf("erro ao configurar interface wireguard: %w", err)
	}
	
	return nil
}

// Remove uma interface WireGuard
func (p *LinuxPlatform) RemoveWireGuardInterface(interfaceName string) error {
	link, err := netlink.LinkByName(interfaceName)
	if err != nil {
		return fmt.Errorf("interface %s não encontrada: %w", interfaceName, err)
	}
	
	// Desativar a interface
	if err := netlink.LinkSetDown(link); err != nil {
		return fmt.Errorf("erro ao desativar interface %s: %w", interfaceName, err)
	}
	
	// Remover a interface
	if err := netlink.LinkDel(link); err != nil {
		return fmt.Errorf("erro ao remover interface %s: %w", interfaceName, err)
	}
	
	return nil
}

// Configura o endereço IP na interface
func (p *LinuxPlatform) ConfigureInterfaceAddress(interfaceName, address, subnet string) error {
	// Obter interface
	link, err := netlink.LinkByName(interfaceName)
	if err != nil {
		return fmt.Errorf("interface %s não encontrada: %w", interfaceName, err)
	}
	
	// Converter endereço para o formato CIDR
	ipNet, err := netlink.ParseIPNet(address + "/" + subnet)
	if err != nil {
		return fmt.Errorf("erro ao analisar endereço IP: %w", err)
	}
	
	// Adicionar endereço à interface
	addr := &netlink.Addr{
		IPNet: ipNet,
	}
	
	if err := netlink.AddrAdd(link, addr); err != nil {
		return fmt.Errorf("erro ao adicionar endereço à interface: %w", err)
	}
	
	return nil
}

// Adiciona um peer à interface WireGuard
func (p *LinuxPlatform) AddPeer(interfaceName, publicKeyStr, allowedIPs, endpointStr string, keepAlive int) error {
	// Decodificar a chave pública
	publicKeyBytes, err := base64.StdEncoding.DecodeString(publicKeyStr)
	if err != nil {
		return fmt.Errorf("erro ao decodificar chave pública: %w", err)
	}
	
	var publicKey wgtypes.Key
	copy(publicKey[:], publicKeyBytes)
	
	// Criar cliente WireGuard
	wgClient, err := wgctrl.New()
	if err != nil {
		return fmt.Errorf("erro ao criar cliente WireGuard: %w", err)
	}
	defer wgClient.Close()
	
	// Resolver endpoint
	var endpoint *net.UDPAddr
	if endpointStr != "" {
		endpoint, err = net.ResolveUDPAddr("udp", endpointStr)
		if err != nil {
			return fmt.Errorf("erro ao resolver endpoint: %w", err)
		}
	}
	
	// Analisar AllowedIPs
	var allowedIPsNet []net.IPNet
	if allowedIPs != "" {
		ipNet, err := netlink.ParseIPNet(allowedIPs)
		if err != nil {
			return fmt.Errorf("erro ao analisar AllowedIPs: %w", err)
		}
		allowedIPsNet = append(allowedIPsNet, *ipNet)
	}
	
	// Configurar keepalive
	var persistentKeepalive *time.Duration
	if keepAlive > 0 {
		keepAliveDuration := time.Duration(keepAlive) * time.Second
		persistentKeepalive = &keepAliveDuration
	}
	
	// Configurar peer
	peerConfig := wgtypes.PeerConfig{
		PublicKey:                   publicKey,
		Endpoint:                    endpoint,
		AllowedIPs:                  allowedIPsNet,
		PersistentKeepaliveInterval: persistentKeepalive,
	}
	
	// Configurar dispositivo
	deviceConfig := wgtypes.Config{
		Peers: []wgtypes.PeerConfig{peerConfig},
	}
	
	if err := wgClient.ConfigureDevice(interfaceName, deviceConfig); err != nil {
		return fmt.Errorf("erro ao adicionar peer: %w", err)
	}
	
	return nil
}

// Remove um peer da interface WireGuard
func (p *LinuxPlatform) RemovePeer(interfaceName, publicKeyStr string) error {
	// Decodificar a chave pública
	publicKeyBytes, err := base64.StdEncoding.DecodeString(publicKeyStr)
	if err != nil {
		return fmt.Errorf("erro ao decodificar chave pública: %w", err)
	}
	
	var publicKey wgtypes.Key
	copy(publicKey[:], publicKeyBytes)
	
	// Criar cliente WireGuard
	wgClient, err := wgctrl.New()
	if err != nil {
		return fmt.Errorf("erro ao criar cliente WireGuard: %w", err)
	}
	defer wgClient.Close()
	
	// Configurar peer para remoção
	peerConfig := wgtypes.PeerConfig{
		PublicKey: publicKey,
		Remove:    true,
	}
	
	// Configurar dispositivo
	deviceConfig := wgtypes.Config{
		Peers: []wgtypes.PeerConfig{peerConfig},
	}
	
	if err := wgClient.ConfigureDevice(interfaceName, deviceConfig); err != nil {
		return fmt.Errorf("erro ao remover peer: %w", err)
	}
	
	return nil
}

// Configura rotas para o tráfego VPN
func (p *LinuxPlatform) ConfigureRouting(interfaceName, vpnCIDR string) error {
	// Obter interface
	link, err := netlink.LinkByName(interfaceName)
	if err != nil {
		return fmt.Errorf("interface %s não encontrada: %w", interfaceName, err)
	}
	
	// Analisar CIDR
	_, dst, err := net.ParseCIDR(vpnCIDR)
	if err != nil {
		return fmt.Errorf("erro ao analisar CIDR: %w", err)
	}
	
	// Adicionar rota
	route := &netlink.Route{
		LinkIndex: link.Attrs().Index,
		Dst:       dst,
	}
	
	if err := netlink.RouteAdd(route); err != nil {
		return fmt.Errorf("erro ao adicionar rota: %w", err)
	}
	
	return nil
}

// Retorna o caminho para a configuração do WireGuard
func (p *LinuxPlatform) WireGuardConfigPath(interfaceName string) string {
	return filepath.Join("/etc/wireguard", interfaceName+".conf")
}

// Obtém o status da interface WireGuard
func (p *LinuxPlatform) GetInterfaceStatus(interfaceName string) (bool, error) {
	// Verificar se a interface existe
	link, err := netlink.LinkByName(interfaceName)
	if err != nil {
		return false, nil // Interface não existe, não é um erro
	}
	
	// Verificar se a interface está ativa
	return link.Attrs().Flags&net.FlagUp != 0, nil
}

// init registra a plataforma Linux
func init() {
	// Substituir a função GetPlatform se estamos em Linux
	if os.Getenv("GOOS") == "linux" || os.Getenv("GOOS") == "" {
		originalGetPlatform := GetPlatform
		GetPlatform = func() (VPNPlatform, error) {
			platform := &LinuxPlatform{}
			if platform.IsSupported() {
				return platform, nil
			}
			return originalGetPlatform()
		}
	}
}

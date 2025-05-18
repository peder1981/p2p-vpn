// +build darwin

package platform

import (
	"encoding/base64"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"golang.zx2c4.com/wireguard/wgctrl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

// DarwinPlatform implementa a interface VPNPlatform para macOS
// DarwinPlatform implements the VPNPlatform interface for macOS
// DarwinPlatform implementa la interfaz VPNPlatform para macOS
type DarwinPlatform struct {
	wireguardGoPath string
}

// Nome da plataforma
func (p *DarwinPlatform) Name() string {
	return "macOS"
}

// Verifica se a plataforma é suportada
func (p *DarwinPlatform) IsSupported() bool {
	// Verificar se o wireguard-go está instalado via Homebrew
	brewWireguardGo := "/usr/local/bin/wireguard-go"
	_, err := os.Stat(brewWireguardGo)
	if err == nil {
		p.wireguardGoPath = brewWireguardGo
		return true
	}
	
	// Verificar no Apple Silicon (M1/M2) com Homebrew
	brewWireguardGoM1 := "/opt/homebrew/bin/wireguard-go"
	_, err = os.Stat(brewWireguardGoM1)
	if err == nil {
		p.wireguardGoPath = brewWireguardGoM1
		return true
	}
	
	// Verificar via PATH
	path, err := exec.LookPath("wireguard-go")
	if err == nil {
		p.wireguardGoPath = path
		return true
	}
	
	return false
}

// Cria e configura uma interface WireGuard
func (p *DarwinPlatform) CreateWireGuardInterface(interfaceName string, listenPort int, privateKeyStr string) error {
	// No macOS, primeiro vamos criar o utun usando wireguard-go
	cmd := exec.Command(p.wireguardGoPath, interfaceName)
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("erro ao iniciar wireguard-go: %w", err)
	}
	
	// Aguardar um momento para a interface ser criada
	time.Sleep(500 * time.Millisecond)
	
	// Verificar se a interface foi criada
	ifconfigCmd := exec.Command("ifconfig", interfaceName)
	if output, err := ifconfigCmd.CombinedOutput(); err != nil {
		cmd.Process.Kill()
		return fmt.Errorf("erro ao verificar interface %s: %w - %s", interfaceName, err, string(output))
	}
	
	// Agora podemos configurar a interface usando wgctrl
	// Decodificar a chave privada
	privateKeyBytes, err := base64.StdEncoding.DecodeString(privateKeyStr)
	if err != nil {
		cmd.Process.Kill()
		return fmt.Errorf("erro ao decodificar chave privada: %w", err)
	}
	
	var privateKey wgtypes.Key
	copy(privateKey[:], privateKeyBytes)
	
	// Criar cliente WireGuard
	wgClient, err := wgctrl.New()
	if err != nil {
		cmd.Process.Kill()
		return fmt.Errorf("erro ao criar cliente WireGuard: %w", err)
	}
	defer wgClient.Close()
	
	// Configurar a interface com chave privada e porta
	deviceConfig := wgtypes.Config{
		PrivateKey: &privateKey,
		ListenPort: &listenPort,
	}
	
	// Aguardar mais um momento para garantir que a interface esteja pronta
	time.Sleep(500 * time.Millisecond)
	
	if err := wgClient.ConfigureDevice(interfaceName, deviceConfig); err != nil {
		cmd.Process.Kill()
		return fmt.Errorf("erro ao configurar interface wireguard: %w", err)
	}
	
	// Salvar o processo do wireguard-go para poder referenciá-lo depois
	// Isso é apenas para desenvolvimento - em produção precisaríamos criar um serviço launchd
	configDir := filepath.Dir(p.WireGuardConfigPath(interfaceName))
	os.MkdirAll(configDir, 0700)
	pidFile := filepath.Join(configDir, interfaceName+".pid")
	os.WriteFile(pidFile, []byte(fmt.Sprintf("%d", cmd.Process.Pid)), 0600)
	
	return nil
}

// Remove uma interface WireGuard
func (p *DarwinPlatform) RemoveWireGuardInterface(interfaceName string) error {
	// Verificar se o arquivo PID existe
	configDir := filepath.Dir(p.WireGuardConfigPath(interfaceName))
	pidFile := filepath.Join(configDir, interfaceName+".pid")
	
	// Tentar encerrar o processo wireguard-go
	pidData, err := os.ReadFile(pidFile)
	if err == nil {
		pidStr := strings.TrimSpace(string(pidData))
		killCmd := exec.Command("kill", pidStr)
		killCmd.Run() // Ignorar erros, o processo pode não existir mais
		os.Remove(pidFile)
	}
	
	// Remover a interface de rede
	cmd := exec.Command("ifconfig", interfaceName, "down")
	cmd.Run() // Ignorar erro se a interface não existir
	
	// Verificar se estava sendo executado pela App WireGuard, nesse caso tentar parar
	// via wg-quick
	wgQuickCmd := exec.Command("wg-quick", "down", interfaceName)
	wgQuickCmd.Run() // Ignorar erros
	
	return nil
}

// Configura o endereço IP na interface
func (p *DarwinPlatform) ConfigureInterfaceAddress(interfaceName, address, subnet string) error {
	// Usar o comando ifconfig para configurar o endereço
	cmd := exec.Command("ifconfig", interfaceName, "inet", address+"/"+subnet)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("erro ao configurar endereço IP (%s): %w", string(output), err)
	}
	
	// Ativar a interface
	cmd = exec.Command("ifconfig", interfaceName, "up")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("erro ao ativar interface (%s): %w", string(output), err)
	}
	
	return nil
}

// Adiciona um peer à interface WireGuard
func (p *DarwinPlatform) AddPeer(interfaceName, publicKeyStr, allowedIPs, endpointStr string, keepAlive int) error {
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
		_, ipNet, err := net.ParseCIDR(allowedIPs)
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
func (p *DarwinPlatform) RemovePeer(interfaceName, publicKeyStr string) error {
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
func (p *DarwinPlatform) ConfigureRouting(interfaceName, vpnCIDR string) error {
	// Extrair rede e máscara do CIDR
	_, ipNet, err := net.ParseCIDR(vpnCIDR)
	if err != nil {
		return fmt.Errorf("erro ao analisar CIDR: %w", err)
	}
	
	// Configurar rota usando o comando route
	cmd := exec.Command("route", "add", "-net", ipNet.String(), "-interface", interfaceName)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("erro ao adicionar rota (%s): %w", string(output), err)
	}
	
	return nil
}

// Retorna o caminho para a configuração do WireGuard
func (p *DarwinPlatform) WireGuardConfigPath(interfaceName string) string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "/Users/" + os.Getenv("USER")
	}
	
	return filepath.Join(homeDir, "Library", "Application Support", "WireGuard", interfaceName+".conf")
}

// Obtém o status da interface WireGuard
func (p *DarwinPlatform) GetInterfaceStatus(interfaceName string) (bool, error) {
	// Verificar se a interface existe e está ativa
	cmd := exec.Command("ifconfig", interfaceName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return false, nil // Interface não existe, não é um erro
	}
	
	// Verificar se a interface está ativa (UP)
	return strings.Contains(string(output), "UP") || strings.Contains(string(output), "RUNNING"), nil
}

// init registra a plataforma macOS
func init() {
	// Registrar a plataforma macOS no sistema
	if os.Getenv("GOOS") == "darwin" {
		RegisterPlatform(func() (VPNPlatform, error) {
			platform := &DarwinPlatform{}
			if platform.IsSupported() {
				return platform, nil
			}
			return nil, fmt.Errorf("plataforma macOS não suportada")
		})
	}
}

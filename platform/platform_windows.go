// +build windows

package platform

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// WindowsPlatform implementa a interface VPNPlatform para Windows
// WindowsPlatform implements the VPNPlatform interface for Windows
// WindowsPlatform implementa la interfaz VPNPlatform para Windows
type WindowsPlatform struct {
	programFilesPath string
	wireguardPath    string
}

// Nome da plataforma
func (p *WindowsPlatform) Name() string {
	return "Windows"
}

// Verifica se a plataforma é suportada
func (p *WindowsPlatform) IsSupported() bool {
	// Verificar se o WireGuard está instalado no Windows
	programFiles := os.Getenv("ProgramFiles")
	if programFiles == "" {
		programFiles = "C:\\Program Files"
	}
	
	wireguardPath := filepath.Join(programFiles, "WireGuard")
	wireguardExe := filepath.Join(wireguardPath, "wireguard.exe")
	
	_, err := os.Stat(wireguardExe)
	if err == nil {
		p.programFilesPath = programFiles
		p.wireguardPath = wireguardPath
		return true
	}
	
	// Também verificar Program Files (x86) para sistemas de 64 bits
	programFilesX86 := os.Getenv("ProgramFiles(x86)")
	if programFilesX86 != "" {
		wireguardPath = filepath.Join(programFilesX86, "WireGuard")
		wireguardExe = filepath.Join(wireguardPath, "wireguard.exe")
		
		_, err = os.Stat(wireguardExe)
		if err == nil {
			p.programFilesPath = programFilesX86
			p.wireguardPath = wireguardPath
			return true
		}
	}
	
	return false
}

// Cria e configura uma interface WireGuard
func (p *WindowsPlatform) CreateWireGuardInterface(interfaceName string, listenPort int, privateKeyStr string) error {
	// No Windows, o WireGuard é baseado em arquivo de configuração
	// Então, vamos criar o arquivo de configuração primeiro
	configPath := p.WireGuardConfigPath(interfaceName)
	
	// Garantir que o diretório exista
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0700); err != nil {
		return fmt.Errorf("erro ao criar diretório para configuração: %w", err)
	}
	
	// Montar arquivo de configuração
	config := fmt.Sprintf("[Interface]\nPrivateKey = %s\nListenPort = %d\n", privateKeyStr, listenPort)
	
	// Gravar arquivo de configuração
	if err := os.WriteFile(configPath, []byte(config), 0600); err != nil {
		return fmt.Errorf("erro ao gravar arquivo de configuração: %w", err)
	}
	
	// Usar o aplicativo wireguard-windows para iniciar o túnel
	cmd := exec.Command(filepath.Join(p.wireguardPath, "wireguard.exe"), "/installtunnelservice", configPath)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("erro ao instalar serviço de túnel (%s): %w", string(output), err)
	}
	
	return nil
}

// Remove uma interface WireGuard
func (p *WindowsPlatform) RemoveWireGuardInterface(interfaceName string) error {
	configPath := p.WireGuardConfigPath(interfaceName)
	
	// Verificar se o arquivo de configuração existe
	if _, err := os.Stat(configPath); err != nil {
		return fmt.Errorf("arquivo de configuração não encontrado para %s: %w", interfaceName, err)
	}
	
	// Usar o aplicativo wireguard-windows para parar e remover o túnel
	cmd := exec.Command(filepath.Join(p.wireguardPath, "wireguard.exe"), "/uninstalltunnelservice", interfaceName)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("erro ao remover serviço de túnel (%s): %w", string(output), err)
	}
	
	// Remover arquivo de configuração
	if err := os.Remove(configPath); err != nil {
		fmt.Printf("Aviso: erro ao remover arquivo de configuração %s: %v\n", configPath, err)
	}
	
	return nil
}

// Configura o endereço IP na interface
func (p *WindowsPlatform) ConfigureInterfaceAddress(interfaceName, address, subnet string) error {
	// No Windows, o endereço IP é configurado no arquivo de configuração do WireGuard
	configPath := p.WireGuardConfigPath(interfaceName)
	
	// Ler configuração atual
	configData, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("erro ao ler arquivo de configuração: %w", err)
	}
	
	config := string(configData)
	
	// Verificar se já existe Address no arquivo
	if strings.Contains(config, "Address = ") {
		// Substituir endereço existente
		lines := strings.Split(config, "\n")
		for i, line := range lines {
			if strings.HasPrefix(line, "Address = ") {
				lines[i] = fmt.Sprintf("Address = %s/%s", address, subnet)
				break
			}
		}
		config = strings.Join(lines, "\n")
	} else {
		// Adicionar endereço à seção [Interface]
		interfaceSection := "[Interface]"
		replacementText := fmt.Sprintf("%s\nAddress = %s/%s", interfaceSection, address, subnet)
		config = strings.Replace(config, interfaceSection, replacementText, 1)
	}
	
	// Gravar configuração atualizada
	if err := os.WriteFile(configPath, []byte(config), 0600); err != nil {
		return fmt.Errorf("erro ao gravar arquivo de configuração: %w", err)
	}
	
	// Reiniciar o serviço para aplicar as alterações
	if err := p.restartWireGuardService(interfaceName); err != nil {
		return fmt.Errorf("erro ao reiniciar serviço: %w", err)
	}
	
	return nil
}

// Adiciona um peer à interface WireGuard
func (p *WindowsPlatform) AddPeer(interfaceName, publicKeyStr, allowedIPs, endpointStr string, keepAlive int) error {
	configPath := p.WireGuardConfigPath(interfaceName)
	
	// Ler configuração atual
	configData, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("erro ao ler arquivo de configuração: %w", err)
	}
	
	config := string(configData)
	
	// Verificar se o peer já existe
	peerSection := fmt.Sprintf("[Peer]\nPublicKey = %s", publicKeyStr)
	if strings.Contains(config, peerSection) {
		// Removemos o peer existente para substituí-lo
		sections := strings.Split(config, "[Peer]")
		for i, section := range sections {
			if strings.Contains(section, fmt.Sprintf("PublicKey = %s", publicKeyStr)) {
				sections = append(sections[:i], sections[i+1:]...)
				break
			}
		}
		config = strings.Join(sections, "[Peer]")
	}
	
	// Montar configuração do peer
	peerConfig := fmt.Sprintf("\n\n[Peer]\nPublicKey = %s\nAllowedIPs = %s", publicKeyStr, allowedIPs)
	
	if endpointStr != "" {
		peerConfig += fmt.Sprintf("\nEndpoint = %s", endpointStr)
	}
	
	if keepAlive > 0 {
		peerConfig += fmt.Sprintf("\nPersistentKeepalive = %d", keepAlive)
	}
	
	// Adicionar peer à configuração
	config += peerConfig
	
	// Gravar configuração atualizada
	if err := os.WriteFile(configPath, []byte(config), 0600); err != nil {
		return fmt.Errorf("erro ao gravar arquivo de configuração: %w", err)
	}
	
	// Reiniciar o serviço para aplicar as alterações
	if err := p.restartWireGuardService(interfaceName); err != nil {
		return fmt.Errorf("erro ao reiniciar serviço: %w", err)
	}
	
	return nil
}

// Remove um peer da interface WireGuard
func (p *WindowsPlatform) RemovePeer(interfaceName, publicKeyStr string) error {
	configPath := p.WireGuardConfigPath(interfaceName)
	
	// Ler configuração atual
	configData, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("erro ao ler arquivo de configuração: %w", err)
	}
	
	config := string(configData)
	
	// Verificar se o peer existe
	if !strings.Contains(config, fmt.Sprintf("PublicKey = %s", publicKeyStr)) {
		return nil // Peer não existe, nada a fazer
	}
	
	// Remover o peer
	sections := strings.Split(config, "[Peer]")
	newSections := []string{sections[0]} // Manter a seção [Interface]
	
	for i := 1; i < len(sections); i++ {
		if !strings.Contains(sections[i], fmt.Sprintf("PublicKey = %s", publicKeyStr)) {
			newSections = append(newSections, "[Peer]"+sections[i])
		}
	}
	
	newConfig := strings.Join(newSections, "")
	
	// Gravar configuração atualizada
	if err := os.WriteFile(configPath, []byte(newConfig), 0600); err != nil {
		return fmt.Errorf("erro ao gravar arquivo de configuração: %w", err)
	}
	
	// Reiniciar o serviço para aplicar as alterações
	if err := p.restartWireGuardService(interfaceName); err != nil {
		return fmt.Errorf("erro ao reiniciar serviço: %w", err)
	}
	
	return nil
}

// Configura rotas para o tráfego VPN
func (p *WindowsPlatform) ConfigureRouting(interfaceName, vpnCIDR string) error {
	// A configuração de rotas no Windows é feita automaticamente pelo WireGuard
	// baseado no AllowedIPs, não precisamos fazer nada aqui
	return nil
}

// Retorna o caminho para a configuração do WireGuard
func (p *WindowsPlatform) WireGuardConfigPath(interfaceName string) string {
	dataDir := os.Getenv("LOCALAPPDATA")
	if dataDir == "" {
		dataDir = filepath.Join(os.Getenv("USERPROFILE"), "AppData", "Local")
	}
	
	return filepath.Join(dataDir, "WireGuard", "Configurations", interfaceName+".conf")
}

// Obtém o status da interface WireGuard
func (p *WindowsPlatform) GetInterfaceStatus(interfaceName string) (bool, error) {
	// Verificar se o arquivo de configuração existe
	configPath := p.WireGuardConfigPath(interfaceName)
	if _, err := os.Stat(configPath); err != nil {
		return false, nil // Interface não existe, não é um erro
	}
	
	// Verificar status do serviço
	cmd := exec.Command("sc", "query", "WireGuardTunnel$"+interfaceName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return false, nil // Serviço não existe ou não está em execução
	}
	
	// Verificar se o serviço está em execução
	return strings.Contains(string(output), "RUNNING"), nil
}

// Reinicia o serviço WireGuard
func (p *WindowsPlatform) restartWireGuardService(interfaceName string) error {
	// Parar o serviço
	cmd := exec.Command("sc", "stop", "WireGuardTunnel$"+interfaceName)
	if _, err := cmd.CombinedOutput(); err != nil {
		// Ignorar erros, pode ser que o serviço não esteja em execução
	}
	
	// Aguardar um momento
	time.Sleep(1 * time.Second)
	
	// Iniciar o serviço
	cmd = exec.Command("sc", "start", "WireGuardTunnel$"+interfaceName)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("erro ao iniciar serviço (%s): %w", string(output), err)
	}
	
	return nil
}

// init registra a plataforma Windows
func init() {
	// Registrar a plataforma Windows no sistema
	if os.Getenv("GOOS") == "windows" {
		RegisterPlatform(func() (VPNPlatform, error) {
			platform := &WindowsPlatform{}
			if platform.IsSupported() {
				return platform, nil
			}
			return nil, fmt.Errorf("plataforma Windows não suportada")
		})
	}
}

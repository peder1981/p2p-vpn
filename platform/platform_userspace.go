package platform

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// UserspaceWireguardPlatform implementa a interface VPNPlatform usando boringtun
// UserspaceWireguardPlatform implements the VPNPlatform interface using boringtun
// UserspaceWireguardPlatform implementa la interfaz VPNPlatform usando boringtun
type UserspaceWireguardPlatform struct {
	boringtunPath  string
	wireguardGoPath string
	tunCtlPath     string
	wgToolPath     string
	ipToolPath     string
	
	// Processo do daemon userspace WireGuard
	wgProcess     *os.Process
	configDir     string
	mutex         sync.Mutex
}

// Nome da plataforma
func (p *UserspaceWireguardPlatform) Name() string {
	if p.boringtunPath != "" {
		return "WireGuard Userspace (boringtun)"
	}
	return "WireGuard Userspace (wireguard-go)"
}

// Verifica se a plataforma é suportada
func (p *UserspaceWireguardPlatform) IsSupported() bool {
	// Verificar se o boringtun está instalado
	boringtunPath, err := exec.LookPath("boringtun")
	if err == nil {
		p.boringtunPath = boringtunPath
		fmt.Println("Encontrado boringtun em:", boringtunPath)
	} else {
		// Verificar se o wireguard-go está instalado
		wireguardGoPath, err := exec.LookPath("wireguard-go")
		if err != nil {
			return false
		}
		p.wireguardGoPath = wireguardGoPath
		fmt.Println("Encontrado wireguard-go em:", wireguardGoPath)
	}
	
	// Verificar ferramentas necessárias (wg, ip, etc.)
	wgToolPath, err := exec.LookPath("wg")
	if err == nil {
		p.wgToolPath = wgToolPath
	} else {
		return false
	}
	
	ipToolPath, err := exec.LookPath("ip")
	if err == nil {
		p.ipToolPath = ipToolPath
	} else {
		return false
	}
	
	// Obter diretório de configuração
	homeDir, err := os.UserHomeDir()
	if err == nil {
		p.configDir = filepath.Join(homeDir, ".wireguard")
	} else {
		p.configDir = "/tmp/wireguard"
	}
	
	// Garantir que o diretório de configuração exista
	os.MkdirAll(p.configDir, 0700)
	
	return true
}

// Cria e configura uma interface WireGuard
func (p *UserspaceWireguardPlatform) CreateWireGuardInterface(interfaceName string, listenPort int, privateKeyStr string) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	
	// Verificar se já existe um processo em execução
	if p.wgProcess != nil {
		return fmt.Errorf("já existe um processo WireGuard em execução")
	}
	
	// Criar arquivo de configuração
	configPath := p.WireGuardConfigPath(interfaceName)
	config := fmt.Sprintf("[Interface]\nPrivateKey = %s\nListenPort = %d\n", privateKeyStr, listenPort)
	
	// Garantir que o diretório exista
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0700); err != nil {
		return fmt.Errorf("erro ao criar diretório para configuração: %w", err)
	}
	
	// Gravar configuração
	if err := os.WriteFile(configPath, []byte(config), 0600); err != nil {
		return fmt.Errorf("erro ao gravar arquivo de configuração: %w", err)
	}
	
	// Iniciar o processo userspace WireGuard
	var cmd *exec.Cmd
	if p.boringtunPath != "" {
		// Usando boringtun
		os.Setenv("WG_QUICK_USERSPACE_IMPLEMENTATION", "boringtun")
		cmd = exec.Command(p.boringtunPath, interfaceName)
	} else {
		// Usando wireguard-go
		cmd = exec.Command(p.wireguardGoPath, interfaceName)
	}
	
	// Redirecionar saída para logs
	stdout, err := os.Create(filepath.Join(p.configDir, fmt.Sprintf("%s-stdout.log", interfaceName)))
	if err != nil {
		return fmt.Errorf("erro ao criar arquivo de log: %w", err)
	}
	
	stderr, err := os.Create(filepath.Join(p.configDir, fmt.Sprintf("%s-stderr.log", interfaceName)))
	if err != nil {
		stdout.Close()
		return fmt.Errorf("erro ao criar arquivo de log de erro: %w", err)
	}
	
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	
	// Iniciar em segundo plano
	if err := cmd.Start(); err != nil {
		stdout.Close()
		stderr.Close()
		return fmt.Errorf("erro ao iniciar processo userspace: %w", err)
	}
	
	// Salvar referência ao processo
	p.wgProcess = cmd.Process
	
	// Guardar PID para referência futura
	pidFile := filepath.Join(p.configDir, fmt.Sprintf("%s.pid", interfaceName))
	if err := os.WriteFile(pidFile, []byte(fmt.Sprintf("%d", cmd.Process.Pid)), 0600); err != nil {
		fmt.Printf("Aviso: não foi possível salvar o PID: %v\n", err)
	}
	
	// Aguardar um momento para a interface ser criada
	time.Sleep(1 * time.Second)
	
	// Configurar a interface usando o utilitário wg
	wgCmd := exec.Command(p.wgToolPath, "setconf", interfaceName, configPath)
	if output, err := wgCmd.CombinedOutput(); err != nil {
		_ = p.RemoveWireGuardInterface(interfaceName) // Tentar limpar
		return fmt.Errorf("erro ao configurar interface WireGuard (%s): %w", string(output), err)
	}
	
	return nil
}

// Remove uma interface WireGuard
func (p *UserspaceWireguardPlatform) RemoveWireGuardInterface(interfaceName string) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	
	// Verificar se temos informação do processo em memória
	if p.wgProcess != nil {
		// Tentar encerrar com SIGTERM primeiro
		if err := p.wgProcess.Signal(os.Interrupt); err != nil {
			fmt.Printf("Aviso: não foi possível enviar sinal SIGTERM: %v\n", err)
			// Tentar matar com SIGKILL
			if err := p.wgProcess.Kill(); err != nil {
				fmt.Printf("Aviso: não foi possível matar o processo: %v\n", err)
			}
		}
		p.wgProcess = nil
	}
	
	// Tentar localizar o PID do arquivo
	pidFile := filepath.Join(p.configDir, fmt.Sprintf("%s.pid", interfaceName))
	if pidData, err := os.ReadFile(pidFile); err == nil {
		pidStr := strings.TrimSpace(string(pidData))
		// Tentar matar o processo
		killCmd := exec.Command("kill", pidStr)
		killCmd.Run() // Ignorar erros
		os.Remove(pidFile)
	}
	
	// Desativar a interface
	ipCmd := exec.Command(p.ipToolPath, "link", "set", interfaceName, "down")
	ipCmd.Run() // Ignorar erros
	
	// Remover a interface
	ipCmd = exec.Command(p.ipToolPath, "link", "delete", interfaceName)
	ipCmd.Run() // Ignorar erros
	
	// Limpar logs
	_ = os.Remove(filepath.Join(p.configDir, fmt.Sprintf("%s-stdout.log", interfaceName)))
	_ = os.Remove(filepath.Join(p.configDir, fmt.Sprintf("%s-stderr.log", interfaceName)))
	
	return nil
}

// Configura o endereço IP na interface
func (p *UserspaceWireguardPlatform) ConfigureInterfaceAddress(interfaceName, address, subnet string) error {
	ipCmd := exec.Command(p.ipToolPath, "addr", "add", fmt.Sprintf("%s/%s", address, subnet), "dev", interfaceName)
	if output, err := ipCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("erro ao configurar endereço IP (%s): %w", string(output), err)
	}
	
	// Ativar a interface
	ipCmd = exec.Command(p.ipToolPath, "link", "set", "dev", interfaceName, "up")
	if output, err := ipCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("erro ao ativar interface (%s): %w", string(output), err)
	}
	
	return nil
}

// Adiciona um peer à interface WireGuard
func (p *UserspaceWireguardPlatform) AddPeer(interfaceName, publicKeyStr, allowedIPs, endpointStr string, keepAlive int) error {
	// Construir comando wg
	args := []string{"set", interfaceName, "peer", publicKeyStr}
	
	if allowedIPs != "" {
		args = append(args, "allowed-ips", allowedIPs)
	}
	
	if endpointStr != "" {
		args = append(args, "endpoint", endpointStr)
	}
	
	if keepAlive > 0 {
		args = append(args, "persistent-keepalive", fmt.Sprintf("%d", keepAlive))
	}
	
	// Executar comando
	wgCmd := exec.Command(p.wgToolPath, args...)
	if output, err := wgCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("erro ao adicionar peer (%s): %w", string(output), err)
	}
	
	return nil
}

// Remove um peer da interface WireGuard
func (p *UserspaceWireguardPlatform) RemovePeer(interfaceName, publicKeyStr string) error {
	// Comando para remover peer
	wgCmd := exec.Command(p.wgToolPath, "set", interfaceName, "peer", publicKeyStr, "remove")
	if output, err := wgCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("erro ao remover peer (%s): %w", string(output), err)
	}
	
	return nil
}

// Configura rotas para o tráfego VPN
func (p *UserspaceWireguardPlatform) ConfigureRouting(interfaceName, vpnCIDR string) error {
	// Adicionar rota para a rede VPN
	ipCmd := exec.Command(p.ipToolPath, "route", "add", vpnCIDR, "dev", interfaceName)
	if output, err := ipCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("erro ao adicionar rota (%s): %w", string(output), err)
	}
	
	return nil
}

// Retorna o caminho para a configuração do WireGuard
func (p *UserspaceWireguardPlatform) WireGuardConfigPath(interfaceName string) string {
	return filepath.Join(p.configDir, fmt.Sprintf("%s.conf", interfaceName))
}

// Obtém o status da interface WireGuard
func (p *UserspaceWireguardPlatform) GetInterfaceStatus(interfaceName string) (bool, error) {
	// Verificar se a interface existe e está ativa
	ipCmd := exec.Command(p.ipToolPath, "link", "show", interfaceName)
	output, err := ipCmd.CombinedOutput()
	if err != nil {
		return false, nil // Interface não existe
	}
	
	// Verificar se a interface está ativa (UP)
	return strings.Contains(string(output), "UP") || strings.Contains(string(output), "UNKNOWN"), nil
}

// init registra a plataforma userspace
func init() {
	// Registrar a plataforma WireGuard Userspace como fallback
	// Esta plataforma é verificada por último, após as implementações específicas
	RegisterPlatform(func() (VPNPlatform, error) {
		platform := &UserspaceWireguardPlatform{}
		if platform.IsSupported() {
			return platform, nil
		}
		return nil, fmt.Errorf("implementação userspace do WireGuard não disponível")
	})
}

package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/p2p-vpn/p2p-vpn/core"
	"github.com/p2p-vpn/p2p-vpn/ui/desktop"
	"github.com/p2p-vpn/p2p-vpn/ui/desktop/assets"
	"github.com/p2p-vpn/p2p-vpn/ui/desktop/shared"
)

// Flags de linha de comando
var (
	configPath      = flag.String("config", "", "Caminho para o arquivo de configuração")
	minimized       = flag.Bool("minimized", false, "Iniciar minimizado na bandeja do sistema")
	language        = flag.String("lang", "pt-br", "Idioma da interface (pt-br, en, es)")
	theme           = flag.String("theme", "system", "Tema da interface (light, dark, system)")
	autostart       = flag.Bool("autostart", false, "Configurar para iniciar automaticamente com o sistema")
)

func main() {
	// Analisar flags de linha de comando
	flag.Parse()

	// Diretório base da aplicação
	appDir, err := getAppDirectory()
	if err != nil {
		log.Fatalf("Erro ao determinar diretório da aplicação: %v", err)
	}

	// Caminho padrão para configuração se não especificado
	if *configPath == "" {
		*configPath = filepath.Join(appDir, "config.json")
	}

	// Verificar se o arquivo de configuração existe
	_, err = os.Stat(*configPath)
	if os.IsNotExist(err) {
		// Criar configuração padrão
		log.Printf("Arquivo de configuração não encontrado, criando padrão em: %s", *configPath)
		createDefaultConfig(*configPath)
	}

	// Carregar configuração da VPN
	config, err := core.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("Erro ao carregar configuração: %v", err)
	}

	// Criar logger para a aplicação
	logFile, err := setupLogging(appDir)
	if err != nil {
		log.Printf("Aviso: não foi possível configurar arquivo de log: %v", err)
	} else {
		defer logFile.Close()
	}

	// Obter configuração padrão para a UI com os caminhos corretos para os ícones
	var uiConfig *shared.UIConfig = assets.GetDefaultConfig()
	
	// Sobrescrever configurações com os valores dos argumentos da linha de comando
	uiConfig.Language = *language
	uiConfig.Theme = *theme
	uiConfig.StartMinimized = *minimized
	uiConfig.AutoStart = false // Será definido mais tarde se necessário

	// Criar uma instância do core VPN
	vpnCore, err := core.NewVPNCoreMulti(config, config.ListenPort)
	if err != nil {
		log.Fatalf("Erro ao inicializar VPN Core: %v", err)
	}

	// Inicializar a UI de desktop
	log.Println("Iniciando interface gráfica...")
	uiManager, err := desktop.NewUIManager(vpnCore, uiConfig)
	if err != nil {
		log.Fatalf("Erro ao inicializar interface gráfica: %v", err)
	}

	// Configurar o início automático se solicitado pela linha de comando
	if *autostart {
		platformUI, err := uiManager.GetPlatformUI()
		if err == nil {
			err = platformUI.SetAutoStart(true)
			if err != nil {
				log.Printf("Aviso: não foi possível configurar o início automático: %v", err)
			} else {
				log.Println("Início automático configurado com sucesso")
			}
		}
	}

	// Iniciar a UI
	err = uiManager.Start()
	if err != nil {
		log.Fatalf("Erro ao executar interface gráfica: %v", err)
	}
}

// getAppDirectory retorna o diretório base da aplicação
// getAppDirectory returns the application base directory
// getAppDirectory devuelve el directorio base de la aplicación
func getAppDirectory() (string, error) {
	// Em produção, usar o diretório do executável
	exePath, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("erro ao obter caminho do executável: %v", err)
	}
	
	return filepath.Dir(exePath), nil
}

// createDefaultConfig cria um arquivo de configuração padrão
// createDefaultConfig creates a default configuration file
// createDefaultConfig crea un archivo de configuración predeterminado
func createDefaultConfig(path string) error {
	// Criar uma configuração padrão
	config := &core.Config{
		NodeID:        "",
		PrivateKey:    "",
		PublicKey:     "",
		VirtualIP:     "10.0.0.1",
		VirtualCIDR:   "10.0.0.0/24",
		InterfaceName: "wg0",
		ListenPort:    51820,
		MTU:           1420,
		DNS:           []string{"1.1.1.1", "8.8.8.8"},
		TrustedPeers:  []core.TrustedPeer{},
	}
	
	// Gerar novas chaves
	err := config.GenerateKeys()
	if err != nil {
		return fmt.Errorf("erro ao gerar chaves: %v", err)
	}
	
	// Gerar node ID baseado na chave pública
	config.GenerateNodeID()
	
	// Salvar a configuração
	return config.Save(path)
}

// setupLogging configura o arquivo de log
// setupLogging sets up the log file
// setupLogging configura el archivo de registro
func setupLogging(appDir string) (*os.File, error) {
	// Criar diretório de logs se não existir
	logDir := filepath.Join(appDir, "logs")
	err := os.MkdirAll(logDir, 0755)
	if err != nil {
		return nil, fmt.Errorf("erro ao criar diretório de logs: %v", err)
	}
	
	// Abrir arquivo de log
	logFile, err := os.OpenFile(
		filepath.Join(logDir, "desktop-ui.log"),
		os.O_CREATE|os.O_WRONLY|os.O_APPEND,
		0644,
	)
	if err != nil {
		return nil, fmt.Errorf("erro ao abrir arquivo de log: %v", err)
	}
	
	// Configurar logger para usar o arquivo
	log.SetOutput(logFile)
	
	return logFile, nil
}

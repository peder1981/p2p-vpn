// +build darwin

package platform

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/getlantern/systray"
	"howett.net/plist"

	"github.com/p2p-vpn/p2p-vpn/core"
	"github.com/p2p-vpn/p2p-vpn/ui/desktop/common"
)

// MacOSUI implementa a interface PlatformUI para macOS
// MacOSUI implements the PlatformUI interface for macOS
// MacOSUI implementa la interfaz PlatformUI para macOS
type MacOSUI struct {
	config            *common.UIConfig
	vpnCore           core.VPNProvider
	trayMenu          *systray.MenuItem
	showWindowMenu    *systray.MenuItem
	connectMenu       *systray.MenuItem
	disconnectMenu    *systray.MenuItem
	settingsMenu      *systray.MenuItem
	exitMenu          *systray.MenuItem
	connectedIcon     []byte
	disconnectedIcon  []byte
	trayInitialized   bool
	onShowWindow      func()
}

// NewMacOSUI cria uma nova instância da UI do macOS
// NewMacOSUI creates a new instance of the macOS UI
// NewMacOSUI crea una nueva instancia de la UI de macOS
func NewMacOSUI(config *common.UIConfig) (*MacOSUI, error) {
	// Carregar ícones para a bandeja
	connectedIcon, err := os.ReadFile(config.Assets.ConnectedIconPath)
	if err != nil {
		return nil, fmt.Errorf("erro ao carregar ícone conectado: %v", err)
	}

	disconnectedIcon, err := os.ReadFile(config.Assets.DisconnectedIconPath)
	if err != nil {
		return nil, fmt.Errorf("erro ao carregar ícone desconectado: %v", err)
	}

	return &MacOSUI{
		config:           config,
		connectedIcon:    connectedIcon,
		disconnectedIcon: disconnectedIcon,
		trayInitialized:  false,
	}, nil
}

// Initialize inicializa os componentes específicos do macOS
// Initialize initializes the macOS-specific components
// Initialize inicializa los componentes específicos de macOS
func (m *MacOSUI) Initialize(vpnCore core.VPNProvider, config *common.UIConfig) error {
	m.vpnCore = vpnCore
	
	// Inicializar a bandeja do sistema em uma goroutine separada
	go m.initializeTray()
	
	return nil
}

// initializeTray inicializa o ícone na bandeja do sistema
// initializeTray initializes the system tray icon
// initializeTray inicializa el icono en la bandeja del sistema
func (m *MacOSUI) initializeTray() {
	// Configurar callbacks para o ícone da bandeja
	systray.Run(m.onTrayReady, m.onTrayExit)
}

// onTrayReady é chamado quando o ícone da bandeja está pronto
// onTrayReady is called when the tray icon is ready
// onTrayReady se llama cuando el icono de la bandeja está listo
func (m *MacOSUI) onTrayReady() {
	// Definir título e ícone
	systray.SetTitle("P2P VPN")
	systray.SetTooltip("P2P VPN")
	systray.SetIcon(m.disconnectedIcon)

	// Criar itens de menu
	language := m.config.Language
	
	// Menu para mostrar/ocultar janela
	m.showWindowMenu = systray.AddMenuItem(getText(language, "showWindow"), getText(language, "showWindow"))
	
	systray.AddSeparator()
	
	// Menus para conectar/desconectar
	m.connectMenu = systray.AddMenuItem(getText(language, "connect"), getText(language, "connect"))
	m.disconnectMenu = systray.AddMenuItem(getText(language, "disconnect"), getText(language, "disconnect"))
	m.disconnectMenu.Hide() // Inicialmente escondido
	
	systray.AddSeparator()
	
	// Menu de configurações
	m.settingsMenu = systray.AddMenuItem(getText(language, "settings"), getText(language, "settings"))
	
	systray.AddSeparator()
	
	// Menu para sair
	m.exitMenu = systray.AddMenuItem(getText(language, "exit"), getText(language, "exit"))
	
	// Marcar como inicializado
	m.trayInitialized = true
	
	// Iniciar goroutine para monitorar cliques no menu
	go m.handleMenuClicks()
}

// handleMenuClicks processa os cliques nos itens do menu
// handleMenuClicks processes clicks on menu items
// handleMenuClicks procesa los clics en los elementos del menú
func (m *MacOSUI) handleMenuClicks() {
	for {
		select {
		case <-m.showWindowMenu.ClickedCh:
			// Enviar evento para mostrar a janela principal
			// Este é um placeholder, deve ser integrado com a UI principal
			log.Println("Evento: Mostrar janela principal")
			if m.onShowWindow != nil {
				m.onShowWindow()
			}
			
		case <-m.connectMenu.ClickedCh:
			// Iniciar VPN
			err := m.vpnCore.Start()
			if err != nil {
				log.Printf("Erro ao iniciar VPN: %v", err)
				m.ShowNotification("Erro", fmt.Sprintf("Falha ao conectar: %v", err), common.PriorityHigh)
			} else {
				m.UpdateTrayIcon(true)
				m.connectMenu.Hide()
				m.disconnectMenu.Show()
			}
			
		case <-m.disconnectMenu.ClickedCh:
			// Parar VPN
			err := m.vpnCore.Stop()
			if err != nil {
				log.Printf("Erro ao parar VPN: %v", err)
				m.ShowNotification("Erro", fmt.Sprintf("Falha ao desconectar: %v", err), common.PriorityHigh)
			} else {
				m.UpdateTrayIcon(false)
				m.disconnectMenu.Hide()
				m.connectMenu.Show()
			}
			
		case <-m.settingsMenu.ClickedCh:
			// Abrir configurações
			log.Println("Evento: Abrir configurações")
			
		case <-m.exitMenu.ClickedCh:
			// Sair da aplicação
			if m.vpnCore.IsRunning() {
				err := m.vpnCore.Stop()
				if err != nil {
					log.Printf("Erro ao parar VPN antes de sair: %v", err)
				}
			}
			
			// Sair da aplicação
			systray.Quit()
			os.Exit(0)
		}
	}
}

// onTrayExit é chamado quando o ícone da bandeja está sendo removido
// onTrayExit is called when the tray icon is being removed
// onTrayExit se llama cuando se está eliminando el icono de la bandeja
func (m *MacOSUI) onTrayExit() {
	// Limpar recursos se necessário
}

// Cleanup limpa recursos específicos do macOS
// Cleanup cleans up macOS-specific resources
// Cleanup limpia recursos específicos de macOS
func (m *MacOSUI) Cleanup() error {
	if m.trayInitialized {
		systray.Quit()
	}
	return nil
}

// ShowNotification exibe uma notificação no macOS
// ShowNotification displays a notification on macOS
// ShowNotification muestra una notificación en macOS
func (m *MacOSUI) ShowNotification(title, content string, priority common.NotificationPriority) {
	// Usar a ferramenta osascript para mostrar notificações no macOS
	script := fmt.Sprintf(`display notification "%s" with title "%s"`, content, title)
	
	cmd := exec.Command("osascript", "-e", script)
	err := cmd.Run()
	if err != nil {
		log.Printf("Erro ao mostrar notificação macOS: %v", err)
	}
}

// UpdateTrayIcon atualiza o ícone na bandeja do sistema
// UpdateTrayIcon updates the system tray icon
// UpdateTrayIcon actualiza el icono en la bandeja del sistema
func (m *MacOSUI) UpdateTrayIcon(connected bool) {
	if !m.trayInitialized {
		return
	}
	
	if connected {
		systray.SetIcon(m.connectedIcon)
		m.connectMenu.Hide()
		m.disconnectMenu.Show()
	} else {
		systray.SetIcon(m.disconnectedIcon)
		m.disconnectMenu.Hide()
		m.connectMenu.Show()
	}
}

// LaunchAgent representa a estrutura de um arquivo LaunchAgent do macOS
// LaunchAgent represents the structure of a macOS LaunchAgent file
// LaunchAgent representa la estructura de un archivo LaunchAgent de macOS
type LaunchAgent struct {
	Label             string   `plist:"Label"`
	ProgramArguments  []string `plist:"ProgramArguments"`
	RunAtLoad         bool     `plist:"RunAtLoad"`
	KeepAlive         bool     `plist:"KeepAlive"`
	StandardErrorPath string   `plist:"StandardErrorPath,omitempty"`
	StandardOutPath   string   `plist:"StandardOutPath,omitempty"`
}

// SetAutoStart configura o início automático com o macOS
// SetAutoStart configures the automatic start with macOS
// SetAutoStart configura el inicio automático con macOS
func (m *MacOSUI) SetAutoStart(enabled bool) error {
	// Caminho para o LaunchAgent
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("erro ao obter diretório home: %v", err)
	}
	
	launchAgentsDir := filepath.Join(homeDir, "Library", "LaunchAgents")
	plistPath := filepath.Join(launchAgentsDir, "com.p2p-vpn.app.plist")
	
	if enabled {
		// Obter caminho do executável
		exePath, err := os.Executable()
		if err != nil {
			return fmt.Errorf("erro ao obter caminho do executável: %v", err)
		}
		
		// Garantir que o diretório LaunchAgents existe
		if _, err := os.Stat(launchAgentsDir); os.IsNotExist(err) {
			err = os.MkdirAll(launchAgentsDir, 0755)
			if err != nil {
				return fmt.Errorf("erro ao criar diretório LaunchAgents: %v", err)
			}
		}
		
		// Criar arquivo LaunchAgent
		agent := LaunchAgent{
			Label:            "com.p2p-vpn.app",
			ProgramArguments: []string{exePath},
			RunAtLoad:        true,
			KeepAlive:        false,
		}
		
		// Serialize para plist
		data, err := plist.MarshalIndent(agent, plist.XMLFormat, "\t")
		if err != nil {
			return fmt.Errorf("erro ao serializar LaunchAgent: %v", err)
		}
		
		// Escrever arquivo
		err = os.WriteFile(plistPath, data, 0644)
		if err != nil {
			return fmt.Errorf("erro ao escrever arquivo LaunchAgent: %v", err)
		}
		
		// Registrar o agente
		cmd := exec.Command("launchctl", "load", plistPath)
		err = cmd.Run()
		if err != nil {
			return fmt.Errorf("erro ao registrar LaunchAgent: %v", err)
		}
	} else {
		// Verificar se o arquivo existe
		if _, err := os.Stat(plistPath); err == nil {
			// Desregistrar o agente
			cmd := exec.Command("launchctl", "unload", plistPath)
			err = cmd.Run()
			if err != nil {
				log.Printf("Aviso: erro ao desregistrar LaunchAgent: %v", err)
			}
			
			// Remover o arquivo
			err = os.Remove(plistPath)
			if err != nil {
				return fmt.Errorf("erro ao remover arquivo LaunchAgent: %v", err)
			}
		}
	}
	
	return nil
}

// Name retorna o nome da plataforma
// Name returns the platform name
// Name devuelve el nombre de la plataforma
func (m *MacOSUI) Name() string {
	return "macOS"
}

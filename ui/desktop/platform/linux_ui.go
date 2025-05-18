// +build linux

package platform

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"

	"github.com/getlantern/systray"

	"github.com/p2p-vpn/p2p-vpn/core"
	"github.com/p2p-vpn/p2p-vpn/ui/desktop/common"
)

// LinuxUI implementa a interface PlatformUI para Linux
// LinuxUI implements the PlatformUI interface for Linux
// LinuxUI implementa la interfaz PlatformUI para Linux
type LinuxUI struct {
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
	desktopEntryPath  string
	hasLibnotify      bool
}

// NewLinuxUI cria uma nova instância da UI do Linux
// NewLinuxUI creates a new instance of the Linux UI
// NewLinuxUI crea una nueva instancia de la UI de Linux
func NewLinuxUI(config *common.UIConfig) (*LinuxUI, error) {
	// Carregar ícones para a bandeja
	connectedIcon, err := os.ReadFile(config.Assets.ConnectedIconPath)
	if err != nil {
		return nil, fmt.Errorf("erro ao carregar ícone conectado: %v", err)
	}

	disconnectedIcon, err := os.ReadFile(config.Assets.DisconnectedIconPath)
	if err != nil {
		return nil, fmt.Errorf("erro ao carregar ícone desconectado: %v", err)
	}

	// Verificar se o libnotify está disponível
	_, hasLibnotify := exec.LookPath("notify-send")

	// Determinar o caminho para o arquivo .desktop
	desktopEntryPath := ""
	usr, err := user.Current()
	if err == nil {
		desktopEntryPath = filepath.Join(usr.HomeDir, ".config", "autostart", "p2p-vpn.desktop")
	}

	return &LinuxUI{
		config:           config,
		connectedIcon:    connectedIcon,
		disconnectedIcon: disconnectedIcon,
		trayInitialized:  false,
		desktopEntryPath: desktopEntryPath,
		hasLibnotify:     hasLibnotify,
	}, nil
}

// Initialize inicializa os componentes específicos do Linux
// Initialize initializes the Linux-specific components
// Initialize inicializa los componentes específicos de Linux
func (l *LinuxUI) Initialize(vpnCore core.VPNProvider, config *common.UIConfig) error {
	l.vpnCore = vpnCore
	
	// Inicializar a bandeja do sistema em uma goroutine separada
	go l.initializeTray()
	
	return nil
}

// initializeTray inicializa o ícone na bandeja do sistema
// initializeTray initializes the system tray icon
// initializeTray inicializa el icono en la bandeja del sistema
func (l *LinuxUI) initializeTray() {
	// Configurar callbacks para o ícone da bandeja
	systray.Run(l.onTrayReady, l.onTrayExit)
}

// onTrayReady é chamado quando o ícone da bandeja está pronto
// onTrayReady is called when the tray icon is ready
// onTrayReady se llama cuando el icono de la bandeja está listo
func (l *LinuxUI) onTrayReady() {
	// Definir título e ícone
	systray.SetTitle("P2P VPN")
	systray.SetTooltip("P2P VPN")
	systray.SetIcon(l.disconnectedIcon)

	// Criar itens de menu
	language := l.config.Language
	
	// Menu para mostrar/ocultar janela
	l.showWindowMenu = systray.AddMenuItem(getText(language, "showWindow"), getText(language, "showWindow"))
	
	systray.AddSeparator()
	
	// Menus para conectar/desconectar
	l.connectMenu = systray.AddMenuItem(getText(language, "connect"), getText(language, "connect"))
	l.disconnectMenu = systray.AddMenuItem(getText(language, "disconnect"), getText(language, "disconnect"))
	l.disconnectMenu.Hide() // Inicialmente escondido
	
	systray.AddSeparator()
	
	// Menu de configurações
	l.settingsMenu = systray.AddMenuItem(getText(language, "settings"), getText(language, "settings"))
	
	systray.AddSeparator()
	
	// Menu para sair
	l.exitMenu = systray.AddMenuItem(getText(language, "exit"), getText(language, "exit"))
	
	// Marcar como inicializado
	l.trayInitialized = true
	
	// Iniciar goroutine para monitorar cliques no menu
	go l.handleMenuClicks()
}

// handleMenuClicks processa os cliques nos itens do menu
// handleMenuClicks processes clicks on menu items
// handleMenuClicks procesa los clics en los elementos del menú
func (l *LinuxUI) handleMenuClicks() {
	for {
		select {
		case <-l.showWindowMenu.ClickedCh:
			// Enviar evento para mostrar a janela principal
			// Este é um placeholder, deve ser integrado com a UI principal
			log.Println("Evento: Mostrar janela principal")
			
		case <-l.connectMenu.ClickedCh:
			// Iniciar VPN
			err := l.vpnCore.Start()
			if err != nil {
				log.Printf("Erro ao iniciar VPN: %v", err)
				l.ShowNotification("Erro", fmt.Sprintf("Falha ao conectar: %v", err), common.PriorityHigh)
			} else {
				l.UpdateTrayIcon(true)
				l.connectMenu.Hide()
				l.disconnectMenu.Show()
			}
			
		case <-l.disconnectMenu.ClickedCh:
			// Parar VPN
			err := l.vpnCore.Stop()
			if err != nil {
				log.Printf("Erro ao parar VPN: %v", err)
				l.ShowNotification("Erro", fmt.Sprintf("Falha ao desconectar: %v", err), common.PriorityHigh)
			} else {
				l.UpdateTrayIcon(false)
				l.disconnectMenu.Hide()
				l.connectMenu.Show()
			}
			
		case <-l.settingsMenu.ClickedCh:
			// Abrir configurações
			log.Println("Evento: Abrir configurações")
			
		case <-l.exitMenu.ClickedCh:
			// Sair da aplicação
			if l.vpnCore.IsRunning() {
				err := l.vpnCore.Stop()
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
func (l *LinuxUI) onTrayExit() {
	// Limpar recursos se necessário
}

// Cleanup limpa recursos específicos do Linux
// Cleanup cleans up Linux-specific resources
// Cleanup limpia recursos específicos de Linux
func (l *LinuxUI) Cleanup() error {
	if l.trayInitialized {
		systray.Quit()
	}
	return nil
}

// ShowNotification exibe uma notificação no Linux
// ShowNotification displays a notification on Linux
// ShowNotification muestra una notificación en Linux
func (l *LinuxUI) ShowNotification(title, content string, priority common.NotificationPriority) {
	// Determinar urgência baseado na prioridade
	urgency := "normal"
	switch priority {
	case common.PriorityLow:
		urgency = "low"
	case common.PriorityNormal:
		urgency = "normal"
	case common.PriorityHigh:
		urgency = "critical"
	}
	
	if l.hasLibnotify {
		// Usar notify-send se disponível
		cmd := exec.Command("notify-send", "-u", urgency, title, content)
		err := cmd.Run()
		if err != nil {
			log.Printf("Erro ao mostrar notificação: %v", err)
		}
	} else {
		// Fallback para log se notify-send não estiver disponível
		log.Printf("Notificação [%s]: %s - %s", urgency, title, content)
	}
}

// UpdateTrayIcon atualiza o ícone na bandeja do sistema
// UpdateTrayIcon updates the system tray icon
// UpdateTrayIcon actualiza el icono en la bandeja del sistema
func (l *LinuxUI) UpdateTrayIcon(connected bool) {
	if !l.trayInitialized {
		return
	}
	
	if connected {
		systray.SetIcon(l.connectedIcon)
		l.connectMenu.Hide()
		l.disconnectMenu.Show()
	} else {
		systray.SetIcon(l.disconnectedIcon)
		l.disconnectMenu.Hide()
		l.connectMenu.Show()
	}
}

// SetAutoStart configura o início automático com o Linux (através de arquivo .desktop)
// SetAutoStart configures the automatic start with Linux (via .desktop file)
// SetAutoStart configura el inicio automático con Linux (mediante archivo .desktop)
func (l *LinuxUI) SetAutoStart(enabled bool) error {
	if l.desktopEntryPath == "" {
		return fmt.Errorf("não foi possível determinar o caminho para o arquivo de autostart")
	}
	
	if enabled {
		// Criar diretório autostart se não existir
		autostartDir := filepath.Dir(l.desktopEntryPath)
		if _, err := os.Stat(autostartDir); os.IsNotExist(err) {
			err := os.MkdirAll(autostartDir, 0755)
			if err != nil {
				return fmt.Errorf("erro ao criar diretório autostart: %v", err)
			}
		}
		
		// Obter caminho do executável
		exePath, err := os.Executable()
		if err != nil {
			return fmt.Errorf("erro ao obter caminho do executável: %v", err)
		}
		
		// Criar arquivo .desktop
		desktopContent := fmt.Sprintf(`[Desktop Entry]
Type=Application
Name=P2P VPN
Exec=%s
Icon=network-vpn
Comment=P2P VPN Client
Categories=Network;
Terminal=false
StartupNotify=false
Hidden=false
X-GNOME-Autostart-enabled=true
`, exePath)
		
		err = os.WriteFile(l.desktopEntryPath, []byte(desktopContent), 0644)
		if err != nil {
			return fmt.Errorf("erro ao escrever arquivo .desktop: %v", err)
		}
	} else {
		// Remover arquivo .desktop se existir
		if _, err := os.Stat(l.desktopEntryPath); err == nil {
			err := os.Remove(l.desktopEntryPath)
			if err != nil {
				return fmt.Errorf("erro ao remover arquivo .desktop: %v", err)
			}
		}
	}
	
	return nil
}

// Name retorna o nome da plataforma
// Name returns the platform name
// Name devuelve el nombre de la plataforma
func (l *LinuxUI) Name() string {
	return "Linux"
}

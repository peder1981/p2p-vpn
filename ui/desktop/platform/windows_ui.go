// +build windows

package platform

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"syscall"

	"github.com/getlantern/systray"
	"github.com/lxn/win"
	"golang.org/x/sys/windows/registry"

	"github.com/p2p-vpn/p2p-vpn/core"
	"github.com/p2p-vpn/p2p-vpn/ui/desktop/common"
)

// WindowsUI implementa a interface PlatformUI para Windows
// WindowsUI implements the PlatformUI interface for Windows
// WindowsUI implementa la interfaz PlatformUI para Windows
type WindowsUI struct {
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
}

// NewWindowsUI cria uma nova instância da UI do Windows
// NewWindowsUI creates a new instance of the Windows UI
// NewWindowsUI crea una nueva instancia de la UI de Windows
func NewWindowsUI(config *common.UIConfig) (*WindowsUI, error) {
	// Carregar ícones para a bandeja
	connectedIcon, err := os.ReadFile(config.Assets.ConnectedIconPath)
	if err != nil {
		return nil, fmt.Errorf("erro ao carregar ícone conectado: %v", err)
	}

	disconnectedIcon, err := os.ReadFile(config.Assets.DisconnectedIconPath)
	if err != nil {
		return nil, fmt.Errorf("erro ao carregar ícone desconectado: %v", err)
	}

	return &WindowsUI{
		config:           config,
		connectedIcon:    connectedIcon,
		disconnectedIcon: disconnectedIcon,
		trayInitialized:  false,
	}, nil
}

// Initialize inicializa os componentes específicos do Windows
// Initialize initializes the Windows-specific components
// Initialize inicializa los componentes específicos de Windows
func (w *WindowsUI) Initialize(vpnCore core.VPNProvider, config *common.UIConfig) error {
	w.vpnCore = vpnCore
	
	// Inicializar a bandeja do sistema em uma goroutine separada
	go w.initializeTray()
	
	return nil
}

// initializeTray inicializa o ícone na bandeja do sistema
// initializeTray initializes the system tray icon
// initializeTray inicializa el icono en la bandeja del sistema
func (w *WindowsUI) initializeTray() {
	// Configurar callbacks para o ícone da bandeja
	systray.Run(w.onTrayReady, w.onTrayExit)
}

// onTrayReady é chamado quando o ícone da bandeja está pronto
// onTrayReady is called when the tray icon is ready
// onTrayReady se llama cuando el icono de la bandeja está listo
func (w *WindowsUI) onTrayReady() {
	// Definir título e ícone
	systray.SetTitle("P2P VPN")
	systray.SetTooltip("P2P VPN")
	systray.SetIcon(w.disconnectedIcon)

	// Criar itens de menu
	language := w.config.Language
	
	// Menu para mostrar/ocultar janela
	w.showWindowMenu = systray.AddMenuItem(getText(language, "showWindow"), getText(language, "showWindow"))
	
	systray.AddSeparator()
	
	// Menus para conectar/desconectar
	w.connectMenu = systray.AddMenuItem(getText(language, "connect"), getText(language, "connect"))
	w.disconnectMenu = systray.AddMenuItem(getText(language, "disconnect"), getText(language, "disconnect"))
	w.disconnectMenu.Hide() // Inicialmente escondido
	
	systray.AddSeparator()
	
	// Menu de configurações
	w.settingsMenu = systray.AddMenuItem(getText(language, "settings"), getText(language, "settings"))
	
	systray.AddSeparator()
	
	// Menu para sair
	w.exitMenu = systray.AddMenuItem(getText(language, "exit"), getText(language, "exit"))
	
	// Marcar como inicializado
	w.trayInitialized = true
	
	// Iniciar goroutine para monitorar cliques no menu
	go w.handleMenuClicks()
}

// handleMenuClicks processa os cliques nos itens do menu
// handleMenuClicks processes clicks on menu items
// handleMenuClicks procesa los clics en los elementos del menú
func (w *WindowsUI) handleMenuClicks() {
	for {
		select {
		case <-w.showWindowMenu.ClickedCh:
			// Enviar evento para mostrar a janela principal
			// Este é um placeholder, deve ser integrado com a UI principal
			log.Println("Evento: Mostrar janela principal")
			
		case <-w.connectMenu.ClickedCh:
			// Iniciar VPN
			err := w.vpnCore.Start()
			if err != nil {
				log.Printf("Erro ao iniciar VPN: %v", err)
				w.ShowNotification("Erro", fmt.Sprintf("Falha ao conectar: %v", err), common.PriorityHigh)
			} else {
				w.UpdateTrayIcon(true)
				w.connectMenu.Hide()
				w.disconnectMenu.Show()
			}
			
		case <-w.disconnectMenu.ClickedCh:
			// Parar VPN
			err := w.vpnCore.Stop()
			if err != nil {
				log.Printf("Erro ao parar VPN: %v", err)
				w.ShowNotification("Erro", fmt.Sprintf("Falha ao desconectar: %v", err), common.PriorityHigh)
			} else {
				w.UpdateTrayIcon(false)
				w.disconnectMenu.Hide()
				w.connectMenu.Show()
			}
			
		case <-w.settingsMenu.ClickedCh:
			// Abrir configurações
			log.Println("Evento: Abrir configurações")
			
		case <-w.exitMenu.ClickedCh:
			// Sair da aplicação
			if w.vpnCore.IsRunning() {
				err := w.vpnCore.Stop()
				if err != nil {
					log.Printf("Erro ao parar VPN antes de sair: %v", err)
				}
			}
			
			// Sair da aplicação - o método real dependerá da integração com a UI principal
			systray.Quit()
			os.Exit(0)
		}
	}
}

// onTrayExit é chamado quando o ícone da bandeja está sendo removido
// onTrayExit is called when the tray icon is being removed
// onTrayExit se llama cuando se está eliminando el icono de la bandeja
func (w *WindowsUI) onTrayExit() {
	// Limpar recursos se necessário
}

// Cleanup limpa recursos específicos do Windows
// Cleanup cleans up Windows-specific resources
// Cleanup limpia recursos específicos de Windows
func (w *WindowsUI) Cleanup() error {
	if w.trayInitialized {
		systray.Quit()
	}
	return nil
}

// ShowNotification exibe uma notificação no Windows
// ShowNotification displays a notification on Windows
// ShowNotification muestra una notificación en Windows
func (w *WindowsUI) ShowNotification(title, content string, priority common.NotificationPriority) {
	// Usar a API do Windows para notificações
	var flags uint32 = win.NIIF_INFO
	
	switch priority {
	case common.PriorityLow:
		flags = win.NIIF_INFO
	case common.PriorityNormal:
		flags = win.NIIF_INFO
	case common.PriorityHigh:
		flags = win.NIIF_WARNING
	}
	
	// Mostrar balão com a notificação (implementação simplificada)
	log.Printf("Notificação Windows [%s]: %s", title, content)
}

// UpdateTrayIcon atualiza o ícone na bandeja do sistema
// UpdateTrayIcon updates the system tray icon
// UpdateTrayIcon actualiza el icono en la bandeja del sistema
func (w *WindowsUI) UpdateTrayIcon(connected bool) {
	if !w.trayInitialized {
		return
	}
	
	if connected {
		systray.SetIcon(w.connectedIcon)
		w.connectMenu.Hide()
		w.disconnectMenu.Show()
	} else {
		systray.SetIcon(w.disconnectedIcon)
		w.disconnectMenu.Hide()
		w.connectMenu.Show()
	}
}

// SetAutoStart configura o início automático com o Windows
// SetAutoStart configures the automatic start with Windows
// SetAutoStart configura el inicio automático con Windows
func (w *WindowsUI) SetAutoStart(enabled bool) error {
	const regPath = `SOFTWARE\Microsoft\Windows\CurrentVersion\Run`
	
	// Abrir ou criar chave de registro
	key, _, err := registry.CreateKey(
		registry.CURRENT_USER, 
		regPath, 
		registry.ALL_ACCESS,
	)
	if err != nil {
		return fmt.Errorf("erro ao acessar registro para autostart: %v", err)
	}
	defer key.Close()
	
	appName := "P2P-VPN"
	
	if enabled {
		// Obter caminho do executável
		exePath, err := os.Executable()
		if err != nil {
			return fmt.Errorf("erro ao obter caminho do executável: %v", err)
		}
		
		// Adicionar ao registro
		err = key.SetStringValue(appName, exePath)
		if err != nil {
			return fmt.Errorf("erro ao configurar autostart: %v", err)
		}
	} else {
		// Remover do registro
		err = key.DeleteValue(appName)
		if err != nil && err != registry.ErrNotExist {
			return fmt.Errorf("erro ao remover autostart: %v", err)
		}
	}
	
	return nil
}

// Name retorna o nome da plataforma
// Name returns the platform name
// Name devuelve el nombre de la plataforma
func (w *WindowsUI) Name() string {
	return "Windows"
}

// Função auxiliar para tradução
// Helper function for translation
// Función auxiliar para traducción
func getText(language, key string) string {
	translations := map[string]map[string]string{
		"pt-br": {
			"showWindow": "Mostrar Janela",
			"connect":    "Conectar",
			"disconnect": "Desconectar",
			"settings":   "Configurações",
			"exit":       "Sair",
		},
		"en": {
			"showWindow": "Show Window",
			"connect":    "Connect",
			"disconnect": "Disconnect",
			"settings":   "Settings",
			"exit":       "Exit",
		},
		"es": {
			"showWindow": "Mostrar Ventana",
			"connect":    "Conectar",
			"disconnect": "Desconectar",
			"settings":   "Ajustes",
			"exit":       "Salir",
		},
	}
	
	if trans, ok := translations[language]; ok {
		if text, ok := trans[key]; ok {
			return text
		}
	}
	
	// Fallback para inglês
	if englishTrans, ok := translations["en"]; ok {
		if text, ok := englishTrans[key]; ok {
			return text
		}
	}
	
	// Se não encontrar, retornar a chave
	return key
}

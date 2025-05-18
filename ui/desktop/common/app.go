package common

import (
	"fmt"
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/data/binding"

	"github.com/p2p-vpn/p2p-vpn/core"
	"github.com/p2p-vpn/p2p-vpn/ui/desktop/platform"
)

// DesktopApp implementa a interface DesktopUI usando a biblioteca Fyne
// DesktopApp implements the DesktopUI interface using the Fyne library
// DesktopApp implementa la interfaz DesktopUI usando la biblioteca Fyne
type DesktopApp struct {
	fyneApp        fyne.App
	mainWindow     fyne.Window
	statusLabel    *widget.Label
	peerList       *widget.List
	peerData       binding.ExternalStringList
	connectButton  *widget.Button
	platformUI     platform.PlatformUI
	vpnCore        core.VPNProvider
	config         *UIConfig
	vpnRunning     bool
	peers          []core.TrustedPeer
}

// Tradução simples para múltiplos idiomas
// Simple translation for multiple languages
// Traducción simple para múltiples idiomas
var translations = map[string]map[string]string{
	"pt-br": {
		"title":       "P2P VPN",
		"status":      "Status: %s",
		"connect":     "Conectar",
		"disconnect":  "Desconectar",
		"peers":       "Peers",
		"settings":    "Configurações",
		"exit":        "Sair",
		"connected":   "Conectado",
		"disconnected": "Desconectado",
	},
	"en": {
		"title":       "P2P VPN",
		"status":      "Status: %s",
		"connect":     "Connect",
		"disconnect":  "Disconnect",
		"peers":       "Peers",
		"settings":    "Settings",
		"exit":        "Exit",
		"connected":   "Connected",
		"disconnected": "Disconnected",
	},
	"es": {
		"title":       "P2P VPN",
		"status":      "Estado: %s",
		"connect":     "Conectar",
		"disconnect":  "Desconectar",
		"peers":       "Peers",
		"settings":    "Ajustes",
		"exit":        "Salir",
		"connected":   "Conectado",
		"disconnected": "Desconectado",
	},
}

// NewDesktopApp cria uma nova instância da aplicação desktop
// NewDesktopApp creates a new instance of the desktop application
// NewDesktopApp crea una nueva instancia de la aplicación de escritorio
func NewDesktopApp(vpnCore core.VPNProvider, platformUI platform.PlatformUI, config *UIConfig) (*DesktopApp, error) {
	// Criar nova aplicação Fyne
	fyneApp := app.New()
	
	// Definir tema baseado na configuração
	if config.Theme == "dark" {
		fyneApp.Settings().SetTheme(theme.DarkTheme())
	} else if config.Theme == "light" {
		fyneApp.Settings().SetTheme(theme.LightTheme())
	} // default: tema do sistema

	// Criar janela principal
	mainWindow := fyneApp.NewWindow(getText(config.Language, "title"))
	
	// Criar binding para lista de peers
	peerData := binding.NewStringList()
	
	// Inicializar a estrutura DesktopApp
	desktopApp := &DesktopApp{
		fyneApp:    fyneApp,
		mainWindow: mainWindow,
		platformUI: platformUI,
		vpnCore:    vpnCore,
		config:     config,
		peerData:   peerData,
		vpnRunning: false,
	}
	
	// Configurar a interface
	desktopApp.setupUI()
	
	// Inicializar estado
	desktopApp.updatePeerList()
	desktopApp.UpdateStatus(core.VPNStatus{Running: false})

	return desktopApp, nil
}

// setupUI configura os elementos da interface
// setupUI sets up the UI elements
// setupUI configura los elementos de la interfaz
func (d *DesktopApp) setupUI() {
	// Status da VPN
	d.statusLabel = widget.NewLabel(fmt.Sprintf(getText(d.config.Language, "status"), 
		getText(d.config.Language, "disconnected")))
	
	// Botão de conectar/desconectar
	d.connectButton = widget.NewButton(getText(d.config.Language, "connect"), func() {
		d.toggleVPNConnection()
	})
	
	// Lista de peers
	d.peerList = widget.NewListWithData(
		d.peerData,
		func() fyne.CanvasObject {
			return widget.NewLabel("Template Item")
		},
		func(i binding.DataItem, o fyne.CanvasObject) {
			o.(*widget.Label).Bind(i.(binding.String))
		},
	)
	
	// Layout principal
	content := container.NewVBox(
		d.statusLabel,
		d.connectButton,
		widget.NewLabel(getText(d.config.Language, "peers")),
		container.New(layout.NewGridWrapLayout(fyne.NewSize(300, 200)), d.peerList),
	)
	
	// Configurar menu
	d.mainWindow.SetMainMenu(fyne.NewMainMenu(
		fyne.NewMenu("File",
			fyne.NewMenuItem(getText(d.config.Language, "settings"), func() {
				d.showSettings()
			}),
			fyne.NewMenuItem(getText(d.config.Language, "exit"), func() {
				d.Close()
			}),
		),
	))
	
	// Configurar a janela principal
	d.mainWindow.SetContent(content)
	d.mainWindow.Resize(fyne.NewSize(400, 500))
	d.mainWindow.SetCloseIntercept(func() {
		// Minimizar para a bandeja em vez de fechar
		d.mainWindow.Hide()
	})
}

// toggleVPNConnection alterna o estado da conexão VPN
// toggleVPNConnection toggles the VPN connection state
// toggleVPNConnection cambia el estado de la conexión VPN
func (d *DesktopApp) toggleVPNConnection() {
	if d.vpnRunning {
		// Desconectar
		err := d.vpnCore.Stop()
		if err != nil {
			// Mostrar erro
			log.Printf("Erro ao desconectar VPN: %v", err)
			d.ShowNotification("Erro", fmt.Sprintf("Falha ao desconectar: %v", err), PriorityHigh)
			return
		}
		
		d.UpdateStatus(core.VPNStatus{Running: false})
	} else {
		// Conectar
		err := d.vpnCore.Start()
		if err != nil {
			// Mostrar erro
			log.Printf("Erro ao conectar VPN: %v", err)
			d.ShowNotification("Erro", fmt.Sprintf("Falha ao conectar: %v", err), PriorityHigh)
			return
		}
		
		d.UpdateStatus(core.VPNStatus{Running: true})
	}
}

// updatePeerList atualiza a lista de peers na interface
// updatePeerList updates the peer list in the interface
// updatePeerList actualiza la lista de peers en la interfaz
func (d *DesktopApp) updatePeerList() {
	// Obter lista de peers do VPNCore
	config := d.vpnCore.GetConfig()
	if config == nil {
		return
	}
	
	d.peers = config.TrustedPeers
	
	// Criar strings para exibição
	peerStrings := make([]string, len(d.peers))
	for i, peer := range d.peers {
		peerStrings[i] = fmt.Sprintf("%s (%s)", peer.NodeID, peer.VirtualIP)
	}
	
	// Atualizar binding
	d.peerData.Set(peerStrings)
}

// showSettings exibe a janela de configurações
// showSettings displays the settings window
// showSettings muestra la ventana de configuración
func (d *DesktopApp) showSettings() {
	// Função a ser implementada - mostrará uma janela de configurações
	// Will be implemented - will show a settings window
	// Se implementará - mostrará una ventana de configuración
	log.Println("Configurações não implementadas ainda")
}

// Run inicia o loop principal da interface gráfica
// Run starts the main GUI loop
// Run inicia el bucle principal de la interfaz gráfica
func (d *DesktopApp) Run() error {
	if d.config.StartMinimized {
		// Iniciar minimizado na bandeja
		d.mainWindow.Hide()
	} else {
		// Mostrar a janela principal
		d.mainWindow.Show()
	}
	
	// Iniciar o loop principal do app
	d.fyneApp.Run()
	
	return nil
}

// Close fecha a interface gráfica
// Close closes the GUI
// Close cierra la interfaz gráfica
func (d *DesktopApp) Close() error {
	d.mainWindow.Close()
	return nil
}

// ShowMainWindow exibe/oculta a janela principal
// ShowMainWindow shows/hides the main window
// ShowMainWindow muestra/oculta la ventana principal
func (d *DesktopApp) ShowMainWindow(show bool) {
	if show {
		d.mainWindow.Show()
		d.mainWindow.RequestFocus()
	} else {
		d.mainWindow.Hide()
	}
}

// ShowNotification exibe uma notificação
// ShowNotification displays a notification
// ShowNotification muestra una notificación
func (d *DesktopApp) ShowNotification(title, content string, priority NotificationPriority) {
	// Delegamos para a implementação específica da plataforma
	d.platformUI.ShowNotification(title, content, priority)
}

// UpdateStatus atualiza o status da VPN na interface
// UpdateStatus updates the VPN status in the interface
// UpdateStatus actualiza el estado de la VPN en la interfaz
func (d *DesktopApp) UpdateStatus(status core.VPNStatus) {
	d.vpnRunning = status.Running
	
	// Atualizar texto do botão e status
	if status.Running {
		d.statusLabel.SetText(fmt.Sprintf(getText(d.config.Language, "status"),
			getText(d.config.Language, "connected")))
		d.connectButton.SetText(getText(d.config.Language, "disconnect"))
	} else {
		d.statusLabel.SetText(fmt.Sprintf(getText(d.config.Language, "status"),
			getText(d.config.Language, "disconnected")))
		d.connectButton.SetText(getText(d.config.Language, "connect"))
	}
	
	// Atualizar ícone na bandeja
	d.platformUI.UpdateTrayIcon(status.Running)
}

// getText obtém um texto traduzido com base no idioma configurado
// getText gets a translated text based on the configured language
// getText obtiene un texto traducido basado en el idioma configurado
func getText(language, key string) string {
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

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
	"github.com/p2p-vpn/p2p-vpn/ui/desktop/shared"
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
	config         *shared.UIConfig
	vpnRunning     bool
	peers          []core.TrustedPeer
	translations   shared.Translations
}

// NewDesktopApp cria uma nova instância da aplicação desktop
// NewDesktopApp creates a new instance of the desktop application
// NewDesktopApp crea una nueva instancia de la aplicación de escritorio
func NewDesktopApp(vpnCore core.VPNProvider, platformUI platform.PlatformUI, config *shared.UIConfig) (*DesktopApp, error) {
	// Carregar traduções para o idioma configurado
	translations := shared.GetTranslations(config.Language)
	// Criar nova aplicação Fyne
	fyneApp := app.New()
	
	// Definir tema baseado na configuração
	if config.Theme == "dark" {
		fyneApp.Settings().SetTheme(theme.DarkTheme())
	} else if config.Theme == "light" {
		fyneApp.Settings().SetTheme(theme.LightTheme())
	} // default: tema do sistema

	// Criar janela principal
	mainWindow := fyneApp.NewWindow("P2P VPN")
	
	// Criar binding para lista de peers
	peerData := binding.NewStringList()
	
	// Inicializar a estrutura DesktopApp
	desktopApp := &DesktopApp{
		fyneApp:        fyneApp,
		mainWindow:     mainWindow,
		platformUI:     platformUI,
		vpnCore:        vpnCore,
		config:         config,
		peerData:       peerData,
		vpnRunning:     false,
		translations:   translations,
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
	d.statusLabel = widget.NewLabel(fmt.Sprintf("%s: %s", d.GetTranslated("common", shared.KeyStatus),
		d.GetTranslated("status", shared.KeyStatusDisconnected)))
	
	// Botão de conectar/desconectar
	d.connectButton = widget.NewButton(d.GetTranslated("common", shared.KeyConnect), func() {
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
		widget.NewLabel("Peers"),
		container.New(layout.NewGridWrapLayout(fyne.NewSize(300, 200)), d.peerList),
	)
	
	// Configurar menu
	d.mainWindow.SetMainMenu(fyne.NewMainMenu(
		fyne.NewMenu("File",
			fyne.NewMenuItem(d.GetTranslated("common", shared.KeySettings), func() {
				d.showSettings()
			}),
			fyne.NewMenuItem(d.GetTranslated("common", shared.KeyExit), func() {
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
			d.ShowNotification("Erro", fmt.Sprintf("Falha ao desconectar: %v", err), shared.PriorityHigh)
			return
		}
		
		d.UpdateStatus(core.VPNStatus{Running: false})
	} else {
		// Conectar
		err := d.vpnCore.Start()
		if err != nil {
			// Mostrar erro
			log.Printf("Erro ao conectar VPN: %v", err)
			d.ShowNotification("Erro", fmt.Sprintf("Falha ao conectar: %v", err), shared.PriorityHigh)
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
func (d *DesktopApp) ShowNotification(title, content string, priority shared.NotificationPriority) {
	d.platformUI.ShowNotification(title, content, priority)
}

// UpdateStatus atualiza o status da VPN na interface
// UpdateStatus updates the VPN status in the interface
// UpdateStatus actualiza el estado de la VPN en la interfaz
func (d *DesktopApp) UpdateStatus(status core.VPNStatus) {
	d.vpnRunning = status.Running
	
	// Atualizar texto do botão e status
	if status.Running {
		d.statusLabel.SetText(fmt.Sprintf("%s: %s", d.GetTranslated("common", shared.KeyStatus),
			d.GetTranslated("status", shared.KeyStatusConnected)))
		d.connectButton.SetText(d.GetTranslated("common", shared.KeyDisconnect))
	} else {
		d.statusLabel.SetText(fmt.Sprintf("%s: %s", d.GetTranslated("common", shared.KeyStatus),
			d.GetTranslated("status", shared.KeyStatusDisconnected)))
		d.connectButton.SetText(d.GetTranslated("common", shared.KeyConnect))
	}
	
	// Atualizar ícone na bandeja
	d.platformUI.UpdateTrayIcon(status.Running)
}

// GetTranslated retorna uma string traduzida com base na chave e categoria
// GetTranslated returns a translated string based on key and category
// GetTranslated devuelve una cadena traducida basada en la clave y categoría
func (d *DesktopApp) GetTranslated(category string, key string) string {
	return shared.GetTranslated(d.translations, category, key)
}

// FormatTranslated retorna uma string traduzida formatada com argumentos
// FormatTranslated returns a translated string formatted with arguments
// FormatTranslated devuelve una cadena traducida formateada con argumentos
func (d *DesktopApp) FormatTranslated(category string, key string, a ...interface{}) string {
	base := shared.GetTranslated(d.translations, category, key)
	return fmt.Sprintf(base, a...)
}

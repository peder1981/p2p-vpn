package common

import (
	"github.com/p2p-vpn/p2p-vpn/core"
)

// DesktopUI define a interface para a UI comum entre todas as plataformas
// DesktopUI defines the interface for common UI across all platforms
// DesktopUI define la interfaz para la UI común entre todas las plataformas
type DesktopUI interface {
	// Run inicia o loop principal da interface gráfica
	// Run starts the main GUI loop
	// Run inicia el bucle principal de la interfaz gráfica
	Run() error

	// Close fecha a interface gráfica
	// Close closes the GUI
	// Close cierra la interfaz gráfica
	Close() error

	// ShowMainWindow exibe/oculta a janela principal
	// ShowMainWindow shows/hides the main window
	// ShowMainWindow muestra/oculta la ventana principal
	ShowMainWindow(show bool)

	// ShowNotification exibe uma notificação
	// ShowNotification displays a notification
	// ShowNotification muestra una notificación
	ShowNotification(title, content string, priority NotificationPriority)

	// UpdateStatus atualiza o status da VPN na interface
	// UpdateStatus updates the VPN status in the interface
	// UpdateStatus actualiza el estado de la VPN en la interfaz
	UpdateStatus(status core.VPNStatus)
}

// NotificationPriority define a prioridade de uma notificação
// NotificationPriority defines the priority of a notification
// NotificationPriority define la prioridad de una notificación
type NotificationPriority int

const (
	// PriorityLow é usado para notificações informativas
	// PriorityLow is used for informational notifications
	// PriorityLow se usa para notificaciones informativas
	PriorityLow NotificationPriority = iota

	// PriorityNormal é usado para notificações normais
	// PriorityNormal is used for normal notifications
	// PriorityNormal se usa para notificaciones normales
	PriorityNormal

	// PriorityHigh é usado para alertas importantes
	// PriorityHigh is used for important alerts
	// PriorityHigh se usa para alertas importantes
	PriorityHigh
)

// UIConfig contém configurações para a interface gráfica
// UIConfig contains configuration for the GUI
// UIConfig contiene configuración para la interfaz gráfica
type UIConfig struct {
	// Language define o idioma da interface (pt-br, en, es)
	// Language defines the interface language (pt-br, en, es)
	// Language define el idioma de la interfaz (pt-br, en, es)
	Language string

	// Theme define o tema da interface (light, dark, system)
	// Theme defines the interface theme (light, dark, system)
	// Theme define el tema de la interfaz (light, dark, system)
	Theme string

	// StartMinimized define se a aplicação deve iniciar minimizada
	// StartMinimized defines if the application should start minimized
	// StartMinimized define si la aplicación debe iniciarse minimizada
	StartMinimized bool

	// AutoStart define se a aplicação deve iniciar automaticamente com o sistema
	// AutoStart defines if the application should start automatically with the system
	// AutoStart define si la aplicación debe iniciarse automáticamente con el sistema
	AutoStart bool

	// Assets contém caminhos para os recursos de UI
	// Assets contains paths to UI resources
	// Assets contiene rutas a los recursos de UI
	Assets UIAssets
}

// UIAssets contém caminhos para os recursos da interface
// UIAssets contains paths to interface resources
// UIAssets contiene rutas a los recursos de la interfaz
type UIAssets struct {
	// IconPath caminho para o ícone da aplicação
	// IconPath path to the application icon
	// IconPath ruta al icono de la aplicación
	IconPath string

	// TrayIconPath caminho para o ícone da bandeja do sistema
	// TrayIconPath path to the system tray icon
	// TrayIconPath ruta al icono de la bandeja del sistema
	TrayIconPath string

	// ConnectedIconPath caminho para o ícone de status conectado
	// ConnectedIconPath path to the connected status icon
	// ConnectedIconPath ruta al icono de estado conectado
	ConnectedIconPath string

	// DisconnectedIconPath caminho para o ícone de status desconectado
	// DisconnectedIconPath path to the disconnected status icon
	// DisconnectedIconPath ruta al icono de estado desconectado
	DisconnectedIconPath string
}

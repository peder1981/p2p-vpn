package shared

import (
	"github.com/p2p-vpn/p2p-vpn/core"
)

// PlatformUI é a interface para implementações específicas de plataforma da UI
// PlatformUI is the interface for platform-specific UI implementations
// PlatformUI es la interfaz para implementaciones de UI específicas de plataforma
type PlatformUI interface {
	// Initialize inicializa os componentes específicos da plataforma
	// Initialize initializes platform-specific components
	// Initialize inicializa los componentes específicos de la plataforma
	Initialize(vpnCore core.VPNProvider, config *UIConfig) error
	
	// Cleanup limpa recursos específicos da plataforma
	// Cleanup cleans up platform-specific resources
	// Cleanup limpia recursos específicos de la plataforma
	Cleanup() error
	
	// ShowNotification exibe uma notificação
	// ShowNotification displays a notification
	// ShowNotification muestra una notificación
	ShowNotification(title, content string, priority NotificationPriority)
	
	// UpdateTrayIcon atualiza o ícone na bandeja do sistema
	// UpdateTrayIcon updates the system tray icon
	// UpdateTrayIcon actualiza el icono en la bandeja del sistema
	UpdateTrayIcon(connected bool)
	
	// SetAutoStart configura o início automático
	// SetAutoStart configures automatic start
	// SetAutoStart configura el inicio automático
	SetAutoStart(enabled bool) error
	
	// Name retorna o nome da plataforma
	// Name returns the platform name
	// Name devuelve el nombre de la plataforma
	Name() string
}

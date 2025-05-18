package platform

import (
	"fmt"
	"runtime"

	"github.com/p2p-vpn/p2p-vpn/core"
	"github.com/p2p-vpn/p2p-vpn/ui/desktop/shared"
)

// PlatformUI define a interface para funcionalidades específicas de cada plataforma
// PlatformUI defines the interface for platform-specific functionalities
// PlatformUI define la interfaz para funcionalidades específicas de cada plataforma
type PlatformUI interface {
	// Initialize inicializa os componentes específicos da plataforma
	// Initialize initializes the platform-specific components
	// Initialize inicializa los componentes específicos de la plataforma
	Initialize(vpnCore core.VPNProvider, config *shared.UIConfig) error

	// Cleanup limpa recursos específicos da plataforma
	// Cleanup cleans up platform-specific resources
	// Cleanup limpia recursos específicos de la plataforma
	Cleanup() error

	// ShowNotification exibe uma notificação específica da plataforma
	// ShowNotification displays a platform-specific notification
	// ShowNotification muestra una notificación específica de la plataforma
	ShowNotification(title, content string, priority shared.NotificationPriority)

	// UpdateTrayIcon atualiza o ícone na bandeja do sistema
	// UpdateTrayIcon updates the system tray icon
	// UpdateTrayIcon actualiza el icono en la bandeja del sistema
	UpdateTrayIcon(connected bool)

	// SetAutoStart configura o início automático com o sistema
	// SetAutoStart configures the automatic start with the system
	// SetAutoStart configura el inicio automático con el sistema
	SetAutoStart(enabled bool) error

	// Name retorna o nome da plataforma
	// Name returns the platform name
	// Name devuelve el nombre de la plataforma
	Name() string
}

// GetPlatformUI retorna a implementação da UI específica para a plataforma atual
// GetPlatformUI returns the UI implementation specific to the current platform
// GetPlatformUI devuelve la implementación de UI específica para la plataforma actual
func GetPlatformUI(config *shared.UIConfig) (PlatformUI, error) {
	switch runtime.GOOS {
	case "windows":
		return NewWindowsUI(config)
	case "darwin":
		return NewMacOSUI(config)
	case "linux":
		return NewLinuxUI(config)
	default:
		return nil, fmt.Errorf("plataforma não suportada: %s", runtime.GOOS)
	}
}

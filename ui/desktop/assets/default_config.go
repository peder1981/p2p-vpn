package assets

import (
	"path/filepath"
	"runtime"
	
	"github.com/p2p-vpn/p2p-vpn/ui/desktop/shared"
)

// GetDefaultConfig retorna a configuração padrão da UI com os caminhos corretos para os ícones
// GetDefaultConfig returns the default UI configuration with the correct paths for icons
// GetDefaultConfig devuelve la configuración predeterminada de la UI con las rutas correctas para los iconos
func GetDefaultConfig() *shared.UIConfig {
	// Determinar caminho base para os ícones
	// Determine base path for icons
	// Determinar ruta base para los iconos
	iconsPath := filepath.Join("ui", "desktop", "assets", "icons", "platforms")

	// Selecionar pasta específica da plataforma
	// Select platform-specific folder
	// Seleccionar carpeta específica de la plataforma
	var platformDir string
	
	switch runtime.GOOS {
	case "windows":
		platformDir = "windows"
	case "darwin":
		platformDir = "macos"
	default:
		platformDir = "linux"
	}
	
	platformPath := filepath.Join(iconsPath, platformDir)
	
	// Configurar extensões de arquivo com base na plataforma
	// Configure file extensions based on platform
	// Configurar extensiones de archivo según la plataforma
	var appIconExt, trayIconExt, statusIconExt string
	
	switch runtime.GOOS {
	case "windows":
		appIconExt = "ico"
		trayIconExt = "ico"
		statusIconExt = "png"
	case "darwin":
		appIconExt = "png" // Idealmente seria icns, mas requer geração em macOS
		trayIconExt = "png"
		statusIconExt = "png"
	default:
		appIconExt = "png"
		trayIconExt = "png"
		statusIconExt = "png"
	}
	
	// Construir configuração com caminhos corretos
	// Build configuration with correct paths
	// Construir configuración con rutas correctas
	return &shared.UIConfig{
		Language: "pt-br", // Idioma padrão
		Theme: "system",   // Tema padrão
		StartMinimized: false,
		AutoStart: false,     // Não iniciar automaticamente por padrão
		Assets: shared.AssetConfig{
			AppIconPath: filepath.Join(platformPath, "app_icon."+appIconExt),
			ConnectedIconPath: filepath.Join(platformPath, "tray_connected."+trayIconExt),
			DisconnectedIconPath: filepath.Join(platformPath, "tray_disconnected."+trayIconExt),
			StatusConnectedIconPath: filepath.Join(platformPath, "status_connected."+statusIconExt),
			StatusDisconnectedIconPath: filepath.Join(platformPath, "status_disconnected."+statusIconExt),
		},
	}
}

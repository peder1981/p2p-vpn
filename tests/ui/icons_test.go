package ui_test

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/p2p-vpn/p2p-vpn/ui/desktop/assets"
)

// TestIconsExist verifica se os ícones originais SVG existem no sistema de arquivos
func TestIconsExist(t *testing.T) {
	// Verificar os SVGs originais
	iconPaths := []string{
		"/home/peder/p2p-vpn/ui/desktop/assets/icons/app_icon.svg",
		"/home/peder/p2p-vpn/ui/desktop/assets/icons/tray_connected.svg",
		"/home/peder/p2p-vpn/ui/desktop/assets/icons/tray_disconnected.svg",
		"/home/peder/p2p-vpn/ui/desktop/assets/icons/status_connected.svg",
		"/home/peder/p2p-vpn/ui/desktop/assets/icons/status_disconnected.svg",
	}
	
	for _, path := range iconPaths {
		_, err := os.Stat(path)
		if os.IsNotExist(err) {
			t.Errorf("Ícone SVG original não encontrado: %s", path)
		}
	}
	
	// Verificar se os diretórios específicos de plataforma existem
	platformDirs := []string{
		"/home/peder/p2p-vpn/ui/desktop/assets/icons/platforms/linux",
		"/home/peder/p2p-vpn/ui/desktop/assets/icons/platforms/macos",
		"/home/peder/p2p-vpn/ui/desktop/assets/icons/platforms/windows",
	}
	
	for _, dir := range platformDirs {
		_, err := os.Stat(dir)
		if os.IsNotExist(err) {
			t.Errorf("Diretório de ícones não encontrado: %s", dir)
		}
	}
}

// TestPlatformSpecificIcons verifica se os ícones específicos da plataforma estão configurados corretamente
func TestPlatformSpecificIcons(t *testing.T) {
	config := assets.GetDefaultConfig()
	
	// Verificar extensões de arquivo apropriadas para a plataforma
	var expectedAppIconExt string
	
	switch runtime.GOOS {
	case "windows":
		expectedAppIconExt = ".ico"
	case "darwin":
		// No macOS, seria .icns, mas estamos usando .png como fallback
		expectedAppIconExt = ".png" 
	default:
		expectedAppIconExt = ".png"
	}
	
	// Verificar se o ícone principal tem a extensão correta para a plataforma
	if !hasExtension(config.Assets.AppIconPath, expectedAppIconExt) {
		t.Errorf("Ícone principal não tem a extensão correta para a plataforma. Esperado: %s, Caminho: %s", 
			expectedAppIconExt, config.Assets.AppIconPath)
	}
	
	// Verificar se os ícones estão no diretório específico da plataforma
	platformDir := ""
	switch runtime.GOOS {
	case "windows":
		platformDir = "windows"
	case "darwin":
		platformDir = "macos"
	default:
		platformDir = "linux"
	}
	
	for _, path := range []string{
		config.Assets.AppIconPath,
		config.Assets.ConnectedIconPath,
		config.Assets.DisconnectedIconPath,
	} {
		if !containsInPath(path, platformDir) {
			t.Errorf("Ícone não está no diretório específico da plataforma %s: %s", platformDir, path)
		}
	}
}

// TestDefaultConfig verifica se a configuração padrão está correta
func TestDefaultConfig(t *testing.T) {
	config := assets.GetDefaultConfig()
	
	// Verificar valores padrão
	if config.Language != "pt-br" {
		t.Errorf("Idioma padrão incorreto: %s", config.Language)
	}
	
	if config.Theme != "system" {
		t.Errorf("Tema padrão incorreto: %s", config.Theme)
	}
	
	if config.StartMinimized {
		t.Error("O valor padrão de StartMinimized deve ser false")
	}
	
	if config.AutoStart {
		t.Error("O valor padrão de AutoStart deve ser false")
	}
}

// Funções auxiliares

// hasExtension verifica se um caminho termina com a extensão especificada
func hasExtension(path, ext string) bool {
	fileExt := filepath.Ext(path)
	return fileExt == ext
}

// containsInPath verifica se um caminho contém o segmento especificado
func containsInPath(path, segment string) bool {
	dirs := filepath.SplitList(path)
	for _, dir := range dirs {
		if dir == segment {
			return true
		}
	}
	
	// Verificação adicional para caminhos usando barras
	return filepath.ToSlash(path) != "" && filepath.Base(filepath.Dir(path)) == segment
}

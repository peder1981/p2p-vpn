// +build windows

package platform_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/p2p-vpn/p2p-vpn/platform"
)

// TestWindowsPlatformDetection verifica a detecção da plataforma Windows
// TestWindowsPlatformDetection verifies Windows platform detection
// TestWindowsPlatformDetection verifica la detección de la plataforma Windows
func TestWindowsPlatformDetection(t *testing.T) {
	// Verificar se estamos executando no Windows
	if os.Getenv("GOOS") != "windows" && os.Getenv("GOOS") != "" {
		t.Skip("Teste específico para Windows, pulando...")
	}

	// Obter a plataforma
	plat, err := platform.GetPlatform()
	if err != nil {
		t.Fatalf("Falha ao detectar plataforma Windows: %v", err)
	}

	// Verificar se a plataforma detectada é Windows
	platName := plat.Name()
	if platName != "Windows" && !strings.Contains(platName, "Windows") {
		t.Errorf("Plataforma Windows não detectada corretamente: %s", platName)
	}

	t.Logf("Plataforma Windows detectada: %s", platName)
}

// TestWindowsDependencies verifica se as dependências do Windows estão disponíveis
// TestWindowsDependencies checks if Windows dependencies are available
// TestWindowsDependencies verifica si las dependencias de Windows están disponibles
func TestWindowsDependencies(t *testing.T) {
	// Verificar se estamos executando no Windows
	if os.Getenv("GOOS") != "windows" && os.Getenv("GOOS") != "" {
		t.Skip("Teste específico para Windows, pulando...")
	}

	// Verificar se o WireGuard está instalado
	programFiles := os.Getenv("ProgramFiles")
	if programFiles == "" {
		programFiles = "C:\\Program Files"
	}
	
	wireguardPath := filepath.Join(programFiles, "WireGuard")
	wireguardExe := filepath.Join(wireguardPath, "wireguard.exe")
	
	_, err := os.Stat(wireguardExe)
	if err != nil {
		// Tentar Program Files (x86)
		programFilesX86 := os.Getenv("ProgramFiles(x86)")
		if programFilesX86 != "" {
			wireguardPath = filepath.Join(programFilesX86, "WireGuard")
			wireguardExe = filepath.Join(wireguardPath, "wireguard.exe")
			
			_, err = os.Stat(wireguardExe)
			if err != nil {
				t.Log("WireGuard não encontrado no caminho padrão.")
			} else {
				t.Logf("WireGuard encontrado em: %s", wireguardPath)
			}
		} else {
			t.Log("WireGuard não encontrado no caminho padrão.")
		}
	} else {
		t.Logf("WireGuard encontrado em: %s", wireguardPath)
	}

	// Verificar se o PowerShell está disponível para scripts
	_, err = exec.LookPath("powershell.exe")
	if err != nil {
		t.Log("PowerShell não encontrado, alguns recursos podem não funcionar corretamente.")
	} else {
		t.Log("PowerShell disponível para automação.")
	}
}

// TestWindowsInterfaceOperations verifica operações básicas de interface no Windows
// TestWindowsInterfaceOperations checks basic interface operations on Windows
// TestWindowsInterfaceOperations verifica operaciones básicas de interfaz en Windows
func TestWindowsInterfaceOperations(t *testing.T) {
	// Verificar se estamos executando no Windows
	if os.Getenv("GOOS") != "windows" && os.Getenv("GOOS") != "" {
		t.Skip("Teste específico para Windows, pulando...")
	}

	// Verificar se o teste está sendo executado como administrador
	isAdmin := false
	cmd := exec.Command("powershell.exe", "-Command", "([Security.Principal.WindowsPrincipal][Security.Principal.WindowsIdentity]::GetCurrent()).IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)")
	output, err := cmd.Output()
	if err == nil && strings.TrimSpace(string(output)) == "True" {
		isAdmin = true
	}

	if !isAdmin {
		t.Skip("Este teste requer privilégios de administrador, pulando...")
	}

	// Obter a plataforma
	plat, err := platform.GetPlatform()
	if err != nil {
		t.Fatalf("Falha ao detectar plataforma Windows: %v", err)
	}

	// Nome da interface de teste
	testInterface := "wgtest0"

	// Limpar qualquer interface de teste anterior
	_ = plat.RemoveWireGuardInterface(testInterface)

	// Testes básicos de criação de interface
	t.Run("CreateInterface", func(t *testing.T) {
		// Criar interface de teste com chave privada dummy
		dummyKey := "2BJtcgPUOMHmHQ4hKMYfwEQhW5Y9XYJHW0C1vQCravU="
		err := plat.CreateWireGuardInterface(testInterface, 51820, dummyKey)
		if err != nil {
			t.Fatalf("Falha ao criar interface WireGuard: %v", err)
		}

		// Verificar se a interface existe
		status, err := plat.GetInterfaceStatus(testInterface)
		if err != nil {
			t.Errorf("Erro ao verificar status da interface: %v", err)
		}
		if !status {
			t.Error("Interface não foi criada corretamente")
		}
	})

	// Limpar a interface de teste após os testes
	t.Cleanup(func() {
		_ = plat.RemoveWireGuardInterface(testInterface)
	})
}

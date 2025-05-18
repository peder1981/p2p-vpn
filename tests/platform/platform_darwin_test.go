// +build darwin

package platform_test

import (
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/p2p-vpn/p2p-vpn/platform"
)

// TestDarwinPlatformDetection verifica a detecção da plataforma macOS
// TestDarwinPlatformDetection verifies macOS platform detection
// TestDarwinPlatformDetection verifica la detección de la plataforma macOS
func TestDarwinPlatformDetection(t *testing.T) {
	// Verificar se estamos executando no macOS
	if os.Getenv("GOOS") != "darwin" && os.Getenv("GOOS") != "" {
		t.Skip("Teste específico para macOS, pulando...")
	}

	// Obter a plataforma
	plat, err := platform.GetPlatform()
	if err != nil {
		t.Fatalf("Falha ao detectar plataforma macOS: %v", err)
	}

	// Verificar se a plataforma detectada é macOS
	platName := plat.Name()
	if platName != "macOS" && !strings.Contains(platName, "macOS") && !strings.Contains(platName, "Darwin") {
		t.Errorf("Plataforma macOS não detectada corretamente: %s", platName)
	}

	t.Logf("Plataforma macOS detectada: %s", platName)
}

// TestDarwinDependencies verifica se as dependências do macOS estão disponíveis
// TestDarwinDependencies checks if macOS dependencies are available
// TestDarwinDependencies verifica si las dependencias de macOS están disponibles
func TestDarwinDependencies(t *testing.T) {
	// Verificar se estamos executando no macOS
	if os.Getenv("GOOS") != "darwin" && os.Getenv("GOOS") != "" {
		t.Skip("Teste específico para macOS, pulando...")
	}

	// Verificar se o Homebrew está instalado
	_, err := exec.LookPath("brew")
	if err != nil {
		t.Log("Homebrew não encontrado. Isso pode afetar a instalação de dependências.")
	} else {
		t.Log("Homebrew disponível para gerenciamento de pacotes.")
	}

	// Verificar se o wireguard-go está instalado
	_, err = exec.LookPath("wireguard-go")
	if err != nil {
		t.Log("wireguard-go não encontrado. O sistema pode precisar usar uma implementação alternativa.")
	} else {
		t.Log("wireguard-go disponível para implementação userspace.")
	}

	// Verificar o ambiente de rede
	cmd := exec.Command("ifconfig")
	output, err := cmd.Output()
	if err != nil {
		t.Logf("Não foi possível verificar interfaces de rede: %v", err)
	} else {
		t.Log("Interfaces de rede disponíveis para configuração.")
	}

	// Verificar permissões para operações de rede
	if os.Geteuid() != 0 {
		t.Log("Teste não está sendo executado como root. Algumas operações de rede podem falhar.")
	} else {
		t.Log("Teste executando com privilégios de root. Operações de rede devem funcionar.")
	}
}

// TestDarwinInterfaceOperations verifica operações básicas de interface no macOS
// TestDarwinInterfaceOperations checks basic interface operations on macOS
// TestDarwinInterfaceOperations verifica operaciones básicas de interfaz en macOS
func TestDarwinInterfaceOperations(t *testing.T) {
	// Verificar se estamos executando no macOS
	if os.Getenv("GOOS") != "darwin" && os.Getenv("GOOS") != "" {
		t.Skip("Teste específico para macOS, pulando...")
	}

	// Verificar se temos permissão root
	if os.Geteuid() != 0 {
		t.Skip("Este teste requer privilégios de root, pulando...")
	}

	// Obter a plataforma
	plat, err := platform.GetPlatform()
	if err != nil {
		t.Fatalf("Falha ao detectar plataforma macOS: %v", err)
	}

	// Nome da interface de teste
	testInterface := "utun7"

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

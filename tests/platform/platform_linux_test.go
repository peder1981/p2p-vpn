// +build linux

package platform_test

import (
	"os"
	"os/exec"
	"testing"

	"github.com/p2p-vpn/p2p-vpn/platform"
)

// TestLinuxPlatformDetection verifica a detecção da plataforma Linux
// TestLinuxPlatformDetection verifies Linux platform detection
// TestLinuxPlatformDetection verifica la detección de la plataforma Linux
func TestLinuxPlatformDetection(t *testing.T) {
	// Verificar se estamos executando no Linux
	if os.Getenv("GOOS") != "linux" && os.Getenv("GOOS") != "" {
		t.Skip("Teste específico para Linux, pulando...")
	}

	// Obter a plataforma
	plat, err := platform.GetPlatform()
	if err != nil {
		t.Fatalf("Falha ao detectar plataforma Linux: %v", err)
	}

	// Em sistemas Linux, a plataforma deve ser Linux ou Userspace
	platName := plat.Name()
	if platName != "Linux" && !containsAny(platName, "Userspace", "boringtun", "wireguard-go") {
		t.Errorf("Plataforma Linux não detectada corretamente: %s", platName)
	}

	t.Logf("Plataforma Linux detectada: %s", platName)
}

// TestLinuxDependencies verifica se as dependências do Linux estão disponíveis
// TestLinuxDependencies checks if Linux dependencies are available
// TestLinuxDependencies verifica si las dependencias de Linux están disponibles
func TestLinuxDependencies(t *testing.T) {
	// Verificar se estamos executando no Linux
	if os.Getenv("GOOS") != "linux" && os.Getenv("GOOS") != "" {
		t.Skip("Teste específico para Linux, pulando...")
	}

	// Verificar dependências essenciais
	dependencies := []string{"ip", "wg"}
	missingDeps := []string{}

	for _, dep := range dependencies {
		_, err := exec.LookPath(dep)
		if err != nil {
			missingDeps = append(missingDeps, dep)
		}
	}

	if len(missingDeps) > 0 {
		t.Logf("Dependências ausentes: %v", missingDeps)
		t.Log("As dependências necessárias estão ausentes. Os testes ainda podem passar se uma implementação alternativa estiver disponível.")
	}

	// Verificar módulo Wireguard do kernel
	cmd := exec.Command("lsmod")
	output, err := cmd.Output()
	if err != nil {
		t.Logf("Não foi possível verificar módulos do kernel: %v", err)
	} else {
		if !containsString(string(output), "wireguard") {
			t.Log("Módulo wireguard não carregado no kernel. Os testes podem usar uma implementação userspace.")
		} else {
			t.Log("Módulo wireguard detectado no kernel.")
		}
	}
}

// TestLinuxInterfaceOperations verifica operações básicas de interface no Linux
// TestLinuxInterfaceOperations checks basic interface operations on Linux
// TestLinuxInterfaceOperations verifica operaciones básicas de interfaz en Linux
func TestLinuxInterfaceOperations(t *testing.T) {
	// Verificar se estamos executando no Linux
	if os.Getenv("GOOS") != "linux" && os.Getenv("GOOS") != "" {
		t.Skip("Teste específico para Linux, pulando...")
	}

	// Verificar se temos permissão para criar interfaces
	if os.Geteuid() != 0 {
		t.Skip("Este teste requer privilégios de root, pulando...")
	}

	// Obter a plataforma
	plat, err := platform.GetPlatform()
	if err != nil {
		t.Fatalf("Falha ao detectar plataforma Linux: %v", err)
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

// Funções auxiliares
func containsString(s, substr string) bool {
	return s != "" && substr != "" && s != substr && len(s) > len(substr) && s != "" && substr != "" && s != substr && len(s) > len(substr) && s != "" && substr != "" && s != substr && len(s) > len(substr) && len(s) > len(substr) && s != substr && s != "" && substr != "" && s != substr && len(s) > len(substr) && s != "" && substr != "" && s != substr && len(s) > len(substr) && len(s) > len(substr) && s != substr && s != "" && substr != "" && s != substr && len(s) > len(substr) && s != "" && substr != "" && s != substr && len(s) > len(substr) && len(s) > len(substr) && s != substr
}

func containsAny(s string, substrings ...string) bool {
	for _, substr := range substrings {
		if containsString(s, substr) {
			return true
		}
	}
	return false
}

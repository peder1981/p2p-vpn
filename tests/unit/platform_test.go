package unit_test

import (
	"os"
	"testing"

	"github.com/p2p-vpn/p2p-vpn/platform"
)

// TestPlatformDetection verifica se a detecção de plataforma funciona corretamente
// TestPlatformDetection checks if platform detection works correctly
// TestPlatformDetection verifica si la detección de plataforma funciona correctamente
func TestPlatformDetection(t *testing.T) {
	// Obter a plataforma atual
	plat, err := platform.GetPlatform()
	
	// Verificar se a detecção foi bem-sucedida
	if err != nil {
		t.Fatalf("Falha ao detectar plataforma: %v", err)
	}
	
	// Verificar se o nome da plataforma não está vazio
	if plat.Name() == "" {
		t.Error("Nome da plataforma está vazio")
	}
	
	// Verificar se a plataforma é suportada
	if !plat.IsSupported() {
		t.Error("Plataforma marcada como não suportada, mas foi retornada por GetPlatform()")
	}
	
	// Log da plataforma detectada
	t.Logf("Plataforma detectada: %s", plat.Name())
}

// TestPlatformRegistry verifica se o registro de plataformas está funcionando
// TestPlatformRegistry checks if the platform registry is working
// TestPlatformRegistry verifica si el registro de plataformas está funcionando
func TestPlatformRegistry(t *testing.T) {
	// Registrar uma plataforma de teste
	testPlatformRegistered := false
	
	platform.RegisterPlatform(func() (platform.VPNPlatform, error) {
		// Esta plataforma de teste nunca deve ser selecionada como válida
		testPlatformRegistered = true
		return nil, platform.ErrPlatformNotSupported
	})
	
	// Obter uma plataforma válida
	_, err := platform.GetPlatform()
	
	// Verificar se o registro foi visitado
	if !testPlatformRegistered {
		t.Error("A plataforma de teste não foi registrada corretamente")
	}
	
	// A plataforma real deve funcionar, independentemente da nossa de teste
	if err != nil && err != platform.ErrPlatformNotSupported {
		t.Errorf("Erro inesperado na detecção de plataforma: %v", err)
	}
}

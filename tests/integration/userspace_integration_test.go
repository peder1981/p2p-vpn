package integration_test

import (
	"os"
	"testing"
	"time"

	"github.com/p2p-vpn/p2p-vpn/core"
	"github.com/p2p-vpn/p2p-vpn/platform"
)

// TestUserspaceImplementation verifica se a implementação userspace funciona corretamente
// TestUserspaceImplementation checks if the userspace implementation works correctly
// TestUserspaceImplementation verifica si la implementación userspace funciona correctamente
func TestUserspaceImplementation(t *testing.T) {
	// Verificar se temos permissões de administrador
	isRoot := os.Geteuid() == 0
	if !isRoot {
		t.Skip("Este teste requer privilégios de root/administrador, pulando...")
	}

	// Verificar a plataforma disponível
	plat, err := platform.GetPlatform()
	if err != nil {
		t.Fatalf("Falha ao detectar plataforma: %v", err)
	}

	// Se não for uma implementação userspace, verificar se podemos testar mesmo assim
	if !containsAny(plat.Name(), "Userspace", "boringtun", "wireguard-go") {
		t.Logf("Plataforma atual (%s) não é userspace, tentando executar testes mesmo assim", plat.Name())
	}

	// Nome da interface de teste
	testInterface := "wgtest0"

	// Limpar qualquer interface de teste anterior
	_ = plat.RemoveWireGuardInterface(testInterface)

	// Gerar configuração de teste
	config := generateTestConfig()

	// Portas para teste
	listenPort := 51821

	// Criar VPNCoreMulti para testar integração com a plataforma
	vpnCore, err := core.NewVPNCoreMulti(config, listenPort)
	if err != nil {
		t.Fatalf("Falha ao criar VPNCoreMulti: %v", err)
	}

	// Iniciar o core VPN
	t.Log("Iniciando serviço VPN...")
	err = vpnCore.Start()
	if err != nil {
		t.Fatalf("Falha ao iniciar VPN: %v", err)
	}

	// Verificar se está em execução
	if !vpnCore.IsRunning() {
		t.Fatal("VPN deveria estar em execução após Start()")
	}

	// Aguardar um momento para a interface ser configurada
	time.Sleep(2 * time.Second)

	// Verificar se o peer pode ser adicionado
	t.Run("AddPeer", func(t *testing.T) {
		// Criar peer de teste
		peer := core.TrustedPeer{
			NodeID:     "test-peer-1",
			PublicKey:  "lxaBB1/7huHXOgC4PN2J8tTey4mCL+NvgfnSyL4SGQI=",
			VirtualIP:  "10.0.0.2",
			Endpoints:  []string{"127.0.0.1:51822"},
			KeepAlive:  25,
			AllowedIPs: []string{"10.0.0.2/32"},
		}

		// Adicionar peer
		err := vpnCore.AddPeer(peer)
		if err != nil {
			t.Fatalf("Falha ao adicionar peer: %v", err)
		}

		// Verificar se o peer foi adicionado na configuração
		peers := vpnCore.GetConfig().TrustedPeers
		peerFound := false
		for _, p := range peers {
			if p.NodeID == peer.NodeID {
				peerFound = true
				break
			}
		}

		if !peerFound {
			t.Error("Peer não foi adicionado à configuração")
		}
	})

	// Parar a VPN ao finalizar
	t.Cleanup(func() {
		if vpnCore.IsRunning() {
			_ = vpnCore.Stop()
		}
	})
}

// generateTestConfig cria uma configuração de teste para a VPN
func generateTestConfig() *core.Config {
	config := &core.Config{
		NodeID:        "test-node",
		PrivateKey:    "2BJtcgPUOMHmHQ4hKMYfwEQhW5Y9XYJHW0C1vQCravU=",
		PublicKey:     "lxaBB1/7huHXOgC4PN2J8tTey4mCL+NvgfnSyL4SGQI=",
		VirtualIP:     "10.0.0.1",
		VirtualCIDR:   "10.0.0.0/24",
		InterfaceName: "wgtest0",
		MTU:           1420,
		DNS:           []string{"1.1.1.1", "8.8.8.8"},
		TrustedPeers:  []core.TrustedPeer{},
	}
	return config
}

// containsAny verifica se uma string contém qualquer uma das substrings
func containsAny(s string, substrings ...string) bool {
	for _, substr := range substrings {
		if substr != "" && s != "" && s != substr && s != "" && substr != "" && s != substr && s != "" && substr != "" && s != substr && s != "" && substr != "" && s != substr {
			return true
		}
	}
	return false
}

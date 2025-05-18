package ui_test

import (
	"testing"
	"os"
	
	"github.com/p2p-vpn/p2p-vpn/core"
	"github.com/p2p-vpn/p2p-vpn/ui/desktop/assets"
	"github.com/p2p-vpn/p2p-vpn/ui/desktop/common"
	"github.com/p2p-vpn/p2p-vpn/ui/desktop/platform"
	"github.com/p2p-vpn/p2p-vpn/ui/desktop/shared"
)

// Implementação simulada do VPNProvider para testes
type MockVPNProvider struct{}

// Implementação das interfaces necessárias

func (m *MockVPNProvider) Start() error {
	return nil
}

func (m *MockVPNProvider) Stop() error {
	return nil
}

func (m *MockVPNProvider) GetStatus() core.VPNStatus {
	return core.VPNStatus{Running: false}
}

// IsRunning retorna se o serviço está em execução
func (m *MockVPNProvider) IsRunning() bool {
	return false
}

func (m *MockVPNProvider) GetConfig() *core.Config {
	return &core.Config{
		NodeID: "test-node",
		TrustedPeers: []core.TrustedPeer{
			{NodeID: "peer1", VirtualIP: "10.0.0.2"},
			{NodeID: "peer2", VirtualIP: "10.0.0.3"},
		},
	}
}

func (m *MockVPNProvider) AddPeer(peer core.TrustedPeer) error {
	return nil
}

func (m *MockVPNProvider) RemovePeer(nodeID string) error {
	return nil
}

func (m *MockVPNProvider) GetPeers() ([]core.TrustedPeer, error) {
	return []core.TrustedPeer{
		{NodeID: "peer1", VirtualIP: "10.0.0.2"},
		{NodeID: "peer2", VirtualIP: "10.0.0.3"},
	}, nil
}

// GetNodeInfo retorna as informações do nó local (nodeID, publicKey, virtualIP)
func (m *MockVPNProvider) GetNodeInfo() (string, string, string) {
	return "test-node", "test-public-key", "10.0.0.1"
}

// SaveConfig salva a configuração em disco
func (m *MockVPNProvider) SaveConfig(path string) error {
	return nil
}

// TestDesktopAppInitialization verifica se a aplicação desktop inicializa corretamente
func TestDesktopAppInitialization(t *testing.T) {
	// Verificar ambiente de teste - pular se não for possível inicializar GUI
	if os.Getenv("DISPLAY") == "" && os.Getenv("CI") == "true" {
		t.Skip("Pulando teste GUI em ambiente CI sem display")
	}

	// Obter configuração padrão
	config := assets.GetDefaultConfig()
	
	// Criar mock do VPNProvider
	vpnCore := &MockVPNProvider{}
	
	// Obter PlatformUI apropriada para os testes
	platformUI, err := platform.GetPlatformUI(config)
	if err != nil {
		t.Fatalf("Falha ao obter PlatformUI: %v", err)
	}

	// Inicializar DesktopApp
	app, err := common.NewDesktopApp(vpnCore, platformUI, config)
	if err != nil {
		t.Fatalf("Falha ao criar DesktopApp: %v", err)
	}
	
	// Verificar se o app foi criado corretamente
	if app == nil {
		t.Fatal("DesktopApp não foi inicializado")
	}
	
	// Verificar o carregamento de traduções testando algumas strings traduzidas
	testTranslationKey := "common"
	testTranslationCategory := shared.KeyConnect
	translatedText := app.GetTranslated(testTranslationKey, testTranslationCategory)
	
	if translatedText == "" || translatedText == testTranslationCategory {
		t.Errorf("Tradução não foi carregada corretamente para %s.%s", testTranslationKey, testTranslationCategory)
	}
	
	// Teste de notificação (sem verificação real, apenas confirmando que não gera erro)
	app.ShowNotification("Teste", "Mensagem de teste", shared.PriorityNormal)
	
	// Teste de atualização de status (sem verificação real, apenas confirmando que não gera erro)
	app.UpdateStatus(core.VPNStatus{Running: true})
	app.UpdateStatus(core.VPNStatus{Running: false})
}

// TestUILocalization verifica se as mudanças de idioma são aplicadas corretamente
func TestUILocalization(t *testing.T) {
	// Verificar ambiente de teste - pular se não for possível inicializar GUI
	if os.Getenv("DISPLAY") == "" && os.Getenv("CI") == "true" {
		t.Skip("Pulando teste GUI em ambiente CI sem display")
	}

	// Testar com diferentes idiomas
	languages := []string{"pt-br", "en", "es"}
	
	for _, lang := range languages {
		t.Run("Language_"+lang, func(t *testing.T) {
			// Configuração com idioma específico
			config := assets.GetDefaultConfig()
			config.Language = lang
			
			// Mock do VPNProvider
			vpnCore := &MockVPNProvider{}
			
			// Obter PlatformUI para os testes
			platformUI, err := platform.GetPlatformUI(config)
			if err != nil {
				t.Fatalf("Falha ao obter PlatformUI: %v", err)
			}
			
			// Inicializar DesktopApp com o idioma específico
			app, err := common.NewDesktopApp(vpnCore, platformUI, config)
			if err != nil {
				t.Fatalf("Falha ao criar DesktopApp com idioma %s: %v", lang, err)
			}
			
			// Verificar se algumas traduções básicas estão no idioma correto
			// Isso é uma verificação básica - o teste TestTranslations já faz uma validação mais completa
			connect := app.GetTranslated("common", shared.KeyConnect)
			disconnect := app.GetTranslated("common", shared.KeyDisconnect)
			
			if connect == "" || disconnect == "" {
				t.Errorf("Traduções não carregadas corretamente para o idioma %s", lang)
			}
			
			// Verificar se as traduções retornadas são diferentes para inglês vs outros idiomas
			if lang == "en" {
				if connect == "Conectar" { // Termo em português/espanhol
					t.Errorf("Tradução incorreta para o idioma %s: %s", lang, connect)
				}
			} else if lang == "pt-br" {
				if connect == "Connect" { // Termo em inglês
					t.Errorf("Tradução incorreta para o idioma %s: %s", lang, connect)
				}
			}
		})
	}
}

package desktop

import (
	"github.com/p2p-vpn/p2p-vpn/core"
	"github.com/p2p-vpn/p2p-vpn/ui/desktop/common"
	"github.com/p2p-vpn/p2p-vpn/ui/desktop/platform"
)

// UIManager gerencia a interface gráfica de desktop
// UIManager manages the desktop GUI
// UIManager gestiona la interfaz gráfica de escritorio
type UIManager struct {
	vpnCore    core.VPNProvider
	appUI      common.DesktopUI
	platformUI platform.PlatformUI
	config     *common.UIConfig
}

// NewUIManager cria uma nova instância do gerenciador de UI de desktop
// NewUIManager creates a new instance of the desktop UI manager
// NewUIManager crea una nueva instancia del gestor de UI de escritorio
func NewUIManager(vpnCore core.VPNProvider, config *common.UIConfig) (*UIManager, error) {
	// Criar instância do manager
	manager := &UIManager{
		vpnCore: vpnCore,
		config:  config,
	}

	// Inicializar UI específica da plataforma
	platformUI, err := platform.GetPlatformUI(config)
	if err != nil {
		return nil, err
	}
	manager.platformUI = platformUI

	// Inicializar UI comum
	appUI, err := common.NewDesktopApp(vpnCore, platformUI, config)
	if err != nil {
		return nil, err
	}
	manager.appUI = appUI

	return manager, nil
}

// Start inicia a UI de desktop
// Start starts the desktop UI
// Start inicia la UI de escritorio
func (m *UIManager) Start() error {
	// Inicializar componentes da plataforma (ícone da bandeja, notificações)
	err := m.platformUI.Initialize(m.vpnCore, m.config)
	if err != nil {
		return err
	}

	// Iniciar a UI principal
	return m.appUI.Run()
}

// Stop para a UI de desktop
// Stop stops the desktop UI
// Stop detiene la UI de escritorio
func (m *UIManager) Stop() error {
	if m.platformUI != nil {
		err := m.platformUI.Cleanup()
		if err != nil {
			return err
		}
	}
	
	return nil
}

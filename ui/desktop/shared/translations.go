package shared

// Translations armazena todas as traduções para diferentes idiomas
// Translations stores all translations for different languages
// Translations almacena todas las traducciones para diferentes idiomas
type Translations struct {
	// Chave de idioma (pt-br, en, es)
	LanguageKey string

	// Traduções de interface comum
	Common map[string]string

	// Traduções específicas do menu
	Menu map[string]string

	// Traduções de status
	Status map[string]string

	// Traduções de notificações
	Notifications map[string]string

	// Traduções de diálogos
	Dialogs map[string]string
}

// Chaves de tradução comuns
// Common translation keys
// Claves de traducción comunes
const (
	// Comum
	KeyAppName       = "app_name"
	KeyConnect       = "connect"
	KeyDisconnect    = "disconnect"
	KeySettings      = "settings"
	KeyExit          = "exit"
	KeyCancel        = "cancel"
	KeySave          = "save"
	KeyAbout         = "about"
	KeyVersion       = "version"
	KeyHelp          = "help"
	KeyStatus        = "status"
	KeyOK            = "ok"

	// Menu
	KeyShowWindow    = "show_window"
	KeyHideWindow    = "hide_window"
	KeyShowConsole   = "show_console"
	KeyStartVPN      = "start_vpn"
	KeyStopVPN       = "stop_vpn"

	// Status
	KeyStatusConnected      = "status_connected"
	KeyStatusDisconnected   = "status_disconnected"
	KeyStatusConnecting     = "status_connecting"
	KeyStatusDisconnecting  = "status_disconnecting"
	KeyStatusError          = "status_error"

	// Notificações
	KeyNotifyConnected       = "notify_connected"
	KeyNotifyDisconnected    = "notify_disconnected"
	KeyNotifyConnectionError = "notify_connection_error"
	KeyNotifyPeerConnected   = "notify_peer_connected"
	KeyNotifyPeerDisconnected = "notify_peer_disconnected"

	// Diálogos
	KeyDialogConfirmExit    = "dialog_confirm_exit"
	KeyDialogConfirmDisconnect = "dialog_confirm_disconnect"
)

// GetTranslations retorna todas as traduções para o idioma especificado
// GetTranslations returns all translations for the specified language
// GetTranslations devuelve todas las traducciones para el idioma especificado
func GetTranslations(languageKey string) Translations {
	// Se o idioma não for suportado, usar português como padrão
	if _, ok := supportedTranslations[languageKey]; !ok {
		languageKey = "pt-br"
	}

	return supportedTranslations[languageKey]
}

// Mapa com todas as traduções suportadas
// Map with all supported translations
// Mapa con todas las traducciones soportadas
var supportedTranslations = map[string]Translations{
	"pt-br": {
		LanguageKey: "pt-br",
		Common: map[string]string{
			KeyAppName:      "P2P VPN",
			KeyConnect:      "Conectar",
			KeyDisconnect:   "Desconectar",
			KeySettings:     "Configurações",
			KeyExit:         "Sair",
			KeyCancel:       "Cancelar",
			KeySave:         "Salvar",
			KeyAbout:        "Sobre",
			KeyVersion:      "Versão",
			KeyHelp:         "Ajuda",
			KeyStatus:       "Status",
			KeyOK:           "OK",
		},
		Menu: map[string]string{
			KeyShowWindow:   "Mostrar Janela",
			KeyHideWindow:   "Ocultar Janela",
			KeyShowConsole:  "Mostrar Console",
			KeyStartVPN:     "Iniciar VPN",
			KeyStopVPN:      "Parar VPN",
		},
		Status: map[string]string{
			KeyStatusConnected:     "Conectado",
			KeyStatusDisconnected:  "Desconectado",
			KeyStatusConnecting:    "Conectando...",
			KeyStatusDisconnecting: "Desconectando...",
			KeyStatusError:         "Erro de Conexão",
		},
		Notifications: map[string]string{
			KeyNotifyConnected:        "VPN Conectada",
			KeyNotifyDisconnected:     "VPN Desconectada",
			KeyNotifyConnectionError:  "Erro ao conectar à VPN",
			KeyNotifyPeerConnected:    "Peer %s conectado",
			KeyNotifyPeerDisconnected: "Peer %s desconectado",
		},
		Dialogs: map[string]string{
			KeyDialogConfirmExit:        "Tem certeza que deseja sair? A conexão VPN será encerrada.",
			KeyDialogConfirmDisconnect:  "Tem certeza que deseja desconectar da VPN?",
		},
	},
	"en": {
		LanguageKey: "en",
		Common: map[string]string{
			KeyAppName:      "P2P VPN",
			KeyConnect:      "Connect",
			KeyDisconnect:   "Disconnect",
			KeySettings:     "Settings",
			KeyExit:         "Exit",
			KeyCancel:       "Cancel",
			KeySave:         "Save",
			KeyAbout:        "About",
			KeyVersion:      "Version",
			KeyHelp:         "Help",
			KeyStatus:       "Status",
			KeyOK:           "OK",
		},
		Menu: map[string]string{
			KeyShowWindow:   "Show Window",
			KeyHideWindow:   "Hide Window",
			KeyShowConsole:  "Show Console",
			KeyStartVPN:     "Start VPN",
			KeyStopVPN:      "Stop VPN",
		},
		Status: map[string]string{
			KeyStatusConnected:     "Connected",
			KeyStatusDisconnected:  "Disconnected",
			KeyStatusConnecting:    "Connecting...",
			KeyStatusDisconnecting: "Disconnecting...",
			KeyStatusError:         "Connection Error",
		},
		Notifications: map[string]string{
			KeyNotifyConnected:        "VPN Connected",
			KeyNotifyDisconnected:     "VPN Disconnected",
			KeyNotifyConnectionError:  "Error connecting to VPN",
			KeyNotifyPeerConnected:    "Peer %s connected",
			KeyNotifyPeerDisconnected: "Peer %s disconnected",
		},
		Dialogs: map[string]string{
			KeyDialogConfirmExit:        "Are you sure you want to exit? The VPN connection will be terminated.",
			KeyDialogConfirmDisconnect:  "Are you sure you want to disconnect from the VPN?",
		},
	},
	"es": {
		LanguageKey: "es",
		Common: map[string]string{
			KeyAppName:      "P2P VPN",
			KeyConnect:      "Conectar",
			KeyDisconnect:   "Desconectar",
			KeySettings:     "Configuración",
			KeyExit:         "Salir",
			KeyCancel:       "Cancelar",
			KeySave:         "Guardar",
			KeyAbout:        "Acerca de",
			KeyVersion:      "Versión",
			KeyHelp:         "Ayuda",
			KeyStatus:       "Estado",
			KeyOK:           "OK",
		},
		Menu: map[string]string{
			KeyShowWindow:   "Mostrar Ventana",
			KeyHideWindow:   "Ocultar Ventana",
			KeyShowConsole:  "Mostrar Consola",
			KeyStartVPN:     "Iniciar VPN",
			KeyStopVPN:      "Detener VPN",
		},
		Status: map[string]string{
			KeyStatusConnected:     "Conectado",
			KeyStatusDisconnected:  "Desconectado",
			KeyStatusConnecting:    "Conectando...",
			KeyStatusDisconnecting: "Desconectando...",
			KeyStatusError:         "Error de Conexión",
		},
		Notifications: map[string]string{
			KeyNotifyConnected:        "VPN Conectada",
			KeyNotifyDisconnected:     "VPN Desconectada",
			KeyNotifyConnectionError:  "Error al conectar a la VPN",
			KeyNotifyPeerConnected:    "Par %s conectado",
			KeyNotifyPeerDisconnected: "Par %s desconectado",
		},
		Dialogs: map[string]string{
			KeyDialogConfirmExit:        "¿Está seguro de que desea salir? La conexión VPN se terminará.",
			KeyDialogConfirmDisconnect:  "¿Está seguro de que desea desconectar de la VPN?",
		},
	},
}

// GetTranslated retorna uma string traduzida com base na chave e categoria
// GetTranslated returns a translated string based on key and category
// GetTranslated devuelve una cadena traducida basada en la clave y categoría
func GetTranslated(translations Translations, category string, key string) string {
	var translatedMap map[string]string

	switch category {
	case "common":
		translatedMap = translations.Common
	case "menu":
		translatedMap = translations.Menu
	case "status":
		translatedMap = translations.Status
	case "notifications":
		translatedMap = translations.Notifications
	case "dialogs":
		translatedMap = translations.Dialogs
	default:
		translatedMap = translations.Common
	}

	if value, exists := translatedMap[key]; exists {
		return value
	}

	// Se a tradução não for encontrada, retornar a chave original
	return key
}

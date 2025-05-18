package shared

// NotificationPriority define níveis de prioridade para notificações
// NotificationPriority defines priority levels for notifications
// NotificationPriority define niveles de prioridad para notificaciones
type NotificationPriority int

const (
	// PriorityLow é usado para notificações de baixa importância
	// PriorityLow is used for low importance notifications
	// PriorityLow se utiliza para notificaciones de baja importancia
	PriorityLow NotificationPriority = iota
	
	// PriorityNormal é usado para notificações de importância normal
	// PriorityNormal is used for normal importance notifications
	// PriorityNormal se utiliza para notificaciones de importancia normal
	PriorityNormal
	
	// PriorityHigh é usado para notificações de alta importância
	// PriorityHigh is used for high importance notifications
	// PriorityHigh se utiliza para notificaciones de alta importancia
	PriorityHigh
)

// AssetConfig configuração de recursos gráficos
// AssetConfig configuration of graphic resources
// AssetConfig configuración de recursos gráficos
type AssetConfig struct {
	// Caminho para os ícones
	ConnectedIconPath        string // Ícone da bandeja para estado conectado
	DisconnectedIconPath     string // Ícone da bandeja para estado desconectado
	AppIconPath              string // Ícone principal da aplicação
	StatusConnectedIconPath    string // Ícone de status para estado conectado
	StatusDisconnectedIconPath string // Ícone de status para estado desconectado
}

// UIConfig configuração da interface do usuário
// UIConfig user interface configuration
// UIConfig configuración de la interfaz de usuario
type UIConfig struct {
	Language       string      // Idioma da interface / Interface language / Idioma de la interfaz
	Theme          string      // Tema da interface / Interface theme / Tema de la interfaz
	StartMinimized bool        // Iniciar minimizado / Start minimized / Iniciar minimizado
	AutoStart      bool        // Iniciar automaticamente com o sistema / Start automatically with system / Iniciar automáticamente con el sistema
	Assets         AssetConfig // Configuração de ativos gráficos / Graphic assets configuration / Configuración de activos gráficos
}

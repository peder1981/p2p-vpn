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
	ConnectedIconPath    string
	DisconnectedIconPath string
	AppIconPath          string
}

// UIConfig configuração da interface do usuário
// UIConfig user interface configuration
// UIConfig configuración de la interfaz de usuario
type UIConfig struct {
	// Idioma da interface (pt-br, en, es)
	Language string
	
	// Tema (light, dark, system)
	Theme string
	
	// Iniciar minimizado na bandeja
	StartMinimized bool
	
	// Recursos gráficos
	Assets AssetConfig
}

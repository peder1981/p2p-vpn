package core

// VPNStatus representa o estado atual da VPN
// VPNStatus represents the current state of the VPN
// VPNStatus representa el estado actual de la VPN
type VPNStatus struct {
	Running       bool   // Se a VPN está em execução
	ConnectionMsg string // Mensagem sobre a conexão atual
	ErrorMsg      string // Mensagem de erro, se houver
	ConnectedPeers int   // Número de peers conectados
	BytesSent     int64  // Bytes enviados
	BytesReceived int64  // Bytes recebidos
	UptimeSeconds int64  // Tempo de atividade em segundos
}

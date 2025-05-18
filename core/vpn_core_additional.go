package core

// GetConfig retorna a configuração atual da VPN
// GetConfig returns the current VPN configuration
// GetConfig devuelve la configuración actual de la VPN
func (v *VPNCore) GetConfig() *Config {
	v.mutex.Lock()
	defer v.mutex.Unlock()
	return v.config
}

// SaveConfig salva a configuração em disco
// SaveConfig saves the configuration to disk
// SaveConfig guarda la configuración en disco
func (v *VPNCore) SaveConfig(path string) error {
	v.mutex.Lock()
	defer v.mutex.Unlock()
	return v.config.SaveConfig(path)
}

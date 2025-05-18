package core

// GetNodeInfo retorna as informações do nó local
// GetNodeInfo returns the node's local information
// GetNodeInfo devuelve la información del nodo local
func (v *VPNCoreMulti) GetNodeInfo() (string, string, string) {
	v.mutex.Lock()
	defer v.mutex.Unlock()
	
	return v.config.NodeID, v.config.PublicKey, v.config.VirtualIP
}

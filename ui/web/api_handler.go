package web

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/p2p-vpn/p2p-vpn/core"
)

// APIHandler gerencia as requisições API para o frontend
// APIHandler manages API requests for the frontend
// APIHandler gestiona las solicitudes API para el frontend
type APIHandler struct {
	vpnCore core.VPNProvider
	config  *core.Config
}

// NewAPIHandler cria um novo manipulador de API
// NewAPIHandler creates a new API handler
// NewAPIHandler crea un nuevo manejador de API
func NewAPIHandler(vpnCore core.VPNProvider, config *core.Config) *APIHandler {
	return &APIHandler{
		vpnCore: vpnCore,
		config:  config,
	}
}

// ServeHTTP implementa a interface http.Handler
// ServeHTTP implements the http.Handler interface
// ServeHTTP implementa la interfaz http.Handler
func (h *APIHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Configurar cabeçalhos para CORS e JSON
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Content-Type", "application/json")

	// Tratar requisições OPTIONS (preflight CORS)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Extrair o path da API removendo o prefixo "/api/"
	path := strings.TrimPrefix(r.URL.Path, "/api/")
	
	// Rotear para o manipulador adequado
	switch {
	case path == "status" && r.Method == "GET":
		h.handleStatus(w, r)
	case path == "peers" && r.Method == "GET":
		h.handleGetPeers(w, r)
	case path == "peers" && r.Method == "POST":
		h.handleAddPeer(w, r)
	case strings.HasPrefix(path, "peers/") && r.Method == "DELETE":
		h.handleRemovePeer(w, r)
	case path == "config" && r.Method == "GET":
		h.handleGetConfig(w, r)
	default:
		// Rota não encontrada
		http.Error(w, `{"error": "Endpoint não encontrado"}`, http.StatusNotFound)
	}
}

// handleStatus retorna o status atual da VPN
func (h *APIHandler) handleStatus(w http.ResponseWriter, r *http.Request) {
	// Verificar se o core da VPN está disponível
	var isRunning bool
	if h.vpnCore != nil {
		isRunning = h.vpnCore.IsRunning()
	} else {
		isRunning = false
	}

	// Preparar resposta
	status := map[string]interface{}{
		"running":       isRunning,
		"node_id":       h.config.NodeID,
		"virtual_ip":    h.config.VirtualIP,
		"virtual_cidr":  h.config.VirtualCIDR,
		"peers_count":   len(h.config.TrustedPeers),
		"interface":     h.config.InterfaceName,
	}

	// Enviar resposta
	json.NewEncoder(w).Encode(status)
}

// handleGetPeers retorna a lista de peers configurados
func (h *APIHandler) handleGetPeers(w http.ResponseWriter, r *http.Request) {
	// Verificar se o core da VPN está disponível
	var activePeers []string
	if h.vpnCore != nil && h.vpnCore.IsRunning() {
		// Em uma implementação real, obteríamos isso do WireGuard
		// Por ora, apenas simulamos
		for _, peer := range h.config.TrustedPeers {
			activePeers = append(activePeers, peer.NodeID)
		}
	}

	// Construir resposta
	peersResponse := make([]map[string]interface{}, 0, len(h.config.TrustedPeers))
	for _, peer := range h.config.TrustedPeers {
		// Verificar se o peer está ativo
		isActive := false
		for _, activeID := range activePeers {
			if activeID == peer.NodeID {
				isActive = true
				break
			}
		}

		// Adicionar info do peer
		peerInfo := map[string]interface{}{
			"node_id":     peer.NodeID,
			"public_key":  peer.PublicKey,
			"virtual_ip":  peer.VirtualIP,
			"endpoints":   peer.Endpoints,
			"active":      isActive,
			"keep_alive":  peer.KeepAlive,
			"allowed_ips": peer.AllowedIPs,
		}
		
		peersResponse = append(peersResponse, peerInfo)
	}

	// Enviar resposta
	json.NewEncoder(w).Encode(peersResponse)
}

// PeerRequest representa uma solicitação para adicionar um peer
type PeerRequest struct {
	NodeID     string   `json:"node_id"`
	PublicKey  string   `json:"public_key"`
	VirtualIP  string   `json:"virtual_ip"`
	Endpoints  []string `json:"endpoints"`
	KeepAlive  int      `json:"keep_alive"`
	AllowedIPs []string `json:"allowed_ips"`
}

// handleAddPeer adiciona um novo peer à configuração
func (h *APIHandler) handleAddPeer(w http.ResponseWriter, r *http.Request) {
	// Decodificar solicitação
	var req PeerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "Erro ao decodificar solicitação"}`, http.StatusBadRequest)
		return
	}

	// Validar dados necessários
	if req.PublicKey == "" || req.VirtualIP == "" {
		http.Error(w, `{"error": "Chave pública e IP virtual são obrigatórios"}`, http.StatusBadRequest)
		return
	}

	// Gerar ID para o peer se não fornecido
	if req.NodeID == "" {
		req.NodeID = "peer-" + strings.Replace(req.VirtualIP, ".", "-", -1)
	}

	// Criar o peer
	peer := core.TrustedPeer{
		NodeID:     req.NodeID,
		PublicKey:  req.PublicKey,
		VirtualIP:  req.VirtualIP,
		Endpoints:  req.Endpoints,
		KeepAlive:  req.KeepAlive,
		AllowedIPs: req.AllowedIPs,
	}

	// Adicionar à configuração
	h.config.AddTrustedPeer(peer)

	// Se o VPN estiver rodando, adicionar o peer em tempo real
	if h.vpnCore != nil && h.vpnCore.IsRunning() {
		if err := h.vpnCore.AddPeer(peer); err != nil {
			// Erro ao adicionar ao WireGuard, mas mantemos na configuração
			response := map[string]interface{}{
				"success": true,
				"message": "Peer adicionado à configuração, mas erro ao ativar: " + err.Error(),
				"peer_id": peer.NodeID,
			}
			json.NewEncoder(w).Encode(response)
			return
		}
	}

	// Resposta de sucesso
	response := map[string]interface{}{
		"success": true,
		"message": "Peer adicionado com sucesso",
		"peer_id": peer.NodeID,
	}
	json.NewEncoder(w).Encode(response)
}

// handleRemovePeer remove um peer da configuração
func (h *APIHandler) handleRemovePeer(w http.ResponseWriter, r *http.Request) {
	// Extrair ID do peer da URL
	path := strings.TrimPrefix(r.URL.Path, "/api/peers/")
	nodeID := path

	if nodeID == "" {
		http.Error(w, `{"error": "ID do peer não fornecido"}`, http.StatusBadRequest)
		return
	}

	// Encontrar o peer na configuração
	var targetPeer *core.TrustedPeer
	for i, peer := range h.config.TrustedPeers {
		if peer.NodeID == nodeID {
			targetPeer = &h.config.TrustedPeers[i]
			break
		}
	}

	if targetPeer == nil {
		http.Error(w, `{"error": "Peer não encontrado"}`, http.StatusNotFound)
		return
	}

	// Remover da configuração WireGuard se estiver rodando
	if h.vpnCore != nil && h.vpnCore.IsRunning() {
		if err := h.vpnCore.RemovePeer(nodeID); err != nil {
			// Erro ao remover, mas continuamos para remover da configuração
			// (não crítico, pois a configuração será usada na próxima inicialização)
		}
	}

	// Remover da configuração
	if !h.config.RemoveTrustedPeer(nodeID) {
		http.Error(w, `{"error": "Erro ao remover peer da configuração"}`, http.StatusInternalServerError)
		return
	}

	// Resposta de sucesso
	response := map[string]interface{}{
		"success": true,
		"message": "Peer removido com sucesso",
	}
	json.NewEncoder(w).Encode(response)
}

// handleGetConfig retorna a configuração atual da VPN
func (h *APIHandler) handleGetConfig(w http.ResponseWriter, r *http.Request) {
	// Ocultar a chave privada por segurança
	configResponse := map[string]interface{}{
		"node_id":       h.config.NodeID,
		"public_key":    h.config.PublicKey,
		"virtual_ip":    h.config.VirtualIP,
		"virtual_cidr":  h.config.VirtualCIDR,
		"interface":     h.config.InterfaceName,
		"mtu":           h.config.MTU,
		"dns":           h.config.DNS,
		"peers_count":   len(h.config.TrustedPeers),
	}

	// Enviar resposta
	json.NewEncoder(w).Encode(configResponse)
}

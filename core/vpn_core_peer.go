package core

import (
	"encoding/base64"
	"fmt"
	"net"
	"strings"
	"time"

	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

// addWireGuardPeer adiciona um peer à configuração do WireGuard
// addWireGuardPeer adds a peer to the WireGuard configuration
// addWireGuardPeer añade un peer a la configuración de WireGuard
func (v *VPNCore) addWireGuardPeer(peer TrustedPeer) error {
	if !v.running {
		return fmt.Errorf("o serviço de VPN não está em execução")
	}

	// Decodificar a chave pública do peer
	peerPublicKeyBytes, err := base64.StdEncoding.DecodeString(peer.PublicKey)
	if err != nil {
		return fmt.Errorf("erro ao decodificar chave pública do peer: %w", err)
	}

	// Criar configuração do peer
	var peerPublicKey wgtypes.Key
	copy(peerPublicKey[:], peerPublicKeyBytes)

	// Definir AllowedIPs (IPs permitidos através deste peer)
	var allowedIPs []net.IPNet
	
	// Se não houver AllowedIPs especificados, usar o IP virtual do peer
	if len(peer.AllowedIPs) == 0 {
		// Usar o IP virtual do peer com máscara /32
		_, ipNet, err := net.ParseCIDR(fmt.Sprintf("%s/32", peer.VirtualIP))
		if err != nil {
			return fmt.Errorf("erro ao analisar IP virtual do peer: %w", err)
		}
		allowedIPs = append(allowedIPs, *ipNet)
	} else {
		// Usar os AllowedIPs fornecidos
		for _, ipStr := range peer.AllowedIPs {
			_, ipNet, err := net.ParseCIDR(ipStr)
			if err != nil {
				fmt.Printf("Aviso: ignorando AllowedIP inválido %s: %v\n", ipStr, err)
				continue
			}
			allowedIPs = append(allowedIPs, *ipNet)
		}
	}

	// Configurar endpoints (se houver)
	var endpoint *net.UDPAddr
	if len(peer.Endpoints) > 0 {
		// Usar o primeiro endpoint disponível
		endpointStr := peer.Endpoints[0]
		if !strings.Contains(endpointStr, ":") {
			// Adicionar a porta padrão se não especificada
			endpointStr = fmt.Sprintf("%s:51820", endpointStr)
		}
		
		endpoint, err = net.ResolveUDPAddr("udp", endpointStr)
		if err != nil {
			fmt.Printf("Aviso: endpoint inválido %s: %v, tentando próximo\n", endpointStr, err)
			endpoint = nil
			
			// Tentar outros endpoints se disponíveis
			for i := 1; i < len(peer.Endpoints) && endpoint == nil; i++ {
				endpointStr = peer.Endpoints[i]
				if !strings.Contains(endpointStr, ":") {
					endpointStr = fmt.Sprintf("%s:51820", endpointStr)
				}
				
				endpoint, err = net.ResolveUDPAddr("udp", endpointStr)
				if err != nil {
					fmt.Printf("Aviso: endpoint alternativo %s inválido: %v\n", endpointStr, err)
					continue
				}
			}
		}
	}

	// Definir configuração do keepalive (se especificado)
	var persistentKeepalive *time.Duration
	if peer.KeepAlive > 0 {
		// Converter de segundos para time.Duration
		keepaliveValue := time.Duration(peer.KeepAlive) * time.Second
		persistentKeepalive = &keepaliveValue
	}

	// Criar configuração do peer
	peerConfig := wgtypes.PeerConfig{
		PublicKey:                   peerPublicKey,
		Endpoint:                    endpoint,
		AllowedIPs:                  allowedIPs,
		PersistentKeepaliveInterval: persistentKeepalive,
	}

	// Atualizar a configuração do dispositivo WireGuard
	deviceConfig := wgtypes.Config{
		Peers: []wgtypes.PeerConfig{peerConfig},
	}

	// Aplicar configuração
	if err := v.wgClient.ConfigureDevice(v.interfaceName, deviceConfig); err != nil {
		return fmt.Errorf("erro ao adicionar peer à interface WireGuard: %w", err)
	}

	fmt.Printf("Peer %s (%s) adicionado com sucesso\n", peer.NodeID, peer.VirtualIP)
	return nil
}

// removeWireGuardPeer remove um peer da configuração do WireGuard
// removeWireGuardPeer removes a peer from the WireGuard configuration
// removeWireGuardPeer elimina un peer de la configuración de WireGuard
func (v *VPNCore) removeWireGuardPeer(peer TrustedPeer) error {
	if !v.running {
		return fmt.Errorf("o serviço de VPN não está em execução")
	}

	// Decodificar a chave pública do peer
	peerPublicKeyBytes, err := base64.StdEncoding.DecodeString(peer.PublicKey)
	if err != nil {
		return fmt.Errorf("erro ao decodificar chave pública do peer: %w", err)
	}

	// Converter para o formato da chave WireGuard
	var peerPublicKey wgtypes.Key
	copy(peerPublicKey[:], peerPublicKeyBytes)

	// Criar configuração para remover o peer
	peerConfig := wgtypes.PeerConfig{
		PublicKey: peerPublicKey,
		Remove:    true,
	}

	// Atualizar a configuração do dispositivo WireGuard
	deviceConfig := wgtypes.Config{
		Peers: []wgtypes.PeerConfig{peerConfig},
	}

	// Aplicar configuração
	if err := v.wgClient.ConfigureDevice(v.interfaceName, deviceConfig); err != nil {
		return fmt.Errorf("erro ao remover peer da interface WireGuard: %w", err)
	}

	fmt.Printf("Peer %s removido com sucesso\n", peer.NodeID)
	return nil
}

// updateWireGuardPeerEndpoint atualiza o endpoint de um peer
// updateWireGuardPeerEndpoint updates a peer's endpoint
// updateWireGuardPeerEndpoint actualiza el endpoint de un peer
func (v *VPNCore) updateWireGuardPeerEndpoint(peer TrustedPeer, endpointStr string) error {
	if !v.running {
		return fmt.Errorf("o serviço de VPN não está em execução")
	}

	// Verificar o formato do endpoint
	if !strings.Contains(endpointStr, ":") {
		// Adicionar a porta padrão se não especificada
		endpointStr = fmt.Sprintf("%s:51820", endpointStr)
	}

	// Resolver o endereço
	endpoint, err := net.ResolveUDPAddr("udp", endpointStr)
	if err != nil {
		return fmt.Errorf("endpoint inválido %s: %w", endpointStr, err)
	}

	// Decodificar a chave pública do peer
	peerPublicKeyBytes, err := base64.StdEncoding.DecodeString(peer.PublicKey)
	if err != nil {
		return fmt.Errorf("erro ao decodificar chave pública do peer: %w", err)
	}

	// Converter para o formato da chave WireGuard
	var peerPublicKey wgtypes.Key
	copy(peerPublicKey[:], peerPublicKeyBytes)

	// Criar configuração para atualizar o peer
	peerConfig := wgtypes.PeerConfig{
		PublicKey: peerPublicKey,
		Endpoint:  endpoint,
	}

	// Atualizar a configuração do dispositivo WireGuard
	deviceConfig := wgtypes.Config{
		Peers: []wgtypes.PeerConfig{peerConfig},
	}

	// Aplicar configuração
	if err := v.wgClient.ConfigureDevice(v.interfaceName, deviceConfig); err != nil {
		return fmt.Errorf("erro ao atualizar endpoint do peer: %w", err)
	}

	fmt.Printf("Endpoint do peer %s atualizado para %s\n", peer.NodeID, endpointStr)
	return nil
}

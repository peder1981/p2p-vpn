package nattraversal

import (
	"fmt"
	"net"
	"strconv"
	"strings"
)

// ParseHostPort analisa uma string de endereço no formato "host:porta"
// ParseHostPort parses an address string in the format "host:port"
// ParseHostPort analiza una cadena de dirección en el formato "host:puerto"
func ParseHostPort(addr string) (string, int, error) {
	// Verificar se contém a porta
	parts := strings.Split(addr, ":")
	if len(parts) != 2 {
		return "", 0, fmt.Errorf("formato de endereço inválido, deve ser host:porta")
	}
	
	// Analisar a porta como número
	port, err := strconv.Atoi(parts[1])
	if err != nil {
		return "", 0, fmt.Errorf("porta inválida: %w", err)
	}
	
	return parts[0], port, nil
}

// ResolveUDPAddr resolve um endereço no formato "host:porta" para um UDPAddr
// ResolveUDPAddr resolves an address in the format "host:port" to a UDPAddr
// ResolveUDPAddr resuelve una dirección en el formato "host:puerto" a un UDPAddr
func ResolveUDPAddr(addr string) (*net.UDPAddr, error) {
	return net.ResolveUDPAddr("udp", addr)
}

// IsPrivateIP verifica se um endereço IP é privado (RFC 1918)
// IsPrivateIP checks if an IP address is private (RFC 1918)
// IsPrivateIP comprueba si una dirección IP es privada (RFC 1918)
func IsPrivateIP(ip net.IP) bool {
	// Verificar se é um IP loopback
	if ip.IsLoopback() {
		return true
	}
	
	// Verificar faixas de IP privado
	// 10.0.0.0/8
	if ip[0] == 10 {
		return true
	}
	
	// 172.16.0.0/12
	if ip[0] == 172 && ip[1] >= 16 && ip[1] <= 31 {
		return true
	}
	
	// 192.168.0.0/16
	if ip[0] == 192 && ip[1] == 168 {
		return true
	}
	
	return false
}

// FindAvailablePort encontra uma porta UDP disponível
// FindAvailablePort finds an available UDP port
// FindAvailablePort encuentra un puerto UDP disponible
func FindAvailablePort() (int, error) {
	// Tentar encontrar uma porta disponível ligando a 0
	addr, err := net.ResolveUDPAddr("udp", "0.0.0.0:0")
	if err != nil {
		return 0, err
	}
	
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return 0, err
	}
	defer conn.Close()
	
	return conn.LocalAddr().(*net.UDPAddr).Port, nil
}

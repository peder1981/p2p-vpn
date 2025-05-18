package nattraversal

import (
	"fmt"
	"net"
	"time"
)

// NATType representa os diferentes tipos de NAT
// NATType represents the different types of NAT
// NATType representa los diferentes tipos de NAT
type NATType int

const (
	NATUnknown NATType = iota
	NATOpen             // Sem NAT, conectividade direta
	NATFullCone         // NAT Full Cone (menos restritivo)
	NATRestrictedCone   // NAT Restricted Cone
	NATPortRestricted   // NAT Port Restricted
	NATSymmetric        // NAT Simétrico (mais restritivo)
)

// String retorna a representação em string do tipo de NAT
func (n NATType) String() string {
	switch n {
	case NATOpen:
		return "Aberto (Sem NAT)"
	case NATFullCone:
		return "NAT Full Cone"
	case NATRestrictedCone:
		return "NAT Restricted Cone"
	case NATPortRestricted:
		return "NAT Port Restricted"
	case NATSymmetric:
		return "NAT Simétrico"
	default:
		return "Desconhecido"
	}
}

// DiagnosticResult contém o resultado da diagnóstico de NAT
// DiagnosticResult contains the result of NAT diagnosis
// DiagnosticResult contiene el resultado del diagnóstico de NAT
type DiagnosticResult struct {
	LocalIP        string    // IP local
	LocalPort      int       // Porta local
	PublicIP       string    // IP público detectado
	PublicPort     int       // Porta pública mapeada
	NATType        NATType   // Tipo de NAT detectado
	STUNServer     string    // Servidor STUN usado
	TestTime       time.Time // Quando o teste foi realizado
	ReachableTests []bool    // Resultados de testes de alcançabilidade
}

// NATDiagnostic implementa diagnóstico de NAT usando STUN
// NATDiagnostic implements NAT diagnosis using STUN
// NATDiagnostic implementa el diagnóstico de NAT usando STUN
type NATDiagnostic struct {
	stunServers []STUNServer
}

// NewNATDiagnostic cria uma nova instância de diagnóstico
// NewNATDiagnostic creates a new diagnostic instance
// NewNATDiagnostic crea una nueva instancia de diagnóstico
func NewNATDiagnostic(stunServers []STUNServer) *NATDiagnostic {
	if len(stunServers) == 0 {
		stunServers = DefaultSTUNServers
	}
	
	return &NATDiagnostic{
		stunServers: stunServers,
	}
}

// RunDiagnosis executa uma série de testes para determinar o tipo de NAT
// RunDiagnosis runs a series of tests to determine the NAT type
// RunDiagnosis ejecuta una serie de pruebas para determinar el tipo de NAT
func (d *NATDiagnostic) RunDiagnosis() (*DiagnosticResult, error) {
	result := &DiagnosticResult{
		TestTime:       time.Now(),
		ReachableTests: make([]bool, 4),
		NATType:        NATUnknown,
	}
	
	// Obter endereço IP local
	localIP, err := getLocalIP()
	if err != nil {
		return nil, fmt.Errorf("erro ao obter IP local: %w", err)
	}
	result.LocalIP = localIP
	
	// Criar socket UDP local
	conn, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv4zero, Port: 0})
	if err != nil {
		return nil, fmt.Errorf("erro ao criar socket UDP: %w", err)
	}
	defer conn.Close()
	
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	result.LocalPort = localAddr.Port
	
	fmt.Printf("Usando socket UDP local %s:%d\n", localIP, result.LocalPort)
	
	// Teste 1: Detectar o endereço IP público e porta usando o primeiro servidor STUN
	if len(d.stunServers) == 0 {
		return nil, fmt.Errorf("nenhum servidor STUN disponível")
	}
	
	stunServer := d.stunServers[0]
	fmt.Printf("Usando servidor STUN primário: %s:%d\n", stunServer.Address, stunServer.Port)
	result.STUNServer = fmt.Sprintf("%s:%d", stunServer.Address, stunServer.Port)
	
	publicIP, publicPort, err := d.detectPublicAddress(conn, stunServer)
	if err != nil {
		fmt.Printf("Erro ao detectar endereço público com servidor STUN primário: %v\n", err)
		
		// Tentar o próximo servidor STUN
		if len(d.stunServers) > 1 {
			stunServer = d.stunServers[1]
			fmt.Printf("Usando servidor STUN alternativo: %s:%d\n", stunServer.Address, stunServer.Port)
			result.STUNServer = fmt.Sprintf("%s:%d", stunServer.Address, stunServer.Port)
			
			publicIP, publicPort, err = d.detectPublicAddress(conn, stunServer)
			if err != nil {
				return nil, fmt.Errorf("erro ao detectar endereço público: %w", err)
			}
		} else {
			return nil, fmt.Errorf("erro ao detectar endereço público: %w", err)
		}
	}
	
	result.PublicIP = publicIP
	result.PublicPort = publicPort
	
	// Se o IP público for igual ao IP local, provavelmente estamos em uma rede com IP público direto
	if publicIP == localIP {
		result.NATType = NATOpen
		fmt.Println("Detectado: Rede com IP público direto (sem NAT)")
		return result, nil
	}
	
	fmt.Printf("IP público detectado: %s:%d\n", publicIP, publicPort)
	
	// Teste 2: Verificar a consistência da porta mapeada com diferentes servidores
	// Este teste diferencia NAT simétrico de outros tipos
	portConsistent := true
	if len(d.stunServers) > 1 {
		stunServer2 := d.stunServers[1]
		_, publicPort2, err := d.detectPublicAddress(conn, stunServer2)
		if err != nil {
			fmt.Printf("Erro no teste de consistência de porta: %v\n", err)
		} else {
			portConsistent = (publicPort == publicPort2)
			fmt.Printf("Teste de consistência de porta: %v (porta1=%d, porta2=%d)\n", 
				portConsistent, publicPort, publicPort2)
		}
	}
	
	// Teste 3: Tentar receber dados de um endpoint não contatado previamente
	// Este teste diferencia Full Cone de Restricted Cone
	canReceiveFromAny := d.testReceiveFromUnknown(conn, d.stunServers)
	result.ReachableTests[0] = canReceiveFromAny
	fmt.Printf("Teste de recebimento de endpoint desconhecido: %v\n", canReceiveFromAny)
	
	// Teste 4: Tentar receber dados da mesma IP mas porta diferente
	// Este teste diferencia Restricted Cone de Port Restricted
	canReceiveFromSameIPDiffPort := d.testReceiveFromSameIPDiffPort(conn, stunServer)
	result.ReachableTests[1] = canReceiveFromSameIPDiffPort
	fmt.Printf("Teste de recebimento de mesma IP, porta diferente: %v\n", canReceiveFromSameIPDiffPort)
	
	// Determinar o tipo de NAT com base nos resultados dos testes
	if !portConsistent {
		result.NATType = NATSymmetric
		fmt.Println("Detectado: NAT Simétrico")
	} else if canReceiveFromAny {
		result.NATType = NATFullCone
		fmt.Println("Detectado: NAT Full Cone")
	} else if canReceiveFromSameIPDiffPort {
		result.NATType = NATRestrictedCone
		fmt.Println("Detectado: NAT Restricted Cone")
	} else {
		result.NATType = NATPortRestricted
		fmt.Println("Detectado: NAT Port Restricted")
	}
	
	return result, nil
}

// detectPublicAddress detecta o endereço IP público e porta usando STUN
func (d *NATDiagnostic) detectPublicAddress(conn *net.UDPConn, stunServer STUNServer) (string, int, error) {
	// Implementação real usaria o protocolo STUN
	// Por ora, usaremos uma versão simplificada
	
	stunAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", stunServer.Address, stunServer.Port))
	if err != nil {
		return "", 0, fmt.Errorf("erro ao resolver endereço STUN: %w", err)
	}
	
	// Enviar solicitação STUN simplificada
	// Na implementação real, usaríamos pacotes STUN formatados corretamente
	_, err = conn.WriteToUDP([]byte("STUN-REQUEST"), stunAddr)
	if err != nil {
		return "", 0, fmt.Errorf("erro ao enviar solicitação STUN: %w", err)
	}
	
	// Configurar timeout para recebimento
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	
	// Receber resposta
	buffer := make([]byte, 1024)
	n, _, err := conn.ReadFromUDP(buffer)
	if err != nil {
		return "", 0, fmt.Errorf("erro ao receber resposta STUN: %w", err)
	}
	
	// Parse da resposta simplificada
	// Em uma implementação real, teríamos que analisar um pacote STUN real
	response := string(buffer[:n])
	
	// Formato simulado: "IP:PORTA"
	// Na implementação real, extrairíamos esses dados do pacote STUN
	var ip string
	var port int
	_, err = fmt.Sscanf(response, "%s:%d", &ip, &port)
	if err != nil {
		return "", 0, fmt.Errorf("erro ao analisar resposta STUN: %w", err)
	}
	
	return ip, port, nil
}

// testReceiveFromUnknown testa se podemos receber pacotes de um endpoint desconhecido
func (d *NATDiagnostic) testReceiveFromUnknown(conn *net.UDPConn, stunServers []STUNServer) bool {
	// Em uma implementação real, solicitaríamos a um servidor STUN
	// que enviasse um pacote de um endpoint desconhecido
	
	// Configurar timeout para recebimento
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	
	// Esperar por um pacote
	buffer := make([]byte, 1024)
	_, _, err := conn.ReadFromUDP(buffer)
	
	return err == nil
}

// testReceiveFromSameIPDiffPort testa se podemos receber de uma IP conhecida mas porta diferente
func (d *NATDiagnostic) testReceiveFromSameIPDiffPort(conn *net.UDPConn, stunServer STUNServer) bool {
	// Em uma implementação real, solicitaríamos ao servidor STUN
	// que enviasse um pacote da mesma IP, mas de uma porta diferente
	
	// Configurar timeout para recebimento
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	
	// Esperar por um pacote
	buffer := make([]byte, 1024)
	_, _, err := conn.ReadFromUDP(buffer)
	
	return err == nil
}

// getLocalIP obtém o endereço IP local preferido
func getLocalIP() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}
	
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String(), nil
			}
		}
	}
	
	return "", fmt.Errorf("não foi possível encontrar um endereço IP local não-loopback")
}

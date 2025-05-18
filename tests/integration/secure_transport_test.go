package integration_test

import (
	"bytes"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/p2p-vpn/p2p-vpn/core"
	"github.com/p2p-vpn/p2p-vpn/security/packets"
)

// mockConn simula uma conexão de rede para testes
// mockConn simulates a network connection for testing
// mockConn simula una conexión de red para pruebas
type mockConn struct {
	readData  []byte
	writeData bytes.Buffer
}

func (m *mockConn) Read(b []byte) (n int, err error) {
	if len(m.readData) == 0 {
		return 0, nil
	}
	n = copy(b, m.readData)
	m.readData = m.readData[n:]
	return n, nil
}

func (m *mockConn) Write(b []byte) (n int, err error) {
	return m.writeData.Write(b)
}

func (m *mockConn) Close() error                       { return nil }
func (m *mockConn) LocalAddr() net.Addr                { return nil }
func (m *mockConn) RemoteAddr() net.Addr               { return nil }
func (m *mockConn) SetDeadline(t time.Time) error      { return nil }
func (m *mockConn) SetReadDeadline(t time.Time) error  { return nil }
func (m *mockConn) SetWriteDeadline(t time.Time) error { return nil }

// TestSecureTransportIntegration testa a integração do transporte seguro com o VPNCore
// TestSecureTransportIntegration tests the integration of secure transport with VPNCore
// TestSecureTransportIntegration prueba la integración del transporte seguro con VPNCore
func TestSecureTransportIntegration(t *testing.T) {
	// Criar diretório temporário para testes
	tempDir := t.TempDir()
	
	// Configuração para os dois nós
	configNode1 := &core.Config{
		NodeID:        "node1",
		InterfaceName: "test1",
		VirtualIP:     "10.0.0.1",
		VirtualCIDR:   "10.0.0.0/24",
		TrustedPeers: []core.TrustedPeer{
			{
				NodeID:    "node2",
				VirtualIP: "10.0.0.2",
				Endpoints: []string{"10.0.0.2:51820"},
			},
		},
	}

	configNode2 := &core.Config{
		NodeID:        "node2",
		InterfaceName: "test2",
		VirtualIP:     "10.0.0.2",
		VirtualCIDR:   "10.0.0.0/24",
		TrustedPeers: []core.TrustedPeer{
			{
				NodeID:    "node1",
				VirtualIP: "10.0.0.1",
				Endpoints: []string{"10.0.0.1:51820"},
			},
		},
	}

	// Salvar configurações em arquivos
	node1ConfigDir := filepath.Join(tempDir, "node1")
	node2ConfigDir := filepath.Join(tempDir, "node2")
	os.MkdirAll(node1ConfigDir, 0755)
	os.MkdirAll(node2ConfigDir, 0755)

	// Criar instâncias VPNCore para testes
	vpnCore1, err := core.NewVPNCore(configNode1, 51820)
	if err != nil {
		t.Fatalf("Erro ao criar VPNCore para node1: %v", err)
	}

	vpnCore2, err := core.NewVPNCore(configNode2, 51821)
	if err != nil {
		t.Fatalf("Erro ao criar VPNCore para node2: %v", err)
	}

	// Criar instâncias de transporte seguro
	secureTransport1, err := core.NewSecureTransport(vpnCore1, node1ConfigDir)
	if err != nil {
		t.Fatalf("Erro ao criar transporte seguro para node1: %v", err)
	}

	secureTransport2, err := core.NewSecureTransport(vpnCore2, node2ConfigDir)
	if err != nil {
		t.Fatalf("Erro ao criar transporte seguro para node2: %v", err)
	}

	// Criar conexões mock
	conn1to2 := &mockConn{}
	conn2to1 := &mockConn{}

	// Registrar conexões
	secureTransport1.RegisterPeerConnection("node2", conn1to2)
	secureTransport2.RegisterPeerConnection("node1", conn2to1)

	// Testar handshake node1 -> node2
	err = secureTransport1.InitiateHandshake("node2")
	if err != nil {
		t.Fatalf("Erro ao iniciar handshake node1 -> node2: %v", err)
	}

	// Processar handshake em node2 (extraindo dados da conexão mock)
	handshakeData := conn1to2.writeData.Bytes()
	err = secureTransport2.HandleIncomingPacket(handshakeData)
	if err != nil {
		t.Fatalf("Erro ao processar handshake em node2: %v", err)
	}

	// Testar handshake node2 -> node1
	err = secureTransport2.InitiateHandshake("node1")
	if err != nil {
		t.Fatalf("Erro ao iniciar handshake node2 -> node1: %v", err)
	}

	// Processar handshake em node1
	handshakeData = conn2to1.writeData.Bytes()
	err = secureTransport1.HandleIncomingPacket(handshakeData)
	if err != nil {
		t.Fatalf("Erro ao processar handshake em node1: %v", err)
	}

	// Testar envio de dados node1 -> node2
	testData := []byte("Teste de dados assinados")
	conn1to2.writeData.Reset() // Limpar dados anteriores
	
	err = secureTransport1.SendSecurePacket("node2", testData, packets.PacketTypeData)
	if err != nil {
		t.Fatalf("Erro ao enviar dados de node1 para node2: %v", err)
	}

	// Verificar se os dados foram enviados
	if conn1to2.writeData.Len() == 0 {
		t.Fatal("Nenhum dado foi enviado")
	}

	// Mock para ProcessIncomingData do VPNCore
	dataProcessed := false
	vpnCore2.ProcessIncomingData = func(peerID string, data []byte) error {
		if peerID == "node1" && bytes.Equal(data, testData) {
			dataProcessed = true
		}
		return nil
	}

	// Processar dados em node2
	dataPacket := conn1to2.writeData.Bytes()
	err = secureTransport2.HandleIncomingPacket(dataPacket)
	if err != nil {
		t.Fatalf("Erro ao processar dados em node2: %v", err)
	}

	// Verificar se os dados foram processados corretamente
	if !dataProcessed {
		t.Fatal("Dados não foram processados corretamente")
	}
}

// TestSecureHandshakeProtocol testa o protocolo de handshake seguro
// TestSecureHandshakeProtocol tests the secure handshake protocol
// TestSecureHandshakeProtocol prueba el protocolo de handshake seguro
func TestSecureHandshakeProtocol(t *testing.T) {
	// Criar diretório temporário para testes
	tempDir := t.TempDir()
	
	// Criar gerenciadores de pacotes seguros para dois nós
	handler1, err := core.NewSecurePacketHandler(filepath.Join(tempDir, "node1"), "node1")
	if err != nil {
		t.Fatalf("Erro ao criar gerenciador de pacotes para node1: %v", err)
	}

	handler2, err := core.NewSecurePacketHandler(filepath.Join(tempDir, "node2"), "node2")
	if err != nil {
		t.Fatalf("Erro ao criar gerenciador de pacotes para node2: %v", err)
	}

	// Obter chave pública do node1 para enviá-la ao node2
	pubKey1, err := handler1.GetPublicKeyPEM()
	if err != nil {
		t.Fatalf("Erro ao obter chave pública de node1: %v", err)
	}

	// Assinar a chave pública com a chave privada do node1
	signedPubKey1, err := handler1.SignPacket("node2", packets.PacketTypeHandshake, pubKey1)
	if err != nil {
		t.Fatalf("Erro ao assinar chave pública de node1: %v", err)
	}

	// Obter chave pública do node2 para enviá-la ao node1
	pubKey2, err := handler2.GetPublicKeyPEM()
	if err != nil {
		t.Fatalf("Erro ao obter chave pública de node2: %v", err)
	}

	// Node1 adiciona a chave pública do Node2
	err = handler1.AddPeerPublicKey("node2", pubKey2)
	if err != nil {
		t.Fatalf("Erro ao adicionar chave pública do node2 ao node1: %v", err)
	}

	// Node2 adiciona a chave pública do Node1
	err = handler2.AddPeerPublicKey("node1", pubKey1)
	if err != nil {
		t.Fatalf("Erro ao adicionar chave pública do node1 ao node2: %v", err)
	}

	// Node2 verifica a chave pública assinada recebida do node1
	senderID, payload, packetType, err := handler2.VerifyAndExtractPacket(signedPubKey1)
	if err != nil {
		t.Fatalf("Erro ao verificar pacote assinado: %v", err)
	}

	// Verificar resultados
	if senderID != "node1" {
		t.Errorf("ID do remetente incorreto: esperado node1, obtido %s", senderID)
	}

	if packetType != packets.PacketTypeHandshake {
		t.Errorf("Tipo de pacote incorreto: esperado %d, obtido %d", packets.PacketTypeHandshake, packetType)
	}

	if !bytes.Equal(payload, pubKey1) {
		t.Error("Conteúdo da chave pública não corresponde após verificação")
	}

	// Testar assinatura e verificação de dados regulares
	testData := []byte("Dados de teste para verificar assinatura digital")
	
	// Node1 assina dados para Node2
	signedData, err := handler1.SignPacket("node2", packets.PacketTypeData, testData)
	if err != nil {
		t.Fatalf("Erro ao assinar dados de teste: %v", err)
	}

	// Node2 verifica e extrai dados
	senderID, extractedData, packetType, err := handler2.VerifyAndExtractPacket(signedData)
	if err != nil {
		t.Fatalf("Erro ao verificar dados assinados: %v", err)
	}

	// Verificar resultados
	if senderID != "node1" {
		t.Errorf("ID do remetente incorreto: esperado node1, obtido %s", senderID)
	}

	if packetType != packets.PacketTypeData {
		t.Errorf("Tipo de pacote incorreto: esperado %d, obtido %d", packets.PacketTypeData, packetType)
	}

	if !bytes.Equal(extractedData, testData) {
		t.Error("Dados extraídos não correspondem aos dados originais")
	}
}

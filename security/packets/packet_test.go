package packets

import (
	"crypto/ed25519"
	"testing"
	"time"
)

// TestPacketSigningAndVerification testa o processo de assinatura e verificação de pacotes
// TestPacketSigningAndVerification tests the packet signing and verification process
// TestPacketSigningAndVerification prueba el proceso de firma y verificación de paquetes
func TestPacketSigningAndVerification(t *testing.T) {
	// Gerar par de chaves para teste
	publicKey, privateKey, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatalf("Erro ao gerar par de chaves: %v", err)
	}

	// Dados para o pacote de teste
	senderID := "node123"
	recipientID := "node456"
	payload := []byte("Dados de teste do pacote P2P VPN")
	packetType := PacketTypeData

	// Criar pacote assinado
	packet, err := NewSignedPacket(senderID, recipientID, packetType, payload, privateKey)
	if err != nil {
		t.Fatalf("Erro ao criar pacote assinado: %v", err)
	}

	// Verificar assinatura com a chave pública correta
	if err := packet.Verify(publicKey); err != nil {
		t.Errorf("Falha na verificação de assinatura válida: %v", err)
	}

	// Teste com chave pública incorreta
	_, incorrectKey, _ := ed25519.GenerateKey(nil)
	if err := packet.Verify(incorrectKey); err != ErrInvalidSignature {
		t.Errorf("Verificação com chave incorreta deveria falhar com ErrInvalidSignature, recebeu: %v", err)
	}

	// Teste de expiração de pacote
	// Forçar timestamp antigo
	packet.Header.Timestamp = time.Now().Unix() - (MaxPacketAge + 10)
	if err := packet.Verify(publicKey); err != ErrExpiredPacket {
		t.Errorf("Verificação de pacote expirado deveria falhar com ErrExpiredPacket, recebeu: %v", err)
	}
}

// TestPacketEncodeAndDecode testa a serialização e desserialização de pacotes
// TestPacketEncodeAndDecode tests packet serialization and deserialization
// TestPacketEncodeAndDecode prueba la serialización y deserialización de paquetes
func TestPacketEncodeAndDecode(t *testing.T) {
	// Gerar par de chaves para teste
	_, privateKey, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatalf("Erro ao gerar par de chaves: %v", err)
	}

	// Criar pacote original
	original, err := NewSignedPacket(
		"origem",
		"destino",
		PacketTypeHandshake,
		[]byte("Conteúdo do pacote"),
		privateKey,
	)
	if err != nil {
		t.Fatalf("Erro ao criar pacote: %v", err)
	}

	// Codificar
	encoded, err := original.Encode()
	if err != nil {
		t.Fatalf("Erro ao codificar pacote: %v", err)
	}

	// Decodificar
	decoded, err := Decode(encoded)
	if err != nil {
		t.Fatalf("Erro ao decodificar pacote: %v", err)
	}

	// Verificar se os dados essenciais são preservados
	if decoded.Header.SenderID != original.Header.SenderID {
		t.Errorf("SenderID não preservado: esperado %s, obtido %s", 
			original.Header.SenderID, decoded.Header.SenderID)
	}

	if decoded.Header.RecipientID != original.Header.RecipientID {
		t.Errorf("RecipientID não preservado: esperado %s, obtido %s", 
			original.Header.RecipientID, decoded.Header.RecipientID)
	}

	if decoded.Header.PacketType != original.Header.PacketType {
		t.Errorf("PacketType não preservado: esperado %d, obtido %d", 
			original.Header.PacketType, decoded.Header.PacketType)
	}

	if len(decoded.Signature) != len(original.Signature) {
		t.Errorf("Tamanho da assinatura não preservado: esperado %d, obtido %d", 
			len(original.Signature), len(decoded.Signature))
	}

	if string(decoded.Payload) != string(original.Payload) {
		t.Errorf("Payload não preservado: esperado %s, obtido %s", 
			string(original.Payload), string(decoded.Payload))
	}
}

// TestKeyManager testa as funcionalidades do gerenciador de chaves
// TestKeyManager tests the key manager functionalities
// TestKeyManager prueba las funcionalidades del gestor de claves
func TestKeyManager(t *testing.T) {
	// Criar diretório temporário para os testes
	tempDir := t.TempDir()

	// Criar gerenciador de chaves
	km, err := NewKeyManager(tempDir)
	if err != nil {
		t.Fatalf("Erro ao criar gerenciador de chaves: %v", err)
	}

	// Verificar se as chaves foram geradas
	privateKey := km.GetPrivateKey()
	publicKey := km.GetPublicKey()

	if privateKey == nil || publicKey == nil {
		t.Fatal("Chaves não foram geradas corretamente")
	}

	// Testar armazenamento de chaves de peers
	peerID := "peer123"
	_, peerKey, _ := ed25519.GenerateKey(nil)

	// Adicionar chave do peer
	km.AddPeerKey(peerID, peerKey)

	// Recuperar e verificar
	retrievedKey, err := km.GetPeerKey(peerID)
	if err != nil {
		t.Fatalf("Erro ao recuperar chave do peer: %v", err)
	}

	if string(retrievedKey) != string(peerKey) {
		t.Error("Chave recuperada não corresponde à chave armazenada")
	}

	// Testar armazenamento em arquivo
	err = km.StorePeerKey(peerID, peerKey)
	if err != nil {
		t.Fatalf("Erro ao armazenar chave do peer: %v", err)
	}

	// Criar novo gerenciador de chaves para testar carregamento
	km2, err := NewKeyManager(tempDir)
	if err != nil {
		t.Fatalf("Erro ao criar segundo gerenciador de chaves: %v", err)
	}

	// Carregar chaves de peers
	err = km2.LoadPeerKeys()
	if err != nil {
		t.Fatalf("Erro ao carregar chaves de peers: %v", err)
	}

	// Verificar se a chave do peer foi carregada corretamente
	retrievedKey2, err := km2.GetPeerKey(peerID)
	if err != nil {
		t.Fatalf("Erro ao recuperar chave do peer do segundo gerenciador: %v", err)
	}

	if string(retrievedKey2) != string(peerKey) {
		t.Error("Chave recuperada do segundo gerenciador não corresponde à original")
	}
}

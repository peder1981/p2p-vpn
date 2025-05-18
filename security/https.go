package security

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"os"
	"time"
)

// TLSConfig contém a configuração para HTTPS
// TLSConfig contains the configuration for HTTPS
// TLSConfig contiene la configuración para HTTPS
type TLSConfig struct {
	CertFile string // Caminho para o arquivo do certificado
	KeyFile  string // Caminho para o arquivo da chave privada
	SelfSign bool   // Gerar certificado autoassinado se os arquivos não existirem
}

// GenerateTLSConfig gera a configuração TLS para o servidor HTTP
// GenerateTLSConfig generates the TLS configuration for the HTTP server
// GenerateTLSConfig genera la configuración TLS para el servidor HTTP
func GenerateTLSConfig(config TLSConfig) (*tls.Config, error) {
	// Verificar se os arquivos existem
	_, certErr := os.Stat(config.CertFile)
	_, keyErr := os.Stat(config.KeyFile)
	
	// Se ambos existem, carregá-los
	if certErr == nil && keyErr == nil {
		cert, err := tls.LoadX509KeyPair(config.CertFile, config.KeyFile)
		if err != nil {
			return nil, fmt.Errorf("erro ao carregar certificado e chave: %w", err)
		}
		
		return &tls.Config{
			Certificates: []tls.Certificate{cert},
			MinVersion:   tls.VersionTLS12,
		}, nil
	}
	
	// Se falta algum dos arquivos e SelfSign está habilitado
	if config.SelfSign {
		// Gerar certificado autoassinado
		cert, key, err := generateSelfSignedCert()
		if err != nil {
			return nil, fmt.Errorf("erro ao gerar certificado autoassinado: %w", err)
		}
		
		// Salvar certificado e chave
		if err := saveCertificateFiles(config.CertFile, config.KeyFile, cert, key); err != nil {
			return nil, fmt.Errorf("erro ao salvar certificado autoassinado: %w", err)
		}
		
		// Carregar certificado e chave
		tlsCert, err := tls.X509KeyPair(cert, key)
		if err != nil {
			return nil, fmt.Errorf("erro ao carregar certificado gerado: %w", err)
		}
		
		return &tls.Config{
			Certificates: []tls.Certificate{tlsCert},
			MinVersion:   tls.VersionTLS12,
		}, nil
	}
	
	// Se os arquivos não existem e não estamos autoassinando
	return nil, fmt.Errorf("arquivos de certificado e/ou chave não encontrados: %s, %s", config.CertFile, config.KeyFile)
}

// generateSelfSignedCert gera um certificado autoassinado
func generateSelfSignedCert() ([]byte, []byte, error) {
	// Gerar chave privada
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, fmt.Errorf("erro ao gerar chave privada: %w", err)
	}
	
	// Definir template para o certificado
	notBefore := time.Now()
	notAfter := notBefore.Add(365 * 24 * time.Hour) // Validade de 1 ano
	
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return nil, nil, fmt.Errorf("erro ao gerar número serial: %w", err)
	}
	
	// Buscar nomes locais da máquina
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "localhost"
	}
	
	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"P2P VPN Universal"},
			CommonName:   hostname,
		},
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}
	
	// Adicionar nomes alternativos ao certificado
	template.DNSNames = append(template.DNSNames, hostname)
	template.DNSNames = append(template.DNSNames, "localhost")
	
	// Adicionar IPs locais
	addrs, err := net.InterfaceAddrs()
	if err == nil {
		for _, addr := range addrs {
			if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
				if ipnet.IP.To4() != nil {
					template.IPAddresses = append(template.IPAddresses, ipnet.IP)
				}
			}
		}
	}
	
	// Sempre adicionar o IP de loopback
	template.IPAddresses = append(template.IPAddresses, net.ParseIP("127.0.0.1"))
	template.IPAddresses = append(template.IPAddresses, net.ParseIP("::1"))
	
	// Criar certificado
	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		return nil, nil, fmt.Errorf("erro ao criar certificado: %w", err)
	}
	
	// Codificar certificado em PEM
	certOut := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	
	// Codificar chave privada em PEM
	keyOut := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)})
	
	return certOut, keyOut, nil
}

// saveCertificateFiles salva os arquivos de certificado e chave no disco
func saveCertificateFiles(certFile, keyFile string, cert, key []byte) error {
	// Criar ou substituir arquivo de certificado
	err := os.WriteFile(certFile, cert, 0600)
	if err != nil {
		return fmt.Errorf("erro ao salvar arquivo de certificado: %w", err)
	}
	
	// Criar ou substituir arquivo de chave
	err = os.WriteFile(keyFile, key, 0600)
	if err != nil {
		return fmt.Errorf("erro ao salvar arquivo de chave: %w", err)
	}
	
	return nil
}

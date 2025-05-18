package security

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

// SecurityConfig representa a configuração completa de segurança
// SecurityConfig represents the complete security configuration
// SecurityConfig representa la configuración completa de seguridad
type SecurityConfig struct {
	Web struct {
		ListenAddr string `yaml:"listen_addr"` // Endereço para escutar
		HTTPS struct {
			Enabled  bool   `yaml:"enabled"`    // Habilitar HTTPS
			CertFile string `yaml:"cert_file"`  // Caminho para o certificado
			KeyFile  string `yaml:"key_file"`   // Caminho para a chave privada
			SelfSign bool   `yaml:"self_sign"`  // Gerar certificado autoassinado
		} `yaml:"https"`
		Auth struct {
			JWTExpiration int    `yaml:"jwt_expiration"` // Tempo de expiração do token em horas
			JWTSecret     string `yaml:"jwt_secret"`     // Segredo para assinatura JWT
			UsersFile     string `yaml:"users_file"`     // Arquivo para armazenar usuários
			DefaultAdmin struct {
				Username string `yaml:"username"` // Nome do usuário admin padrão
				Password string `yaml:"password"` // Senha do usuário admin padrão
			} `yaml:"default_admin"`
		} `yaml:"auth"`
	} `yaml:"web"`
	VPN struct {
		Encryption struct {
			Curve string `yaml:"curve"` // Curva de criptografia
		} `yaml:"encryption"`
		Access struct {
			TrustedIPs          []string `yaml:"trusted_ips"`           // IPs confiáveis
			AuthenticatedOnly   bool     `yaml:"authenticated_peers_only"` // Apenas peers autenticados
			VerifyPeerKeys      bool     `yaml:"verify_peer_keys"`      // Verificar chaves de peers
		} `yaml:"access"`
	} `yaml:"vpn"`
	Audit struct {
		Enabled     bool   `yaml:"enabled"`     // Habilitar auditoria
		Level       string `yaml:"level"`       // Nível de log
		LogFile     string `yaml:"log_file"`    // Arquivo de log
		LogRotation int    `yaml:"log_rotation"` // Rotação de logs em dias
		Events struct {
			Auth  bool `yaml:"auth"`  // Eventos de autenticação
			VPN   bool `yaml:"vpn"`   // Eventos de VPN
			Admin bool `yaml:"admin"` // Eventos de administração
			API   bool `yaml:"api"`   // Eventos de API
		} `yaml:"events"`
	} `yaml:"audit"`
}

// LoadSecurityConfig carrega a configuração de segurança de um arquivo YAML
// LoadSecurityConfig loads the security configuration from a YAML file
// LoadSecurityConfig carga la configuración de seguridad desde un archivo YAML
func LoadSecurityConfig(configPath string) (*SecurityConfig, error) {
	// Verificar se o arquivo existe
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("arquivo de configuração não encontrado: %s", configPath)
	}

	// Ler o conteúdo do arquivo
	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("erro ao ler arquivo de configuração: %w", err)
	}

	// Decodificar o YAML
	var config SecurityConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("erro ao decodificar configuração: %w", err)
	}

	// Validar configuração
	if err := validateSecurityConfig(&config); err != nil {
		return nil, fmt.Errorf("configuração inválida: %w", err)
	}

	return &config, nil
}

// validateSecurityConfig valida e preenche valores padrão para a configuração
// validateSecurityConfig validates and fills default values for the configuration
// validateSecurityConfig valida y rellena valores predeterminados para la configuración
func validateSecurityConfig(config *SecurityConfig) error {
	// Endereço para escutar
	if config.Web.ListenAddr == "" {
		config.Web.ListenAddr = "localhost:8080"
	}

	// Configuração HTTPS
	if config.Web.HTTPS.Enabled {
		// Verificar caminhos de certificados
		if config.Web.HTTPS.CertFile == "" {
			config.Web.HTTPS.CertFile = "certs/server.crt"
		}
		if config.Web.HTTPS.KeyFile == "" {
			config.Web.HTTPS.KeyFile = "certs/server.key"
		}

		// Criar diretório de certificados se necessário
		certDir := filepath.Dir(config.Web.HTTPS.CertFile)
		if err := os.MkdirAll(certDir, 0755); err != nil {
			return fmt.Errorf("erro ao criar diretório de certificados: %w", err)
		}
	}

	// Configuração JWT
	if config.Web.Auth.JWTExpiration <= 0 {
		config.Web.Auth.JWTExpiration = 24 // 24 horas por padrão
	}

	// Arquivo de usuários
	if config.Web.Auth.UsersFile == "" {
		config.Web.Auth.UsersFile = "data/users.json"
	}

	// Criar diretório para o arquivo de usuários
	usersDir := filepath.Dir(config.Web.Auth.UsersFile)
	if err := os.MkdirAll(usersDir, 0755); err != nil {
		return fmt.Errorf("erro ao criar diretório para usuários: %w", err)
	}

	// Configuração de auditoria
	if config.Audit.Enabled {
		if config.Audit.Level == "" {
			config.Audit.Level = "info"
		}
		if config.Audit.LogFile == "" {
			config.Audit.LogFile = "logs/security.log"
		}

		// Criar diretório de logs
		logDir := filepath.Dir(config.Audit.LogFile)
		if err := os.MkdirAll(logDir, 0755); err != nil {
			return fmt.Errorf("erro ao criar diretório de logs: %w", err)
		}
	}

	return nil
}

// ToAuthConfig converte a configuração para AuthConfig
// ToAuthConfig converts the configuration to AuthConfig
// ToAuthConfig convierte la configuración a AuthConfig
func (c *SecurityConfig) ToAuthConfig() AuthConfig {
	return AuthConfig{
		JWTSecret:        c.Web.Auth.JWTSecret,
		JWTExpiration:    time.Duration(c.Web.Auth.JWTExpiration) * time.Hour,
		DefaultAdminUser: c.Web.Auth.DefaultAdmin.Username,
		DefaultAdminPass: c.Web.Auth.DefaultAdmin.Password,
		UsersFile:        c.Web.Auth.UsersFile,
	}
}

// ToJWTConfig converte a configuração para JWTConfig
// ToJWTConfig converts the configuration to JWTConfig
// ToJWTConfig convierte la configuración a JWTConfig
func (c *SecurityConfig) ToJWTConfig() JWTConfig {
	return JWTConfig{
		Secret:     c.Web.Auth.JWTSecret,
		Expiration: time.Duration(c.Web.Auth.JWTExpiration) * time.Hour,
	}
}

// ToTLSConfig converte a configuração para TLSConfig
// ToTLSConfig converts the configuration to TLSConfig
// ToTLSConfig convierte la configuración a TLSConfig
func (c *SecurityConfig) ToTLSConfig() TLSConfig {
	return TLSConfig{
		CertFile: c.Web.HTTPS.CertFile,
		KeyFile:  c.Web.HTTPS.KeyFile,
		SelfSign: c.Web.HTTPS.SelfSign,
	}
}

// ToWebConfig converte a configuração para web.Config
// ToWebConfig converts the configuration to web.Config
// ToWebConfig convierte la configuración a web.Config
func (c *SecurityConfig) ToWebConfig() map[string]interface{} {
	return map[string]interface{}{
		"ListenAddr":       c.Web.ListenAddr,
		"UseHTTPS":         c.Web.HTTPS.Enabled,
		"TLSConfig":        c.ToTLSConfig(),
		"JWTSecret":        c.Web.Auth.JWTSecret,
		"JWTExpiration":    time.Duration(c.Web.Auth.JWTExpiration) * time.Hour,
		"DefaultAdminUser": c.Web.Auth.DefaultAdmin.Username,
		"DefaultAdminPass": c.Web.Auth.DefaultAdmin.Password,
		"UsersFile":        c.Web.Auth.UsersFile,
	}
}

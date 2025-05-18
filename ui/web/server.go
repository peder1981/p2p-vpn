package web

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"net/http"
	"path/filepath"
	"time"

	"github.com/p2p-vpn/p2p-vpn/core"
	"github.com/p2p-vpn/p2p-vpn/security"
)

//go:embed static
var staticFiles embed.FS

// Config contém a configuração para o servidor web
// Config contains the configuration for the web server
// Config contiene la configuración para el servidor web
type Config struct {
	ListenAddr     string           // Endereço para escutar (ex: localhost:8080)
	CoreVPN        core.VPNProvider // Referência para o core da VPN
	Config         *core.Config     // Configuração geral
	TLSConfig      security.TLSConfig // Configuração TLS para HTTPS
	JWTSecret      string          // Segredo para JWT (opcional, será gerado aleatoriamente se vazio)
	JWTExpiration  time.Duration   // Tempo de expiração do token JWT (padrão: 24h)
	DefaultAdminUser string        // Nome de usuário padrão para admin (opcional)
	DefaultAdminPass string        // Senha padrão para admin (opcional)
	UsersFile      string          // Arquivo para armazenar dados de usuários
	UseHTTPS       bool            // Usar HTTPS em vez de HTTP
}

// StartServer inicia o servidor web para a interface de usuário
// StartServer starts the web server for the user interface
// StartServer inicia el servidor web para la interfaz de usuario
func StartServer(config Config) error {
	if config.ListenAddr == "" {
		config.ListenAddr = "localhost:8080"
	}

	// Configurar o serviço de autenticação
	authConfig := security.AuthConfig{
		JWTSecret:        config.JWTSecret,
		JWTExpiration:    config.JWTExpiration,
		DefaultAdminUser: config.DefaultAdminUser,
		DefaultAdminPass: config.DefaultAdminPass,
		UsersFile:        filepath.Join(filepath.Dir(config.UsersFile), "users.json"),
	}

	authService, err := security.NewAuthService(authConfig)
	if err != nil {
		return fmt.Errorf("erro ao inicializar serviço de autenticação: %w", err)
	}

	// Configurar JWT
	jwtConfig := security.JWTConfig{
		Secret:     authConfig.JWTSecret,
		Expiration: authConfig.JWTExpiration,
	}

	// Configurar o manipulador de arquivos estáticos
	staticFS, err := fs.Sub(staticFiles, "static")
	if err != nil {
		return fmt.Errorf("erro ao configurar sistema de arquivos estáticos: %w", err)
	}
	
	// Criar o roteador HTTP
	mux := http.NewServeMux()
	
	// Configurar middleware de autenticação
	authMiddleware := &security.AuthMiddleware{
		JWTConfig:     jwtConfig,
		AuthService:   authService,
		AllowedPaths:  []string{
			"/",              // Página inicial
			"/index.html",     // Página inicial
			"/css/",          // Recursos CSS
			"/js/",           // Recursos JavaScript
			"/img/",          // Imagens
			"/favicon.ico",    // Favicon
			"/api/auth/login", // Endpoint de login
		},
		RequiredPerms: map[string]security.Permission{
			"/api/status":         security.PermReadOnly,
			"/api/peers":          security.PermReadOnly,
			"/api/config":         security.PermReadOnly,
			"/api/peers/add":      security.PermReadWrite,
			"/api/peers/remove":   security.PermReadWrite,
			"/api/users":          security.PermAdmin,
			"/api/auth/register":  security.PermAdmin,
		},
	}
	
	// Arquivos estáticos (sem necessidade de autenticação)
	fileServer := http.FileServer(http.FS(staticFS))
	mux.Handle("/", fileServer)
	
	// Endpoints de autenticação
	mux.Handle("/api/auth/login", security.LoginHandler(authService, jwtConfig))
	mux.Handle("/api/auth/logout", security.LogoutHandler())
	mux.Handle("/api/auth/register", authMiddleware.Middleware(security.RegisterHandler(authService)))
	
	// API para gerenciamento da VPN (protegida por autenticação)
	apiHandler := NewAPIHandler(config.CoreVPN, config.Config)
	mux.Handle("/api/", authMiddleware.Middleware(apiHandler))
	
	// Criar servidor com timeout
	server := &http.Server{
		Addr:         config.ListenAddr,
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Verificar se deve usar HTTPS
	if config.UseHTTPS {
		// Gerar configuração TLS
		tlsConfig, err := security.GenerateTLSConfig(config.TLSConfig)
		if err != nil {
			return fmt.Errorf("erro ao configurar TLS: %w", err)
		}
		
		server.TLSConfig = tlsConfig
		
		// Iniciar servidor HTTPS
		fmt.Printf("Interface web iniciada em https://%s\n", config.ListenAddr)
		fmt.Println("AVISO: Este servidor usa um certificado autoassinado. Você pode receber avisos de segurança no navegador.")
		return server.ListenAndServeTLS(config.TLSConfig.CertFile, config.TLSConfig.KeyFile)
	}
	
	// Iniciar servidor HTTP
	fmt.Printf("Interface web iniciada em http://%s\n", config.ListenAddr)
	fmt.Println("AVISO: Esta conexão não é segura. Para maior segurança, habilite HTTPS.")
	return server.ListenAndServe()
}

// Gracefully encerra o servidor
// Gracefully stops the server
// Detiene el servidor de forma ordenada
func GracefulShutdown(server *http.Server) {
	// Criar contexto com timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	// Tentar encerrar o servidor graciosamente
	if err := server.Shutdown(ctx); err != nil {
		fmt.Printf("Erro ao encerrar servidor web: %v\n", err)
	}
}

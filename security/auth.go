package security

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"sync"
	"time"

	"golang.org/x/crypto/argon2"
)

// Configuração do hash Argon2
const (
	argonTime    = 1
	argonMemory  = 64 * 1024
	argonThreads = 4
	argonKeyLen  = 32
)

// Erros comuns
var (
	ErrInvalidCredentials = errors.New("credenciais inválidas")
	ErrUserAlreadyExists  = errors.New("usuário já existe")
	ErrUserNotFound       = errors.New("usuário não encontrado")
	ErrInvalidToken       = errors.New("token inválido ou expirado")
)

// Tipos de autenticação
type AuthMethod string

const (
	AuthLocal   AuthMethod = "local"  // Autenticação local (usuário/senha)
	AuthKey     AuthMethod = "apikey" // Autenticação por API key
	AuthToken   AuthMethod = "token"  // Autenticação por JWT token
)

// Permissões
type Permission string

const (
	PermReadOnly     Permission = "read"      // Somente leitura
	PermReadWrite    Permission = "write"     // Leitura e escrita
	PermAdmin        Permission = "admin"     // Acesso administrativo completo
)

// User representa um usuário do sistema
// User represents a system user
// User representa un usuario del sistema
type User struct {
	Username     string      `json:"username"`
	PasswordHash string      `json:"password_hash,omitempty"`
	Salt         string      `json:"salt,omitempty"`
	APIKeys      []APIKey    `json:"api_keys,omitempty"`
	Permissions  []Permission `json:"permissions"`
	LastLogin    time.Time   `json:"last_login,omitempty"`
	CreatedAt    time.Time   `json:"created_at"`
}

// APIKey representa uma chave de API para autenticação
// APIKey represents an API key for authentication
// APIKey representa una clave de API para autenticación
type APIKey struct {
	Name      string    `json:"name"`
	Key       string    `json:"key,omitempty"` // Hash da chave
	CreatedAt time.Time `json:"created_at"`
	LastUsed  time.Time `json:"last_used,omitempty"`
	ExpiresAt time.Time `json:"expires_at,omitempty"`
}

// AuthConfig contém a configuração para o sistema de autenticação
// AuthConfig contains the configuration for the authentication system
// AuthConfig contiene la configuración para el sistema de autenticación
type AuthConfig struct {
	JWTSecret        string        `json:"jwt_secret"`
	JWTExpiration    time.Duration `json:"jwt_expiration"`
	DefaultAdminUser string        `json:"default_admin_user"`
	DefaultAdminPass string        `json:"default_admin_pass"`
	UsersFile        string        `json:"users_file"`
}

// AuthService gerencia autenticação de usuários e autorização
// AuthService manages user authentication and authorization
// AuthService gestiona la autenticación y autorización de usuarios
type AuthService struct {
	config      AuthConfig
	users       map[string]User
	usersMutex  sync.RWMutex
	initialized bool
}

// NewAuthService cria um novo serviço de autenticação
// NewAuthService creates a new authentication service
// NewAuthService crea un nuevo servicio de autenticación
func NewAuthService(config AuthConfig) (*AuthService, error) {
	if config.JWTSecret == "" {
		// Gerar um segredo aleatório para JWT se não fornecido
		secret := make([]byte, 32)
		if _, err := rand.Read(secret); err != nil {
			return nil, fmt.Errorf("erro ao gerar segredo JWT: %w", err)
		}
		config.JWTSecret = base64.StdEncoding.EncodeToString(secret)
	}

	if config.JWTExpiration == 0 {
		// Padrão: 24 horas
		config.JWTExpiration = 24 * time.Hour
	}

	if config.UsersFile == "" {
		config.UsersFile = "users.json"
	}

	service := &AuthService{
		config: config,
		users: make(map[string]User),
	}

	// Carregar usuários existentes ou criar admin padrão
	if err := service.loadUsers(); err != nil {
		return nil, err
	}

	return service, nil
}

// loadUsers carrega os usuários do arquivo ou cria o admin padrão
func (s *AuthService) loadUsers() error {
	s.usersMutex.Lock()
	defer s.usersMutex.Unlock()

	// Verificar se o arquivo de usuários existe
	if _, err := os.Stat(s.config.UsersFile); os.IsNotExist(err) {
		// Criar usuário admin padrão se configurado
		if s.config.DefaultAdminUser != "" && s.config.DefaultAdminPass != "" {
			admin := User{
				Username:    s.config.DefaultAdminUser,
				Permissions: []Permission{PermAdmin},
				CreatedAt:   time.Now(),
			}
			
			// Gerar hash da senha
			hash, salt, err := s.hashPassword(s.config.DefaultAdminPass)
			if err != nil {
				return fmt.Errorf("erro ao criar hash da senha: %w", err)
			}
			
			admin.PasswordHash = hash
			admin.Salt = salt
			s.users[admin.Username] = admin
			
			// Salvar o usuário admin
			return s.saveUsers()
		}
		
		// Se não tem admin padrão, apenas inicializa com lista vazia
		return s.saveUsers()
	}
	
	// Ler o arquivo
	data, err := ioutil.ReadFile(s.config.UsersFile)
	if err != nil {
		return fmt.Errorf("erro ao ler arquivo de usuários: %w", err)
	}
	
	// Decodificar JSON
	var users []User
	if err := json.Unmarshal(data, &users); err != nil {
		return fmt.Errorf("erro ao decodificar usuários: %w", err)
	}
	
	// Mapear usuários por nome
	for _, user := range users {
		s.users[user.Username] = user
	}
	
	s.initialized = true
	return nil
}

// saveUsers salva os usuários no arquivo
func (s *AuthService) saveUsers() error {
	s.usersMutex.RLock()
	defer s.usersMutex.RUnlock()
	
	// Converter mapa para slice
	users := make([]User, 0, len(s.users))
	for _, user := range s.users {
		// Omitir campos sensíveis ao salvar
		users = append(users, user)
	}
	
	// Codificar como JSON
	data, err := json.MarshalIndent(users, "", "  ")
	if err != nil {
		return fmt.Errorf("erro ao codificar usuários: %w", err)
	}
	
	// Escrever no arquivo
	if err := ioutil.WriteFile(s.config.UsersFile, data, 0600); err != nil {
		return fmt.Errorf("erro ao salvar arquivo de usuários: %w", err)
	}
	
	return nil
}

// hashPassword gera um hash da senha usando Argon2
func (s *AuthService) hashPassword(password string) (string, string, error) {
	// Gerar salt aleatório
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", "", err
	}
	
	// Gerar o hash usando Argon2id
	hash := argon2.IDKey([]byte(password), salt, argonTime, argonMemory, argonThreads, argonKeyLen)
	
	// Codificar em base64
	hashStr := base64.StdEncoding.EncodeToString(hash)
	saltStr := base64.StdEncoding.EncodeToString(salt)
	
	return hashStr, saltStr, nil
}

// verifyPassword verifica se o hash da senha corresponde à senha
func (s *AuthService) verifyPassword(password, hash, saltStr string) bool {
	// Decodificar o salt
	salt, err := base64.StdEncoding.DecodeString(saltStr)
	if err != nil {
		return false
	}
	
	// Calcular o hash com a senha fornecida
	computedHash := argon2.IDKey([]byte(password), salt, argonTime, argonMemory, argonThreads, argonKeyLen)
	computedHashStr := base64.StdEncoding.EncodeToString(computedHash)
	
	// Comparar os hashes usando comparação de tempo constante
	return subtle.ConstantTimeCompare([]byte(hash), []byte(computedHashStr)) == 1
}

// CreateUser cria um novo usuário
// CreateUser creates a new user
// CreateUser crea un nuevo usuario
func (s *AuthService) CreateUser(username, password string, permissions []Permission) error {
	s.usersMutex.Lock()
	defer s.usersMutex.Unlock()
	
	// Verificar se o usuário já existe
	if _, exists := s.users[username]; exists {
		return ErrUserAlreadyExists
	}
	
	// Gerar hash da senha
	hash, salt, err := s.hashPassword(password)
	if err != nil {
		return fmt.Errorf("erro ao criar hash da senha: %w", err)
	}
	
	// Criar usuário
	user := User{
		Username:     username,
		PasswordHash: hash,
		Salt:         salt,
		Permissions:  permissions,
		CreatedAt:    time.Now(),
	}
	
	// Adicionar ao mapa
	s.users[username] = user
	
	// Salvar alterações
	return s.saveUsers()
}

// UpdateUserPassword atualiza a senha de um usuário
// UpdateUserPassword updates a user's password
// UpdateUserPassword actualiza la contraseña de un usuario
func (s *AuthService) UpdateUserPassword(username, newPassword string) error {
	s.usersMutex.Lock()
	defer s.usersMutex.Unlock()
	
	// Verificar se o usuário existe
	user, exists := s.users[username]
	if !exists {
		return ErrUserNotFound
	}
	
	// Gerar hash da nova senha
	hash, salt, err := s.hashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("erro ao criar hash da senha: %w", err)
	}
	
	// Atualizar usuário
	user.PasswordHash = hash
	user.Salt = salt
	s.users[username] = user
	
	// Salvar alterações
	return s.saveUsers()
}

// Authenticate autentica um usuário com nome/senha
// Authenticate authenticates a user with username/password
// Authenticate autentica un usuario con nombre/contraseña
func (s *AuthService) Authenticate(username, password string) (string, error) {
	s.usersMutex.RLock()
	user, exists := s.users[username]
	s.usersMutex.RUnlock()
	
	if !exists {
		return "", ErrInvalidCredentials
	}
	
	// Verificar a senha
	if !s.verifyPassword(password, user.PasswordHash, user.Salt) {
		return "", ErrInvalidCredentials
	}
	
	// Atualizar último login
	s.usersMutex.Lock()
	user.LastLogin = time.Now()
	s.users[username] = user
	s.usersMutex.Unlock()
	
	// Salvar alterações (em segundo plano para não bloquear)
	go s.saveUsers()
	
	return username, nil
}

// GetUser retorna um usuário específico (sem senha/chaves)
// GetUser returns a specific user (without password/keys)
// GetUser devuelve un usuario específico (sin contraseña/claves)
func (s *AuthService) GetUser(username string) (User, error) {
	s.usersMutex.RLock()
	defer s.usersMutex.RUnlock()
	
	// Buscar usuário
	user, exists := s.users[username]
	if !exists {
		return User{}, ErrUserNotFound
	}
	
	// Limpar dados sensíveis
	userCopy := user
	userCopy.PasswordHash = ""
	userCopy.Salt = ""
	
	return userCopy, nil
}

// HasPermission verifica se um usuário tem uma permissão específica
// HasPermission checks if a user has a specific permission
// HasPermission comprueba si un usuario tiene un permiso específico
func (s *AuthService) HasPermission(username string, permission Permission) bool {
	s.usersMutex.RLock()
	defer s.usersMutex.RUnlock()
	
	// Verificar se o usuário existe
	user, exists := s.users[username]
	if !exists {
		return false
	}
	
	// Usuários com permissão admin têm todas as permissões
	for _, p := range user.Permissions {
		if p == PermAdmin || p == permission {
			return true
		}
	}
	
	return false
}

package security

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	jwt "github.com/golang-jwt/jwt/v4"
)

// JWTConfig contém as configurações para geração de token JWT
// JWTConfig contains settings for JWT token generation
// JWTConfig contiene la configuración para la generación de tokens JWT
type JWTConfig struct {
	Secret     string        // Segredo para assinatura
	Expiration time.Duration // Tempo de expiração
}

// JWTClaims representa as claims do token JWT
// JWTClaims represents JWT token claims
// JWTClaims representa las claims del token JWT
type JWTClaims struct {
	Username    string       `json:"username"`
	Permissions []Permission `json:"permissions"`
	jwt.RegisteredClaims
}

// NewJWTConfig cria uma nova configuração JWT com valores padrão ou aleatórios
// NewJWTConfig creates a new JWT configuration with default or random values
// NewJWTConfig crea una nueva configuración JWT con valores predeterminados o aleatorios
func NewJWTConfig(secret string, expiration time.Duration) JWTConfig {
	if secret == "" {
		// Gerar um segredo aleatório
		secretBytes := make([]byte, 32)
		rand.Read(secretBytes)
		secret = base64.StdEncoding.EncodeToString(secretBytes)
	}

	if expiration == 0 {
		// Padrão: 24 horas
		expiration = 24 * time.Hour
	}

	return JWTConfig{
		Secret:     secret,
		Expiration: expiration,
	}
}

// GenerateToken gera um token JWT para o usuário especificado
// GenerateToken generates a JWT token for the specified user
// GenerateToken genera un token JWT para el usuario especificado
func GenerateToken(config JWTConfig, username string, permissions []Permission) (string, error) {
	// Definir tempo de expiração
	expiresAt := time.Now().Add(config.Expiration)

	// Criar claims
	claims := JWTClaims{
		Username:    username,
		Permissions: permissions,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   username,
		},
	}

	// Criar token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Assinar token
	signedToken, err := token.SignedString([]byte(config.Secret))
	if err != nil {
		return "", fmt.Errorf("erro ao assinar token: %w", err)
	}

	return signedToken, nil
}

// ValidateToken valida um token JWT e retorna as claims
// ValidateToken validates a JWT token and returns the claims
// ValidateToken valida un token JWT y devuelve las claims
func ValidateToken(config JWTConfig, tokenString string) (*JWTClaims, error) {
	// Analisar token
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Verificar se o método de assinatura é o esperado
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("método de assinatura inválido: %v", token.Header["alg"])
		}
		return []byte(config.Secret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("erro ao validar token: %w", err)
	}

	// Verificar se o token é válido
	if !token.Valid {
		return nil, errors.New("token inválido")
	}

	// Extrair claims
	claims, ok := token.Claims.(*JWTClaims)
	if !ok {
		return nil, errors.New("não foi possível extrair claims do token")
	}

	return claims, nil
}

// HasPermission verifica se um conjunto de permissões inclui uma permissão específica
// HasPermission checks if a set of permissions includes a specific permission
// HasPermission comprueba si un conjunto de permisos incluye un permiso específico
func HasPermission(permissions []Permission, permission Permission) bool {
	for _, p := range permissions {
		if p == PermAdmin || p == permission {
			return true
		}
	}
	return false
}

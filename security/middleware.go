package security

import (
	"context"
	"net/http"
	"strings"
)

// AuthMiddleware representa um middleware HTTP para autenticação
// AuthMiddleware represents an HTTP middleware for authentication
// AuthMiddleware representa un middleware HTTP para autenticación
type AuthMiddleware struct {
	JWTConfig      JWTConfig
	AuthService    *AuthService
	AllowedPaths   []string          // Caminhos que não requerem autenticação
	RequiredPerms  map[string]Permission // Mapeamento de caminhos para permissão necessária
}

// contextKey é um tipo para chaves de contexto
type contextKey string

const (
	// UserContextKey é a chave para armazenar o usuário no contexto da requisição
	UserContextKey contextKey = "user"
)

// WithUser retorna o usuário armazenado no contexto
// WithUser returns the user stored in the context
// WithUser devuelve el usuario almacenado en el contexto
func WithUser(r *http.Request) (string, []Permission, bool) {
	username, ok := r.Context().Value(UserContextKey).(string)
	if !ok {
		return "", nil, false
	}

	// Obter permissões do usuário
	perms, ok := r.Context().Value(contextKey("permissions")).([]Permission)
	if !ok {
		return username, nil, false
	}

	return username, perms, true
}

// Middleware retorna uma função handler que autentica as requisições
// Middleware returns a handler function that authenticates requests
// Middleware devuelve una función handler que autentica las solicitudes
func (m *AuthMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verificar se o caminho está na lista de permitidos
		for _, path := range m.AllowedPaths {
			if strings.HasPrefix(r.URL.Path, path) {
				next.ServeHTTP(w, r)
				return
			}
		}

		// Obter token do cabeçalho Authorization
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			// Tentar obter da cookie
			cookie, err := r.Cookie("auth_token")
			if err == nil {
				authHeader = "Bearer " + cookie.Value
			}
		}

		// Verificar se o cabeçalho existe e está no formato correto
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			http.Error(w, "Autenticação necessária", http.StatusUnauthorized)
			return
		}

		// Extrair token
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		// Validar token
		claims, err := ValidateToken(m.JWTConfig, tokenString)
		if err != nil {
			http.Error(w, "Token inválido ou expirado", http.StatusUnauthorized)
			return
		}

		// Verificar permissão para o caminho atual
		if requiredPerm, exists := m.RequiredPerms[r.URL.Path]; exists {
			hasPermission := false
			for _, perm := range claims.Permissions {
				if perm == PermAdmin || perm == requiredPerm {
					hasPermission = true
					break
				}
			}

			if !hasPermission {
				http.Error(w, "Permissão insuficiente", http.StatusForbidden)
				return
			}
		}

		// Adicionar usuário ao contexto da requisição
		ctx := context.WithValue(r.Context(), UserContextKey, claims.Username)
		ctx = context.WithValue(ctx, contextKey("permissions"), claims.Permissions)

		// Chamar o próximo handler com o contexto atualizado
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// PermissionsContextKey é a chave para armazenar as permissões no contexto da requisição
const PermissionsContextKey contextKey = "permissions"

package security

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// LoginResponse é a resposta para uma autenticação bem-sucedida
// LoginResponse is the response for a successful authentication
// LoginResponse es la respuesta para una autenticación exitosa
type LoginResponse struct {
	Username    string       `json:"username"`     // Nome do usuário
	Token       string       `json:"token"`        // Token JWT
	ExpiresAt   int64        `json:"expires_at"`   // Timestamp de expiração
	Permissions []Permission `json:"permissions"`  // Permissões do usuário
}

// LoginRequest é a requisição para autenticação
// LoginRequest is the request for authentication
// LoginRequest es la solicitud para autenticación
type LoginRequest struct {
	Username string `json:"username"` // Nome do usuário
	Password string `json:"password"` // Senha
}

// RegisterRequest é a requisição para registrar um novo usuário
// RegisterRequest is the request to register a new user
// RegisterRequest es la solicitud para registrar un nuevo usuario
type RegisterRequest struct {
	Username    string       `json:"username"`     // Nome do usuário
	Password    string       `json:"password"`     // Senha
	Permissions []Permission `json:"permissions"`  // Permissões do usuário (opcional)
}

// LoginHandler cria um handler para autenticação
// LoginHandler creates a handler for authentication
// LoginHandler crea un manejador para autenticación
func LoginHandler(authService *AuthService, jwtConfig JWTConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Apenas método POST é permitido
		if r.Method != http.MethodPost {
			http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
			return
		}

		// Decodificar requisição JSON
		var req LoginRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Formato de requisição inválido", http.StatusBadRequest)
			return
		}

		// Validar campos obrigatórios
		if req.Username == "" || req.Password == "" {
			http.Error(w, "Usuário e senha são obrigatórios", http.StatusBadRequest)
			return
		}

		// Tentar autenticar usuário
		token, err := authService.Authenticate(req.Username, req.Password)
		if err != nil {
			// Registrar tentativa de login falha
			fmt.Printf("Falha de login para usuário %s: %v\n", req.Username, err)
			http.Error(w, "Credenciais inválidas", http.StatusUnauthorized)
			return
		}

		// Obter informações do usuário
		user, err := authService.GetUser(req.Username)
		if err != nil {
			http.Error(w, "Erro ao obter informações do usuário", http.StatusInternalServerError)
			return
		}

		// Calcular tempo de expiração
		expiresAt := time.Now().Add(jwtConfig.Expiration).Unix()

		// Preparar resposta
		resp := LoginResponse{
			Username:    user.Username,
			Token:       token,
			ExpiresAt:   expiresAt,
			Permissions: user.Permissions,
		}

		// Registrar login bem-sucedido
		fmt.Printf("Login bem-sucedido para usuário %s\n", req.Username)

		// Retornar resposta JSON
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			http.Error(w, "Erro ao gerar resposta", http.StatusInternalServerError)
			return
		}
	}
}

// LogoutHandler cria um handler para logout
// LogoutHandler creates a handler for logout
// LogoutHandler crea un manejador para cierre de sesión
func LogoutHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Apenas método POST é permitido
		if r.Method != http.MethodPost {
			http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
			return
		}

		// Obter token do cabeçalho Authorization
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Token não fornecido", http.StatusBadRequest)
			return
		}

		// Extrair token do cabeçalho "Bearer <token>"
		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			http.Error(w, "Formato de token inválido", http.StatusBadRequest)
			return
		}

		// O token poderia ser adicionado a uma lista de tokens inválidos (blacklist)
		// Esta é uma implementação simplificada - em ambiente de produção, considere
		// implementar um sistema de revogação de tokens ou usar tokens com
		// expiração curta e refresh tokens.

		// Retornar sucesso
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]bool{"success": true})
	}
}

// RegisterHandler cria um handler para registrar novos usuários
// RegisterHandler creates a handler for registering new users
// RegisterHandler crea un manejador para registrar nuevos usuarios
func RegisterHandler(authService *AuthService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Apenas método POST é permitido
		if r.Method != http.MethodPost {
			http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
			return
		}

		// Decodificar requisição JSON
		var req RegisterRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Formato de requisição inválido", http.StatusBadRequest)
			return
		}

		// Validar campos obrigatórios
		if req.Username == "" || req.Password == "" {
			http.Error(w, "Usuário e senha são obrigatórios", http.StatusBadRequest)
			return
		}

		// Se não foram fornecidas permissões, definir como leitura apenas
		if len(req.Permissions) == 0 {
			req.Permissions = []Permission{PermReadOnly}
		}

		// Verificar se o usuário já existe
		_, err := authService.GetUser(req.Username)
		if err == nil {
			http.Error(w, "Usuário já existe", http.StatusConflict)
			return
		}

		// Criar usuário
		err = authService.CreateUser(req.Username, req.Password, req.Permissions)

		if err != nil {
			http.Error(w, fmt.Sprintf("Erro ao criar usuário: %v", err), http.StatusInternalServerError)
			return
		}

		// Registrar criação de usuário
		fmt.Printf("Usuário %s criado com permissões %v\n", req.Username, req.Permissions)

		// Retornar sucesso
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]bool{"success": true})
	}
}

// ChangePasswordRequest é a requisição para alterar a senha
// ChangePasswordRequest is the request to change the password
// ChangePasswordRequest es la solicitud para cambiar la contraseña
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password"` // Senha atual
	NewPassword     string `json:"new_password"`     // Nova senha
}

// ChangePasswordHandler cria um handler para alterar a senha do usuário
// ChangePasswordHandler creates a handler for changing the user's password
// ChangePasswordHandler crea un manejador para cambiar la contraseña del usuario
func ChangePasswordHandler(authService *AuthService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Apenas método POST é permitido
		if r.Method != http.MethodPost {
			http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
			return
		}

		// Obter usuário do contexto (adicionado pelo middleware de autenticação)
		userCtx := r.Context().Value(UserContextKey)
		if userCtx == nil {
			http.Error(w, "Usuário não autenticado", http.StatusUnauthorized)
			return
		}

		username := userCtx.(string)

		// Decodificar requisição JSON
		var req ChangePasswordRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Formato de requisição inválido", http.StatusBadRequest)
			return
		}

		// Validar campos obrigatórios
		if req.CurrentPassword == "" || req.NewPassword == "" {
			http.Error(w, "Senha atual e nova senha são obrigatórias", http.StatusBadRequest)
			return
		}

		// Verificar se a senha atual está correta
		if _, err := authService.Authenticate(username, req.CurrentPassword); err != nil {
			http.Error(w, "Senha atual incorreta", http.StatusUnauthorized)
			return
		}

		// Alterar a senha
		if err := authService.UpdateUserPassword(username, req.NewPassword); err != nil {
			http.Error(w, fmt.Sprintf("Erro ao alterar senha: %v", err), http.StatusInternalServerError)
			return
		}

		// Retornar sucesso
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]bool{"success": true})
	}
}

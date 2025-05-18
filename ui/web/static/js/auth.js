/**
 * VPN P2P Universal - Módulo de autenticação
 * VPN P2P Universal - Authentication module
 * VPN P2P Universal - Módulo de autenticación
 */

// Gestão de autenticação
const Auth = {
    // Estado de autenticação
    isLoggedIn: false,
    username: null,
    permissions: [],
    token: null,
    tokenExpiry: null,

    // Inicializar módulo de autenticação
    init() {
        // Verificar se há token armazenado
        this.token = localStorage.getItem('auth_token');
        
        if (this.token) {
            try {
                // Decodificar token (apenas a parte de payload)
                const tokenParts = this.token.split('.');
                if (tokenParts.length === 3) {
                    const payload = JSON.parse(atob(tokenParts[1]));
                    
                    // Verificar expiração
                    const expiryTime = payload.exp * 1000; // Converter para milissegundos
                    if (expiryTime > Date.now()) {
                        // Token ainda é válido
                        this.isLoggedIn = true;
                        this.username = payload.username || payload.sub;
                        this.permissions = payload.permissions || [];
                        this.tokenExpiry = new Date(expiryTime);
                        
                        // Configurar timer para renovação ou logout
                        this._setupAutoRefresh(expiryTime);
                        
                        return true;
                    }
                }
            } catch (e) {
                console.error('Erro ao processar token armazenado:', e);
            }
            
            // Se chegou aqui, o token não é válido
            this.logout();
        }
        
        return false;
    },

    // Tentativa de login
    async login(username, password) {
        try {
            const response = await fetch('/api/auth/login', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({ username, password })
            });

            if (!response.ok) {
                const error = await response.json();
                throw new Error(error.message || 'Falha na autenticação');
            }

            const data = await response.json();
            
            // Armazenar informações de autenticação
            this.isLoggedIn = true;
            this.username = data.username;
            this.permissions = data.permissions;
            this.token = data.token;
            this.tokenExpiry = new Date(data.expires_at * 1000);
            
            // Armazenar token
            localStorage.setItem('auth_token', this.token);
            
            // Configurar timer para renovação ou logout
            this._setupAutoRefresh(data.expires_at * 1000);
            
            return true;
        } catch (error) {
            console.error('Erro de login:', error);
            throw error;
        }
    },

    // Logout
    async logout() {
        try {
            // Tentar fazer logout no servidor
            if (this.token) {
                await fetch('/api/auth/logout', {
                    method: 'POST',
                    headers: {
                        'Authorization': `Bearer ${this.token}`
                    }
                }).catch(() => {
                    // Ignorar erros, apenas fazer logout localmente
                });
            }
        } finally {
            // Limpar dados de autenticação
            this.isLoggedIn = false;
            this.username = null;
            this.permissions = [];
            this.token = null;
            this.tokenExpiry = null;
            
            // Remover token armazenado
            localStorage.removeItem('auth_token');
            
            // Redirecionar para a página de login, se não estiver nela
            if (!window.location.pathname.includes('login.html')) {
                window.location.href = '/login.html';
            }
        }
    },

    // Verificar se o usuário tem uma permissão específica
    hasPermission(permission) {
        if (!this.isLoggedIn) return false;
        
        // Admin tem todas as permissões
        if (this.permissions.includes('admin')) return true;
        
        return this.permissions.includes(permission);
    },

    // Verificar se o usuário está autenticado
    checkAuth() {
        if (!this.isLoggedIn) {
            // Redirecionar para a página de login
            window.location.href = '/login.html';
            return false;
        }
        return true;
    },

    // Configurar timer para renovação ou logout automático
    _setupAutoRefresh(expiryTime) {
        // Limpar timer existente
        if (this._refreshTimer) {
            clearTimeout(this._refreshTimer);
        }
        
        // Calcular tempo até expiração
        const timeToExpiry = expiryTime - Date.now();
        
        // Se faltar menos de 5 minutos, tentar renovar agora
        if (timeToExpiry < 5 * 60 * 1000) {
            this._refreshToken();
            return;
        }
        
        // Configurar timer para renovar 5 minutos antes da expiração
        this._refreshTimer = setTimeout(() => {
            this._refreshToken();
        }, timeToExpiry - 5 * 60 * 1000);
    },

    // Renovar token
    async _refreshToken() {
        try {
            // Implementar lógica de renovação do token quando necessário
            // Por enquanto, apenas verificar se precisa fazer logout
            if (this.tokenExpiry && this.tokenExpiry.getTime() < Date.now()) {
                this.logout();
            }
        } catch (error) {
            console.error('Erro ao renovar token:', error);
            this.logout();
        }
    },

    // Obter cabeçalhos para requisições autenticadas
    getHeaders() {
        const headers = {
            'Content-Type': 'application/json'
        };
        
        if (this.token) {
            headers['Authorization'] = `Bearer ${this.token}`;
        }
        
        return headers;
    },

    // Envolver fetch com autenticação
    async fetch(url, options = {}) {
        if (!this.isLoggedIn) {
            throw new Error('Usuário não autenticado');
        }
        
        // Adicionar cabeçalhos de autenticação
        const authOptions = {
            ...options,
            headers: {
                ...options.headers,
                'Authorization': `Bearer ${this.token}`
            }
        };
        
        const response = await fetch(url, authOptions);
        
        // Verificar se o token expirou
        if (response.status === 401) {
            this.logout();
            throw new Error('Sessão expirada. Por favor, faça login novamente.');
        }
        
        return response;
    }
};

// Inicializar ao carregar a página
document.addEventListener('DOMContentLoaded', () => {
    Auth.init();
});

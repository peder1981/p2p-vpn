/**
 * VPN P2P Universal - Código JavaScript da interface web
 * VPN P2P Universal - Web interface JavaScript code
 * VPN P2P Universal - Código JavaScript de la interfaz web
 */

// Configurações
const API_BASE_URL = '/api';
const UPDATE_INTERVAL = 10000; // 10 segundos

// Elementos DOM principais
const elements = {
    // Navegação
    navLinks: document.querySelectorAll('nav a'),
    sections: document.querySelectorAll('.section'),
    languageSelector: document.getElementById('language'),
    
    // Status
    connectionState: document.getElementById('connection-state'),
    connectionStatus: document.getElementById('connection-status'),
    interfaceName: document.getElementById('interface-name'),
    nodeId: document.getElementById('node-id'),
    virtualIp: document.getElementById('virtual-ip'),
    peersCount: document.getElementById('peers-count'),
    virtualNetwork: document.getElementById('virtual-network'),
    publicKey: document.getElementById('public-key'),
    
    // Ações de Status
    startVpnBtn: document.getElementById('start-vpn'),
    stopVpnBtn: document.getElementById('stop-vpn'),
    refreshStatusBtn: document.getElementById('refresh-status'),
    
    // Peers
    peersList: document.getElementById('peers-list'),
    addPeerBtn: document.getElementById('add-peer'),
    refreshPeersBtn: document.getElementById('refresh-peers'),
    
    // Modal para adicionar peer
    addPeerModal: document.getElementById('add-peer-modal'),
    addPeerForm: document.getElementById('add-peer-form'),
    cancelAddPeerBtn: document.getElementById('cancel-add-peer'),
    closeModalBtn: document.querySelector('.close'),
    
    // Configurações
    networkSettingsForm: document.getElementById('network-settings-form'),
    discoverySettingsForm: document.getElementById('discovery-settings-form')
};

// Estado da aplicação
const state = {
    vpnRunning: false,
    vpnConfig: null,
    peers: [],
    language: 'pt-br'
};

// Textos traduzidos
const translations = {
    'pt-br': {
        connectionStates: {
            connected: 'Conectado',
            disconnected: 'Desconectado',
            loading: 'Carregando...'
        },
        messages: {
            peerAdded: 'Peer adicionado com sucesso!',
            peerRemoved: 'Peer removido com sucesso!',
            connectionError: 'Erro de conexão com a API',
            confirmDelete: 'Tem certeza que deseja remover este peer?'
        }
    },
    'en': {
        connectionStates: {
            connected: 'Connected',
            disconnected: 'Disconnected',
            loading: 'Loading...'
        },
        messages: {
            peerAdded: 'Peer added successfully!',
            peerRemoved: 'Peer removed successfully!',
            connectionError: 'API connection error',
            confirmDelete: 'Are you sure you want to remove this peer?'
        }
    },
    'es': {
        connectionStates: {
            connected: 'Conectado',
            disconnected: 'Desconectado',
            loading: 'Cargando...'
        },
        messages: {
            peerAdded: 'Par añadido con éxito!',
            peerRemoved: 'Par eliminado con éxito!',
            connectionError: 'Error de conexión con la API',
            confirmDelete: '¿Está seguro de que desea eliminar este par?'
        }
    }
};

// Função para obter texto traduzido
function getTranslation(key, subKey) {
    return translations[state.language][key][subKey];
}

// Inicializar a aplicação
function init() {
    // Configurar navegação
    setupNavigation();
    
    // Configurar seletor de idioma
    setupLanguageSelector();
    
    // Carregar dados iniciais
    loadStatus();
    loadPeers();
    
    // Configurar modais
    setupModals();
    
    // Configurar formulários
    setupForms();
    
    // Configurar atualização automática
    setInterval(() => {
        if (isActiveSection('status')) {
            loadStatus();
        } else if (isActiveSection('peers')) {
            loadPeers();
        }
    }, UPDATE_INTERVAL);
}

// Verifica se uma seção está ativa
function isActiveSection(sectionId) {
    return document.getElementById(sectionId).classList.contains('active');
}

// Configurar navegação entre seções
function setupNavigation() {
    elements.navLinks.forEach(link => {
        link.addEventListener('click', (e) => {
            e.preventDefault();
            
            // Remover classe ativa de todos os links e seções
            elements.navLinks.forEach(l => l.classList.remove('active'));
            elements.sections.forEach(s => s.classList.remove('active'));
            
            // Adicionar classe ativa ao link clicado
            link.classList.add('active');
            
            // Mostrar a seção correspondente
            const sectionId = link.getAttribute('data-section');
            document.getElementById(sectionId).classList.add('active');
        });
    });
}

// Configurar seletor de idioma
function setupLanguageSelector() {
    elements.languageSelector.addEventListener('change', (e) => {
        state.language = e.target.value;
        updateUILanguage();
    });
}

// Atualizar idioma da interface
function updateUILanguage() {
    // Atualizar textos dinâmicos
    if (state.vpnRunning) {
        elements.connectionState.textContent = getTranslation('connectionStates', 'connected');
    } else {
        elements.connectionState.textContent = getTranslation('connectionStates', 'disconnected');
    }
    
    // Atualizar a lista de peers para aplicar novos textos
    refreshPeersList();
}

// Função para obter status atual da VPN
async function fetchStatus() {
    try {
        const response = await fetch(`${API_BASE_URL}/status`, {
            headers: Auth.getHeaders()
        });
        
        if (response.status === 401) {
            // Token expirado ou inválido
            Auth.logout();
            return { running: false, error: 'Sessão expirada' };
        }
        
        if (!response.ok) throw new Error('Erro ao obter status');
        return await response.json();
    } catch (error) {
        console.error('Erro ao buscar status:', error);
        return { running: false, error: error.message };
    }
}

// Carregar o status da VPN
async function loadStatus() {
    try {
        const data = await fetchStatus();
        
        // Atualizar o estado da aplicação
        state.vpnRunning = data.running;
        
        // Atualizar a interface
        updateStatusUI(data);
    } catch (error) {
        console.error('Error loading status:', error);
        elements.connectionState.textContent = getTranslation('connectionStates', 'disconnected');
        elements.connectionStatus.classList.remove('connected');
        elements.connectionStatus.classList.add('disconnected');
    }
}

// Atualizar a interface de status
function updateStatusUI(data) {
    // Status de conexão
    if (data.running) {
        elements.connectionState.textContent = getTranslation('connectionStates', 'connected');
        elements.connectionStatus.classList.remove('disconnected');
        elements.connectionStatus.classList.add('connected');
        
        // Habilitar/desabilitar botões
        elements.startVpnBtn.disabled = true;
        elements.stopVpnBtn.disabled = false;
    } else {
        elements.connectionState.textContent = getTranslation('connectionStates', 'disconnected');
        elements.connectionStatus.classList.remove('connected');
        elements.connectionStatus.classList.add('disconnected');
        
        // Habilitar/desabilitar botões
        elements.startVpnBtn.disabled = false;
        elements.stopVpnBtn.disabled = true;
    }
    
    // Informações do nó
    elements.interfaceName.textContent = data.interface || 'wg0';
    elements.nodeId.textContent = data.node_id || '-';
    elements.virtualIp.textContent = data.virtual_ip || '-';
    
    // Informações adicionais
    elements.peersCount.textContent = data.peers_count || '0';
    elements.virtualNetwork.textContent = data.virtual_cidr || '-';
    elements.publicKey.textContent = data.public_key || '-';
}

// Função para obter a lista de peers
async function fetchPeers() {
    try {
        const response = await fetch(`${API_BASE_URL}/peers`, {
            headers: Auth.getHeaders()
        });
        
        if (response.status === 401) {
            // Token expirado ou inválido
            Auth.logout();
            return [];
        }
        
        if (!response.ok) throw new Error('Erro ao obter peers');
        return await response.json();
    } catch (error) {
        console.error('Erro ao buscar peers:', error);
        return [];
    }
}

// Carregar a lista de peers
async function loadPeers() {
    try {
        const data = await fetchPeers();
        
        // Atualizar o estado da aplicação
        state.peers = data;
        
        // Atualizar a interface
        refreshPeersList();
    } catch (error) {
        console.error('Error loading peers:', error);
    }
}

// Atualizar a lista de peers na interface
function refreshPeersList() {
    // Limpar a lista atual
    elements.peersList.innerHTML = '';
    
    if (state.peers.length === 0) {
        // Mostrar mensagem se não houver peers
        const emptyRow = document.createElement('tr');
        emptyRow.innerHTML = `
            <td colspan="5" style="text-align: center; padding: 2rem;">
                Nenhum peer configurado
            </td>
        `;
        elements.peersList.appendChild(emptyRow);
        return;
    }
    
    // Adicionar cada peer à lista
    state.peers.forEach(peer => {
        const peerRow = document.createElement('tr');
        
        // Status do peer (ativo/inativo)
        const statusClass = peer.active ? 'active' : 'inactive';
        
        // Endpoints formatados
        const endpoints = peer.endpoints && peer.endpoints.length > 0
            ? peer.endpoints.join(', ')
            : '-';
        
        peerRow.innerHTML = `
            <td>
                <span class="peer-status ${statusClass}"></span>
                ${peer.active ? 'Ativo' : 'Inativo'}
            </td>
            <td>${peer.node_id}</td>
            <td>${peer.virtual_ip}</td>
            <td>${endpoints}</td>
            <td class="peer-actions">
                <button class="connect" title="Conectar" data-id="${peer.node_id}">
                    <i class="fas fa-plug"></i>
                </button>
                <button class="edit" title="Editar" data-id="${peer.node_id}">
                    <i class="fas fa-edit"></i>
                </button>
                <button class="delete" title="Remover" data-id="${peer.node_id}">
                    <i class="fas fa-trash"></i>
                </button>
            </td>
        `;
        
        elements.peersList.appendChild(peerRow);
    });
    
    // Adicionar event listeners para as ações
    document.querySelectorAll('.peer-actions .delete').forEach(button => {
        button.addEventListener('click', handleDeletePeer);
    });
    
    document.querySelectorAll('.peer-actions .connect').forEach(button => {
        button.addEventListener('click', handleConnectPeer);
    });
    
    document.querySelectorAll('.peer-actions .edit').forEach(button => {
        button.addEventListener('click', handleEditPeer);
    });
}

// Configurar modais
function setupModals() {
    // Abrir modal para adicionar peer
    elements.addPeerBtn.addEventListener('click', () => {
        elements.addPeerModal.style.display = 'block';
    });
    
    // Fechar modal (X)
    elements.closeModalBtn.addEventListener('click', () => {
        elements.addPeerModal.style.display = 'none';
    });
    
    // Fechar modal (botão cancelar)
    elements.cancelAddPeerBtn.addEventListener('click', () => {
        elements.addPeerModal.style.display = 'none';
    });
    
    // Fechar modal clicando fora
    window.addEventListener('click', (event) => {
        if (event.target === elements.addPeerModal) {
            elements.addPeerModal.style.display = 'none';
        }
    });
}

// Configurar formulários
function setupForms() {
    // Formulário para adicionar peer
    elements.addPeerForm.addEventListener('submit', handleAddPeer);
    
    // Botão de atualizar status
    elements.refreshStatusBtn.addEventListener('click', loadStatus);
    
    // Botão de atualizar peers
    elements.refreshPeersBtn.addEventListener('click', loadPeers);
}

// Manipular adição de peer
async function handleAddPeer(e) {
    e.preventDefault();
    
    // Obter dados do formulário
    const peerData = {
        node_id: document.getElementById('peer-id').value,
        public_key: document.getElementById('peer-public-key').value,
        virtual_ip: document.getElementById('peer-virtual-ip').value,
        endpoints: []
    };
    
    // Validar dados obrigatórios
    if (!peerData.public_key || !peerData.virtual_ip) {
        alert('Chave pública e IP virtual são obrigatórios!');
        return;
    }
    
    // Adicionar endpoint se fornecido
    const endpoint = document.getElementById('peer-endpoint').value;
    if (endpoint) {
        peerData.endpoints.push(endpoint);
    }
    
    // Adicionar keepalive se fornecido
    const keepalive = document.getElementById('peer-keepalive').value;
    if (keepalive) {
        peerData.keep_alive = parseInt(keepalive, 10);
    }
    
    try {
        const response = await addPeer(peerData);
        
        // Fechar o modal
        elements.addPeerModal.style.display = 'none';
        
        // Limpar o formulário
        elements.addPeerForm.reset();
        
        // Recarregar lista de peers
        loadPeers();
        
        // Mostrar mensagem de sucesso
        alert(getTranslation('messages', 'peerAdded'));
    } catch (error) {
        console.error('Erro ao adicionar peer:', error);
        alert(getTranslation('messages', 'connectionError'));
    }
}

// Função para adicionar um novo peer
async function addPeer(peer) {
    try {
        // Verificar se tem permissão para modificar peers
        if (!Auth.hasPermission('write')) {
            throw new Error('Você não tem permissão para adicionar peers');
        }
        
        const response = await fetch(`${API_BASE_URL}/peers`, {
            method: 'POST',
            headers: {
                ...Auth.getHeaders(),
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(peer)
        });

        if (response.status === 401) {
            // Token expirado ou inválido
            Auth.logout();
            throw new Error('Sessão expirada');
        }
        
        if (response.status === 403) {
            throw new Error('Permissão negada');
        }

        if (!response.ok) {
            const error = await response.json();
            throw new Error(error.message || 'Erro ao adicionar peer');
        }

        return await response.json();
    } catch (error) {
        console.error('Erro ao adicionar peer:', error);
        throw error;
    }
}

// Manipular remoção de peer
async function handleDeletePeer(e) {
    const peerId = e.currentTarget.getAttribute('data-id');
    
    // Confirmar remoção
    if (!confirm(getTranslation('messages', 'confirmDelete'))) {
        return;
    }
    
    try {
        await removePeer(peerId);
        
        // Recarregar lista de peers
        loadPeers();
        
        // Mostrar mensagem de sucesso
        alert(getTranslation('messages', 'peerRemoved'));
    } catch (error) {
        console.error('Erro ao remover peer:', error);
        alert(getTranslation('messages', 'connectionError'));
    }
}

// Função para remover um peer existente
async function removePeer(peerId) {
    try {
        // Verificar se tem permissão para modificar peers
        if (!Auth.hasPermission('write')) {
            throw new Error('Você não tem permissão para remover peers');
        }
        
        const response = await fetch(`${API_BASE_URL}/peers/${peerId}`, {
            method: 'DELETE',
            headers: Auth.getHeaders()
        });

        if (response.status === 401) {
            // Token expirado ou inválido
            Auth.logout();
            throw new Error('Sessão expirada');
        }
        
        if (response.status === 403) {
            throw new Error('Permissão negada');
        }

        if (!response.ok) {
            const error = await response.json();
            throw new Error(error.message || 'Erro ao remover peer');
        }

        return true;
    } catch (error) {
        console.error('Erro ao remover peer:', error);
        throw error;
    }
}

// Manipular conexão a um peer
function handleConnectPeer(e) {
    const peerId = e.currentTarget.getAttribute('data-id');
    
    // Encontrar o peer na lista
    const peer = state.peers.find(p => p.node_id === peerId);
    
    if (!peer) {
        return;
    }
    
    // Solicitar o endpoint para conexão
    let endpoint = prompt('Digite o endpoint do peer (IP:porta):', 
        peer.endpoints && peer.endpoints.length > 0 ? peer.endpoints[0] : '');
    
    if (!endpoint) {
        return;
    }
    
    // Em uma implementação real, enviar solicitação para API
    // Por ora, apenas simular
    alert(`Conectando ao peer ${peerId} em ${endpoint}...`);
}

// Manipular edição de peer (não implementado)
function handleEditPeer(e) {
    const peerId = e.currentTarget.getAttribute('data-id');
    alert(`Edição de peer não implementada: ${peerId}`);
}

// Inicializar a aplicação quando a página carregar
document.addEventListener('DOMContentLoaded', function() {
    // Verificar se o usuário está autenticado
    if (!Auth.isLoggedIn) {
        // Redirecionar para a página de login
        window.location.href = '/login.html';
        return;
    }
    
    // Mostrar informações do usuário
    document.getElementById('username-display').textContent = Auth.username;
    
    // Exibir ou ocultar elementos baseado nas permissões
    if (Auth.hasPermission('admin')) {
        // Exibir todos os elementos administrativos
        document.querySelectorAll('.admin-only').forEach(el => {
            el.style.display = 'block';
        });
    }
    
    // Configurar o botão de logout
    document.getElementById('logout-button').addEventListener('click', function() {
        Auth.logout();
    });
    
    // Inicializar o resto da aplicação
    init();
});

#!/bin/bash

# Script de instalação para macOS do P2P-VPN
# macOS installation script for P2P-VPN
# Script de instalación para macOS de P2P-VPN

set -e

# Cores para saída
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Diretórios
INSTALL_DIR="/usr/local/p2p-vpn"
CONFIG_DIR="/etc/p2p-vpn"
LAUNCHD_DIR="/Library/LaunchDaemons"
LAUNCHD_FILE="com.p2p-vpn.service.plist"

echo -e "${GREEN}Iniciando instalação do P2P-VPN para macOS...${NC}"

# Verificar se está rodando como root
if [ "$EUID" -ne 0 ]; then
  echo -e "${RED}Este script deve ser executado como root!${NC}"
  echo "Por favor, execute novamente com sudo ou como usuário root."
  exit 1
fi

# Verificar dependências
echo -e "${YELLOW}Verificando dependências...${NC}"

# Verificar se o Homebrew está instalado
if ! command -v brew &> /dev/null; then
    echo -e "${BLUE}Homebrew não encontrado.${NC}"
    echo -e "${YELLOW}Deseja instalar o Homebrew? [S/n]${NC}"
    read -r response
    response=${response:-S} # Padrão para Sim
    
    if [[ "$response" =~ ^([sS]|[sS][iI]|[yY]|[yY][eE][sS])$ ]]; then
        echo -e "${YELLOW}Instalando Homebrew...${NC}"
        /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
    else
        echo -e "${RED}Homebrew é necessário para a instalação das dependências.${NC}"
        echo "Por favor, instale manualmente e tente novamente."
        exit 1
    fi
fi

# Verificar e instalar WireGuard se necessário
if ! command -v wireguard-go &> /dev/null; then
    echo -e "${YELLOW}WireGuard não encontrado. Instalando via Homebrew...${NC}"
    brew install wireguard-go wireguard-tools
fi

# Verificar e instalar dependências adicionais se necessárias
brew install jq # Para manipulação de JSON

# Criar diretórios de instalação
echo -e "${YELLOW}Criando diretórios...${NC}"
mkdir -p "$INSTALL_DIR"
mkdir -p "$CONFIG_DIR"
mkdir -p "$CONFIG_DIR/users"

# Copiar arquivos
echo -e "${YELLOW}Copiando arquivos...${NC}"
cp -f ./p2p-vpn "$INSTALL_DIR/"
cp -f ./config.yaml.example "$CONFIG_DIR/config.yaml"

# Configurar permissões
echo -e "${YELLOW}Configurando permissões...${NC}"
chmod +x "$INSTALL_DIR/p2p-vpn"
chmod 640 "$CONFIG_DIR/config.yaml"

# Criar arquivo de usuários se não existir
touch "$CONFIG_DIR/users/users.json"
chmod 640 "$CONFIG_DIR/users/users.json"

# Instalar arquivo LaunchDaemon
echo -e "${YELLOW}Instalando arquivo LaunchDaemon...${NC}"
cp -f ./packaging/macos/$LAUNCHD_FILE "$LAUNCHD_DIR/"
chmod 644 "$LAUNCHD_DIR/$LAUNCHD_FILE"

# Carregar serviço
echo -e "${YELLOW}Carregando serviço LaunchDaemon...${NC}"
launchctl unload -w "$LAUNCHD_DIR/$LAUNCHD_FILE" 2>/dev/null || true
launchctl load -w "$LAUNCHD_DIR/$LAUNCHD_FILE"

# Verificar se o módulo do kernel WireGuard está carregado
echo -e "${YELLOW}Verificando módulo WireGuard...${NC}"
if ! kextstat | grep -q wireguard; then
    echo -e "${BLUE}Módulo WireGuard não encontrado.${NC}"
    echo -e "${YELLOW}Usando implementação em userspace (wireguard-go)${NC}"
fi

# Configurar firewall se estiver habilitado
if command -v /usr/libexec/ApplicationFirewall/socketfilterfw &> /dev/null; then
    echo -e "${YELLOW}Configurando firewall...${NC}"
    /usr/libexec/ApplicationFirewall/socketfilterfw --add "$INSTALL_DIR/p2p-vpn"
    /usr/libexec/ApplicationFirewall/socketfilterfw --unblockapp "$INSTALL_DIR/p2p-vpn"
fi

echo -e "${GREEN}Instalação concluída!${NC}"
echo -e "${YELLOW}Interface web disponível em:${NC} http://localhost:8080"
echo -e "${YELLOW}Arquivos de configuração em:${NC} $CONFIG_DIR"
echo -e "${YELLOW}Para gerenciar o serviço, use:${NC}"
echo -e "  sudo launchctl start/stop $LAUNCHD_FILE"

exit 0

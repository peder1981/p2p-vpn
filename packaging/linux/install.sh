#!/bin/bash

# Script de instalação para Linux do P2P-VPN
# Linux installation script for P2P-VPN
# Script de instalación para Linux de P2P-VPN

set -e

# Cores para saída
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Diretórios
INSTALL_DIR="/opt/p2p-vpn"
CONFIG_DIR="/etc/p2p-vpn"
SYSTEMD_DIR="/etc/systemd/system"

echo -e "${GREEN}Iniciando instalação do P2P-VPN para Linux...${NC}"

# Verificar se está rodando como root
if [ "$EUID" -ne 0 ]; then
  echo -e "${RED}Este script deve ser executado como root!${NC}"
  echo "Por favor, execute novamente com sudo ou como usuário root."
  exit 1
fi

# Verificar dependências
echo -e "${YELLOW}Verificando dependências...${NC}"

DEPS_TO_INSTALL=()
USE_BORINGTUN=false

check_and_install_dependencies() {
    local deps=("wireguard" "wireguard-tools" "iproute2")
    local deps_names=("WireGuard" "Ferramentas WireGuard" "Ferramentas de rede")
    
    for i in "${!deps[@]}"; do
        local pkg="${deps[$i]}"
        local name="${deps_names[$i]}"
        
        echo -n "Verificando $name... "
        if ! command -v wg &> /dev/null && [[ "$pkg" == "wireguard-tools" ]]; then
            echo -e "${RED}Não encontrado!${NC}"
            DEPS_TO_INSTALL+=("$pkg")
        elif ! ip -V &> /dev/null && [[ "$pkg" == "iproute2" ]]; then
            echo -e "${RED}Não encontrado!${NC}"
            DEPS_TO_INSTALL+=("$pkg")
        elif [[ "$pkg" == "wireguard" ]] && ! grep -q wireguard /proc/modules && ! grep -q wireguard /lib/modules/$(uname -r)/modules.builtin; then
            echo -e "${YELLOW}Módulo não carregado!${NC}"
            DEPS_TO_INSTALL+=("$pkg")
            
            # Se não encontramos o módulo do kernel, verificar se queremos instalar boringtun
            echo -e "${YELLOW}Módulo WireGuard não encontrado no kernel.${NC}"
            echo -e "${YELLOW}Deseja instalar o boringtun como fallback? [S/n]${NC}"
            read -r response
            response=${response:-S} # Padrão para Sim
            
            if [[ "$response" =~ ^([sS]|[sS][iI]|[yY]|[yY][eE][sS])$ ]]; then
                USE_BORINGTUN=true
            fi
        else
            echo -e "${GREEN}OK!${NC}"
        fi
    done
    
    # Verificar o Rust para o boringtun se necessário
    if [[ "$USE_BORINGTUN" == true ]]; then
        echo -n "Verificando Rust... "
        if ! command -v rustc &> /dev/null || ! command -v cargo &> /dev/null; then
            echo -e "${RED}Não encontrado!${NC}"
            echo -e "${YELLOW}Rust é necessário para compilar o boringtun.${NC}"
            echo -e "${YELLOW}Deseja instalar o Rust? [S/n]${NC}"
            read -r response
            response=${response:-S} # Padrão para Sim
            
            if [[ "$response" =~ ^([sS]|[sS][iI]|[yY]|[yY][eE][sS])$ ]]; then
                echo -e "${YELLOW}Instalando Rust...${NC}"
                curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh -s -- -y
                source "$HOME/.cargo/env"
            else
                echo -e "${RED}Abortando instalação do boringtun.${NC}"
                USE_BORINGTUN=false
            fi
        else
            echo -e "${GREEN}OK!${NC}"
        fi
    fi
}

check_and_install_dependencies

# Instalar dependências se necessário
if [ ${#DEPS_TO_INSTALL[@]} -gt 0 ]; then
    echo -e "${YELLOW}Instalando dependências...${NC}"
    
    # Detectar gerenciador de pacotes
    if command -v apt-get &> /dev/null; then
        apt-get update
        apt-get install -y "${DEPS_TO_INSTALL[@]}"
    elif command -v dnf &> /dev/null; then
        dnf install -y "${DEPS_TO_INSTALL[@]}"
    elif command -v yum &> /dev/null; then
        yum install -y "${DEPS_TO_INSTALL[@]}"
    elif command -v pacman &> /dev/null; then
        pacman -Sy --noconfirm "${DEPS_TO_INSTALL[@]}"
    else
        echo -e "${RED}Não foi possível detectar o gerenciador de pacotes.${NC}"
        echo "Por favor, instale manualmente os seguintes pacotes:"
        for pkg in "${DEPS_TO_INSTALL[@]}"; do
            echo "- $pkg"
        done
        exit 1
    fi
fi

# Instalar boringtun se necessário
if [ "$USE_BORINGTUN" = true ]; then
    echo -e "${YELLOW}Instalando boringtun como implementação de fallback...${NC}"
    
    # Verificar se já está instalado
    if command -v boringtun &> /dev/null; then
        echo -e "${GREEN}boringtun já está instalado!${NC}"
    else
        # Criar diretório temporário para compilação
        TEMP_DIR=$(mktemp -d)
        cd "$TEMP_DIR"
        
        echo -e "${YELLOW}Baixando código-fonte do boringtun...${NC}"
        git clone --depth=1 https://github.com/cloudflare/boringtun.git
        cd boringtun
        
        echo -e "${YELLOW}Compilando boringtun (isso pode levar alguns minutos)...${NC}"
        cargo build --release
        
        echo -e "${YELLOW}Instalando boringtun...${NC}"
        cp target/release/boringtun /usr/local/bin/
        chmod +x /usr/local/bin/boringtun
        
        echo -e "${GREEN}boringtun instalado com sucesso!${NC}"
        
        # Voltar ao diretório original e limpar
        cd - > /dev/null
        rm -rf "$TEMP_DIR"
    fi
    
    # Configurar variável de ambiente para habilitar boringtun
    echo 'export WG_QUICK_USERSPACE_IMPLEMENTATION=boringtun' >> /etc/profile.d/boringtun.sh
    chmod +x /etc/profile.d/boringtun.sh
    
    # Verificar se wireguard-go também está instalado como backup adicional
    if ! command -v wireguard-go &> /dev/null; then
        echo -e "${YELLOW}Instalando wireguard-go como backup adicional...${NC}"
        
        # Instalar wireguard-go usando o gerenciador de pacotes
        if command -v apt-get &> /dev/null; then
            apt-get install -y wireguard-tools
        elif command -v dnf &> /dev/null; then
            dnf install -y wireguard-tools
        elif command -v yum &> /dev/null; then
            yum install -y wireguard-tools
        elif command -v pacman &> /dev/null; then
            pacman -Sy --noconfirm wireguard-tools
        fi
        
        # Se não conseguimos instalar via gerenciador de pacotes, compilar do código fonte
        if ! command -v wireguard-go &> /dev/null; then
            if command -v go &> /dev/null; then
                echo -e "${YELLOW}Compilando wireguard-go do código-fonte...${NC}"
                TEMP_DIR=$(mktemp -d)
                cd "$TEMP_DIR"
                git clone --depth=1 https://git.zx2c4.com/wireguard-go
                cd wireguard-go
                make
                cp wireguard-go /usr/local/bin/
                chmod +x /usr/local/bin/wireguard-go
                cd - > /dev/null
                rm -rf "$TEMP_DIR"
            else
                echo -e "${YELLOW}Go não encontrado, pulando instalação de wireguard-go.${NC}"
            fi
        fi
    fi
fi

# Criar diretórios de instalação
echo -e "${YELLOW}Criando diretórios...${NC}"
mkdir -p "$INSTALL_DIR"
mkdir -p "$CONFIG_DIR"

# Copiar arquivos
echo -e "${YELLOW}Copiando arquivos...${NC}"
cp -f ./p2p-vpn "$INSTALL_DIR/"
cp -f ./config.yaml.example "$CONFIG_DIR/config.yaml"
cp -f ./packaging/linux/p2p-vpn.service "$SYSTEMD_DIR/"

# Configurar permissões
echo -e "${YELLOW}Configurando permissões...${NC}"
chmod +x "$INSTALL_DIR/p2p-vpn"
chmod 640 "$CONFIG_DIR/config.yaml"

# Criar diretório para usuários se não existir
mkdir -p "$CONFIG_DIR/users"
touch "$CONFIG_DIR/users/users.json"
chmod 640 "$CONFIG_DIR/users/users.json"

# Configurar serviço systemd
echo -e "${YELLOW}Configurando serviço systemd...${NC}"
systemctl daemon-reload
systemctl enable p2p-vpn.service

echo -e "${GREEN}Instalação concluída!${NC}"
echo -e "${YELLOW}Para iniciar o serviço, execute:${NC} sudo systemctl start p2p-vpn"
echo -e "${YELLOW}Para verificar o status, execute:${NC} sudo systemctl status p2p-vpn"
echo -e "${YELLOW}Configuração disponível em:${NC} $CONFIG_DIR/config.yaml"

exit 0

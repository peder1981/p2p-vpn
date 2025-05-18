#!/bin/bash

# Script de instalação para P2P VPN Universal
# Este script instala automaticamente o P2P VPN em sistemas Linux

# Cores para saída no terminal
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # Sem cor

echo -e "${GREEN}=========================================${NC}"
echo -e "${GREEN}   Instalador do P2P VPN Universal      ${NC}"
echo -e "${GREEN}=========================================${NC}"
echo ""

# Verificar privilégios de root
if [ "$EUID" -ne 0 ]; then
  echo -e "${RED}Este script precisa ser executado como root.${NC}"
  echo "Por favor, execute com sudo ou como usuário root."
  exit 1
fi

# Detectar distribuição Linux
if [ -f /etc/os-release ]; then
    . /etc/os-release
    DISTRO=$ID
    VERSION=$VERSION_ID
else
    echo -e "${RED}Não foi possível determinar a distribuição Linux.${NC}"
    exit 1
fi

echo -e "Distribuição detectada: ${YELLOW}$DISTRO $VERSION${NC}"

# Definir o repositório GitHub
REPO_URL="https://github.com/peder1981/p2p-vpn"
RELEASE_URL="$REPO_URL/releases/latest"

# Função para baixar a versão mais recente
download_latest() {
    echo "Baixando a versão mais recente do P2P VPN Universal..."
    
    # Determinar a arquitetura
    ARCH=$(uname -m)
    case $ARCH in
        x86_64)
            ARCH_NAME="amd64"
            ;;
        aarch64|arm64)
            ARCH_NAME="arm64"
            ;;
        *)
            echo -e "${RED}Arquitetura não suportada: $ARCH${NC}"
            exit 1
            ;;
    esac
    
    # Criar diretório temporário
    TMP_DIR=$(mktemp -d)
    
    # Determinar o pacote correto com base na distribuição
    case $DISTRO in
        ubuntu|debian|linuxmint|pop|elementary)
            # Debian/Ubuntu e derivados usam .deb
            PACKAGE_URL="$RELEASE_URL/download/p2p-vpn_latest_${ARCH_NAME}.deb"
            PACKAGE_FILE="$TMP_DIR/p2p-vpn.deb"
            
            echo "Baixando pacote .deb..."
            wget -q --show-progress -O "$PACKAGE_FILE" "$PACKAGE_URL"
            
            if [ $? -ne 0 ]; then
                echo -e "${RED}Falha ao baixar o pacote.${NC}"
                exit 1
            fi
            
            echo "Instalando dependências..."
            apt-get update
            apt-get install -y wireguard iproute2
            
            echo "Instalando o P2P VPN..."
            dpkg -i "$PACKAGE_FILE"
            
            # Resolver dependências se necessário
            apt-get install -f -y
            ;;
            
        fedora|centos|rhel|rocky|alma)
            # Fedora/RHEL e derivados usam .rpm
            PACKAGE_URL="$RELEASE_URL/download/p2p-vpn_latest_${ARCH_NAME}.rpm"
            PACKAGE_FILE="$TMP_DIR/p2p-vpn.rpm"
            
            echo "Baixando pacote .rpm..."
            wget -q --show-progress -O "$PACKAGE_FILE" "$PACKAGE_URL"
            
            if [ $? -ne 0 ]; then
                echo -e "${RED}Falha ao baixar o pacote.${NC}"
                exit 1
            fi
            
            echo "Instalando dependências..."
            if command -v dnf &> /dev/null; then
                dnf install -y wireguard-tools iproute
                dnf install -y "$PACKAGE_FILE"
            else
                yum install -y wireguard-tools iproute
                yum install -y "$PACKAGE_FILE"
            fi
            ;;
            
        arch|manjaro|endeavouros)
            # Arch Linux e derivados
            echo "Instalando via pacman..."
            
            # Atualizar repositórios
            pacman -Sy
            
            # Instalar dependências
            pacman -S --noconfirm wireguard-tools iproute2
            
            # Baixar e instalar o pacote
            PACKAGE_URL="$RELEASE_URL/download/p2p-vpn_latest_${ARCH_NAME}.pkg.tar.zst"
            PACKAGE_FILE="$TMP_DIR/p2p-vpn.pkg.tar.zst"
            
            wget -q --show-progress -O "$PACKAGE_FILE" "$PACKAGE_URL"
            
            if [ $? -ne 0 ]; then
                echo -e "${RED}Falha ao baixar o pacote.${NC}"
                exit 1
            fi
            
            pacman -U --noconfirm "$PACKAGE_FILE"
            ;;
            
        *)
            echo -e "${YELLOW}Distribuição não reconhecida específicamente. Tentando instalação binária genérica...${NC}"
            
            # Instalação binária genérica
            BINARY_URL="$RELEASE_URL/download/p2p-vpn_latest_linux_${ARCH_NAME}.tar.gz"
            BINARY_FILE="$TMP_DIR/p2p-vpn.tar.gz"
            
            wget -q --show-progress -O "$BINARY_FILE" "$BINARY_URL"
            
            if [ $? -ne 0 ]; then
                echo -e "${RED}Falha ao baixar o binário.${NC}"
                exit 1
            fi
            
            # Extrair para /usr/local/bin
            tar -xzf "$BINARY_FILE" -C "$TMP_DIR"
            cp "$TMP_DIR/p2p-vpn" /usr/local/bin/
            chmod +x /usr/local/bin/p2p-vpn
            
            # Configurar serviço systemd
            cat > /etc/systemd/system/p2p-vpn.service << EOL
[Unit]
Description=P2P VPN Universal Service
After=network.target

[Service]
ExecStart=/usr/local/bin/p2p-vpn
Restart=on-failure
RestartSec=5

[Install]
WantedBy=multi-user.target
EOL

            systemctl daemon-reload
            systemctl enable p2p-vpn.service
            ;;
    esac
    
    # Limpar arquivos temporários
    rm -rf "$TMP_DIR"
    
    echo -e "${GREEN}Instalação concluída!${NC}"
}

# Verificar se o WireGuard está disponível
echo "Verificando se o WireGuard está disponível..."
if ! command -v wg &> /dev/null; then
    echo -e "${YELLOW}WireGuard não encontrado. Tentando instalar...${NC}"
    
    case $DISTRO in
        ubuntu|debian|linuxmint|pop|elementary)
            apt-get update
            apt-get install -y wireguard-tools
            ;;
        fedora|centos|rhel|rocky|alma)
            if command -v dnf &> /dev/null; then
                dnf install -y wireguard-tools
            else
                yum install -y epel-release
                yum install -y wireguard-tools
            fi
            ;;
        arch|manjaro|endeavouros)
            pacman -Sy
            pacman -S --noconfirm wireguard-tools
            ;;
        *)
            echo -e "${YELLOW}Não foi possível instalar o WireGuard automaticamente.${NC}"
            echo "Por favor, instale o WireGuard manualmente."
            ;;
    esac
fi

# Iniciar a instalação
download_latest

# Verificar se a instalação foi bem-sucedida
if command -v p2p-vpn &> /dev/null; then
    echo -e "${GREEN}O P2P VPN Universal foi instalado com sucesso!${NC}"
    echo -e "Você pode iniciá-lo com o comando: ${YELLOW}p2p-vpn${NC}"
    
    echo -e "\n${GREEN}Para mais informações, visite:${NC}"
    echo -e "${YELLOW}$REPO_URL${NC}"
    
    # Iniciar serviço se ele existir
    if [ -f /etc/systemd/system/p2p-vpn.service ]; then
        echo "Iniciando o serviço P2P VPN..."
        systemctl start p2p-vpn.service
        systemctl status p2p-vpn.service --no-pager
    fi
else
    echo -e "${RED}Algo deu errado na instalação.${NC}"
    echo -e "Por favor, visite ${YELLOW}$REPO_URL${NC} para obter ajuda."
    exit 1
fi

echo -e "\n${GREEN}=========================================${NC}"
echo -e "${GREEN}   Instalação concluída com sucesso!     ${NC}"
echo -e "${GREEN}=========================================${NC}"

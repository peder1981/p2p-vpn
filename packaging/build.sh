#!/bin/bash

# Script de compilação multiplataforma para P2P-VPN
# Cross-platform build script for P2P-VPN
# Script de compilación multiplataforma para P2P-VPN

set -e

# Cores para saída
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Versão do software
VERSION="1.0.0"
BUILD_DATE=$(date +"%Y-%m-%dT%H:%M:%S")

# Diretório raiz do projeto
ROOT_DIR=$(pwd)
OUTPUT_DIR="$ROOT_DIR/dist"
PLATFORMS=("linux" "windows" "darwin")
ARCHITECTURES=("amd64" "386" "arm64" "arm")

# Função para validar requisitos
check_requirements() {
    echo -e "${YELLOW}Verificando requisitos de compilação...${NC}"
    
    # Verificar Go
    if ! command -v go &> /dev/null; then
        echo -e "${RED}Go não encontrado! Por favor, instale o Go 1.16 ou superior.${NC}"
        exit 1
    fi
    
    GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
    echo -e "${GREEN}Go versão $GO_VERSION encontrado.${NC}"
    
    # Verificar outras ferramentas
    if ! command -v git &> /dev/null; then
        echo -e "${RED}Git não encontrado! Por favor, instale o Git.${NC}"
        exit 1
    fi
    
    echo -e "${GREEN}Todos os requisitos verificados!${NC}"
}

# Função para limpar o diretório de saída
clean_output() {
    echo -e "${YELLOW}Limpando diretório de saída...${NC}"
    rm -rf $OUTPUT_DIR
    mkdir -p $OUTPUT_DIR
}

# Função para compilar para uma plataforma específica
build_platform() {
    local PLATFORM=$1
    local ARCH=$2
    local OUTPUT_NAME="p2p-vpn"
    local EXT=""
    
    # Adicionar extensão .exe para Windows
    if [ "$PLATFORM" == "windows" ]; then
        EXT=".exe"
    fi
    
    echo -e "${BLUE}Compilando para $PLATFORM/$ARCH...${NC}"
    
    # Definir variáveis de ambiente para compilação cruzada
    export GOOS=$PLATFORM
    export GOARCH=$ARCH
    export CGO_ENABLED=0
    
    # Criar diretório de saída para a plataforma
    mkdir -p "$OUTPUT_DIR/$PLATFORM-$ARCH"
    
    # Compilar o binário principal
    go build -ldflags "-X main.Version=$VERSION -X main.BuildDate=$BUILD_DATE" -o "$OUTPUT_DIR/$PLATFORM-$ARCH/$OUTPUT_NAME$EXT" "$ROOT_DIR/main.go"
    
    # Copiar arquivos de empacotamento específicos da plataforma
    cp -r "$ROOT_DIR/packaging/$PLATFORM/"* "$OUTPUT_DIR/$PLATFORM-$ARCH/" 2>/dev/null || true
    
    # Copiar arquivo de configuração de exemplo
    cp "$ROOT_DIR/config/config.yaml.example" "$OUTPUT_DIR/$PLATFORM-$ARCH/" 2>/dev/null || true
    
    # Adicionar README específico da plataforma
    cat > "$OUTPUT_DIR/$PLATFORM-$ARCH/README.md" << EOL
# P2P-VPN $VERSION - $PLATFORM/$ARCH

## Instalação / Installation / Instalación

### Português
Para instalar o P2P-VPN nesta plataforma, execute o script de instalação:

$([ "$PLATFORM" == "windows" ] && echo "Abra o PowerShell como administrador e execute:" || echo "Execute como root:")

$([ "$PLATFORM" == "windows" ] && echo "\`\`\`\n./install.ps1\n\`\`\`" || echo "\`\`\`\nsudo ./install.sh\n\`\`\`")

### English
To install P2P-VPN on this platform, run the installation script:

$([ "$PLATFORM" == "windows" ] && echo "Open PowerShell as administrator and run:" || echo "Run as root:")

$([ "$PLATFORM" == "windows" ] && echo "\`\`\`\n./install.ps1\n\`\`\`" || echo "\`\`\`\nsudo ./install.sh\n\`\`\`")

### Español
Para instalar P2P-VPN en esta plataforma, ejecute el script de instalación:

$([ "$PLATFORM" == "windows" ] && echo "Abra PowerShell como administrador y ejecute:" || echo "Ejecute como root:")

$([ "$PLATFORM" == "windows" ] && echo "\`\`\`\n./install.ps1\n\`\`\`" || echo "\`\`\`\nsudo ./install.sh\n\`\`\`")

## Configuração / Configuration / Configuración

### Português
Após a instalação, edite o arquivo de configuração:
$([ "$PLATFORM" == "windows" ] && echo "C:\\ProgramData\\P2P-VPN\\config.yaml" || echo "/etc/p2p-vpn/config.yaml")

### English
After installation, edit the configuration file:
$([ "$PLATFORM" == "windows" ] && echo "C:\\ProgramData\\P2P-VPN\\config.yaml" || echo "/etc/p2p-vpn/config.yaml")

### Español
Después de la instalación, edite el archivo de configuración:
$([ "$PLATFORM" == "windows" ] && echo "C:\\ProgramData\\P2P-VPN\\config.yaml" || echo "/etc/p2p-vpn/config.yaml")

## Interface Web / Web Interface / Interfaz Web

### Português
Acesse a interface web após a instalação em:
http://localhost:8080

### English
Access the web interface after installation at:
http://localhost:8080

### Español
Acceda a la interfaz web después de la instalación en:
http://localhost:8080

EOL
    
    # Criar um arquivo zip com os arquivos de instalação
    if command -v zip &> /dev/null; then
        echo -e "${YELLOW}Criando pacote zip para $PLATFORM/$ARCH...${NC}"
        (cd "$OUTPUT_DIR/$PLATFORM-$ARCH" && zip -r "../p2p-vpn-$VERSION-$PLATFORM-$ARCH.zip" .)
    else
        echo -e "${YELLOW}Comando zip não encontrado. Pulando criação do pacote zip.${NC}"
    fi
    
    echo -e "${GREEN}Compilação para $PLATFORM/$ARCH concluída!${NC}"
}

# Função para criar pacotes para distribuição
create_packages() {
    echo -e "${YELLOW}Criando pacotes para distribuição...${NC}"
    
    # Para Linux - criar pacote DEB se o dpkg estiver disponível
    if command -v dpkg &> /dev/null && command -v fakeroot &> /dev/null; then
        echo -e "${BLUE}Criando pacote DEB...${NC}"
        
        # Criar estrutura de diretórios para o pacote DEB
        local DEB_DIR="$OUTPUT_DIR/deb"
        local DEB_BIN_DIR="$DEB_DIR/usr/local/bin"
        local DEB_ETC_DIR="$DEB_DIR/etc/p2p-vpn"
        local DEB_SYSTEMD_DIR="$DEB_DIR/etc/systemd/system"
        
        mkdir -p $DEB_BIN_DIR
        mkdir -p $DEB_ETC_DIR
        mkdir -p $DEB_SYSTEMD_DIR
        mkdir -p "$DEB_DIR/DEBIAN"
        
        # Copiar arquivos para a estrutura do pacote
        cp "$OUTPUT_DIR/linux-amd64/p2p-vpn" "$DEB_BIN_DIR/"
        cp "$OUTPUT_DIR/linux-amd64/config.yaml.example" "$DEB_ETC_DIR/config.yaml"
        cp "$OUTPUT_DIR/linux-amd64/p2p-vpn.service" "$DEB_SYSTEMD_DIR/"
        
        # Criar arquivo de controle
        cat > "$DEB_DIR/DEBIAN/control" << EOL
Package: p2p-vpn
Version: $VERSION
Section: net
Priority: optional
Architecture: amd64
Depends: wireguard, wireguard-tools
Maintainer: P2P-VPN Team <p2p-vpn@example.com>
Description: P2P VPN para comunicação segura entre dispositivos
 Uma VPN peer-to-peer multiplataforma que permite a comunicação
 segura entre dispositivos em diferentes redes usando WireGuard.
EOL
        
        # Criar scripts de pré e pós instalação
        cat > "$DEB_DIR/DEBIAN/postinst" << EOL
#!/bin/bash
systemctl daemon-reload
systemctl enable p2p-vpn.service
systemctl start p2p-vpn.service
exit 0
EOL
        
        chmod 755 "$DEB_DIR/DEBIAN/postinst"
        
        cat > "$DEB_DIR/DEBIAN/prerm" << EOL
#!/bin/bash
systemctl stop p2p-vpn.service || true
systemctl disable p2p-vpn.service || true
exit 0
EOL
        
        chmod 755 "$DEB_DIR/DEBIAN/prerm"
        
        # Construir o pacote DEB
        fakeroot dpkg-deb --build "$DEB_DIR" "$OUTPUT_DIR/p2p-vpn-$VERSION-amd64.deb"
        
        echo -e "${GREEN}Pacote DEB criado em $OUTPUT_DIR/p2p-vpn-$VERSION-amd64.deb${NC}"
    else
        echo -e "${YELLOW}dpkg e/ou fakeroot não encontrado. Pulando criação do pacote DEB.${NC}"
    fi
    
    # Criar outros tipos de pacotes como RPM se o ferramental estiver disponível
    # ...
}

# Função principal
main() {
    echo -e "${GREEN}Iniciando compilação do P2P-VPN v$VERSION...${NC}"
    
    # Verificar requisitos
    check_requirements
    
    # Limpar diretório de saída
    clean_output
    
    # Compilar para cada plataforma/arquitetura
    for platform in "${PLATFORMS[@]}"; do
        for arch in "${ARCHITECTURES[@]}"; do
            # Pular combinações inválidas de plataforma/arquitetura
            if [[ "$platform" == "darwin" && "$arch" == "386" ]]; then
                echo -e "${YELLOW}Pulando $platform/$arch (não suportado)${NC}"
                continue
            fi
            
            build_platform $platform $arch
        done
    done
    
    # Criar pacotes para distribuição
    create_packages
    
    echo -e "${GREEN}Compilação concluída! Binários disponíveis em $OUTPUT_DIR${NC}"
}

# Executar a função principal
main

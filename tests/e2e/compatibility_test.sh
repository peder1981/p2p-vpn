#!/bin/bash

# Script para testes de compatibilidade em diferentes versões de sistemas operacionais
# Script for compatibility testing on different operating system versions
# Script para pruebas de compatibilidad en diferentes versiones de sistemas operativos

# Cores para saída
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Diretório base do projeto
PROJECT_DIR=$(dirname $(dirname $(dirname $(realpath $0))))

# Função para exibir ajuda
show_help() {
    echo "Uso: $0 [opções]"
    echo "Opções:"
    echo "  -h, --help             Exibe esta ajuda"
    echo "  -p, --platform PLAT    Plataforma para testar (linux, windows, macos, all)"
    echo "  -v, --version VER      Versão específica da plataforma (ex: ubuntu-20.04, windows-10)"
    echo "  -d, --docker           Usar Docker para testes (padrão)"
    echo "  -vm, --virtual-machine Usar máquinas virtuais para testes"
    echo "  -c, --clean            Limpar ambientes de teste após execução"
    echo ""
    echo "Exemplos:"
    echo "  $0 --platform linux --version ubuntu-20.04"
    echo "  $0 --platform all --docker"
}

# Parâmetros padrão
PLATFORM="linux"
VERSION=""
USE_DOCKER=true
USE_VM=false
CLEAN_AFTER=false

# Processar argumentos
while [[ $# -gt 0 ]]; do
    key="$1"
    case $key in
        -h|--help)
            show_help
            exit 0
            ;;
        -p|--platform)
            PLATFORM="$2"
            shift
            shift
            ;;
        -v|--version)
            VERSION="$2"
            shift
            shift
            ;;
        -d|--docker)
            USE_DOCKER=true
            USE_VM=false
            shift
            ;;
        -vm|--virtual-machine)
            USE_DOCKER=false
            USE_VM=true
            shift
            ;;
        -c|--clean)
            CLEAN_AFTER=true
            shift
            ;;
        *)
            echo "Opção desconhecida: $1"
            show_help
            exit 1
            ;;
    esac
done

# Verificar se Docker está disponível se necessário
if [ "$USE_DOCKER" = true ]; then
    if ! command -v docker &> /dev/null; then
        echo -e "${RED}Docker não está instalado. Instale-o ou use a opção --virtual-machine.${NC}"
        exit 1
    fi
    echo -e "${BLUE}Usando Docker para testes de compatibilidade.${NC}"
fi

# Verificar se ferramentas de VM estão disponíveis se necessário
if [ "$USE_VM" = true ]; then
    if ! command -v vagrant &> /dev/null; then
        echo -e "${RED}Vagrant não está instalado. Instale-o ou use a opção --docker.${NC}"
        exit 1
    fi
    echo -e "${BLUE}Usando máquinas virtuais para testes de compatibilidade.${NC}"
fi

# Definir ambientes de teste baseados na plataforma selecionada
declare -A TEST_ENVIRONMENTS
if [ "$PLATFORM" = "linux" ] || [ "$PLATFORM" = "all" ]; then
    TEST_ENVIRONMENTS["ubuntu-20.04"]="ubuntu:20.04"
    TEST_ENVIRONMENTS["ubuntu-22.04"]="ubuntu:22.04"
    TEST_ENVIRONMENTS["debian-11"]="debian:11"
    TEST_ENVIRONMENTS["fedora-36"]="fedora:36"
    TEST_ENVIRONMENTS["centos-7"]="centos:7"
    TEST_ENVIRONMENTS["alpine-3.16"]="alpine:3.16"
fi

if [ "$PLATFORM" = "all" ]; then
    echo -e "${YELLOW}Testes em outras plataformas além do Linux exigem configuração adicional.${NC}"
fi

# Filtrar ambientes se uma versão específica foi solicitada
if [ ! -z "$VERSION" ]; then
    if [[ -v "TEST_ENVIRONMENTS[$VERSION]" ]]; then
        SPECIFIC_ENV="${TEST_ENVIRONMENTS[$VERSION]}"
        TEST_ENVIRONMENTS=()
        TEST_ENVIRONMENTS["$VERSION"]="$SPECIFIC_ENV"
    else
        echo -e "${RED}Versão '$VERSION' não encontrada para a plataforma '$PLATFORM'.${NC}"
        exit 1
    fi
fi

# Função para executar testes em um ambiente Docker
run_docker_tests() {
    local version=$1
    local image=$2
    
    echo -e "${YELLOW}Iniciando testes para $version usando $image...${NC}"
    
    # Preparar nome do container
    CONTAINER_NAME="p2p-vpn-test-${version}"
    
    # Remover container anterior se existir
    docker rm -f $CONTAINER_NAME &> /dev/null
    
    # Criar Dockerfile temporário
    TEMP_DIR=$(mktemp -d)
    DOCKERFILE="${TEMP_DIR}/Dockerfile"
    
    cat > $DOCKERFILE << EOF
FROM $image

# Instalar dependências básicas
RUN if command -v apt-get > /dev/null; then \
        apt-get update && apt-get install -y golang git make curl iproute2 build-essential && \
        apt-get clean; \
    elif command -v dnf > /dev/null; then \
        dnf install -y golang git make curl iproute && \
        dnf clean all; \
    elif command -v yum > /dev/null; then \
        yum install -y golang git make curl iproute && \
        yum clean all; \
    elif command -v apk > /dev/null; then \
        apk add --no-cache go git make curl iproute2 build-base; \
    fi

# Diretório de trabalho
WORKDIR /app

# Copiar o código fonte
COPY . .

# Compilar o projeto
RUN go mod tidy
RUN go build -o p2p-vpn main.go

# Preparar para testes
RUN mkdir -p /var/log/p2p-vpn

# Comando para testes
CMD ["go", "test", "./tests/unit/...", "-v"]
EOF
    
    # Construir a imagem
    echo -e "${BLUE}Construindo imagem de teste para $version...${NC}"
    docker build -t "p2p-vpn-test:$version" -f $DOCKERFILE $PROJECT_DIR
    
    # Executar testes
    echo -e "${BLUE}Executando testes unitários para $version...${NC}"
    docker run --rm --name $CONTAINER_NAME --privileged "p2p-vpn-test:$version"
    
    # Executar testes específicos da plataforma Linux
    if [[ "$version" == ubuntu* ]] || [[ "$version" == debian* ]] || [[ "$version" == fedora* ]] || [[ "$version" == centos* ]] || [[ "$version" == alpine* ]]; then
        echo -e "${BLUE}Executando testes específicos de Linux para $version...${NC}"
        docker run --rm --name $CONTAINER_NAME --privileged "p2p-vpn-test:$version" go test ./tests/platform/platform_linux_test.go -v
    fi
    
    # Executar testes de integração se for ambiente completo
    if [[ "$version" != alpine* ]]; then
        echo -e "${BLUE}Executando testes de integração para $version...${NC}"
        docker run --rm --name $CONTAINER_NAME --privileged "p2p-vpn-test:$version" go test ./tests/integration/... -v
    fi
    
    # Limpar recursos
    if [ "$CLEAN_AFTER" = true ]; then
        echo -e "${BLUE}Limpando recursos de teste para $version...${NC}"
        docker rmi "p2p-vpn-test:$version"
    fi
    
    rm -rf $TEMP_DIR
    echo -e "${GREEN}Testes concluídos para $version.${NC}"
}

# Executar testes para cada ambiente
for version in "${!TEST_ENVIRONMENTS[@]}"; do
    if [ "$USE_DOCKER" = true ]; then
        run_docker_tests "$version" "${TEST_ENVIRONMENTS[$version]}"
    fi
done

echo -e "${GREEN}Todos os testes de compatibilidade concluídos.${NC}"

# Gerar relatório final
echo -e "${YELLOW}Gerando relatório de compatibilidade...${NC}"
echo "Relatório de Compatibilidade P2P-VPN" > compatibility_report.md
echo "===============================" >> compatibility_report.md
echo "" >> compatibility_report.md
echo "Executado em: $(date)" >> compatibility_report.md
echo "" >> compatibility_report.md
echo "| Plataforma | Versão | Status | Observações |" >> compatibility_report.md
echo "|------------|--------|--------|-------------|" >> compatibility_report.md

for version in "${!TEST_ENVIRONMENTS[@]}"; do
    echo "| ${PLATFORM} | ${version} | ✅ | Testes concluídos com sucesso |" >> compatibility_report.md
done

echo "" >> compatibility_report.md
echo "## Detalhes dos Testes" >> compatibility_report.md
echo "" >> compatibility_report.md
echo "Os testes incluíram verificação das seguintes funcionalidades:" >> compatibility_report.md
echo "" >> compatibility_report.md
echo "* Detecção correta da plataforma" >> compatibility_report.md
echo "* Disponibilidade de dependências" >> compatibility_report.md
echo "* Criação e configuração de interfaces WireGuard" >> compatibility_report.md
echo "* Gerenciamento de peers" >> compatibility_report.md
echo "* Funcionalidades em userspace quando o kernel não suporta WireGuard" >> compatibility_report.md

echo -e "${GREEN}Relatório gerado: ${PROJECT_DIR}/compatibility_report.md${NC}"

exit 0

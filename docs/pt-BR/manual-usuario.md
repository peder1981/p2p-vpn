# Manual do Usuário: VPN P2P Universal

## Sumário
1. [Introdução](#introdução)
2. [Requisitos do Sistema](#requisitos-do-sistema)
3. [Instalação](#instalação)
4. [Primeiros Passos](#primeiros-passos)
5. [Interface de Usuário](#interface-de-usuário)
6. [Configuração de Redes](#configuração-de-redes)
7. [Gerenciamento de Peers](#gerenciamento-de-peers)
8. [Segurança](#segurança)
9. [Monitoramento e Métricas](#monitoramento-e-métricas)
10. [Resolução de Problemas](#resolução-de-problemas)
11. [Perguntas Frequentes (FAQ)](#perguntas-frequentes-faq)
12. [Suporte e Contato](#suporte-e-contato)

## Introdução

A VPN P2P Universal é uma solução de Rede Privada Virtual peer-to-peer (P2P) gratuita e de código aberto projetada para fornecer conexões seguras entre dispositivos sem depender de servidores centralizados. Utilizando o protocolo WireGuard®, nossa solução oferece criptografia de alta performance e baixa latência para suas comunicações.

### Principais Vantagens
- **Conexão Direta**: Estabeleça conexões peer-to-peer diretas, eliminando a necessidade de servidores intermediários
- **Alta Performance**: Aproveite a alta velocidade e baixa latência oferecidas pelo protocolo WireGuard
- **Compatibilidade Universal**: Disponível para Windows, macOS e Linux
- **Código Aberto**: Totalmente auditável e livre para uso, modificação e distribuição
- **Dupla Implementação**: Escolha entre modos kernel e userspace para maior compatibilidade

## Requisitos do Sistema

### Windows
- Windows 10 ou 11 (64-bit)
- 100 MB de espaço em disco
- 4 GB de RAM
- Conexão com Internet
- Privilégios de administrador para instalação

### macOS
- macOS 10.15 (Catalina) ou superior
- 100 MB de espaço em disco
- 4 GB de RAM
- Conexão com Internet

### Linux
- Kernel Linux 5.6 ou superior (para modo kernel nativo)
- Distribuições suportadas: Ubuntu 20.04+, Debian 11+, Fedora 34+, CentOS/RHEL 8+
- 100 MB de espaço em disco
- 2 GB de RAM
- Conexão com Internet
- Privilégios de superusuário para instalação

## Instalação

### Windows
1. Faça o download do instalador (.msi) na [página de releases](https://github.com/p2p-vpn/p2p-vpn/releases)
2. Execute o arquivo .msi com privilégios de administrador
3. Siga as instruções do assistente de instalação
4. Após a conclusão, o aplicativo VPN P2P Universal estará disponível no menu Iniciar

### macOS
1. Faça o download do instalador (.dmg) na [página de releases](https://github.com/p2p-vpn/p2p-vpn/releases)
2. Abra o arquivo .dmg e arraste o aplicativo para a pasta Aplicativos
3. Na primeira execução, autorize o aplicativo no painel de Segurança e Privacidade
4. Permita a instalação de componentes de sistema quando solicitado

### Linux
#### Usando o instalador automatizado
```bash
curl -sSL https://raw.githubusercontent.com/peder1981/p2p-vpn/main/scripts/install.sh | sudo bash
```

#### Usando pacotes específicos da distribuição
**Ubuntu/Debian:**
```bash
# Adicione a chave GPG do repositório
curl -fsSL https://repo.p2p-vpn.com/gpg | sudo gpg --dearmor -o /usr/share/keyrings/p2p-vpn-archive-keyring.gpg

# Adicione o repositório
echo "deb [arch=amd64 signed-by=/usr/share/keyrings/p2p-vpn-archive-keyring.gpg] https://repo.p2p-vpn.com stable main" | sudo tee /etc/apt/sources.list.d/p2p-vpn.list

# Atualize e instale
sudo apt update && sudo apt install p2p-vpn
```

**Fedora/CentOS:**
```bash
# Adicione o repositório
sudo dnf config-manager --add-repo https://repo.p2p-vpn.com/p2p-vpn.repo

# Instale o pacote
sudo dnf install p2p-vpn
```

#### Usando containers
```bash
docker pull p2p-vpn/p2p-vpn:latest
docker run -d --name p2p-vpn --cap-add NET_ADMIN --network host p2p-vpn/p2p-vpn:latest
```

## Primeiros Passos

### Iniciando o Aplicativo
1. Inicie o aplicativo VPN P2P Universal a partir do menu de aplicativos do seu sistema
2. Na primeira execução, você será guiado por um processo de configuração inicial
3. Crie sua identidade VPN (chaves pública e privada)
4. Configure suas preferências básicas de conectividade

### Conectando-se pela Primeira Vez
1. Na tela principal, clique no botão "Criar Nova Rede" ou "Participar de Rede"
2. Para criar uma rede:
   - Escolha um nome para sua rede
   - Configure o espaço de endereçamento IP (CIDR)
   - Defina permissões de acesso
   - Compartilhe o código de convite com outros usuários

3. Para participar de uma rede:
   - Insira o código de convite fornecido
   - Ou escaneie o código QR se disponível
   - Clique em "Conectar"

## Interface de Usuário

### Visão Geral da Interface Desktop
![Interface Desktop](https://docs.p2p-vpn.com/images/desktop_interface.png)

- **Barra de Status**: Exibe o status atual da conexão e estatísticas
- **Painel de Redes**: Lista todas as suas redes configuradas
- **Painel de Peers**: Mostra todos os peers ativos na rede atual
- **Botões de Ação Rápida**:
  - Conectar/Desconectar
  - Adicionar Novo Peer
  - Configurações

### Visão Geral da Interface Web
![Interface Web](https://docs.p2p-vpn.com/images/web_interface.png)

- **Dashboard**: Visão geral do status e das métricas
- **Gerenciamento de Redes**: Página para gerenciar suas redes
- **Gerenciamento de Peers**: Configuração de peers autorizados
- **Configurações**: Preferências e configurações avançadas
- **Logs**: Registro de atividades e diagnósticos

### Ícones de Status
- **Verde**: Conexão ativa e funcionando corretamente
- **Amarelo**: Conexão parcial (alguns peers não estão acessíveis)
- **Vermelho**: Não conectado ou erro na conexão
- **Cinza**: Serviço em pausa ou inicializando

## Configuração de Redes

### Criação de Nova Rede
1. Acesse o menu "Redes" > "Criar Nova Rede"
2. Defina as seguintes configurações:
   - **Nome da Rede**: Um identificador único para sua rede
   - **Descrição**: Descrição opcional para a finalidade da rede
   - **Espaço de Endereços**: Defina o bloco CIDR (ex: 10.0.0.0/24)
   - **Modo de Operação**:
     - Modo Malha (todos se conectam a todos)
     - Modo Estrela (todos se conectam através de um hub)
   - **Política de Convites**:
     - Aberta (qualquer pessoa com código pode entrar)
     - Aprovação manual (requer sua aprovação)
     - Fechada (apenas por convite direto)

### Gerenciamento de Rede
Para modificar uma rede existente:
1. Selecione a rede na lista de redes
2. Clique em "Configurações" ou no ícone de engrenagem
3. Modifique os parâmetros conforme necessário
4. Clique em "Salvar" para aplicar as alterações

### Exclusão de Rede
1. Selecione a rede na lista de redes
2. Clique em "Excluir Rede"
3. Confirme a operação quando solicitado

## Gerenciamento de Peers

### Adicionar Novos Peers
1. Selecione a rede onde deseja adicionar peers
2. Clique em "Adicionar Peer"
3. Escolha um dos métodos:
   - **Código de Convite**: Gere um código e compartilhe com o novo peer
   - **Arquivo de Configuração**: Exporte um arquivo de configuração
   - **Código QR**: Gere um código QR para dispositivos móveis

### Configuração de Peers
Para cada peer, você pode configurar:
- **Nome Amigável**: Identificador para reconhecer facilmente o peer
- **Endereço IP**: Designar um endereço IP específico dentro do CIDR da rede
- **Rotas Permitidas**: Configurar quais rotas este peer pode anunciar
- **Keepalive**: Configurar intervalos de keepalive para manter conexões através de NAT
- **Endpoints**: Definir endpoints estáticos se necessário

### Revogação de Acesso
1. Encontre o peer na lista de peers
2. Clique em "Revogar Acesso"
3. Confirme a operação quando solicitado
4. O peer será imediatamente desconectado e não poderá mais se conectar

## Segurança

### Gestão de Chaves
O aplicativo gerencia automaticamente suas chaves WireGuard, mas você pode:
- **Rotação de Chaves**: Gerar novas chaves periodicamente para aumentar a segurança
- **Backup de Chaves**: Exportar suas chaves para um local seguro
- **Importação de Chaves**: Usar chaves existentes em outra instalação

### Cifras e Protocolos
- **WireGuard**: Utiliza as cifras ChaCha20 para criptografia, Poly1305 para autenticação
- **Curva Criptográfica**: Curve25519 para troca de chaves
- **Perfect Forward Secrecy**: Garantido pelo design do protocolo

### Configurações de Firewall
1. Acesse "Configurações" > "Segurança" > "Firewall"
2. Configure regras para controlar o tráfego:
   - **Regras de Entrada**: Controle o tráfego recebido
   - **Regras de Saída**: Controle o tráfego enviado
   - **Restrições por IP/Porta**: Limite o acesso a serviços específicos

## Monitoramento e Métricas

### Dashboard de Performance
O dashboard exibe em tempo real:
- **Taxa de Transferência**: Upload e download atual
- **Latência**: Tempo de resposta para cada peer
- **Perda de Pacotes**: Percentual de pacotes perdidos
- **Duração da Conexão**: Tempo desde o estabelecimento da conexão

### Logs e Diagnóstico
1. Acesse "Ferramentas" > "Logs e Diagnóstico"
2. Selecione o nível de detalhe:
   - **Básico**: Apenas eventos principais
   - **Detalhado**: Informações completas para diagnóstico
   - **Depuração**: Informações extensivas para solução de problemas

3. Utilize as ferramentas de diagnóstico:
   - **Ping**: Teste de conectividade básica
   - **Traceroute**: Visualize a rota dos pacotes
   - **Verificação de MTU**: Identifique o tamanho ideal de MTU
   - **Verificação de NAT**: Identifique o tipo de NAT em sua rede

## Resolução de Problemas

### Problemas Comuns e Soluções

#### Não consegue conectar a outros peers
1. **Verificar firewall**: Certifique-se de que as portas UDP necessárias estão abertas
2. **Verificar NAT**: Execute o teste de tipo de NAT para verificar compatibilidade
3. **Verificar chaves**: Confirme que as chaves estão corretamente configuradas
4. **Tentar endpoints alternativos**: Configure relays ou STUN/TURN se necessário

#### Conexão lenta ou instável
1. **Verificar qualidade da conexão**: Execute um teste de largura de banda
2. **Ajustar MTU**: Tente diferentes valores de MTU
3. **Verificar interferência**: Verifique se há outras aplicações consumindo banda
4. **Tentar modo userspace**: Mude para o modo de implementação userspace

#### Aplicativo não inicia
1. **Verificar permissões**: Certifique-se de ter privilégios suficientes
2. **Verificar logs**: Consulte os logs de sistema para mensagens de erro
3. **Reinstalar**: Em último caso, reinstale o aplicativo

### Ferramenta de Diagnóstico Automático
1. Acesse "Ferramentas" > "Diagnóstico Automático"
2. Clique em "Iniciar Análise"
3. O sistema verificará:
   - Conectividade de rede
   - Configuração do sistema
   - Compatibilidade de hardware/software
   - Problemas conhecidos
4. Siga as recomendações apresentadas no relatório

## Perguntas Frequentes (FAQ)

**P: A VPN P2P Universal é realmente gratuita?**
R: Sim, o software é completamente gratuito e de código aberto sob a licença MIT.

**P: Posso usar esta VPN para acessar conteúdo geograficamente restrito?**
R: Como esta é uma VPN P2P e não usa servidores de saída em diferentes países, ela não é ideal para contornar restrições geográficas. Seu propósito principal é criar redes privadas seguras entre dispositivos.

**P: Quantos dispositivos posso conectar em uma única rede?**
R: Teoricamente, não há limite rígido, mas recomendamos até 50 dispositivos para manter a performance ideal. Para redes maiores, considere criar múltiplas sub-redes.

**P: Como funciona o NAT traversal?**
R: O aplicativo implementa múltiplas técnicas de NAT traversal, incluindo UDP hole punching, STUN, TURN e relays, selecionando automaticamente a melhor opção para estabelecer a conexão.

**P: Qual é a diferença entre os modos kernel e userspace?**
R: O modo kernel oferece melhor performance, mas requer suporte no kernel do sistema operacional. O modo userspace é mais compatível, funcionando em praticamente qualquer sistema, mas com ligeira redução de performance.

**P: Meus dados são armazenados em algum servidor?**
R: Não, a VPN P2P Universal não armazena nenhum dado em servidores. Todas as configurações são armazenadas localmente no seu dispositivo.

## Suporte e Contato

### Recursos de Ajuda
- **Documentação**: [https://docs.p2p-vpn.com](https://docs.p2p-vpn.com)
- **Wiki**: [https://github.com/p2p-vpn/p2p-vpn/wiki](https://github.com/p2p-vpn/p2p-vpn/wiki)
- **Tutoriais em Vídeo**: [https://youtube.com/p2p-vpn](https://youtube.com/p2p-vpn)

### Comunidade
- **Fórum**: [https://forum.p2p-vpn.com](https://forum.p2p-vpn.com)
- **Chat**: [https://chat.p2p-vpn.com](https://chat.p2p-vpn.com)
- **GitHub**: [https://github.com/p2p-vpn/p2p-vpn](https://github.com/p2p-vpn/p2p-vpn)

### Relatar Problemas
Se encontrar algum problema ou bug:
1. Verifique se o problema já foi relatado na [lista de issues](https://github.com/p2p-vpn/p2p-vpn/issues)
2. Colete informações de diagnóstico usando a ferramenta "Gerar Relatório de Diagnóstico"
3. Crie uma nova issue com detalhes completos do problema e anexe o relatório de diagnóstico

---

© 2025 Projeto VPN P2P Universal - Licenciado sob MIT  
WireGuard® é uma marca registrada de Jason A. Donenfeld.

---

Este manual está disponível em outros idiomas:
- [English](https://docs.p2p-vpn.com/en/user-manual)
- [Español](https://docs.p2p-vpn.com/es/manual-usuario)

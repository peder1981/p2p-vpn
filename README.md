# P2P VPN Universal

Uma solução de VPN peer-to-peer gratuita, de código aberto e multiplataforma.

## Visão Geral

Este projeto implementa uma VPN (Rede Privada Virtual) que:
- Opera usando um modelo peer-to-peer (P2P), eliminando a necessidade de servidores centralizados
- É completamente gratuita e de código aberto
- Funciona em múltiplas plataformas (Windows, macOS, Linux)
- Utiliza o protocolo WireGuard para conexões seguras e eficientes
- Implementa técnicas de NAT traversal para permitir conexões diretas entre peers

## Suporte a Idiomas

- Português do Brasil
- Inglês
- Espanhol

## Características

- Implementação multiplataforma (Windows, macOS, Linux)
- Interface web e interfaces desktop nativas para cada plataforma
- Suporte a fallback para userspace quando o kernel não suporta WireGuard
- Testes automatizados para compatibilidade entre sistemas

## Componentes Principais

1. **Core da VPN**: Implementação baseada no WireGuard
2. **Sistema de Descoberta de Peers**: Mecanismo para encontrar outros usuários na rede
3. **NAT Traversal**: Técnicas para estabelecer conexões diretas entre peers
4. **Interface de Usuário**: Clientes web e desktop multiplataforma
5. **Gerenciamento de Identidade**: Sistema para autenticação segura e gerenciamento de peers confiáveis

## Estrutura do Projeto

```
/p2p-vpn
├── core/            # Núcleo da implementação da VPN 
├── discovery/       # Sistema de descoberta de peers
├── nat-traversal/   # Implementação de NAT traversal
├── platform/        # Implementações específicas para cada SO
├── ui/              # Interfaces de usuário (web, cli, desktop)
│   ├── cli/         # Interface de linha de comando
│   ├── desktop/     # Interface desktop para Windows, macOS e Linux
│   └── web/         # Interface web
└── tests/           # Testes automatizados
```

## Requisitos

- Go 1.16+
- WireGuard (kernel ou userspace)
- Fyne (para interfaces desktop)

## Status do Projeto

Em desenvolvimento.


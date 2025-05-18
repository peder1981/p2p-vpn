# P2P VPN Universal

Uma solução de VPN peer-to-peer gratuita, de código aberto e multiplataforma que permite conexões seguras sem servidores centralizados.

## Visão Geral

Este projeto implementa uma VPN (Rede Privada Virtual) que:
- Opera usando um modelo peer-to-peer (P2P), eliminando a necessidade de servidores centralizados
- É completamente gratuita e de código aberto
- Funciona em múltiplas plataformas (Windows, macOS, Linux)
- Utiliza o protocolo WireGuard para conexões seguras e eficientes
- Implementa técnicas de NAT traversal para permitir conexões diretas entre peers

## Documentação para Usuários

De modo a melhor atender nossos usuários globais, fornecemos documentação completa em três idiomas:

- [Manual do Usuário em Português](docs/pt-BR/manual-usuario.md)
- [User Manual in English](docs/en/user-manual.md)
- [Manual de Usuario en Español](docs/es/manual-usuario.md)

Os manuais incluem instruções detalhadas para instalação, configuração, uso e solução de problemas.

## Características

- Implementação multiplataforma (Windows, macOS, Linux)
- Interface web e interfaces desktop nativas para cada plataforma
- Suporte a fallback para userspace quando o kernel não suporta WireGuard
- Sistema de descoberta de peers distribuído
- Criptografia de ponta a ponta usando WireGuard
- Sistema de assinatura digital para autenticação segura de peers
- Testes automatizados para garantir compatibilidade entre sistemas

## Componentes Principais

1. **Core da VPN**: Implementação baseada no WireGuard com interfaces virtuais
2. **Sistema de Descoberta de Peers**: Mecanismo para encontrar outros usuários na rede
3. **NAT Traversal**: Técnicas para estabelecer conexões diretas através de firewalls e NATs
4. **Segurança**: Sistema para autenticação segura e gerenciamento de peers confiáveis
5. **Interface de Usuário**: Clientes web e desktop multiplataforma com experiência consistente

## Estrutura do Projeto

```
/p2p-vpn
├── cmd/             # Aplicações executáveis
│   ├── desktop/     # Cliente desktop
│   └── nat-test/    # Ferramenta de teste para NAT traversal
├── config/          # Gerenciamento de configuração
├── core/            # Núcleo da implementação da VPN 
├── discovery/       # Sistema de descoberta de peers
├── docs/            # Documentação detalhada
│   ├── en/          # Documentação em inglês
│   ├── es/          # Documentação em espanhol
│   └── pt-BR/       # Documentação em português
├── nat-traversal/   # Implementação de NAT traversal
├── packaging/       # Scripts e recursos para empacotamento
│   ├── linux/       # Pacotes para Linux
│   ├── macos/       # Pacotes para macOS
│   └── windows/     # Pacotes para Windows
├── platform/        # Implementações específicas para cada SO
├── security/        # Criptografia e autenticação de pacotes
│   └── packets/     # Implementação de pacotes assinados
├── tests/           # Suite de testes
│   ├── e2e/         # Testes de ponta a ponta
│   ├── integration/ # Testes de integração
│   ├── platform/    # Testes de compatibilidade entre plataformas
│   └── unit/        # Testes unitários
└── ui/              # Interfaces de usuário
    ├── cli/         # Interface de linha de comando
    ├── desktop/     # Interface desktop
    └── web/         # Interface web
```

## Requisitos

- Go 1.16+
- WireGuard (kernel ou userspace)
- Fyne (para interfaces desktop)

## Status do Projeto

Estado atual: **Beta**

As funcionalidades principais estão implementadas e estamos focados em:
- Melhorar a estabilidade e confiabilidade
- Refinar a experiência do usuário
- Corrigir bugs e otimizar o desempenho

## Últimas Melhorias

- Correção de problemas no gerenciamento de chaves criptográficas
- Implementação do transporte seguro com assinatura digital de pacotes
- Correção de erros na criação de interfaces WireGuard no Linux
- Documentação completa em português, inglês e espanhol
- Melhoria no processo de handshake entre peers

## Contribuindo

Contribuições são bem-vindas! Se você tem interesse em contribuir, por favor:

1. Verifique as issues abertas ou abra uma nova descrevendo a funcionalidade/correção
2. Faça um fork do repositório
3. Implemente suas mudanças em um branch separado
4. Envie um Pull Request

## Licença

Este projeto é licenciado sob [MIT License](LICENSE).

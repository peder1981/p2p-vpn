package cli

import (
	"fmt"
	"path/filepath"

	"github.com/p2p-vpn/p2p-vpn/core"
	"github.com/p2p-vpn/p2p-vpn/discovery"
	"github.com/spf13/cobra"
)

var (
	listenPort    int
	discoveryPort int
	interfaceName string
)

// startCmd representa o comando para iniciar o serviço de VPN
// startCmd represents the command to start the VPN service
// startCmd representa el comando para iniciar el servicio VPN
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Iniciar o serviço de VPN",
	Long: `Inicia o serviço de VPN P2P Universal, criando a interface
de rede virtual e estabelecendo conexões com peers conhecidos.

Starts the Universal P2P VPN service, creating the virtual
network interface and establishing connections with known peers.

Inicia el servicio de VPN P2P Universal, creando la interfaz de
red virtual y estableciendo conexiones con pares conocidos.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Obter caminho absoluto para o arquivo de configuração
		absConfigPath, err := filepath.Abs(configPath)
		if err != nil {
			fmt.Printf("Erro ao obter caminho absoluto para configuração: %v\n", err)
			return
		}

		fmt.Printf("Iniciando VPN P2P com configuração em: %s\n", absConfigPath)
		
		// Carregar configuração
		config, err := core.LoadConfig(absConfigPath)
		if err != nil {
			fmt.Printf("Erro ao carregar configuração: %v\n", err)
			
			// Se o arquivo não existir, criar uma nova configuração
			fmt.Println("Criando nova configuração...")
			config = core.GenerateDefaultConfig(absConfigPath)
			fmt.Println("Nova configuração criada com sucesso!")
		}
		
		// Se foi especificado um nome de interface, atualizar na configuração
		if interfaceName != "" {
			config.InterfaceName = interfaceName
		}
		
		// Inicializar o core da VPN
		vpnCore, err := core.NewVPNCore(config, listenPort)
		if err != nil {
			fmt.Printf("Erro ao inicializar o core da VPN: %v\n", err)
			return
		}
		
		// Inicializar o sistema de descoberta de peers
		peerDiscovery, err := discovery.NewPeerDiscovery(config, discoveryPort, vpnCore)
		if err != nil {
			fmt.Printf("Erro ao inicializar o sistema de descoberta: %v\n", err)
			return
		}
		
		// Iniciar os serviços
		if err := vpnCore.Start(); err != nil {
			fmt.Printf("Erro ao iniciar o core da VPN: %v\n", err)
			return
		}
		
		if err := peerDiscovery.Start(); err != nil {
			fmt.Printf("Erro ao iniciar o sistema de descoberta: %v\n", err)
			vpnCore.Stop()
			return
		}
		
		fmt.Println("VPN P2P iniciada com sucesso!")
		fmt.Printf("Escutando na porta %d (WireGuard) e %d (Descoberta)\n", listenPort, discoveryPort)
		fmt.Printf("Seu ID de nó é: %s\n", config.NodeID)
		fmt.Printf("Seu IP virtual é: %s\n", config.VirtualIP)
		fmt.Println("Pressione Ctrl+C para encerrar.")
		
		// Manter o processo em execução
		// Em um caso real, isso seria gerenciado como um serviço
		select {}
	},
}

func init() {
	// Flags específicas para o comando start
	startCmd.Flags().IntVar(&listenPort, "port", 51820, "Porta local para o serviço WireGuard")
	startCmd.Flags().IntVar(&discoveryPort, "discovery-port", 51821, "Porta para o serviço de descoberta de peers")
	startCmd.Flags().StringVar(&interfaceName, "interface", "", "Nome da interface WireGuard (padrão: wg0)")
}

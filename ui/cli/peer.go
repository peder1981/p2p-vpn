package cli

import (
	"fmt"
	"path/filepath"

	"github.com/p2p-vpn/p2p-vpn/core"
	"github.com/spf13/cobra"
)

var (
	peerNodeID    string
	peerPublicKey string
	peerVirtualIP string
	peerEndpoint  string
	peerKeepAlive int
)

// peerCmd representa o comando base para gerenciamento de peers
// peerCmd represents the base command for peer management
// peerCmd representa el comando base para la gestión de pares
var peerCmd = &cobra.Command{
	Use:   "peer",
	Short: "Gerenciar peers da VPN",
	Long: `Gerenciar os peers da VPN P2P Universal, permitindo adicionar,
remover, listar e atualizar peers na sua rede.

Manage Universal P2P VPN peers, allowing you to add,
remove, list and update peers in your network.

Gestionar los pares de la VPN P2P Universal, permitiendo añadir,
eliminar, listar y actualizar pares en su red.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Se chamado sem subcomandos, mostrar a ajuda
		cmd.Help()
	},
}

// Subcomandos para gerenciamento de peers
var peerAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Adicionar um novo peer",
	Run: func(cmd *cobra.Command, args []string) {
		// Validar parâmetros obrigatórios
		if peerPublicKey == "" || peerVirtualIP == "" {
			fmt.Println("Erro: chave pública e IP virtual são obrigatórios.")
			cmd.Help()
			return
		}

		// Carregar configuração
		absConfigPath, err := filepath.Abs(configPath)
		if err != nil {
			fmt.Printf("Erro ao obter caminho absoluto para configuração: %v\n", err)
			return
		}

		config, err := core.LoadConfig(absConfigPath)
		if err != nil {
			fmt.Printf("Erro ao carregar configuração: %v\n", err)
			return
		}

		// Gerar um ID para o peer se não fornecido
		if peerNodeID == "" {
			peerNodeID = fmt.Sprintf("peer-%s", peerVirtualIP)
		}

		// Criar o peer
		peer := core.TrustedPeer{
			NodeID:    peerNodeID,
			PublicKey: peerPublicKey,
			VirtualIP: peerVirtualIP,
		}

		// Adicionar endpoint se fornecido
		if peerEndpoint != "" {
			peer.Endpoints = []string{peerEndpoint}
		}

		// Configurar keepalive se fornecido
		if peerKeepAlive > 0 {
			peer.KeepAlive = peerKeepAlive
		}

		// Adicionar peer à configuração
		config.AddTrustedPeer(peer)

		// Salvar configuração
		if err := config.SaveConfig(absConfigPath); err != nil {
			fmt.Printf("Erro ao salvar configuração: %v\n", err)
			return
		}

		fmt.Printf("Peer %s adicionado com sucesso!\n", peerNodeID)
		fmt.Println("Reinicie o serviço de VPN para aplicar as alterações.")
	},
}

var peerRemoveCmd = &cobra.Command{
	Use:   "remove",
	Short: "Remover um peer existente",
	Run: func(cmd *cobra.Command, args []string) {
		if peerNodeID == "" {
			fmt.Println("Erro: ID do peer é obrigatório.")
			cmd.Help()
			return
		}

		// Carregar configuração
		absConfigPath, err := filepath.Abs(configPath)
		if err != nil {
			fmt.Printf("Erro ao obter caminho absoluto para configuração: %v\n", err)
			return
		}

		config, err := core.LoadConfig(absConfigPath)
		if err != nil {
			fmt.Printf("Erro ao carregar configuração: %v\n", err)
			return
		}

		// Remover peer
		if !config.RemoveTrustedPeer(peerNodeID) {
			fmt.Printf("Erro: peer %s não encontrado na configuração.\n", peerNodeID)
			return
		}

		// Salvar configuração
		if err := config.SaveConfig(absConfigPath); err != nil {
			fmt.Printf("Erro ao salvar configuração: %v\n", err)
			return
		}

		fmt.Printf("Peer %s removido com sucesso!\n", peerNodeID)
		fmt.Println("Reinicie o serviço de VPN para aplicar as alterações.")
	},
}

var peerListCmd = &cobra.Command{
	Use:   "list",
	Short: "Listar peers configurados",
	Run: func(cmd *cobra.Command, args []string) {
		// Carregar configuração
		absConfigPath, err := filepath.Abs(configPath)
		if err != nil {
			fmt.Printf("Erro ao obter caminho absoluto para configuração: %v\n", err)
			return
		}

		config, err := core.LoadConfig(absConfigPath)
		if err != nil {
			fmt.Printf("Erro ao carregar configuração: %v\n", err)
			return
		}

		// Mostrar peers
		if len(config.TrustedPeers) == 0 {
			fmt.Println("Nenhum peer configurado.")
			return
		}

		fmt.Println("Peers configurados:")
		fmt.Println("--------------------------------------------------")
		for i, peer := range config.TrustedPeers {
			fmt.Printf("%d. ID: %s\n", i+1, peer.NodeID)
			fmt.Printf("   IP virtual: %s\n", peer.VirtualIP)
			fmt.Printf("   Chave pública: %s\n", peer.PublicKey)
			
			if len(peer.Endpoints) > 0 {
				fmt.Printf("   Endpoints: %s\n", peer.Endpoints)
			}
			
			if peer.KeepAlive > 0 {
				fmt.Printf("   KeepAlive: %d segundos\n", peer.KeepAlive)
			}
			
			fmt.Println("--------------------------------------------------")
		}
	},
}

func init() {
	// Adicionar subcomandos ao comando peer
	peerCmd.AddCommand(peerAddCmd)
	peerCmd.AddCommand(peerRemoveCmd)
	peerCmd.AddCommand(peerListCmd)

	// Flags para o comando add
	peerAddCmd.Flags().StringVar(&peerNodeID, "id", "", "ID do peer (opcional)")
	peerAddCmd.Flags().StringVar(&peerPublicKey, "pubkey", "", "Chave pública do peer (obrigatório)")
	peerAddCmd.Flags().StringVar(&peerVirtualIP, "ip", "", "IP virtual do peer (obrigatório)")
	peerAddCmd.Flags().StringVar(&peerEndpoint, "endpoint", "", "Endpoint do peer (ex: 123.45.67.89:51820)")
	peerAddCmd.Flags().IntVar(&peerKeepAlive, "keepalive", 0, "Intervalo de keepalive em segundos")

	// Flags para o comando remove
	peerRemoveCmd.Flags().StringVar(&peerNodeID, "id", "", "ID do peer a ser removido (obrigatório)")
}

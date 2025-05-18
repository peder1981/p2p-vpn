package cli

import (
	"fmt"
	"path/filepath"

	"github.com/p2p-vpn/p2p-vpn/core"
	"github.com/spf13/cobra"
)

var (
	targetEndpoint string
)

// connectCmd representa o comando para se conectar a um peer remoto
// connectCmd represents the command to connect to a remote peer
// connectCmd representa el comando para conectarse a un par remoto
var connectCmd = &cobra.Command{
	Use:   "connect",
	Short: "Conectar a um peer remoto",
	Long: `Conecta-se a um peer remoto da VPN P2P Universal
através do seu endereço ou identificador.

Connects to a remote peer of the Universal P2P VPN
through its address or identifier.

Conecta a un par remoto de la VPN P2P Universal
a través de su dirección o identificador.`,
	Run: func(cmd *cobra.Command, args []string) {
		if peerNodeID == "" || targetEndpoint == "" {
			fmt.Println("Erro: ID do peer e endpoint de destino são obrigatórios.")
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

		// Verificar se o peer existe
		var targetPeer *core.TrustedPeer
		for i, peer := range config.TrustedPeers {
			if peer.NodeID == peerNodeID {
				targetPeer = &config.TrustedPeers[i]
				break
			}
		}

		if targetPeer == nil {
			fmt.Printf("Erro: peer %s não encontrado na configuração.\n", peerNodeID)
			return
		}

		// Atualizar ou adicionar o endpoint
		endpointExists := false
		for _, ep := range targetPeer.Endpoints {
			if ep == targetEndpoint {
				endpointExists = true
				break
			}
		}

		if !endpointExists {
			// Adicionar o novo endpoint
			targetPeer.Endpoints = append(targetPeer.Endpoints, targetEndpoint)
			fmt.Printf("Adicionado novo endpoint %s para o peer %s.\n", targetEndpoint, peerNodeID)
		} else {
			fmt.Printf("Endpoint %s já está configurado para o peer %s.\n", targetEndpoint, peerNodeID)
		}

		// Salvar configuração
		if err := config.SaveConfig(absConfigPath); err != nil {
			fmt.Printf("Erro ao salvar configuração: %v\n", err)
			return
		}

		// Tentar estabelecer conexão
		fmt.Printf("Conexão com %s em %s configurada.\n", peerNodeID, targetEndpoint)
		fmt.Println("Para atualizar a conexão, reinicie o serviço de VPN ou aguarde a próxima atualização automática.")
		
		// Instrução para conexão manual
		fmt.Println("\nPara conectar imediatamente, você também pode usar o seguinte comando:")
		fmt.Printf("wg set wg0 peer %s endpoint %s\n", targetPeer.PublicKey, targetEndpoint)
	},
}

func init() {
	// Flags para o comando connect
	connectCmd.Flags().StringVar(&peerNodeID, "id", "", "ID do peer a se conectar (obrigatório)")
	connectCmd.Flags().StringVar(&targetEndpoint, "endpoint", "", "Endpoint do peer (ex: 123.45.67.89:51820) (obrigatório)")
}

package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	configPath string
	verbose    bool
)

// rootCmd representa o comando base para a CLI
// rootCmd represents the base command for the CLI
// rootCmd representa el comando base para la CLI
var rootCmd = &cobra.Command{
	Use:   "p2p-vpn",
	Short: "VPN P2P Universal - Cliente de linha de comando",
	Long: `Cliente de linha de comando para VPN P2P Universal.
Esta ferramenta permite gerenciar facilmente conexões seguras 
peer-to-peer sem depender de servidores centralizados.

Universal P2P VPN - Command line client.
This tool allows you to easily manage secure 
peer-to-peer connections without relying on centralized servers.

VPN P2P Universal - Cliente de línea de comandos.
Esta herramienta le permite administrar fácilmente conexiones 
seguras peer-to-peer sin depender de servidores centralizados.`,
	// Não executa nada no comando raiz
	Run: func(cmd *cobra.Command, args []string) {
		// Se não houver subcomandos, mostra a ajuda
		cmd.Help()
	},
}

// Execute adiciona todos os comandos filhos ao comando raiz e configura flags
// Execute adds all child commands to the root command and sets flags
// Execute agrega todos los comandos secundarios al comando raíz y configura las banderas
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	// Flags globais persistentes
	// Estas flags serão globais para a aplicação
	rootCmd.PersistentFlags().StringVar(&configPath, "config", "config.yaml", "Caminho para o arquivo de configuração")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Ativar saída detalhada")

	// Adicionar comandos
	rootCmd.AddCommand(startCmd)
	rootCmd.AddCommand(stopCmd)
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(peerCmd)
	rootCmd.AddCommand(connectCmd)
}

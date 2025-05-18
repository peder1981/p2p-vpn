package cli

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

// statusCmd representa o comando para verificar o status do serviço VPN
// statusCmd represents the command to check the VPN service status
// statusCmd representa el comando para verificar el estado del servicio VPN
var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Verificar o status da VPN",
	Long: `Verifica o status atual do serviço de VPN P2P Universal,
mostrando informações sobre conexões ativas, interface de rede
e peers conectados.

Checks the current status of the Universal P2P VPN service,
showing information about active connections, network interface
and connected peers.

Verifica el estado actual del servicio VPN P2P Universal,
mostrando información sobre conexiones activas, interfaz de red
y pares conectados.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Verificar se o processo está em execução
		processCmd := exec.Command("pgrep", "-f", "p2p-vpn.+start")
		output, err := processCmd.Output()
		
		// Verificar status do processo
		if err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
				fmt.Println("Status: INATIVO")
				fmt.Println("O serviço de VPN P2P não está em execução.")
				return
			}
			fmt.Printf("Erro ao verificar status do processo: %v\n", err)
			return
		}
		
		// Processo está em execução
		fmt.Println("Status: ATIVO")
		pidLines := strings.Split(strings.TrimSpace(string(output)), "\n")
		fmt.Printf("Processos em execução: %d\n", len(pidLines))
		
		// Verificar interfaces de rede WireGuard
		wgCmd := exec.Command("ip", "link", "show", "type", "wireguard")
		wgOutput, err := wgCmd.Output()
		
		if err != nil {
			fmt.Println("Interfaces WireGuard: Não foi possível verificar")
		} else {
			wgInterfaces := strings.TrimSpace(string(wgOutput))
			if wgInterfaces == "" {
				fmt.Println("Interfaces WireGuard: Nenhuma encontrada")
			} else {
				fmt.Println("Interfaces WireGuard encontradas:")
				
				// Processar a saída para mostrar as interfaces
				lines := strings.Split(wgInterfaces, "\n")
				for _, line := range lines {
					if strings.Contains(line, ": ") {
						parts := strings.SplitN(line, ": ", 2)
						if len(parts) >= 2 {
							interfaceName := strings.TrimSpace(parts[1])
							interfaceName = strings.Split(interfaceName, "@")[0]
							fmt.Printf("  - %s\n", interfaceName)
							
							// Obter informações da interface
							wgShowCmd := exec.Command("wg", "show", interfaceName)
							wgShowOutput, err := wgShowCmd.Output()
							if err == nil {
								fmt.Println(strings.TrimSpace(string(wgShowOutput)))
							}
						}
					}
				}
			}
		}
		
		// Verificar conexões estabelecidas (simplificado)
		fmt.Println("\nVPN pronta para conectar peers.")
	},
}

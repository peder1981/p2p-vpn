package cli

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

// stopCmd representa o comando para parar o serviço de VPN
// stopCmd represents the command to stop the VPN service
// stopCmd representa el comando para detener el servicio VPN
var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Parar o serviço de VPN",
	Long: `Para o serviço de VPN P2P Universal, desativando a
interface de rede virtual e encerrando conexões com peers.

Stops the Universal P2P VPN service, deactivating the
virtual network interface and terminating connections with peers.

Detiene el servicio VPN P2P Universal, desactivando la
interfaz de red virtual y terminando las conexiones con pares.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Em uma implementação completa, teríamos uma maneira adequada de
		// comunicar com o processo em execução via sockets ou arquivo PID.
		// Esta implementação simplificada verifica processos pelo nome

		fmt.Println("Procurando processos da VPN P2P...")
		
		// Encontrar PIDs dos processos
		processCmd := exec.Command("pgrep", "-f", "p2p-vpn.+start")
		output, err := processCmd.Output()
		if err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
				fmt.Println("Nenhum processo de VPN P2P encontrado em execução.")
				return
			}
			fmt.Printf("Erro ao procurar processos: %v\n", err)
			return
		}
		
		// Processar a saída para obter os PIDs
		pidLines := strings.Split(strings.TrimSpace(string(output)), "\n")
		
		if len(pidLines) == 0 || (len(pidLines) == 1 && pidLines[0] == "") {
			fmt.Println("Nenhum processo de VPN P2P encontrado em execução.")
			return
		}
		
		// Encerrar cada processo encontrado
		for _, pidStr := range pidLines {
			if pidStr == "" {
				continue
			}
			
			pid, err := strconv.Atoi(pidStr)
			if err != nil {
				fmt.Printf("Erro ao converter PID '%s': %v\n", pidStr, err)
				continue
			}
			
			fmt.Printf("Encerrando processo VPN P2P com PID %d...\n", pid)
			
			// Enviar sinal SIGTERM
			process, err := os.FindProcess(pid)
			if err != nil {
				fmt.Printf("Erro ao encontrar processo com PID %d: %v\n", pid, err)
				continue
			}
			
			if err := process.Signal(os.Interrupt); err != nil {
				fmt.Printf("Erro ao enviar sinal para processo %d: %v\n", pid, err)
				
				// Tentar com SIGKILL se SIGTERM falhar
				if err := process.Kill(); err != nil {
					fmt.Printf("Erro ao encerrar forçadamente o processo %d: %v\n", pid, err)
				} else {
					fmt.Printf("Processo %d encerrado forçadamente com sucesso.\n", pid)
				}
			} else {
				fmt.Printf("Sinal enviado com sucesso para processo %d.\n", pid)
			}
		}
		
		fmt.Println("Todos os processos VPN P2P foram encerrados.")
	},
}

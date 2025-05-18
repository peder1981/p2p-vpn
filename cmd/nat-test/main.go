package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	nattraversal "github.com/p2p-vpn/p2p-vpn/nat-traversal"
)

func main() {
	// Comandos disponíveis
	diagnoseCmd := flag.NewFlagSet("diagnose", flag.ExitOnError)
	serverCmd := flag.NewFlagSet("server", flag.ExitOnError)
	clientCmd := flag.NewFlagSet("client", flag.ExitOnError)
	punchtestCmd := flag.NewFlagSet("punchtest", flag.ExitOnError)
	simulateCmd := flag.NewFlagSet("simulate", flag.ExitOnError)

	// Flags para o comando "diagnose"
	diagnoseStunServer := diagnoseCmd.String("stun", "stun.l.google.com:19302", "Servidor STUN para diagnóstico")
	
	// Flags para o comando "server"
	serverPort := serverCmd.Int("port", 8888, "Porta para escutar conexões")
	
	// Flags para o comando "client"
	clientTarget := clientCmd.String("target", "", "Endereço do servidor (host:porta)")
	clientPort := clientCmd.Int("port", 0, "Porta local (0 para automática)")
	
	// Flags para o comando "punchtest"
	punchRendezvous := punchtestCmd.String("rendezvous", "", "Servidor de rendezvous (host:porta)")
	punchID := punchtestCmd.String("id", "", "ID único para identificação no teste")
	punchPort := punchtestCmd.Int("port", 9999, "Porta local para o teste")
	
	// Flags para o comando "simulate"
	simulateType := simulateCmd.Int("type", 0, "Tipo de NAT a simular (0=Full Cone, 1=Restricted Cone, 2=Port Restricted Cone, 3=Symmetric)")
	simulateExternalIP := simulateCmd.String("external", "203.0.113.1", "IP externo simulado")
	simulateInternalNet := simulateCmd.String("internal", "192.168.0.0/24", "Rede interna simulada (CIDR)")
	simulateExternalPort := simulateCmd.Int("extport", 11000, "Porta externa para o simulador")
	simulateInternalPort := simulateCmd.Int("intport", 11001, "Porta interna para o simulador")

	// Verificar se foi passado um subcomando
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	// Analisar o subcomando
	switch os.Args[1] {
	case "diagnose":
		diagnoseCmd.Parse(os.Args[2:])
		runDiagnose(*diagnoseStunServer)
	case "server":
		serverCmd.Parse(os.Args[2:])
		runServer(*serverPort)
	case "client":
		clientCmd.Parse(os.Args[2:])
		if *clientTarget == "" {
			fmt.Println("Erro: O parâmetro --target é obrigatório")
			clientCmd.Usage()
			os.Exit(1)
		}
		runClient(*clientTarget, *clientPort)
	case "punchtest":
		punchtestCmd.Parse(os.Args[2:])
		if *punchRendezvous == "" || *punchID == "" {
			fmt.Println("Erro: Os parâmetros --rendezvous e --id são obrigatórios")
			punchtestCmd.Usage()
			os.Exit(1)
		}
		runPunchTest(*punchRendezvous, *punchID, *punchPort)
	case "simulate":
		simulateCmd.Parse(os.Args[2:])
		simType := nattraversal.NATSimulatorType(*simulateType)
		runSimulator(simType, *simulateExternalIP, *simulateInternalNet, 
			*simulateExternalPort, *simulateInternalPort)
	default:
		printUsage()
		os.Exit(1)
	}
}

// printUsage mostra as instruções de uso da ferramenta
func printUsage() {
	fmt.Println("Ferramenta de Teste para NAT Traversal")
	fmt.Println("Uso:")
	fmt.Println("  nat-test [comando] [opções]")
	fmt.Println("\nComandos disponíveis:")
	fmt.Println("  diagnose    Executa diagnóstico do tipo de NAT")
	fmt.Println("  server      Inicia um servidor de teste")
	fmt.Println("  client      Conecta a um servidor de teste")
	fmt.Println("  punchtest   Executa um teste de UDP hole punching")
	fmt.Println("  simulate    Simula diferentes tipos de NAT")
	fmt.Println("\nExecute 'nat-test [comando] --help' para ver as opções específicas de cada comando")
}

// runDiagnose executa o diagnóstico de NAT
func runDiagnose(stunServerAddr string) {
	fmt.Println("Iniciando diagnóstico de NAT...")
	fmt.Printf("Usando servidor STUN: %s\n", stunServerAddr)
	
	// Analisar o endereço do servidor STUN
	host, port, err := nattraversal.ParseHostPort(stunServerAddr)
	if err != nil {
		fmt.Printf("Erro ao analisar endereço do servidor STUN: %v\n", err)
		os.Exit(1)
	}
	
	// Criar servidor STUN para o teste
	stunServer := nattraversal.STUNServer{
		Address: host,
		Port:    port,
	}
	
	// Executar diagnóstico
	diagnostic := nattraversal.NewNATDiagnostic([]nattraversal.STUNServer{stunServer})
	result, err := diagnostic.RunDiagnosis()
	if err != nil {
		fmt.Printf("Erro durante o diagnóstico: %v\n", err)
		os.Exit(1)
	}
	
	// Mostrar resultados
	fmt.Println("\n=== Resultado do Diagnóstico de NAT ===")
	fmt.Printf("IP Local: %s:%d\n", result.LocalIP, result.LocalPort)
	fmt.Printf("IP Público: %s:%d\n", result.PublicIP, result.PublicPort)
	fmt.Printf("Tipo de NAT: %s\n", result.NATType)
	fmt.Printf("Servidor STUN: %s\n", result.STUNServer)
	fmt.Printf("Data/Hora: %s\n", result.TestTime.Format(time.RFC1123))
	
	// Recomendações com base no tipo de NAT
	fmt.Println("\n=== Recomendações para NAT Traversal ===")
	switch result.NATType {
	case nattraversal.NATOpen:
		fmt.Println("Você está em uma rede aberta sem NAT. Conexões diretas devem funcionar sem problemas.")
	case nattraversal.NATFullCone:
		fmt.Println("NAT Full Cone detectado. Técnicas básicas de hole punching devem funcionar bem.")
	case nattraversal.NATRestrictedCone:
		fmt.Println("NAT Restricted Cone detectado. Hole punching deve funcionar, mas exige troca prévia de IPs.")
	case nattraversal.NATPortRestricted:
		fmt.Println("NAT Port Restricted detectado. Hole punching é possível, mas requer sincronização precisa.")
	case nattraversal.NATSymmetric:
		fmt.Println("NAT Simétrico detectado. Hole punching direto pode não funcionar. Considere usar um relay.")
	default:
		fmt.Println("Tipo de NAT desconhecido. Recomenda-se testes adicionais.")
	}
}

// runServer inicia um servidor de teste para NAT traversal
func runServer(port int) {
	fmt.Printf("Iniciando servidor de teste na porta %d...\n", port)
	
	// Criar o servidor de teste
	server, err := nattraversal.NewTestServer(port)
	if err != nil {
		fmt.Printf("Erro ao iniciar servidor: %v\n", err)
		os.Exit(1)
	}
	
	// Configurar manipulador de sinais para encerramento gracioso
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	
	// Iniciar o servidor em uma goroutine
	go func() {
		if err := server.Start(); err != nil {
			fmt.Printf("Erro ao executar servidor: %v\n", err)
			os.Exit(1)
		}
	}()
	
	fmt.Printf("Servidor iniciado com sucesso. Escutando na porta %d\n", port)
	fmt.Println("Pressione Ctrl+C para encerrar.")
	
	// Aguardar sinal de encerramento
	<-sigCh
	fmt.Println("\nEncerrando servidor...")
	
	// Parar o servidor
	if err := server.Stop(); err != nil {
		fmt.Printf("Erro ao encerrar servidor: %v\n", err)
	}
	
	fmt.Println("Servidor encerrado com sucesso.")
}

// runClient executa um cliente de teste para conectar ao servidor
func runClient(target string, localPort int) {
	fmt.Printf("Iniciando cliente de teste...\n")
	fmt.Printf("Destino: %s\n", target)
	
	// Criar o cliente de teste
	client, err := nattraversal.NewTestClient(localPort)
	if err != nil {
		fmt.Printf("Erro ao criar cliente: %v\n", err)
		os.Exit(1)
	}
	
	// Conectar ao servidor
	fmt.Printf("Conectando ao servidor %s...\n", target)
	if err := client.Connect(target); err != nil {
		fmt.Printf("Erro ao conectar ao servidor: %v\n", err)
		os.Exit(1)
	}
	
	fmt.Println("Conexão estabelecida com sucesso!")
	fmt.Printf("Endereço local: %s\n", client.LocalAddr())
	fmt.Printf("Endereço remoto: %s\n", client.RemoteAddr())
	
	// Iniciar loop de envio/recebimento de mensagens
	go client.ReceiveLoop()
	
	fmt.Println("\nDigite mensagens para enviar ou Ctrl+C para sair.")
	
	// Configurar manipulador de sinais para encerramento gracioso
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	
	// Loop principal
	go func() {
		for {
			// Ler mensagem do usuário
			var message string
			fmt.Print("> ")
			fmt.Scanln(&message)
			
			if message == "" {
				continue
			}
			
			// Enviar mensagem
			if err := client.SendMessage(message); err != nil {
				fmt.Printf("Erro ao enviar mensagem: %v\n", err)
			}
		}
	}()
	
	// Aguardar sinal de encerramento
	<-sigCh
	fmt.Println("\nEncerrando cliente...")
	
	// Encerrar o cliente
	client.Close()
	
	fmt.Println("Cliente encerrado com sucesso.")
}

// runPunchTest executa um teste de UDP hole punching
func runPunchTest(rendezvousServer, id string, localPort int) {
	fmt.Printf("Iniciando teste de UDP hole punching...\n")
	fmt.Printf("Servidor de rendezvous: %s\n", rendezvousServer)
	fmt.Printf("ID: %s\n", id)
	
	// Criar cliente de teste para simular hole punching
	client, err := nattraversal.NewTestClient(localPort)
	if err != nil {
		fmt.Printf("Erro ao iniciar cliente para teste de hole punching: %v\n", err)
		os.Exit(1)
	}
	
	// Tentar conectar ao servidor de rendezvous
	fmt.Printf("Conectando ao servidor de rendezvous %s...\n", rendezvousServer)
	if err := client.Connect(rendezvousServer); err != nil {
		fmt.Printf("Erro ao conectar ao servidor de rendezvous: %v\n", err)
		os.Exit(1)
	}
	
	// Iniciar loop de recebimento de mensagens
	go client.ReceiveLoop()
	
	// Registrar-se no servidor com ID fornecido
	fmt.Printf("Registrando-se no servidor como '%s'...\n", id)
	if err := client.SendMessage("/register " + id); err != nil {
		fmt.Printf("Erro ao registrar-se: %v\n", err)
		os.Exit(1)
	}
	
	// Listar outros clientes disponíveis
	time.Sleep(500 * time.Millisecond) // Pequena pausa para dar tempo ao registro
	if err := client.SendMessage("/list"); err != nil {
		fmt.Printf("Erro ao solicitar lista de clientes: %v\n", err)
	}
	
	// Configurar manipulador de sinais para encerramento gracioso
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	
	fmt.Println("\nTeste iniciado. Comandos disponíveis:")
	fmt.Println("  /list                  - Listar clientes conectados")
	fmt.Println("  /connect <nome>        - Conectar a outro cliente")
	fmt.Println("  Qualquer outra mensagem será enviada como chat")
	fmt.Println("Pressione Ctrl+C para encerrar.")
	
	// Loop de leitura das mensagens do usuário
	go func() {
		for {
			fmt.Print("> ")
			var message string
			fmt.Scanln(&message)
			
			if message == "" {
				continue
			}
			
			if err := client.SendMessage(message); err != nil {
				fmt.Printf("Erro ao enviar mensagem: %v\n", err)
			}
		}
	}()
	
	// Aguardar sinal de encerramento
	<-sigCh
	fmt.Println("\nEncerrando teste...")
	
	// Encerrar o cliente
	client.Close()
	
	fmt.Println("\n=== Teste de Hole Punching Encerrado ===")
	fmt.Printf("Endereço local: %s\n", client.LocalAddr())
	if client.RemoteAddr() != "não conectado" {
		fmt.Printf("Último endereço conectado: %s\n", client.RemoteAddr())
	} else {
		fmt.Println("Nenhuma conexão peer-to-peer estabelecida")
	}
}

// runSimulator executa um simulador de NAT
func runSimulator(natType nattraversal.NATSimulatorType, externalIP, internalNet string, 
	externalPort, internalPort int) {
	
	// Criar o simulador de NAT
	natTypeStr := ""
	switch natType {
	case nattraversal.SimulateFullCone:
		natTypeStr = "Full Cone"
	case nattraversal.SimulateRestrictedCone:
		natTypeStr = "Restricted Cone"
	case nattraversal.SimulatePortRestrictedCone:
		natTypeStr = "Port Restricted Cone"
	case nattraversal.SimulateSymmetric:
		natTypeStr = "Symmetric"
	default:
		natTypeStr = "Desconhecido"
	}
	
	fmt.Printf("Iniciando simulador de NAT tipo %s (%d)\n", natTypeStr, natType)
	fmt.Printf("IP Externo: %s\n", externalIP)
	fmt.Printf("Rede Interna: %s\n", internalNet)
	fmt.Printf("Porta Externa: %d, Porta Interna: %d\n", externalPort, internalPort)
	
	simulator, err := nattraversal.NewNATSimulator(natType, externalIP, internalNet)
	if err != nil {
		fmt.Printf("Erro ao criar simulador: %v\n", err)
		os.Exit(1)
	}
	
	// Iniciar o simulador
	err = simulator.Start(externalPort, internalPort)
	if err != nil {
		fmt.Printf("Erro ao iniciar simulador: %v\n", err)
		os.Exit(1)
	}
	
	// Criar canal para comandos
	cmdChan := make(chan string, 1)
	
	// Iniciar goroutine para ler comandos do usuário
	go func() {
		reader := bufio.NewReader(os.Stdin)
		for {
			fmt.Print("> ")
			cmdLine, err := reader.ReadString('\n')
			if err != nil {
				fmt.Printf("Erro ao ler comando: %v\n", err)
				continue
			}
			
			cmdLine = strings.TrimSpace(cmdLine)
			cmdChan <- cmdLine
		}
	}()
	
	// Configurar manipulador de sinais para encerramento gracioso
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	
	fmt.Println("\nSimulador iniciado. Comandos disponíveis:")
	fmt.Println("  list     - Lista os mapeamentos de NAT ativos")
	fmt.Println("  help     - Mostra esta ajuda")
	fmt.Println("  quit     - Encerra o simulador")
	fmt.Println("Pressione Ctrl+C para encerrar.\n")
	
	// Loop principal
	running := true
	for running {
		select {
		case <-sigChan:
			fmt.Println("\nRecebido sinal de interrupção, encerrando...")
			running = false
			
		case cmd := <-cmdChan:
			switch cmd {
			case "quit", "exit":
				fmt.Println("Encerrando simulador...")
				running = false
				
			case "list":
				// Listar mapeamentos ativos
				mappings := simulator.ListMappings()
				fmt.Printf("\n=== Mapeamentos Ativos (%d) ===\n", len(mappings))
				for i, mapping := range mappings {
					fmt.Printf("%d. Interno: %s -> Externo: %s\n", 
						i+1, mapping["internal_addr"], mapping["external_addr"])
					fmt.Printf("   Destinos (%d): ", len(mapping["destinations"].([]string)))
					for j, dest := range mapping["destinations"].([]string) {
						if j > 0 {
							fmt.Print(", ")
						}
						fmt.Print(dest)
						if j >= 4 {
							fmt.Printf(" e mais %d...", len(mapping["destinations"].([]string))-5)
							break
						}
					}
					fmt.Printf("\n   Última atividade: %s (%s atrás)\n\n", 
						mapping["last_activity"], mapping["idle_time"])
				}
				
			case "help":
				fmt.Println("\nComandos disponíveis:")
				fmt.Println("  list     - Lista os mapeamentos de NAT ativos")
				fmt.Println("  help     - Mostra esta ajuda")
				fmt.Println("  quit     - Encerra o simulador\n")
				
			case "":
				// Ignorar linha vazia
				continue
				
			default:
				// Tentar analisar como um comando com parâmetros
				parts := strings.Split(cmd, " ")
				if len(parts) > 0 {
					switch parts[0] {
					case "add":
						if len(parts) >= 3 {
							// Exemplo de comando para adicionar mapeamento manualmente
							// add 192.168.0.2:1234 5678
							internalAddr := parts[1]
							externalPort, err := strconv.Atoi(parts[2])
							if err != nil {
								fmt.Printf("Porta externa inválida: %v\n", err)
								continue
							}
							
							fmt.Printf("Comando ainda não implementado: add %s %d\n", 
								internalAddr, externalPort)
						} else {
							fmt.Println("Uso: add <endereço_interno> <porta_externa>")
						}
						
					default:
						fmt.Printf("Comando desconhecido: %s\n", cmd)
						fmt.Println("Digite 'help' para ver os comandos disponíveis.")
					}
				}
			}
		}
	}
	
	// Parar o simulador
	simulator.Stop()
	fmt.Println("Simulador encerrado com sucesso.")
}

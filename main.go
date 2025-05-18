package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/p2p-vpn/p2p-vpn/core"
	"github.com/p2p-vpn/p2p-vpn/discovery"
	"github.com/p2p-vpn/p2p-vpn/platform"
	"github.com/p2p-vpn/p2p-vpn/security"
	"github.com/p2p-vpn/p2p-vpn/ui/web"
)

func main() {
	// Configurações via linha de comando
	listenPort := flag.Int("port", 51820, "Porta local para o serviço WireGuard")
	discoveryPort := flag.Int("discovery-port", 51821, "Porta para o serviço de descoberta de peers")
	configPath := flag.String("config", "config.yaml", "Caminho para o arquivo de configuração")
	securityConfigPath := flag.String("security-config", "config/server_security.yaml", "Caminho para o arquivo de configuração de segurança")
	webPort := flag.String("web-port", "8080", "Porta para a interface web")
	flag.Parse()

	// Inicializar o logger
	fmt.Println("Inicializando VPN P2P...")
	
	// Carregar configuração
	config, err := core.LoadConfig(*configPath)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("Arquivo de configuração não encontrado. Criando nova configuração...")
			config = core.GenerateDefaultConfig(*configPath)
			fmt.Println("Nova configuração criada com sucesso!")
		} else {
			fmt.Printf("Erro ao carregar configuração: %v\n", err)
			os.Exit(1)
		}
	}

	// Verificar a plataforma atual
	plat, err := platform.GetPlatform()
	if err != nil {
		fmt.Printf("Erro ao detectar plataforma: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Plataforma detectada: %s\n", plat.Name())

	// Inicializar o core da VPN multiplataforma
	var vpnCore core.VPNProvider
	
	// Tentativa de criar a versão multiplataforma primeiro
	vpnMulti, err := core.NewVPNCoreMulti(config, *listenPort)
	if err == nil {
		vpnCore = vpnMulti
		fmt.Println("Usando implementação VPN multiplataforma")
	} else {
		// Se falhar, tentar a versão específica para Linux
		fmt.Printf("Não foi possível inicializar o core multiplataforma: %v\n", err)
		fmt.Println("Tentando inicializar o core específico para Linux...")
		
		vpnLinux, err := core.NewVPNCore(config, *listenPort)
		if err != nil {
			fmt.Printf("Erro ao inicializar o core da VPN: %v\n", err)
			os.Exit(1)
		}
		vpnCore = vpnLinux
		fmt.Println("Usando implementação VPN específica para Linux")
	}

	// Inicializar o sistema de descoberta de peers
	peerDiscovery, err := discovery.NewPeerDiscovery(config, *discoveryPort, vpnCore)
	if err != nil {
		fmt.Printf("Erro ao inicializar o sistema de descoberta: %v\n", err)
		os.Exit(1)
	}

	// Carregar configuração de segurança
	fmt.Println("Carregando configuração de segurança...")
	securityConfig, err := security.LoadSecurityConfig(*securityConfigPath)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("Arquivo de configuração de segurança não encontrado. Usando configuração padrão.")
			// Criar diretório para o arquivo de configuração
			if err := os.MkdirAll(filepath.Dir(*securityConfigPath), 0755); err != nil {
				fmt.Printf("Erro ao criar diretório de configuração: %v\n", err)
				os.Exit(1)
			}
		} else {
			fmt.Printf("Erro ao carregar configuração de segurança: %v\n", err)
			os.Exit(1)
		}
	}

	// Iniciar os serviços
	if err := vpnCore.Start(); err != nil {
		fmt.Printf("Erro ao iniciar o core da VPN: %v\n", err)
		os.Exit(1)
	}
	
	if err := peerDiscovery.Start(); err != nil {
		fmt.Printf("Erro ao iniciar o sistema de descoberta: %v\n", err)
		vpnCore.Stop()
		os.Exit(1)
	}

	// Iniciar servidor web com HTTPS e autenticação
	webAddr := fmt.Sprintf("0.0.0.0:%s", *webPort)
	if securityConfig != nil {
		webAddr = securityConfig.Web.ListenAddr
	}
	
	webConfig := web.Config{
		ListenAddr:       webAddr,
		CoreVPN:          vpnCore, // VPNProvider já é aceito aqui
		Config:           config,
		UseHTTPS:         securityConfig != nil && securityConfig.Web.HTTPS.Enabled,
		TLSConfig:        securityConfig.ToTLSConfig(),
		JWTSecret:        securityConfig.Web.Auth.JWTSecret,
		JWTExpiration:    time.Duration(securityConfig.Web.Auth.JWTExpiration) * time.Hour,
		DefaultAdminUser: securityConfig.Web.Auth.DefaultAdmin.Username,
		DefaultAdminPass: securityConfig.Web.Auth.DefaultAdmin.Password,
		UsersFile:        securityConfig.Web.Auth.UsersFile,
	}
	
	// Iniciar o servidor web em uma goroutine separada
	go func() {
		fmt.Println("Iniciando servidor web...")
		if err := web.StartServer(webConfig); err != nil {
			fmt.Printf("Erro ao iniciar servidor web: %v\n", err)
		}
	}()

	fmt.Println("VPN P2P iniciada com sucesso!")
	fmt.Printf("Escutando na porta %d (WireGuard), %d (Descoberta) e %s (Web)\n", 
		*listenPort, *discoveryPort, webAddr)
	fmt.Println("Pressione Ctrl+C para encerrar.")

	// Configurar tratamento de sinais para encerramento gracioso
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	// Encerrar os serviços
	fmt.Println("\nEncerrando...")
	peerDiscovery.Stop()
	vpnCore.Stop()
	fmt.Println("VPN P2P encerrada com sucesso!")
}

package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

// Definições de cores usadas nos ícones
const (
	// Cores do escudo
	ShieldFillConnected    = "#27ae60" // Verde
	ShieldFillDisconnected = "#e74c3c" // Vermelho
	ShieldStroke           = "#2c3e50" // Azul escuro
	
	// Cores do globo
	GlobeStroke    = "#2c3e50" // Azul escuro
	GlobeFill      = "#3498db" // Azul claro
	
	// Cor do texto
	TextColor = "#2c3e50" // Azul escuro
)

// Dimensões principais
const (
	DefaultSize = 256 // Tamanho padrão do ícone
)

// Função principal para gerar todos os ícones
func main() {
	// Cria o diretório de ícones se não existir
	iconsDir := "icons"
	if _, err := os.Stat(iconsDir); os.IsNotExist(err) {
		os.MkdirAll(iconsDir, 0755)
	}

	// Gerar os ícones principais
	generateAppIcon(filepath.Join(iconsDir, "app_icon.svg"))
	generateTrayIconConnected(filepath.Join(iconsDir, "tray_connected.svg"))
	generateTrayIconDisconnected(filepath.Join(iconsDir, "tray_disconnected.svg"))
	generateStatusIconConnected(filepath.Join(iconsDir, "status_connected.svg"))
	generateStatusIconDisconnected(filepath.Join(iconsDir, "status_disconnected.svg"))

	fmt.Println("Ícones gerados com sucesso no diretório:", iconsDir)
	fmt.Println("Por favor, converti-los para os formatos adequados (PNG, ICO, ICNS) usando ferramentas apropriadas.")
}

// Gerar ícone principal da aplicação
func generateAppIcon(filePath string) {
	width := DefaultSize
	height := DefaultSize

	// Criar o documento SVG
	svg := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8" standalone="no"?>
<svg width="%d" height="%d" viewBox="0 0 %d %d" xmlns="http://www.w3.org/2000/svg">
  <!-- Fundo circular -->
  <circle cx="%d" cy="%d" r="%d" fill="white" />
  
  <!-- Globo -->
  <circle cx="%d" cy="%d" r="%d" fill="%s" stroke="%s" stroke-width="8" />
  
  <!-- Linhas do globo (meridianos) -->
  <path d="M %d,%d C %d,%d %d,%d %d,%d" 
        fill="none" stroke="%s" stroke-width="4" />
  <path d="M %d,%d C %d,%d %d,%d %d,%d" 
        fill="none" stroke="%s" stroke-width="4" />
  
  <!-- Escudo de segurança -->
  <path d="M %d,%d L %d,%d L %d,%d L %d,%d z" 
        fill="%s" stroke="%s" stroke-width="4" />
  
  <!-- "VPN" no escudo -->
  <text x="%d" y="%d" font-family="Arial" font-weight="bold" font-size="%d" 
        text-anchor="middle" fill="%s">VPN</text>
</svg>`,
		width, height, width, height,
		width/2, height/2, width*0.45, // Fundo circular
		width/2, height/2, width*0.4, GlobeFill, GlobeStroke, // Globo
		width*0.2, height*0.3, width*0.5, height*0.2, width*0.8, height*0.3, width*0.5, height*0.7, GlobeStroke, // Meridiano 1
		width*0.3, height*0.2, width*0.5, height*0.6, width*0.7, height*0.2, width*0.5, height*0.8, GlobeStroke, // Meridiano 2
		width*0.3, height*0.45, width*0.5, height*0.25, width*0.7, height*0.45, width*0.5, height*0.7, ShieldFillConnected, ShieldStroke, // Escudo
		width*0.5, height*0.53, width*0.12, TextColor, // Texto VPN
	)

	// Salvar o arquivo SVG
	err := ioutil.WriteFile(filePath, []byte(svg), 0644)
	if err != nil {
		log.Fatalf("Erro ao salvar o ícone da aplicação: %v", err)
	}
}

// Gerar ícone da bandeja para VPN conectada
func generateTrayIconConnected(filePath string) {
	size := 22 // Tamanho padrão para ícones de bandeja
	
	svg := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8" standalone="no"?>
<svg width="%d" height="%d" viewBox="0 0 %d %d" xmlns="http://www.w3.org/2000/svg">
  <!-- Escudo conectado -->
  <path d="M %d,%d L %d,%d L %d,%d L %d,%d z" 
        fill="%s" stroke="%s" stroke-width="1.5" />
  
  <!-- Símbolo de verificação -->
  <path d="M %d,%d L %d,%d L %d,%d" 
        fill="none" stroke="white" stroke-width="1.5" stroke-linecap="round" />
</svg>`,
		size, size, size, size,
		size*0.2, size*0.3, size*0.5, size*0.1, size*0.8, size*0.3, size*0.5, size*0.9, ShieldFillConnected, ShieldStroke, // Escudo
		size*0.35, size*0.55, size*0.45, size*0.7, size*0.65, size*0.4, // Símbolo de verificação
	)

	// Salvar o arquivo SVG
	err := ioutil.WriteFile(filePath, []byte(svg), 0644)
	if err != nil {
		log.Fatalf("Erro ao salvar o ícone da bandeja conectado: %v", err)
	}
}

// Gerar ícone da bandeja para VPN desconectada
func generateTrayIconDisconnected(filePath string) {
	size := 22 // Tamanho padrão para ícones de bandeja
	
	svg := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8" standalone="no"?>
<svg width="%d" height="%d" viewBox="0 0 %d %d" xmlns="http://www.w3.org/2000/svg">
  <!-- Escudo desconectado -->
  <path d="M %d,%d L %d,%d L %d,%d L %d,%d z" 
        fill="%s" stroke="%s" stroke-width="1.5" />
  
  <!-- Símbolo X -->
  <path d="M %d,%d L %d,%d M %d,%d L %d,%d" 
        fill="none" stroke="white" stroke-width="1.5" stroke-linecap="round" />
</svg>`,
		size, size, size, size,
		size*0.2, size*0.3, size*0.5, size*0.1, size*0.8, size*0.3, size*0.5, size*0.9, ShieldFillDisconnected, ShieldStroke, // Escudo
		size*0.35, size*0.4, size*0.65, size*0.7, size*0.35, size*0.7, size*0.65, size*0.4, // Símbolo X
	)

	// Salvar o arquivo SVG
	err := ioutil.WriteFile(filePath, []byte(svg), 0644)
	if err != nil {
		log.Fatalf("Erro ao salvar o ícone da bandeja desconectado: %v", err)
	}
}

// Gerar ícone de status para VPN conectada
func generateStatusIconConnected(filePath string) {
	size := 32 // Tamanho padrão para ícones de status
	
	svg := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8" standalone="no"?>
<svg width="%d" height="%d" viewBox="0 0 %d %d" xmlns="http://www.w3.org/2000/svg">
  <!-- Círculo de fundo -->
  <circle cx="%d" cy="%d" r="%d" fill="%s" />
  
  <!-- Símbolo de verificação -->
  <path d="M %d,%d L %d,%d L %d,%d" 
        fill="none" stroke="white" stroke-width="3" stroke-linecap="round" />
</svg>`,
		size, size, size, size,
		size/2, size/2, size*0.4, ShieldFillConnected, // Círculo
		size*0.3, size*0.5, size*0.45, size*0.7, size*0.7, size*0.35, // Símbolo de verificação
	)

	// Salvar o arquivo SVG
	err := ioutil.WriteFile(filePath, []byte(svg), 0644)
	if err != nil {
		log.Fatalf("Erro ao salvar o ícone de status conectado: %v", err)
	}
}

// Gerar ícone de status para VPN desconectada
func generateStatusIconDisconnected(filePath string) {
	size := 32 // Tamanho padrão para ícones de status
	
	svg := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8" standalone="no"?>
<svg width="%d" height="%d" viewBox="0 0 %d %d" xmlns="http://www.w3.org/2000/svg">
  <!-- Círculo de fundo -->
  <circle cx="%d" cy="%d" r="%d" fill="%s" />
  
  <!-- Símbolo X -->
  <path d="M %d,%d L %d,%d M %d,%d L %d,%d" 
        fill="none" stroke="white" stroke-width="3" stroke-linecap="round" />
</svg>`,
		size, size, size, size,
		size/2, size/2, size*0.4, ShieldFillDisconnected, // Círculo
		size*0.3, size*0.3, size*0.7, size*0.7, size*0.3, size*0.7, size*0.7, size*0.3, // Símbolo X
	)

	// Salvar o arquivo SVG
	err := ioutil.WriteFile(filePath, []byte(svg), 0644)
	if err != nil {
		log.Fatalf("Erro ao salvar o ícone de status desconectado: %v", err)
	}
}

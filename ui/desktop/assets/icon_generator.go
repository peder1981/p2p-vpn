package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
)

// Definições de cores usadas nos ícones
// Color definitions used in the icons
// Definiciones de colores usados en los iconos
const (
	// Cores do escudo / Shield colors / Colores del escudo
	ShieldFillConnected    = "#27ae60" // Verde / Green / Verde
	ShieldFillDisconnected = "#e74c3c" // Vermelho / Red / Rojo
	ShieldStroke           = "#2c3e50" // Azul escuro / Dark blue / Azul oscuro
	
	// Cores do globo / Globe colors / Colores del globo
	GlobeStroke    = "#2c3e50" // Azul escuro / Dark blue / Azul oscuro
	GlobeFill      = "#3498db" // Azul claro / Light blue / Azul claro
	
	// Cor do texto / Text color / Color del texto
	TextColor = "#2c3e50" // Azul escuro / Dark blue / Azul oscuro
)

// Função principal para gerar todos os ícones
// Main function to generate all icons
// Función principal para generar todos los iconos
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
	fmt.Println("Por favor, converta-os para os formatos adequados (PNG, ICO, ICNS) usando ferramentas apropriadas.")
}

// Gerar ícone principal da aplicação
// Generate main application icon
// Generar icono principal de la aplicación
func generateAppIcon(filePath string) {
	width := 256.0
	height := 256.0

	// Criar o documento SVG com valores float explícitos
	svg := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8" standalone="no"?>
<svg width="%.0f" height="%.0f" viewBox="0 0 %.0f %.0f" xmlns="http://www.w3.org/2000/svg">
  <!-- Fundo circular / Circular background / Fondo circular -->
  <circle cx="%.1f" cy="%.1f" r="%.1f" fill="white" />
  
  <!-- Globo / Globe / Globo -->
  <circle cx="%.1f" cy="%.1f" r="%.1f" fill="%s" stroke="%s" stroke-width="8" />
  
  <!-- Linhas do globo (meridianos) / Globe lines (meridians) / Líneas del globo (meridianos) -->
  <path d="M %.1f,%.1f C %.1f,%.1f %.1f,%.1f %.1f,%.1f" 
        fill="none" stroke="%s" stroke-width="4" />
  <path d="M %.1f,%.1f C %.1f,%.1f %.1f,%.1f %.1f,%.1f" 
        fill="none" stroke="%s" stroke-width="4" />
  
  <!-- Escudo de segurança / Security shield / Escudo de seguridad -->
  <path d="M %.1f,%.1f L %.1f,%.1f L %.1f,%.1f L %.1f,%.1f z" 
        fill="%s" stroke="%s" stroke-width="4" />
  
  <!-- "VPN" no escudo / "VPN" in shield / "VPN" en el escudo -->
  <text x="%.1f" y="%.1f" font-family="Arial" font-weight="bold" font-size="%.0f" 
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
	err := os.WriteFile(filePath, []byte(svg), 0644)
	if err != nil {
		log.Fatalf("Erro ao salvar o ícone da aplicação: %v", err)
	}
}

// Gerar ícone da bandeja para VPN conectada
// Generate tray icon for connected VPN
// Generar icono de bandeja para VPN conectada
func generateTrayIconConnected(filePath string) {
	size := 22.0 // Tamanho padrão para ícones de bandeja
	
	svg := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8" standalone="no"?>
<svg width="%.0f" height="%.0f" viewBox="0 0 %.0f %.0f" xmlns="http://www.w3.org/2000/svg">
  <!-- Escudo conectado / Connected shield / Escudo conectado -->
  <path d="M %.1f,%.1f L %.1f,%.1f L %.1f,%.1f L %.1f,%.1f z" 
        fill="%s" stroke="%s" stroke-width="1.5" />
  
  <!-- Símbolo de verificação / Check symbol / Símbolo de verificación -->
  <path d="M %.1f,%.1f L %.1f,%.1f L %.1f,%.1f" 
        fill="none" stroke="white" stroke-width="1.5" stroke-linecap="round" />
</svg>`,
		size, size, size, size,
		size*0.2, size*0.3, size*0.5, size*0.1, size*0.8, size*0.3, size*0.5, size*0.9, ShieldFillConnected, ShieldStroke, // Escudo
		size*0.35, size*0.55, size*0.45, size*0.7, size*0.65, size*0.4, // Símbolo de verificação
	)

	// Salvar o arquivo SVG
	err := os.WriteFile(filePath, []byte(svg), 0644)
	if err != nil {
		log.Fatalf("Erro ao salvar o ícone da bandeja conectado: %v", err)
	}
}

// Gerar ícone da bandeja para VPN desconectada
// Generate tray icon for disconnected VPN
// Generar icono de bandeja para VPN desconectada
func generateTrayIconDisconnected(filePath string) {
	size := 22.0 // Tamanho padrão para ícones de bandeja
	
	svg := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8" standalone="no"?>
<svg width="%.0f" height="%.0f" viewBox="0 0 %.0f %.0f" xmlns="http://www.w3.org/2000/svg">
  <!-- Escudo desconectado / Disconnected shield / Escudo desconectado -->
  <path d="M %.1f,%.1f L %.1f,%.1f L %.1f,%.1f L %.1f,%.1f z" 
        fill="%s" stroke="%s" stroke-width="1.5" />
  
  <!-- Símbolo X / X symbol / Símbolo X -->
  <path d="M %.1f,%.1f L %.1f,%.1f M %.1f,%.1f L %.1f,%.1f" 
        fill="none" stroke="white" stroke-width="1.5" stroke-linecap="round" />
</svg>`,
		size, size, size, size,
		size*0.2, size*0.3, size*0.5, size*0.1, size*0.8, size*0.3, size*0.5, size*0.9, ShieldFillDisconnected, ShieldStroke, // Escudo
		size*0.35, size*0.4, size*0.65, size*0.7, size*0.35, size*0.7, size*0.65, size*0.4, // Símbolo X
	)

	// Salvar o arquivo SVG
	err := os.WriteFile(filePath, []byte(svg), 0644)
	if err != nil {
		log.Fatalf("Erro ao salvar o ícone da bandeja desconectado: %v", err)
	}
}

// Gerar ícone de status para VPN conectada
// Generate status icon for connected VPN
// Generar icono de estado para VPN conectada
func generateStatusIconConnected(filePath string) {
	size := 32.0 // Tamanho padrão para ícones de status
	
	svg := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8" standalone="no"?>
<svg width="%.0f" height="%.0f" viewBox="0 0 %.0f %.0f" xmlns="http://www.w3.org/2000/svg">
  <!-- Círculo de fundo / Background circle / Círculo de fondo -->
  <circle cx="%.1f" cy="%.1f" r="%.1f" fill="%s" />
  
  <!-- Símbolo de verificação / Check symbol / Símbolo de verificación -->
  <path d="M %.1f,%.1f L %.1f,%.1f L %.1f,%.1f" 
        fill="none" stroke="white" stroke-width="3" stroke-linecap="round" />
</svg>`,
		size, size, size, size,
		size/2, size/2, size*0.4, ShieldFillConnected, // Círculo
		size*0.3, size*0.5, size*0.45, size*0.7, size*0.7, size*0.35, // Símbolo de verificação
	)

	// Salvar o arquivo SVG
	err := os.WriteFile(filePath, []byte(svg), 0644)
	if err != nil {
		log.Fatalf("Erro ao salvar o ícone de status conectado: %v", err)
	}
}

// Gerar ícone de status para VPN desconectada
// Generate status icon for disconnected VPN
// Generar icono de estado para VPN desconectada
func generateStatusIconDisconnected(filePath string) {
	size := 32.0 // Tamanho padrão para ícones de status
	
	svg := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8" standalone="no"?>
<svg width="%.0f" height="%.0f" viewBox="0 0 %.0f %.0f" xmlns="http://www.w3.org/2000/svg">
  <!-- Círculo de fundo / Background circle / Círculo de fondo -->
  <circle cx="%.1f" cy="%.1f" r="%.1f" fill="%s" />
  
  <!-- Símbolo X / X symbol / Símbolo X -->
  <path d="M %.1f,%.1f L %.1f,%.1f M %.1f,%.1f L %.1f,%.1f" 
        fill="none" stroke="white" stroke-width="3" stroke-linecap="round" />
</svg>`,
		size, size, size, size,
		size/2, size/2, size*0.4, ShieldFillDisconnected, // Círculo
		size*0.3, size*0.3, size*0.7, size*0.7, size*0.3, size*0.7, size*0.7, size*0.3, // Símbolo X
	)

	// Salvar o arquivo SVG
	err := os.WriteFile(filePath, []byte(svg), 0644)
	if err != nil {
		log.Fatalf("Erro ao salvar o ícone de status desconectado: %v", err)
	}
}

package platform

import (
	"fmt"
	"os"
	"path/filepath"
)

// O tipo LinuxPlatform está definido no arquivo platform_linux.go
// LinuxPlatform type is defined in platform_linux.go
// El tipo LinuxPlatform está definido en el archivo platform_linux.go

// init registra a plataforma Linux
func init() {
	// Registrar a plataforma Linux no sistema
	RegisterPlatform(func() (VPNPlatform, error) {
		// Verificar se o módulo wireguard está carregado
		_, err := os.Stat("/sys/module/wireguard")
		if err != nil {
			// Tentar verificar se o kernel tem suporte embutido para WireGuard
			_, err = os.Stat("/proc/sys/net/ipv4/conf/all/forwarding")
			if err != nil {
				return nil, fmt.Errorf("módulo wireguard não encontrado e verificação de suporte de kernel falhou: %w", err)
			}
		}
		
		// Verificar se os utilitários necessários estão disponíveis
		utilsDir := "/usr/bin"
		requiredUtils := []string{"ip", "wg"}
		
		for _, util := range requiredUtils {
			_, err := os.Stat(filepath.Join(utilsDir, util))
			if err != nil {
				return nil, fmt.Errorf("utilitário %s não encontrado: %w", util, err)
			}
		}
		
		// Linux suportado, retornar nova instância
		return &LinuxPlatform{}, nil
	})
}

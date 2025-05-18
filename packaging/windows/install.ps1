# Script de instalação para Windows do P2P-VPN
# Windows installation script for P2P-VPN
# Script de instalación para Windows de P2P-VPN

# Requer privilégios de administrador
if (-NOT ([Security.Principal.WindowsPrincipal][Security.Principal.WindowsIdentity]::GetCurrent()).IsInRole([Security.Principal.WindowsBuiltInRole] "Administrator")) {
    Write-Warning "Este script requer privilégios de administrador!"
    Write-Warning "Por favor, execute o PowerShell como administrador e tente novamente."
    exit 1
}

# Configuração de diretórios
$InstallDir = "C:\Program Files\P2P-VPN"
$ConfigDir = "C:\ProgramData\P2P-VPN"
$LogDir = "$ConfigDir\logs"
$BinDir = "$InstallDir\bin"
$ServiceName = "P2P-VPN"

Write-Host "Iniciando instalação do P2P-VPN para Windows..." -ForegroundColor Green

# Verificar dependências
Write-Host "Verificando dependências..." -ForegroundColor Yellow

# Verificar instalação do WireGuard
if (-Not (Test-Path "C:\Program Files\WireGuard\wireguard.exe")) {
    Write-Host "WireGuard não encontrado! Iniciando download..." -ForegroundColor Red
    
    # URL de download do instalador do WireGuard
    $WireGuardURL = "https://download.wireguard.com/windows-client/wireguard-installer.exe"
    $WireGuardInstaller = "$env:TEMP\wireguard-installer.exe"
    
    # Baixar o instalador
    try {
        Invoke-WebRequest -Uri $WireGuardURL -OutFile $WireGuardInstaller
        
        # Executar o instalador
        Write-Host "Executando instalador do WireGuard..." -ForegroundColor Yellow
        Start-Process -FilePath $WireGuardInstaller -Args "/S" -Wait
        
        # Verificar novamente se o WireGuard foi instalado
        if (-Not (Test-Path "C:\Program Files\WireGuard\wireguard.exe")) {
            Write-Host "Falha ao instalar o WireGuard. Por favor, instale manualmente e tente novamente." -ForegroundColor Red
            exit 1
        }
    }
    catch {
        Write-Host "Erro ao baixar o instalador do WireGuard: $_" -ForegroundColor Red
        Write-Host "Por favor, instale o WireGuard manualmente em: https://www.wireguard.com/install/" -ForegroundColor Yellow
        exit 1
    }
}
else {
    Write-Host "WireGuard encontrado! OK!" -ForegroundColor Green
}

# Criar diretórios de instalação
Write-Host "Criando diretórios..." -ForegroundColor Yellow
New-Item -ItemType Directory -Force -Path $InstallDir | Out-Null
New-Item -ItemType Directory -Force -Path $ConfigDir | Out-Null
New-Item -ItemType Directory -Force -Path $LogDir | Out-Null
New-Item -ItemType Directory -Force -Path $BinDir | Out-Null

# Copiar arquivos
Write-Host "Copiando arquivos..." -ForegroundColor Yellow
Copy-Item -Path ".\p2p-vpn.exe" -Destination "$BinDir\p2p-vpn.exe" -Force
Copy-Item -Path ".\config.yaml.example" -Destination "$ConfigDir\config.yaml" -Force

# Criar usuários se não existir
if (-Not (Test-Path "$ConfigDir\users")) {
    New-Item -ItemType Directory -Force -Path "$ConfigDir\users" | Out-Null
    New-Item -ItemType File -Force -Path "$ConfigDir\users\users.json" | Out-Null
}

# Instalar o serviço Windows
Write-Host "Instalando serviço Windows..." -ForegroundColor Yellow

# Remover serviço se já existir
if (Get-Service -Name $ServiceName -ErrorAction SilentlyContinue) {
    Write-Host "Serviço já existe. Removendo..." -ForegroundColor Yellow
    Stop-Service -Name $ServiceName -Force -ErrorAction SilentlyContinue
    Start-Sleep -Seconds 2
    sc.exe delete $ServiceName
    Start-Sleep -Seconds 2
}

# Criar novo serviço
$ServicePath = "$BinDir\p2p-vpn.exe --config $ConfigDir\config.yaml --security-config $ConfigDir\security.yaml --web-port 8080"
sc.exe create $ServiceName binPath= $ServicePath start= auto DisplayName= "P2P VPN Service"
sc.exe description $ServiceName "Serviço para a VPN P2P multiplataforma"

# Definir dependências e privilégios
sc.exe privs $ServiceName SeChangeNotifyPrivilege/SeImpersonatePrivilege/SeAssignPrimaryTokenPrivilege/SeIncreaseQuotaPrivilege

# Iniciar o serviço
Write-Host "Iniciando serviço P2P-VPN..." -ForegroundColor Yellow
Start-Service -Name $ServiceName -ErrorAction SilentlyContinue

# Verificar se o serviço está em execução
$Service = Get-Service -Name $ServiceName -ErrorAction SilentlyContinue
if ($Service -and $Service.Status -eq "Running") {
    Write-Host "Serviço iniciado com sucesso!" -ForegroundColor Green
} 
else {
    Write-Host "Atenção: O serviço não foi iniciado automaticamente." -ForegroundColor Yellow
    Write-Host "Você pode iniciá-lo manualmente através do Gerenciador de Serviços." -ForegroundColor Yellow
}

# Adicionar regras de firewall
Write-Host "Configurando regras de firewall..." -ForegroundColor Yellow
New-NetFirewallRule -DisplayName "P2P-VPN Web Interface" -Direction Inbound -Action Allow -Protocol TCP -LocalPort 8080 -Program "$BinDir\p2p-vpn.exe" -ErrorAction SilentlyContinue | Out-Null
New-NetFirewallRule -DisplayName "P2P-VPN WireGuard" -Direction Inbound -Action Allow -Protocol UDP -LocalPort 51820 -Program "$BinDir\p2p-vpn.exe" -ErrorAction SilentlyContinue | Out-Null
New-NetFirewallRule -DisplayName "P2P-VPN Discovery" -Direction Inbound -Action Allow -Protocol UDP -LocalPort 51821 -Program "$BinDir\p2p-vpn.exe" -ErrorAction SilentlyContinue | Out-Null

Write-Host "Instalação concluída!" -ForegroundColor Green
Write-Host "Interface web disponível em: http://localhost:8080" -ForegroundColor Yellow
Write-Host "Arquivos de configuração em: $ConfigDir" -ForegroundColor Yellow

# Criar atalho na área de trabalho
$WshShell = New-Object -ComObject WScript.Shell
$Shortcut = $WshShell.CreateShortcut("$env:PUBLIC\Desktop\P2P-VPN.lnk")
$Shortcut.TargetPath = "http://localhost:8080"
$Shortcut.Save()

Write-Host "Um atalho para a interface web foi criado na área de trabalho." -ForegroundColor Green

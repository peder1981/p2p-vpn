# Ícones da Aplicação P2P VPN

Este diretório contém os ícones utilizados pela interface gráfica da aplicação P2P VPN. É necessário criar ou incluir os seguintes arquivos de ícones:

- `app_icon.png` - Ícone principal da aplicação (256x256 pixels recomendado)
- `tray_disconnected.png` - Ícone para a bandeja do sistema quando desconectado (22x22 pixels)
- `tray_connected.png` - Ícone para a bandeja do sistema quando conectado (22x22 pixels)
- `status_connected.png` - Ícone para indicar status conectado (32x32 pixels)
- `status_disconnected.png` - Ícone para indicar status desconectado (32x32 pixels)

## Formatos suportados

- Em sistemas Linux: PNG e SVG são recomendados
- Em macOS: PNG e ICNS são recomendados
- Em Windows: PNG e ICO são recomendados

## Como gerar ícones para diferentes plataformas

### Convertendo para ICO (Windows)
```bash
convert app_icon.png -define icon:auto-resize=256,128,64,48,32,16 app_icon.ico
```

### Convertendo para ICNS (macOS)
```bash
# Requer o utilitário iconutil do macOS
mkdir MyIcon.iconset
sips -z 16 16     app_icon.png --out MyIcon.iconset/icon_16x16.png
sips -z 32 32     app_icon.png --out MyIcon.iconset/icon_16x16@2x.png
sips -z 32 32     app_icon.png --out MyIcon.iconset/icon_32x32.png
sips -z 64 64     app_icon.png --out MyIcon.iconset/icon_32x32@2x.png
sips -z 128 128   app_icon.png --out MyIcon.iconset/icon_128x128.png
sips -z 256 256   app_icon.png --out MyIcon.iconset/icon_128x128@2x.png
sips -z 256 256   app_icon.png --out MyIcon.iconset/icon_256x256.png
sips -z 512 512   app_icon.png --out MyIcon.iconset/icon_256x256@2x.png
sips -z 512 512   app_icon.png --out MyIcon.iconset/icon_512x512.png
sips -z 1024 1024 app_icon.png --out MyIcon.iconset/icon_512x512@2x.png
iconutil -c icns MyIcon.iconset
```

## Recomendações para os ícones

- Utilize ícones com fundo transparente (PNG)
- Mantenha um design consistente em todos os ícones
- Para o ícone da bandeja, use cores distintas para os estados conectado (verde) e desconectado (cinza/vermelho)
- Considere as diretrizes de design para cada sistema operacional

## Recursos para criação de ícones

- [Inkscape](https://inkscape.org/) - Editor SVG gratuito
- [GIMP](https://www.gimp.org/) - Editor de imagens gratuito
- [Figma](https://www.figma.com/) - Ferramenta de design online
- [Material Design Icons](https://material.io/resources/icons/) - Coleção de ícones gratuitos

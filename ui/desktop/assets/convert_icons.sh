#!/bin/bash

# Script para converter ícones SVG para formatos específicos das plataformas
# Script to convert SVG icons to platform-specific formats
# Script para convertir iconos SVG a formatos específicos de las plataformas

# Verifica dependências / Check dependencies / Verificar dependencias
echo "Verificando dependências..."
MISSING_DEPS=0

if ! command -v convert &> /dev/null; then
    echo "⚠️ ImageMagick não encontrado! Por favor, instale-o com:"
    echo "    sudo apt-get install imagemagick"
    MISSING_DEPS=1
fi

if ! command -v inkscape &> /dev/null; then
    echo "⚠️ Inkscape não encontrado! Por favor, instale-o com:"
    echo "    sudo apt-get install inkscape"
    MISSING_DEPS=1
fi

if [ $MISSING_DEPS -eq 1 ]; then
    echo "❌ Dependências faltando. Por favor, instale-as e tente novamente."
    exit 1
fi

# Diretórios
ICONS_DIR="icons"
OUTPUT_DIR="$ICONS_DIR/platforms"

# Criar diretórios de saída
mkdir -p "$OUTPUT_DIR/windows"
mkdir -p "$OUTPUT_DIR/macos"
mkdir -p "$OUTPUT_DIR/linux"

echo "🔄 Convertendo ícones para diferentes plataformas..."

# Converter para PNG em várias resoluções (para todas as plataformas)
echo "📏 Gerando PNGs em várias resoluções..."
for ICON in app_icon tray_connected tray_disconnected status_connected status_disconnected; do
    inkscape -w 16 -h 16 "$ICONS_DIR/${ICON}.svg" -o "$OUTPUT_DIR/${ICON}_16.png"
    inkscape -w 22 -h 22 "$ICONS_DIR/${ICON}.svg" -o "$OUTPUT_DIR/${ICON}_22.png"
    inkscape -w 24 -h 24 "$ICONS_DIR/${ICON}.svg" -o "$OUTPUT_DIR/${ICON}_24.png"
    inkscape -w 32 -h 32 "$ICONS_DIR/${ICON}.svg" -o "$OUTPUT_DIR/${ICON}_32.png"
    inkscape -w 48 -h 48 "$ICONS_DIR/${ICON}.svg" -o "$OUTPUT_DIR/${ICON}_48.png"
    inkscape -w 64 -h 64 "$ICONS_DIR/${ICON}.svg" -o "$OUTPUT_DIR/${ICON}_64.png"
    inkscape -w 128 -h 128 "$ICONS_DIR/${ICON}.svg" -o "$OUTPUT_DIR/${ICON}_128.png"
    inkscape -w 256 -h 256 "$ICONS_DIR/${ICON}.svg" -o "$OUTPUT_DIR/${ICON}_256.png"
done

# Criar ícone ICO para Windows
echo "🪟 Gerando ICO para Windows..."
convert "$OUTPUT_DIR/app_icon_16.png" "$OUTPUT_DIR/app_icon_24.png" "$OUTPUT_DIR/app_icon_32.png" \
        "$OUTPUT_DIR/app_icon_48.png" "$OUTPUT_DIR/app_icon_64.png" "$OUTPUT_DIR/app_icon_128.png" \
        "$OUTPUT_DIR/app_icon_256.png" \
        "$OUTPUT_DIR/windows/app_icon.ico"

convert "$OUTPUT_DIR/tray_connected_16.png" "$OUTPUT_DIR/tray_connected_22.png" "$OUTPUT_DIR/tray_connected_24.png" \
        "$OUTPUT_DIR/windows/tray_connected.ico"

convert "$OUTPUT_DIR/tray_disconnected_16.png" "$OUTPUT_DIR/tray_disconnected_22.png" "$OUTPUT_DIR/tray_disconnected_24.png" \
        "$OUTPUT_DIR/windows/tray_disconnected.ico"

# Copiar PNGs para Linux
echo "🐧 Preparando ícones para Linux..."
cp "$OUTPUT_DIR/app_icon_256.png" "$OUTPUT_DIR/linux/app_icon.png"
cp "$OUTPUT_DIR/tray_connected_22.png" "$OUTPUT_DIR/linux/tray_connected.png"
cp "$OUTPUT_DIR/tray_disconnected_22.png" "$OUTPUT_DIR/linux/tray_disconnected.png"
cp "$OUTPUT_DIR/status_connected_32.png" "$OUTPUT_DIR/linux/status_connected.png"
cp "$OUTPUT_DIR/status_disconnected_32.png" "$OUTPUT_DIR/linux/status_disconnected.png"

# Também copiar os SVGs originais para Linux
cp "$ICONS_DIR/"*.svg "$OUTPUT_DIR/linux/"

# Preparar para macOS (macOS pode usar PNG, mas ICNS é preferível se disponível)
echo "🍎 Preparando ícones para macOS..."
cp "$OUTPUT_DIR/app_icon_256.png" "$OUTPUT_DIR/macos/app_icon.png"
cp "$OUTPUT_DIR/tray_connected_22.png" "$OUTPUT_DIR/macos/tray_connected.png"
cp "$OUTPUT_DIR/tray_disconnected_22.png" "$OUTPUT_DIR/macos/tray_disconnected.png"
cp "$OUTPUT_DIR/status_connected_32.png" "$OUTPUT_DIR/macos/status_connected.png"
cp "$OUTPUT_DIR/status_disconnected_32.png" "$OUTPUT_DIR/macos/status_disconnected.png"

echo "✅ Conversão concluída! Ícones disponíveis em $OUTPUT_DIR"
echo "   - Windows: $OUTPUT_DIR/windows/"
echo "   - macOS:   $OUTPUT_DIR/macos/"
echo "   - Linux:   $OUTPUT_DIR/linux/"
echo ""
echo "📋 Nota: Para criar um arquivo ICNS para macOS, é necessário um ambiente macOS."
echo "   Em um ambiente macOS, você pode usar o comando iconutil."

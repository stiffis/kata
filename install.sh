#!/bin/bash

# Colores para el output
GREEN='\033[0;32m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}🥋 KATA Installer${NC}"
echo "-----------------------------------"

# 1. Verificar si Go está instalado
if ! command -v go &> /dev/null; then
    echo -e "${RED}Error: Go no está instalado.${NC}"
    echo "Por favor instala Go 1.21+ para continuar."
    exit 1
fi

# 2. Compilar el proyecto
echo -e "🔨 Compilando KATA..."
go build -o kata ./cmd/kata

if [ $? -ne 0 ]; then
    echo -e "${RED}❌ Error en la compilación.${NC}"
    exit 1
fi
echo -e "${GREEN}✓ Compilación exitosa.${NC}"

# 3. Determinar directorio de instalación
INSTALL_DIR="/usr/local/bin"
echo -e "📦 Instalando en $INSTALL_DIR..."

# 4. Mover el binario (usando sudo si es necesario)
if [ -w "$INSTALL_DIR" ]; then
    mv kata "$INSTALL_DIR/kata"
else
    echo "⚠️  Se requieren permisos de administrador para escribir en $INSTALL_DIR"
    sudo mv kata "$INSTALL_DIR/kata"
fi

if [ $? -eq 0 ]; then
    echo "-----------------------------------"
    echo -e "${GREEN}✅ ¡Instalación completada con éxito!${NC}"
    echo -e "🚀 Ahora puedes ejecutar ${BLUE}kata${NC} desde cualquier lugar."
else
    echo -e "${RED}❌ Falló la instalación.${NC}"
    exit 1
fi

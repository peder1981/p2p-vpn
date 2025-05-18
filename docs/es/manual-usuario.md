# Manual de Usuario: VPN P2P Universal

## Índice
1. [Introducción](#introducción)
2. [Requisitos del Sistema](#requisitos-del-sistema)
3. [Instalación](#instalación)
4. [Primeros Pasos](#primeros-pasos)
5. [Interfaz de Usuario](#interfaz-de-usuario)
6. [Configuración de Redes](#configuración-de-redes)
7. [Gestión de Peers](#gestión-de-peers)
8. [Seguridad](#seguridad)
9. [Monitoreo y Métricas](#monitoreo-y-métricas)
10. [Resolución de Problemas](#resolución-de-problemas)
11. [Preguntas Frecuentes (FAQ)](#preguntas-frecuentes-faq)
12. [Soporte y Contacto](#soporte-y-contacto)

## Introducción

VPN P2P Universal es una solución de Red Privada Virtual peer-to-peer (P2P) gratuita y de código abierto diseñada para proporcionar conexiones seguras entre dispositivos sin depender de servidores centralizados. Utilizando el protocolo WireGuard®, nuestra solución ofrece cifrado de alto rendimiento y comunicaciones de baja latencia.

### Principales Ventajas
- **Conexión Directa**: Establece conexiones peer-to-peer directas, eliminando la necesidad de servidores intermediarios
- **Alto Rendimiento**: Aprovecha la alta velocidad y baja latencia que ofrece el protocolo WireGuard
- **Compatibilidad Universal**: Disponible para Windows, macOS y Linux
- **Código Abierto**: Totalmente auditable y libre para usar, modificar y distribuir
- **Implementación Dual**: Elige entre modos kernel y userspace para mayor compatibilidad

## Requisitos del Sistema

### Windows
- Windows 10 o 11 (64-bit)
- 100 MB de espacio en disco
- 4 GB de RAM
- Conexión a Internet
- Privilegios de administrador para la instalación

### macOS
- macOS 10.15 (Catalina) o superior
- 100 MB de espacio en disco
- 4 GB de RAM
- Conexión a Internet

### Linux
- Kernel Linux 5.6 o superior (para modo kernel nativo)
- Distribuciones soportadas: Ubuntu 20.04+, Debian 11+, Fedora 34+, CentOS/RHEL 8+
- 100 MB de espacio en disco
- 2 GB de RAM
- Conexión a Internet
- Privilegios de superusuario para la instalación

## Instalación

### Windows
1. Descarga el instalador (.msi) desde la [página de releases](https://github.com/p2p-vpn/p2p-vpn/releases)
2. Ejecuta el archivo .msi con privilegios de administrador
3. Sigue las instrucciones del asistente de instalación
4. Al finalizar, la aplicación VPN P2P Universal estará disponible en el menú Inicio

### macOS
1. Descarga el instalador (.dmg) desde la [página de releases](https://github.com/p2p-vpn/p2p-vpn/releases)
2. Abre el archivo .dmg y arrastra la aplicación a la carpeta Aplicaciones
3. En la primera ejecución, autoriza la aplicación en el panel de Seguridad y Privacidad
4. Permite la instalación de componentes del sistema cuando se te solicite

### Linux
#### Usando el instalador automatizado
```bash
curl -sSL https://install.p2p-vpn.com | sudo bash
```

#### Usando paquetes específicos de la distribución
**Ubuntu/Debian:**
```bash
# Añade la clave GPG del repositorio
curl -fsSL https://repo.p2p-vpn.com/gpg | sudo gpg --dearmor -o /usr/share/keyrings/p2p-vpn-archive-keyring.gpg

# Añade el repositorio
echo "deb [arch=amd64 signed-by=/usr/share/keyrings/p2p-vpn-archive-keyring.gpg] https://repo.p2p-vpn.com stable main" | sudo tee /etc/apt/sources.list.d/p2p-vpn.list

# Actualiza e instala
sudo apt update && sudo apt install p2p-vpn
```

**Fedora/CentOS:**
```bash
# Añade el repositorio
sudo dnf config-manager --add-repo https://repo.p2p-vpn.com/p2p-vpn.repo

# Instala el paquete
sudo dnf install p2p-vpn
```

#### Usando contenedores
```bash
docker pull p2p-vpn/p2p-vpn:latest
docker run -d --name p2p-vpn --cap-add NET_ADMIN --network host p2p-vpn/p2p-vpn:latest
```

## Primeros Pasos

### Iniciando la Aplicación
1. Inicia la aplicación VPN P2P Universal desde el menú de aplicaciones de tu sistema
2. En la primera ejecución, serás guiado a través de un proceso de configuración inicial
3. Crea tu identidad VPN (claves pública y privada)
4. Configura tus preferencias básicas de conectividad

### Conectándose por Primera Vez
1. En la pantalla principal, haz clic en el botón "Crear Nueva Red" o "Unirse a Red"
2. Para crear una red:
   - Elige un nombre para tu red
   - Configura el espacio de direcciones IP (CIDR)
   - Establece permisos de acceso
   - Comparte el código de invitación con otros usuarios

3. Para unirse a una red:
   - Introduce el código de invitación proporcionado
   - O escanea el código QR si está disponible
   - Haz clic en "Conectar"

## Interfaz de Usuario

### Visión General de la Interfaz de Escritorio
![Interfaz de Escritorio](https://docs.p2p-vpn.com/images/desktop_interface.png)

- **Barra de Estado**: Muestra el estado actual de la conexión y estadísticas
- **Panel de Redes**: Lista todas tus redes configuradas
- **Panel de Peers**: Muestra todos los peers activos en la red actual
- **Botones de Acción Rápida**:
  - Conectar/Desconectar
  - Añadir Nuevo Peer
  - Configuración

### Visión General de la Interfaz Web
![Interfaz Web](https://docs.p2p-vpn.com/images/web_interface.png)

- **Dashboard**: Visión general del estado y métricas
- **Gestión de Redes**: Página para gestionar tus redes
- **Gestión de Peers**: Configuración de peers autorizados
- **Configuración**: Preferencias y configuraciones avanzadas
- **Logs**: Registro de actividades y diagnósticos

### Iconos de Estado
- **Verde**: Conexión activa y funcionando correctamente
- **Amarillo**: Conexión parcial (algunos peers no están accesibles)
- **Rojo**: No conectado o error en la conexión
- **Gris**: Servicio en pausa o inicializando

## Configuración de Redes

### Creación de Nueva Red
1. Accede al menú "Redes" > "Crear Nueva Red"
2. Define las siguientes configuraciones:
   - **Nombre de la Red**: Un identificador único para tu red
   - **Descripción**: Descripción opcional para el propósito de la red
   - **Espacio de Direcciones**: Define el bloque CIDR (ej: 10.0.0.0/24)
   - **Modo de Operación**:
     - Modo Malla (todos se conectan a todos)
     - Modo Estrella (todos se conectan a través de un hub)
   - **Política de Invitaciones**:
     - Abierta (cualquier persona con código puede entrar)
     - Aprobación manual (requiere tu aprobación)
     - Cerrada (solo por invitación directa)

### Gestión de Red
Para modificar una red existente:
1. Selecciona la red en la lista de redes
2. Haz clic en "Configuración" o en el icono de engranaje
3. Modifica los parámetros según sea necesario
4. Haz clic en "Guardar" para aplicar los cambios

### Eliminación de Red
1. Selecciona la red en la lista de redes
2. Haz clic en "Eliminar Red"
3. Confirma la operación cuando se te solicite

## Gestión de Peers

### Añadir Nuevos Peers
1. Selecciona la red donde deseas añadir peers
2. Haz clic en "Añadir Peer"
3. Elige uno de los métodos:
   - **Código de Invitación**: Genera un código y compártelo con el nuevo peer
   - **Archivo de Configuración**: Exporta un archivo de configuración
   - **Código QR**: Genera un código QR para dispositivos móviles

### Configuración de Peers
Para cada peer, puedes configurar:
- **Nombre Amigable**: Identificador para reconocer fácilmente el peer
- **Dirección IP**: Asignar una dirección IP específica dentro del CIDR de la red
- **Rutas Permitidas**: Configurar qué rutas puede anunciar este peer
- **Keepalive**: Configurar intervalos de keepalive para mantener conexiones a través de NAT
- **Endpoints**: Definir endpoints estáticos si es necesario

### Revocación de Acceso
1. Encuentra el peer en la lista de peers
2. Haz clic en "Revocar Acceso"
3. Confirma la operación cuando se te solicite
4. El peer será inmediatamente desconectado y no podrá volver a conectarse

## Seguridad

### Gestión de Claves
La aplicación gestiona automáticamente tus claves WireGuard, pero puedes:
- **Rotación de Claves**: Generar nuevas claves periódicamente para aumentar la seguridad
- **Backup de Claves**: Exportar tus claves a una ubicación segura
- **Importación de Claves**: Usar claves existentes en otra instalación

### Cifrados y Protocolos
- **WireGuard**: Utiliza los cifrados ChaCha20 para cifrado, Poly1305 para autenticación
- **Curva Criptográfica**: Curve25519 para intercambio de claves
- **Perfect Forward Secrecy**: Garantizado por el diseño del protocolo

### Configuración de Firewall
1. Accede a "Configuración" > "Seguridad" > "Firewall"
2. Configura reglas para controlar el tráfico:
   - **Reglas de Entrada**: Controla el tráfico recibido
   - **Reglas de Salida**: Controla el tráfico enviado
   - **Restricciones por IP/Puerto**: Limita el acceso a servicios específicos

## Monitoreo y Métricas

### Dashboard de Rendimiento
El dashboard muestra en tiempo real:
- **Tasa de Transferencia**: Subida y descarga actual
- **Latencia**: Tiempo de respuesta para cada peer
- **Pérdida de Paquetes**: Porcentaje de paquetes perdidos
- **Duración de la Conexión**: Tiempo desde el establecimiento de la conexión

### Logs y Diagnóstico
1. Accede a "Herramientas" > "Logs y Diagnóstico"
2. Selecciona el nivel de detalle:
   - **Básico**: Solo eventos principales
   - **Detallado**: Información completa para diagnóstico
   - **Depuración**: Información extensiva para solución de problemas

3. Utiliza las herramientas de diagnóstico:
   - **Ping**: Prueba de conectividad básica
   - **Traceroute**: Visualiza la ruta de los paquetes
   - **Verificación de MTU**: Identifica el tamaño óptimo de MTU
   - **Verificación de NAT**: Identifica el tipo de NAT en tu red

## Resolución de Problemas

### Problemas Comunes y Soluciones

#### No puede conectarse a otros peers
1. **Verificar firewall**: Asegúrate de que los puertos UDP necesarios estén abiertos
2. **Verificar NAT**: Ejecuta la prueba de tipo de NAT para verificar compatibilidad
3. **Verificar claves**: Confirma que las claves estén correctamente configuradas
4. **Probar endpoints alternativos**: Configura relays o STUN/TURN si es necesario

#### Conexión lenta o inestable
1. **Verificar calidad de la conexión**: Ejecuta una prueba de ancho de banda
2. **Ajustar MTU**: Prueba diferentes valores de MTU
3. **Verificar interferencia**: Comprueba si hay otras aplicaciones consumiendo ancho de banda
4. **Probar modo userspace**: Cambia al modo de implementación userspace

#### La aplicación no inicia
1. **Verificar permisos**: Asegúrate de tener privilegios suficientes
2. **Verificar logs**: Consulta los logs del sistema para mensajes de error
3. **Reinstalar**: Como último recurso, reinstala la aplicación

### Herramienta de Diagnóstico Automático
1. Accede a "Herramientas" > "Diagnóstico Automático"
2. Haz clic en "Iniciar Análisis"
3. El sistema verificará:
   - Conectividad de red
   - Configuración del sistema
   - Compatibilidad de hardware/software
   - Problemas conocidos
4. Sigue las recomendaciones presentadas en el informe

## Preguntas Frecuentes (FAQ)

**P: ¿La VPN P2P Universal es realmente gratuita?**
R: Sí, el software es completamente gratuito y de código abierto bajo la licencia MIT.

**P: ¿Puedo usar esta VPN para acceder a contenido geográficamente restringido?**
R: Como esta es una VPN P2P y no utiliza servidores de salida en diferentes países, no es ideal para eludir restricciones geográficas. Su propósito principal es crear redes privadas seguras entre dispositivos.

**P: ¿Cuántos dispositivos puedo conectar en una sola red?**
R: Teóricamente, no hay un límite estricto, pero recomendamos hasta 50 dispositivos para mantener un rendimiento óptimo. Para redes más grandes, considera crear múltiples sub-redes.

**P: ¿Cómo funciona el NAT traversal?**
R: La aplicación implementa múltiples técnicas de NAT traversal, incluyendo UDP hole punching, STUN, TURN y relays, seleccionando automáticamente la mejor opción para establecer la conexión.

**P: ¿Cuál es la diferencia entre los modos kernel y userspace?**
R: El modo kernel ofrece mejor rendimiento, pero requiere soporte en el kernel del sistema operativo. El modo userspace es más compatible, funcionando en prácticamente cualquier sistema, pero con una ligera reducción del rendimiento.

**P: ¿Mis datos se almacenan en algún servidor?**
R: No, la VPN P2P Universal no almacena ningún dato en servidores. Todas las configuraciones se almacenan localmente en tu dispositivo.

## Soporte y Contacto

### Recursos de Ayuda
- **Documentación**: [https://docs.p2p-vpn.com](https://docs.p2p-vpn.com)
- **Wiki**: [https://github.com/p2p-vpn/p2p-vpn/wiki](https://github.com/p2p-vpn/p2p-vpn/wiki)
- **Tutoriales en Vídeo**: [https://youtube.com/p2p-vpn](https://youtube.com/p2p-vpn)

### Comunidad
- **Foro**: [https://forum.p2p-vpn.com](https://forum.p2p-vpn.com)
- **Chat**: [https://chat.p2p-vpn.com](https://chat.p2p-vpn.com)
- **GitHub**: [https://github.com/p2p-vpn/p2p-vpn](https://github.com/p2p-vpn/p2p-vpn)

### Reportar Problemas
Si encuentras algún problema o bug:
1. Verifica si el problema ya ha sido reportado en la [lista de issues](https://github.com/p2p-vpn/p2p-vpn/issues)
2. Recopila información de diagnóstico usando la herramienta "Generar Informe de Diagnóstico"
3. Crea una nueva issue con detalles completos del problema y adjunta el informe de diagnóstico

---

© 2025 Proyecto VPN P2P Universal - Licenciado bajo MIT  
WireGuard® es una marca registrada de Jason A. Donenfeld.

---

Este manual está disponible en otros idiomas:
- [Português](https://docs.p2p-vpn.com/pt-BR/manual-usuario)
- [English](https://docs.p2p-vpn.com/en/user-manual)

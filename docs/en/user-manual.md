# User Manual: P2P Universal VPN

## Table of Contents
1. [Introduction](#introduction)
2. [System Requirements](#system-requirements)
3. [Installation](#installation)
4. [Getting Started](#getting-started)
5. [User Interface](#user-interface)
6. [Network Configuration](#network-configuration)
7. [Peer Management](#peer-management)
8. [Security](#security)
9. [Monitoring and Metrics](#monitoring-and-metrics)
10. [Troubleshooting](#troubleshooting)
11. [Frequently Asked Questions (FAQ)](#frequently-asked-questions-faq)
12. [Support and Contact](#support-and-contact)

## Introduction

P2P Universal VPN is a free and open-source peer-to-peer (P2P) Virtual Private Network solution designed to provide secure connections between devices without relying on centralized servers. Using the WireGuard® protocol, our solution offers high-performance encryption and low-latency communications.

### Key Advantages
- **Direct Connection**: Establish direct peer-to-peer connections, eliminating the need for intermediary servers
- **High Performance**: Leverage the high speed and low latency offered by the WireGuard protocol
- **Universal Compatibility**: Available for Windows, macOS, and Linux
- **Open Source**: Fully auditable and free to use, modify, and distribute
- **Dual Implementation**: Choose between kernel and userspace modes for greater compatibility

## System Requirements

### Windows
- Windows 10 or 11 (64-bit)
- 100 MB of disk space
- 4 GB of RAM
- Internet connection
- Administrator privileges for installation

### macOS
- macOS 10.15 (Catalina) or higher
- 100 MB of disk space
- 4 GB of RAM
- Internet connection

### Linux
- Linux Kernel 5.6 or higher (for native kernel mode)
- Supported distributions: Ubuntu 20.04+, Debian 11+, Fedora 34+, CentOS/RHEL 8+
- 100 MB of disk space
- 2 GB of RAM
- Internet connection
- Superuser privileges for installation

## Installation

### Windows
1. Download the installer (.msi) from the [releases page](https://github.com/p2p-vpn/p2p-vpn/releases)
2. Run the .msi file with administrator privileges
3. Follow the installation wizard instructions
4. After completion, the P2P Universal VPN application will be available in the Start menu

### macOS
1. Download the installer (.dmg) from the [releases page](https://github.com/p2p-vpn/p2p-vpn/releases)
2. Open the .dmg file and drag the application to the Applications folder
3. On first run, authorize the application in the Security & Privacy panel
4. Allow system components installation when prompted

### Linux
#### Using the automated installer
```bash
curl -sSL https://install.p2p-vpn.com | sudo bash
```

#### Using distribution-specific packages
**Ubuntu/Debian:**
```bash
# Add the GPG key for the repository
curl -fsSL https://repo.p2p-vpn.com/gpg | sudo gpg --dearmor -o /usr/share/keyrings/p2p-vpn-archive-keyring.gpg

# Add the repository
echo "deb [arch=amd64 signed-by=/usr/share/keyrings/p2p-vpn-archive-keyring.gpg] https://repo.p2p-vpn.com stable main" | sudo tee /etc/apt/sources.list.d/p2p-vpn.list

# Update and install
sudo apt update && sudo apt install p2p-vpn
```

**Fedora/CentOS:**
```bash
# Add the repository
sudo dnf config-manager --add-repo https://repo.p2p-vpn.com/p2p-vpn.repo

# Install the package
sudo dnf install p2p-vpn
```

#### Using containers
```bash
docker pull p2p-vpn/p2p-vpn:latest
docker run -d --name p2p-vpn --cap-add NET_ADMIN --network host p2p-vpn/p2p-vpn:latest
```

## Getting Started

### Starting the Application
1. Launch the P2P Universal VPN application from your system's application menu
2. On first run, you'll be guided through an initial setup process
3. Create your VPN identity (public and private keys)
4. Configure your basic connectivity preferences

### Connecting for the First Time
1. On the main screen, click the "Create New Network" or "Join Network" button
2. To create a network:
   - Choose a name for your network
   - Configure the IP address space (CIDR)
   - Set access permissions
   - Share the invitation code with other users

3. To join a network:
   - Enter the provided invitation code
   - Or scan the QR code if available
   - Click "Connect"

## User Interface

### Desktop Interface Overview
![Desktop Interface](https://docs.p2p-vpn.com/images/desktop_interface.png)

- **Status Bar**: Displays current connection status and statistics
- **Networks Panel**: Lists all your configured networks
- **Peers Panel**: Shows all active peers in the current network
- **Quick Action Buttons**:
  - Connect/Disconnect
  - Add New Peer
  - Settings

### Web Interface Overview
![Web Interface](https://docs.p2p-vpn.com/images/web_interface.png)

- **Dashboard**: Overview of status and metrics
- **Network Management**: Page for managing your networks
- **Peer Management**: Configuration of authorized peers
- **Settings**: Preferences and advanced configurations
- **Logs**: Activity logs and diagnostics

### Status Icons
- **Green**: Active and properly functioning connection
- **Yellow**: Partial connection (some peers are not accessible)
- **Red**: Not connected or connection error
- **Gray**: Service paused or initializing

## Network Configuration

### Creating a New Network
1. Access the "Networks" > "Create New Network" menu
2. Define the following settings:
   - **Network Name**: A unique identifier for your network
   - **Description**: Optional description for the network's purpose
   - **Address Space**: Define the CIDR block (e.g., 10.0.0.0/24)
   - **Operation Mode**:
     - Mesh Mode (all connect to all)
     - Star Mode (all connect through a hub)
   - **Invitation Policy**:
     - Open (anyone with code can join)
     - Manual approval (requires your approval)
     - Closed (by direct invitation only)

### Network Management
To modify an existing network:
1. Select the network in the network list
2. Click "Settings" or the gear icon
3. Modify parameters as needed
4. Click "Save" to apply the changes

### Deleting a Network
1. Select the network in the network list
2. Click "Delete Network"
3. Confirm the operation when prompted

## Peer Management

### Adding New Peers
1. Select the network where you want to add peers
2. Click "Add Peer"
3. Choose one of the methods:
   - **Invitation Code**: Generate a code and share it with the new peer
   - **Configuration File**: Export a configuration file
   - **QR Code**: Generate a QR code for mobile devices

### Peer Configuration
For each peer, you can configure:
- **Friendly Name**: Identifier to easily recognize the peer
- **IP Address**: Assign a specific IP address within the network's CIDR
- **Allowed Routes**: Configure which routes this peer can announce
- **Keepalive**: Configure keepalive intervals to maintain connections through NAT
- **Endpoints**: Define static endpoints if necessary

### Revoking Access
1. Find the peer in the peers list
2. Click "Revoke Access"
3. Confirm the operation when prompted
4. The peer will be immediately disconnected and will no longer be able to connect

## Security

### Key Management
The application automatically manages your WireGuard keys, but you can:
- **Key Rotation**: Generate new keys periodically to enhance security
- **Key Backup**: Export your keys to a secure location
- **Key Import**: Use existing keys in another installation

### Ciphers and Protocols
- **WireGuard**: Uses ChaCha20 for encryption, Poly1305 for authentication
- **Cryptographic Curve**: Curve25519 for key exchange
- **Perfect Forward Secrecy**: Guaranteed by the protocol design

### Firewall Settings
1. Access "Settings" > "Security" > "Firewall"
2. Configure rules to control traffic:
   - **Inbound Rules**: Control received traffic
   - **Outbound Rules**: Control sent traffic
   - **IP/Port Restrictions**: Limit access to specific services

## Monitoring and Metrics

### Performance Dashboard
The dashboard displays in real-time:
- **Throughput**: Current upload and download rates
- **Latency**: Response time for each peer
- **Packet Loss**: Percentage of lost packets
- **Connection Duration**: Time since the connection was established

### Logs and Diagnostics
1. Access "Tools" > "Logs and Diagnostics"
2. Select the detail level:
   - **Basic**: Only main events
   - **Detailed**: Complete information for diagnostics
   - **Debug**: Extensive information for troubleshooting

3. Use the diagnostic tools:
   - **Ping**: Basic connectivity test
   - **Traceroute**: Visualize the packet route
   - **MTU Check**: Identify the optimal MTU size
   - **NAT Check**: Identify the NAT type in your network

## Troubleshooting

### Common Issues and Solutions

#### Cannot connect to other peers
1. **Check firewall**: Ensure that the necessary UDP ports are open
2. **Check NAT**: Run the NAT type test to verify compatibility
3. **Check keys**: Confirm that the keys are correctly configured
4. **Try alternative endpoints**: Configure relays or STUN/TURN if necessary

#### Slow or unstable connection
1. **Check connection quality**: Run a bandwidth test
2. **Adjust MTU**: Try different MTU values
3. **Check interference**: Verify if there are other applications consuming bandwidth
4. **Try userspace mode**: Switch to userspace implementation mode

#### Application does not start
1. **Check permissions**: Make sure you have sufficient privileges
2. **Check logs**: Consult system logs for error messages
3. **Reinstall**: As a last resort, reinstall the application

### Automatic Diagnostic Tool
1. Access "Tools" > "Automatic Diagnostic"
2. Click "Start Analysis"
3. The system will check:
   - Network connectivity
   - System configuration
   - Hardware/software compatibility
   - Known issues
4. Follow the recommendations presented in the report

## Frequently Asked Questions (FAQ)

**Q: Is P2P Universal VPN really free?**
A: Yes, the software is completely free and open-source under the MIT license.

**Q: Can I use this VPN to access geographically restricted content?**
A: As this is a P2P VPN and does not use exit servers in different countries, it is not ideal for bypassing geographical restrictions. Its main purpose is to create secure private networks between devices.

**Q: How many devices can I connect in a single network?**
A: Theoretically, there is no hard limit, but we recommend up to 50 devices to maintain optimal performance. For larger networks, consider creating multiple sub-networks.

**Q: How does NAT traversal work?**
A: The application implements multiple NAT traversal techniques, including UDP hole punching, STUN, TURN, and relays, automatically selecting the best option to establish the connection.

**Q: What is the difference between kernel and userspace modes?**
A: Kernel mode offers better performance but requires support in the operating system kernel. Userspace mode is more compatible, working on virtually any system, but with a slight performance reduction.

**Q: Is my data stored on any server?**
A: No, P2P Universal VPN does not store any data on servers. All configurations are stored locally on your device.

## Support and Contact

### Help Resources
- **Documentation**: [https://docs.p2p-vpn.com](https://docs.p2p-vpn.com)
- **Wiki**: [https://github.com/p2p-vpn/p2p-vpn/wiki](https://github.com/p2p-vpn/p2p-vpn/wiki)
- **Video Tutorials**: [https://youtube.com/p2p-vpn](https://youtube.com/p2p-vpn)

### Community
- **Forum**: [https://forum.p2p-vpn.com](https://forum.p2p-vpn.com)
- **Chat**: [https://chat.p2p-vpn.com](https://chat.p2p-vpn.com)
- **GitHub**: [https://github.com/p2p-vpn/p2p-vpn](https://github.com/p2p-vpn/p2p-vpn)

### Reporting Issues
If you find any issues or bugs:
1. Check if the issue has already been reported in the [issues list](https://github.com/p2p-vpn/p2p-vpn/issues)
2. Collect diagnostic information using the "Generate Diagnostic Report" tool
3. Create a new issue with full details of the problem and attach the diagnostic report

---

© 2025 P2P Universal VPN Project - Licensed under MIT  
WireGuard® is a registered trademark of Jason A. Donenfeld.

---

This manual is available in other languages:
- [Português](https://docs.p2p-vpn.com/pt-BR/manual-usuario)
- [Español](https://docs.p2p-vpn.com/es/manual-usuario)

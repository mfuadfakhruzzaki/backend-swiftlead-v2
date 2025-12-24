#!/bin/bash
# ===========================================
# Swiftlet Backend - AWS EC2 Server Setup
# ===========================================
# Run this script on AWS CloudShell or directly on EC2 instance
# Usage: ./setup-server.sh

set -e

echo "🚀 Swiftlet Server Setup Script"
echo "================================"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Variables
DEPLOY_USER="deploy"
APP_DIR="/opt/swiftlet"
DOMAIN="api.swiftlead.id"  # Change this

# Check if running as root
if [ "$EUID" -ne 0 ]; then
    echo -e "${RED}Please run as root (sudo)${NC}"
    exit 1
fi

echo -e "${YELLOW}Step 1: System Update${NC}"
apt-get update && apt-get upgrade -y

echo -e "${YELLOW}Step 2: Install Dependencies${NC}"
apt-get install -y \
    apt-transport-https \
    ca-certificates \
    curl \
    gnupg \
    lsb-release \
    git \
    htop \
    vim \
    ufw \
    fail2ban

echo -e "${YELLOW}Step 3: Install Docker${NC}"
if ! command -v docker &> /dev/null; then
    curl -fsSL https://get.docker.com -o get-docker.sh
    sh get-docker.sh
    rm get-docker.sh
    systemctl enable docker
    systemctl start docker
fi

echo -e "${YELLOW}Step 4: Install Docker Compose${NC}"
if ! command -v docker-compose &> /dev/null; then
    curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
    chmod +x /usr/local/bin/docker-compose
fi

echo -e "${YELLOW}Step 5: Create Deploy User${NC}"
if ! id "$DEPLOY_USER" &>/dev/null; then
    useradd -m -s /bin/bash -G docker "$DEPLOY_USER"
    echo -e "${GREEN}Created user: $DEPLOY_USER${NC}"
else
    usermod -aG docker "$DEPLOY_USER"
    echo -e "${GREEN}User $DEPLOY_USER already exists, added to docker group${NC}"
fi

echo -e "${YELLOW}Step 6: Setup SSH for Deploy User${NC}"
mkdir -p /home/$DEPLOY_USER/.ssh
chmod 700 /home/$DEPLOY_USER/.ssh
touch /home/$DEPLOY_USER/.ssh/authorized_keys
chmod 600 /home/$DEPLOY_USER/.ssh/authorized_keys
chown -R $DEPLOY_USER:$DEPLOY_USER /home/$DEPLOY_USER/.ssh

echo -e "${YELLOW}Step 7: Create Application Directory${NC}"
mkdir -p $APP_DIR
chown -R $DEPLOY_USER:$DEPLOY_USER $APP_DIR

echo -e "${YELLOW}Step 8: Configure Firewall (UFW)${NC}"
ufw --force reset
ufw default deny incoming
ufw default allow outgoing
ufw allow ssh
ufw allow 80/tcp
ufw allow 443/tcp
ufw allow 8080/tcp   # Backend (temporary, remove after nginx setup)
ufw allow 9090/tcp   # Prometheus (internal only)
ufw allow 3000/tcp   # Grafana
ufw --force enable

echo -e "${YELLOW}Step 9: Configure Fail2Ban${NC}"
cat > /etc/fail2ban/jail.local << 'EOF'
[DEFAULT]
bantime = 3600
findtime = 600
maxretry = 5

[sshd]
enabled = true
port = ssh
filter = sshd
logpath = /var/log/auth.log
maxretry = 3
EOF
systemctl enable fail2ban
systemctl restart fail2ban

echo -e "${YELLOW}Step 10: Install Certbot for SSL${NC}"
apt-get install -y certbot

echo -e "${YELLOW}Step 11: Create Docker Network${NC}"
docker network create swiftlet-network 2>/dev/null || true

echo -e "${YELLOW}Step 12: Setup Log Rotation${NC}"
cat > /etc/logrotate.d/docker-containers << 'EOF'
/var/lib/docker/containers/*/*.log {
    rotate 7
    daily
    compress
    missingok
    delaycompress
    copytruncate
}
EOF

echo ""
echo -e "${GREEN}✅ Server Setup Complete!${NC}"
echo ""
echo "📋 Next Steps:"
echo "1. Add your deploy SSH public key:"
echo "   echo 'YOUR_PUBLIC_KEY' >> /home/$DEPLOY_USER/.ssh/authorized_keys"
echo ""
echo "2. Clone the repository:"
echo "   cd $APP_DIR"
echo "   git clone https://github.com/mfuadfakhruzzaki/backend-swiftlead-v2.git ."
echo ""
echo "3. Create .env file:"
echo "   cp .env.prod.example .env"
echo "   nano .env  # Edit with production values"
echo ""
echo "4. Get SSL certificate:"
echo "   certbot certonly --standalone -d $DOMAIN"
echo "   mkdir -p $APP_DIR/docker/nginx/ssl"
echo "   ln -s /etc/letsencrypt/live/$DOMAIN/fullchain.pem $APP_DIR/docker/nginx/ssl/"
echo "   ln -s /etc/letsencrypt/live/$DOMAIN/privkey.pem $APP_DIR/docker/nginx/ssl/"
echo ""
echo "5. Start the application:"
echo "   cd $APP_DIR"
echo "   docker-compose -f docker-compose.prod.yml up -d"
echo ""
echo "6. Setup SSL auto-renewal:"
echo "   echo '0 0 * * * certbot renew --quiet' | crontab -"
echo ""
echo -e "${YELLOW}Server IP: $(curl -s ifconfig.me)${NC}"

#!/bin/bash
# ================================
# Optional setup of firewall rules on VPS
#
# scp ufw.sh deploy@46.224.208.70:/opt/
# ./ufw.sh
# ================================

echo "=== Setting default policies ==="
sudo ufw default deny incoming
sudo ufw default allow outgoing

echo "=== Allow SSH for deploy user ==="
sudo ufw allow OpenSSH

echo "=== Allow HTTP/HTTPS for Caddy ==="
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp

# Optional: open additional host ports if you expose them in the future
# sudo ufw allow 8080/tcp   # Example: backend API host port
# sudo ufw allow 5432/tcp   # Example: Postgres remote (only if needed)

echo "=== Enabling UFW safely ==="
# Force enable, keep current session open
sudo ufw --force enable

echo "=== Status check ==="
sudo ufw status verbose
sudo ufw status numbered

echo "=== ✅ Done ==="
echo "Test SSH from another terminal before closing this session."


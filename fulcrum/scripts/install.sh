mkdir -p /opt/fulcrum
cp devfest2025 /opt/fulcrum/
cp .env /opt/fulcrum/
chown -R www-data:www-data /opt/fulcrum
chmod +x /opt/fulcrum/devfest2025
cp devfest2025.service /etc/systemd/system/
systemctl daemon-reload
systemctl enable devfest2025.service
systemctl start devfest2025.service

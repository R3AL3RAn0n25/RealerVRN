sudo mkdir -p /etc/vrn
sudo openssl req -x509 -nodes -days 365 -newkey rsa:4096 \
  -keyout /etc/vrn/privkey.pem -out /etc/vrn/fullchain.pem \
  -subj "/CN=www.microsoft.com"
# Gitomatically

<div align="center">
  <img src="https://github.com/khouwdevin/gitomatically/blob/master/public/gitomatically.png" height="300px"/>
</div>

Gitomatically is an app designed to implement CI/CD with ease. You can write the configuration, and Gitomatically will handle the rest.

## Setup

To configure Gitomatically for your desired repositories, create a `config.yaml` file in your project folder using the following guide:

```yaml
repositories:
  { repository-name (you can name it whatever you want) }:
    url: { github repository url }
    clone: { github clone url (using SSH is suggested) }
    branch: { which branch you want to pull }
    path: { path to your apps }
    builds:
      - { build commands, you can leave it empty if you don't need to build }
  example.com:
    url: https://github.com/example/example.com
    clone: git@github.com:example/example.com.git
    branch: main
    path: /home/khouwdevin/apps/example.com/
    builds:
      - docker compose up --build -d
```

Additionally, you need to create an `.env` file with the following content:

```env
GIN_MODE=release
GITHUB_WEBHOOK_SECRET="helloworld" # you can create a secret when you register the webhook
```

### (Optional) Enabling SSL with Certbot (Using Nginx Reverse Proxy)

To secure your Gitomatically application with SSL/TLS (HTTPS) using a free Let's Encrypt certificate, you'll typically set up a reverse proxy like Nginx to handle the SSL termination and forward requests to your Gitomatically app running on port 8080.

**Step 1: Create an Nginx Server Block Configuration**

This configuration tells Nginx to listen for requests to your domain and forward them to your Gitomatically application.

Create a new Nginx configuration file. The name `gitomatically.conf` is a good choice.

```bash
sudo nano /etc/nginx/sites-available/gitomatically.conf
```

Paste the following content into the file. Remember to replace `gitomatically.example.com` with your actual domain name.

```nginx
# /etc/nginx/sites-available/gitomatically.conf

server {
    server_name gitomatically.example.com;

    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

**Step 2: Enable the Nginx Server Block**

Nginx uses a system of sites-available (where configs are stored) and sites-enabled (where active configs are symlinked).

```bash
sudo ln -s /etc/nginx/sites-available/gitomatically.conf /etc/nginx/sites-enabled/gitomatically.conf
```

**Step 3: Test Nginx Configuration and Restart**

```bash
sudo nginx -t
sudo systemctl restart nginx
```

**Step 4: Install Certbot**

You can install certbot in a few different ways, the instruction will install certbot with snap.

```bash
sudo snap install --classic certbot
sudo ln -s /snap/bin/certbot /usr/bin/certbot
sudo certbot --nginx
```

## Run app

To run Gitomatically, you can use the provided `install.sh` script. You can also uninstall it using `uninstall.sh`.

## Notes

Currently, only GitHub is supported. This is because I primarily use GitHub. However, if you're interested in using Gitomatically with GitLab, please let me know by opening an issue.

Pull requests are always welcome!

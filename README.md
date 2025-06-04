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
    path: ~/apps/example.com/
    builds:
      - docker compose up --build -d
```

Additionally, you need to create an `.env` file with the following content:

```env
GIN_MODE=release
GITHUB_WEBHOOK_SECRET="helloworld" # you can get the secret sha256 when you register the webhook
```

## Run app

To run Gitomatically, you can use the provided `install.sh` script. You can also uninstall it using `uninstall.sh`.

## Notes

Currently, only GitHub is supported. This is because I primarily use GitHub. However, if you're interested in using Gitomatically with GitLab, please let me know by opening an issue.

Pull requests are always welcome!

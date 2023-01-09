# Azure ACR credential helper

Azure ACR credential helper is a [credential helper](https://github.com/docker/docker-credential-helpers) for docker that makes it easier to use the [Azure Container Registry](https://azure.microsoft.com/en-us/products/container-registry/#overview).

## Prerequisites

Docker credential helpers were introduced in Docker 1.11 which is the minimum version required.

You must be logged into either Azure CLI or have a managed identity on an Azure VM prior to use the credential helper.

## Installing

You can download the most recent version of the credential helper on the [releases page](https://github.com/ffwdeng/azure-acr-credential-helper/releases) for your desired operating system and architecture and make sure it's available in path for docker.

## Configuration

Put the `docker-credential-acr-login` binary in your `PATH` and add the following to your `~/.docker/config.json`.

```json
{
	"credsStore": "acr-login"
}
```

This will configure acr-login as your credential store for docker. This however will use acr-login as the credential store for every image you try to pull and `docker-credential-acr-login` only supports acr registries.

Starting with docker 1.13 you can set a credential store per registry. To use `docker-credential-acr-login` for specific acr registry put the following in `~/.docker/config.json` instead.

```json
{
	"credHelpers" {
		"<acr_id>.azurecr.io": "acr-login"
	}
}
```

### Azure credentials

The credential helper will try the following methods in the following order to obtain credentials.

- **Environment** - Reads account information specified in [environment variables](https://pkg.go.dev/github.com/Azure/azure-sdk-for-go/sdk/azidentity@v1.2.0#readme-environment-variables) and uses it to authenticate.
- **Managed Identity** - If the app is deployed to an Azure host with managed identity enabled, it will be used.
- **Azure CLI** - If a user or service principal has authenticated via the Azure CLI `az login` command, it will be used.

## Usage

`docker pull <acr_id>.azurecr.io/image:version`

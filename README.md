# DNS Sync

A simple declarative tool for managing DNS records.

DNS Sync is _alpha_ and under active development.

# Downloads

   * [Linux](https://github.com/brendandburns/dns-sync/releases/download/0.1.0/linux-amd64.tgz)
   * [OS X](https://github.com/brendandburns/dns-sync/releases/download/0.1.0/darwin-amd64.tgz)
   * [Windows](https://github.com/brendandburns/dns-sync/releases/download/0.1.0/windows-amd64.zip)

# Usage

Create a declarative YAML or JSON file:

```yaml
zone:
  # metadata
  name: example
  description: test-sync is changed

  # this needs to be a valid dns zone ending with a 'dot'
  dnsName: sync.contuso.io.

# supported record types A, CNAME, NS
records:
- kind: A
  ttl: 350
  name: www.sync.contuso.io.
  addresses:
  - 1.2.3.4
  - 2.3.4.5
- kind: CNAME
  ttl: 200
  name: cname.sync.contuso.io.
  canonicalName: some.other.company.com.
```

Then you can synchronize this as follows:

```sh
$ dns-sync --config sample.yaml
```

# Configuring cloud providers

DNS sync works can work with any DNS provider. Currently Google and Azure are supported.

You can use the `--cloud` flag to determine which you use.

Each cloud provider is configured differently.

## Google
The Google CloudDNS provider expects two environment variables:

   * `GOOGLE_APPLICATION_CREDENTIALS` should point to a JSON credentials file [details here](https://cloud.google.com/genomics/docs/how-tos/getting-started#download_credentials_for_api_access)
   * `GOOGLE_PROJECT` should have the name of the project where the DNS records should be created.

## Azure
The Azure DNS provider expects expects three environment variables:

   * `AZURE_SUBSCRIPTION` should point to the Azure subscription you want to use.
   * `AZURE_RESOURCE_GROUP` should point to the resource group you want the records to be placed in.
   * `AZURE_AUTH_LOCATION` should point to an auth file [details here](https://docs.microsoft.com/en-us/go/azure/azure-sdk-go-authorization#use-file-based-authentication)

# Building

For now, building is pretty manual.

```sh
go build cmd/dns-sync.go
```

You will likely need to `go get ...` a number of dependencies first. A better build system is coming...

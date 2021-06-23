Terraform Provider
==================

- Website: https://www.terraform.io
- [![Gitter chat](https://badges.gitter.im/hashicorp-terraform/Lobby.png)](https://gitter.im/hashicorp-terraform/Lobby)
- Mailing list: [Google Groups](http://groups.google.com/group/terraform-tool)

<img src="https://cdn.rawgit.com/hashicorp/terraform-website/master/content/source/assets/images/logo-hashicorp.svg" width="600px">

Requirements
------------

-	[Terraform](https://www.terraform.io/downloads.html) 0.10.x
-	[Go](https://golang.org/doc/install) 1.11 (to build the provider plugin)

Building The Provider
---------------------

Clone repository to: `$GOPATH/src/github.com/terraform-providers/terraform-provider-cloudscale`

```sh
$ mkdir -p $GOPATH/src/github.com/terraform-providers; cd $GOPATH/src/github.com/terraform-providers
$ git clone git@github.com:terraform-providers/terraform-provider-cloudscale
```

Enter the provider directory and build the provider

```sh
$ cd $GOPATH/src/github.com/terraform-providers/terraform-provider-cloudscale
$ make build
```

Using the provider
----------------------
1. Download and install [Terraform](https://www.terraform.io/downloads.html)
2. Move `terraform-provider-cloudscale` executable to the root folder where main.tf is. 

```shell
  $ mv terraform-provider-cloudscale /path/to/main.tf/folder
``` 

3. Execute `terraform plan`
4. `terraform init`

Developing the Provider
---------------------------

If you wish to work on the provider, you'll first need [Go](http://www.golang.org) installed on your machine (see [Requirements](#requirements) above).

To compile the provider, run `go install`. This will build the provider and put the provider binary in the `$GOPATH/bin` directory.

To generate or update documentation, run `go generate`.

In order to run the full suite of Acceptance tests, run `make testacc`.

*Note:* Acceptance tests create real resources, and often cost money to run.

```sh
$ make testacc
```

In order to run run a subset of the tests:

``` sh
$ TESTARGS="-run TestAccCloudscaleSubnet" make testacc
```

In order to upgrade the `cloudscale-go-sdk`.

```sh
go get -u github.com/cloudscale-ch/cloudscale-go-sdk
go mod vendor
```

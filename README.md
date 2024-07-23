cloudscale.ch Terraform Provider
==================

- Website: https://www.terraform.io
- [![Gitter chat](https://badges.gitter.im/hashicorp-terraform/Lobby.png)](https://gitter.im/hashicorp-terraform/Lobby)
- Mailing list: [Google Groups](http://groups.google.com/group/terraform-tool)

Requirements
------------

-	[Terraform](https://www.terraform.io/downloads.html) 0.12 or higher
-	[Go](https://golang.org/doc/install) to build the provider plugin

Developing the Provider
---------------------------

If you wish to work on the provider, you'll first need [Go](http://www.golang.org) installed on your machine (see [Requirements](#requirements) above).

To compile the provider, run `go install -mod vendor`. This will build the provider and put the provider binary in the `$GOPATH/bin` directory.

To generate or update documentation, run `go generate`.

In order to run the full suite of Acceptance tests, run `make testacc`.

*Notes:* 
 * Acceptance tests create real resources, and often cost money to run.
 * [See here](https://www.terraform.io/plugin/sdkv2/testing/acceptance-tests#terraform-cli-installation-behaviors)
   to understand which version of Terraform is used in your tests.

```sh
$ make testacc
```

In order to run a subset of the tests:

``` sh
$ TESTARGS="-run TestAccCloudscaleSubnet" make testacc
```

In order to upgrade the `cloudscale-go-sdk`.

```sh
go get -u github.com/cloudscale-ch/cloudscale-go-sdk
go mod vendor
```


Use the following commands to switch to a local version of the go-sdk and back:
```sh
go mod edit -replace "github.com/cloudscale-ch/cloudscale-go-sdk/v4=../cloudscale-go-sdk/"
go mod vendor
git commit -m "drop: Use local version of cloudscale-go-sdk"
```
```sh
go mod edit -dropreplace "github.com/cloudscale-ch/cloudscale-go-sdk/v4"
go mod vendor
```

To test unreleased driver versions locally add the following to your `~/.terraformrc`

```hcl
provider_installation {
  # Use go/bin as an overridden package directory
  # for the cloudscale-ch/cloudscale provider. This disables the version and checksum
  # verifications for this provider and forces Terraform to look for the
  # null provider plugin in the given directory.
  dev_overrides {
    "cloudscale-ch/cloudscale" = "/Users/alain/go/bin"
  }

  # For all other providers, install them directly from their origin provider
  # registries as normal. If you omit this, Terraform will _only_ use
  # the dev_overrides block, and so no other providers will be available.
  direct {}
}
```

To cross-compile a local build, run:

```
# goreleaser v1.x
docker run -it --rm -v $PWD:/app --workdir=/app goreleaser/goreleaser:v1.26.2 release --snapshot --rm-dist --skip-sign
# goreleaser v2.x
docker run -it --rm -v $PWD:/app --workdir=/app goreleaser/goreleaser:v2.1.0 release --snapshot --clean --skip=publish,sign
```

Releasing the Provider
---------------------------

 1. Ensure the `CHANGELOG.md` is up-to-date.
 2.  Create a new release [on GitHub](https://github.com/cloudscale-ch/terraform-provider-cloudscale/releases/new).
    Both the tag and release title must follow this pattern: `v<<SEMVER>>`.
    Examples: `v42.43.44` or `v1.33.7-rc.1`.
 3. It might take a moment until the release appears in the [Terraform registry](https://registry.terraform.io/providers/cloudscale-ch/cloudscale/latest).
    You can manually resync the provider when you are logged in to the registry. 


Developing the Documentation Website
------------------------------------

Use the Terraform [doc preview tool](https://registry.terraform.io/tools/doc-preview) to test markdown rendering.

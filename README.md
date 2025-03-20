# cloudscale.ch Terraform Provider

- Website: https://www.terraform.io
- [![Gitter chat](https://badges.gitter.im/hashicorp-terraform/Lobby.png)](https://gitter.im/hashicorp-terraform/Lobby)
- Mailing list: [Google Groups](http://groups.google.com/group/terraform-tool)

## Using the Provider

For detailed usage instructions and examples, please refer to the official documentation available
at [Terraform Registry: cloudscale-ch/cloudscale](https://registry.terraform.io/providers/cloudscale-ch/cloudscale/latest).

## Developing the Provider

Before you begin, make sure you have [Go](http://golang.org) installed on your machine.

### 1. Compile the Provider

Run the following command to compile the provider. The binary will be placed in your `$GOPATH/bin` directory.

```sh
go install
```

To create builds for different platforms, you can use [goreleaser](https://goreleaser.com/):

- **For goreleaser v1.x:**

  ```sh
  docker run -it --rm -v $PWD:/app --workdir=/app goreleaser/goreleaser:v1.26.2 release --snapshot --rm-dist --skip-sign
  ```

- **For goreleaser v2.x:**

  ```sh
  docker run -it --rm -v $PWD:/app --workdir=/app goreleaser/goreleaser:v2.1.0 release --snapshot --clean --skip=publish,sign
  ```

### 2. Testing Unreleased Driver Versions

To test unreleased driver versions, add the following to your `~/.terraformrc` file.
This configuration directs Terraform to use your local `go/bin` directory for the cloudscale provider:

```hcl
provider_installation {
  # Use go/bin as an overridden package directory
  # for the cloudscale-ch/cloudscale provider. This disables the version and checksum
  # verifications for this provider and forces Terraform to look for the
  # null provider plugin in the given directory.
  dev_overrides {
    "cloudscale-ch/cloudscale" = "/Users/[your-username]/go/bin"
  }

  # For all other providers, install them directly from their origin provider
  # registries as normal. If you omit this, Terraform will _only_ use
  # the dev_overrides block, and so no other providers will be available.
  direct {}
}
```

*Remember to replace `[your-username]` with your actual username.*

### 3. Generate or Update Documentation

Update the documentation by running:

```sh
go generate
```

### 4. Running Acceptance Tests

Acceptance tests create real resources and might incur costs. They also use a specific version of Terraform (see [Terraform CLI Installation Behaviors](https://www.terraform.io/plugin/sdkv2/testing/acceptance-tests#terraform-cli-installation-behaviors)).

- **Run all tests:**

  ```sh
  make testacc
  ```

- **Run a subset of the tests (e.g., tests for subnet):**

  ```sh
  TESTARGS="-run TestAccCloudscaleSubnet" make testacc
  ```

### 5. Upgrading the cloudscale-go-sdk

- **Upgrade to the latest version:**

  ```sh
  go get -u github.com/cloudscale-ch/cloudscale-go-sdk/v5
  go mod tidy
  ```

### 6. Working with Different Versions of the cloudscale-go-sdk

If you want to work with a local version or a specific version of the cloudscale-go-sdk during development, use the
following commands:

- **Replace with a local version:**

  ```sh
  go mod edit -replace "github.com/cloudscale-ch/cloudscale-go-sdk/v4=../cloudscale-go-sdk/"
  go mod tidy
  git commit -m "drop: Use local version of cloudscale-go-sdk"
  ```

- **Pin to a specific commit:**

  ```sh
  go mod edit -replace "github.com/cloudscale-ch/cloudscale-go-sdk/v4=github.com/cloudscale-ch/cloudscale-go-sdk/v4@<commit-hash>"
  go mod tidy
  git commit -m "drop: Pin specific commit of cloudscale-go-sdk"
  ```

- **Switch back to the upstream version:**

  ```sh
  go mod edit -dropreplace "github.com/cloudscale-ch/cloudscale-go-sdk/v4"
  go mod tidy
  ```

## Releasing the Provider

1. Ensure the `CHANGELOG.md` is up-to-date.
1. Ensure the `.github/workflows/terraform-integration-tests.yml` tests the 3 most recent Terraform versions.
1. Create a new release [on GitHub](https://github.com/cloudscale-ch/terraform-provider-cloudscale/releases/new).
   Both the tag and release title must follow this pattern: `v<<SEMVER>>`.
   Examples: `v42.43.44` or `v1.33.7-rc.1`.
1. It might take a moment until the release appears in the [Terraform registry](https://registry.terraform.io/providers/cloudscale-ch/cloudscale/latest).
   You can manually resync the provider when you are logged in to the registry.

## Developing the Documentation Website

Use the Terraform [doc preview tool](https://registry.terraform.io/tools/doc-preview) to test markdown rendering.

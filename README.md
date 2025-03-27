# Configuring Terraform Provider Overrides Locally

To set up and test a Terraform provider locally, you can configure a `.terraformrc` file in your home directory (`~`). Follow the steps below to configure and validate the setup.

## Steps to Configure

1. **Create the `.terraformrc` File**

   Create a new file named `.terraformrc` in your home directory:

   ```bash
   touch ~/.terraformrc
   ```

2. **Add the `dev_overrides` Block**

   Add the following block to the `.terraformrc` file. Replace `<PATH>` with the value returned by the `go env GOBIN` command:

   ```
   provider_installation {
       dev_overrides {
           "registry.terraform.io/terraform-provider-vnpaycloud/vnpaycloud" = "/path/to/go/bin"
       }
   }
   ```

   To find the correct path, run:

   ```bash
   go env GOBIN
   ```

   Use the returned value in place of `<PATH>`.

## Verifying the Configuration

Run a Terraform plan with a non-existent data source to verify the configuration. For example, using the following Terraform code in `main.tf`:

```hcl
provider "vnpaycloud" {}

data "vnpaycloud_volume" "example" {}
```

Run the command:

```bash
terraform plan
```

## Expected Output

You should see the following output, indicating that the development override is in effect and that the data source is invalid:

```
╷
│ Warning: Provider development overrides are in effect
│
│ The following provider development overrides are set in the CLI
│ configuration:
│  - registry.terraform.io/terraform-provider-vnpaycloud/vnpaycloud in /path/to/go/bin
│
│ The behavior may therefore not match any released version of the provider and
│ applying changes may cause the state to become incompatible with published
│ releases.
╵
╷
│ Error: Invalid data source
│
│   on main.tf line 2, in data "vnpaycloud_volume" "example":
│   2: data "vnpaycloud_volume" "example" {}
│
│ The provider registry.terraform.io/terraform-provider-vnpaycloud does not support data source
│ "vnpaycloud_volume".
╵
```

This output confirms that the local provider override is correctly configured but that the data source is not implemented in the provider.


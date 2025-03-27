terraform {
  required_providers {
    vnpaycloud = {
      source = "terraform-provider-vnpaycloud/vnpaycloud"
      version = "1.0.0"
    }
  }
}

provider "vnpaycloud" {
  auth_url = "https://compute-hcm.infiniband.vn/v3"
  application_credential_id = "678794411ee745f1b950e12e91dd80ce"
  application_credential_name = "tf_dev_for_code"
  application_credential_secret = "HCM01-JPt_j46fmaEN_Wja2HwkG3rMBS6CAF4-mScdDchgfXhaVYJtErCdyM1WYXyZxfu5CeeCvtc2DJMT1z2Y"
}

resource "vnpaycloud_volume" "volume_created_by_tf_thuannt" {
  name = "volume_created_by_tf_thuannt"
  description = "volume_created_by_tf_thuannt"
  size = 30
  volume_type = "c1-standard"
}
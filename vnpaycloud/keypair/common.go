package keypair

import (
	"github.com/vnpaycloud-console/gophercloud/v2/openstack/compute/v2/keypairs"
	"terraform-provider-vnpaycloud/vnpaycloud/util"
)

const (
	computeKeyPairV2UserIDMicroversion = "2.10"
)

// ComputeKeyPairV2CreateOpts is a custom KeyPair struct to include the ValueSpecs field.
type ComputeKeyPairV2CreateOpts struct {
	keypairs.CreateOpts
	ValueSpecs map[string]string `json:"value_specs,omitempty"`
}

// ToKeyPairCreateMap casts a CreateOpts struct to a map.
// It overrides keypairs.ToKeyPairCreateMap to add the ValueSpecs field.
func (opts ComputeKeyPairV2CreateOpts) ToKeyPairCreateMap() (map[string]interface{}, error) {
	return util.BuildRequest(opts, "keypair")
}

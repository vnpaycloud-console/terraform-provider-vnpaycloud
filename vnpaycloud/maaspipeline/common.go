package maaspipeline

import (
	"regexp"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

// The backend accepts letters, digits, spaces and the three separators for both name and description.
var nameRegex = regexp.MustCompile(`^[a-zA-Z0-9-_. ]*$`)

var nameValidation = validation.All(
	validation.StringLenBetween(3, 255),
	validation.StringMatch(nameRegex, "must contain only letters, digits, spaces, hyphens, underscores and dots"),
)

var descriptionValidation = validation.All(
	validation.StringLenBetween(0, 255),
	validation.StringMatch(nameRegex, "must contain only letters, digits, spaces, hyphens, underscores and dots"),
)

func pipelineComputedSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"id":          {Type: schema.TypeString, Computed: true},
		"name":        {Type: schema.TypeString, Computed: true},
		"type":        {Type: schema.TypeString, Computed: true},
		"description": {Type: schema.TypeString, Computed: true},
		"status":      {Type: schema.TypeString, Computed: true},
		"created_at":  {Type: schema.TypeString, Computed: true},
	}
}

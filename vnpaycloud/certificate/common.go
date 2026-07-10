package certificate

import "regexp"

var nameRegexp = regexp.MustCompile(`^[a-zA-Z0-9-_. ]{3,255}$`)

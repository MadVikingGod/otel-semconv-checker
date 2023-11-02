package servers

import (
	"regexp"
)

type matchDef struct {
	name   *regexp.Regexp
	group  []string
	ignore []string
}

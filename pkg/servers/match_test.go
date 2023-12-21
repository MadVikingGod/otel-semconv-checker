package servers

import "regexp"

func newTestMatchDef(groups []string, ignore []string) matchDef {
	return matchDef{
		name:   regexp.MustCompile(`.*`),
		group:  groups,
		ignore: ignore,
	}
}

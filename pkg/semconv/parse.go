package semconv

import (
	"embed"
	"fmt"
	"io/fs"
	"strings"

	"gopkg.in/yaml.v3"
)

//go:embed src/*
var files embed.FS

type File struct {
	Groups []Group
}

func fileError(path string, err error) error {
	return fmt.Errorf("error parsing %s: %w", path, err.Error())
}

func ParseGroups() (map[string]Group, error) {
	groups := make(map[string]Group)
	err := fs.WalkDir(files, "src", func(path string, d fs.DirEntry, err error) error {
		if !strings.HasSuffix(d.Name(), ".yaml") {
			return nil
		}
		var raw File
		b, err := files.ReadFile(path)
		if err != nil {
			return fileError(path, err)
		}
		if err := yaml.Unmarshal(b, &raw); err != nil {
			return fileError(path, err)
		}
		for _, g := range raw.Groups {
			if _, ok := groups[g.Id]; ok {
				return fmt.Errorf("duplicate group id %s", g.Id)
			}
			groups[g.Id] = g
		}
		return nil
	})
	return flattenGroup(groups), err
}

func flattenGroup(groups map[string]Group) map[string]Group {
	attributes := map[string]Attribute{}
	for _, g := range groups {
		for _, a := range g.Attributes {
			if a.Id == "" {
				continue
			}

			a.CanonicalId = canonicalName(g.Prefix, a.Id)
			attributes[a.CanonicalId] = a
		}
	}

	for id, g := range groups {
		for g.Extends != "" {
			prefix := groups[g.Extends].Prefix
			for _, a := range groups[g.Extends].Attributes {
				if a.Id == "" {
					g.Attributes = append(g.Attributes, attributes[a.Ref])
					continue
				}
				if a.CanonicalId != "" {
					g.Attributes = append(g.Attributes, a)
					continue
				}
				a.CanonicalId = canonicalName(prefix, a.Id)
				g.Attributes = append(g.Attributes, a)
			}
			g.Extends = groups[g.Extends].Extends
		}
		for i, a := range g.Attributes {
			if a.Id == "" {
				g.Attributes = append(g.Attributes, attributes[a.Ref])
				continue
			}
			if a.CanonicalId != "" {
				continue
			}
			a.CanonicalId = canonicalName(g.Prefix, a.Id)
			g.Attributes[i] = a
		}
		groups[id] = g
	}

	return groups
}
func canonicalName(prefix, name string) string {
	if prefix != "" {
		return fmt.Sprintf("%s.%s", prefix, name)
	}
	return name
}

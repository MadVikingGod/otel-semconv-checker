package semconv

import (
	"embed"
	"fmt"
	"io/fs"
	"path"
	"strings"

	"gopkg.in/yaml.v3"
)

const DefaultVersion = "https://opentelemetry.io/schemas/1.21.0"

//go:embed src/*
var files embed.FS

type versions struct {
	Versions []SemanticVersion `yaml:"versions"`
}

type SemanticVersion struct {
	Url    string `yaml:"url"`
	Dir    string `yaml:"dir"`
	Groups map[string]Group
}

type File struct {
	Groups []Group
}

func fileError(path string, err error) error {
	return fmt.Errorf("error parsing %s: %w", path, err)
}

func ParseSemanticVersion() (map[string]SemanticVersion, error) {
	var v versions
	b, err := files.ReadFile("src/versions.yaml")
	if err != nil {
		return nil, err
	}
	if err := yaml.Unmarshal(b, &v); err != nil {
		return nil, err
	}
	versions := make(map[string]SemanticVersion)
	for _, v := range v.Versions {
		groups, err := ParseGroups(path.Join("src", v.Dir))
		if err != nil {
			return nil, err
		}
		v.Groups = groups
		versions[v.Url] = v
	}
	return versions, nil
}

func ParseGroups(dir string) (map[string]Group, error) {
	groups := make(map[string]Group)
	err := fs.WalkDir(files, dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
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
	return denormalizeGroups(groups), err
}

func denormalizeGroups(groups map[string]Group) map[string]Group {
	attributes := map[string]Attribute{}

	// For all non-reference attributes:
	// - Make their CanonicalId (prefix+id)
	// - Create a global lookup for that attribute
	for _, g := range groups {
		for _, a := range g.Attributes {
			if a.Id == "" {
				continue
			}

			a.CanonicalId = canonicalName(g.Prefix, a.Id)
			attributes[a.CanonicalId] = a
		}
	}

	for _, g := range groups {
		for i, a := range g.Attributes {
			if a.Ref == "" {
				continue
			}
			g.Attributes[i] = attributes[a.Ref]
		}
	}

	for id, g := range groups {
		for g.Extends != "" {
			g.Attributes = append(g.Attributes, groups[g.Extends].Attributes...)
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

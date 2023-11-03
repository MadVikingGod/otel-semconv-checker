package semconv

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseGroups(t *testing.T) {
	tests := []string{
		"src/v1.20.0",
		"src/v1.21.0",
		"src/v1.22.0",
	}

	for _, dir := range tests {
		t.Run(dir, func(t *testing.T) {
			testParseGroups(t, dir)
		})
	}

}

func testParseGroups(t *testing.T, dir string) {

	groups, err := ParseGroups(dir)

	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(groups), 1)

	// All groups have names
	for gName, g := range groups {
		assert.NotEmpty(t, gName)

		// All Attributes have been denormalized
		// All Attributes have CanonicalID (prefix.id)
		// No Attribute is a Ref to another attribute.
		for _, attr := range g.Attributes {
			assert.NotEmpty(t, attr.CanonicalId)
			assert.Empty(t, attr.Ref)
		}
	}
}

func TestParseSemanticVersion(t *testing.T) {
	versions, err := ParseSemanticVersion()

	assert.NoError(t, err)
	assert.Len(t, versions, 3)
}

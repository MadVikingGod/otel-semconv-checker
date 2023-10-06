package semconv

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseGroups(t *testing.T) {
	groups, err := ParseGroups()

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

package build

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetVersion(t *testing.T) {
	result := "v" + version + "-" + Commit

	assert.Equal(t, GetVersion(), result, "Version is not formatted correctly")
}

package disk

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetMounts(t *testing.T) {
	mounts, err := getMounts(context.Background(), ignoredMounts)
	require.NoError(t, err)
	require.NotEmpty(t, mounts)

	var hasRoot bool

	for _, m := range mounts {
		if m.Mountpoint == "/" {
			hasRoot = true
			break
		}
	}

	assert.True(t, hasRoot, "expected a mount at /")
}

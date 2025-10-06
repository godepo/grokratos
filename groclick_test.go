package grokratos

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	tc := suite.Case(t)
	require.NotNil(t, tc)
	require.NotNil(t, tc.Deps.Admin)
	require.NotNil(t, tc.Deps.Front)
}

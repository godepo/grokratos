package grokratos

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	tckratos "github.com/godepo/grokratos/pkg/tc-kratos"
)

func TestBootstrapper(t *testing.T) {
	t.Run("should be able can't run", func(t *testing.T) {
		t.Run("when is not specified config path", func(t *testing.T) {
			exp := errors.New("unexpected error")
			var cfg config
			cfg.runner = func(ctx context.Context, opts ...tckratos.Option) (KratosContainer, error) {
				return nil, exp
			}
			_, err := bootstrapper[Deps](cfg)(t.Context())
			require.ErrorIs(t, err, exp)

		})
	})
}

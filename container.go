package grokratos

import (
	"context"
	"sync/atomic"
	"testing"

	"github.com/godepo/groat/pkg/generics"
	client "github.com/ory/kratos-client-go"
)

func newContainer[T any](
	ctx context.Context,
	cfg config,
	adminDSN, publicDSN string,
) *Container[T] {
	container := &Container[T]{
		forks:            &atomic.Int32{},
		adminDSN:         adminDSN,
		publicDSN:        publicDSN,
		ctx:              ctx,
		injectLabel:      cfg.injectLabel,
		frontInjectLabel: cfg.frontInjectLabel,
	}

	return container
}

func (c *Container[T]) Injector(t *testing.T, to T) T {
	t.Helper()

	cfg := client.NewConfiguration()
	cfg.Host = c.adminDSN
	cfg.Scheme = "http"

	adminClient := client.NewAPIClient(cfg)

	res := generics.Injector(t, adminClient, to, c.injectLabel)

	cfgFront := client.NewConfiguration()
	cfgFront.Host = c.publicDSN
	cfgFront.Scheme = "http"

	frontClient := client.NewAPIClient(cfg)

	res = generics.Injector(t, frontClient, res, c.frontInjectLabel)

	return res
}

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
	click KratosContainer,
	cfg config,
) *Container[T] {
	container := &Container[T]{
		forks:            &atomic.Int32{},
		kratosContainer:  click,
		ctx:              ctx,
		injectLabel:      cfg.injectLabel,
		frontInjectLabel: cfg.frontInjectLabel,
	}

	return container
}

func (c *Container[T]) Injector(t *testing.T, to T) T {
	t.Helper()

	cfg := client.NewConfiguration()
	cfg.Host = c.kratosContainer.AdminConnectionString(c.ctx)

	adminClient := client.NewAPIClient(cfg)

	res := generics.Injector(t, adminClient, to, c.injectLabel)

	cfgFront := client.NewConfiguration()
	cfgFront.Host = c.kratosContainer.PublicConnectionString(c.ctx)

	frontClient := client.NewAPIClient(cfg)

	res = generics.Injector(t, frontClient, res, c.frontInjectLabel)

	return res
}

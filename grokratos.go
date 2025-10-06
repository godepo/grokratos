package grokratos

import (
	"context"
	"fmt"
	"os"
	"sync/atomic"

	"github.com/godepo/groat/integration"
	"github.com/godepo/groat/pkg/ctxgroup"
	"github.com/testcontainers/testcontainers-go"

	"github.com/godepo/grokratos/internal/containersync"
	tckratos "github.com/godepo/grokratos/pkg/tc-kratos"
)

type (
	KratosContainer interface {
		PublicConnectionString(ctx context.Context) string
		AdminConnectionString(ctx context.Context) string
		Terminate(ctx context.Context, opts ...testcontainers.TerminateOption) error
	}

	containerRunner func(
		ctx context.Context,
		opts ...tckratos.Option,
	) (KratosContainer, error)

	Container[T any] struct {
		forks            *atomic.Int32
		kratosContainer  KratosContainer
		ctx              context.Context
		injectLabel      string
		frontInjectLabel string
	}
	config struct {
		containerImage   string
		imageEnvValue    string
		injectLabel      string
		frontInjectLabel string
		runner           containerRunner
		userSchemaPath   string
		kratosConfig     string
	}

	Option func(*config)
)

func WithUserSchemaPath(path string) Option {
	return func(c *config) {
		c.userSchemaPath = path
	}
}

func WithConfig(path string) Option {
	return func(c *config) {
		c.kratosConfig = path
	}
}

func WithContainerImage(image string) Option {
	return func(c *config) {
		c.containerImage = image
	}
}

func WithInjectLabel(label string) Option {
	return func(c *config) {
		c.injectLabel = label
	}
}

func WithFrontInjectLabel(label string) Option {
	return func(c *config) {
		c.frontInjectLabel = label
	}
}

func New[T any](options ...Option) integration.Bootstrap[T] {
	cfg := config{
		containerImage: "oryd/kratos:v1.3.1",
		imageEnvValue:  "GROAT_I9N_KR_IMAGE",

		injectLabel:      "grokratos",
		frontInjectLabel: "grokratos.front",
		runner: func(
			ctx context.Context,
			opts ...tckratos.Option,
		) (KratosContainer, error) {
			return tckratos.Run(ctx, opts...)
		},
	}

	for _, op := range options {
		op(&cfg)
	}

	if env := os.Getenv(cfg.imageEnvValue); env != "" {
		cfg.containerImage = env
	}

	return bootstrapper[T](cfg)
}

func bootstrapper[T any](cfg config) integration.Bootstrap[T] {
	return func(ctx context.Context) (integration.Injector[T], error) {
		kratosContainer, err := cfg.runner(
			ctx,
			tckratos.WithKratosConfig(cfg.kratosConfig),
			tckratos.WithUserSchemaPath(cfg.userSchemaPath),
			tckratos.WithKratosImage(cfg.containerImage),
		)
		if err != nil {
			return nil, fmt.Errorf("kratos container failed to run: %w", err)
		}

		ctxgroup.IncAt(ctx)

		go containersync.Terminator(ctx, kratosContainer.Terminate)()

		container := newContainer[T](ctx, kratosContainer, cfg)

		return container.Injector, nil
	}
}

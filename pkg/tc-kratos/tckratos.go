//go:generate go tool mockery
package tckratos

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

var (
	ErrConfigNotFound     = errors.New("kratos config not found")
	ErrUserSchemaNotFound = errors.New("user schema not found")
)

const readOnlyRights int64 = 0644

type Option func(*KratosConfig)

type KratosContainer struct {
	KratosContainer testcontainers.Container
	PublicURL       string
	AdminURL        string
	DSN             string
}

func (kc *KratosContainer) PublicConnectionString(ctx context.Context) string {
	return kc.PublicURL
}

func (kc *KratosContainer) AdminConnectionString(ctx context.Context) string {
	return kc.AdminURL
}

func (kc *KratosContainer) Terminate(ctx context.Context, opts ...testcontainers.TerminateOption) error {
	err := kc.KratosContainer.Terminate(ctx, opts...)
	if err != nil {
		return fmt.Errorf("failed to terminate kratos container: %w", err)
	}

	return nil
}

type KratosConfig struct {
	userSchemaPath       string
	kratosConfig         string
	containerConstructor func(
		ctx context.Context,
		req testcontainers.GenericContainerRequest,
	) (testcontainers.Container, error)
	kratosImage string
}

func WithUserSchemaPath(path string) func(*KratosConfig) {
	return func(c *KratosConfig) {
		c.userSchemaPath = path
	}
}

func WithKratosImage(image string) func(*KratosConfig) {
	return func(c *KratosConfig) {
		c.kratosImage = image
	}
}

func WithKratosConfig(config string) func(*KratosConfig) {
	return func(c *KratosConfig) {
		c.kratosConfig = config
	}
}

func WithContainerConstructor(
	fn func(ctx context.Context, req testcontainers.GenericContainerRequest,
	) (testcontainers.Container, error)) func(*KratosConfig) {
	return func(c *KratosConfig) {
		c.containerConstructor = fn
	}
}

func Run(ctx context.Context, opts ...Option) (*KratosContainer, error) {
	cfg := KratosConfig{
		kratosConfig:         "",
		userSchemaPath:       "",
		containerConstructor: testcontainers.GenericContainer,
		kratosImage:          "oryd/kratos:v1.3.1",
	}

	for _, fn := range opts {
		fn(&cfg)
	}

	if cfg.kratosConfig == "" {
		return nil, ErrConfigNotFound
	}

	if cfg.userSchemaPath == "" {
		return nil, ErrUserSchemaNotFound
	}

	kratosReq := containerRequest(cfg)

	kratosContainer, err := cfg.containerConstructor(ctx,
		testcontainers.GenericContainerRequest{
			ContainerRequest: kratosReq,
			Started:          true,
		})
	if err != nil {
		return nil, fmt.Errorf("failed to start kratos: %w", err)
	}

	// Получаем порты Kratos
	publicPort, err := kratosContainer.MappedPort(ctx, "4433")
	if err != nil {
		return nil, fmt.Errorf("failed to get kratos public port: %w", err)
	}

	adminPort, err := kratosContainer.MappedPort(ctx, "4434")
	if err != nil {
		return nil, fmt.Errorf("failed to get kratos admin port: %w", err)
	}

	kratosHost, err := kratosContainer.Host(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get kratos host: %w", err)
	}

	publicURL := net.JoinHostPort(kratosHost, publicPort.Port())
	adminURL := net.JoinHostPort(kratosHost, adminPort.Port())

	return &KratosContainer{
		KratosContainer: kratosContainer,
		PublicURL:       publicURL,
		AdminURL:        adminURL,
	}, nil
}

func containerRequest(cfg KratosConfig) testcontainers.ContainerRequest {
	return testcontainers.ContainerRequest{
		Image:        cfg.kratosImage,
		ExposedPorts: []string{"4433/tcp", "4434/tcp"},
		Cmd:          []string{"serve", "-c", "/etc/config/kratos/kratos.yaml", "--dev"},
		Env: map[string]string{
			"LOG_LEVEL":             "trace",
			"LOG_FORMAT":            "text",
			"DSN":                   "memory",
			"SERVE_PUBLIC_BASE_URL": "http://127.0.0.1:4433/",
			"SERVE_ADMIN_BASE_URL":  "http://127.0.0.1:4434/",
		},
		Files: []testcontainers.ContainerFile{
			{
				HostFilePath:      cfg.kratosConfig,
				ContainerFilePath: "/etc/config/kratos/kratos.yaml",
				FileMode:          readOnlyRights,
			},
			{
				HostFilePath:      cfg.userSchemaPath,
				ContainerFilePath: "/etc/config/kratos/user.schema.json",
				FileMode:          readOnlyRights,
			},
		},

		WaitingFor: wait.ForHTTP("/health/ready").
			WithPort("4433/tcp").
			WithStartupTimeout(time.Minute).
			WithStatusCodeMatcher(func(status int) bool {
				return status == http.StatusOK
			}),
	}
}

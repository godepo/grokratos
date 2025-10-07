//go:generate go tool mockery
package tckratos

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/go-connections/nat"
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
	kratosImage              string
	adminPort                int
	frontPort                int
	adminListenerConstructor func(network string, address string) (net.Listener, error)
	frontListenerConstructor func(network string, address string) (net.Listener, error)
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

func WithAdminListenerConstructor(
	fn func(network string, address string) (net.Listener, error)) func(*KratosConfig) {
	return func(c *KratosConfig) {
		c.adminListenerConstructor = fn
	}
}

func WithFrontListenerConstructor(
	fn func(network string, address string) (net.Listener, error)) func(*KratosConfig) {
	return func(c *KratosConfig) {
		c.frontListenerConstructor = fn
	}
}

func Run(ctx context.Context, opts ...Option) (*KratosContainer, error) {
	cfg := KratosConfig{
		kratosConfig:             "",
		userSchemaPath:           "",
		containerConstructor:     testcontainers.GenericContainer,
		kratosImage:              "oryd/kratos:v1.3.1",
		adminListenerConstructor: net.Listen,
		frontListenerConstructor: net.Listen,
	}

	for _, fn := range opts {
		fn(&cfg)
	}

	if cfg.kratosConfig == "" {
		return nil, ErrConfigNotFound
	}

	adminLn, err := cfg.adminListenerConstructor("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, fmt.Errorf("failed to listen on admin port: %w", err)
	}

	cfg.adminPort = adminLn.Addr().(*net.TCPAddr).Port

	frontLn, err := cfg.frontListenerConstructor("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, fmt.Errorf("failed to listen on admin port: %w", err)
	}

	cfg.frontPort = frontLn.Addr().(*net.TCPAddr).Port

	if cfg.userSchemaPath == "" {
		return nil, ErrUserSchemaNotFound
	}

	_ = frontLn.Close()
	_ = adminLn.Close()

	kratosReq := containerRequest(cfg)

	kratosContainer, err := cfg.containerConstructor(ctx,
		testcontainers.GenericContainerRequest{
			ContainerRequest: kratosReq,
			Started:          true,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to start kratos: %w", err)
	}

	kratosHost := "localhost"

	publicURL := net.JoinHostPort(kratosHost, strconv.Itoa(cfg.frontPort))
	adminURL := net.JoinHostPort(kratosHost, strconv.Itoa(cfg.adminPort))

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
			"SERVE_PUBLIC_BASE_URL": "http://localhost:" + strconv.Itoa(cfg.frontPort) + "/",
			"SERVE_ADMIN_BASE_URL":  "http://localhost:" + strconv.Itoa(cfg.adminPort) + "/",
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
		HostConfigModifier: func(hc *container.HostConfig) {
			adminPort, _ := nat.NewPort("tcp", "4434")
			hc.PortBindings = nat.PortMap{
				"4433/tcp": []nat.PortBinding{
					{
						HostIP:   "127.0.0.1",
						HostPort: strconv.Itoa(cfg.frontPort),
					},
				},
				adminPort: []nat.PortBinding{
					{
						HostIP:   "127.0.0.1",
						HostPort: strconv.Itoa(cfg.adminPort),
					},
				},
			}
		},
		WaitingFor: wait.ForHTTP("/health/ready").
			WithPort("4433/tcp").
			WithStartupTimeout(time.Minute).
			WithStatusCodeMatcher(func(status int) bool {
				return status == http.StatusOK
			}),
	}
}

package tckratos

import (
	"context"
	"errors"
	"net"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
)

func TestStartKratosWithTestContainers(t *testing.T) {
	t.Run("should be able to be able", func(t *testing.T) {

		container, err := Run(
			t.Context(),
			WithKratosConfig("etc/kratos.yaml"),
			WithUserSchemaPath("etc/user.schema.json"),
			WithKratosImage("oryd/kratos:v1.3.1"),
		)
		require.NoError(t, err)
		require.NotNil(t, container)
		require.NoError(t, container.Terminate(t.Context()))
		assert.NotEmpty(t, container.AdminConnectionString(t.Context()))
		assert.NotEmpty(t, container.PublicConnectionString(t.Context()))
	})
	t.Run("should be able to be failed", func(t *testing.T) {
		t.Run("when is not specified config path", func(t *testing.T) {
			_, err := Run(
				t.Context(),
			)
			require.ErrorIs(t, err, ErrConfigNotFound)
		})

		t.Run("when is not specified user schema path", func(t *testing.T) {
			_, err := Run(
				t.Context(),
				WithKratosConfig("etc/kratos.yaml"),
			)
			require.ErrorIs(t, err, ErrUserSchemaNotFound)
		})

		t.Run("when cant allocate admin port", func(t *testing.T) {
			expErr := errors.New(uuid.NewString())
			_, err := Run(
				t.Context(),
				WithKratosConfig("etc/kratos.yaml"),
				WithUserSchemaPath("etc/user.schema.json"),
				WithAdminListenerConstructor(func(network string, address string) (net.Listener, error) {
					return nil, expErr
				}),
			)
			require.ErrorIs(t, err, expErr)
		})

		t.Run("when cant allocate front port", func(t *testing.T) {
			expErr := errors.New(uuid.NewString())
			_, err := Run(
				t.Context(),
				WithKratosConfig("etc/kratos.yaml"),
				WithUserSchemaPath("etc/user.schema.json"),
				WithFrontListenerConstructor(func(network string, address string) (net.Listener, error) {
					return nil, expErr
				}),
			)
			require.ErrorIs(t, err, expErr)
		})

		t.Run("when run container will be failed", func(t *testing.T) {
			expErr := errors.New(uuid.NewString())
			_, err := Run(
				t.Context(),
				WithKratosConfig("etc/kratos.yaml"),
				WithUserSchemaPath("etc/user.schema.json"),
				WithContainerConstructor(
					func(
						ctx context.Context,

						req testcontainers.GenericContainerRequest,
					) (testcontainers.Container, error) {

						return nil, expErr
					}),
			)
			require.ErrorIs(t, err, expErr)
		})
	})
}

func TestKratosContainer_Terminate(t *testing.T) {
	expErr := errors.New(uuid.NewString())
	mock := NewMockContainer(t)

	mock.EXPECT().Terminate(t.Context()).Return(expErr)

	container := &KratosContainer{
		KratosContainer: mock,
	}
	err := container.Terminate(t.Context())
	require.ErrorIs(t, err, expErr)
}

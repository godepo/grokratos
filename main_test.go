package grokratos

import (
	"os"
	"testing"

	"github.com/godepo/groat"
	"github.com/godepo/groat/integration"
	client "github.com/ory/kratos-client-go"
)

type (
	SystemUnderTest struct {
	}

	State struct {
	}
	Deps struct {
		Front *client.APIClient `groat:"grokratos"`
		Admin *client.APIClient `groat:"grokratos.front"`
	}
)

var suite *integration.Container[Deps, State, *SystemUnderTest]

func TestMain(m *testing.M) {
	_ = os.Setenv("GROAT_I9N_KR_IMAGE", "oryd/kratos:v1.3.1")

	suite = integration.New[Deps, State, *SystemUnderTest](
		m,
		func(t *testing.T) *groat.Case[Deps, State, *SystemUnderTest] {
			tcs := groat.New[Deps, State, *SystemUnderTest](t, func(t *testing.T, deps Deps) *SystemUnderTest {
				return &SystemUnderTest{}
			})
			return tcs
		},
		New[Deps](
			WithInjectLabel("grokratos"),
			WithContainerImage("oryd/kratos:v1.3.1"),
			WithFrontInjectLabel("grokratos.front"),
			WithUserSchemaPath("pkg/tc-kratos/etc/user.schema.json"),
			WithConfig("pkg/tc-kratos/etc/kratos.yaml"),
		),
	)
	os.Exit(suite.Go())
}

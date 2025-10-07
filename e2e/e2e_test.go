package e2e

import (
	"os"
	"testing"

	"github.com/godepo/groat"
	"github.com/godepo/groat/integration"
	"github.com/jaswdr/faker/v2"
	client "github.com/ory/kratos-client-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/godepo/grokratos"
)

type (
	Deps struct {
		Client *client.APIClient `groat:"grokratos"`
		Front  *client.APIClient `groat:"grokratos.front"`
		Faker  faker.Faker
	}
	State struct {
	}
	Service struct {
		client *client.APIClient
	}
)

var suite *integration.Container[Deps, State, *Service]

func TestMain(m *testing.M) {
	suite = integration.New[Deps, State, *Service](
		m,
		func(t *testing.T) *groat.Case[Deps, State, *Service] {
			return groat.New[Deps, State, *Service](t, func(t *testing.T, deps Deps) *Service {
				return &Service{
					client: deps.Client,
				}
			})
		},
		grokratos.New[Deps](
			grokratos.WithInjectLabel("grokratos"),
			grokratos.WithContainerImage("oryd/kratos:v1.3.1"),
			grokratos.WithFrontInjectLabel("grokratos.front"),
			grokratos.WithUserSchemaPath("../pkg/tc-kratos/etc/user.schema.json"),
			grokratos.WithConfig("../pkg/tc-kratos/etc/kratos.yaml"),
		),
	)
	os.Exit(suite.Go())
}

func TestCreateUser(t *testing.T) {
	t.Run("should be able to create user", func(t *testing.T) {
		tc := suite.Case(t)

		fkr := faker.New()

		credentials := client.IdentityWithCredentials{
			Password: client.NewIdentityWithCredentialsPassword(),
		}

		credentials.Password = &client.IdentityWithCredentialsPassword{
			Config: client.NewIdentityWithCredentialsPasswordConfig(),
		}

		credentials.Password.Config.Password = client.PtrString(fkr.Internet().Password())

		login := fkr.Internet().Email()
		traits := map[string]interface{}{
			"email": login,
		}

		createBody := client.CreateIdentityBody{
			SchemaId:    "user",
			Traits:      traits,
			Credentials: &credentials,
		}

		id, _, err := tc.SUT.client.IdentityAPI.
			CreateIdentity(t.Context()).
			CreateIdentityBody(createBody).
			Execute()
		require.NoError(t, err)
		require.NotEmpty(t, id)

		flow, resp, err := tc.Deps.Front.FrontendAPI.CreateNativeLoginFlow(t.Context()).Execute()
		require.NoError(t, err)

		defer func() {
			_ = resp.Body.Close()
		}()

		updateBody := client.UpdateLoginFlowBody{
			UpdateLoginFlowWithPasswordMethod: &client.UpdateLoginFlowWithPasswordMethod{
				Method:     "password",
				Identifier: login,
				Password:   *credentials.Password.Config.Password,
			},
		}

		loginResult, resp, err := tc.SUT.client.FrontendAPI.
			UpdateLoginFlow(t.Context()).Flow(flow.Id).
			UpdateLoginFlowBody(updateBody).
			Execute()
		require.NoError(t, err)

		defer func() {
			_ = resp.Body.Close()
		}()

		sessionToken := loginResult.GetSessionToken()
		require.NotEmpty(t, sessionToken)

		assert.Equal(t, id.Id, loginResult.GetSession().Identity.Id)
	})
}

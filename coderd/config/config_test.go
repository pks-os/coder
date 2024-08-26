package config_test

import (
	"testing"

	"github.com/coder/serpent"
	"github.com/stretchr/testify/require"

	"github.com/coder/coder/v2/coderd/coderdtest"
	"github.com/coder/coder/v2/coderd/config"
	"github.com/coder/coder/v2/codersdk"
	"github.com/coder/coder/v2/enterprise/coderd/coderdenttest"
	"github.com/coder/coder/v2/enterprise/coderd/license"
)

// TestConfig demonstrates creating org-level overrides for deployment-level settings.
func TestConfig(t *testing.T) {
	t.Parallel()

	vals := coderdtest.DeploymentValues(t)
	vals.Experiments = []string{string(codersdk.ExperimentMultiOrganization)}
	adminClient, _, _, _ := coderdenttest.NewWithAPI(t, &coderdenttest.Options{
		Options: &coderdtest.Options{DeploymentValues: vals},
		LicenseOptions: &coderdenttest.LicenseOptions{
			Features: license.Features{
				codersdk.FeatureMultipleOrganizations: 1,
			},
		},
	})
	altOrg := coderdenttest.CreateOrganization(t, adminClient, coderdenttest.CreateOrganizationOptions{})

	t.Run("string", func(t *testing.T) {
		t.Parallel()

		store := config.NewFakeStore()

		var (
			base     = serpent.String("system@dev.coder.com")
			override = serpent.String("dogfood@dev.coder.com")
		)

		field := &vals.Notifications.SMTP.From
		// Check that no default has been set.
		require.Empty(t, field.GlobalValue().String())
		// Initialize the value, it has no default.
		require.NoError(t, field.Set(base.String()))
		// Validate that it returns that value.
		require.Equal(t, base.String(), field.String())
		// Validate that there is no org-level override right now.
		_, err := field.OrgValue(store, altOrg.ID)
		require.ErrorIs(t, err, config.EntryNotFound)
		// Coalesce returns the deployment-wide value.
		val, err := field.Coalesce(store, altOrg.ID)
		require.NoError(t, err)
		require.Equal(t, base.String(), val.String())
		// Set an org-level override.
		require.NoError(t, field.Override(store, altOrg.ID, &override))
		// Coalesce now returns the org-level value.
		val, err = field.Coalesce(store, altOrg.ID)
		require.NoError(t, err)
		require.Equal(t, override.String(), val.String())
	})

	t.Run("struct", func(t *testing.T) {
		t.Parallel()

		store := config.NewFakeStore()

		field := &vals.OIDC.AuthURLParams
		var (
			base     = map[string]string{"access_type": "offline"}
			override = serpent.Struct[map[string]string]{
				Value: map[string]string{
					"a": "b",
					"c": "d",
				},
			}
		)
		// Validate that the default value was set (see codersdk/deployment.go).
		require.Equal(t, base, field.GlobalValue().Value)
		// Validate that there is no org-level override right now.
		_, err := field.OrgValue(store, altOrg.ID)
		require.ErrorIs(t, err, config.EntryNotFound)
		// Coalesce returns the deployment-wide value.
		val, err := field.Coalesce(store, altOrg.ID)
		require.NoError(t, err)
		require.Equal(t, base, val.Value)
		// Set an org-level override.
		require.NoError(t, field.Override(store, altOrg.ID, &override))
		// Coalesce now returns the org-level value.
		structVal, err := field.OrgValue(store, altOrg.ID)
		require.NoError(t, err)
		require.Equal(t, override.Value, structVal.Value)
	})
}

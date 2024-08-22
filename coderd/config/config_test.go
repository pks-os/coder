package config_test

import (
	"testing"
	"time"

	"github.com/coder/serpent"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/coder/coder/v2/coderd/coderdtest"
	"github.com/coder/coder/v2/coderd/config"
	"github.com/coder/coder/v2/codersdk"
	"github.com/coder/coder/v2/enterprise/coderd/coderdenttest"
	"github.com/coder/coder/v2/enterprise/coderd/license"
)

// TestResolveByField demonstrates creating org-level overrides for deployment-level settings.
func TestResolveByField(t *testing.T) {
	t.Parallel()

	vals := coderdtest.DeploymentValues(t)
	vals.Experiments = []string{string(codersdk.ExperimentMultiOrganization)}
	adminClient, _, _, _ := coderdenttest.NewWithAPI(t, &coderdenttest.Options{
		Options: &coderdtest.Options{DeploymentValues: vals},
		LicenseOptions: &coderdenttest.LicenseOptions{
			Features: license.Features{
				codersdk.FeatureExternalProvisionerDaemons: 1,
				codersdk.FeatureMultipleOrganizations:      1,
			},
		},
	})
	altOrg := coderdenttest.CreateOrganization(t, adminClient, coderdenttest.CreateOrganizationOptions{
		IncludeProvisionerDaemon: true,
	})

	tests := []struct {
		name          string
		field         config.Option
		initialValue  string
		overrideValue string
		expectedValue any
		equalityFn    func(t testing.TB, expected any, actual config.Option) bool
	}{
		// The following test-cases are not meant to exhaustively check all types which satisfy
		// the config.Option interface, but rather to just be demonstrative.
		{
			name:          "default value (string)",
			field:         &vals.Notifications.SMTP.From,
			initialValue:  "system@dev.coder.com",
			expectedValue: "system@dev.coder.com",
			equalityFn: func(t testing.TB, expected any, actual config.Option) bool {
				return assert.Equal(t, expected, actual.(*serpent.String).Value())
			},
		},
		{
			name:          "overridden value (string)",
			field:         &vals.Notifications.SMTP.From,
			initialValue:  "system@dev.coder.com",
			overrideValue: "dogfood@dev.coder.com",
			expectedValue: "dogfood@dev.coder.com",
			equalityFn: func(t testing.TB, expected any, actual config.Option) bool {
				return assert.Equal(t, expected, actual.(*serpent.String).Value())
			},
		},
		{
			name:          "default value (hostport)",
			field:         &vals.Notifications.SMTP.Smarthost,
			initialValue:  "localhost:587",
			expectedValue: "localhost:587",
			equalityFn: func(t testing.TB, expected any, actual config.Option) bool {
				return assert.Equal(t, expected, actual.(*serpent.HostPort).String())
			},
		},
		{
			name:          "overridden value (hostport)",
			field:         &vals.Notifications.SMTP.Smarthost,
			initialValue:  "localhost:587",
			overrideValue: "localhost:25",
			expectedValue: "localhost:25",
			equalityFn: func(t testing.TB, expected any, actual config.Option) bool {
				return assert.Equal(t, expected, actual.(*serpent.HostPort).String())
			},
		},
		{
			name:          "default value (int64)",
			field:         &vals.Provisioner.Daemons,
			initialValue:  "3",
			expectedValue: int64(3),
			equalityFn: func(t testing.TB, expected any, actual config.Option) bool {
				return assert.Equal(t, expected, actual.(*serpent.Int64).Value())
			},
		},
		{
			name:          "overridden value (int64)",
			field:         &vals.Provisioner.Daemons,
			initialValue:  "3",
			overrideValue: "0",
			expectedValue: int64(0),
			equalityFn: func(t testing.TB, expected any, actual config.Option) bool {
				return assert.Equal(t, expected, actual.(*serpent.Int64).Value())
			},
		},
		{
			name:          "default value (string array)",
			field:         &vals.ProxyTrustedHeaders,
			initialValue: "Origin",
			expectedValue: []string{"Origin"},
			equalityFn: func(t testing.TB, expected any, actual config.Option) bool {
				return assert.Equal(t, expected, actual.(*serpent.StringArray).Value())
			},
		},
		{
			name:          "overridden value (string array)",
			field:         &vals.ProxyTrustedHeaders,
			initialValue: "Origin",
			overrideValue: "Origin,Content-Type",
			expectedValue: []string{"Origin", "Content-Type"},
			equalityFn: func(t testing.TB, expected any, actual config.Option) bool {
				return assert.Equal(t, expected, actual.(*serpent.StringArray).Value())
			},
		},
	}

	for _, tc := range tests {
		// Tests are _not_ run in parallel because some refer to the same field.
		t.Run(tc.name, func(t *testing.T) {
			mgr := config.NewManager(vals.Options())

			// Set the initial value in the field.
			require.NoError(t, tc.field.Set(tc.initialValue))

			if tc.overrideValue != "" {
				// Set an overridden value in the store.
				require.NoError(t, mgr.AddOrgSettingOverride(altOrg.ID, tc.field, tc.overrideValue))
			}

			out, err := config.ResolveForOrgByOption(mgr, altOrg.ID, tc.field)
			require.NoError(t, err)
			// assert.True(t,
			tc.equalityFn(t, tc.expectedValue, out)
			// )
		})
	}
}

// TestResolveByName demonstrates create an org-level setting with any key.
func TestResolveByName(t *testing.T) {
	t.Parallel()

	vals := coderdtest.DeploymentValues(t)
	vals.Experiments = []string{string(codersdk.ExperimentMultiOrganization)}
	adminClient, _, _, _ := coderdenttest.NewWithAPI(t, &coderdenttest.Options{
		Options: &coderdtest.Options{DeploymentValues: vals},
		LicenseOptions: &coderdenttest.LicenseOptions{
			Features: license.Features{
				codersdk.FeatureExternalProvisionerDaemons: 1,
				codersdk.FeatureMultipleOrganizations:      1,
			},
		},
	})
	altOrg := coderdenttest.CreateOrganization(t, adminClient, coderdenttest.CreateOrganizationOptions{
		IncludeProvisionerDaemon: true,
	})

	mgr := config.NewManager(vals.Options())
	const (
		key   = "my-custom-setting"
		value = "my-custom-value"
	)
	require.NoError(t, mgr.AddOrgSettingByName(altOrg.ID, key, value))
	setting, err := config.ResolveForOrgByName(mgr, altOrg.ID, key)
	require.NoError(t, err)
	require.Equal(t, value, setting)
}

// TestResolveByNameUnsafe demonstrates that you _can_ retrieve a setting by a well-known name, but this should be avoided
// because it lacks any type-safety.
func TestResolveByNameUnsafe(t *testing.T) {
	t.Parallel()

	vals := coderdtest.DeploymentValues(t)
	vals.Experiments = []string{string(codersdk.ExperimentMultiOrganization)}
	adminClient, _, _, _ := coderdenttest.NewWithAPI(t, &coderdenttest.Options{
		Options: &coderdtest.Options{DeploymentValues: vals},
		LicenseOptions: &coderdenttest.LicenseOptions{
			Features: license.Features{
				codersdk.FeatureExternalProvisionerDaemons: 1,
				codersdk.FeatureMultipleOrganizations:      1,
			},
		},
	})
	altOrg := coderdenttest.CreateOrganization(t, adminClient, coderdenttest.CreateOrganizationOptions{
		IncludeProvisionerDaemon: true,
	})

	vals.Notifications.FetchInterval = serpent.Duration(time.Minute)

	mgr := config.NewManager(vals.Options())

	// This field, as defined in codersdk/deployment.go, has the env of "CODER_NOTIFICATIONS_FETCH_INTERVAL".
	field := &vals.Notifications.FetchInterval

	require.NoError(t, mgr.AddOrgSettingOverride(altOrg.ID, field, "2m"))

	// DON'T DO THIS!!!
	setting, err := config.ResolveForOrgByName(mgr, altOrg.ID, "CODER_NOTIFICATIONS_FETCH_INTERVAL")
	require.NoError(t, err)
	require.Equal(t, "2m", setting)

	// DO THIS INSTEAD!
	dur, err := config.ResolveForOrgByOption(mgr, altOrg.ID, field)
	require.NoError(t, err)
	require.Equal(t, time.Minute*2, dur.Value())
}

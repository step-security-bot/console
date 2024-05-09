package devices

import (
	"testing"

	"github.com/open-amt-cloud-toolkit/console/internal/entity/dto"
	"github.com/open-amt-cloud-toolkit/go-wsman-messages/v2/pkg/wsman/amt/boot"
	"github.com/open-amt-cloud-toolkit/go-wsman-messages/v2/pkg/wsman/cim/software"
	"github.com/stretchr/testify/require"
)

type powerTest struct {
	name string
	res  any
	err  error

	bootSettings dto.BootSetting
	version      []software.SoftwareIdentity
}

func TestDetermineIDERBootDevice(t *testing.T) {
	t.Parallel()

	tests := []powerTest{
		{
			name: "Master Bus Reset",
			res: boot.BootSettingDataRequest{
				IDERBootDevice: 1,
			},
			bootSettings: dto.BootSetting{
				Action: 202,
			},
		},
		{
			name: "Power On",
			res: boot.BootSettingDataRequest{
				IDERBootDevice: 0,
			},
			bootSettings: dto.BootSetting{
				Action: 999,
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result := boot.BootSettingDataRequest{
				IDERBootDevice: 999,
			}

			determineIDERBootDevice(tc.bootSettings, &result)

			require.Equal(t, tc.res, result)
		})
	}
}

// CRAIG - changeBootOrder - TestChangeBootOrder

func TestDetermineBootAction(t *testing.T) {
	t.Parallel()

	tests := []powerTest{
		{
			name: "Master Bus Reset",
			res:  10,
			bootSettings: dto.BootSetting{
				Action: 200,
			},
		},
		{
			name: "Power On",
			res:  2,
			bootSettings: dto.BootSetting{
				Action: 999,
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			determineBootAction(&tc.bootSettings)

			require.Equal(t, tc.res, tc.bootSettings.Action)
		})
	}
}

func TestParseVersion(t *testing.T) {
	t.Parallel()

	tests := []powerTest{
		{
			name: "success",
			res:  12,
			err:  nil,
			version: []software.SoftwareIdentity{
				{
					InstanceID:    "AMT",
					VersionString: "12.2.67",
				},
			},
		},
		{
			name: "Instance id not AMT",
			res:  0,
			err:  nil,
			version: []software.SoftwareIdentity{
				{
					InstanceID:    "NOT",
					VersionString: "12.2.67",
				},
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			res, err := parseVersion(tc.version)

			require.Equal(t, tc.res, res)
			require.Equal(t, tc.err, err)
		})
	}
}

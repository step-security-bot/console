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

func TestDeterminePowerCapabilities(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		amtVersion       int
		expectedResponse map[string]interface{}
		expectedError    error
	}{
		{
			name:       "AMT version 10",
			amtVersion: 10,
			expectedResponse: map[string]interface{}{
				"Power up":                 2,
				"Power cycle":              5,
				"Power down":               8,
				"Reset":                    10,
				"Soft-off":                 12,
				"Soft-reset":               14,
				"Sleep":                    4,
				"Hibernate":                7,
				"Power on to IDE-R Floppy": 201,
				"Reset to IDE-R CDROM":     202,
				"Power on to IDE-R CDROM":  203,
				"Reset to IDE-R Floppy":    200,
				"Power on to diagnostic":   300,
				"Reset to diagnostic":      301,
				"Reset to PXE":             400,
				"Power on to PXE":          401,
			},
			expectedError: nil,
		},
	}

	for _, tc := range tests {
		tc := tc // capture range variable
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			capabilities := boot.BootCapabilitiesResponse{
				BIOSReflash:         true,
				BIOSSetup:           false,
				SecureErase:         false,
				ForceDiagnosticBoot: true,
			}

			res := determinePowerCapabilities(tc.amtVersion, capabilities)

			require.Equal(t, tc.expectedResponse, res)
			require.Equal(t, tc.expectedError, nil)
		})
	}
}

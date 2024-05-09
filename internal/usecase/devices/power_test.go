package devices_test

import (
	"context"
	"errors"
	"testing"

	"github.com/open-amt-cloud-toolkit/console/internal/entity"
	devices "github.com/open-amt-cloud-toolkit/console/internal/usecase/devices"
	"github.com/open-amt-cloud-toolkit/console/internal/usecase/utils"
	"github.com/open-amt-cloud-toolkit/console/pkg/logger"
	"github.com/open-amt-cloud-toolkit/go-wsman-messages/v2/pkg/wsman/amt/boot"
	"github.com/open-amt-cloud-toolkit/go-wsman-messages/v2/pkg/wsman/cim/power"
	"github.com/open-amt-cloud-toolkit/go-wsman-messages/v2/pkg/wsman/cim/service"
	"github.com/stretchr/testify/require"
	gomock "go.uber.org/mock/gomock"
)

var ErrGeneral = errors.New("general error")

type powerTest struct {
	name     string
	action   int
	manMock  func(*MockManagement)
	repoMock func(*MockRepository)
	res      any
	err      error
}

func initPowerTest(t *testing.T) (*devices.UseCase, *MockManagement, *MockRepository) {
	t.Helper()

	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()

	repo := NewMockRepository(mockCtl)
	management := NewMockManagement(mockCtl)
	log := logger.New("error")
	u := devices.New(repo, management, NewMockRedirection(mockCtl), log)

	return u, management, repo
}

func TestSendPowerAction(t *testing.T) {
	t.Parallel()

	device := &entity.Device{
		GUID:     "device-guid-123",
		TenantID: "tenant-id-456",
	}

	powerActionRes := power.PowerActionResponse{
		ReturnValue: 0,
	}

	tests := []powerTest{
		{
			name:   "success",
			action: 0,
			manMock: func(man *MockManagement) {
				man.EXPECT().
					SetupWsmanClient(gomock.Any(), false, true).
					Return()
				man.EXPECT().
					SendPowerAction(0).
					Return(powerActionRes, nil)
			},
			repoMock: func(repo *MockRepository) {
				repo.EXPECT().
					GetByID(gomock.Any(), device.GUID, "").
					Return(device, nil)
			},
			res: powerActionRes,
			err: nil,
		},
		{
			name:    "GetById fails",
			action:  0,
			manMock: func(man *MockManagement) {},
			repoMock: func(repo *MockRepository) {
				repo.EXPECT().
					GetByID(gomock.Any(), device.GUID, "").
					Return(nil, errTest)
			},
			res: power.PowerActionResponse{},
			err: utils.ErrNotFound,
		},
		{
			name:   "SendPowerAction fails",
			action: 0,
			manMock: func(man *MockManagement) {
				man.EXPECT().
					SetupWsmanClient(gomock.Any(), false, true).
					Return()
				man.EXPECT().
					SendPowerAction(0).
					Return(power.PowerActionResponse{}, ErrGeneral)
			},
			repoMock: func(repo *MockRepository) {
				repo.EXPECT().
					GetByID(gomock.Any(), device.GUID, "").
					Return(device, nil)
			},
			res: power.PowerActionResponse{},
			err: ErrGeneral,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			useCase, management, repo := initPowerTest(t)

			tc.manMock(management)
			tc.repoMock(repo)

			res, err := useCase.SendPowerAction(context.Background(), device.GUID, tc.action)

			require.Equal(t, tc.res, res)
			require.Equal(t, tc.err, err)
		})
	}
}

func TestGetPowerState(t *testing.T) {
	t.Parallel()

	device := &entity.Device{
		GUID:     "device-guid-123",
		TenantID: "tenant-id-456",
	}

	ourRes := []service.CIM_AssociatedPowerManagementService{
		{
			PowerState: 0,
		},
	}

	tests := []powerTest{
		{
			name: "success",
			manMock: func(man *MockManagement) {
				man.EXPECT().
					SetupWsmanClient(gomock.Any(), false, true).
					Return()
				man.EXPECT().
					GetPowerState().
					Return(ourRes, nil)
			},
			repoMock: func(repo *MockRepository) {
				repo.EXPECT().
					GetByID(gomock.Any(), device.GUID, "").
					Return(device, nil)
			},
			res: map[string]interface{}{
				"powerstate": service.PowerState(0),
			},
			err: nil,
		},
		{
			name:    "GetById fails",
			manMock: func(man *MockManagement) {},
			repoMock: func(repo *MockRepository) {
				repo.EXPECT().
					GetByID(gomock.Any(), device.GUID, "").
					Return(nil, errTest)
			},
			res: (map[string]interface{})(nil),
			err: utils.ErrNotFound,
		},
		{
			name: "GetPowerState fails",
			manMock: func(man *MockManagement) {
				man.EXPECT().
					SetupWsmanClient(gomock.Any(), false, true).
					Return()
				man.EXPECT().
					GetPowerState().
					Return([]service.CIM_AssociatedPowerManagementService{}, ErrGeneral)
			},
			repoMock: func(repo *MockRepository) {
				repo.EXPECT().
					GetByID(gomock.Any(), device.GUID, "").
					Return(device, nil)
			},
			res: (map[string]interface{})(nil),
			err: ErrGeneral,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			useCase, management, repo := initPowerTest(t)

			tc.manMock(management)
			tc.repoMock(repo)

			res, err := useCase.GetPowerState(context.Background(), device.GUID)

			require.Equal(t, tc.res, res)
			require.Equal(t, tc.err, err)
		})
	}
}

func TestGetPowerCapabilities(t *testing.T) {
	t.Parallel()
	device := &entity.Device{
		GUID:     "device-guid-123",
		TenantID: "tenant-id-456",
	}
	ourRes := boot.BootCapabilitiesResponse{}
	tests := []powerTest{
		{
			name: "success",
			manMock: func(man *MockManagement) {
				man.EXPECT().
					SetupWsmanClient(gomock.Any(), false, true).
					Return()
				man.EXPECT().
					GetAMTVersion().
					Return(nil, nil)
				man.EXPECT().
					GetPowerCapabilities().
					Return(ourRes, nil)
			},
			repoMock: func(repo *MockRepository) {
				repo.EXPECT().
					GetByID(gomock.Any(), device.GUID, "").
					Return(device, nil)
			},
			res: map[string]interface{}(map[string]interface{}{"Power cycle": 5, "Power down": 8, "Power on to IDE-R CDROM": 203, "Power on to IDE-R Floppy": 201, "Power on to PXE": 401, "Power up": 2, "Reset": 10, "Reset to IDE-R CDROM": 202, "Reset to IDE-R Floppy": 200, "Reset to PXE": 400}),
			err: nil,
		},
		{
			name:    "GetById fails",
			manMock: func(man *MockManagement) {},
			repoMock: func(repo *MockRepository) {
				repo.EXPECT().
					GetByID(gomock.Any(), device.GUID, "").
					Return(nil, errTest)
			},
			res: (map[string]interface{})(nil),
			err: utils.ErrNotFound,
		},
		{
			name: "GetPowerCapabilities fails",
			manMock: func(man *MockManagement) {
				man.EXPECT().
					SetupWsmanClient(gomock.Any(), false, true).
					Return()
				man.EXPECT().
					GetPowerCapabilities().
					Return(boot.BootCapabilitiesResponse{}, ErrGeneral)
				man.EXPECT().
					GetAMTVersion().
					Return(nil, nil)
			},
			repoMock: func(repo *MockRepository) {
				repo.EXPECT().
					GetByID(gomock.Any(), device.GUID, "").
					Return(device, nil)
			},
			res: (map[string]interface{})(nil),
			err: ErrGeneral,
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			useCase, management, repo := initPowerTest(t)
			tc.manMock(management)
			tc.repoMock(repo)
			res, err := useCase.GetPowerCapabilities(context.Background(), device.GUID)
			require.Equal(t, tc.res, res)
			require.Equal(t, tc.err, err)
		})
	}
}

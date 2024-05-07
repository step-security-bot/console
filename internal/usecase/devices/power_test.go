package devices_test

import (
	"context"
	"testing"

	"github.com/open-amt-cloud-toolkit/console/internal/entity"
	devices "github.com/open-amt-cloud-toolkit/console/internal/usecase/devices"
	"github.com/open-amt-cloud-toolkit/console/pkg/logger"
	"github.com/open-amt-cloud-toolkit/go-wsman-messages/v2/pkg/wsman/cim/power"
	"github.com/stretchr/testify/require"
	gomock "go.uber.org/mock/gomock"
)

type powerTestType struct {
	name     string
	guid     string
	action   int
	manMock  func(*MockManagement)
	repoMock func(*MockRepository)
	res      power.PowerActionResponse
	err      error
}

func powerTest(t *testing.T) (*devices.UseCase, *MockManagement, *MockRepository) {
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

	tests := []powerTestType{
		{
			name:   "success",
			guid:   "test-1234",
			action: 0,
			manMock: func(man *MockManagement) {
				man.EXPECT().
					SetupWsmanClient(gomock.Any(), false, true).
					Return()
				man.EXPECT().
					SendPowerAction(0).
					Return()
			},
			repoMock: func(repo *MockRepository) {
				repo.EXPECT().
					GetByID(gomock.Any(), "device-guid-123", "tenant-id-456").
					Return(device, nil)
			},
			res: power.PowerActionResponse{},
			err: nil,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			useCase, management, repo := powerTest(t)

			tc.manMock(management)
			tc.repoMock(repo)

			res, err := useCase.SendPowerAction(context.Background(), tc.guid, tc.action)

			require.Equal(t, tc.res, res)
			require.Equal(t, tc.err, err)
		})
	}
}

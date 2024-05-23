package devices_test

import (
	"testing"

	"github.com/open-amt-cloud-toolkit/console/internal/entity/dto"
	devices "github.com/open-amt-cloud-toolkit/console/internal/usecase/devices"
	"github.com/open-amt-cloud-toolkit/go-wsman-messages/v2/pkg/wsman"
	"github.com/stretchr/testify/require"
	gomock "go.uber.org/mock/gomock"
)

func initRedirectionTest(t *testing.T) (*devices.Redirector, *MockRedirection, *MockRepository) {
	t.Helper()

	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()

	repo := NewMockRepository(mockCtl)
	redirect := NewMockRedirection(mockCtl)
	u := &devices.Redirector{}

	return u, redirect, repo
}

type redTest struct {
	name    string
	redMock func(*MockRedirection)
	//repoMock func(*MockRepository)
	res any
}

func TestSetupWsmanClient(t *testing.T) {
	t.Parallel()

	device := &dto.Device{
		GUID:     "device-guid-123",
		TenantID: "tenant-id-456",
	}

	tests := []redTest{
		{
			name: "success",
			redMock: func(redirect *MockRedirection) {
				redirect.EXPECT().
					SetupWsmanClient(gomock.Any(), false, true).
					Return(wsman.Messages{})
			},
			res: wsman.Messages{},
		},
		{
			name: "fail",
			redMock: func(redirect *MockRedirection) {
				redirect.EXPECT().
					SetupWsmanClient(gomock.Any(), true, true).
					Return(wsman.Messages{})
			},
			res: wsman.Messages{},
		},
	}

	for _, tc := range tests {
		tc := tc // Necessary for proper parallel execution
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			redirector, redirect, _ := initRedirectionTest(t)

			tc.redMock(redirect)

			res := redirector.SetupWsmanClient(*device, true, true)

			require.IsType(t, tc.res, res)
		})
	}
}

func TestNewRedirector(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
	}{
		{
			name: "success",
		},
	}

	for _, tc := range tests {
		tc := tc // Necessary for proper parallel execution
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Call the function under test
			redirector := devices.NewRedirector()

			// Assert that the returned redirector is not nil
			require.NotNil(t, redirector)
		})
	}
}

// func TestRedirectConnect(t *testing.T) {
// 	t.Parallel()

// 	device := &dto.Device{
// 		GUID:     "device-guid-123",
// 		TenantID: "tenant-id-456",
// 	}
// 	conn := &websocket.Conn{}
// 	deviceConnection := &devices.DeviceConnection{
// 		Conn:      conn,
// 		Device:    *device,
// 		Direct:    false,
// 		Challenge: client.AuthChallenge{},
// 	}

// 	tests := []redTest{
// 		{
// 			name: "success",
// 			redMock: func(redirect *MockRedirection) {
// 				redirect.EXPECT().
// 					SetupWsmanClient(gomock.Any(), false, true).
// 					Return(wsman.Messages{})
// 			},
// 			res: nil,
// 		},
// 	}

// 	for _, tc := range tests {
// 		tc := tc // Necessary for proper parallel execution
// 		t.Run(tc.name, func(t *testing.T) {
// 			t.Parallel()

// 			redirector, redirect, _ := initRedirectionTest(t)

// 			tc.redMock(redirect)

// 			err := redirector.RedirectConnect(context.Background(), deviceConnection)

// 			require.IsType(t, tc.err, err)
// 		})
// 	}
// }

package devices_test

import (
	"context"
	"testing"

	"github.com/gorilla/websocket"
	"github.com/open-amt-cloud-toolkit/console/internal/entity/dto"
	devices "github.com/open-amt-cloud-toolkit/console/internal/usecase/devices"
	"github.com/open-amt-cloud-toolkit/console/pkg/logger"
	"github.com/open-amt-cloud-toolkit/go-wsman-messages/v2/pkg/wsman/client"
	gomock "go.uber.org/mock/gomock"
)

func initInterceptorTest(t *testing.T) (*devices.UseCase, *MockRedirection, *MockRepository) {
	t.Helper()

	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()

	repo := NewMockRepository(mockCtl)
	redirect := NewMockRedirection(mockCtl)
	log := logger.New("error")
	u := devices.New(repo, NewMockManagement(mockCtl), redirect, log)

	return u, redirect, repo
}

type redirectTest struct {
	name     string
	redMock  func(*MockRedirection)
	repoMock func(*MockRepository)
	//res      any
	err error
}

func TestListenToDevice(t *testing.T) {
	ctx := context.Background()
	conn := &websocket.Conn{}
	device := &dto.Device{
		GUID:     "test-guid",
		Username: "user",
		Password: "password",
	}
	deviceConnection := &devices.DeviceConnection{
		Conn:      conn,
		Device:    *device,
		Direct:    false,
		Challenge: client.AuthChallenge{},
	}

	useCase, redirect, _ := initInterceptorTest(t)

	redirect.EXPECT().RedirectListen(gomock.Any(), deviceConnection).Return([]byte{}, nil)
	redirect.EXPECT().RedirectListen(gomock.Any(), deviceConnection).Return([]byte{1, 2, 3}, nil)

	useCase.ListenToDevice(ctx, deviceConnection)

}

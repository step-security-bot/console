package devices_test

import (
	"testing"

	devices "github.com/open-amt-cloud-toolkit/console/internal/usecase/devices"
	"github.com/open-amt-cloud-toolkit/console/pkg/logger"
	gomock "go.uber.org/mock/gomock"
)

func initRedirectionTest(t *testing.T) (*devices.UseCase, *MockRedirection, *MockRepository) {
	t.Helper()

	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()

	repo := NewMockRepository(mockCtl)
	redirect := NewMockRedirection(mockCtl)
	log := logger.New("error")
	u := devices.New(repo, nil, redirect, log)

	return u, redirect, repo
}

type redTest struct {
	name     string
	redMock  func(*MockRedirection)
	repoMock func(*MockRepository)
	res      any
	err      error
}

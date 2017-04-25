package spies

import (
	"github.com/continuouspipe/remote-environment-client/sync/options"
	"github.com/stretchr/testify/mock"
)

//Mock for RsyncFetch
type SpyRsyncFetch struct {
	mock.Mock
}

func NewSpyRsyncFetch() *SpyRsyncFetch {
	return &SpyRsyncFetch{}
}

func (s SpyRsyncFetch) Fetch(filePath string) error {
	s.Called(filePath)
	return nil
}

func (s SpyRsyncFetch) SetOptions(syncOptions options.SyncOptions) {
	s.Called(syncOptions)
}

//Mock for RsyncPush
type SpyRsyncSyncer struct {
	mock.Mock
}

func NewSpyRsyncSyncer() *SpyRsyncSyncer {
	return &SpyRsyncSyncer{}
}

func (s SpyRsyncSyncer) Sync(filePaths []string) error {
	s.Called(filePaths)
	return nil
}

func (s SpyRsyncSyncer) SetOptions(syncOptions options.SyncOptions) {
	s.Called(syncOptions)
}

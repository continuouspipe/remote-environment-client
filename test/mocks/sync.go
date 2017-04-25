package mocks

import (
	"github.com/continuouspipe/remote-environment-client/sync/options"
	"github.com/stretchr/testify/mock"
)

//Spy for Syncer interface
type SpySyncer struct {
	mock.Mock
}

func NewSpySyncer() *SpySyncer {
	return &SpySyncer{}
}

func (s SpySyncer) Sync(filePaths []string) error {
	s.Called(filePaths)
	return nil
}

func (s SpySyncer) SetOptions(syncOptions options.SyncOptions) {
	s.Called(syncOptions)
}

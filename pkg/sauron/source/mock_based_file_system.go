package source

import (
	"context"
	"io"

	"github.com/stretchr/testify/mock"
)

// MockBasedFileSystem is a testify mock implementing FileSystem.
type MockBasedFileSystem struct {
	mock.Mock
}

// List records the call and returns the configured values.
func (m *MockBasedFileSystem) List(ctx context.Context, uri string, opts ...Option) ([]File, error) {
	args := m.Called(ctx, uri, opts)

	var files []File
	if v := args.Get(0); v != nil {
		files = v.([]File)
	}

	return files, args.Error(1)
}

// Describe records the call and returns the configured values.
func (m *MockBasedFileSystem) Describe(ctx context.Context, uri string) (Stat, error) {
	args := m.Called(ctx, uri)

	var stat Stat
	if v := args.Get(0); v != nil {
		stat = v.(Stat)
	}

	return stat, args.Error(1)
}

// Get records the call and returns the configured values.
func (m *MockBasedFileSystem) Get(ctx context.Context, uri string) (File, error) {
	args := m.Called(ctx, uri)

	var file File
	if v := args.Get(0); v != nil {
		file = v.(File)
	}

	return file, args.Error(1)
}

// Fetch records the call and returns the configured values. Each returned
// File's Name() is its path relative to the artifact directory.
func (m *MockBasedFileSystem) Fetch(ctx context.Context, uri string) ([]File, error) {
	args := m.Called(ctx, uri)

	var files []File
	if v := args.Get(0); v != nil {
		files = v.([]File)
	}

	return files, args.Error(1)
}

// MockBasedStat is a testify mock implementing Stat.
type MockBasedStat struct {
	mock.Mock
}

// Name records the call and returns the configured value.
func (m *MockBasedStat) Name() string {
	return m.Called().String(0)
}

// IsDirectory records the call and returns the configured value.
func (m *MockBasedStat) IsDirectory() bool {
	return m.Called().Bool(0)
}

// Size records the call and returns the configured value.
func (m *MockBasedStat) Size() int64 {
	return m.Called().Get(0).(int64)
}

// Version records the call and returns the configured value.
func (m *MockBasedStat) Version() string {
	return m.Called().String(0)
}

// MockBasedFile is a testify mock implementing File.
type MockBasedFile struct {
	MockBasedStat
}

// Read records the call and returns the configured values.
func (m *MockBasedFile) Read(ctx context.Context) (io.ReadCloser, error) {
	args := m.Called(ctx)

	var rc io.ReadCloser
	if v := args.Get(0); v != nil {
		rc = v.(io.ReadCloser)
	}

	return rc, args.Error(1)
}

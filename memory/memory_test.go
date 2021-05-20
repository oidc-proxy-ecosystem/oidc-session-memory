package memory

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOutputOfLogFile(t *testing.T) {
	ctx := context.Background()
	memory := newMemorySession()
	memory.Init(ctx, map[string]interface{}{
		"prefix":        "/memory",
		"filename":      "./test.log",
		"loglevel":      "debug",
		"logformat":     "short",
		"logdateformat": "date",
	})
	defer memory.Close(ctx)
	assert.IsType(t, &os.File{}, memory.writer)
	tests := []struct {
		name string
		fn   func(t *testing.T)
	}{
		{
			name: "put no error",
			fn: func(t *testing.T) {
				err := memory.Put(ctx, "key", "test")
				assert.NoError(t, err)
			},
		},
		{
			name: "get is exists",
			fn: func(t *testing.T) {
				item, err := memory.Get(ctx, "key")
				assert.NoError(t, err)
				assert.Equal(t, "test", item)
			},
		},
		{
			name: "delete no error",
			fn: func(t *testing.T) {
				err := memory.Delete(ctx, "key")
				assert.NoError(t, err)
			},
		},
		{
			name: "get is not exists",
			fn: func(t *testing.T) {
				item, err := memory.Get(ctx, "key")
				assert.NoError(t, err)
				assert.Equal(t, "", item)
			},
		},
		{
			name: "log file is exists",
			fn: func(t *testing.T) {
				assert.FileExists(t, "./test.log")
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, tt.fn)
	}
	os.Remove("./test.log")
}

func TestOutputOfStdout(t *testing.T) {
	ctx := context.Background()
	memory := newMemorySession()
	memory.Init(ctx, map[string]interface{}{
		"prefix":        "/memory",
		"filename":      "",
		"loglevel":      "debug",
		"logformat":     "short",
		"logdateformat": "date",
	})
	defer memory.Close(ctx)
	assert.IsType(t, &os.File{}, memory.writer)
	tests := []struct {
		name string
		fn   func(t *testing.T)
	}{
		{
			name: "put no error",
			fn: func(t *testing.T) {
				err := memory.Put(ctx, "key", "test")
				assert.NoError(t, err)
			},
		},
		{
			name: "get is exists",
			fn: func(t *testing.T) {
				item, err := memory.Get(ctx, "key")
				assert.NoError(t, err)
				assert.Equal(t, "test", item)
			},
		},
		{
			name: "delete no error",
			fn: func(t *testing.T) {
				err := memory.Delete(ctx, "key")
				assert.NoError(t, err)
			},
		},
		{
			name: "get is not exists",
			fn: func(t *testing.T) {
				item, err := memory.Get(ctx, "key")
				assert.NoError(t, err)
				assert.Equal(t, "", item)
			},
		},
		{
			name: "log file is exists",
			fn: func(t *testing.T) {
				assert.NoFileExists(t, "./test.log")
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, tt.fn)
	}
}

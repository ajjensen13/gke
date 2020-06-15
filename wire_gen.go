// Code generated by Wire. DO NOT EDIT.

//go:generate wire
//+build !wireinject

package gke

import (
	"context"
	"net/http"
)

// Injectors from log_wireinject.go:

func NewLogger(ctx context.Context) (Logger, func(), error) {
	logClient, cleanup, err := NewLogClient(ctx)
	if err != nil {
		return Logger{}, nil, err
	}
	string2 := _wireStringValue
	logger := provideDefaultLogger(logClient, string2)
	return logger, func() {
		cleanup()
	}, nil
}

var (
	_wireStringValue = DefaultLogID
)

// Injectors from server_wireinject.go:

func NewServer(ctx context.Context, handler http.Handler, lg Logger) (*http.Server, error) {
	server := provideServer(lg, handler)
	return server, nil
}

// Injectors from storage_wireinject.go:

func NewStorageClient(ctx context.Context) (StorageClient, func(), error) {
	storageClient, cleanup, err := provideStorageClient(ctx)
	if err != nil {
		return nil, nil, err
	}
	return storageClient, func() {
		cleanup()
	}, nil
}

// log_wireinject.go:

func provideDefaultLogger(client LogClient, logId string) Logger {
	return client.Logger(logId)
}

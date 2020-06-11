// Code generated by Wire. DO NOT EDIT.

//go:generate wire
//+build !wireinject

package gke

import (
	"cloud.google.com/go/logging"
	"cloud.google.com/go/storage"
	"context"
	"fmt"
	"github.com/ajjensen13/config"
	"log"
	"net"
	"net/http"
	"os"
	"path"
	"runtime/debug"
	"sync"
	"time"
)

// Injectors from wire.go:

func NewLogClient(ctx context.Context) (LogClient, func(), error) {
	config, err := provideConfig()
	if err != nil {
		return nil, nil, err
	}
	logParentId := provideLogParentId(config)
	v := _wireValue
	gkeLogClient, cleanup, err := NewLogClientWithOptions(ctx, logParentId, v...)
	if err != nil {
		return nil, nil, err
	}
	return gkeLogClient, func() {
		cleanup()
	}, nil
}

var (
	_wireValue = []logging.LoggerOption{}
)

func NewLogClientWithOptions(ctx context.Context, parent LogParentId, opts ...logging.LoggerOption) (LogClient, func(), error) {
	client, cleanup, err := provideLoggingClient(ctx, parent)
	if err != nil {
		return nil, nil, err
	}
	config, err := provideConfig()
	if err != nil {
		cleanup()
		return nil, nil, err
	}
	gkeLogClient, err := provideLogClient(client, config, opts)
	if err != nil {
		cleanup()
		return nil, nil, err
	}
	return gkeLogClient, func() {
		cleanup()
	}, nil
}

func NewLogger(logc LogClient) (Logger, func(), error) {
	logId := NewLogId()
	v := _wireValue2
	gkeLogger, cleanup, err := NewLoggerWithOptions(logc, logId, v...)
	if err != nil {
		return nil, nil, err
	}
	return gkeLogger, func() {
		cleanup()
	}, nil
}

var (
	_wireValue2 = []logging.LoggerOption{}
)

func NewLoggerWithOptions(logc LogClient, logId LogId, opts ...logging.LoggerOption) (Logger, func(), error) {
	gkeLogger, cleanup := provideLogger(logc, logId, opts...)
	return gkeLogger, func() {
		cleanup()
	}, nil
}

func NewStorageClient(ctx context.Context) (StorageClient, func(), error) {
	storageClient, cleanup, err := provideStorageClient(ctx)
	if err != nil {
		return nil, nil, err
	}
	return storageClient, func() {
		cleanup()
	}, nil
}

func NewServer(ctx context.Context, handler http.Handler, lg Logger) (*http.Server, error) {
	server := provideServer(lg, handler)
	return server, nil
}

func NewLogParentId() (LogParentId, error) {
	config, err := provideConfig()
	if err != nil {
		return "", err
	}
	logParentId := provideLogParentId(config)
	return logParentId, nil
}

// wire.go:

type LogParentId string

func provideLogParentId(cfg *Config) LogParentId {
	return LogParentId(cfg.ProjectId)
}

type LogId string

func NewLogId() LogId {
	result, ok := os.LookupEnv("GKE_LOG_ID")
	if ok {
		return LogId(result)
	}

	info, ok := debug.ReadBuildInfo()
	if ok {
		return LogId(path.Base(info.Path))
	}

	return LogId(os.Args[0])
}

func provideLogger(logc LogClient, logId LogId, opt ...logging.LoggerOption) (Logger, func()) {
	l := logc.Logger(string(logId), opt...)
	return l, func() { _ = l.Flush() }
}

func provideServer(lg Logger, handler http.Handler) *http.Server {
	return &http.Server{
		Handler:           handler,
		ReadTimeout:       time.Second * 30,
		ReadHeaderTimeout: time.Second * 5,
		WriteTimeout:      time.Second * 30,
		IdleTimeout:       time.Second * 60,
		MaxHeaderBytes:    http.DefaultMaxHeaderBytes,
		ErrorLog:          lg.StandardLogger(logging.Error),
		BaseContext: func(_ net.Listener) (ctx context.Context) {
			ctx, _ = Alive()
			return
		},
	}
}

var (
	pkgConfigOnce sync.Once
	pkgConfig     *Config
	pkgConfigErr  error
)

type Config struct {
	ProjectId       string            `yaml:"projectId"`
	CommonLogLabels map[string]string `yaml:"commonLogLabels"`
}

func provideConfig() (*Config, error) {
	pkgConfigOnce.Do(func() {
		var c Config
		pkgConfigErr = config.InterfaceYaml("gke.yaml", &c)
		if pkgConfigErr != nil {
			return
		}
		pkgConfig = &c
	})
	return pkgConfig, pkgConfigErr
}

func provideLoggingClient(ctx context.Context, parent LogParentId) (*logging.Client, func(), error) {
	result, err := logging.NewClient(ctx, string(parent))
	if err != nil {
		return nil, nil, err
	}

	result.OnError = func(err error) {
		log.Printf("%v", err)
	}

	err = result.Ping(ctx)
	if err != nil {
		return nil, nil, err
	}

	return result, func() { _ = result.Close() }, nil
}

func provideLogClient(client *logging.Client, config2 *Config, opts []logging.LoggerOption) (LogClient, error) {
	var defaultOpts []logging.LoggerOption
	if config2.CommonLogLabels != nil && len(config2.CommonLogLabels) > 0 {
		defaultOpts = append(defaultOpts, logging.CommonLabels(config2.CommonLogLabels))
	}
	if len(opts) > 0 {
		defaultOpts = append(defaultOpts, opts...)
	}

	return &logClient{
		Client: client,
		opts:   defaultOpts,
	}, nil
}

type logClient struct {
	opts []logging.LoggerOption
	*logging.Client
}

func (l *logClient) Logger(logID string, opts ...logging.LoggerOption) Logger {
	copyOpts := append(l.opts, opts...)
	return &logger{
		logId:  logID,
		client: l,
		opts:   copyOpts,
		Logger: l.Client.Logger(
			logID,
			copyOpts...,
		),
	}
}

type logger struct {
	logId  string
	client *logClient
	opts   []logging.LoggerOption
	*logging.Logger
}

type fmtPayload struct {
	Message string
	Args    []interface{}
}

type errPayload struct {
	Message string
	Err     error
}

func (l *logger) Logf(severity logging.Severity, format string, args ...interface{}) string {
	message := fmt.Sprintf(format, args...)
	l.Log(logging.Entry{Severity: severity, Payload: fmtPayload{
		Message: message,
		Args:    args,
	}})
	return message
}

func (l *logger) Infof(format string, args ...interface{}) string {
	return l.Logf(logging.Info, format, args...)
}

func (l *logger) Noticef(format string, args ...interface{}) string {
	return l.Logf(logging.Notice, format, args...)
}

func (l *logger) Warnf(format string, args ...interface{}) string {
	return l.Logf(logging.Warning, format, args...)
}

func (l *logger) Errorf(format string, args ...interface{}) string {
	return l.Logf(logging.Error, format, args...)
}

func (l *logger) InfoErr(err error) error {
	l.Log(logging.Entry{Severity: logging.Info, Payload: errPayload{
		Message: fmt.Sprintf("%v", err),
		Err:     err,
	}})
	return err
}

func (l *logger) NoticeErr(err error) error {
	l.Log(logging.Entry{Severity: logging.Notice, Payload: errPayload{
		Message: fmt.Sprintf("%v", err),
		Err:     err,
	}})
	return err
}

func (l *logger) WarnErr(err error) error {
	l.Log(logging.Entry{Severity: logging.Warning, Payload: errPayload{
		Message: fmt.Sprintf("%v", err),
		Err:     err,
	}})
	return err
}

func (l *logger) ErrorErr(err error) error {
	l.Log(logging.Entry{Severity: logging.Error, Payload: errPayload{
		Message: fmt.Sprintf("%v", err),
		Err:     err,
	}})
	return err
}

type LogClient interface {
	Logger(logID string, opts ...logging.LoggerOption) Logger
}

type Logger interface {
	StandardLogger(severity logging.Severity) *log.Logger
	Log(entry logging.Entry)
	Logf(severity logging.Severity, format string, args ...interface{}) string
	Infof(format string, args ...interface{}) string
	Noticef(format string, args ...interface{}) string
	Warnf(format string, args ...interface{}) string
	Errorf(format string, args ...interface{}) string
	InfoErr(err error) error
	NoticeErr(err error) error
	WarnErr(err error) error
	ErrorErr(err error) error
	Flush() error
}

func provideStorageClient(ctx context.Context) (StorageClient, func(), error) {
	result, err := storage.NewClient(ctx)
	if err != nil {
		return nil, nil, err
	}
	return result, func() { _ = result.Close() }, nil
}

type StorageClient interface {
	HMACKeyHandle(projectID, accessID string) *storage.HMACKeyHandle
	CreateHMACKey(ctx context.Context, projectID, serviceAccountEmail string, opts ...storage.HMACKeyOption) (*storage.HMACKey, error)
	ListHMACKeys(ctx context.Context, projectID string, opts ...storage.HMACKeyOption) *storage.HMACKeysIterator
	ServiceAccount(ctx context.Context, projectID string) (string, error)
	Bucket(name string) *storage.BucketHandle
	Buckets(ctx context.Context, projectID string) *storage.BucketIterator
}

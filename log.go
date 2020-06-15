package gke

import (
	"cloud.google.com/go/logging"
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	stdlog "log"
	"os"
	"path"
	"runtime"
	"runtime/debug"

	logpb "google.golang.org/genproto/googleapis/logging/v2"

	"github.com/ajjensen13/gke/internal/log"
	"github.com/ajjensen13/gke/internal/metadata"
)

var (
	// True to log to GKE. This will default to true if running on GCE.
	LogGke bool
	// True to log to the standard logger. This will default to true if not running on GCE.
	LogStd bool
)

func init() {
	md, err := Metadata()
	switch {
	case err == nil:
		LogGke = true
		DefaultLogID = md.ContainerName
	case errors.Is(err, metadata.ErrNotOnGCE):
		LogStd = true
		bi, ok := debug.ReadBuildInfo()
		if ok {
			DefaultLogID = path.Base(bi.Path)
		}
		if DefaultLogID == "" {
			DefaultLogID = os.Args[0]
		}
	default:
		panic(fmt.Errorf("failed to setup logging: %w", err))
	}
}

// DefaultLogID will be metadata.Metadata().ContainerName if running on GCE.
// Otherwise, it will attempt to detect the name from the build info or the
// program arguments.
var DefaultLogID string

// NewLogClient returns a log client. The context should remain open for the life of the log client.
// Note: ctx should usually be context.Background() to ensure that the logging
// events occur event after AliveContext() is canceled.
func NewLogClient(ctx context.Context) (LogClient, func(), error) {
	var result log.MultiClient
	var cleanup = func() {}

	if LogGke {
		md, err := Metadata()
		if err != nil {
			return LogClient{}, nil, fmt.Errorf("LogGke is true, but metadata cannot be found: %w", err)
		}

		parent := md.ProjectID
		client, err := log.NewGkeClient(ctx, "projects/"+parent)
		if err != nil {
			return LogClient{}, func() {}, err
		}
		err = client.Ping(ctx)
		if err != nil {
			return LogClient{}, func() {}, err
		}
		result = append(result, client)
		prevCleanup := cleanup
		cleanup = func() { prevCleanup(); _ = client.Close() }
	}

	if LogStd {
		client := log.NewStandardClient(os.Stderr)
		result = append(result, client)
		prevCleanup := cleanup
		cleanup = func() { prevCleanup(); _ = client.Close() }
	}

	if len(result) == 0 {
		result = append(result, log.NewStandardClient(ioutil.Discard))
	}

	return LogClient{result}, cleanup, nil
}

type LogClient struct {
	log.Client
}

// Logger returns a new Logger. If logId is empty, then DefaultLogID is used.
func (lc LogClient) Logger(logId string) Logger {
	if logId == "" {
		logId = DefaultLogID
	}
	return Logger{lc.Client.Logger(logId)}
}

// Logger logs entries to a single log.
type Logger struct {
	log.Logger
}

// StandardLogger returns a *log.Logger for a given severity.
func (l Logger) StandardLogger(severity logging.Severity) *stdlog.Logger {
	return l.Logger.StandardLogger(severity)
}

func (l Logger) logPayload(severity logging.Severity, payload interface{}) {
	var sl *logpb.LogEntrySourceLocation
	if _, file, line, ok := runtime.Caller(2); ok {
		sl = &logpb.LogEntrySourceLocation{File: file, Line: int64(line)}
	}
	l.Logger.Log(logging.Entry{Severity: severity, Payload: payload, SourceLocation: sl})
}

// MsgData is a convenience type for logging a message with additional data.
// It is provided for consistency in logging across GKE applications.
type MsgData struct {
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

// NewMsgData returns a MsgData.
// If len(data) == 0, then result.Data will be nil.
// If len(data) == 1, then result.Data will be data[0] (interface{}).
// Otherwise, result.Data will be data ([]interface{}).
func NewMsgData(msg string, data ...interface{}) (result MsgData) {
	switch len(data) {
	case 0:
		return MsgData{msg, nil}
	case 1:
		return MsgData{msg, data[0]}
	default:
		return MsgData{msg, data}
	}
}

// NewFmtMsgData is equivalent to gke.NewMsgData(fmt.Sprintf(msg, data...), data...).
func NewFmtMsgData(msg string, data ...interface{}) MsgData {
	return NewMsgData(fmt.Sprintf(msg, data...), data...)
}

func (l Logger) log(severity logging.Severity, payload interface{}) {
	l.logPayload(severity, payload)
}

// Default creates a log entry with a Default severity.
//
// Note: Default means the log entry has no assigned severity level.
func (l Logger) Default(payload interface{}) {
	l.log(logging.Default, payload)
}

// Default creates a log entry with a Debug severity.
//
// Note: Debug means debug or trace information.
func (l Logger) Debug(payload interface{}) {
	l.log(logging.Debug, payload)
}

// Default creates a log entry with a Info severity.
//
// Note: Info means routine information, such as ongoing status or performance.
func (l Logger) Info(payload interface{}) {
	l.log(logging.Info, payload)
}

// Default creates a log entry with a Notice severity.
//
// Note: Notice means normal but significant events, such as start up, shut down, or configuration.
func (l Logger) Notice(payload interface{}) {
	l.log(logging.Notice, payload)
}

// Default creates a log entry with a Warning severity.
//
// Note: Warning means events that might cause problems.
func (l Logger) Warning(payload interface{}) {
	l.log(logging.Warning, payload)
}

// Default creates a log entry with an Error severity.
//
// Note: Error means events that are likely to cause problems.
func (l Logger) Error(payload interface{}) {
	l.log(logging.Error, payload)
}

// Default creates a log entry with a Critical severity.
//
// Note: Critical means events that cause more severe problems or brief outages.
func (l Logger) Critical(payload interface{}) {
	l.log(logging.Critical, payload)
}

// Default creates a log entry with an Alert severity.
//
// Note: Alert means a person must take an action immediately.
func (l Logger) Alert(payload interface{}) {
	l.log(logging.Alert, payload)
}

// Default creates a log entry with an Emergency severity.
//
// Note: Emergency means one or more systems are unusable.
func (l Logger) Emergency(payload interface{}) {
	l.log(logging.Emergency, payload)
}

func (l Logger) logf(severity logging.Severity, format string, args ...interface{}) string {
	str := fmt.Sprintf(format, args...)
	l.logPayload(severity, str)
	return str
}

// Defaultf creates a log entry with a Default severity with a formatted string payload.
// The return is the formatted string as created by fmt.Sprintf(format, args...)
//
// Note: Default means the log entry has no assigned severity level.
func (l Logger) Defaultf(format string, args ...interface{}) string {
	return l.logf(logging.Default, format, args...)
}

// Debugf creates a log entry with a Debug severity with a formatted string payload.
// The return is the formatted string as created by fmt.Sprintf(format, args...)
//
// Note: Debug means debug or trace information.
func (l Logger) Debugf(format string, args ...interface{}) string {
	return l.logf(logging.Debug, format, args...)
}

// Infof creates a log entry with a Info severity with a formatted string payload.
// The return is the formatted string as created by fmt.Sprintf(format, args...)
//
// Note: Info means routine information, such as ongoing status or performance.
func (l Logger) Infof(format string, args ...interface{}) string {
	return l.logf(logging.Info, format, args...)
}

// Noticef creates a log entry with a Notice severity with a formatted string payload.
// The return is the formatted string as created by fmt.Sprintf(format, args...)
//
// Note: Notice means normal but significant events, such as start up, shut down, or configuration.
func (l Logger) Noticef(format string, args ...interface{}) string {
	return l.logf(logging.Notice, format, args...)
}

// Warningf creates a log entry with a Warning severity with a formatted string payload.
// The return is the formatted string as created by fmt.Sprintf(format, args...)
//
// Note: Warning means events that might cause problems.
func (l Logger) Warningf(format string, args ...interface{}) string {
	return l.logf(logging.Warning, format, args...)
}

// Errorf creates a log entry with an Error severity with a formatted string payload.
// The return is the formatted string as created by fmt.Sprintf(format, args...)
//
// Note: Error means events that are likely to cause problems.
func (l Logger) Errorf(format string, args ...interface{}) string {
	return l.logf(logging.Error, format, args...)
}

// Criticalf creates a log entry with a Critical severity with a formatted string payload.
// The return is the formatted string as created by fmt.Sprintf(format, args...)
//
// Note: Critical means events that cause more severe problems or brief outages.
func (l Logger) Criticalf(format string, args ...interface{}) string {
	return l.logf(logging.Critical, format, args...)
}

// Alertf creates a log entry with an Alert severity with a formatted string payload.
// The return is the formatted string as created by fmt.Sprintf(format, args...)
//
// Note: Alert means a person must take an action immediately.
func (l Logger) Alertf(format string, args ...interface{}) string {
	return l.logf(logging.Alert, format, args...)
}

// Emergencyf creates a log entry with an Emergency severity with a formatted string payload.
// The return is the formatted string as created by fmt.Sprintf(format, args...)
//
// Note: Emergency means one or more systems are unusable.
func (l Logger) Emergencyf(format string, args ...interface{}) string {
	return l.logf(logging.Emergency, format, args...)
}

func (l Logger) logErr(severity logging.Severity, err error) error {
	if err != nil {
		str := fmt.Sprintf("%v", err)
		l.logPayload(severity, str)
	}
	return err
}

// DefaultErr creates a log entry with a Default severity with an error as its payload.
// The error is converted into a string via fmt.Sprintf("%v", err) before sending to
// avoid possible serialization errors. The return value is err.
//
// Note: Default means the log entry has no assigned severity level.
func (l Logger) DefaultErr(err error) error {
	return l.logErr(logging.Default, err)
}

// DebugErr creates a log entry with a Debug severity with an error as its payload.
// The error is converted into a string via fmt.Sprintf("%v", err) before sending to
// avoid possible serialization errors. The return value is err.
//
// Note: Debug means debug or trace information.
func (l Logger) DebugErr(err error) error {
	return l.logErr(logging.Debug, err)
}

// InfoErr creates a log entry with a Info severity with an error as its payload.
// The error is converted into a string via fmt.Sprintf("%v", err) before sending to
// avoid possible serialization errors. The return value is err.
//
// Note: Info means routine information, such as ongoing status or performance.
func (l Logger) InfoErr(err error) error {
	return l.logErr(logging.Info, err)
}

// NoticeErr creates a log entry with a Notice severity with an error as its payload.
// The error is converted into a string via fmt.Sprintf("%v", err) before sending to
// avoid possible serialization errors. The return value is err.
//
// Note: Notice means normal but significant events, such as start up, shut down, or configuration.
func (l Logger) NoticeErr(err error) error {
	return l.logErr(logging.Notice, err)
}

// WarningErr creates a log entry with a Warning severity with an error as its payload.
// The error is converted into a string via fmt.Sprintf("%v", err) before sending to
// avoid possible serialization errors. The return value is err.
//
// Note: Warning means events that might cause problems.
func (l Logger) WarningErr(err error) error {
	return l.logErr(logging.Warning, err)
}

// ErrorErr creates a log entry with an Error severity with an error as its payload.
// The error is converted into a string via fmt.Sprintf("%v", err) before sending to
// avoid possible serialization errors. The return value is err.
//
// Note: Error means events that are likely to cause problems.
func (l Logger) ErrorErr(err error) error {
	return l.logErr(logging.Error, err)
}

// CriticalErr creates a log entry with a Critical severity with an error as its payload.
// The error is converted into a string via fmt.Sprintf("%v", err) before sending to
// avoid possible serialization errors. The return value is err.
//
// Note: Critical means events that cause more severe problems or brief outages.
func (l Logger) CriticalErr(err error) error {
	return l.logErr(logging.Critical, err)
}

// AlertErr creates a log entry with an Alert severity with an error as its payload.
// The error is converted into a string via fmt.Sprintf("%v", err) before sending to
// avoid possible serialization errors. The return value is err.
//
// Note: Alert means a person must take an action immediately.
func (l Logger) AlertErr(err error) error {
	return l.logErr(logging.Alert, err)
}

// EmergencyErr creates a log entry with an Emergency severity with an error as its payload.
// The error is converted into a string via fmt.Sprintf("%v", err) before sending to
// avoid possible serialization errors. The return value is err.
//
// Note: Emergency means one or more systems are unusable.
func (l Logger) EmergencyErr(err error) error {
	return l.logErr(logging.Emergency, err)
}

package gke

import (
	"cloud.google.com/go/logging"
	"context"
	"fmt"
	"io/ioutil"
	stdlog "log"
	"os"
	"runtime"

	logpb "google.golang.org/genproto/googleapis/logging/v2"

	"github.com/ajjensen13/gke/internal/log"
)

var (
	// True to log to GKE. This will default to true if OnGCE() returns true
	LogGke bool
	// True to log to the standard logger. This will default to true if OnGCE() returns false
	LogStd bool
)

func init() {
	if Metadata().OnGCE {
		LogGke = true
		return
	}
	LogStd = true
}

// NewLogClient returns a log client. The context should remain open for the life of the log client.
func NewLogClient(ctx context.Context) (LogClient, func(), error) {
	var result log.MultiClient
	var cleanup = func() {}

	if LogGke {
		parent := Metadata().ProjectID
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

func (lc LogClient) Logger(logId string, opts ...logging.LoggerOption) Logger {
	return Logger{lc.Client.Logger(logId, opts...)}
}

type Logger struct {
	log.Logger
}

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

func (l Logger) log(severity logging.Severity, args ...interface{}) {
	l.logPayload(severity, args)
}

// Default creates a log entry with a Default severity and the args as the payload ([]interface{}).
//
// Note: Default means the log entry has no assigned severity level.
func (l Logger) Default(args ...interface{}) {
	l.log(logging.Default, args)
}

// Default creates a log entry with a Debug severity and the args as the payload ([]interface{}).
//
// Note: Debug means debug or trace information.
func (l Logger) Debug(args ...interface{}) {
	l.log(logging.Debug, args)
}

// Default creates a log entry with a Info severity and the args as the payload ([]interface{}).
//
// Note: Info means routine information, such as ongoing status or performance.
func (l Logger) Info(args ...interface{}) {
	l.log(logging.Info, args)
}

// Default creates a log entry with a Notice severity and the args as the payload ([]interface{}).
//
// Note: Notice means normal but significant events, such as start up, shut down, or configuration.
func (l Logger) Notice(args ...interface{}) {
	l.log(logging.Notice, args)
}

// Default creates a log entry with a Warning severity and the args as the payload ([]interface{}).
//
// Note: Warning means events that might cause problems.
func (l Logger) Warning(args ...interface{}) {
	l.log(logging.Warning, args)
}

// Default creates a log entry with an Error severity and the args as the payload ([]interface{}).
//
// Note: Error means events that are likely to cause problems.
func (l Logger) Error(args ...interface{}) {
	l.log(logging.Error, args)
}

// Default creates a log entry with a Critical severity and the args as the payload ([]interface{}).
//
// Note: Critical means events that cause more severe problems or brief outages.
func (l Logger) Critical(args ...interface{}) {
	l.log(logging.Critical, args)
}

// Default creates a log entry with an Alert severity and the args as the payload ([]interface{}).
//
// Note: Alert means a person must take an action immediately.
func (l Logger) Alert(args ...interface{}) {
	l.log(logging.Alert, args)
}

// Default creates a log entry with an Emergency severity and the args as the payload ([]interface{}).
//
// Note: Emergency means one or more systems are unusable.
func (l Logger) Emergency(args ...interface{}) {
	l.log(logging.Emergency, args)
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

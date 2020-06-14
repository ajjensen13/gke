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
	LogGke bool
	LogStd bool
)

func init() {
	if OnGCE() {
		LogGke = true
		return
	}
	LogStd = true
}

func NewLogClient(ctx context.Context) (LogClient, func(), error) {
	var result log.MultiClient
	var cleanup = func() {}

	if LogGke {
		parent := Metadata().ProjectID
		client, err := log.NewGkeClient(ctx, "projects/"+parent)
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

func (l Logger) Default(args ...interface{}) {
	l.log(logging.Default, args)
}

func (l Logger) Debug(args ...interface{}) {
	l.log(logging.Debug, args)
}

func (l Logger) Info(args ...interface{}) {
	l.log(logging.Info, args)
}

func (l Logger) Notice(args ...interface{}) {
	l.log(logging.Notice, args)
}

func (l Logger) Warning(args ...interface{}) {
	l.log(logging.Warning, args)
}

func (l Logger) Error(args ...interface{}) {
	l.log(logging.Error, args)
}

func (l Logger) Alert(args ...interface{}) {
	l.log(logging.Alert, args)
}

func (l Logger) Critical(args ...interface{}) {
	l.log(logging.Critical, args)
}

func (l Logger) Emergency(args ...interface{}) {
	l.log(logging.Emergency, args)
}

func (l Logger) logf(severity logging.Severity, format string, args ...interface{}) string {
	str := fmt.Sprintf(format, args...)
	l.logPayload(severity, str)
	return str
}

func (l Logger) Defaultf(format string, args ...interface{}) string {
	return l.logf(logging.Default, format, args...)
}

func (l Logger) Debugf(format string, args ...interface{}) string {
	return l.logf(logging.Debug, format, args...)
}

func (l Logger) Infof(format string, args ...interface{}) string {
	return l.logf(logging.Info, format, args...)
}

func (l Logger) Noticef(format string, args ...interface{}) string {
	return l.logf(logging.Notice, format, args...)
}

func (l Logger) Warningf(format string, args ...interface{}) string {
	return l.logf(logging.Warning, format, args...)
}

func (l Logger) Errorf(format string, args ...interface{}) string {
	return l.logf(logging.Error, format, args...)
}

func (l Logger) Alertf(format string, args ...interface{}) string {
	return l.logf(logging.Alert, format, args...)
}

func (l Logger) Criticalf(format string, args ...interface{}) string {
	return l.logf(logging.Critical, format, args...)
}

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

func (l Logger) DefaultErr(err error) error {
	return l.logErr(logging.Default, err)
}

func (l Logger) DebugErr(err error) error {
	return l.logErr(logging.Debug, err)
}

func (l Logger) InfoErr(err error) error {
	return l.logErr(logging.Info, err)
}

func (l Logger) NoticeErr(err error) error {
	return l.logErr(logging.Notice, err)
}

func (l Logger) WarningErr(err error) error {
	return l.logErr(logging.Warning, err)
}

func (l Logger) ErrorErr(err error) error {
	return l.logErr(logging.Error, err)
}

func (l Logger) AlertErr(err error) error {
	return l.logErr(logging.Alert, err)
}

func (l Logger) CriticalErr(err error) error {
	return l.logErr(logging.Critical, err)
}

func (l Logger) EmergencyErr(err error) error {
	return l.logErr(logging.Emergency, err)
}

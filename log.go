/*
Copyright © 2020 A. Jensen <jensen.aaro@gmail.com>

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program. If not, see <http://www.gnu.org/licenses/>.
*/

package gke

import (
	"cloud.google.com/go/logging"
	"context"
	"errors"
	"fmt"
	stdlog "log"
	"os"
	"path"
	"runtime/debug"

	"github.com/ajjensen13/gke/internal/log"
)

// DefaultLogID will be metadata.Metadata().ContainerName if running on GCE.
// Otherwise, it will attempt to detect the name from the build info or the
// program arguments.
func DefaultLogID() (string, error) {
	md, err := Metadata()
	switch {
	case errors.Is(err, ErrNotOnGCE):
		bi, ok := debug.ReadBuildInfo()
		if ok {
			return path.Base(bi.Path), nil
		}
		return os.Args[0], nil
	case err == nil:
		return md.ContainerName, nil
	default:
		return "", fmt.Errorf("failed to determine default log id: %w", err)
	}
}

// NewLogClient returns a log client. The context should remain open for the life of the log client.
// Note: ctx should usually be context.Background() to ensure that the logging
// events occur event after AliveContext() is canceled.
func NewLogClient(ctx context.Context) (client LogClient, cleanup func(), err error) {
	md, err := Metadata()
	switch {
	case errors.Is(err, ErrNotOnGCE):
		client := log.NewStandardClient(os.Stderr)
		return LogClient{client}, func() { _ = client.Close() }, nil
	case err == nil:
		parent := md.ProjectID
		client, err := log.NewGkeClient(ctx, "projects/"+parent)
		if err != nil {
			return LogClient{}, func() {}, err
		}
		err = client.Ping(ctx)
		if err != nil {
			return LogClient{}, func() {}, err
		}
		return LogClient{client}, func() { _ = client.Close() }, nil
	default:
		return LogClient{}, func() {}, fmt.Errorf("failed to create logging client: %w", err)
	}
}

// LogClient is used to provision new loggers and close underlying connections during shutdown.
type LogClient struct {
	log.Client
}

// Logger returns a new Logger.
func (lc LogClient) Logger(logId string) Logger {
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
	entry := logging.Entry{Severity: severity, Payload: payload}
	SetupSourceLocation(&entry, 3)
	l.Logger.Log(entry)
}

func (l Logger) logPayloadSync(ctx context.Context, entry logging.Entry) error {
	SetupSourceLocation(&entry, 2)
	return l.Logger.LogSync(ctx, entry)
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

func (l Logger) LogSync(ctx context.Context, payload logging.Entry) error {
	return l.logPayloadSync(ctx, payload)
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

// SetupSourceLocation sets up the entry.SourceLocation field if it is not already set. If callDepth is 0, then
// the source location of the caller to SetupSourceLocation will be used. If 1, then the caller of that caller, etc, etc.
func SetupSourceLocation(entry *logging.Entry, callDepth int) {
	log.SetupSourceLocation(entry, 1+callDepth)
}

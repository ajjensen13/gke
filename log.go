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
	"fmt"
	"log"
	"os"
	"path"
	"runtime/debug"
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

type LogParentId string

func NewLogParentId() LogParentId {
	if OnGCE() {
		return LogParentId("projects/" + ProjectID())
	}
	return LogParentId("")
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

func provideLogger(logc LogClient, logId LogId) (Logger, func()) {
	l := logc.Logger(string(logId))
	return l, func() { _ = l.Flush() }
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

func provideLogClient(client *logging.Client) (LogClient, error) {
	return &logClient{Client: client}, nil
}

type logClient struct {
	*logging.Client
}

func (l *logClient) Logger(logID string) Logger {
	return &logger{
		logId:  logID,
		client: l,
		Logger: l.Client.Logger(logID),
	}
}

type logger struct {
	logId  string
	client *logClient
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
	if LogGke {
		l.Log(logging.Entry{Severity: severity, Payload: fmtPayload{
			Message: message,
			Args:    args,
		}})
	}
	if LogStd {
		log.Printf("%v %s", severity, message)
	}
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

func (l *logger) LogErr(severity logging.Severity, err error) error {
	if LogGke {
		l.Log(logging.Entry{Severity: severity, Payload: errPayload{
			Message: fmt.Sprintf("%v", err),
			Err:     err,
		}})
	}
	if LogStd {
		log.Printf("%v %v", severity, err)
	}
	return err
}

func (l *logger) InfoErr(err error) error {
	return l.LogErr(logging.Info, err)
}

func (l *logger) NoticeErr(err error) error {
	return l.LogErr(logging.Notice, err)
}

func (l *logger) WarnErr(err error) error {
	return l.LogErr(logging.Warning, err)
}

func (l *logger) ErrorErr(err error) error {
	return l.LogErr(logging.Error, err)
}

type LogClient interface {
	Logger(logID string) Logger
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

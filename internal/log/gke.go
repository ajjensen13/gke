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
package log

import (
	"cloud.google.com/go/logging"
	"context"
	mrpb "google.golang.org/genproto/googleapis/api/monitoredres"
	"strings"

	"github.com/ajjensen13/gke/internal/metadata"
)

// NewGkeClient returns a new logging client associated with the provided parent.
// A parent can take any of the following forms:
//    projects/PROJECT_ID
//    folders/FOLDER_ID
//    billingAccounts/ACCOUNT_ID
//    organizations/ORG_ID
// for backwards compatibility, a string with no '/' is also allowed and is interpreted
// as a project ID.
//
// Note: NewGkeClient uses WriteScope.
func NewGkeClient(ctx context.Context, parent string) (GkeClient, error) {
	client, err := logging.NewClient(ctx, parent)
	if err != nil {
		return GkeClient{}, err
	}
	return GkeClient{client}, nil
}

// GkeClient is a Logging client. A Client is associated with a single Cloud project.
type GkeClient struct {
	client *logging.Client
}

// Logger returns a Logger that will write entries with the given log ID, such as
// "syslog". A log ID must be less than 512 characters long and can only
// include the following characters: upper and lower case alphanumeric
// characters: [A-Za-z0-9]; and punctuation characters: forward-slash,
// underscore, hyphen, and period.
func (g GkeClient) Logger(logID string) Logger {
	md, _ := metadata.Metadata()
	labels := make(map[string]string, len(md.PodLabels))
	for k, v := range md.PodLabels {
		k = "k8s-pod/" + strings.ReplaceAll(k, ".", "_")
		labels[k] = v
	}
	return g.client.Logger(
		logID,
		logging.CommonResource(&mrpb.MonitoredResource{
			Type: "k8s_container",
			Labels: map[string]string{
				"pod_name":       md.PodName,
				"cluster_name":   md.ClusterName,
				"location":       md.Zone,
				"project_id":     md.ProjectID,
				"namespace_name": md.PodNamespace,
				"container_name": md.ContainerName,
			},
		}),
		logging.CommonLabels(labels),
	)
}

// Close waits for all opened loggers to be flushed and closes the client.
func (g GkeClient) Close() error {
	return g.client.Close()
}

// Ping reports whether the client's connection to the logging service and the
// authentication configuration are valid. To accomplish this, Ping writes a
// log entry "ping" to a log named "ping".
func (g GkeClient) Ping(ctx context.Context) error {
	return g.client.Ping(ctx)
}

// A GkeLogger is used to write log messages to a single log.
type GkeLogger struct {
	*logging.Logger
}

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
	"os"

	"github.com/ajjensen13/gke/internal/metadata"
)

func NewGkeClient(ctx context.Context, parent string) (GkeClient, error) {
	client, err := logging.NewClient(ctx, parent)
	if err != nil {
		return GkeClient{}, err
	}
	return GkeClient{client}, nil
}

type GkeClient struct {
	client *logging.Client
}

func (g GkeClient) Logger(logID string) Logger {
	md, _ := metadata.Metadata()
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
				// TODO "container_name": md.ContainerName,
			},
		}),
		logging.CommonLabels(md.PodLabels),
	)
}

func (g GkeClient) Close() error {
	return g.client.Close()
}

func (g GkeClient) Ping(ctx context.Context) error {
	return g.client.Ping(ctx)
}

type GkeLogger struct {
	*logging.Logger
}

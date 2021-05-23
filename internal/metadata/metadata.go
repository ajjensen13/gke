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

package metadata

import (
	"bufio"
	"bytes"
	"cloud.google.com/go/compute/metadata"
	"errors"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
	"sync"
)

// MetadataType is structured GCE metadata.
type MetadataType struct {
	// OnGCE comes from the metadata server
	OnGCE bool `json:""`
	// ProjectID comes from the metadata server
	ProjectID string `json:",omitempty"`
	// InstanceID comes from the metadata server
	InstanceID string `json:",omitempty"`
	// Hostname comes from the metadata server
	Hostname string `json:",omitempty"`
	// Zone comes from the metadata server
	Zone string `json:",omitempty"`
	// ClusterName comes from the metadata server
	ClusterName string `json:",omitempty"`
	// ClusterLocation comes from the metadata server
	ClusterLocation string `json:",omitempty"`
	// ClusterUid comes from the metadata server
	ClusterUid string
	// InstanceAttributes comes from the metadata server. It is not
	// included in JSON serialization to prevent sending sensitive data.
	InstanceAttributes map[string]string `json:"-"`
	// ProjectAttributes comes from the metadata server. It is not
	// included in JSON serialization to prevent sending sensitive data.
	ProjectAttributes map[string]string `json:"-"`
	// PodName comes from /etc/k8info/pod_name.
	// PodName can be provided by a Downward API volume.
	PodName string `json:",omitempty"`
	// PodNamespace comes from /etc/k8info/pod_namespace.
	// PodNamespace can be provided by a Downward API volume.
	PodNamespace string `json:",omitempty"`
	// PodLabels comes from /etc/k8info/pod_labels.
	// PodLabels can be provided by a Downward API volume.
	PodLabels map[string]string `json:",omitempty"`
	// ContainerName comes from /etc/k8info/container_name
	ContainerName string `json:",omitempty"`
}

var (
	pkgMetadata     MetadataType
	pkgMetadataErr  error
	pkgMetadataOnce sync.Once
)

// ErrNotOnGCE is returned when requesting metadata while not on GCE.
var ErrNotOnGCE = errors.New("not on GCE")

func initMetadata() {
	pkgMetadata.OnGCE = metadata.OnGCE()
	if !pkgMetadata.OnGCE {
		pkgMetadataErr = ErrNotOnGCE
		return
	}

	pid, err := metadata.ProjectID()
	if err != nil {
		pkgMetadataErr = fmt.Errorf("failed to initialize ProjectID: %w", err)
		return
	}
	pkgMetadata.ProjectID = pid

	iid, err := metadata.InstanceID()
	if err != nil {
		pkgMetadataErr = fmt.Errorf("failed to initialize InstanceID: %w", err)
		return
	}
	pkgMetadata.InstanceID = iid

	host, err := metadata.Hostname()
	if err != nil {
		pkgMetadataErr = fmt.Errorf("failed to initialize Hostname: %w", err)
		return
	}
	pkgMetadata.Hostname = host

	vals, err := metadata.InstanceAttributes()
	if err != nil {
		pkgMetadataErr = fmt.Errorf("failed to initialize InstanceAttributes: %w", err)
		return
	}
	pkgMetadata.InstanceAttributes = make(map[string]string, len(vals))

	for _, val := range vals {
		v, err := metadata.InstanceAttributeValue(val)
		if err != nil {
			pkgMetadataErr = fmt.Errorf("failed to initialize InstanceAttributesValue %s: %w", v, err)
			return
		}
		pkgMetadata.InstanceAttributes[val] = v
	}

	vals, _ = metadata.ProjectAttributes()
	pkgMetadata.ProjectAttributes = make(map[string]string, len(vals))

	for _, val := range vals {
		v, err := metadata.ProjectAttributeValue(val)
		if err != nil {
			pkgMetadataErr = fmt.Errorf("failed to initialize ProjectAttributeValue %s: %w", v, err)
			return
		}
		pkgMetadata.ProjectAttributes[val] = v
	}

	zone, _ := metadata.Zone()
	pkgMetadata.Zone = zone

	clusterLocation, err := metadata.Get("instance/attributes/cluster-location")
	if err != nil {
		return
	}
	pkgMetadata.ClusterLocation = clusterLocation

	clusterName, err := metadata.Get("instance/attributes/cluster-name")
	if err != nil {
		return
	}
	pkgMetadata.ClusterName = clusterName

	clusterUid, err := metadata.Get("instance/attributes/cluster-uid")
	if err != nil {
		return
	}
	pkgMetadata.ClusterUid = clusterUid

	// pkgMetadata.ClusterName, err = readK8InfoValue("cluster_name")
	// if err != nil {
	// 	pkgMetadataErr = fmt.Errorf("failed to initialize ClusterName: %w", err)
	// 	return
	// }

	pkgMetadata.PodName, err = readK8InfoValue("pod_name")
	if err != nil {
		pkgMetadataErr = fmt.Errorf("failed to initialize PodName: %w", err)
		return
	}

	pkgMetadata.PodNamespace, err = readK8InfoValue("pod_namespace")
	if err != nil {
		pkgMetadataErr = fmt.Errorf("failed to initialize PodNamespace: %w", err)
		return
	}

	pkgMetadata.PodLabels, err = readK8InfoValues("pod_labels")
	if err != nil {
		pkgMetadataErr = fmt.Errorf("failed to initialize PodLabels: %w", err)
		return
	}

	pkgMetadata.ContainerName, err = readK8InfoValue("container_name")
	if err != nil {
		pkgMetadataErr = fmt.Errorf("failed to initialize ContainerName: %w", err)
		return
	}
}

func readK8InfoValue(name string) (string, error) {
	p := filepath.Join("/etc/k8info", name)
	val, err := ioutil.ReadFile(p)
	if err != nil {
		return "", err
	}
	return string(val), nil
}

func readK8InfoValues(name string) (map[string]string, error) {
	v, err := readK8InfoValue(name)
	if err != nil {
		return nil, err
	}

	result := map[string]string{}
	for s := bufio.NewScanner(strings.NewReader(v)); s.Scan(); {
		t := s.Bytes()
		eq := bytes.IndexRune(t, '=')
		key := string(t[:eq])
		val := string(bytes.Trim(bytes.TrimSpace(t[eq+1:]), "\""))
		result[key] = val
	}

	return result, nil
}

// Metadata returns a cached instance of the GCE metadata. If not on GCE, Metadata() returns ErrNotOnGCE.
// The data comes from various sources including the GCE metadata server and K8 downward API volumes.
func Metadata() (md *MetadataType, err error) {
	pkgMetadataOnce.Do(initMetadata)
	return &pkgMetadata, pkgMetadataErr
}

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
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
	"sync"
)

type MetadataType struct {
	OnGCE              bool
	ProjectID          string
	InstanceID         string
	InstanceName       string
	ExternalIP         string
	InternalIP         string
	Hostname           string
	Zone               string
	InstanceTags       []string
	InstanceAttributes map[string]string
	ProjectAttributes  map[string]string
	PodName            string
	PodNamespace       string
	PodLabels          map[string]string
}

var (
	pkgMetadata     MetadataType
	pkgMetadataOnce sync.Once
)

func initMetadata() {
	pkgMetadata.OnGCE = metadata.OnGCE()
	if !pkgMetadata.OnGCE {
		return
	}

	pid, err := metadata.ProjectID()
	if err != nil {
		panic(err)
	}
	pkgMetadata.ProjectID = pid

	iid, err := metadata.InstanceID()
	if err != nil {
		panic(err)
	}
	pkgMetadata.InstanceID = iid

	instName, err := metadata.InstanceName()
	if err != nil {
		panic(err)
	}
	pkgMetadata.InstanceName = instName

	exIP, err := metadata.ExternalIP()
	if err != nil {
		panic(err)
	}
	pkgMetadata.ExternalIP = exIP

	inIP, err := metadata.InternalIP()
	if err != nil {
		panic(err)
	}
	pkgMetadata.InternalIP = inIP

	host, err := metadata.Hostname()
	if err != nil {
		panic(err)
	}
	pkgMetadata.Hostname = host

	vals, err := metadata.InstanceAttributes()
	if err != nil {
		panic(err)
	}
	pkgMetadata.InstanceAttributes = make(map[string]string, len(vals))

	for _, val := range vals {
		v, err := metadata.InstanceAttributeValue(val)
		if err != nil {
			panic(err)
		}
		pkgMetadata.InstanceAttributes[val] = v
	}

	instTags, err := metadata.InstanceTags()
	if err != nil {
		panic(err)
	}
	pkgMetadata.InstanceTags = instTags

	vals, err = metadata.ProjectAttributes()
	if err != nil {
		panic(err)
	}
	pkgMetadata.ProjectAttributes = make(map[string]string, len(vals))
	for _, val := range vals {
		v, err := metadata.ProjectAttributeValue(val)
		if err != nil {
			panic(err)
		}
		pkgMetadata.ProjectAttributes[val] = v
	}

	zone, err := metadata.Zone()
	if err != nil {
		panic(err)
	}
	pkgMetadata.Zone = zone

	pkgMetadata.PodName = readK8InfoValue("pod_name")
	pkgMetadata.PodNamespace = readK8InfoValue("pod_namespace")
	pkgMetadata.PodLabels = readK8InfoValues("pod_labels")
}

func readK8InfoValue(name string) string {
	p := filepath.Join("/etc", name)
	val, err := ioutil.ReadFile(p)
	if err != nil {
		panic(fmt.Errorf("error reading %s: %w", name, err))
	}
	return string(val)
}

func readK8InfoValues(name string) map[string]string {
	v := readK8InfoValue(name)

	result := map[string]string{}
	for s := bufio.NewScanner(strings.NewReader(v)); s.Scan(); {
		t := s.Bytes()
		eq := bytes.IndexRune(t, '=')
		key := string(t[:eq])
		val := string(bytes.Trim(bytes.TrimSpace(t[eq+1:]), "\""))
		result[key] = val
	}

	return result
}

func Metadata() (*MetadataType, bool) {
	pkgMetadataOnce.Do(initMetadata)
	if !pkgMetadata.OnGCE {
		return nil, false
	}
	return &pkgMetadata, true
}

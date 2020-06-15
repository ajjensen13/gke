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
	"os"

	"github.com/ajjensen13/gke/internal/metadata"
)

// ErrNotOnGCE is returned when requesting metadata while not on GCE.
var ErrNotOnGCE = metadata.ErrNotOnGCE

// MetadataType is structured GCE metadata.
type MetadataType = metadata.MetadataType

// Metadata returns a cached instance of the GCE metadata. If not on GCE, Metadata() returns ErrNotOnGCE.
// The data comes from various sources including the GCE metadata server and K8 downward API volumes.
func Metadata() (md *MetadataType, err error) {
	return metadata.Metadata()
}

// LogMetadata logs the metadata at Info severity.
// It is provided for consistency in logging across GKE applications.
func LogMetadata(lg Logger) {
	md, ok := Metadata()
	lg.Info(NewMsgData("gke.Metadata()", md, ok))
}

// LogEnv logs the environment at Info severity.
// It is provided for consistency in logging across GKE applications.
func LogEnv(lg Logger) {
	lg.Info(NewMsgData("os.Environ()", os.Environ()))
}

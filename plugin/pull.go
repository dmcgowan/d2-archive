package plugin

import "io"

// PullLayerPlugin pulls image layers based on id
// returning a stream of the image contents as a tar archive.
type PullLayerPlugin interface {
	PullLayer(id string) (io.ReadCloser, error)
}

package internal

import "io"

type ProjectImage struct {
	Name  string
	Image io.ReadCloser
}

type ProjectManager interface {
	AddImage(scanner string, image []byte) (*ProjectImage, error)
	Get(scanner string) ([]ProjectImage, error)
}

type Exporter interface {
	Export(images []ProjectImage) (string, io.ReadCloser, error)
}

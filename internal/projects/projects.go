package projects

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/tariel-x/scan/internal"
)

type Projects struct {
	imageStorage string
}

func NewProjects(imageStorage string) (*Projects, error) {
	return &Projects{
		imageStorage: imageStorage,
	}, nil
}

func (p *Projects) AddImage(name string, image []byte) (*internal.ProjectImage, error) {
	scanner := p.encodeScannerName(name)
	imageNumber := 0

	existingImages, err := p.listImages(scanner)
	if err != nil {
		return nil, fmt.Errorf("can not list existing images: %w", err)
	}
	if len(existingImages) > 0 {
		imageNumber = len(existingImages)
	}

	filename := p.getFileName(scanner, imageNumber, time.Now())
	if err := os.WriteFile(filename, image, 0666); err != nil {
		return nil, fmt.Errorf("can not write image to storage: %w", err)
	}
	return &internal.ProjectImage{
		Name:  filename,
		Image: io.NopCloser(bytes.NewBuffer(image)),
	}, nil
}

func (p *Projects) Get(name string) ([]internal.ProjectImage, error) {
	scanner := p.encodeScannerName(name)

	filenames, err := p.listImages(scanner)
	if err != nil {
		return nil, fmt.Errorf("can not list existing images: %w", err)
	}

	images := make([]internal.ProjectImage, 0, len(filenames))
	for _, filename := range filenames {
		f, err := p.openImage(filename)
		if err != nil {
			return nil, fmt.Errorf("can not open image file: %w", err)
		}
		images = append(images, internal.ProjectImage{
			Name:  filename,
			Image: f,
		})
	}
	return images, nil
}

func (p *Projects) openImage(filename string) (io.ReadCloser, error) {
	return os.Open(path.Join(p.imageStorage, filename))
}

func (p *Projects) listImages(scanner string) ([]string, error) {
	entries, err := os.ReadDir(p.imageStorage)
	if err != nil {
		return nil, fmt.Errorf("can not list project images: %w", err)
	}
	imageNames := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		imageScanner, _, _, err := p.extractNameParts(entry.Name())
		if err != nil {
			continue
		}
		if imageScanner == scanner {
			imageNames = append(imageNames, entry.Name())
		}
	}
	return imageNames, nil
}

func (p *Projects) encodeScannerName(scanner string) string {
	scanner = strings.Replace(scanner, "/", "_", -1)
	scanner = strings.Replace(scanner, ";;;", "_", -1)
	return strings.Replace(scanner, " ", "_", -1)
}

func (p *Projects) getStoragePath(scanner string) string {
	return path.Join(p.imageStorage, scanner)
}

func (p *Projects) getFileName(scanner string, number int, createdAt time.Time) string {
	return fmt.Sprintf("img;;;%s;;;%d;;;%d", scanner, number, createdAt.Unix())
}

func (p *Projects) extractNameParts(filename string) (string, int, time.Time, error) {
	parts := strings.Split(filename, ";;;")
	if len(parts) != 4 {
		return "", 0, time.Time{}, fmt.Errorf("broken filename %q", filename)
	}
	scanner := parts[1]
	number, err := strconv.Atoi(parts[2])
	if err != nil {
		return "", 0, time.Time{}, fmt.Errorf("broken number: %w", err)
	}
	createdAtUnix, err := strconv.Atoi(parts[3])
	if err != nil {
		return "", 0, time.Time{}, fmt.Errorf("broken created at: %w", err)
	}
	createdAt := time.Unix(int64(createdAtUnix), 0)

	return scanner, number, createdAt, nil
}

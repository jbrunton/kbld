package config

import (
	"fmt"

	"github.com/ghodss/yaml"
	ctlres "github.com/k14s/kbld/pkg/kbld/resources"
)

const (
	configAPIVersion      = "kbld.k14s.io/v1alpha1"
	sourcesKind           = "Sources"           // specify list of sources for building images
	imageOverridesKind    = "ImageOverrides"    // specify alternative image urls
	imageDestinationsKind = "ImageDestinations" // specify image push destinations
	imageKeysKind         = "ImageKeys"
)

var (
	configKinds = []string{sourcesKind, imageOverridesKind, imageDestinationsKind, imageKeysKind}
)

type Config struct {
	APIVersion string `json:"apiVersion"`
	Kind       string

	Sources      []Source
	Overrides    []ImageOverride
	Destinations []ImageDestination
	Keys         []string
}

type Source struct {
	ImageRef
	Path string

	Docker *SourceDockerOpts
	Pack   *SourcePackOpts
}

type ImageOverride struct {
	ImageRef
	NewImage    string `json:"newImage"`
	Preresolved bool
}

type ImageDestination struct {
	ImageRef
	NewImage string `json:"newImage"`
}

type ImageRef struct {
	Image     string
	ImageRepo string `json:"imageRepo"`
}

func NewConfigFromResource(res ctlres.Resource) (Config, error) {
	bs, err := res.AsYAMLBytes()
	if err != nil {
		return Config{}, err
	}

	var config Config

	err = yaml.Unmarshal(bs, &config)
	if err != nil {
		return Config{}, fmt.Errorf("Unmarshaling %s: %s", res.Description(), err)
	}

	err = config.Validate()
	if err != nil {
		return Config{}, fmt.Errorf("Validating %s: %s", res.Description(), err)
	}

	for i, imageDst := range config.Destinations {
		if len(imageDst.NewImage) == 0 {
			imageDst.NewImage = imageDst.Image
		}
		config.Destinations[i] = imageDst
	}

	return config, nil
}

func (d Config) Validate() error {
	for i, src := range d.Sources {
		err := src.Validate()
		if err != nil {
			return fmt.Errorf("Validating Sources[%d]: %s", i, err)
		}
	}

	for i, override := range d.Overrides {
		err := override.Validate()
		if err != nil {
			return fmt.Errorf("Validating Overrides[%d]: %s", i, err)
		}
	}

	for i, dst := range d.Destinations {
		err := dst.Validate()
		if err != nil {
			return fmt.Errorf("Validating Destinations[%d]: %s", i, err)
		}
	}

	for i, key := range d.Keys {
		if len(key) == 0 {
			return fmt.Errorf("Validating Destinations[%d]: Expected to be non-empty", i)
		}
	}

	return nil
}

func (d Source) Validate() error {
	err := d.ImageRef.Validate()
	if err != nil {
		return err
	}
	if len(d.Path) == 0 {
		return fmt.Errorf("Expected Path to be non-empty")
	}
	return nil
}

func (d ImageOverride) Validate() error {
	err := d.ImageRef.Validate()
	if err != nil {
		return err
	}
	if len(d.NewImage) == 0 {
		return fmt.Errorf("Expected NewImage to be non-empty")
	}
	return nil
}

func (d ImageDestination) Validate() error {
	return d.ImageRef.Validate()
}

func (r ImageRef) Validate() error {
	if len(r.Image) == 0 && len(r.ImageRepo) == 0 {
		return fmt.Errorf("Expected Image or ImageRepo to be non-empty")
	}
	return nil
}

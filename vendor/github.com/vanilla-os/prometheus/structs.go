package prometheus

/*	License: GPLv3
	Authors:
		Mirko Brombin <mirko@fabricators.ltd>
		Vanilla OS Contributors <https://github.com/vanilla-os/>
	Copyright: 2023
	Description:
		Prometheus is a simple and accessible library for pulling and mounting
		container images. It is designed to be used as a dependency in ABRoot
		and Albius.
*/

import cstorage "github.com/containers/storage"

type Prometheus struct {
	Store  cstorage.Store
	Config PrometheusConfig
}

type OciManifest struct {
	SchemaVersion int                 `json:"schemaVersion"`
	MediaType     string              `json:"mediaType"`
	Config        OciManifestConfig   `json:"config"`
	Layers        []OciManifestConfig `json:"layers"`
}

type OciManifestConfig struct {
	MediaType string `json:"mediaType"`
	Size      int    `json:"size"`
	Digest    string `json:"digest"`
}

type PrometheusConfig struct {
	Root                 string
	GraphDriverName      string
	MaxParallelDownloads uint
}

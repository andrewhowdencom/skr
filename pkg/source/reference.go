package source

import (
	"strings"
)

const (
	SchemaOCI  = "oci"
	SchemaFile = "file"
	SchemaGit  = "git"
	SchemaHTTP = "http"
)

// Reference represents a parsed skill source reference.
type Reference struct {
	Original string
	Schema   string
	Path     string
	Spec     string
}

// ParseReference parses a raw reference string into its Schema, Path, and Spec components.
func ParseReference(raw string) Reference {
	ref := Reference{Original: raw}
	workingRaw := raw

	// 1. Detect Schema
	if strings.HasPrefix(workingRaw, "file://") {
		ref.Schema = SchemaFile
		workingRaw = strings.TrimPrefix(workingRaw, "file://")
	} else if strings.HasPrefix(workingRaw, "git://") || strings.HasPrefix(workingRaw, "git@") {
		ref.Schema = SchemaGit
	} else if strings.HasPrefix(workingRaw, "git+https://") || strings.HasPrefix(workingRaw, "git+http://") || strings.HasPrefix(workingRaw, "git+ssh://") {
		ref.Schema = SchemaGit
		workingRaw = strings.TrimPrefix(workingRaw, "git+")
	} else if strings.HasPrefix(workingRaw, "https://") || strings.HasPrefix(workingRaw, "http://") {
		ref.Schema = SchemaHTTP
	} else if strings.HasPrefix(workingRaw, "oci://") {
		ref.Schema = SchemaOCI
		workingRaw = strings.TrimPrefix(workingRaw, "oci://")
	} else {
		// Rule: in the absence of a schema, it's OCI.
		ref.Schema = SchemaOCI
	}

	// 2. Extract Spec
	switch ref.Schema {
	case SchemaOCI:
		// OCI uses ':' for tags
		idx := strings.LastIndex(workingRaw, ":")
		if idx != -1 && !strings.Contains(workingRaw[idx:], "/") {
			ref.Path = workingRaw[:idx]
			ref.Spec = workingRaw[idx+1:]
		} else {
			ref.Path = workingRaw
			ref.Spec = "latest"
		}
	case SchemaGit:
		// Git references use '#' for refspecs
		idx := strings.LastIndex(workingRaw, "#")
		if idx != -1 {
			ref.Path = workingRaw[:idx]
			ref.Spec = workingRaw[idx+1:]
		} else {
			ref.Path = workingRaw
			// For git, empty implies default branch behavior
			ref.Spec = ""
		}
	case SchemaFile:
		ref.Path = workingRaw
		ref.Spec = "latest"
	case SchemaHTTP:
		ref.Path = workingRaw
		ref.Spec = "latest"
	}

	return ref
}

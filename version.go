package main

import "fmt"

const (
	MajorVersion  = 0     //MajorVersion when you make incompatible changes .
	MinorVersion  = 0     //MinorVersion when you add functionality in a backwards-compatible manner .
	PatchVersion  = 1     //PatchVersion  when you fix bugs .
	VersionSuffix = "dev" // -alpha -alpha.1 -beta -rc -rc.1
)

// Version returns the version of server.
func Version() string {
	return fmt.Sprintf("v%d.%d.%d-%s", MajorVersion, MinorVersion, PatchVersion, VersionSuffix)
}

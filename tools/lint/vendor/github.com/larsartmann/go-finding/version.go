package finding

import "fmt"

// VersionMajor is the major version number.
const VersionMajor = 1

// VersionMinor is the minor version number.
const VersionMinor = 15

// VersionPatch is the patch version number.
const VersionPatch = 0

// Version is the semantic version string, computed from components.
var Version = fmt.Sprintf("%d.%d.%d", VersionMajor, VersionMinor, VersionPatch)

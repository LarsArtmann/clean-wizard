package domain

import (
	"gopkg.in/yaml.v3"
)

// CacheCleanupMode represents cache cleanup behavior as a type-safe enum.
type CacheCleanupMode int

const (
	// CacheCleanupDisabled represents disabled cache cleanup.
	CacheCleanupDisabled CacheCleanupMode = iota
	// CacheCleanupEnabled represents enabled cache cleanup.
	CacheCleanupEnabled
)

// String constants for cache cleanup mode representation.
const (
	stringDisabled = "DISABLED"
	stringEnabled  = "ENABLED"
	stringUnknown  = "UNKNOWN"
)

// String returns string representation of cache cleanup mode.
func (cm CacheCleanupMode) String() string {
	switch cm {
	case CacheCleanupDisabled:
		return stringDisabled
	case CacheCleanupEnabled:
		return stringEnabled
	default:
		return stringUnknown
	}
}

// IsValid checks if cache cleanup mode is valid.
func (cm CacheCleanupMode) IsValid() bool {
	return cm >= CacheCleanupDisabled && cm <= CacheCleanupEnabled
}

// Values returns all possible cache cleanup modes.
func (cm CacheCleanupMode) Values() []CacheCleanupMode {
	return []CacheCleanupMode{
		CacheCleanupDisabled,
		CacheCleanupEnabled,
	}
}

// IsEnabled checks if cache cleanup is enabled.
func (cm CacheCleanupMode) IsEnabled() bool {
	return cm == CacheCleanupEnabled
}

// MarshalYAML implements yaml.Marshaler interface for CacheCleanupMode.
func (cm CacheCleanupMode) MarshalYAML() (any, error) {
	return int(cm), nil
}

// UnmarshalYAML implements yaml.Unmarshaler interface for CacheCleanupMode.
// Accepts both string and integer representations.
func (cm *CacheCleanupMode) UnmarshalYAML(value *yaml.Node) error {
	return UnmarshalYAMLEnum(value, cm, map[string]CacheCleanupMode{
		"DISABLED": CacheCleanupDisabled,
		"ENABLED":  CacheCleanupEnabled,
	}, "invalid cache cleanup mode")
}

// DockerPruneMode represents Docker prune behavior as a type-safe enum.
type DockerPruneMode int

const (
	// DockerPruneAll represents pruning all resources.
	DockerPruneAll DockerPruneMode = iota
	// DockerPruneImages represents pruning only images.
	DockerPruneImages
	// DockerPruneContainers represents pruning only containers.
	DockerPruneContainers
	// DockerPruneVolumes represents pruning only volumes.
	DockerPruneVolumes
	// DockerPruneBuilds represents pruning only build cache.
	DockerPruneBuilds
)

// String returns string representation of Docker prune mode.
func (pm DockerPruneMode) String() string {
	switch pm {
	case DockerPruneAll:
		return "ALL"
	case DockerPruneImages:
		return "IMAGES"
	case DockerPruneContainers:
		return "CONTAINERS"
	case DockerPruneVolumes:
		return "VOLUMES"
	case DockerPruneBuilds:
		return "BUILDS"
	default:
		return stringUnknown
	}
}

// IsValid checks if Docker prune mode is valid.
func (pm DockerPruneMode) IsValid() bool {
	return pm >= DockerPruneAll && pm <= DockerPruneBuilds
}

// Values returns all possible Docker prune modes.
func (pm DockerPruneMode) Values() []DockerPruneMode {
	return []DockerPruneMode{
		DockerPruneAll,
		DockerPruneImages,
		DockerPruneContainers,
		DockerPruneVolumes,
		DockerPruneBuilds,
	}
}

// MarshalYAML implements yaml.Marshaler interface for DockerPruneMode.
func (pm DockerPruneMode) MarshalYAML() (any, error) {
	return int(pm), nil
}

// UnmarshalYAML implements yaml.Unmarshaler interface for DockerPruneMode.
func (pm *DockerPruneMode) UnmarshalYAML(value *yaml.Node) error {
	return UnmarshalYAMLEnum(value, pm, map[string]DockerPruneMode{
		"ALL":        DockerPruneAll,
		"IMAGES":     DockerPruneImages,
		"CONTAINERS": DockerPruneContainers,
		"VOLUMES":    DockerPruneVolumes,
		"BUILDS":     DockerPruneBuilds,
	}, "invalid docker prune mode")
}

// BuildToolType represents build tool types as a type-safe enum.
type BuildToolType int

const (
	// BuildToolGo represents Go build tools.
	BuildToolGo BuildToolType = iota
	// BuildToolRust represents Rust/Cargo build tools.
	BuildToolRust
	// BuildToolNode represents Node.js build tools.
	BuildToolNode
	// BuildToolPython represents Python build tools.
	BuildToolPython
	// BuildToolJava represents Java build tools (Maven, Gradle).
	BuildToolJava
	// BuildToolScala represents Scala build tools (SBT).
	BuildToolScala
)

// String returns string representation of build tool type.
func (bt BuildToolType) String() string {
	switch bt {
	case BuildToolGo:
		return "GO"
	case BuildToolRust:
		return "RUST"
	case BuildToolNode:
		return "NODE"
	case BuildToolPython:
		return "PYTHON"
	case BuildToolJava:
		return "JAVA"
	case BuildToolScala:
		return "SCALA"
	default:
		return stringUnknown
	}
}

// IsValid checks if build tool type is valid.
func (bt BuildToolType) IsValid() bool {
	return bt >= BuildToolGo && bt <= BuildToolScala
}

// Values returns all possible build tool types.
func (bt BuildToolType) Values() []BuildToolType {
	return []BuildToolType{
		BuildToolGo,
		BuildToolRust,
		BuildToolNode,
		BuildToolPython,
		BuildToolJava,
		BuildToolScala,
	}
}

// MarshalYAML implements yaml.Marshaler interface for BuildToolType.
func (bt BuildToolType) MarshalYAML() (any, error) {
	return int(bt), nil
}

// UnmarshalYAML implements yaml.Unmarshaler interface for BuildToolType.
func (bt *BuildToolType) UnmarshalYAML(value *yaml.Node) error {
	return UnmarshalYAMLEnum(value, bt, map[string]BuildToolType{
		"GO":     BuildToolGo,
		"RUST":   BuildToolRust,
		"NODE":   BuildToolNode,
		"PYTHON": BuildToolPython,
		"JAVA":   BuildToolJava,
		"SCALA":  BuildToolScala,
	}, "invalid build tool type")
}

// CacheType represents system cache types as a type-safe enum.
type CacheType int

const (
	// CacheTypeSpotlight represents macOS Spotlight cache.
	CacheTypeSpotlight CacheType = iota
	// CacheTypeXcode represents Xcode derived data cache.
	CacheTypeXcode
	// CacheTypeCocoapods represents CocoaPods cache.
	CacheTypeCocoapods
	// CacheTypeHomebrew represents Homebrew cache.
	CacheTypeHomebrew
	// CacheTypePip represents Python pip cache.
	CacheTypePip
	// CacheTypeNpm represents Node.js npm cache.
	CacheTypeNpm
	// CacheTypeYarn represents Yarn cache.
	CacheTypeYarn
	// CacheTypeCcache represents ccache.
	CacheTypeCcache
	// CacheTypeXdgCache represents Linux XDG cache directory (~/.cache).
	CacheTypeXdgCache
	// CacheTypeThumbnails represents Linux thumbnail cache.
	CacheTypeThumbnails
	// CacheTypePuppeteer represents Puppeteer browser cache.
	CacheTypePuppeteer
	// CacheTypeTerraform represents Terraform plugin cache.
	CacheTypeTerraform
	// CacheTypeGradleWrapper represents Gradle wrapper distributions cache.
	CacheTypeGradleWrapper
	// CacheTypeKonan represents Kotlin/Native toolchain dependencies.
	CacheTypeKonan
	// CacheTypeRustup represents Rust toolchain cache.
	CacheTypeRustup
)

// String returns string representation of cache type.
func (ct CacheType) String() string {
	switch ct {
	case CacheTypeSpotlight:
		return "SPOTLIGHT"
	case CacheTypeXcode:
		return "XCODE"
	case CacheTypeCocoapods:
		return "COCOAPODS"
	case CacheTypeHomebrew:
		return "HOMEBREW"
	case CacheTypePip:
		return "PIP"
	case CacheTypeNpm:
		return "NPM"
	case CacheTypeYarn:
		return "YARN"
	case CacheTypeCcache:
		return "CCACHE"
	case CacheTypeXdgCache:
		return "XDG_CACHE"
	case CacheTypeThumbnails:
		return "THUMBNAILS"
	case CacheTypePuppeteer:
		return "PUPPETEER"
	case CacheTypeTerraform:
		return "TERRAFORM"
	case CacheTypeGradleWrapper:
		return "GRADLE_WRAPPER"
	case CacheTypeKonan:
		return "KONAN"
	case CacheTypeRustup:
		return "RUSTUP"
	default:
		return stringUnknown
	}
}

// IsValid checks if cache type is valid.
func (ct CacheType) IsValid() bool {
	return ct >= CacheTypeSpotlight && ct <= CacheTypeRustup
}

// Values returns all possible cache types.
func (ct CacheType) Values() []CacheType {
	return []CacheType{
		CacheTypeSpotlight,
		CacheTypeXcode,
		CacheTypeCocoapods,
		CacheTypeHomebrew,
		CacheTypePip,
		CacheTypeNpm,
		CacheTypeYarn,
		CacheTypeCcache,
		CacheTypeXdgCache,
		CacheTypeThumbnails,
		CacheTypePuppeteer,
		CacheTypeTerraform,
		CacheTypeGradleWrapper,
		CacheTypeKonan,
		CacheTypeRustup,
	}
}

// MarshalYAML implements yaml.Marshaler interface for CacheType.
func (ct CacheType) MarshalYAML() (any, error) {
	return int(ct), nil
}

// UnmarshalYAML implements yaml.Unmarshaler interface for CacheType.
func (ct *CacheType) UnmarshalYAML(value *yaml.Node) error {
	return UnmarshalYAMLEnum(value, ct, map[string]CacheType{
		"SPOTLIGHT":      CacheTypeSpotlight,
		"XCODE":          CacheTypeXcode,
		"COCOAPODS":      CacheTypeCocoapods,
		"HOMEBREW":       CacheTypeHomebrew,
		"PIP":            CacheTypePip,
		"NPM":            CacheTypeNpm,
		"YARN":           CacheTypeYarn,
		"CCACHE":         CacheTypeCcache,
		"XDG_CACHE":      CacheTypeXdgCache,
		"THUMBNAILS":     CacheTypeThumbnails,
		"PUPPETEER":      CacheTypePuppeteer,
		"TERRAFORM":      CacheTypeTerraform,
		"GRADLE_WRAPPER": CacheTypeGradleWrapper,
		"KONAN":          CacheTypeKonan,
		"RUSTUP":         CacheTypeRustup,
	}, "invalid cache type")
}

// PackageManagerType represents Node.js package manager types as a type-safe enum.
type PackageManagerType int

const (
	// PackageManagerNpm represents npm.
	PackageManagerNpm PackageManagerType = iota
	// PackageManagerPnpm represents pnpm.
	PackageManagerPnpm
	// PackageManagerYarn represents Yarn.
	PackageManagerYarn
	// PackageManagerBun represents Bun.
	PackageManagerBun
)

// String returns string representation of package manager type.
func (pm PackageManagerType) String() string {
	switch pm {
	case PackageManagerNpm:
		return "NPM"
	case PackageManagerPnpm:
		return "PNPM"
	case PackageManagerYarn:
		return "YARN"
	case PackageManagerBun:
		return "BUN"
	default:
		return stringUnknown
	}
}

// IsValid checks if package manager type is valid.
func (pm PackageManagerType) IsValid() bool {
	return pm >= PackageManagerNpm && pm <= PackageManagerBun
}

// Values returns all possible package manager types.
func (pm PackageManagerType) Values() []PackageManagerType {
	return []PackageManagerType{
		PackageManagerNpm,
		PackageManagerPnpm,
		PackageManagerYarn,
		PackageManagerBun,
	}
}

// MarshalYAML implements yaml.Marshaler interface for PackageManagerType.
func (pm PackageManagerType) MarshalYAML() (any, error) {
	return int(pm), nil
}

// UnmarshalYAML implements yaml.Unmarshaler interface for PackageManagerType.
func (pm *PackageManagerType) UnmarshalYAML(value *yaml.Node) error {
	return UnmarshalYAMLEnum(value, pm, map[string]PackageManagerType{
		"NPM":  PackageManagerNpm,
		"PNPM": PackageManagerPnpm,
		"YARN": PackageManagerYarn,
		"BUN":  PackageManagerBun,
	}, "invalid package manager type")
}

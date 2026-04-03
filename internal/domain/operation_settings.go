package domain

import "gopkg.in/yaml.v3"

// CacheCleanupMode represents cache cleanup behavior as a type-safe enum.
type CacheCleanupMode int

const (
	// CacheCleanupDisabled represents disabled cache cleanup.
	CacheCleanupDisabled CacheCleanupMode = iota
	// CacheCleanupEnabled represents enabled cache cleanup.
	CacheCleanupEnabled
)

var cacheCleanupModeStrings = []string{"DISABLED", "ENABLED"}

func (cm CacheCleanupMode) String() string { return EnumString(cm, cacheCleanupModeStrings) }
func (cm CacheCleanupMode) IsValid() bool  { return EnumIsValid(cm, CacheCleanupEnabled) }
func (cm CacheCleanupMode) Values() []CacheCleanupMode {
	return EnumValues[CacheCleanupMode](CacheCleanupEnabled)
}
func (cm CacheCleanupMode) IsEnabled() bool { return cm == CacheCleanupEnabled }
func (cm CacheCleanupMode) MarshalYAML() (any, error) {
	return EnumMarshalYAML(cm, cacheCleanupModeStrings)
}

func (cm *CacheCleanupMode) UnmarshalYAML(value *yaml.Node) error {
	return EnumUnmarshalYAML(value, (*int)(cm), cacheCleanupModeStrings, "cache cleanup mode")
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

var dockerPruneModeStrings = []string{"ALL", "IMAGES", "CONTAINERS", "VOLUMES", "BUILDS"}

func (pm DockerPruneMode) String() string { return EnumString(pm, dockerPruneModeStrings) }
func (pm DockerPruneMode) IsValid() bool  { return EnumIsValid(pm, DockerPruneBuilds) }
func (pm DockerPruneMode) Values() []DockerPruneMode {
	return EnumValues[DockerPruneMode](DockerPruneBuilds)
}

func (pm DockerPruneMode) MarshalYAML() (any, error) {
	return EnumMarshalYAML(pm, dockerPruneModeStrings)
}

func (pm *DockerPruneMode) UnmarshalYAML(value *yaml.Node) error {
	return EnumUnmarshalYAML(value, (*int)(pm), dockerPruneModeStrings, "docker prune mode")
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

var buildToolTypeStrings = []string{"GO", "RUST", "NODE", "PYTHON", "JAVA", "SCALA"}

func (bt BuildToolType) String() string { return EnumString(bt, buildToolTypeStrings) }
func (bt BuildToolType) IsValid() bool  { return EnumIsValid(bt, BuildToolScala) }

func (bt BuildToolType) Values() []BuildToolType { return EnumValues[BuildToolType](BuildToolScala) }

func (bt BuildToolType) MarshalYAML() (any, error) { return EnumMarshalYAML(bt, buildToolTypeStrings) }
func (bt *BuildToolType) UnmarshalYAML(value *yaml.Node) error {
	return EnumUnmarshalYAML(value, (*int)(bt), buildToolTypeStrings, "build tool type")
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

var cacheTypeStrings = []string{
	"SPOTLIGHT", "XCODE", "COCOAPODS", "HOMEBREW", "PIP",
	"NPM", "YARN", "CCACHE", "XDG_CACHE", "THUMBNAILS",
	"PUPPETEER", "TERRAFORM", "GRADLE_WRAPPER", "KONAN", "RUSTUP",
}

func (ct CacheType) String() string            { return EnumString(ct, cacheTypeStrings) }
func (ct CacheType) IsValid() bool             { return EnumIsValid(ct, CacheTypeRustup) }
func (ct CacheType) Values() []CacheType       { return EnumValues[CacheType](CacheTypeRustup) }
func (ct CacheType) MarshalYAML() (any, error) { return EnumMarshalYAML(ct, cacheTypeStrings) }
func (ct *CacheType) UnmarshalYAML(value *yaml.Node) error {
	return EnumUnmarshalYAML(value, (*int)(ct), cacheTypeStrings, "cache type")
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

var packageManagerTypeStrings = []string{"NPM", "PNPM", "YARN", "BUN"}

func (pm PackageManagerType) String() string { return EnumString(pm, packageManagerTypeStrings) }
func (pm PackageManagerType) IsValid() bool  { return EnumIsValid(pm, PackageManagerBun) }
func (pm PackageManagerType) Values() []PackageManagerType {
	return EnumValues[PackageManagerType](PackageManagerBun)
}

func (pm PackageManagerType) MarshalYAML() (any, error) {
	return EnumMarshalYAML(pm, packageManagerTypeStrings)
}

func (pm *PackageManagerType) UnmarshalYAML(value *yaml.Node) error {
	return EnumUnmarshalYAML(value, (*int)(pm), packageManagerTypeStrings, "package manager type")
}

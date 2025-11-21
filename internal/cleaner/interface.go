package cleaner

import (
	"github.com/LarsArtmann/clean-wizard/internal/domain"
)

// Use domain.Cleaner interface directly
type CleanerInterface = domain.Cleaner

// Ensure all cleaners implement interface at compile time
var (
	_ CleanerInterface = (*NixCleaner)(nil)
	_ CleanerInterface = (*HomebrewCleaner)(nil)
	_ CleanerInterface = (*NpmCleaner)(nil)
	_ CleanerInterface = (*PnpmCleaner)(nil)
	_ CleanerInterface = (*TempFileCleaner)(nil)
)
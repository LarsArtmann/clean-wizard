package di

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/LarsArtmann/clean-wizard/internal/api"
	"github.com/LarsArtmann/clean-wizard/internal/config"
	"github.com/LarsArtmann/clean-wizard/internal/domain"
)

// ConfigAPIService handles API operations for configuration
// Replaces manual HTTP handling with proper service architecture
type ConfigAPIService struct {
	configService     *config.Service
	validationService *domain.ValidationService
	mapper           *api.Mapper
}

// NewConfigAPIService creates a new config API service
func NewConfigAPIService(configService *config.Service, validationService *domain.ValidationService) *ConfigAPIService {
	return &ConfigAPIService{
		configService:     configService,
		validationService: validationService,
		mapper:           api.NewMapper(),
	}
}

// GetConfig retrieves current configuration
// Returns API-serializable configuration
func (s *ConfigAPIService) GetConfig(ctx context.Context) (*api.PublicConfig, error) {
	// Get domain configuration
	domainConfig, err := s.configService.GetConfig(ctx)
	if err != nil {
		return nil, err
	}
	
	// Validate configuration
	if err := s.validationService.ValidateConfig(ctx, domainConfig); err != nil {
		return nil, err
	}
	
	// Convert to API format
	publicConfig, err := s.mapper.MapConfigToPublic(domainConfig)
	if err != nil {
		return nil, err
	}
	
	return publicConfig.Unwrap(), nil
}

// UpdateConfig updates configuration with validation
// Accepts API format and converts to domain model
func (s *ConfigAPIService) UpdateConfig(ctx context.Context, publicConfig *api.PublicConfig) (*api.PublicConfig, error) {
	// Convert API config to domain model
	domainConfig, err := s.mapper.MapConfigToDomain(publicConfig)
	if err != nil {
		return nil, err
	}
	
	// Validate updated configuration
	if err := s.validationService.ValidateConfig(ctx, domainConfig.Unwrap()); err != nil {
		return nil, err
	}
	
	// Update configuration
	updatedConfig, err := s.configService.UpdateConfig(ctx, domainConfig.Unwrap())
	if err != nil {
		return nil, err
	}
	
	// Convert back to API format
	returnPublicConfig, err := s.mapper.MapConfigToPublic(updatedConfig)
	if err != nil {
		return nil, err
	}
	
	return returnPublicConfig.Unwrap(), nil
}

// ScanConfig performs configuration scanning
// Returns scan results in API format
func (s *ConfigAPIService) ScanConfig(ctx context.Context, publicConfig *api.PublicConfig) (*api.PublicScanResult, error) {
	// Convert API config to domain model
	domainConfig, err := s.mapper.MapConfigToDomain(publicConfig)
	if err != nil {
		return nil, err
	}
	
	// Validate configuration
	if err := s.validationService.ValidateConfig(ctx, domainConfig.Unwrap()); err != nil {
		return nil, err
	}
	
	// Perform scan
	scanResult, err := s.configService.ScanConfig(ctx, domainConfig.Unwrap())
	if err != nil {
		return nil, err
	}
	
	// Convert to API format
	publicScanResult := s.mapper.MapScanResultToPublic(scanResult)
	return publicScanResult.Unwrap(), nil
}

// WriteJSONResponse is a utility for HTTP response writing
// Will be replaced by gin middleware in the future
func (s *ConfigAPIService) WriteJSONResponse(w http.ResponseWriter, status int, data interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(data)
}

// WriteErrorResponse is a utility for error response writing
// Will be replaced by gin middleware in the future
func (s *ConfigAPIService) WriteErrorResponse(w http.ResponseWriter, status int, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	
	errorResponse := map[string]interface{}{
		"error": err.Error(),
		"code":  status,
	}
	
	json.NewEncoder(w).Encode(errorResponse)
}
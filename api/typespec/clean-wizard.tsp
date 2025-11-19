// === CLEAN WIZARD PUBLIC API SPECIFICATION ===
// Defines external contracts for all client integrations

import "@typespec/http";
import "@typespec/rest";

// Core configuration types for external clients
model PublicConfig {
  version: string;
  safeMode: boolean;
  maxDiskUsage: int32;
  protectedPaths: string[];
  profiles: Record<PublicProfile>;
}

model PublicProfile {
  name: string;
  description: string;
  enabled: boolean;
  operations: PublicOperation[];
}

model PublicOperation {
  name: string;
  description: string;
  riskLevel: RiskLevel;
  enabled: boolean;
  settings: OperationSettings;
}

model OperationSettings {
  dryRun: boolean;
  verbose: boolean;
  timeoutSeconds: int32;
  confirmBeforeDelete: boolean;
}

// Result types for API responses
model PublicCleanResult {
  success: boolean;
  freedBytes: uint64;
  itemsRemoved: uint32;
  itemsFailed: uint32;
  cleanTime: string; // ISO 8601 duration
  cleanedAt: string;  // ISO 8601 timestamp
  strategy: CleanStrategy;
  errors?: string[];
}

model PublicScanResult {
  success: boolean;
  scannedItems: uint64;
  estimatedFreedBytes: uint64;
  operationResults: PublicOperationResult[];
  errors?: string[];
}

model PublicOperationResult {
  operationName: string;
  success: boolean;
  estimatedImpact: uint64;
  warnings?: string[];
}

// Enum types with compile-time safety
enum RiskLevel {
  LOW: "LOW";
  MEDIUM: "MEDIUM"; 
  HIGH: "HIGH";
  CRITICAL: "CRITICAL";
}

enum CleanStrategy {
  AGGRESSIVE: "aggressive";
  CONSERVATIVE: "conservative";
  DRY_RUN: "dry-run";
}

enum OperationType {
  NIX_GENERATIONS: "nix-generations";
  TEMP_FILES: "temp-files";
  LOG_FILES: "log-files";
  CACHE_FILES: "cache-files";
}

// API Routes
@route("/api/v1")
namespace CleanWizardAPI {
  @get
  @route("/config")
  op getConfig(): PublicConfig;

  @post
  @route("/config")
  op updateConfig(@body config: PublicConfig): PublicConfig;

  @post  
  @route("/scan")
  op scanConfig(@body config: PublicConfig): PublicScanResult;

  @post
  @route("/clean")
  op cleanSystem(@body request: CleanRequest): PublicCleanResult;
}

model CleanRequest {
  config: PublicConfig;
  strategy: CleanStrategy;
  operations: OperationType[]; // operations to execute
  dryRun?: boolean;
}
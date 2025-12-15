# TypeSpec vs Go Types Strategic Decision

## EXECUTIVE SUMMARY

**DECISION**: Hybrid approach - TypeSpec for public APIs, Go for internal domain models

## RATIONALE

### 1. EXTERNAL CONTRACT STABILITY

- Public APIs need stable contracts across languages
- TypeSpec provides automatic schema validation and documentation
- Future Rust/Python/etc. clients need consistent interfaces

### 2. INTERNAL FLEXIBILITY

- Domain models benefit from Go's strong typing and performance
- Complex business logic easier in native Go
- No generation overhead for internal operations

### 3. PRAGMATIC COMPROMISE

- Avoids full generation complexity
- Maintains Go ecosystem benefits
- Enables cross-language capabilities where needed

## IMPLEMENTATION STRATEGY

### PHASE 1: Define Public API Contracts

```typescript
// config.typespec - Public API
model Config {
  version: string;
  safeMode: boolean;
  maxDiskUsage: int32;
  protectedPaths: string[];
}

model CleanResult {
  freedBytes: uint64;
  itemsRemoved: uint32;
  strategy: CleanStrategy;
}
```

### PHASE 2: Generate API Layer

```go
// Generated Go API for external clients
type ConfigAPIClient struct {
  client *http.Client
  baseURL string
}
```

### PHASE 3: Internal Go Domain Models

```go
// Rich domain models with business logic
type Config struct {
  version  string
  safeMode bool

  // Business methods
  func (c *Config) Validate() Result[ValidationContext]
  func (c *Config) Sanitize() SanitizationResult
}
```

### PHASE 4: Mapping Layer

```go
// Convert between API types and domain models
func APIConfigToDomain(apiConfig apitypes.Config) domain.Config
func DomainConfigToAPI(domainConfig domain.Config) apitypes.Config
```

## BENEFITS

1. **Future-Proof**: Ready for multiple language clients
2. **Performance**: Internal operations remain fully native Go
3. **Documentation**: Auto-generated API docs
4. **Validation**: TypeSpec schema validation for external inputs
5. **Evolution**: Can migrate more types to TypeSpec gradually

## COSTS

1. **Complexity**: Two type systems to maintain
2. **Mapping**: Conversion layer between API and domain types
3. **Learning**: Team needs to understand both systems

## MITIGATION STRATEGIES

1. **Automated Mapping**: Code generation for type conversions
2. **Testing**: Comprehensive mapping layer tests
3. **Documentation**: Clear separation of API vs domain types
4. **Gradual Migration**: Start with core types, expand gradually

## NEXT STEPS

1. Create `api/typespec/` directory
2. Define core API types in TypeSpec
3. Set up code generation pipeline
4. Implement mapping layer
5. Migrate one API endpoint as proof of concept

## TIMELINE

- **Week 1**: Define TypeSpec schemas and generation setup
- **Week 2**: Implement mapping layer and proof of concept
- **Week 3**: Migrate remaining public APIs
- **Week 4**: Documentation and team training

---

This decision balances long-term architectural goals with immediate practical needs,
providing foundation for multi-language support while maintaining Go's strengths.

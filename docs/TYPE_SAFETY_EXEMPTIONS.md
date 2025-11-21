# Type Safety Exemption Markers

This document explains the type safety exemption markers used in this codebase.

## TYPE-SAFE-EXEMPT Marker

The `// TYPE-SAFE-EXEMPT: <reason>` marker can be used to intentionally exempt specific code from type safety checks in CI/CD.

### When to Use

This marker should only be used in exceptional circumstances:

1. **Legacy Compatibility**: When maintaining backward compatibility with older APIs that require less strict typing
2. **External Dependencies**: When working with third-party libraries that force the use of `interface{}`, `map[string]any`, or similar constructs
3. **Performance Critical Code**: In rare cases where reflection or unsafe operations are necessary for performance reasons
4. **Serialization/Deserialization**: When working with external data formats that require flexible typing

### Examples

```go
// TYPE-SAFE-EXEMPT: Legacy compatibility method using map[string]any intentionally
func (cfg *EnvironmentConfig) ToMap() map[string]any {
    // Implementation
}
```

### Review Process

Any use of `TYPE-SAFE-EXEMPT` must:
1. Have a clear, specific reason
2. Be reviewed by a senior team member
3. Include a TODO ticket to address the technical debt
4. Be limited in scope and well-documented

### Automated Enforcement

Our CI/CD pipeline (`type-safety.yml`) automatically skips files containing this marker when checking for:
- `map[string]any` usage
- `interface{}` usage  
- `unsafe` package usage
- `reflect` package usage

## Alternatives to Exemptions

Before using an exemption marker, consider these alternatives:

1. **Strong Typing**: Use proper Go types and interfaces
2. **Code Generation**: Generate type-safe code instead of using reflection
3. **Adapter Pattern**: Create type-safe adapters for external dependencies
4. **Schema Validation**: Use proper validation for external data

Remember: Type safety is a core principle of this codebase. Exemptions should be rare and well-justified.
# ROADMAP

**Last Updated:** 2026-04-05
**Focus:** Aspirational items - no timeline commitment

---

## Future Enhancements

### Potential Improvements

| Category             | Item                          | Notes                                                                                  |
| -------------------- | ----------------------------- | -------------------------------------------------------------------------------------- |
| Dependency Injection | Consider samber/do/v2         | Evaluated: Current simple constructor pattern sufficient, DI would be over-engineering |
| Plugin Architecture  | Plugin system for cleaners    | Future enhancement, not required for v1                                                |
| RiskLevelType        | Auto mapstructure decode hook | Investigated: Works correctly, would need decode hook for auto-conversion              |

---

## Deferred Decisions

These items were evaluated and deferred, but may be revisited in future versions:

1. **Plugin Architecture for Cleaners** - Would allow third-party cleaners, but not needed for v1 scope
2. **Advanced DI Container** - Current dependency injection via constructors is sufficient
3. **Automated RiskLevel Configuration** - Manual mapstructure processing works; auto-conversion needs additional hooks

---

## Long-term Vision

- Expand cleaner ecosystem with community-contributed plugins
- Web UI for configuration and monitoring
- Remote execution capabilities
- Integration with cloud storage providers

---

**Note:** Items here are aspirational. No timeline commitments.

# Documentation Generation Implementation TODO

This TODO list tracks the implementation of the documentation generation feature as outlined in `DOCUMENTATION_GENERATION_PLAN.md`.

## Phase 1: Registry & Core Infrastructure

### Core Types and Registry
- [ ] 1. Create Domain type and core type definitions
- [ ] 2. **BREAKING CHANGE**: Remove Description() method from Rule[T] interface in rule.go
- [ ] 3. Create registry.go with Registry interface and implementation
- [ ] 4. Add Description field to RegisteredRule struct (single source of truth)
- [ ] 5. Implement Register(), AllRules(), RulesByDomain(), RulesByGroup() functions
- [ ] 6. Add domain and group tagging via RegistrationOption pattern
- [ ] 7. Implement pointer-equality based rule lookup in registry

### Auto-Registration
- [ ] 8. Implement auto-registering rule creation functions (NewWithDomain, NewWithGroup)
- [ ] 9. Implement WithDescription helper (updates registry, returns same rule)
- [ ] 10. Implement GetDescription helper (retrieves from registry by pointer)

### Domain Inheritance for Hierarchical Rules
- [ ] 11. Modify And(), Or(), Not() to collect and deduplicate domains from children
- [ ] 12. Update quantifiers (AtLeast, Exactly, AtMost) to inherit domains
- [ ] 13. Update helper functions (AllOf, AnyOf, NoneOf) to inherit domains
- [ ] 14. Ensure hierarchical rules auto-register if they have domains

### Documentation Infrastructure
- [ ] 15. Create documenter.go with base interfaces and DocumentOptions
- [ ] 16. Implement rule introspection (extract structure, children, type, domains, groups)
- [ ] 17. Add Metadata() and Examples() optional interfaces
- [ ] 18. Create internal representation of rule hierarchy with domain and group info

### Testing
- [ ] 19. Write comprehensive tests for registry (including group operations and auto-registration)
- [ ] 20. Write tests for domain inheritance in hierarchical rules (And, Or, Not, quantifiers)
- [ ] 21. Write tests for WithDescription/GetDescription (pointer-based lookup)

## Phase 2: Markdown Generator

### Core Implementation
- [x] 22. Implement Markdown generator with tree-based output
- [x] 23. Add domain grouping and filtering to Markdown generator
- [x] 24. Use group names as headers in Markdown output

### Enhancement
- [x] 25. Add collapsible sections and metadata to Markdown output

### Testing
- [x] 26. Write tests for Markdown generator

## Phase 3: JSON Generator

### Core Implementation
- [x] 27. Define JSON schema for rule documentation (including domains and groups)
- [x] 28. Implement JSON generator with full metadata serialization
- [ ] 29. Add JSON Schema validation

### Enhancement
- [x] 30. Support domain and group filtering in JSON output

### Testing
- [x] 31. Write tests for JSON generator

## Phase 4: HTML Generator

### Core Implementation
- [x] 32. Create HTML template system
- [x] 33. Implement interactive UI with JavaScript (search/filter by domain and group)

### Enhancement
- [x] 34. Add domain and group navigation sidebar to HTML output
- [x] 35. Style HTML output with responsive CSS

### Testing
- [x] 36. Write tests for HTML generator

## Phase 5: Mermaid Generator

### Core Implementation
- [x] 37. Implement Mermaid diagram generation
- [x] 38. Color-code or group by domain in Mermaid diagrams
- [x] 39. Label cross-domain subgraphs with group names in Mermaid

### Testing
- [x] 40. Write tests for Mermaid generator

## Phase 6: Polish & Documentation

### Documentation
- [ ] 41. Add comprehensive package documentation
- [ ] 42. Create examples for each format with multi-domain scenarios
- [ ] 43. Document domain-driven architecture patterns (including domain inheritance)
- [ ] 44. Update README with documentation generation section and auto-registration
- [ ] 45. Create best practices guide for domain organization

### Testing & Quality
- [ ] 46. Add integration tests with multi-package scenarios
- [ ] 47. Performance optimization and benchmarks
- [ ] 48. Ensure all code passes golangci-lint
- [ ] 49. Verify 100% test coverage on public APIs

---

## Progress Summary

- **Total Tasks**: 49
- **Completed**: 0
- **In Progress**: 0
- **Pending**: 49

## Notes

- All tasks are organized by implementation phase as outlined in the plan
- Mark tasks with `[x]` when completed
- Update the Progress Summary as tasks are completed
- Refer to `DOCUMENTATION_GENERATION_PLAN.md` for detailed specifications


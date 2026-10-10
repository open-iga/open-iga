# Trade-off

This is to document the list of trade-offs taken during the MVP

- Core trust the validation by plugin. It does not cross-check if all the template strings are mentioned in the configs. This will cause runtime issues, but now this is fine as plugins are expected to be built by connector SDK for MVP

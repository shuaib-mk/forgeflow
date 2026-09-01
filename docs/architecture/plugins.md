# Plugin contracts

Plugins expose a versioned manifest and register implementations through a constrained registrar. The current contracts cover workflow steps, notification handlers, analytics processors, and repository integrations. Registration is atomic: invalid or conflicting plugins never become visible.

Plugins run in-process and are therefore trusted code. Administrators must review and compile plugins with ForgeFlow; arbitrary uploaded binaries are intentionally unsupported. The `run-summary` example demonstrates a useful custom workflow step and sanitizes its output filename.


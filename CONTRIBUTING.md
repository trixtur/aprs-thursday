# Contributing

Keep changes small enough to review. Before opening a pull request, run the checks listed in the README and make sure the sandbox tests pass when loopback networking is available.

Every commit must be cryptographically signed. Pull requests target `main`; direct pushes to `main` are disabled. The required GitHub Actions checks must pass before merging, and at least one approval is required.

Do not commit APRS passcodes, OAuth tokens, server-specific credentials, private planning notes, or generated card archives. Use the local sandbox for dynamic tests and keep it disconnected from live APRS-IS.

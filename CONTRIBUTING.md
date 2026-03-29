# Contributing to Traza Work

We welcome contributions to Traza Work. Before you start, please review this guide.

## Contributor License Agreement

All contributors must agree to our [Contributor License Agreement](CLA.md) before their first pull request can be merged. This is required because Traza Work is dual-licensed:

- **Traza Work** is available under [BSL 1.1](LICENSE) — free to use, modify, and self-host. Converts to Apache 2.0 after 4 years per version.
- **Traza Work Cloud** is a commercial SaaS product maintained by Start Codex SAS, built on top of Traza Work.

The CLA ensures that contributions to the open source project can also be included in Traza Work Cloud. Your contribution will always remain available under BSL 1.1 in this repository.

When you open your first pull request, the CLA Assistant bot will ask you to sign. This is a one-time step.

## How to Contribute

1. Fork the repository.
2. Create a branch for your change.
3. Make your changes following the project conventions (see [docs/04-go-conventions.md](docs/04-go-conventions.md)).
4. Run tests: `go test ./internal/...`
5. Open a pull request against `main` on the upstream repository.

## Code Style

- Go: follow the conventions in [docs/04-go-conventions.md](docs/04-go-conventions.md).
- Frontend: SvelteKit + Svelte 5 + Tailwind 4. Run `cd front && pnpm check` before submitting.
- Keep pull requests focused — one logical change per PR.

## Reporting Issues

Open an issue on GitHub with a clear description of the problem, steps to reproduce, and expected behavior.

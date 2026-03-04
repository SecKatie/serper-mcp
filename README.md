# serper-mcp

> **Note:** This project was generated with [Claude Code](https://claude.ai/claude-code) (claude-sonnet-4-6).

An MCP server that exposes Google search and web scraping via the [Serper API](https://serper.dev).

## Tools

- **search** — performs a Google search and returns results as JSON
- **scrape** — fetches and returns the content of a webpage as JSON

## Configuration

Set your Serper API key via environment variable:

```sh
export SERPER_API_KEY=your_key_here
```

Or place it in `~/.serper-mcp.yaml`:

```yaml
serper_api_key: your_key_here
```

## Usage

```sh
serper-mcp
```

The server communicates over stdio using the MCP protocol. Configure it in your MCP client (e.g. Claude Desktop) as a stdio server.

## Development

```sh
make build   # compile
make test    # run tests with race detector
make lint    # golangci-lint
make audit   # govulncheck + gosec
```

## License

MIT — see [LICENSE](LICENSE).

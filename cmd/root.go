/*
Copyright © 2026 Katie Mulliken <katie@mulliken.net>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/

// Package cmd implements the serper-mcp CLI and MCP server entrypoint.
package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/SecKatie/serper-mcp/internal/serper"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile string
	version = "dev"
)

type searchInput struct {
	Query string `json:"q" jsonschema:"Search query string"`
}

type scrapeInput struct {
	URL string `json:"url" jsonschema:"URL of the page to scrape"`
}

var rootCmd = &cobra.Command{
	Use:     "serper-mcp",
	Version: version,
	Short: "MCP server for Google search and web scraping via Serper",
	Long: `serper-mcp is an MCP server that exposes two tools:
  - search: performs a Google search via the Serper API
  - scrape: fetches and returns the content of a webpage via the Serper scrape API

The Serper API key must be set via the SERPER_API_KEY environment variable
or the serper_api_key key in the config file.`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		apiKey := viper.GetString("serper_api_key")
		if apiKey == "" {
			return fmt.Errorf("SERPER_API_KEY is not set; provide it via environment variable or config file")
		}

		client := serper.NewClient(apiKey)

		server := mcp.NewServer(
			&mcp.Implementation{Name: "serper-mcp", Version: version},
			nil,
		)

		mcp.AddTool(server, &mcp.Tool{
			Name:        "search",
			Description: "Perform a Google search via the Serper API and return results as JSON.",
		}, func(ctx context.Context, _ *mcp.CallToolRequest, in searchInput) (*mcp.CallToolResult, any, error) {
			result, err := client.Search(ctx, in.Query)
			if err != nil {
				return nil, nil, fmt.Errorf("search failed: %w", err)
			}
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: string(result)},
				},
			}, nil, nil
		})

		mcp.AddTool(server, &mcp.Tool{
			Name:        "scrape",
			Description: "Scrape a webpage and return its content as JSON via the Serper scrape API.",
		}, func(ctx context.Context, _ *mcp.CallToolRequest, in scrapeInput) (*mcp.CallToolResult, any, error) {
			result, err := client.Scrape(ctx, in.URL)
			if err != nil {
				return nil, nil, fmt.Errorf("scrape failed: %w", err)
			}
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: string(result)},
				},
			}, nil, nil
		})

		return server.Run(cmd.Context(), &mcp.StdioTransport{})
	},
}

// Execute runs the root command and exits on error.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.serper-mcp.yaml)")
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".serper-mcp")
	}

	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}

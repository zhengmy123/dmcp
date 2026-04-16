package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"
	"time"

	"dynamic_mcp_go_server/internal/infrastructure/store/tooldef"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/client/transport"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/redis/go-redis/v9"
)

type options struct {
	dataFile  string
	redisAddr string
	redisKey  string
	mcpURL    string
	skipSeed  bool
	timeout   time.Duration
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	opts := loadOptions()
	expected, err := loadExpectedFlowMapping(opts.dataFile)
	if err != nil {
		return err
	}
	if len(expected) == 0 {
		return errors.New("no enabled tools found in data file")
	}

	if !opts.skipSeed {
		fmt.Println("Seeding test data via cmd/seedredis ...")
		if err := seedData(opts); err != nil {
			return err
		}
	}

	if err := waitRedisReady(opts); err != nil {
		return err
	}

	fmt.Printf("Connecting to streamable HTTP MCP endpoint: %s\n", opts.mcpURL)
	httpTransport, err := transport.NewStreamableHTTP(opts.mcpURL)
	if err != nil {
		return fmt.Errorf("create streamable HTTP transport failed: %w", err)
	}
	cli := client.NewClient(httpTransport)
	defer func() { _ = cli.Close() }()

	ctx, cancel := context.WithTimeout(context.Background(), opts.timeout)
	defer cancel()

	initReq := mcp.InitializeRequest{
		Params: mcp.InitializeParams{
			ProtocolVersion: "2024-11-05",
			Capabilities:    mcp.ClientCapabilities{},
			ClientInfo: mcp.Implementation{
				Name:    "mcp-protocol-validator",
				Version: "0.1.0",
			},
		},
	}
	if _, err := cli.Initialize(ctx, initReq); err != nil {
		return fmt.Errorf("initialize failed: %w", err)
	}

	listResp, err := cli.ListTools(ctx, mcp.ListToolsRequest{})
	if err != nil {
		return fmt.Errorf("tools/list failed: %w", err)
	}
	if err := validateToolsList(listResp, expected); err != nil {
		return err
	}

	if err := validateToolCalls(ctx, cli, expected); err != nil {
		return err
	}

	fmt.Println("Success: MCP protocol validation passed with streamable HTTP client.")
	return nil
}

func validateToolsList(resp *mcp.ListToolsResult, expected map[string]string) error {
	if resp == nil {
		return errors.New("tools/list result is nil")
	}
	got := make(map[string]struct{}, len(resp.Tools))
	for _, t := range resp.Tools {
		got[t.Name] = struct{}{}
	}
	for tool := range expected {
		if _, ok := got[tool]; !ok {
			return fmt.Errorf("tools/list missing expected tool: %s", tool)
		}
	}
	return nil
}

func validateToolCalls(ctx context.Context, cli *client.Client, expected map[string]string) error {
	argsByTool := map[string]map[string]any{
		"search_users": {
			"query": "alice",
			"limit": 3,
		},
		"set_user_flag": {
			"user_id":  "u-1",
			"flag_key": "beta",
			"enabled":  true,
		},
		"get_order_summary": {
			"order_id":      "o-1",
			"include_items": true,
		},
	}

	toolNames := make([]string, 0, len(expected))
	for name := range expected {
		toolNames = append(toolNames, name)
	}
	sort.Strings(toolNames)

	for _, name := range toolNames {
		callReq := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name:      name,
				Arguments: argsByTool[name],
			},
		}
		callResp, err := cli.CallTool(ctx, callReq)
		if err != nil {
			return fmt.Errorf("tools/call failed for %s: %w", name, err)
		}
		flow, tool, err := parseFlowAndTool(callResp)
		if err != nil {
			return fmt.Errorf("parse tools/call result failed for %s: %w", name, err)
		}
		if flow != expected[name] || tool != name {
			return fmt.Errorf("tools/call mismatch for %s: flow=%s tool=%s", name, flow, tool)
		}
	}
	return nil
}

func parseFlowAndTool(result *mcp.CallToolResult) (string, string, error) {
	if result.StructuredContent != nil {
		if flow, tool, ok := extractFlowAndTool(result.StructuredContent); ok {
			return flow, tool, nil
		}
	}
	for _, c := range result.Content {
		tc, ok := c.(mcp.TextContent)
		if !ok {
			continue
		}
		var payload any
		if err := json.Unmarshal([]byte(tc.Text), &payload); err != nil {
			continue
		}
		if flow, tool, ok := extractFlowAndTool(payload); ok {
			return flow, tool, nil
		}
	}
	return "", "", errors.New("flow/tool not found in call result")
}

func extractFlowAndTool(v any) (string, string, bool) {
	obj, ok := v.(map[string]any)
	if !ok {
		return "", "", false
	}
	flow, _ := obj["flow"].(string)
	tool, _ := obj["tool"].(string)
	flow = strings.TrimSpace(flow)
	tool = strings.TrimSpace(tool)
	if flow == "" || tool == "" {
		return "", "", false
	}
	return flow, tool, true
}

func loadExpectedFlowMapping(path string) (map[string]string, error) {
	if strings.TrimSpace(path) == "" {
		path = "docs/redis-flow-test-data.json"
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	defs, err := tooldef.ParseToolDefinitions(raw)
	if err != nil {
		return nil, err
	}
	out := make(map[string]string)
	for _, def := range defs {
		out[def.Name] = def.VAuthKey
	}
	return out, nil
}

func seedData(opts options) error {
	cmd := exec.Command("go", "run", "./cmd/seedredis")
	cmd.Env = append(os.Environ(),
		"REDIS_ADDR="+opts.redisAddr,
		"REDIS_KEY="+opts.redisKey,
		"DATA_FILE="+opts.dataFile,
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func waitRedisReady(opts options) error {
	redisClient := redis.NewClient(&redis.Options{Addr: opts.redisAddr, DB: 0})
	defer redisClient.Close()

	ctx, cancel := context.WithTimeout(context.Background(), opts.timeout)
	defer cancel()

	for {
		if err := redisClient.Ping(ctx).Err(); err == nil {
			return nil
		}
		select {
		case <-ctx.Done():
			return fmt.Errorf("redis not ready at %s within %s", opts.redisAddr, opts.timeout)
		case <-time.After(250 * time.Millisecond):
		}
	}
}

func loadOptions() options {
	return options{
		dataFile:  getenv("DATA_FILE", "docs/redis-flow-test-data.json"),
		redisAddr: getenv("REDIS_ADDR", "127.0.0.1:6379"),
		redisKey:  getenv("REDIS_KEY", "mcp:tool-definitions"),
		mcpURL:    getenv("MCP_URL", "http://127.0.0.1:18080/mcp/user-service"),
		skipSeed:  getenv("SKIP_SEED", "0") == "1",
		timeout:   time.Duration(getenvInt("MCP_TIMEOUT_SECONDS", 20)) * time.Second,
	}
}

func getenv(key, fallback string) string {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return fallback
	}
	return v
}

func getenvInt(key string, fallback int) int {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback
	}
	var v int
	if _, err := fmt.Sscanf(raw, "%d", &v); err != nil || v <= 0 {
		return fallback
	}
	return v
}

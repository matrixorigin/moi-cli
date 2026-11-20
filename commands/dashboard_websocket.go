package commands

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"time"

	sdk "github.com/matrixorigin/moi-go-sdk/v5"
	"google.golang.org/protobuf/encoding/protojson"
)

// 执行一次明确的消息并打印响应；不重连或重放写操作。
func runDashboardServiceConnectWebSocket(ctx *Context, args []string) (resultErr error) {
	fs := flag.NewFlagSet("dashboard connect-web-socket", flag.ContinueOnError)
	workspaceID := fs.String("workspace-id", "", "工作区 ID，省略时读取配置")
	action := fs.String("action", "ping", "ping、preview、create-chart 或 update-chart")
	messageID := fs.String("message-id", "cli-dashboard", "本次连接中的消息 ID")
	dashboardID := fs.String("dashboard-id", "", "看板 ID")
	chartID := fs.String("chart-id", "", "图表 ID")
	sqlText := fs.String("sql-text", "", "预览 SQL")
	bodyJSON := fs.String("json", "", "图表请求 JSON")
	bodyFile := fs.String("json-file", "", "图表请求 JSON 文件")
	timeout := fs.Duration("timeout", 30*time.Second, "连接、发送及读取的总期限")
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: moi-cli dashboard connect-web-socket [options]\n\nProduct SDK: DashboardService.ConnectWebSocket\n\nOptions:")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	var request sdk.DataDashboardChartMutationRequest
	switch *action {
	case "ping", "preview":
	case "create-chart", "update-chart":
		rawBody, err := loadJSONBytes(*bodyJSON, *bodyFile)
		if err != nil {
			return err
		}
		if err := json.Unmarshal(rawBody, &request); err != nil {
			return fmt.Errorf("parse --json: %w", err)
		}
	default:
		return fmt.Errorf("unknown dashboard WebSocket action: %s", *action)
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	operation, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()
	session, err := ctx.Client.Dashboard(wid).ConnectWebSocket(operation)
	if err != nil {
		return err
	}
	defer func() { resultErr = errors.Join(resultErr, session.Close()) }()
	switch *action {
	case "ping":
		err = session.Ping(operation, *messageID)
	case "preview":
		err = session.PreviewSQL(operation, *messageID, *dashboardID, *chartID, *sqlText)
	case "create-chart":
		err = session.CreateChart(operation, *messageID, *dashboardID, &request)
	case "update-chart":
		err = session.UpdateChart(operation, *messageID, *dashboardID, *chartID, &request)
	}
	if err != nil {
		return err
	}
	frame, err := session.Receive(operation)
	if err != nil {
		return err
	}
	payload, err := (protojson.MarshalOptions{UseProtoNames: true}).Marshal(frame)
	if err != nil {
		return err
	}
	return printJSON(json.RawMessage(payload))
}

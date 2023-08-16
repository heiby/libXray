package libXray

import (
	"context"
	"fmt"
	"reflect"
	"time"

	statsService "github.com/xtls/xray-core/app/stats/command"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

func isNil(i interface{}) bool {
	vi := reflect.ValueOf(i)
	if vi.Kind() == reflect.Ptr {
		return vi.IsNil()
	}
	return i == nil
}

func convertToJsonString(m proto.Message) (string, error) {
	if isNil(m) {
		return "", fmt.Errorf("m is nil")
	}
	ops := protojson.MarshalOptions{}
	b, err := ops.Marshal(m)
	if err != nil {
		return "", err
	}
	str := string(b)
	return str, nil
}

// query system stats and outbound stats.
// server means The API server address, like "127.0.0.1:8080".
// return json string with success: {"code":0,"data":{"sysStats":{...}},"stats":{...}}}
// return json string with failed: {"code":-1,"message":"failed reason"}
func QueryStats(server string) string {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(3)*time.Second)
	conn, err := grpc.DialContext(ctx, server, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	close := func() {
		cancel()
		if !isNil(conn) {
			conn.Close()
		}
	}
	defer close()
	if err != nil {
		return fmt.Sprintf(`{"code":-1,"message":"%s"}`, err)
	}

	client := statsService.NewStatsServiceClient(conn)

	sysStatsReq := &statsService.SysStatsRequest{}
	sysStatsRes, err := client.GetSysStats(ctx, sysStatsReq)
	if err != nil {
		return fmt.Sprintf(`{"code":-1,"message":"%s"}`, err)
	}
	sysStatsStr, err := convertToJsonString(sysStatsRes)
	if err != nil {
		return fmt.Sprintf(`{"code":-1,"message":"%s"}`, err)
	}

	statsReq := &statsService.QueryStatsRequest{
		Pattern: "",
		Reset_:  false,
	}
	statsRes, err := client.QueryStats(ctx, statsReq)
	if err != nil {
		return fmt.Sprintf(`{"code":-1,"message":"%s"}`, err)
	}
	statsStr, err := convertToJsonString(statsRes)
	if err != nil {
		return fmt.Sprintf(`{"code":-1,"message":"%s"}`, err)
	}
	retVal := fmt.Sprintf(`{"code":0,"data":{"sysStats":%s,"stats":%s}}`, sysStatsStr, statsStr)
	return retVal
}

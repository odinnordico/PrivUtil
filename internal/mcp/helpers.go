package mcp

import (
	"encoding/json"
	"strings"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	pb "github.com/odinnordico/privutil/proto"
)

// structured converts a protobuf response into a generic JSON object for use as
// an MCP tool's structured output. The in-band "error" field is dropped (errors
// are surfaced via errResult before this is called), so a successful result
// carries only data.
func structured(m proto.Message) (map[string]any, error) {
	b, err := protojson.Marshal(m)
	if err != nil {
		return nil, err
	}
	var out map[string]any
	if err := json.Unmarshal(b, &out); err != nil {
		return nil, err
	}
	delete(out, "error")
	return out, nil
}

// String-to-enum helpers (case-insensitive). Unknown values map to the zero enum
// value, which the underlying handlers treat as their documented default.
func dataFormat(s string) pb.DataFormat {
	return pb.DataFormat(pb.DataFormat_value[strings.ToUpper(strings.TrimSpace(s))])
}
func textAction(s string) pb.TextAction {
	return pb.TextAction(pb.TextAction_value[strings.ToUpper(strings.TrimSpace(s))])
}
func listAction(s string) pb.ListAction {
	return pb.ListAction(pb.ListAction_value["LIST_"+strings.ToUpper(strings.TrimSpace(s))])
}
func percentMode(s string) pb.PercentMode {
	return pb.PercentMode(pb.PercentMode_value["PCT_"+strings.ToUpper(strings.TrimSpace(s))])
}
func unitCategory(s string) pb.UnitCategory {
	return pb.UnitCategory(pb.UnitCategory_value["UNIT_"+strings.ToUpper(strings.TrimSpace(s))])
}

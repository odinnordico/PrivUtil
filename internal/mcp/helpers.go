package mcp

import (
	"encoding/json"
	"strings"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	pb "github.com/odinnordico/privutil/proto"
)

// structured converts a protobuf response into a generic JSON object for use as
// an MCP tool's structured output. EmitUnpopulated keeps false/zero values so a
// negative result (e.g. valid:false) is not silently dropped. The in-band
// "error" field is removed (errors are surfaced via errResult before this is
// called); tools whose "error" field is meaningful re-add it themselves.
func structured(m proto.Message) (map[string]any, error) {
	b, err := protojson.MarshalOptions{EmitUnpopulated: true}.Marshal(m)
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

// String-to-enum helpers (case-insensitive). The bool reports whether the value
// was recognized, so tools can reject a typo instead of silently using the zero
// enum (which some handlers treat as a default, others as the first action).
func dataFormat(s string) (pb.DataFormat, bool) {
	v, ok := pb.DataFormat_value[strings.ToUpper(strings.TrimSpace(s))]
	return pb.DataFormat(v), ok
}
func textAction(s string) (pb.TextAction, bool) {
	v, ok := pb.TextAction_value[strings.ToUpper(strings.TrimSpace(s))]
	return pb.TextAction(v), ok
}
func listAction(s string) (pb.ListAction, bool) {
	v, ok := pb.ListAction_value["LIST_"+strings.ToUpper(strings.TrimSpace(s))]
	return pb.ListAction(v), ok
}
func percentMode(s string) (pb.PercentMode, bool) {
	v, ok := pb.PercentMode_value["PCT_"+strings.ToUpper(strings.TrimSpace(s))]
	return pb.PercentMode(v), ok
}
func unitCategory(s string) (pb.UnitCategory, bool) {
	v, ok := pb.UnitCategory_value["UNIT_"+strings.ToUpper(strings.TrimSpace(s))]
	return pb.UnitCategory(v), ok
}

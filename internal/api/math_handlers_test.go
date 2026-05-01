package api

import (
	"context"
	"math"
	"testing"

	pb "github.com/odinnordico/privutil/proto"
)

var mathSrv = &Server{}

// ─── Math expression evaluator ────────────────────────────────────────────────

func TestMathEval_Arithmetic(t *testing.T) {
	cases := []struct {
		expr string
		want float64
	}{
		{"2 + 3", 5},
		{"10 - 4", 6},
		{"3 * 4", 12},
		{"15 / 3", 5},
		{"10 % 3", 1},
		{"2 ^ 10", 1024},
		{"2 ^ 3 ^ 2", 512}, // right-assoc: 2^(3^2) = 2^9
		{"-5 + 3", -2},
		{"(2 + 3) * 4", 20},
	}
	for _, tc := range cases {
		t.Run(tc.expr, func(t *testing.T) {
			resp, err := mathSrv.MathEval(context.Background(), &pb.MathEvalRequest{
				Expression: tc.expr, Precision: 10,
			})
			if err != nil {
				t.Fatal(err)
			}
			if resp.Error != "" {
				t.Fatalf("unexpected error: %s", resp.Error)
			}
			if math.Abs(resp.RawValue-tc.want) > 1e-9 {
				t.Errorf("expr=%q: got %v, want %v", tc.expr, resp.RawValue, tc.want)
			}
		})
	}
}

func TestMathEval_Functions(t *testing.T) {
	cases := []struct {
		expr string
		want float64
		tol  float64
	}{
		{"sqrt(9)", 3, 1e-10},
		{"abs(-5)", 5, 1e-10},
		{"floor(3.7)", 3, 1e-10},
		{"ceil(3.2)", 4, 1e-10},
		{"round(3.5)", 4, 1e-10},
		{"log(e)", 1, 1e-10},
		{"log2(8)", 3, 1e-10},
		{"log10(1000)", 3, 1e-10},
		{"sin(0)", 0, 1e-10},
		{"cos(0)", 1, 1e-10},
		{"max(3, 7, 2)", 7, 1e-10},
		{"min(3, 7, 2)", 2, 1e-10},
		{"pow(2, 8)", 256, 1e-10},
		{"factorial(5)", 120, 1e-10},
		{"gcd(12, 8)", 4, 1e-10},
		{"lcm(4, 6)", 12, 1e-10},
		{"hypot(3, 4)", 5, 1e-10},
		{"cbrt(27)", 3, 1e-10},
	}
	for _, tc := range cases {
		t.Run(tc.expr, func(t *testing.T) {
			resp, _ := mathSrv.MathEval(context.Background(), &pb.MathEvalRequest{
				Expression: tc.expr, Precision: 10,
			})
			if resp.Error != "" {
				t.Fatalf("error: %s", resp.Error)
			}
			if math.Abs(resp.RawValue-tc.want) > tc.tol {
				t.Errorf("expr=%q: got %.10f, want %.10f", tc.expr, resp.RawValue, tc.want)
			}
		})
	}
}

func TestMathEval_Constants(t *testing.T) {
	resp, _ := mathSrv.MathEval(context.Background(), &pb.MathEvalRequest{Expression: "pi"})
	if math.Abs(resp.RawValue-math.Pi) > 1e-10 {
		t.Errorf("pi: got %v", resp.RawValue)
	}
	resp, _ = mathSrv.MathEval(context.Background(), &pb.MathEvalRequest{Expression: "e"})
	if math.Abs(resp.RawValue-math.E) > 1e-10 {
		t.Errorf("e: got %v", resp.RawValue)
	}
	resp, _ = mathSrv.MathEval(context.Background(), &pb.MathEvalRequest{Expression: "tau"})
	if math.Abs(resp.RawValue-2*math.Pi) > 1e-10 {
		t.Errorf("tau: got %v", resp.RawValue)
	}
}

func TestMathEval_Variables(t *testing.T) {
	resp, _ := mathSrv.MathEval(context.Background(), &pb.MathEvalRequest{
		Expression: "pi * r ^ 2",
		Variables:  []*pb.MathVariable{{Name: "r", Value: 5}},
		Precision:  10,
	})
	want := math.Pi * 25
	if math.Abs(resp.RawValue-want) > 1e-9 {
		t.Errorf("circle area: got %v, want %v", resp.RawValue, want)
	}
}

func TestMathEval_Degrees(t *testing.T) {
	// sin(90°) = 1
	resp, _ := mathSrv.MathEval(context.Background(), &pb.MathEvalRequest{
		Expression: "sin(90)", Degrees: true,
	})
	if math.Abs(resp.RawValue-1) > 1e-10 {
		t.Errorf("sin(90deg) = %v, want 1", resp.RawValue)
	}
}

func TestMathEval_DivisionByZero(t *testing.T) {
	resp, _ := mathSrv.MathEval(context.Background(), &pb.MathEvalRequest{Expression: "1 / 0"})
	if resp.Error == "" {
		t.Error("expected division by zero error")
	}
}

func TestMathEval_UnknownVariable(t *testing.T) {
	resp, _ := mathSrv.MathEval(context.Background(), &pb.MathEvalRequest{Expression: "x + 1"})
	if resp.Error == "" {
		t.Error("expected undefined variable error")
	}
}

func TestMathEval_InvalidExpression(t *testing.T) {
	cases := []string{"2 +", "(1 + 2", "1 @ 2"}
	for _, expr := range cases {
		resp, _ := mathSrv.MathEval(context.Background(), &pb.MathEvalRequest{Expression: expr})
		if resp.Error == "" {
			t.Errorf("expected error for %q", expr)
		}
	}
}

func TestMathEval_Empty(t *testing.T) {
	resp, _ := mathSrv.MathEval(context.Background(), &pb.MathEvalRequest{Expression: ""})
	if resp.Error == "" {
		t.Error("expected error for empty expression")
	}
}

func TestMathEval_Precision(t *testing.T) {
	resp, _ := mathSrv.MathEval(context.Background(), &pb.MathEvalRequest{
		Expression: "1 / 3", Precision: 4,
	})
	if resp.Result != "0.3333" {
		t.Errorf("precision=4 for 1/3: got %q, want %q", resp.Result, "0.3333")
	}
}

// ─── Percentage calculator ────────────────────────────────────────────────────

func TestPercentageCalc_XofY(t *testing.T) {
	resp, err := mathSrv.PercentageCalc(context.Background(), &pb.PercentageRequest{
		Mode: pb.PercentMode_PCT_X_OF_Y, A: 20, B: 150,
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.Error != "" {
		t.Fatal(resp.Error)
	}
	if math.Abs(resp.Result-30) > 1e-9 {
		t.Errorf("20%% of 150 = %v, want 30", resp.Result)
	}
}

func TestPercentageCalc_WhatPercent(t *testing.T) {
	resp, _ := mathSrv.PercentageCalc(context.Background(), &pb.PercentageRequest{
		Mode: pb.PercentMode_PCT_WHAT, A: 30, B: 150,
	})
	if math.Abs(resp.Result-20) > 1e-9 {
		t.Errorf("30 is %%?? of 150 = %v, want 20", resp.Result)
	}
}

func TestPercentageCalc_PercentChange_Increase(t *testing.T) {
	resp, _ := mathSrv.PercentageCalc(context.Background(), &pb.PercentageRequest{
		Mode: pb.PercentMode_PCT_CHANGE, A: 100, B: 150,
	})
	if math.Abs(resp.Result-50) > 1e-9 {
		t.Errorf("%% change 100→150 = %v, want 50", resp.Result)
	}
}

func TestPercentageCalc_PercentChange_Decrease(t *testing.T) {
	resp, _ := mathSrv.PercentageCalc(context.Background(), &pb.PercentageRequest{
		Mode: pb.PercentMode_PCT_CHANGE, A: 200, B: 100,
	})
	if math.Abs(resp.Result-(-50)) > 1e-9 {
		t.Errorf("%% change 200→100 = %v, want -50", resp.Result)
	}
}

func TestPercentageCalc_Reverse(t *testing.T) {
	// 30 is 20% of what?  → 150
	resp, _ := mathSrv.PercentageCalc(context.Background(), &pb.PercentageRequest{
		Mode: pb.PercentMode_PCT_REVERSE, A: 30, B: 20,
	})
	if math.Abs(resp.Result-150) > 1e-9 {
		t.Errorf("30 is 20%% of ?? = %v, want 150", resp.Result)
	}
}

func TestPercentageCalc_DivisionByZero(t *testing.T) {
	resp, _ := mathSrv.PercentageCalc(context.Background(), &pb.PercentageRequest{
		Mode: pb.PercentMode_PCT_WHAT, A: 10, B: 0,
	})
	if resp.Error == "" {
		t.Error("expected error when B=0")
	}
}

// ─── Temperature converter ────────────────────────────────────────────────────

func TestTempConvert_CtoAll(t *testing.T) {
	resp, err := mathSrv.TempConvert(context.Background(), &pb.TempConvertRequest{Value: 100, FromUnit: "c"})
	if err != nil {
		t.Fatal(err)
	}
	if resp.Error != "" {
		t.Fatal(resp.Error)
	}
	if math.Abs(resp.Celsius-100) > 1e-9 {
		t.Errorf("celsius: got %v", resp.Celsius)
	}
	if math.Abs(resp.Fahrenheit-212) > 1e-9 {
		t.Errorf("fahrenheit: got %v, want 212", resp.Fahrenheit)
	}
	if math.Abs(resp.Kelvin-373.15) > 1e-9 {
		t.Errorf("kelvin: got %v, want 373.15", resp.Kelvin)
	}
}

func TestTempConvert_FtoAll(t *testing.T) {
	resp, _ := mathSrv.TempConvert(context.Background(), &pb.TempConvertRequest{Value: 32, FromUnit: "f"})
	if math.Abs(resp.Celsius-0) > 1e-9 {
		t.Errorf("32F → 0C: got %v", resp.Celsius)
	}
	if math.Abs(resp.Kelvin-273.15) > 1e-9 {
		t.Errorf("32F → 273.15K: got %v", resp.Kelvin)
	}
}

func TestTempConvert_KtoAll(t *testing.T) {
	resp, _ := mathSrv.TempConvert(context.Background(), &pb.TempConvertRequest{Value: 0, FromUnit: "k"})
	if math.Abs(resp.Celsius-(-273.15)) > 1e-9 {
		t.Errorf("0K → -273.15C: got %v", resp.Celsius)
	}
}

func TestTempConvert_NegativeKelvin(t *testing.T) {
	resp, _ := mathSrv.TempConvert(context.Background(), &pb.TempConvertRequest{Value: -1, FromUnit: "k"})
	if resp.Error == "" {
		t.Error("expected error for negative Kelvin")
	}
}

func TestTempConvert_UnknownUnit(t *testing.T) {
	resp, _ := mathSrv.TempConvert(context.Background(), &pb.TempConvertRequest{Value: 100, FromUnit: "x"})
	if resp.Error == "" {
		t.Error("expected error for unknown unit")
	}
}

// ─── Unit converter ───────────────────────────────────────────────────────────

func TestUnitConvert_Bytes(t *testing.T) {
	resp, err := mathSrv.UnitConvert(context.Background(), &pb.UnitConvertRequest{
		Value: 1, FromUnit: "GB", Category: pb.UnitCategory_UNIT_BYTES,
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.Error != "" {
		t.Fatal(resp.Error)
	}
	// Find MB result
	for _, r := range resp.Results {
		if r.Unit == "MB" {
			if math.Abs(r.Value-1000) > 1e-6 {
				t.Errorf("1 GB = %v MB, want 1000", r.Value)
			}
		}
		if r.Unit == "GiB" {
			// 1 GB = 1e9 bytes / 1024^3 ≈ 0.931323 GiB
			if math.Abs(r.Value-0.9313225746) > 1e-6 {
				t.Errorf("1 GB = %v GiB, want ~0.931", r.Value)
			}
		}
	}
}

func TestUnitConvert_Length(t *testing.T) {
	resp, _ := mathSrv.UnitConvert(context.Background(), &pb.UnitConvertRequest{
		Value: 1, FromUnit: "km", Category: pb.UnitCategory_UNIT_LENGTH,
	})
	for _, r := range resp.Results {
		switch r.Unit {
		case "m":
			if math.Abs(r.Value-1000) > 1e-6 {
				t.Errorf("1km = %vm, want 1000", r.Value)
			}
		case "mi":
			if math.Abs(r.Value-0.621371) > 1e-4 {
				t.Errorf("1km = %vmi, want ~0.621", r.Value)
			}
		}
	}
}

func TestUnitConvert_Mass(t *testing.T) {
	resp, _ := mathSrv.UnitConvert(context.Background(), &pb.UnitConvertRequest{
		Value: 1, FromUnit: "kg", Category: pb.UnitCategory_UNIT_MASS,
	})
	for _, r := range resp.Results {
		if r.Unit == "lb" {
			if math.Abs(r.Value-2.20462) > 1e-4 {
				t.Errorf("1kg = %vlb, want ~2.205", r.Value)
			}
		}
	}
}

func TestUnitConvert_Temperature_Speed(t *testing.T) {
	resp, _ := mathSrv.UnitConvert(context.Background(), &pb.UnitConvertRequest{
		Value: 100, FromUnit: "km/h", Category: pb.UnitCategory_UNIT_SPEED,
	})
	for _, r := range resp.Results {
		if r.Unit == "mph" {
			if math.Abs(r.Value-62.1371) > 1e-3 {
				t.Errorf("100km/h = %vmph, want ~62.14", r.Value)
			}
		}
	}
}

func TestUnitConvert_UnknownUnit(t *testing.T) {
	resp, _ := mathSrv.UnitConvert(context.Background(), &pb.UnitConvertRequest{
		Value: 1, FromUnit: "xyz", Category: pb.UnitCategory_UNIT_LENGTH,
	})
	if resp.Error == "" {
		t.Error("expected error for unknown unit")
	}
}

func TestUnitConvert_Area(t *testing.T) {
	resp, _ := mathSrv.UnitConvert(context.Background(), &pb.UnitConvertRequest{
		Value: 1, FromUnit: "m²", Category: pb.UnitCategory_UNIT_AREA,
	})
	for _, r := range resp.Results {
		if r.Unit == "cm²" {
			if math.Abs(r.Value-10000) > 1e-6 {
				t.Errorf("1m² = %vcm², want 10000", r.Value)
			}
		}
	}
}

func TestUnitConvert_Volume(t *testing.T) {
	resp, _ := mathSrv.UnitConvert(context.Background(), &pb.UnitConvertRequest{
		Value: 1, FromUnit: "l", Category: pb.UnitCategory_UNIT_VOLUME,
	})
	for _, r := range resp.Results {
		if r.Unit == "ml" {
			if math.Abs(r.Value-1000) > 1e-6 {
				t.Errorf("1l = %vml, want 1000", r.Value)
			}
		}
	}
}

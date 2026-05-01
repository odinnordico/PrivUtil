package api

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
	"unicode"

	pb "github.com/odinnordico/privutil/proto"
)

// ─── Math expression evaluator ────────────────────────────────────────────────
//
// Self-contained recursive descent parser (no external deps).
// Operator precedence: + - < * / % < ^ (right-assoc) < unary - < primary
// Supports: arithmetic, parentheses, user variables, constants (pi,e,phi,tau),
// and a full suite of math functions.

type tokKind int

const (
	tokNum tokKind = iota
	tokIdent
	tokPlus
	tokMinus
	tokStar
	tokSlash
	tokCaret
	tokPercent
	tokLParen
	tokRParen
	tokComma
	tokEOF
)

type token struct {
	kind tokKind
	text string
	num  float64
}

// tokenize converts an expression string into tokens.
func tokenize(expr string) ([]token, error) {
	var tokens []token
	i := 0
	runes := []rune(expr)
	n := len(runes)
	for i < n {
		ch := runes[i]
		if unicode.IsSpace(ch) {
			i++
			continue
		}
		switch {
		case ch >= '0' && ch <= '9' || ch == '.':
			j := i
			hasDot := false
			hasExp := false
			for j < n && (runes[j] >= '0' && runes[j] <= '9' || runes[j] == '.' ||
				(runes[j] == 'e' || runes[j] == 'E') ||
				((runes[j] == '+' || runes[j] == '-') && j > 0 && (runes[j-1] == 'e' || runes[j-1] == 'E'))) {
				if runes[j] == '.' {
					if hasDot {
						break
					}
					hasDot = true
				}
				if runes[j] == 'e' || runes[j] == 'E' {
					if hasExp {
						break
					}
					hasExp = true
				}
				j++
			}
			s := string(runes[i:j])
			f, err := strconv.ParseFloat(s, 64)
			if err != nil {
				return nil, fmt.Errorf("invalid number %q", s)
			}
			tokens = append(tokens, token{kind: tokNum, text: s, num: f})
			i = j
		case unicode.IsLetter(ch) || ch == '_':
			j := i
			for j < n && (unicode.IsLetter(runes[j]) || unicode.IsDigit(runes[j]) || runes[j] == '_') {
				j++
			}
			tokens = append(tokens, token{kind: tokIdent, text: string(runes[i:j])})
			i = j
		case ch == '+':
			tokens = append(tokens, token{kind: tokPlus, text: "+"})
			i++
		case ch == '-':
			tokens = append(tokens, token{kind: tokMinus, text: "-"})
			i++
		case ch == '*':
			tokens = append(tokens, token{kind: tokStar, text: "*"})
			i++
		case ch == '/':
			tokens = append(tokens, token{kind: tokSlash, text: "/"})
			i++
		case ch == '^':
			tokens = append(tokens, token{kind: tokCaret, text: "^"})
			i++
		case ch == '%':
			tokens = append(tokens, token{kind: tokPercent, text: "%"})
			i++
		case ch == '(':
			tokens = append(tokens, token{kind: tokLParen, text: "("})
			i++
		case ch == ')':
			tokens = append(tokens, token{kind: tokRParen, text: ")"})
			i++
		case ch == ',':
			tokens = append(tokens, token{kind: tokComma, text: ","})
			i++
		default:
			return nil, fmt.Errorf("unexpected character %q", string(ch))
		}
	}
	tokens = append(tokens, token{kind: tokEOF})
	return tokens, nil
}

// parser is a recursive descent parser with operator-precedence.
type mathParser struct {
	tokens  []token
	pos     int
	vars    map[string]float64
	degrees bool // if true, trig functions receive/return degrees
}

func (p *mathParser) peek() token {
	if p.pos >= len(p.tokens) {
		return token{kind: tokEOF}
	}
	return p.tokens[p.pos]
}

func (p *mathParser) consume() token {
	t := p.peek()
	p.pos++
	return t
}

func (p *mathParser) expect(k tokKind) (token, error) {
	t := p.consume()
	if t.kind != k {
		return token{}, fmt.Errorf("expected %d, got %q", k, t.text)
	}
	return t, nil
}

// parseExpr is the entry point.
func (p *mathParser) parseExpr() (float64, error) {
	return p.parseAddSub()
}

func (p *mathParser) parseAddSub() (float64, error) {
	left, err := p.parseMulDiv()
	if err != nil {
		return 0, err
	}
	for p.peek().kind == tokPlus || p.peek().kind == tokMinus {
		op := p.consume()
		right, err := p.parseMulDiv()
		if err != nil {
			return 0, err
		}
		if op.kind == tokPlus {
			left += right
		} else {
			left -= right
		}
	}
	return left, nil
}

func (p *mathParser) parseMulDiv() (float64, error) {
	left, err := p.parseUnary()
	if err != nil {
		return 0, err
	}
	for {
		switch p.peek().kind {
		case tokStar:
			p.consume()
			right, err := p.parseUnary()
			if err != nil {
				return 0, err
			}
			left *= right
		case tokSlash:
			p.consume()
			right, err := p.parseUnary()
			if err != nil {
				return 0, err
			}
			if right == 0 {
				return 0, fmt.Errorf("division by zero")
			}
			left /= right
		case tokPercent:
			p.consume()
			right, err := p.parseUnary()
			if err != nil {
				return 0, err
			}
			if right == 0 {
				return 0, fmt.Errorf("modulo by zero")
			}
			left = math.Mod(left, right)
		default:
			return left, nil
		}
	}
}

// parseUnary handles prefix + and -.
func (p *mathParser) parseUnary() (float64, error) {
	if p.peek().kind == tokMinus {
		p.consume()
		v, err := p.parseUnary()
		return -v, err
	}
	if p.peek().kind == tokPlus {
		p.consume()
		return p.parseUnary()
	}
	return p.parsePower()
}

// parsePower handles ^ (right-associative).
func (p *mathParser) parsePower() (float64, error) {
	base, err := p.parsePrimary()
	if err != nil {
		return 0, err
	}
	if p.peek().kind == tokCaret {
		p.consume()
		// right-associative: recurse on parseUnary
		exp, err := p.parseUnary()
		if err != nil {
			return 0, err
		}
		return math.Pow(base, exp), nil
	}
	return base, nil
}

func (p *mathParser) parsePrimary() (float64, error) {
	t := p.peek()
	switch t.kind {
	case tokNum:
		p.consume()
		return t.num, nil

	case tokLParen:
		p.consume()
		v, err := p.parseExpr()
		if err != nil {
			return 0, err
		}
		if _, err := p.expect(tokRParen); err != nil {
			return 0, err
		}
		return v, nil

	case tokIdent:
		p.consume()
		name := t.text

		// Check for function call
		if p.peek().kind == tokLParen {
			p.consume() // consume '('
			var args []float64
			if p.peek().kind != tokRParen {
				arg, err := p.parseExpr()
				if err != nil {
					return 0, err
				}
				args = append(args, arg)
				for p.peek().kind == tokComma {
					p.consume()
					arg, err := p.parseExpr()
					if err != nil {
						return 0, err
					}
					args = append(args, arg)
				}
			}
			if _, err := p.expect(tokRParen); err != nil {
				return 0, err
			}
			return p.callFunc(name, args)
		}

		// Variable or constant lookup
		return p.lookupVar(name)

	case tokEOF:
		return 0, fmt.Errorf("unexpected end of expression")

	default:
		return 0, fmt.Errorf("unexpected token %q", t.text)
	}
}

func (p *mathParser) lookupVar(name string) (float64, error) {
	// Built-in constants
	switch strings.ToLower(name) {
	case "pi":
		return math.Pi, nil
	case "e":
		return math.E, nil
	case "phi":
		return (1 + math.Sqrt(5)) / 2, nil
	case "tau":
		return 2 * math.Pi, nil
	case "inf", "infinity":
		return math.Inf(1), nil
	case "nan":
		return math.NaN(), nil
	}

	// User variables
	if v, ok := p.vars[name]; ok {
		return v, nil
	}
	return 0, fmt.Errorf("undefined variable %q", name)
}

func (p *mathParser) toRad(deg float64) float64 {
	if p.degrees {
		return deg * math.Pi / 180
	}
	return deg
}

func (p *mathParser) fromRad(rad float64) float64 {
	if p.degrees {
		return rad * 180 / math.Pi
	}
	return rad
}

func (p *mathParser) callFunc(name string, args []float64) (float64, error) {
	argc := func(n int) error {
		if len(args) != n {
			return fmt.Errorf("%s() expects %d argument(s), got %d", name, n, len(args))
		}
		return nil
	}
	argc2 := func(min, max int) error {
		if len(args) < min || len(args) > max {
			return fmt.Errorf("%s() expects %d-%d arguments, got %d", name, min, max, len(args))
		}
		return nil
	}

	switch strings.ToLower(name) {
	case "sqrt":
		if err := argc(1); err != nil {
			return 0, err
		}
		if args[0] < 0 {
			return 0, fmt.Errorf("sqrt of negative number")
		}
		return math.Sqrt(args[0]), nil
	case "cbrt":
		if err := argc(1); err != nil {
			return 0, err
		}
		return math.Cbrt(args[0]), nil
	case "abs":
		if err := argc(1); err != nil {
			return 0, err
		}
		return math.Abs(args[0]), nil
	case "floor":
		if err := argc(1); err != nil {
			return 0, err
		}
		return math.Floor(args[0]), nil
	case "ceil":
		if err := argc(1); err != nil {
			return 0, err
		}
		return math.Ceil(args[0]), nil
	case "round":
		if err := argc(1); err != nil {
			return 0, err
		}
		return math.Round(args[0]), nil
	case "trunc":
		if err := argc(1); err != nil {
			return 0, err
		}
		return math.Trunc(args[0]), nil
	case "sign", "sgn":
		if err := argc(1); err != nil {
			return 0, err
		}
		if args[0] < 0 {
			return -1, nil
		}
		if args[0] > 0 {
			return 1, nil
		}
		return 0, nil
	case "exp":
		if err := argc(1); err != nil {
			return 0, err
		}
		return math.Exp(args[0]), nil
	case "exp2":
		if err := argc(1); err != nil {
			return 0, err
		}
		return math.Exp2(args[0]), nil
	case "log", "ln":
		if err := argc2(1, 2); err != nil {
			return 0, err
		}
		if args[0] <= 0 {
			return 0, fmt.Errorf("log of non-positive number")
		}
		if len(args) == 2 {
			if args[1] <= 0 || args[1] == 1 {
				return 0, fmt.Errorf("invalid log base")
			}
			return math.Log(args[0]) / math.Log(args[1]), nil
		}
		return math.Log(args[0]), nil
	case "log2":
		if err := argc(1); err != nil {
			return 0, err
		}
		return math.Log2(args[0]), nil
	case "log10":
		if err := argc(1); err != nil {
			return 0, err
		}
		return math.Log10(args[0]), nil
	case "sin":
		if err := argc(1); err != nil {
			return 0, err
		}
		return math.Sin(p.toRad(args[0])), nil
	case "cos":
		if err := argc(1); err != nil {
			return 0, err
		}
		return math.Cos(p.toRad(args[0])), nil
	case "tan":
		if err := argc(1); err != nil {
			return 0, err
		}
		return math.Tan(p.toRad(args[0])), nil
	case "asin", "arcsin":
		if err := argc(1); err != nil {
			return 0, err
		}
		return p.fromRad(math.Asin(args[0])), nil
	case "acos", "arccos":
		if err := argc(1); err != nil {
			return 0, err
		}
		return p.fromRad(math.Acos(args[0])), nil
	case "atan", "arctan":
		if err := argc(1); err != nil {
			return 0, err
		}
		return p.fromRad(math.Atan(args[0])), nil
	case "atan2":
		if err := argc(2); err != nil {
			return 0, err
		}
		return p.fromRad(math.Atan2(args[0], args[1])), nil
	case "sinh":
		if err := argc(1); err != nil {
			return 0, err
		}
		return math.Sinh(args[0]), nil
	case "cosh":
		if err := argc(1); err != nil {
			return 0, err
		}
		return math.Cosh(args[0]), nil
	case "tanh":
		if err := argc(1); err != nil {
			return 0, err
		}
		return math.Tanh(args[0]), nil
	case "pow":
		if err := argc(2); err != nil {
			return 0, err
		}
		return math.Pow(args[0], args[1]), nil
	case "min":
		if len(args) < 2 {
			return 0, fmt.Errorf("min() requires at least 2 arguments")
		}
		m := args[0]
		for _, v := range args[1:] {
			if v < m {
				m = v
			}
		}
		return m, nil
	case "max":
		if len(args) < 2 {
			return 0, fmt.Errorf("max() requires at least 2 arguments")
		}
		m := args[0]
		for _, v := range args[1:] {
			if v > m {
				m = v
			}
		}
		return m, nil
	case "gcd":
		if err := argc(2); err != nil {
			return 0, err
		}
		a, b := math.Abs(math.Round(args[0])), math.Abs(math.Round(args[1]))
		for b != 0 {
			a, b = b, math.Mod(a, b)
		}
		return a, nil
	case "lcm":
		if err := argc(2); err != nil {
			return 0, err
		}
		a, b := math.Abs(math.Round(args[0])), math.Abs(math.Round(args[1]))
		if a == 0 || b == 0 {
			return 0, nil
		}
		g := a
		bb := b
		for bb != 0 {
			g, bb = bb, math.Mod(g, bb)
		}
		return a * b / g, nil
	case "hypot":
		if err := argc(2); err != nil {
			return 0, err
		}
		return math.Hypot(args[0], args[1]), nil
	case "clamp":
		if err := argc(3); err != nil {
			return 0, err
		}
		v, lo, hi := args[0], args[1], args[2]
		if v < lo {
			return lo, nil
		}
		if v > hi {
			return hi, nil
		}
		return v, nil
	case "lerp":
		if err := argc(3); err != nil {
			return 0, err
		}
		a, b, t := args[0], args[1], args[2]
		return a + (b-a)*t, nil
	case "factorial", "fact":
		if err := argc(1); err != nil {
			return 0, err
		}
		n := math.Round(args[0])
		if n < 0 {
			return 0, fmt.Errorf("factorial of negative number")
		}
		if n > 170 {
			return math.Inf(1), nil
		}
		result := 1.0
		for i := 2.0; i <= n; i++ {
			result *= i
		}
		return result, nil
	default:
		return 0, fmt.Errorf("unknown function %q", name)
	}
}

// evalExpr evaluates an expression string with optional variables.
func evalExpr(expr string, vars map[string]float64, degrees bool) (float64, error) {
	tokens, err := tokenize(expr)
	if err != nil {
		return 0, err
	}
	p := &mathParser{tokens: tokens, vars: vars, degrees: degrees}
	result, err := p.parseExpr()
	if err != nil {
		return 0, err
	}
	if p.peek().kind != tokEOF {
		return 0, fmt.Errorf("unexpected token %q after expression", p.peek().text)
	}
	return result, nil
}

// formatFloat formats a float with up to `prec` significant decimal places.
func formatFloat(v float64, prec int) string {
	if math.IsInf(v, 1) {
		return "∞"
	}
	if math.IsInf(v, -1) {
		return "-∞"
	}
	if math.IsNaN(v) {
		return "NaN"
	}
	if prec <= 0 || prec > 15 {
		prec = 10
	}
	s := strconv.FormatFloat(v, 'f', prec, 64)
	// Trim trailing zeros after decimal point
	if strings.Contains(s, ".") {
		s = strings.TrimRight(s, "0")
		s = strings.TrimRight(s, ".")
	}
	return s
}

func (s *Server) MathEval(_ context.Context, req *pb.MathEvalRequest) (*pb.MathEvalResponse, error) {
	expr := strings.TrimSpace(req.Expression)
	if expr == "" {
		return &pb.MathEvalResponse{Error: "expression is required"}, nil
	}

	vars := make(map[string]float64, len(req.Variables))
	for _, v := range req.Variables {
		vars[v.Name] = v.Value
	}

	result, err := evalExpr(expr, vars, req.Degrees)
	if err != nil {
		return &pb.MathEvalResponse{Error: err.Error()}, nil
	}

	prec := int(req.Precision)
	if prec <= 0 {
		prec = 10
	}

	return &pb.MathEvalResponse{
		Result:   formatFloat(result, prec),
		RawValue: result,
	}, nil
}

// ─── Percentage calculator ────────────────────────────────────────────────────

func (s *Server) PercentageCalc(_ context.Context, req *pb.PercentageRequest) (*pb.PercentageResponse, error) {
	a, b := req.A, req.B

	var result float64
	var formatted, formula string

	switch req.Mode {
	case pb.PercentMode_PCT_X_OF_Y:
		// What is A% of B?
		if b == 0 {
			return &pb.PercentageResponse{Error: "B cannot be zero"}, nil
		}
		result = (a / 100) * b
		formula = fmt.Sprintf("(%s / 100) × %s", fmtNum(a), fmtNum(b))
		formatted = fmt.Sprintf("%s%% of %s = %s", fmtNum(a), fmtNum(b), fmtNum(result))

	case pb.PercentMode_PCT_WHAT:
		// A is what % of B?
		if b == 0 {
			return &pb.PercentageResponse{Error: "B cannot be zero"}, nil
		}
		result = (a / b) * 100
		formula = fmt.Sprintf("(%s / %s) × 100", fmtNum(a), fmtNum(b))
		formatted = fmt.Sprintf("%s is %s%% of %s", fmtNum(a), fmtNum(result), fmtNum(b))

	case pb.PercentMode_PCT_CHANGE:
		// % change from A to B
		if a == 0 {
			return &pb.PercentageResponse{Error: "starting value (A) cannot be zero"}, nil
		}
		result = ((b - a) / math.Abs(a)) * 100
		formula = fmt.Sprintf("((%s − %s) / |%s|) × 100", fmtNum(b), fmtNum(a), fmtNum(a))
		dir := "increase"
		if result < 0 {
			dir = "decrease"
		}
		formatted = fmt.Sprintf("%s → %s = %s%% %s", fmtNum(a), fmtNum(b), fmtNum(result), dir)

	case pb.PercentMode_PCT_REVERSE:
		// A is B% of what?
		if b == 0 {
			return &pb.PercentageResponse{Error: "percentage (B) cannot be zero"}, nil
		}
		result = a / (b / 100)
		formula = fmt.Sprintf("%s / (%s / 100)", fmtNum(a), fmtNum(b))
		formatted = fmt.Sprintf("%s is %s%% of %s", fmtNum(a), fmtNum(b), fmtNum(result))

	default:
		return &pb.PercentageResponse{Error: "unknown mode"}, nil
	}

	return &pb.PercentageResponse{
		Result:    result,
		Formatted: formatted,
		Formula:   formula,
	}, nil
}

// fmtNum formats a float64 nicely for display in percentage formulas.
func fmtNum(v float64) string {
	if v == math.Trunc(v) {
		return strconv.FormatFloat(v, 'f', 0, 64)
	}
	s := strconv.FormatFloat(v, 'f', 8, 64)
	s = strings.TrimRight(s, "0")
	return strings.TrimRight(s, ".")
}

// ─── Temperature converter ────────────────────────────────────────────────────

func (s *Server) TempConvert(_ context.Context, req *pb.TempConvertRequest) (*pb.TempConvertResponse, error) {
	from := strings.ToLower(strings.TrimSpace(req.FromUnit))
	v := req.Value

	var celsius, fahrenheit, kelvin float64

	switch from {
	case "c", "celsius":
		celsius = v
		fahrenheit = v*9/5 + 32
		kelvin = v + 273.15
	case "f", "fahrenheit":
		celsius = (v - 32) * 5 / 9
		fahrenheit = v
		kelvin = (v-32)*5/9 + 273.15
	case "k", "kelvin":
		if v < 0 {
			return &pb.TempConvertResponse{Error: "Kelvin cannot be negative"}, nil
		}
		celsius = v - 273.15
		fahrenheit = (v-273.15)*9/5 + 32
		kelvin = v
	default:
		return &pb.TempConvertResponse{Error: fmt.Sprintf("unknown unit %q — use c, f, or k", from)}, nil
	}

	return &pb.TempConvertResponse{
		Celsius:    celsius,
		Fahrenheit: fahrenheit,
		Kelvin:     kelvin,
	}, nil
}

// ─── Unit converter ───────────────────────────────────────────────────────────

type unitDef struct {
	unit   string
	label  string
	toBase float64 // multiply input by this to get the base unit
}

// byteDefs: all units with byte as base.
var siUnits = []unitDef{
	{"B", "Byte", 1},
	{"KB", "Kilobyte", 1e3},
	{"MB", "Megabyte", 1e6},
	{"GB", "Gigabyte", 1e9},
	{"TB", "Terabyte", 1e12},
	{"PB", "Petabyte", 1e15},
	{"EB", "Exabyte", 1e18},
}

var binaryUnits = []unitDef{
	{"B", "Byte", 1},
	{"KiB", "Kibibyte", 1024},
	{"MiB", "Mebibyte", 1024 * 1024},
	{"GiB", "Gibibyte", 1024 * 1024 * 1024},
	{"TiB", "Tebibyte", 1024 * 1024 * 1024 * 1024},
	{"PiB", "Pebibyte", 1024 * 1024 * 1024 * 1024 * 1024},
	{"EiB", "Exbibyte", 1024 * 1024 * 1024 * 1024 * 1024 * 1024},
}

// Length: base = metre
var lengthUnits = []unitDef{
	{"nm", "Nanometre", 1e-9},
	{"µm", "Micrometre", 1e-6},
	{"mm", "Millimetre", 1e-3},
	{"cm", "Centimetre", 1e-2},
	{"dm", "Decimetre", 1e-1},
	{"m", "Metre", 1},
	{"km", "Kilometre", 1e3},
	{"in", "Inch", 0.0254},
	{"ft", "Foot", 0.3048},
	{"yd", "Yard", 0.9144},
	{"mi", "Mile", 1609.344},
	{"nmi", "Nautical mile", 1852},
	{"ly", "Light-year", 9.461e15},
}

// Mass: base = gram
var massUnits = []unitDef{
	{"µg", "Microgram", 1e-6},
	{"mg", "Milligram", 1e-3},
	{"g", "Gram", 1},
	{"kg", "Kilogram", 1e3},
	{"t", "Metric ton", 1e6},
	{"oz", "Ounce", 28.3495},
	{"lb", "Pound", 453.592},
	{"st", "Stone", 6350.29},
	{"ton", "Short ton (US)", 907185},
	{"lt", "Long ton (UK)", 1016047},
}

// Area: base = m²
var areaUnits = []unitDef{
	{"mm²", "Square millimetre", 1e-6},
	{"cm²", "Square centimetre", 1e-4},
	{"m²", "Square metre", 1},
	{"km²", "Square kilometre", 1e6},
	{"ha", "Hectare", 1e4},
	{"in²", "Square inch", 6.4516e-4},
	{"ft²", "Square foot", 0.092903},
	{"yd²", "Square yard", 0.836127},
	{"ac", "Acre", 4046.86},
	{"mi²", "Square mile", 2.59e6},
}

// Volume: base = litre
var volumeUnits = []unitDef{
	{"ml", "Millilitre", 1e-3},
	{"cl", "Centilitre", 1e-2},
	{"dl", "Decilitre", 1e-1},
	{"l", "Litre", 1},
	{"m³", "Cubic metre", 1000},
	{"in³", "Cubic inch", 0.016387},
	{"ft³", "Cubic foot", 28.3168},
	{"tsp", "Teaspoon (US)", 0.00492892},
	{"tbsp", "Tablespoon (US)", 0.0147868},
	{"fl oz", "Fluid ounce (US)", 0.0295735},
	{"cup", "Cup (US)", 0.236588},
	{"pt", "Pint (US)", 0.473176},
	{"qt", "Quart (US)", 0.946353},
	{"gal", "Gallon (US)", 3.78541},
	{"imp gal", "Gallon (Imperial)", 4.54609},
}

// Speed: base = m/s
var speedUnits = []unitDef{
	{"m/s", "Metre per second", 1},
	{"km/h", "Kilometre per hour", 1.0 / 3.6},
	{"mph", "Mile per hour", 0.44704},
	{"ft/s", "Foot per second", 0.3048},
	{"kn", "Knot", 0.514444},
	{"mach", "Mach (at sea level)", 343},
	{"c", "Speed of light", 299792458},
}

func unitsByCategory(cat pb.UnitCategory) []unitDef {
	switch cat {
	case pb.UnitCategory_UNIT_BYTES:
		// Special-cased: return combined SI + binary
		return nil
	case pb.UnitCategory_UNIT_LENGTH:
		return lengthUnits
	case pb.UnitCategory_UNIT_MASS:
		return massUnits
	case pb.UnitCategory_UNIT_AREA:
		return areaUnits
	case pb.UnitCategory_UNIT_VOLUME:
		return volumeUnits
	case pb.UnitCategory_UNIT_SPEED:
		return speedUnits
	}
	return nil
}

func findUnit(units []unitDef, name string) (unitDef, bool) {
	nameLow := strings.ToLower(strings.TrimSpace(name))
	for _, u := range units {
		if strings.ToLower(u.unit) == nameLow || strings.ToLower(u.label) == nameLow {
			return u, true
		}
	}
	return unitDef{}, false
}

func formatUnitValue(v float64) string {
	if math.IsInf(v, 0) || math.IsNaN(v) {
		return "—"
	}
	abs := math.Abs(v)
	switch {
	case abs == 0:
		return "0"
	case abs >= 1e15 || (abs < 1e-9 && abs > 0):
		return strconv.FormatFloat(v, 'e', 6, 64)
	case abs >= 1:
		s := strconv.FormatFloat(v, 'f', 8, 64)
		s = strings.TrimRight(s, "0")
		return strings.TrimRight(s, ".")
	default:
		s := strconv.FormatFloat(v, 'f', 12, 64)
		s = strings.TrimRight(s, "0")
		return strings.TrimRight(s, ".")
	}
}

func (s *Server) UnitConvert(_ context.Context, req *pb.UnitConvertRequest) (*pb.UnitConvertResponse, error) {
	if req.Category == pb.UnitCategory_UNIT_BYTES {
		return s.convertBytes(req.Value, req.FromUnit)
	}

	units := unitsByCategory(req.Category)
	from, ok := findUnit(units, req.FromUnit)
	if !ok {
		// Build valid unit list for the error message
		names := make([]string, len(units))
		for i, u := range units {
			names[i] = u.unit
		}
		return &pb.UnitConvertResponse{
			Error: fmt.Sprintf("unknown unit %q — valid units: %s", req.FromUnit, strings.Join(names, ", ")),
		}, nil
	}

	// Convert input to base unit, then to all other units
	baseValue := req.Value * from.toBase

	results := make([]*pb.UnitResult, 0, len(units))
	for _, u := range units {
		converted := baseValue / u.toBase
		results = append(results, &pb.UnitResult{
			Unit:      u.unit,
			Label:     u.label,
			Value:     converted,
			Formatted: formatUnitValue(converted),
		})
	}

	return &pb.UnitConvertResponse{Results: results}, nil
}

func (s *Server) convertBytes(value float64, fromUnit string) (*pb.UnitConvertResponse, error) {
	// Build combined lookup
	allUnits := append(siUnits, binaryUnits...) // nolint:gocritic
	from, ok := findUnit(allUnits, fromUnit)
	if !ok {
		names := make([]string, 0, len(allUnits))
		seen := map[string]bool{}
		for _, u := range allUnits {
			if !seen[u.unit] {
				seen[u.unit] = true
				names = append(names, u.unit)
			}
		}
		sort.Strings(names)
		return &pb.UnitConvertResponse{
			Error: fmt.Sprintf("unknown byte unit %q — valid units: %s", fromUnit, strings.Join(names, ", ")),
		}, nil
	}

	bytes := value * from.toBase

	results := make([]*pb.UnitResult, 0, len(siUnits)+len(binaryUnits))
	for _, u := range siUnits {
		converted := bytes / u.toBase
		results = append(results, &pb.UnitResult{
			Unit:      u.unit,
			Label:     u.label + " (SI)",
			Value:     converted,
			Formatted: formatUnitValue(converted),
		})
	}
	for _, u := range binaryUnits {
		converted := bytes / u.toBase
		results = append(results, &pb.UnitResult{
			Unit:      u.unit,
			Label:     u.label + " (binary)",
			Value:     converted,
			Formatted: formatUnitValue(converted),
		})
	}

	return &pb.UnitConvertResponse{Results: results}, nil
}

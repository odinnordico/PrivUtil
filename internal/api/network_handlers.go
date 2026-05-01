package api

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"math/big"
	"net"
	"regexp"
	"slices"
	"strconv"
	"strings"

	pb "github.com/odinnordico/privutil/proto"
)

// ─── chmod ────────────────────────────────────────────────────────────────────

var (
	octalRe      = regexp.MustCompile(`^0?[0-7]{1,4}$`)
	symbolic9Re  = regexp.MustCompile(`^[r-][w-][xsS-][r-][w-][xsS-][r-][w-][xtT-]$`)
	symbolic10Re = regexp.MustCompile(`^[-dlcbps][r-][w-][xsS-][r-][w-][xsS-][r-][w-][xtT-]$`)
)

func (s *Server) ChmodCalc(_ context.Context, req *pb.ChmodRequest) (*pb.ChmodResponse, error) {
	input := strings.TrimSpace(req.Input)
	if input == "" {
		return &pb.ChmodResponse{Error: "input is required"}, nil
	}

	var mode uint32
	switch {
	case octalRe.MatchString(input):
		v, err := strconv.ParseUint(input, 8, 32)
		if err != nil {
			return &pb.ChmodResponse{Error: "invalid octal: " + err.Error()}, nil
		}
		mode = uint32(v)
	case symbolic9Re.MatchString(input):
		mode = chmodFromSymbolic(input)
	case symbolic10Re.MatchString(input):
		mode = chmodFromSymbolic(input[1:])
	default:
		return &pb.ChmodResponse{Error: `invalid format — use octal (e.g. "755") or symbolic (e.g. "rwxr-xr-x")`}, nil
	}

	return buildChmodResponse(mode), nil
}

func chmodFromSymbolic(s string) uint32 {
	// s is exactly 9 chars: owner(rwx) group(rwx) other(rwt)
	var mode uint32
	perm := [9]struct {
		bit  uint32
		char byte
	}{
		{0o400, 'r'}, {0o200, 'w'}, {0o100, 'x'},
		{0o040, 'r'}, {0o020, 'w'}, {0o010, 'x'},
		{0o004, 'r'}, {0o002, 'w'}, {0o001, 'x'},
	}
	for i, p := range perm {
		if s[i] == p.char {
			mode |= p.bit
		}
	}
	// Special bits
	switch s[2] {
	case 's':
		mode |= 0o4100 // setuid + execute
	case 'S':
		mode |= 0o4000 // setuid only
	}
	switch s[5] {
	case 's':
		mode |= 0o2010 // setgid + execute
	case 'S':
		mode |= 0o2000 // setgid only
	}
	switch s[8] {
	case 't':
		mode |= 0o1001 // sticky + execute
	case 'T':
		mode |= 0o1000 // sticky only
	}
	return mode
}

func buildChmodResponse(mode uint32) *pb.ChmodResponse {
	r := &pb.ChmodResponse{
		Value:        int32(mode), // #nosec G115
		Octal:        fmt.Sprintf("%o", mode),
		OwnerRead:    mode&0o400 != 0,
		OwnerWrite:   mode&0o200 != 0,
		OwnerExecute: mode&0o100 != 0,
		GroupRead:    mode&0o040 != 0,
		GroupWrite:   mode&0o020 != 0,
		GroupExecute: mode&0o010 != 0,
		OtherRead:    mode&0o004 != 0,
		OtherWrite:   mode&0o002 != 0,
		OtherExecute: mode&0o001 != 0,
		Setuid:       mode&0o4000 != 0,
		Setgid:       mode&0o2000 != 0,
		Sticky:       mode&0o1000 != 0,
	}

	// Symbolic notation
	sym := [9]byte{'-', '-', '-', '-', '-', '-', '-', '-', '-'}
	if r.OwnerRead {
		sym[0] = 'r'
	}
	if r.OwnerWrite {
		sym[1] = 'w'
	}
	switch {
	case r.Setuid && r.OwnerExecute:
		sym[2] = 's'
	case r.Setuid:
		sym[2] = 'S'
	case r.OwnerExecute:
		sym[2] = 'x'
	}
	if r.GroupRead {
		sym[3] = 'r'
	}
	if r.GroupWrite {
		sym[4] = 'w'
	}
	switch {
	case r.Setgid && r.GroupExecute:
		sym[5] = 's'
	case r.Setgid:
		sym[5] = 'S'
	case r.GroupExecute:
		sym[5] = 'x'
	}
	if r.OtherRead {
		sym[6] = 'r'
	}
	if r.OtherWrite {
		sym[7] = 'w'
	}
	switch {
	case r.Sticky && r.OtherExecute:
		sym[8] = 't'
	case r.Sticky:
		sym[8] = 'T'
	case r.OtherExecute:
		sym[8] = 'x'
	}
	r.Symbolic = string(sym[:])

	// Human-readable description
	var parts []string
	if ops := chmodPermDesc("owner", r.OwnerRead, r.OwnerWrite, r.OwnerExecute); ops != "" {
		parts = append(parts, ops)
	}
	if ops := chmodPermDesc("group", r.GroupRead, r.GroupWrite, r.GroupExecute); ops != "" {
		parts = append(parts, ops)
	}
	if ops := chmodPermDesc("others", r.OtherRead, r.OtherWrite, r.OtherExecute); ops != "" {
		parts = append(parts, ops)
	}
	if r.Setuid {
		parts = append(parts, "setuid")
	}
	if r.Setgid {
		parts = append(parts, "setgid")
	}
	if r.Sticky {
		parts = append(parts, "sticky bit")
	}
	if len(parts) == 0 {
		r.Description = "no permissions"
	} else {
		r.Description = strings.Join(parts, "; ")
	}
	return r
}

func chmodPermDesc(who string, read, write, execute bool) string {
	var ops []string
	if read {
		ops = append(ops, "read")
	}
	if write {
		ops = append(ops, "write")
	}
	if execute {
		ops = append(ops, "execute")
	}
	if len(ops) == 0 {
		return ""
	}
	return who + ": " + strings.Join(ops, "+")
}

// ─── IPv4 converter ───────────────────────────────────────────────────────────

var (
	binaryIPRe  = regexp.MustCompile(`^[01]{32}$`)              // exactly 32 binary digits
	binaryDotRe = regexp.MustCompile(`^[01]{8}(\.[01]{8}){3}$`) // exactly 8 bits per octet
	hexNoPrefix = regexp.MustCompile(`(?i)^[0-9a-f]{8}$`)       // exactly 8 hex digits
)

func (s *Server) Ipv4Convert(_ context.Context, req *pb.Ipv4ConvertRequest) (*pb.Ipv4ConvertResponse, error) {
	input := strings.TrimSpace(req.Input)
	if input == "" {
		return &pb.Ipv4ConvertResponse{Error: "input is required"}, nil
	}

	v, err := parseAnyIPv4(input)
	if err != nil {
		return &pb.Ipv4ConvertResponse{Error: err.Error()}, nil
	}

	ip := make(net.IP, 4)
	binary.BigEndian.PutUint32(ip, v)

	octets := []byte(ip)
	binParts := make([]string, 4)
	for i, b := range octets {
		binParts[i] = fmt.Sprintf("%08b", b)
	}

	return &pb.Ipv4ConvertResponse{
		Dotted:  ip.String(),
		Decimal: strconv.FormatUint(uint64(v), 10),
		Hex:     fmt.Sprintf("0x%08X", v),
		Binary:  strings.Join(binParts, "."),
	}, nil
}

func parseAnyIPv4(input string) (uint32, error) {
	// 1. Dotted input (3 dots): try dotted decimal first, then binary-dotted
	if strings.Count(input, ".") == 3 {
		if ip := net.ParseIP(input); ip != nil {
			ip4 := ip.To4()
			if ip4 == nil {
				return 0, fmt.Errorf("not an IPv4 address")
			}
			return binary.BigEndian.Uint32(ip4), nil
		}
		if binaryDotRe.MatchString(input) {
			parts := strings.Split(input, ".")
			var result uint32
			for _, p := range parts {
				b, err := strconv.ParseUint(p, 2, 8)
				if err != nil {
					return 0, fmt.Errorf("invalid binary octet %q", p)
				}
				result = (result << 8) | uint32(b)
			}
			return result, nil
		}
		return 0, fmt.Errorf("invalid dotted IPv4")
	}

	// 2. Explicit hex prefix
	if strings.HasPrefix(strings.ToLower(input), "0x") {
		clean := input[2:]
		if matched, _ := regexp.MatchString(`(?i)^[0-9a-f]{1,8}$`, clean); matched {
			v, err := strconv.ParseUint(clean, 16, 64)
			if err == nil {
				return uint32(v), nil
			}
		}
		return 0, fmt.Errorf("invalid hex IPv4")
	}

	// 3. Exactly 32 binary digits
	if binaryIPRe.MatchString(input) {
		v, err := strconv.ParseUint(input, 2, 32)
		if err == nil {
			return uint32(v), nil
		}
	}

	// 4. Exactly 8 hex digits containing at least one a-f letter (unambiguous hex)
	if hexNoPrefix.MatchString(input) && strings.ContainsAny(strings.ToLower(input), "abcdef") {
		v, err := strconv.ParseUint(input, 16, 64)
		if err == nil {
			return uint32(v), nil
		}
	}

	// 5. Decimal integer
	if v, err := strconv.ParseUint(input, 10, 64); err == nil && v <= 0xFFFFFFFF {
		return uint32(v), nil
	}

	return 0, fmt.Errorf("unrecognized format — use dotted decimal, integer, hex (0x...) or binary")
}

// ─── IPv4 range expander ──────────────────────────────────────────────────────

const maxIPList = 256

func (s *Server) Ipv4RangeExpand(_ context.Context, req *pb.Ipv4RangeRequest) (*pb.Ipv4RangeResponse, error) {
	startIP, err := parseAnyIPv4(strings.TrimSpace(req.Start))
	if err != nil {
		return &pb.Ipv4RangeResponse{Error: "start: " + err.Error()}, nil
	}
	endIP, err := parseAnyIPv4(strings.TrimSpace(req.End))
	if err != nil {
		return &pb.Ipv4RangeResponse{Error: "end: " + err.Error()}, nil
	}
	if startIP > endIP {
		return &pb.Ipv4RangeResponse{Error: "start must be ≤ end"}, nil
	}

	total := int64(endIP-startIP) + 1
	cidrs := rangeToCIDRs(startIP, endIP)

	var addrs []string
	if total <= maxIPList {
		addrs = make([]string, total)
		buf := make(net.IP, 4)
		for i := range total {
			binary.BigEndian.PutUint32(buf, startIP+uint32(i))
			addrs[i] = buf.String()
		}
	}

	return &pb.Ipv4RangeResponse{
		Addresses: addrs,
		Cidrs:     cidrs,
		Total:     total,
	}, nil
}

func rangeToCIDRs(start, end uint32) []string {
	var result []string
	for start <= end {
		bits := 32
		for bits > 0 {
			newBits := bits - 1
			blockSize := uint64(1) << uint(32-newBits)
			if uint64(start)&(blockSize-1) != 0 {
				break
			}
			if uint64(start)+blockSize-1 > uint64(end) {
				break
			}
			bits = newBits
		}
		blockSize := uint64(1) << uint(32-bits)
		ip := make(net.IP, 4)
		binary.BigEndian.PutUint32(ip, start)
		result = append(result, fmt.Sprintf("%s/%d", ip.String(), bits))
		next := uint64(start) + blockSize
		if next > 0xFFFFFFFF {
			break
		}
		start = uint32(next)
	}
	return result
}

// ─── port generator ───────────────────────────────────────────────────────────

const (
	maxPortGenCount = 100
	sysPortMax      = 1023
	portMax         = 65535
)

func (s *Server) GeneratePort(_ context.Context, req *pb.PortRequest) (*pb.PortResponse, error) {
	count := int(req.Count)
	if count <= 0 {
		count = 1
	}
	if count > maxPortGenCount {
		count = maxPortGenCount
	}

	lo := int(req.Min)
	hi := int(req.Max)
	if lo <= 0 {
		lo = 0
	}
	if hi <= 0 || hi > portMax {
		hi = portMax
	}
	if req.ExcludeSystem && lo <= sysPortMax {
		lo = sysPortMax + 1
	}
	if lo > hi {
		return &pb.PortResponse{Error: fmt.Sprintf("no valid ports in range %d–%d", lo, hi)}, nil
	}

	span := int64(hi - lo + 1)
	seen := make(map[int32]struct{}, count)
	ports := make([]int32, 0, count)

	maxAttempts := count * 10
	for len(ports) < count && maxAttempts > 0 {
		maxAttempts--
		n, err := rand.Int(rand.Reader, big.NewInt(span))
		if err != nil {
			return &pb.PortResponse{Error: "random generation failed"}, nil
		}
		p := int32(lo) + int32(n.Int64()) // #nosec G115
		if _, dup := seen[p]; !dup {
			seen[p] = struct{}{}
			ports = append(ports, p)
		}
	}
	slices.Sort(ports)
	return &pb.PortResponse{Ports: ports}, nil
}

// ─── MAC address generator ────────────────────────────────────────────────────

const maxMacCount = 100

func (s *Server) GenerateMac(_ context.Context, req *pb.MacRequest) (*pb.MacResponse, error) {
	count := int(req.Count)
	if count <= 0 {
		count = 1
	}
	if count > maxMacCount {
		count = maxMacCount
	}

	sep := req.Separator
	if sep == "" {
		sep = ":"
	}

	// Parse OUI if provided
	var ouiBytes []byte
	if req.Oui != "" {
		b, err := parseMACBytes(req.Oui)
		if err != nil || len(b) != 3 {
			return &pb.MacResponse{Error: "invalid OUI — expected 3 bytes e.g. 00:1A:2B"}, nil
		}
		ouiBytes = b
	}

	addrs := make([]string, count)
	for i := range addrs {
		mac := make([]byte, 6)
		if len(ouiBytes) == 3 {
			copy(mac[:3], ouiBytes)
		} else {
			if _, err := rand.Read(mac[:3]); err != nil {
				return &pb.MacResponse{Error: "random generation failed"}, nil
			}
		}
		if _, err := rand.Read(mac[3:]); err != nil {
			return &pb.MacResponse{Error: "random generation failed"}, nil
		}

		// Apply unicast / locally-administered flags on first byte
		if req.Unicast {
			mac[0] &^= 0x01 // clear multicast bit
		}
		if req.Local {
			mac[0] |= 0x02 // set locally-administered bit
		}

		addrs[i] = formatMAC(mac, sep, req.Uppercase)
	}
	return &pb.MacResponse{Addresses: addrs}, nil
}

func parseMACBytes(s string) ([]byte, error) {
	s = strings.ReplaceAll(s, ":", "")
	s = strings.ReplaceAll(s, "-", "")
	s = strings.ReplaceAll(s, ".", "")
	if len(s)%2 != 0 {
		return nil, fmt.Errorf("odd length hex string")
	}
	out := make([]byte, len(s)/2)
	for i := range out {
		b, err := strconv.ParseUint(s[i*2:i*2+2], 16, 8)
		if err != nil {
			return nil, err
		}
		out[i] = byte(b)
	}
	return out, nil
}

func formatMAC(mac []byte, sep string, upper bool) string {
	if sep == "." {
		// Cisco notation: XXXX.XXXX.XXXX
		parts := make([]string, 3)
		for i := range parts {
			group := mac[i*2 : i*2+2]
			parts[i] = fmt.Sprintf("%02x%02x", group[0], group[1])
		}
		result := strings.Join(parts, ".")
		if upper {
			return strings.ToUpper(result)
		}
		return result
	}

	parts := make([]string, 6)
	for i, b := range mac {
		parts[i] = fmt.Sprintf("%02x", b)
	}
	result := strings.Join(parts, sep)
	if upper {
		return strings.ToUpper(result)
	}
	return result
}

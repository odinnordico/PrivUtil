package api

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/gorhill/cronexpr"

	pb "github.com/odinnordico/privutil/proto"
)

func (s *Server) TimeConvert(ctx context.Context, req *pb.TimeRequest) (*pb.TimeResponse, error) {
	input := strings.TrimSpace(req.Input)
	var t time.Time

	if input == "" || strings.EqualFold(input, "now") {
		t = time.Now()
	} else {
		if ts, err := strconv.ParseInt(input, 10, 64); err == nil {
			if ts > 10000000000 {
				t = time.UnixMilli(ts)
			} else {
				t = time.Unix(ts, 0)
			}
		} else {
			formats := []string{
				time.RFC3339,
				time.RFC3339Nano,
				time.Layout,
				"2006-01-02T15:04:05",
				"2006-01-02 15:04:05",
				time.RubyDate,
				time.UnixDate,
			}
			var parsed bool
			for _, layout := range formats {
				if pt, err := time.Parse(layout, input); err == nil {
					t = pt
					parsed = true
					break
				}
			}
			if !parsed {
				return &pb.TimeResponse{Iso: "Invalid input format"}, nil
			}
		}
	}

	return &pb.TimeResponse{
		Unix:  t.Unix(),
		Utc:   t.UTC().Format(time.RFC3339),
		Local: t.Local().Format("2006-01-02 15:04:05 -0700 MST"),
		Iso:   t.Format(time.RFC3339),
	}, nil
}

func (s *Server) CronExplain(ctx context.Context, req *pb.CronRequest) (*pb.CronResponse, error) {
	expr, err := cronexpr.Parse(req.Expression)
	if err != nil {
		return &pb.CronResponse{Error: fmt.Sprintf("Invalid cron expression: %v", err)}, nil
	}

	nextTimes := expr.NextN(time.Now(), 5)
	var nextRuns []string
	for _, t := range nextTimes {
		nextRuns = append(nextRuns, t.Format(time.RFC3339))
	}

	desc := describeCron(req.Expression)

	return &pb.CronResponse{
		Description: desc,
		NextRuns:    strings.Join(nextRuns, "\n"),
	}, nil
}

func (s *Server) IpCalc(ctx context.Context, req *pb.IpRequest) (*pb.IpResponse, error) {
	ip, ipnet, err := net.ParseCIDR(req.Cidr)
	if err != nil {
		if ip2 := net.ParseIP(req.Cidr); ip2 != nil {
			if ip2.To4() != nil {
				ip, ipnet, _ = net.ParseCIDR(req.Cidr + "/32")
			} else {
				ip, ipnet, _ = net.ParseCIDR(req.Cidr + "/128")
			}
		}

		if ip == nil {
			return &pb.IpResponse{Error: "Invalid IP or CIDR"}, nil
		}
	}

	ones, bits := ipnet.Mask.Size()
	network := ipnet.IP
	netmask := net.IP(ipnet.Mask)

	ipLen := len(ipnet.Mask)
	lastIP := make(net.IP, ipLen)
	for i := range ipLen {
		lastIP[i] = network[i] | ^ipnet.Mask[i]
	}

	var broadcastStr string
	if ip.To4() != nil {
		broadcastStr = lastIP.String()
	}

	var count int64
	if bits-ones < 63 {
		count = 1 << uint(bits-ones) // #nosec G115
	}

	return &pb.IpResponse{
		Network:   network.String(),
		Broadcast: broadcastStr,
		Netmask:   netmask.String(),
		NumHosts:  count,
		FirstIp:   network.String(),
		LastIp:    lastIP.String(),
	}, nil
}

func describeCron(expr string) string {
	parts := strings.Fields(expr)
	var min, hour, dom, month, dow string
	switch len(parts) {
	case 5:
		min, hour, dom, month, dow = parts[0], parts[1], parts[2], parts[3], parts[4]
	case 6:
		// 6-field form: seconds minute hour dom month dow
		min, hour, dom, month, dow = parts[1], parts[2], parts[3], parts[4], parts[5]
	default:
		return "Invalid cron expression format"
	}

	if min == "*" && hour == "*" && dom == "*" && month == "*" && dow == "*" {
		return "Every minute"
	}
	if strings.HasPrefix(min, "*/") && hour == "*" && dom == "*" && month == "*" && dow == "*" {
		return fmt.Sprintf("Every %s minutes", min[2:])
	}
	if min == "0" && hour == "*" && dom == "*" && month == "*" && dow == "*" {
		return "At the start of every hour"
	}
	if min == "0" && strings.HasPrefix(hour, "*/") && dom == "*" && month == "*" && dow == "*" {
		return fmt.Sprintf("At minute 0 past every %s hours", hour[2:])
	}
	if min == "0" && hour == "0" && dom == "*" && month == "*" && dow == "*" {
		return "At 00:00 every day"
	}

	desc := "Run "
	if min != "*" {
		desc += fmt.Sprintf("at minute %s", min)
	} else {
		desc += "every minute"
	}

	if hour != "*" {
		desc += fmt.Sprintf(" of hour %s", hour)
	}

	if dom != "*" {
		desc += fmt.Sprintf(" on day-of-month %s", dom)
	}

	if dow != "*" {
		desc += fmt.Sprintf(" on day-of-week %s", dow)
	}

	return desc
}

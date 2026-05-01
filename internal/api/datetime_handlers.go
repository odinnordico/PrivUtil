package api

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"
	_ "time/tzdata" // embed IANA timezone database

	pb "github.com/odinnordico/privutil/proto"
)

// ─── Date parsing ─────────────────────────────────────────────────────────────

var dateLayouts = []string{
	time.RFC3339,
	time.RFC3339Nano,
	"2006-01-02T15:04:05",
	"2006-01-02 15:04:05",
	"2006-01-02",
	"January 2, 2006",
	"January 2 2006",
	"Jan 2, 2006",
	"Jan 2 2006",
	"2 January 2006",
	"2 Jan 2006",
	"2006.01.02",
	"20060102",
	"01/02/2006",
	"1/2/2006",
	"02.01.2006",
}

func parseDate(s string) (time.Time, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return time.Time{}, fmt.Errorf("date is required")
	}

	now := time.Now().UTC()
	switch strings.ToLower(s) {
	case "today", "now":
		return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC), nil
	case "yesterday":
		t := now.AddDate(0, 0, -1)
		return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC), nil
	case "tomorrow":
		t := now.AddDate(0, 0, 1)
		return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC), nil
	}

	// Try "DD/MM/YYYY" only when first component > 12 (unambiguously a day)
	parts := strings.FieldsFunc(s, func(r rune) bool { return r == '/' || r == '-' })
	if len(parts) == 3 && len(parts[0]) <= 2 && len(parts[2]) == 4 {
		if d, err := strconv.Atoi(parts[0]); err == nil && d > 12 {
			if t, err := time.Parse("02/01/2006", strings.Join(parts, "/")); err == nil {
				return t, nil
			}
		}
	}

	for _, layout := range dateLayouts {
		if t, err := time.Parse(layout, s); err == nil {
			return t, nil
		}
	}

	// Try Unix timestamp
	if ts, err := strconv.ParseInt(s, 10, 64); err == nil {
		if ts > 1e10 {
			return time.UnixMilli(ts).UTC(), nil
		}
		return time.Unix(ts, 0).UTC(), nil
	}

	return time.Time{}, fmt.Errorf("cannot parse date %q — try YYYY-MM-DD or RFC 3339", s)
}

// ─── Calendar-aware diff ──────────────────────────────────────────────────────

// calendarDiff computes (years, months, days) from `from` to `to` (from ≤ to).
// Uses an iterative advance so edge cases like Jan 31 → Mar 1 are correct.
func calendarDiff(from, to time.Time) (years, months, days int64) {
	// Estimate full years — loop handles AddDate overflow (e.g. Feb 29 in non-leap years)
	y := int64(to.Year() - from.Year())
	for y > 0 && from.AddDate(int(y), 0, 0).After(to) {
		y--
	}
	years = y
	from = from.AddDate(int(y), 0, 0)

	// Estimate full months — loop because AddDate can overflow (e.g. Jan 31 + 1mo = Mar 2)
	m := int64(to.Month()) - int64(from.Month())
	if m < 0 {
		m += 12
	}
	for m > 0 && from.AddDate(0, int(m), 0).After(to) {
		m--
	}
	months = m
	from = from.AddDate(0, int(m), 0)

	// Remaining days (always non-negative after the loops above)
	days = int64(to.Sub(from).Hours() / 24)
	return years, months, days
}

func humanDiff(years, months, weeks, days, hours, minutes, seconds int64) string {
	var parts []string
	add := func(n int64, sing, plur string) {
		if n == 0 {
			return
		}
		if n == 1 {
			parts = append(parts, fmt.Sprintf("1 %s", sing))
		} else {
			parts = append(parts, fmt.Sprintf("%d %s", n, plur))
		}
	}
	add(years, "year", "years")
	add(months, "month", "months")
	// show weeks only if no years/months and days ≥ 7
	if years == 0 && months == 0 && weeks > 0 && days == 0 {
		add(weeks, "week", "weeks")
	}
	add(days, "day", "days")
	add(hours, "hour", "hours")
	add(minutes, "minute", "minutes")
	add(seconds, "second", "seconds")
	if len(parts) == 0 {
		return "0 seconds"
	}
	return strings.Join(parts, ", ")
}

// ─── Date diff handler ────────────────────────────────────────────────────────

func (s *Server) DateDiff(_ context.Context, req *pb.DateDiffRequest) (*pb.DateDiffResponse, error) {
	from, err := parseDate(req.FromDate)
	if err != nil {
		return &pb.DateDiffResponse{Error: "from: " + err.Error()}, nil
	}
	to, err := parseDate(req.ToDate)
	if err != nil {
		return &pb.DateDiffResponse{Error: "to: " + err.Error()}, nil
	}

	negative := from.After(to)
	if negative {
		from, to = to, from
	}

	// Calendar-level breakdown (date-only)
	fromDay := time.Date(from.Year(), from.Month(), from.Day(), 0, 0, 0, 0, time.UTC)
	toDay := time.Date(to.Year(), to.Month(), to.Day(), 0, 0, 0, 0, time.UTC)
	years, months, days := calendarDiff(fromDay, toDay)

	// Total metrics
	totalDur := to.Sub(from)
	totalSecs := int64(totalDur.Seconds())
	totalHours := totalSecs / 3600
	totalDays := int64(toDay.Sub(fromDay).Hours()) / 24
	totalWeeks := totalDays / 7

	// Sub-day time components from the remainder
	adjusted := from.AddDate(int(years), int(months), int(days))
	remaining := to.Sub(adjusted)
	if remaining < 0 {
		days--
		adjusted = from.AddDate(int(years), int(months), int(days))
		remaining = to.Sub(adjusted)
	}
	hrs := int64(remaining.Hours())
	mins := int64(remaining.Minutes()) % 60
	secs := int64(remaining.Seconds()) % 60

	human := humanDiff(years, months, totalWeeks, days, hrs, mins, secs)

	return &pb.DateDiffResponse{
		Years:      years,
		Months:     months,
		Weeks:      totalWeeks,
		Days:       days,
		Hours:      hrs,
		Minutes:    mins,
		Seconds:    secs,
		TotalDays:  totalDays,
		TotalHours: totalHours,
		TotalSecs:  totalSecs,
		Negative:   negative,
		Human:      human,
	}, nil
}

// ─── Leap year handler ────────────────────────────────────────────────────────

func isLeapYear(year int) bool {
	return (year%4 == 0 && year%100 != 0) || year%400 == 0
}

func (s *Server) LeapYear(_ context.Context, req *pb.LeapYearRequest) (*pb.LeapYearResponse, error) {
	input := strings.TrimSpace(req.Input)
	if input == "" {
		return &pb.LeapYearResponse{Error: "input is required"}, nil
	}

	var years []int

	// Check for range "YYYY-YYYY"
	if rangeParts := strings.SplitN(input, "-", 2); len(rangeParts) == 2 {
		start, err1 := strconv.Atoi(strings.TrimSpace(rangeParts[0]))
		end, err2 := strconv.Atoi(strings.TrimSpace(rangeParts[1]))
		if err1 == nil && err2 == nil && start > 0 && end >= start {
			if end-start > 400 {
				return &pb.LeapYearResponse{Error: "range too large — max 400 years"}, nil
			}
			for y := start; y <= end; y++ {
				years = append(years, y)
			}
		}
	}

	// Comma-separated list (also handles single year)
	if len(years) == 0 {
		for _, part := range strings.Split(input, ",") {
			part = strings.TrimSpace(part)
			if part == "" {
				continue
			}
			y, err := strconv.Atoi(part)
			if err != nil {
				return &pb.LeapYearResponse{Error: fmt.Sprintf("invalid year %q", part)}, nil
			}
			years = append(years, y)
		}
	}

	if len(years) == 0 {
		return &pb.LeapYearResponse{Error: "no valid years found"}, nil
	}

	results := make([]*pb.LeapYearEntry, len(years))
	leapCount := int32(0)
	for i, y := range years {
		leap := isLeapYear(y)
		if leap {
			leapCount++
		}
		results[i] = &pb.LeapYearEntry{
			Year:   int32(y), // #nosec G115
			IsLeap: leap,
		}
	}

	return &pb.LeapYearResponse{Results: results, LeapCount: leapCount}, nil
}

// ─── Date add handler ─────────────────────────────────────────────────────────

func (s *Server) DateAdd(_ context.Context, req *pb.DateAddRequest) (*pb.DateAddResponse, error) {
	t, err := parseDate(req.Date)
	if err != nil {
		return &pb.DateAddResponse{Error: err.Error()}, nil
	}

	t = t.AddDate(int(req.Years), int(req.Months), int(req.Weeks)*7+int(req.Days))
	t = t.Add(time.Duration(req.Hours)*time.Hour +
		time.Duration(req.Minutes)*time.Minute +
		time.Duration(req.Seconds)*time.Second)

	isoYear, isoWeek := t.ISOWeek()
	doy := t.YearDay()

	_ = isoYear // field not in response proto

	return &pb.DateAddResponse{
		Iso:       t.Format("2006-01-02"),
		IsoFull:   t.UTC().Format(time.RFC3339),
		Unix:      strconv.FormatInt(t.Unix(), 10),
		Weekday:   t.Weekday().String(),
		DayOfYear: int32(doy),     // #nosec G115
		IsoWeek:   int32(isoWeek), // #nosec G115
		Formatted: t.Format("Monday, 2 January 2006"),
	}, nil
}

// ─── Date formatter handler ───────────────────────────────────────────────────

func (s *Server) DateFormat(_ context.Context, req *pb.DateFormatRequest) (*pb.DateFormatResponse, error) {
	t, err := parseDate(req.DateStr)
	if err != nil {
		return &pb.DateFormatResponse{Error: err.Error()}, nil
	}

	// Apply timezone
	tz := strings.TrimSpace(req.Timezone)
	if tz == "" {
		tz = "UTC"
	}
	loc, err := time.LoadLocation(tz)
	if err != nil {
		return &pb.DateFormatResponse{Error: fmt.Sprintf("unknown timezone %q", tz)}, nil
	}
	t = t.In(loc)

	isoWeek, isoWeekYear := t.ISOWeek()
	epoch := time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)
	daysSinceEpoch := int(t.UTC().Sub(epoch).Hours() / 24)

	formats := []*pb.DateFormatEntry{
		{Label: "ISO 8601 date",          Result: t.Format("2006-01-02")},
		{Label: "ISO 8601 datetime",      Result: t.UTC().Format(time.RFC3339)},
		{Label: "RFC 2822",               Result: t.Format("Mon, 02 Jan 2006 15:04:05 -0700")},
		{Label: "RFC 850 (obsolete)",     Result: t.Format("Monday, 02-Jan-06 15:04:05 MST")},
		{Label: "Unix timestamp (s)",     Result: strconv.FormatInt(t.Unix(), 10)},
		{Label: "Unix timestamp (ms)",    Result: strconv.FormatInt(t.UnixMilli(), 10)},
		{Label: "Unix timestamp (µs)",    Result: strconv.FormatInt(t.UnixMicro(), 10)},
		{Label: "Unix timestamp (ns)",    Result: strconv.FormatInt(t.UnixNano(), 10)},
		{Label: "US long",                Result: t.Format("January 2, 2006")},
		{Label: "US short",               Result: t.Format("Jan 2, 2006")},
		{Label: "US numeric",             Result: t.Format("01/02/2006")},
		{Label: "EU long",                Result: t.Format("2 January 2006")},
		{Label: "EU short",               Result: t.Format("2 Jan 2006")},
		{Label: "EU numeric",             Result: t.Format("02.01.2006")},
		{Label: "Full weekday",           Result: t.Format("Monday, 2 January 2006")},
		{Label: "With time (12h)",        Result: t.Format("January 2, 2006 3:04:05 PM")},
		{Label: "With time (24h)",        Result: t.Format("2006-01-02 15:04:05")},
		{Label: "SQL datetime",           Result: t.UTC().Format("2006-01-02 15:04:05")},
		{Label: "ISO 8601 week date",     Result: fmt.Sprintf("%04d-W%02d-%d", isoWeekYear, isoWeek, int(t.Weekday()))},
		{Label: "Ordinal date",           Result: fmt.Sprintf("%04d-%03d", t.Year(), t.YearDay())},
		{Label: "Days since Unix epoch",  Result: strconv.Itoa(daysSinceEpoch)},
		{Label: "Timezone",               Result: t.Format("MST (UTC-07:00)")},
	}

	return &pb.DateFormatResponse{Formats: formats}, nil
}

// ─── Date info handler ────────────────────────────────────────────────────────

func zodiacSign(month time.Month, day int) string {
	switch {
	case (month == time.March && day >= 21) || (month == time.April && day <= 19):
		return "Aries ♈"
	case (month == time.April && day >= 20) || (month == time.May && day <= 20):
		return "Taurus ♉"
	case (month == time.May && day >= 21) || (month == time.June && day <= 20):
		return "Gemini ♊"
	case (month == time.June && day >= 21) || (month == time.July && day <= 22):
		return "Cancer ♋"
	case (month == time.July && day >= 23) || (month == time.August && day <= 22):
		return "Leo ♌"
	case (month == time.August && day >= 23) || (month == time.September && day <= 22):
		return "Virgo ♍"
	case (month == time.September && day >= 23) || (month == time.October && day <= 22):
		return "Libra ♎"
	case (month == time.October && day >= 23) || (month == time.November && day <= 21):
		return "Scorpio ♏"
	case (month == time.November && day >= 22) || (month == time.December && day <= 21):
		return "Sagittarius ♐"
	case (month == time.December && day >= 22) || (month == time.January && day <= 19):
		return "Capricorn ♑"
	case (month == time.January && day >= 20) || (month == time.February && day <= 18):
		return "Aquarius ♒"
	default:
		return "Pisces ♓"
	}
}

func season(month time.Month, day int) string {
	// Northern hemisphere seasons
	switch {
	case (month == time.March && day >= 20) || month == time.April || month == time.May ||
		(month == time.June && day < 21):
		return "Spring 🌸"
	case (month == time.June && day >= 21) || month == time.July || month == time.August ||
		(month == time.September && day < 22):
		return "Summer ☀️"
	case (month == time.September && day >= 22) || month == time.October || month == time.November ||
		(month == time.December && day < 21):
		return "Autumn 🍂"
	default:
		return "Winter ❄️"
	}
}

func daysInMonthFn(year int, month time.Month) int {
	return time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC).Day()
}

func (s *Server) DateInfo(_ context.Context, req *pb.DateInfoRequest) (*pb.DateInfoResponse, error) {
	t, err := parseDate(req.Date)
	if err != nil {
		return &pb.DateInfoResponse{Error: err.Error()}, nil
	}

	year, month, day := t.Date()
	weekday := t.Weekday()
	isWeekend := weekday == time.Saturday || weekday == time.Sunday
	doy := t.YearDay()
	isoYear, isoWeek := t.ISOWeek()

	daysInYear := 365
	if isLeapYear(year) {
		daysInYear = 366
	}
	daysLeftYear := daysInYear - doy
	daysInMonth := daysInMonthFn(year, month)
	daysLeftMonth := daysInMonth - day

	var quarter string
	switch {
	case month <= 3:
		quarter = "Q1"
	case month <= 6:
		quarter = "Q2"
	case month <= 9:
		quarter = "Q3"
	default:
		quarter = "Q4"
	}

	epoch := time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)
	daysSinceEpoch := int64(time.Date(year, month, day, 0, 0, 0, 0, time.UTC).Sub(epoch).Hours() / 24)

	return &pb.DateInfoResponse{
		Weekday:       weekday.String(),
		IsWeekend:     isWeekend,
		DayOfYear:     int32(doy),            // #nosec G115
		DaysInYear:    int32(daysInYear),      // #nosec G115
		DaysLeftYear:  int32(daysLeftYear),    // #nosec G115
		IsoWeek:       int32(isoWeek),         // #nosec G115
		IsoYear:       int32(isoYear),         // #nosec G115
		Quarter:       quarter,
		DaysInMonth:   int32(daysInMonth),     // #nosec G115
		DaysLeftMonth: int32(daysLeftMonth),   // #nosec G115
		UnixSec:       t.Unix(),
		Zodiac:        zodiacSign(month, day),
		Season:        season(month, day),
		DaysSinceEpoch: daysSinceEpoch,
	}, nil
}

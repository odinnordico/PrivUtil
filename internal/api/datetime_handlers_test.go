package api

import (
	"context"
	"testing"

	pb "github.com/odinnordico/privutil/proto"
)

var dtSrv = &Server{}

// ─── parseDate ────────────────────────────────────────────────────────────────

func TestParseDate_ISO(t *testing.T) {
	cases := []string{
		"2024-01-15",
		"2024-01-15T14:30:00Z",
		"2024-01-15 14:30:00",
		"January 15, 2024",
		"Jan 15, 2024",
		"15 January 2024",
		"15 Jan 2024",
		"20240115",
	}
	for _, s := range cases {
		t.Run(s, func(t *testing.T) {
			d, err := parseDate(s)
			if err != nil {
				t.Fatalf("parseDate(%q): %v", s, err)
			}
			if d.Year() != 2024 || d.Month() != 1 || d.Day() != 15 {
				t.Errorf("parseDate(%q) = %v", s, d)
			}
		})
	}
}

func TestParseDate_Relative(t *testing.T) {
	for _, s := range []string{"today", "now", "tomorrow", "yesterday"} {
		d, err := parseDate(s)
		if err != nil {
			t.Fatalf("parseDate(%q): %v", s, err)
		}
		if d.IsZero() {
			t.Errorf("parseDate(%q) returned zero time", s)
		}
	}
}

func TestParseDate_UnixTimestamp(t *testing.T) {
	d, err := parseDate("1705276800") // 2024-01-15 00:00:00 UTC
	if err != nil {
		t.Fatal(err)
	}
	if d.Year() != 2024 || d.Month() != 1 || d.Day() != 15 {
		t.Errorf("unix parse: got %v", d)
	}
}

func TestParseDate_EU_Day(t *testing.T) {
	// Day > 12, unambiguously DD/MM/YYYY
	d, err := parseDate("25/12/2024")
	if err != nil {
		t.Fatal(err)
	}
	if d.Month() != 12 || d.Day() != 25 {
		t.Errorf("EU date: got %v", d)
	}
}

func TestParseDate_Empty(t *testing.T) {
	_, err := parseDate("")
	if err == nil {
		t.Error("expected error for empty input")
	}
}

// ─── isLeapYear ───────────────────────────────────────────────────────────────

func TestIsLeapYear(t *testing.T) {
	cases := []struct{ year int; want bool }{
		{2024, true},  // div by 4
		{2023, false}, // not div by 4
		{1900, false}, // div by 100, not 400
		{2000, true},  // div by 400
		{1600, true},  // div by 400
	}
	for _, tc := range cases {
		if got := isLeapYear(tc.year); got != tc.want {
			t.Errorf("isLeapYear(%d) = %v, want %v", tc.year, got, tc.want)
		}
	}
}

// ─── calendarDiff ─────────────────────────────────────────────────────────────

func TestCalendarDiff_ExactYears(t *testing.T) {
	from, _ := parseDate("2020-01-01")
	to, _ := parseDate("2024-01-01")
	y, m, d := calendarDiff(from, to)
	if y != 4 || m != 0 || d != 0 {
		t.Errorf("4 exact years: got %d y %d m %d d", y, m, d)
	}
}

func TestCalendarDiff_Mixed(t *testing.T) {
	from, _ := parseDate("2024-01-15")
	to, _ := parseDate("2024-03-20")
	y, m, d := calendarDiff(from, to)
	if y != 0 || m != 2 || d != 5 {
		t.Errorf("2mo 5d: got %d y %d m %d d", y, m, d)
	}
}

func TestCalendarDiff_EdgeCase_Jan31_to_Mar1(t *testing.T) {
	// Go's AddDate overflows: Jan 31 + 1 month = Mar 2 (> Mar 1), so 0 complete months.
	// Result: 0 years, 0 months, 30 days.
	from, _ := parseDate("2024-01-31")
	to, _ := parseDate("2024-03-01")
	y, m, d := calendarDiff(from, to)
	if y != 0 || m != 0 || d != 30 {
		t.Errorf("Jan31→Mar1 edge: got %d y %d m %d d", y, m, d)
	}
}

func TestCalendarDiff_SameDate(t *testing.T) {
	from, _ := parseDate("2024-06-15")
	to, _ := parseDate("2024-06-15")
	y, m, d := calendarDiff(from, to)
	if y != 0 || m != 0 || d != 0 {
		t.Errorf("same date: got %d y %d m %d d", y, m, d)
	}
}

// ─── DateDiff handler ─────────────────────────────────────────────────────────

func TestDateDiff_Basic(t *testing.T) {
	resp, err := dtSrv.DateDiff(context.Background(), &pb.DateDiffRequest{
		FromDate: "2020-01-01", ToDate: "2024-06-15",
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.Error != "" {
		t.Fatal(resp.Error)
	}
	if resp.Years != 4 {
		t.Errorf("years: got %d, want 4", resp.Years)
	}
	if resp.TotalDays <= 0 {
		t.Error("total days should be positive")
	}
	if resp.Human == "" {
		t.Error("human string should not be empty")
	}
}

func TestDateDiff_Negative(t *testing.T) {
	// from > to → negative flag
	resp, _ := dtSrv.DateDiff(context.Background(), &pb.DateDiffRequest{
		FromDate: "2024-12-31", ToDate: "2024-01-01",
	})
	if !resp.Negative {
		t.Error("expected negative flag when from > to")
	}
	if resp.TotalDays <= 0 {
		t.Error("total days should still be positive")
	}
}

func TestDateDiff_SameDates(t *testing.T) {
	resp, _ := dtSrv.DateDiff(context.Background(), &pb.DateDiffRequest{
		FromDate: "2024-06-01", ToDate: "2024-06-01",
	})
	if resp.TotalDays != 0 || resp.Years != 0 {
		t.Errorf("same date should give zero diff, got %+v", resp)
	}
}

func TestDateDiff_InvalidDate(t *testing.T) {
	resp, _ := dtSrv.DateDiff(context.Background(), &pb.DateDiffRequest{
		FromDate: "not-a-date", ToDate: "2024-01-01",
	})
	if resp.Error == "" {
		t.Error("expected error for invalid date")
	}
}

// ─── LeapYear handler ─────────────────────────────────────────────────────────

func TestLeapYear_Single(t *testing.T) {
	resp, err := dtSrv.LeapYear(context.Background(), &pb.LeapYearRequest{Input: "2024"})
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.Results) != 1 || !resp.Results[0].IsLeap {
		t.Error("2024 should be a leap year")
	}
}

func TestLeapYear_NotLeap(t *testing.T) {
	resp, _ := dtSrv.LeapYear(context.Background(), &pb.LeapYearRequest{Input: "2023"})
	if resp.Results[0].IsLeap {
		t.Error("2023 should not be a leap year")
	}
}

func TestLeapYear_CommaList(t *testing.T) {
	resp, _ := dtSrv.LeapYear(context.Background(), &pb.LeapYearRequest{Input: "2020, 2021, 2022, 2023, 2024"})
	if len(resp.Results) != 5 {
		t.Fatalf("expected 5 results, got %d", len(resp.Results))
	}
	if resp.LeapCount != 2 { // 2020 and 2024
		t.Errorf("leap count: got %d, want 2", resp.LeapCount)
	}
}

func TestLeapYear_Range(t *testing.T) {
	resp, _ := dtSrv.LeapYear(context.Background(), &pb.LeapYearRequest{Input: "2000-2004"})
	if len(resp.Results) != 5 {
		t.Fatalf("expected 5 results, got %d", len(resp.Results))
	}
	// 2000 (leap), 2001, 2002, 2003, 2004 (leap)
	if resp.LeapCount != 2 {
		t.Errorf("leap count: got %d, want 2", resp.LeapCount)
	}
}

func TestLeapYear_RangeTooLarge(t *testing.T) {
	resp, _ := dtSrv.LeapYear(context.Background(), &pb.LeapYearRequest{Input: "1600-2100"})
	if resp.Error == "" {
		t.Error("expected error for range > 400")
	}
}

func TestLeapYear_Empty(t *testing.T) {
	resp, _ := dtSrv.LeapYear(context.Background(), &pb.LeapYearRequest{Input: ""})
	if resp.Error == "" {
		t.Error("expected error for empty input")
	}
}

// ─── DateAdd handler ──────────────────────────────────────────────────────────

func TestDateAdd_Basic(t *testing.T) {
	resp, err := dtSrv.DateAdd(context.Background(), &pb.DateAddRequest{
		Date: "2024-01-15", Years: 1, Months: 2, Days: 5,
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.Error != "" {
		t.Fatal(resp.Error)
	}
	// 2024-01-15 + 1yr + 2mo + 5d = 2025-03-20
	if resp.Iso != "2025-03-20" {
		t.Errorf("expected 2025-03-20, got %s", resp.Iso)
	}
}

func TestDateAdd_Subtract(t *testing.T) {
	resp, _ := dtSrv.DateAdd(context.Background(), &pb.DateAddRequest{
		Date: "2024-03-01", Days: -1,
	})
	// Leap year: 2024-02-29
	if resp.Iso != "2024-02-29" {
		t.Errorf("expected 2024-02-29, got %s", resp.Iso)
	}
}

func TestDateAdd_Weekday(t *testing.T) {
	resp, _ := dtSrv.DateAdd(context.Background(), &pb.DateAddRequest{
		Date: "2024-01-15", // Monday
	})
	if resp.Weekday != "Monday" {
		t.Errorf("weekday: got %s, want Monday", resp.Weekday)
	}
}

func TestDateAdd_WeekInput(t *testing.T) {
	resp, _ := dtSrv.DateAdd(context.Background(), &pb.DateAddRequest{
		Date: "2024-01-01", Weeks: 2,
	})
	// 2024-01-01 + 14 days = 2024-01-15
	if resp.Iso != "2024-01-15" {
		t.Errorf("expected 2024-01-15, got %s", resp.Iso)
	}
}

func TestDateAdd_InvalidDate(t *testing.T) {
	resp, _ := dtSrv.DateAdd(context.Background(), &pb.DateAddRequest{Date: "not-a-date"})
	if resp.Error == "" {
		t.Error("expected error for invalid date")
	}
}

// ─── DateFormat handler ───────────────────────────────────────────────────────

func TestDateFormat_Basic(t *testing.T) {
	resp, err := dtSrv.DateFormat(context.Background(), &pb.DateFormatRequest{
		DateStr: "2024-01-15", Timezone: "UTC",
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.Error != "" {
		t.Fatal(resp.Error)
	}
	if len(resp.Formats) == 0 {
		t.Error("expected format results")
	}
	// Find ISO 8601 date
	for _, f := range resp.Formats {
		if f.Label == "ISO 8601 date" && f.Result != "2024-01-15" {
			t.Errorf("ISO 8601 date: got %s", f.Result)
		}
	}
}

func TestDateFormat_UnixTimestamp(t *testing.T) {
	resp, _ := dtSrv.DateFormat(context.Background(), &pb.DateFormatRequest{
		DateStr: "2024-01-15T00:00:00Z", Timezone: "UTC",
	})
	for _, f := range resp.Formats {
		if f.Label == "Unix timestamp (s)" {
			if f.Result == "" {
				t.Error("unix timestamp should not be empty")
			}
		}
	}
}

func TestDateFormat_InvalidTimezone(t *testing.T) {
	resp, _ := dtSrv.DateFormat(context.Background(), &pb.DateFormatRequest{
		DateStr: "2024-01-15", Timezone: "Not/Real",
	})
	if resp.Error == "" {
		t.Error("expected error for invalid timezone")
	}
}

func TestDateFormat_DefaultUTC(t *testing.T) {
	resp, _ := dtSrv.DateFormat(context.Background(), &pb.DateFormatRequest{
		DateStr: "2024-01-15",
	})
	if resp.Error != "" {
		t.Fatalf("empty timezone should default to UTC: %s", resp.Error)
	}
}

// ─── DateInfo handler ─────────────────────────────────────────────────────────

func TestDateInfo_Basic(t *testing.T) {
	resp, err := dtSrv.DateInfo(context.Background(), &pb.DateInfoRequest{Date: "2024-01-15"})
	if err != nil {
		t.Fatal(err)
	}
	if resp.Error != "" {
		t.Fatal(resp.Error)
	}
	if resp.Weekday != "Monday" {
		t.Errorf("weekday: got %s, want Monday", resp.Weekday)
	}
	if resp.IsWeekend {
		t.Error("2024-01-15 (Monday) should not be weekend")
	}
	if resp.DayOfYear != 15 {
		t.Errorf("day of year: got %d, want 15", resp.DayOfYear)
	}
	if resp.Quarter != "Q1" {
		t.Errorf("quarter: got %s, want Q1", resp.Quarter)
	}
	if resp.IsoWeek != 3 {
		t.Errorf("ISO week: got %d, want 3", resp.IsoWeek)
	}
}

func TestDateInfo_Weekend(t *testing.T) {
	resp, _ := dtSrv.DateInfo(context.Background(), &pb.DateInfoRequest{Date: "2024-01-20"}) // Saturday
	if !resp.IsWeekend {
		t.Error("2024-01-20 (Saturday) should be weekend")
	}
}

func TestDateInfo_LeapYear(t *testing.T) {
	resp, _ := dtSrv.DateInfo(context.Background(), &pb.DateInfoRequest{Date: "2024-12-31"})
	if resp.DaysInYear != 366 {
		t.Errorf("2024 is leap (366 days): got %d", resp.DaysInYear)
	}
	if resp.DaysLeftYear != 0 {
		t.Errorf("Dec 31 has 0 days left: got %d", resp.DaysLeftYear)
	}
}

func TestDateInfo_Quarters(t *testing.T) {
	cases := []struct{ date, want string }{
		{"2024-01-01", "Q1"},
		{"2024-04-01", "Q2"},
		{"2024-07-01", "Q3"},
		{"2024-10-01", "Q4"},
	}
	for _, tc := range cases {
		resp, _ := dtSrv.DateInfo(context.Background(), &pb.DateInfoRequest{Date: tc.date})
		if resp.Quarter != tc.want {
			t.Errorf("%s: quarter got %s, want %s", tc.date, resp.Quarter, tc.want)
		}
	}
}

func TestDateInfo_Zodiac(t *testing.T) {
	cases := []struct{ date, want string }{
		{"2024-01-15", "Capricorn ♑"},
		{"2024-03-25", "Aries ♈"},
		{"2024-06-21", "Cancer ♋"},
		{"2024-12-25", "Capricorn ♑"},
	}
	for _, tc := range cases {
		resp, _ := dtSrv.DateInfo(context.Background(), &pb.DateInfoRequest{Date: tc.date})
		if resp.Zodiac != tc.want {
			t.Errorf("%s: zodiac got %q, want %q", tc.date, resp.Zodiac, tc.want)
		}
	}
}

func TestDateInfo_InvalidDate(t *testing.T) {
	resp, _ := dtSrv.DateInfo(context.Background(), &pb.DateInfoRequest{Date: "garbage"})
	if resp.Error == "" {
		t.Error("expected error for invalid date")
	}
}

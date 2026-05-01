//go:build manual

package api

import (
	"context"
	"strings"
	"testing"

	pb "github.com/odinnordico/privutil/proto"
)

var wdSrv = &Server{}
var wdCtx = context.Background()

// ─── URL Parser ───────────────────────────────────────────────────────────────

func TestUrlParse_Full(t *testing.T) {
	res, err := wdSrv.UrlParse(wdCtx, &pb.UrlParseRequest{Url: "https://user:pass@example.com:8080/path/to/page?foo=1&bar=2#section"})
	if err != nil {
		t.Fatal(err)
	}
	if !res.IsValid {
		t.Fatal("expected valid")
	}
	if res.Scheme != "https" {
		t.Errorf("scheme got %q", res.Scheme)
	}
	if res.Username != "user" {
		t.Errorf("username got %q", res.Username)
	}
	if res.Password != "pass" {
		t.Errorf("password got %q", res.Password)
	}
	if res.Hostname != "example.com" {
		t.Errorf("hostname got %q", res.Hostname)
	}
	if res.Port != "8080" {
		t.Errorf("port got %q", res.Port)
	}
	if res.Path != "/path/to/page" {
		t.Errorf("path got %q", res.Path)
	}
	if len(res.QueryParams) != 2 {
		t.Errorf("query params len got %d", len(res.QueryParams))
	}
	if res.Fragment != "section" {
		t.Errorf("fragment got %q", res.Fragment)
	}
}

func TestUrlParse_NoScheme(t *testing.T) {
	res, err := wdSrv.UrlParse(wdCtx, &pb.UrlParseRequest{Url: "example.com/page"})
	if err != nil {
		t.Fatal(err)
	}
	if !res.IsValid {
		t.Fatal("expected valid")
	}
	if res.Scheme != "" {
		t.Errorf("scheme should be empty for no-scheme input, got %q", res.Scheme)
	}
	if res.Hostname != "example.com" {
		t.Errorf("hostname got %q", res.Hostname)
	}
}

func TestUrlParse_QueryParams(t *testing.T) {
	res, err := wdSrv.UrlParse(wdCtx, &pb.UrlParseRequest{Url: "https://api.example.com/v1/search?q=hello+world&lang=en&page=2"})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.QueryParams) != 3 {
		t.Errorf("expected 3 params, got %d", len(res.QueryParams))
	}
}

func TestUrlParse_Empty(t *testing.T) {
	_, err := wdSrv.UrlParse(wdCtx, &pb.UrlParseRequest{Url: ""})
	if err == nil {
		t.Fatal("expected error for empty url")
	}
}

// ─── User-Agent Parser ────────────────────────────────────────────────────────

func TestUserAgentParse_Chrome(t *testing.T) {
	ua := "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"
	res, err := wdSrv.UserAgentParse(wdCtx, &pb.UserAgentParseRequest{UserAgent: ua})
	if err != nil {
		t.Fatal(err)
	}
	if res.BrowserName != "Chrome" {
		t.Errorf("browser got %q", res.BrowserName)
	}
	if !strings.HasPrefix(res.BrowserVersion, "120") {
		t.Errorf("version got %q", res.BrowserVersion)
	}
	if res.OsName != "Windows" {
		t.Errorf("os got %q", res.OsName)
	}
	if res.DeviceType != "desktop" {
		t.Errorf("device got %q", res.DeviceType)
	}
	if res.IsBot {
		t.Error("should not be bot")
	}
}

func TestUserAgentParse_Firefox(t *testing.T) {
	ua := "Mozilla/5.0 (X11; Linux x86_64; rv:120.0) Gecko/20100101 Firefox/120.0"
	res, err := wdSrv.UserAgentParse(wdCtx, &pb.UserAgentParseRequest{UserAgent: ua})
	if err != nil {
		t.Fatal(err)
	}
	if res.BrowserName != "Firefox" {
		t.Errorf("browser got %q", res.BrowserName)
	}
	if res.Engine != "Gecko" {
		t.Errorf("engine got %q", res.Engine)
	}
	if res.OsName != "Linux" {
		t.Errorf("os got %q", res.OsName)
	}
}

func TestUserAgentParse_Mobile(t *testing.T) {
	ua := "Mozilla/5.0 (Linux; Android 13; Pixel 7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/112.0.0.0 Mobile Safari/537.36"
	res, err := wdSrv.UserAgentParse(wdCtx, &pb.UserAgentParseRequest{UserAgent: ua})
	if err != nil {
		t.Fatal(err)
	}
	if res.DeviceType != "mobile" {
		t.Errorf("device got %q", res.DeviceType)
	}
	if !res.IsMobile {
		t.Error("should be mobile")
	}
	if res.OsName != "Android" {
		t.Errorf("os got %q", res.OsName)
	}
}

func TestUserAgentParse_Bot(t *testing.T) {
	ua := "Mozilla/5.0 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)"
	res, err := wdSrv.UserAgentParse(wdCtx, &pb.UserAgentParseRequest{UserAgent: ua})
	if err != nil {
		t.Fatal(err)
	}
	if !res.IsBot {
		t.Error("should be bot")
	}
	if res.DeviceType != "bot" {
		t.Errorf("device got %q", res.DeviceType)
	}
}

func TestUserAgentParse_Edge(t *testing.T) {
	ua := "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36 Edg/120.0.0.0"
	res, err := wdSrv.UserAgentParse(wdCtx, &pb.UserAgentParseRequest{UserAgent: ua})
	if err != nil {
		t.Fatal(err)
	}
	if res.BrowserName != "Edge" {
		t.Errorf("browser got %q", res.BrowserName)
	}
}

func TestUserAgentParse_Empty(t *testing.T) {
	_, err := wdSrv.UserAgentParse(wdCtx, &pb.UserAgentParseRequest{UserAgent: ""})
	if err == nil {
		t.Fatal("expected error for empty UA")
	}
}

// ─── HTTP Status Codes ────────────────────────────────────────────────────────

func TestHttpStatusSearch_All(t *testing.T) {
	res, err := wdSrv.HttpStatusSearch(wdCtx, &pb.HttpStatusSearchRequest{})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Entries) < 50 {
		t.Errorf("expected at least 50 entries, got %d", len(res.Entries))
	}
}

func TestHttpStatusSearch_ByCode(t *testing.T) {
	res, err := wdSrv.HttpStatusSearch(wdCtx, &pb.HttpStatusSearchRequest{Query: "404"})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Entries) != 1 {
		t.Fatalf("expected 1 entry for 404, got %d", len(res.Entries))
	}
	if res.Entries[0].Code != 404 {
		t.Errorf("code got %d", res.Entries[0].Code)
	}
	if res.Entries[0].Name != "Not Found" {
		t.Errorf("name got %q", res.Entries[0].Name)
	}
}

func TestHttpStatusSearch_ByCategory(t *testing.T) {
	res, err := wdSrv.HttpStatusSearch(wdCtx, &pb.HttpStatusSearchRequest{Category: "2xx"})
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range res.Entries {
		if e.Code < 200 || e.Code >= 300 {
			t.Errorf("unexpected code %d in 2xx category", e.Code)
		}
	}
	if len(res.Entries) == 0 {
		t.Error("expected 2xx entries")
	}
}

func TestHttpStatusSearch_ByKeyword(t *testing.T) {
	res, err := wdSrv.HttpStatusSearch(wdCtx, &pb.HttpStatusSearchRequest{Query: "redirect"})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Entries) == 0 {
		t.Error("expected redirect-related entries")
	}
}

func TestHttpStatusSearch_Teapot(t *testing.T) {
	res, err := wdSrv.HttpStatusSearch(wdCtx, &pb.HttpStatusSearchRequest{Query: "418"})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Entries) != 1 || res.Entries[0].Code != 418 {
		t.Error("expected 418 teapot")
	}
}

// ─── MIME Type Lookup ─────────────────────────────────────────────────────────

func TestMimeLookup_All(t *testing.T) {
	res, err := wdSrv.MimeLookup(wdCtx, &pb.MimeLookupRequest{Query: ""})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Entries) < 50 {
		t.Errorf("expected at least 50 entries, got %d", len(res.Entries))
	}
}

func TestMimeLookup_ByExtension(t *testing.T) {
	res, err := wdSrv.MimeLookup(wdCtx, &pb.MimeLookupRequest{Query: "pdf"})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Entries) == 0 {
		t.Fatal("expected pdf entry")
	}
	if res.Entries[0].MimeType != "application/pdf" {
		t.Errorf("mime got %q", res.Entries[0].MimeType)
	}
}

func TestMimeLookup_ByExtensionWithDot(t *testing.T) {
	res, err := wdSrv.MimeLookup(wdCtx, &pb.MimeLookupRequest{Query: ".json"})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Entries) == 0 {
		t.Fatal("expected json entry")
	}
}

func TestMimeLookup_ByMimeType(t *testing.T) {
	res, err := wdSrv.MimeLookup(wdCtx, &pb.MimeLookupRequest{Query: "image/png"})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Entries) == 0 {
		t.Fatal("expected png entry")
	}
}

func TestMimeLookup_ByCategory(t *testing.T) {
	res, err := wdSrv.MimeLookup(wdCtx, &pb.MimeLookupRequest{Query: "video"})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Entries) < 3 {
		t.Errorf("expected at least 3 video entries, got %d", len(res.Entries))
	}
}

// ─── Docker run → Compose ─────────────────────────────────────────────────────

func TestDockerRunToCompose_Basic(t *testing.T) {
	res, err := wdSrv.DockerRunToCompose(wdCtx, &pb.DockerRunToComposeRequest{
		Command: "docker run -d --name myapp -p 8080:80 nginx:latest",
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.Error != "" {
		t.Fatal(res.Error)
	}
	if !strings.Contains(res.ComposeYaml, "nginx:latest") {
		t.Error("image not in compose")
	}
	if !strings.Contains(res.ComposeYaml, "8080:80") {
		t.Error("port not in compose")
	}
	if !strings.Contains(res.ComposeYaml, "container_name: myapp") {
		t.Error("container name not in compose")
	}
}

func TestDockerRunToCompose_Full(t *testing.T) {
	res, err := wdSrv.DockerRunToCompose(wdCtx, &pb.DockerRunToComposeRequest{
		Command: `docker run -d --name web --restart always -p 443:443 -p 80:80 -v /data:/app/data -e DATABASE_URL=postgres://localhost/db -e SECRET_KEY=abc123 --network mynet --memory 512m postgres:15`,
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.Error != "" {
		t.Fatal(res.Error)
	}
	yaml := res.ComposeYaml
	if !strings.Contains(yaml, "restart: always") {
		t.Error("restart not in compose")
	}
	if !strings.Contains(yaml, "mem_limit: 512m") {
		t.Error("mem_limit not in compose")
	}
	if !strings.Contains(yaml, "networks:") {
		t.Error("network not in compose")
	}
	if !strings.Contains(yaml, "DATABASE_URL=postgres://localhost/db") {
		t.Error("env var not in compose")
	}
}

func TestDockerRunToCompose_Privileged(t *testing.T) {
	res, err := wdSrv.DockerRunToCompose(wdCtx, &pb.DockerRunToComposeRequest{
		Command: "docker run --privileged --read-only alpine sh",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(res.ComposeYaml, "privileged: true") {
		t.Error("privileged not in compose")
	}
	if !strings.Contains(res.ComposeYaml, "read_only: true") {
		t.Error("read_only not in compose")
	}
}

func TestDockerRunToCompose_RmWarning(t *testing.T) {
	res, err := wdSrv.DockerRunToCompose(wdCtx, &pb.DockerRunToComposeRequest{
		Command: "docker run --rm ubuntu bash",
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Warnings) == 0 {
		t.Error("expected warning for --rm flag")
	}
}

func TestDockerRunToCompose_NoImage(t *testing.T) {
	res, err := wdSrv.DockerRunToCompose(wdCtx, &pb.DockerRunToComposeRequest{
		Command: "docker run -e FOO=bar",
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.Error == "" {
		t.Error("expected error for missing image")
	}
}

func TestDockerRunToCompose_Empty(t *testing.T) {
	_, err := wdSrv.DockerRunToCompose(wdCtx, &pb.DockerRunToComposeRequest{Command: ""})
	if err == nil {
		t.Fatal("expected error for empty command")
	}
}

func TestDockerRunToCompose_QuotedValues(t *testing.T) {
	res, err := wdSrv.DockerRunToCompose(wdCtx, &pb.DockerRunToComposeRequest{
		Command: `docker run -e "MESSAGE=hello world" -e EMPTY="" myimage`,
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.Error != "" {
		t.Fatal(res.Error)
	}
	if !strings.Contains(res.ComposeYaml, "MESSAGE=hello world") {
		t.Error("quoted env var not parsed correctly")
	}
}

// ─── Git Cheat Sheet ──────────────────────────────────────────────────────────

func TestGitCheatSheet_All(t *testing.T) {
	res, err := wdSrv.GitCheatSheet(wdCtx, &pb.GitCheatSheetRequest{})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Categories) < 8 {
		t.Errorf("expected at least 8 categories, got %d", len(res.Categories))
	}
	total := 0
	for _, c := range res.Categories {
		total += len(c.Commands)
	}
	if total < 50 {
		t.Errorf("expected at least 50 commands, got %d", total)
	}
}

func TestGitCheatSheet_SearchByCommand(t *testing.T) {
	res, err := wdSrv.GitCheatSheet(wdCtx, &pb.GitCheatSheetRequest{Query: "stash"})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Categories) == 0 {
		t.Fatal("expected results for 'stash'")
	}
	for _, cat := range res.Categories {
		for _, cmd := range cat.Commands {
			found := strings.Contains(strings.ToLower(cmd.Command), "stash") ||
				strings.Contains(strings.ToLower(cmd.Description), "stash")
			if !found {
				t.Errorf("unrelated command in results: %q", cmd.Command)
			}
		}
	}
}

func TestGitCheatSheet_SearchByDescription(t *testing.T) {
	res, err := wdSrv.GitCheatSheet(wdCtx, &pb.GitCheatSheetRequest{Query: "revert"})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Categories) == 0 {
		t.Fatal("expected results for 'revert'")
	}
}

func TestGitCheatSheet_CategoryFilter(t *testing.T) {
	res, err := wdSrv.GitCheatSheet(wdCtx, &pb.GitCheatSheetRequest{Category: "stash"})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Categories) != 1 {
		t.Errorf("expected 1 category, got %d", len(res.Categories))
	}
	if !strings.Contains(strings.ToLower(res.Categories[0].Name), "stash") {
		t.Errorf("wrong category: %q", res.Categories[0].Name)
	}
}

func TestGitCheatSheet_NoMatch(t *testing.T) {
	res, err := wdSrv.GitCheatSheet(wdCtx, &pb.GitCheatSheetRequest{Query: "zzznomatch"})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Categories) != 0 {
		t.Error("expected no results for non-existent query")
	}
}

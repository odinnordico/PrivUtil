package api

// This file adapts the existing gRPC-style *Server handlers to the connect-go
// PrivUtilServiceHandler interface. Each method unwraps the connect request,
// invokes the matching *Server method, and wraps the response. Generated from
// the rpc definitions in proto/privutil.proto; regenerate with `make proto`.

import (
	"context"
	"errors"

	connect "connectrpc.com/connect"
	pb "github.com/odinnordico/privutil/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ConnectServer adapts *Server to the connect PrivUtilServiceHandler interface.
type ConnectServer struct {
	s *Server
}

// NewConnectServer wraps an existing *Server for use with connect-go.
func NewConnectServer(s *Server) *ConnectServer {
	return &ConnectServer{s: s}
}

// toConnectError preserves gRPC status codes (e.g. InvalidArgument) emitted by
// the handlers by mapping them to the numerically-identical connect codes. Plain
// errors are returned unchanged, which connect reports as CodeUnknown — matching
// the previous gRPC behavior.
func toConnectError(err error) error {
	if err == nil {
		return nil
	}
	if st, ok := status.FromError(err); ok && st.Code() != codes.OK {
		return connect.NewError(connect.Code(st.Code()), errors.New(st.Message()))
	}
	return err
}

func (a *ConnectServer) Diff(ctx context.Context, r *connect.Request[pb.DiffRequest]) (*connect.Response[pb.DiffResponse], error) {
	resp, err := a.s.Diff(ctx, r.Msg)
	if err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(resp), nil
}

func (a *ConnectServer) Base64Encode(ctx context.Context, r *connect.Request[pb.Base64Request]) (*connect.Response[pb.Base64Response], error) {
	resp, err := a.s.Base64Encode(ctx, r.Msg)
	if err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(resp), nil
}

func (a *ConnectServer) Base64Decode(ctx context.Context, r *connect.Request[pb.Base64Request]) (*connect.Response[pb.Base64Response], error) {
	resp, err := a.s.Base64Decode(ctx, r.Msg)
	if err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(resp), nil
}

func (a *ConnectServer) JsonFormat(ctx context.Context, r *connect.Request[pb.JsonFormatRequest]) (*connect.Response[pb.JsonFormatResponse], error) {
	resp, err := a.s.JsonFormat(ctx, r.Msg)
	if err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(resp), nil
}

func (a *ConnectServer) Convert(ctx context.Context, r *connect.Request[pb.ConvertRequest]) (*connect.Response[pb.ConvertResponse], error) {
	resp, err := a.s.Convert(ctx, r.Msg)
	if err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(resp), nil
}

func (a *ConnectServer) ValidateData(ctx context.Context, r *connect.Request[pb.ValidateRequest]) (*connect.Response[pb.ValidateResponse], error) {
	resp, err := a.s.ValidateData(ctx, r.Msg)
	if err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(resp), nil
}

func (a *ConnectServer) GenerateUuid(ctx context.Context, r *connect.Request[pb.UuidRequest]) (*connect.Response[pb.UuidResponse], error) {
	resp, err := a.s.GenerateUuid(ctx, r.Msg)
	if err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(resp), nil
}

func (a *ConnectServer) GenerateLorem(ctx context.Context, r *connect.Request[pb.LoremRequest]) (*connect.Response[pb.LoremResponse], error) {
	resp, err := a.s.GenerateLorem(ctx, r.Msg)
	if err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(resp), nil
}

func (a *ConnectServer) CalculateHash(ctx context.Context, r *connect.Request[pb.HashRequest]) (*connect.Response[pb.HashResponse], error) {
	resp, err := a.s.CalculateHash(ctx, r.Msg)
	if err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(resp), nil
}

func (a *ConnectServer) TextInspect(ctx context.Context, r *connect.Request[pb.TextInspectRequest]) (*connect.Response[pb.TextInspectResponse], error) {
	resp, err := a.s.TextInspect(ctx, r.Msg)
	if err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(resp), nil
}

func (a *ConnectServer) TextManipulate(ctx context.Context, r *connect.Request[pb.TextManipulateRequest]) (*connect.Response[pb.TextManipulateResponse], error) {
	resp, err := a.s.TextManipulate(ctx, r.Msg)
	if err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(resp), nil
}

func (a *ConnectServer) UrlEncode(ctx context.Context, r *connect.Request[pb.TextRequest]) (*connect.Response[pb.TextResponse], error) {
	resp, err := a.s.UrlEncode(ctx, r.Msg)
	if err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(resp), nil
}

func (a *ConnectServer) UrlDecode(ctx context.Context, r *connect.Request[pb.TextRequest]) (*connect.Response[pb.TextResponse], error) {
	resp, err := a.s.UrlDecode(ctx, r.Msg)
	if err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(resp), nil
}

func (a *ConnectServer) HtmlEncode(ctx context.Context, r *connect.Request[pb.TextRequest]) (*connect.Response[pb.TextResponse], error) {
	resp, err := a.s.HtmlEncode(ctx, r.Msg)
	if err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(resp), nil
}

func (a *ConnectServer) HtmlDecode(ctx context.Context, r *connect.Request[pb.TextRequest]) (*connect.Response[pb.TextResponse], error) {
	resp, err := a.s.HtmlDecode(ctx, r.Msg)
	if err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(resp), nil
}

func (a *ConnectServer) TimeConvert(ctx context.Context, r *connect.Request[pb.TimeRequest]) (*connect.Response[pb.TimeResponse], error) {
	resp, err := a.s.TimeConvert(ctx, r.Msg)
	if err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(resp), nil
}

func (a *ConnectServer) JwtDecode(ctx context.Context, r *connect.Request[pb.JwtRequest]) (*connect.Response[pb.JwtResponse], error) {
	resp, err := a.s.JwtDecode(ctx, r.Msg)
	if err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(resp), nil
}

func (a *ConnectServer) RegexTest(ctx context.Context, r *connect.Request[pb.RegexRequest]) (*connect.Response[pb.RegexResponse], error) {
	resp, err := a.s.RegexTest(ctx, r.Msg)
	if err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(resp), nil
}

func (a *ConnectServer) JsonToGo(ctx context.Context, r *connect.Request[pb.JsonToGoRequest]) (*connect.Response[pb.JsonToGoResponse], error) {
	resp, err := a.s.JsonToGo(ctx, r.Msg)
	if err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(resp), nil
}

func (a *ConnectServer) CronExplain(ctx context.Context, r *connect.Request[pb.CronRequest]) (*connect.Response[pb.CronResponse], error) {
	resp, err := a.s.CronExplain(ctx, r.Msg)
	if err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(resp), nil
}

func (a *ConnectServer) CertParse(ctx context.Context, r *connect.Request[pb.CertRequest]) (*connect.Response[pb.CertResponse], error) {
	resp, err := a.s.CertParse(ctx, r.Msg)
	if err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(resp), nil
}

func (a *ConnectServer) ColorConvert(ctx context.Context, r *connect.Request[pb.ColorRequest]) (*connect.Response[pb.ColorResponse], error) {
	resp, err := a.s.ColorConvert(ctx, r.Msg)
	if err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(resp), nil
}

func (a *ConnectServer) CaseConvert(ctx context.Context, r *connect.Request[pb.CaseRequest]) (*connect.Response[pb.CaseResponse], error) {
	resp, err := a.s.CaseConvert(ctx, r.Msg)
	if err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(resp), nil
}

func (a *ConnectServer) StringEscape(ctx context.Context, r *connect.Request[pb.EscapeRequest]) (*connect.Response[pb.EscapeResponse], error) {
	resp, err := a.s.StringEscape(ctx, r.Msg)
	if err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(resp), nil
}

func (a *ConnectServer) TextSimilarity(ctx context.Context, r *connect.Request[pb.SimilarityRequest]) (*connect.Response[pb.SimilarityResponse], error) {
	resp, err := a.s.TextSimilarity(ctx, r.Msg)
	if err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(resp), nil
}

func (a *ConnectServer) SqlFormat(ctx context.Context, r *connect.Request[pb.SqlRequest]) (*connect.Response[pb.SqlResponse], error) {
	resp, err := a.s.SqlFormat(ctx, r.Msg)
	if err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(resp), nil
}

func (a *ConnectServer) IpCalc(ctx context.Context, r *connect.Request[pb.IpRequest]) (*connect.Response[pb.IpResponse], error) {
	resp, err := a.s.IpCalc(ctx, r.Msg)
	if err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(resp), nil
}

func (a *ConnectServer) GeneratePassword(ctx context.Context, r *connect.Request[pb.PasswordRequest]) (*connect.Response[pb.PasswordResponse], error) {
	resp, err := a.s.GeneratePassword(ctx, r.Msg)
	if err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(resp), nil
}

func (a *ConnectServer) GenerateRsaKeyPair(ctx context.Context, r *connect.Request[pb.RsaKeyRequest]) (*connect.Response[pb.RsaKeyResponse], error) {
	resp, err := a.s.GenerateRsaKeyPair(ctx, r.Msg)
	if err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(resp), nil
}

func (a *ConnectServer) BaseConvert(ctx context.Context, r *connect.Request[pb.BaseConvertRequest]) (*connect.Response[pb.BaseConvertResponse], error) {
	resp, err := a.s.BaseConvert(ctx, r.Msg)
	if err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(resp), nil
}

func (a *ConnectServer) MarkdownToHtml(ctx context.Context, r *connect.Request[pb.TextRequest]) (*connect.Response[pb.TextResponse], error) {
	resp, err := a.s.MarkdownToHtml(ctx, r.Msg)
	if err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(resp), nil
}

func (a *ConnectServer) HtmlToMarkdown(ctx context.Context, r *connect.Request[pb.TextRequest]) (*connect.Response[pb.TextResponse], error) {
	resp, err := a.s.HtmlToMarkdown(ctx, r.Msg)
	if err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(resp), nil
}

func (a *ConnectServer) HmacGenerate(ctx context.Context, r *connect.Request[pb.HmacRequest]) (*connect.Response[pb.HmacResponse], error) {
	resp, err := a.s.HmacGenerate(ctx, r.Msg)
	if err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(resp), nil
}

func (a *ConnectServer) OtpGenerate(ctx context.Context, r *connect.Request[pb.OtpRequest]) (*connect.Response[pb.OtpResponse], error) {
	resp, err := a.s.OtpGenerate(ctx, r.Msg)
	if err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(resp), nil
}

func (a *ConnectServer) OtpValidate(ctx context.Context, r *connect.Request[pb.OtpValidateRequest]) (*connect.Response[pb.OtpValidateResponse], error) {
	resp, err := a.s.OtpValidate(ctx, r.Msg)
	if err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(resp), nil
}

func (a *ConnectServer) UlidGenerate(ctx context.Context, r *connect.Request[pb.UlidRequest]) (*connect.Response[pb.UlidResponse], error) {
	resp, err := a.s.UlidGenerate(ctx, r.Msg)
	if err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(resp), nil
}

func (a *ConnectServer) CaesarCipher(ctx context.Context, r *connect.Request[pb.CaesarRequest]) (*connect.Response[pb.CaesarResponse], error) {
	resp, err := a.s.CaesarCipher(ctx, r.Msg)
	if err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(resp), nil
}

func (a *ConnectServer) TextEncode(ctx context.Context, r *connect.Request[pb.TextEncodeRequest]) (*connect.Response[pb.TextEncodeResponse], error) {
	resp, err := a.s.TextEncode(ctx, r.Msg)
	if err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(resp), nil
}

func (a *ConnectServer) MorseCode(ctx context.Context, r *connect.Request[pb.MorseRequest]) (*connect.Response[pb.MorseResponse], error) {
	resp, err := a.s.MorseCode(ctx, r.Msg)
	if err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(resp), nil
}

func (a *ConnectServer) BasicAuthGenerate(ctx context.Context, r *connect.Request[pb.BasicAuthRequest]) (*connect.Response[pb.BasicAuthResponse], error) {
	resp, err := a.s.BasicAuthGenerate(ctx, r.Msg)
	if err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(resp), nil
}

func (a *ConnectServer) ChmodCalc(ctx context.Context, r *connect.Request[pb.ChmodRequest]) (*connect.Response[pb.ChmodResponse], error) {
	resp, err := a.s.ChmodCalc(ctx, r.Msg)
	if err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(resp), nil
}

func (a *ConnectServer) Ipv4Convert(ctx context.Context, r *connect.Request[pb.Ipv4ConvertRequest]) (*connect.Response[pb.Ipv4ConvertResponse], error) {
	resp, err := a.s.Ipv4Convert(ctx, r.Msg)
	if err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(resp), nil
}

func (a *ConnectServer) Ipv4RangeExpand(ctx context.Context, r *connect.Request[pb.Ipv4RangeRequest]) (*connect.Response[pb.Ipv4RangeResponse], error) {
	resp, err := a.s.Ipv4RangeExpand(ctx, r.Msg)
	if err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(resp), nil
}

func (a *ConnectServer) GeneratePort(ctx context.Context, r *connect.Request[pb.PortRequest]) (*connect.Response[pb.PortResponse], error) {
	resp, err := a.s.GeneratePort(ctx, r.Msg)
	if err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(resp), nil
}

func (a *ConnectServer) GenerateMac(ctx context.Context, r *connect.Request[pb.MacRequest]) (*connect.Response[pb.MacResponse], error) {
	resp, err := a.s.GenerateMac(ctx, r.Msg)
	if err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(resp), nil
}

func (a *ConnectServer) Slugify(ctx context.Context, r *connect.Request[pb.SlugifyRequest]) (*connect.Response[pb.SlugifyResponse], error) {
	resp, err := a.s.Slugify(ctx, r.Msg)
	if err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(resp), nil
}

func (a *ConnectServer) HiddenChars(ctx context.Context, r *connect.Request[pb.HiddenCharsRequest]) (*connect.Response[pb.HiddenCharsResponse], error) {
	resp, err := a.s.HiddenChars(ctx, r.Msg)
	if err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(resp), nil
}

func (a *ConnectServer) TextReplace(ctx context.Context, r *connect.Request[pb.TextReplaceRequest]) (*connect.Response[pb.TextReplaceResponse], error) {
	resp, err := a.s.TextReplace(ctx, r.Msg)
	if err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(resp), nil
}

func (a *ConnectServer) StringObfuscate(ctx context.Context, r *connect.Request[pb.StringObfuscateRequest]) (*connect.Response[pb.StringObfuscateResponse], error) {
	resp, err := a.s.StringObfuscate(ctx, r.Msg)
	if err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(resp), nil
}

func (a *ConnectServer) NumeronymGenerate(ctx context.Context, r *connect.Request[pb.NumeronymRequest]) (*connect.Response[pb.NumeronymResponse], error) {
	resp, err := a.s.NumeronymGenerate(ctx, r.Msg)
	if err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(resp), nil
}

func (a *ConnectServer) NatoAlphabet(ctx context.Context, r *connect.Request[pb.NatoRequest]) (*connect.Response[pb.NatoResponse], error) {
	resp, err := a.s.NatoAlphabet(ctx, r.Msg)
	if err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(resp), nil
}

func (a *ConnectServer) ListProcess(ctx context.Context, r *connect.Request[pb.ListRequest]) (*connect.Response[pb.ListResponse], error) {
	resp, err := a.s.ListProcess(ctx, r.Msg)
	if err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(resp), nil
}

func (a *ConnectServer) MathEval(ctx context.Context, r *connect.Request[pb.MathEvalRequest]) (*connect.Response[pb.MathEvalResponse], error) {
	resp, err := a.s.MathEval(ctx, r.Msg)
	if err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(resp), nil
}

func (a *ConnectServer) PercentageCalc(ctx context.Context, r *connect.Request[pb.PercentageRequest]) (*connect.Response[pb.PercentageResponse], error) {
	resp, err := a.s.PercentageCalc(ctx, r.Msg)
	if err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(resp), nil
}

func (a *ConnectServer) TempConvert(ctx context.Context, r *connect.Request[pb.TempConvertRequest]) (*connect.Response[pb.TempConvertResponse], error) {
	resp, err := a.s.TempConvert(ctx, r.Msg)
	if err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(resp), nil
}

func (a *ConnectServer) UnitConvert(ctx context.Context, r *connect.Request[pb.UnitConvertRequest]) (*connect.Response[pb.UnitConvertResponse], error) {
	resp, err := a.s.UnitConvert(ctx, r.Msg)
	if err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(resp), nil
}

func (a *ConnectServer) DateDiff(ctx context.Context, r *connect.Request[pb.DateDiffRequest]) (*connect.Response[pb.DateDiffResponse], error) {
	resp, err := a.s.DateDiff(ctx, r.Msg)
	if err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(resp), nil
}

func (a *ConnectServer) LeapYear(ctx context.Context, r *connect.Request[pb.LeapYearRequest]) (*connect.Response[pb.LeapYearResponse], error) {
	resp, err := a.s.LeapYear(ctx, r.Msg)
	if err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(resp), nil
}

func (a *ConnectServer) DateAdd(ctx context.Context, r *connect.Request[pb.DateAddRequest]) (*connect.Response[pb.DateAddResponse], error) {
	resp, err := a.s.DateAdd(ctx, r.Msg)
	if err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(resp), nil
}

func (a *ConnectServer) DateFormat(ctx context.Context, r *connect.Request[pb.DateFormatRequest]) (*connect.Response[pb.DateFormatResponse], error) {
	resp, err := a.s.DateFormat(ctx, r.Msg)
	if err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(resp), nil
}

func (a *ConnectServer) DateInfo(ctx context.Context, r *connect.Request[pb.DateInfoRequest]) (*connect.Response[pb.DateInfoResponse], error) {
	resp, err := a.s.DateInfo(ctx, r.Msg)
	if err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(resp), nil
}

func (a *ConnectServer) UrlParse(ctx context.Context, r *connect.Request[pb.UrlParseRequest]) (*connect.Response[pb.UrlParseResponse], error) {
	resp, err := a.s.UrlParse(ctx, r.Msg)
	if err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(resp), nil
}

func (a *ConnectServer) UserAgentParse(ctx context.Context, r *connect.Request[pb.UserAgentParseRequest]) (*connect.Response[pb.UserAgentParseResponse], error) {
	resp, err := a.s.UserAgentParse(ctx, r.Msg)
	if err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(resp), nil
}

func (a *ConnectServer) HttpStatusSearch(ctx context.Context, r *connect.Request[pb.HttpStatusSearchRequest]) (*connect.Response[pb.HttpStatusSearchResponse], error) {
	resp, err := a.s.HttpStatusSearch(ctx, r.Msg)
	if err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(resp), nil
}

func (a *ConnectServer) MimeLookup(ctx context.Context, r *connect.Request[pb.MimeLookupRequest]) (*connect.Response[pb.MimeLookupResponse], error) {
	resp, err := a.s.MimeLookup(ctx, r.Msg)
	if err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(resp), nil
}

func (a *ConnectServer) DockerRunToCompose(ctx context.Context, r *connect.Request[pb.DockerRunToComposeRequest]) (*connect.Response[pb.DockerRunToComposeResponse], error) {
	resp, err := a.s.DockerRunToCompose(ctx, r.Msg)
	if err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(resp), nil
}

func (a *ConnectServer) GitCheatSheet(ctx context.Context, r *connect.Request[pb.GitCheatSheetRequest]) (*connect.Response[pb.GitCheatSheetResponse], error) {
	resp, err := a.s.GitCheatSheet(ctx, r.Msg)
	if err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(resp), nil
}

func (a *ConnectServer) SvgOptimize(ctx context.Context, r *connect.Request[pb.SvgOptimizeRequest]) (*connect.Response[pb.SvgOptimizeResponse], error) {
	resp, err := a.s.SvgOptimize(ctx, r.Msg)
	if err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(resp), nil
}

func (a *ConnectServer) ExifRead(ctx context.Context, r *connect.Request[pb.ExifReadRequest]) (*connect.Response[pb.ExifReadResponse], error) {
	resp, err := a.s.ExifRead(ctx, r.Msg)
	if err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(resp), nil
}

func (a *ConnectServer) FileToBase64(ctx context.Context, r *connect.Request[pb.FileToBase64Request]) (*connect.Response[pb.FileToBase64Response], error) {
	resp, err := a.s.FileToBase64(ctx, r.Msg)
	if err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(resp), nil
}

func (a *ConnectServer) Base64ToFile(ctx context.Context, r *connect.Request[pb.Base64ToFileRequest]) (*connect.Response[pb.Base64ToFileResponse], error) {
	resp, err := a.s.Base64ToFile(ctx, r.Msg)
	if err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(resp), nil
}

func (a *ConnectServer) TokenCount(ctx context.Context, r *connect.Request[pb.TokenCountRequest]) (*connect.Response[pb.TokenCountResponse], error) {
	resp, err := a.s.TokenCount(ctx, r.Msg)
	if err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(resp), nil
}

func (a *ConnectServer) SpellCheck(ctx context.Context, r *connect.Request[pb.SpellCheckRequest]) (*connect.Response[pb.SpellCheckResponse], error) {
	resp, err := a.s.SpellCheck(ctx, r.Msg)
	if err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(resp), nil
}

func (a *ConnectServer) SpellLanguages(ctx context.Context, r *connect.Request[pb.SpellLanguagesRequest]) (*connect.Response[pb.SpellLanguagesResponse], error) {
	resp, err := a.s.SpellLanguages(ctx, r.Msg)
	if err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(resp), nil
}

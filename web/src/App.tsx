import { lazy } from 'react';
import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { Layout } from './components/Layout';

const DiffTool        = lazy(() => import('./components/DiffTool').then(m => ({ default: m.DiffTool })));
const Base64Tool      = lazy(() => import('./components/Base64Tool').then(m => ({ default: m.Base64Tool })));
const JsonTool        = lazy(() => import('./components/JsonTool').then(m => ({ default: m.JsonTool })));
const ConverterTool   = lazy(() => import('./components/ConverterTool').then(m => ({ default: m.ConverterTool })));
const BaseTool        = lazy(() => import('./components/BaseTool').then(m => ({ default: m.BaseTool })));
const GeneratorTool   = lazy(() => import('./components/GeneratorTool').then(m => ({ default: m.GeneratorTool })));
const TextTool        = lazy(() => import('./components/TextTool').then(m => ({ default: m.TextTool })));
const EncoderTool     = lazy(() => import('./components/EncoderTool').then(m => ({ default: m.EncoderTool })));
const TimeTool        = lazy(() => import('./components/TimeTool').then(m => ({ default: m.TimeTool })));
const DevTools        = lazy(() => import('./components/DevTools').then(m => ({ default: m.DevTools })));
const CronTool        = lazy(() => import('./components/CronTool').then(m => ({ default: m.CronTool })));
const CertTool        = lazy(() => import('./components/CertTool').then(m => ({ default: m.CertTool })));
const ColorTool       = lazy(() => import('./components/ColorTool').then(m => ({ default: m.ColorTool })));
const StringTool      = lazy(() => import('./components/StringTool').then(m => ({ default: m.StringTool })));
const SimilarityTool  = lazy(() => import('./components/SimilarityTool').then(m => ({ default: m.SimilarityTool })));
const SqlTool         = lazy(() => import('./components/SqlTool').then(m => ({ default: m.SqlTool })));
const IpTool          = lazy(() => import('./components/IpTool').then(m => ({ default: m.IpTool })));
const PasswordTool    = lazy(() => import('./components/PasswordTool').then(m => ({ default: m.PasswordTool })));
const MarkdownTool       = lazy(() => import('./components/MarkdownTool').then(m => ({ default: m.MarkdownTool })));
const DataValidatorTool  = lazy(() => import('./components/DataValidatorTool').then(m => ({ default: m.DataValidatorTool })));
const NetworkTool          = lazy(() => import('./components/NetworkTool').then(m => ({ default: m.NetworkTool })));
const EncodingCryptoTool   = lazy(() => import('./components/EncodingCryptoTool').then(m => ({ default: m.EncodingCryptoTool })));
const TextStringTool     = lazy(() => import('./components/TextStringTool').then(m => ({ default: m.TextStringTool })));
const MathUnitTool       = lazy(() => import('./components/MathUnitTool').then(m => ({ default: m.MathUnitTool })));
const DateTimeTool       = lazy(() => import('./components/DateTimeTool').then(m => ({ default: m.DateTimeTool })));
const WebDevOpsTool      = lazy(() => import('./components/WebDevOpsTool').then(m => ({ default: m.WebDevOpsTool })));
const MediaTool          = lazy(() => import('./components/MediaTool').then(m => ({ default: m.MediaTool })));
const Dashboard          = lazy(() => import('./components/Dashboard').then(m => ({ default: m.Dashboard })));
const HtmlMarkdownViewer = lazy(() => import('./components/HtmlMarkdownViewer').then(m => ({ default: m.HtmlMarkdownViewer })));
const TokenCounterTool   = lazy(() => import('./components/TokenCounterTool').then(m => ({ default: m.TokenCounterTool })));
const SpellCheckTool     = lazy(() => import('./components/SpellCheckTool').then(m => ({ default: m.SpellCheckTool })));

// A single Suspense fallback and ErrorBoundary wrap the routed Outlet in Layout,
// so routes render their lazy component directly.
function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/" element={<Layout />}>
          <Route index element={<Dashboard />} />
          <Route path="diff"      element={<DiffTool />} />
          <Route path="base64"    element={<Base64Tool />} />
          <Route path="json"      element={<JsonTool />} />
          <Route path="convert"   element={<ConverterTool />} />
          <Route path="base"      element={<BaseTool />} />
          <Route path="generators" element={<GeneratorTool />} />
          <Route path="text"      element={<TextTool />} />
          <Route path="encoder"   element={<EncoderTool />} />
          <Route path="time"      element={<TimeTool />} />
          <Route path="dev"       element={<DevTools />} />
          <Route path="cron"      element={<CronTool />} />
          <Route path="cert"      element={<CertTool />} />
          <Route path="color"     element={<ColorTool />} />
          <Route path="string"    element={<StringTool />} />
          <Route path="diff-text" element={<SimilarityTool />} />
          <Route path="sql"       element={<SqlTool />} />
          <Route path="ip"        element={<IpTool />} />
          <Route path="password"  element={<PasswordTool />} />
          <Route path="markdown"  element={<MarkdownTool />} />
          <Route path="validate"  element={<DataValidatorTool />} />
          <Route path="network"   element={<NetworkTool />} />
          <Route path="crypto"       element={<EncodingCryptoTool />} />
          <Route path="text-string"  element={<TextStringTool />} />
          <Route path="math"         element={<MathUnitTool />} />
          <Route path="datetime"     element={<DateTimeTool />} />
          <Route path="webdevops"    element={<WebDevOpsTool />} />
          <Route path="media"        element={<MediaTool />} />
          <Route path="viewer"       element={<HtmlMarkdownViewer />} />
          <Route path="tokens"      element={<TokenCounterTool />} />
          <Route path="spell"       element={<SpellCheckTool />} />
          <Route path="*"         element={<Navigate to="/" replace />} />
        </Route>
      </Routes>
    </BrowserRouter>
  );
}

export default App;

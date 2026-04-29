import { lazy, Suspense } from 'react';
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
const MarkdownTool    = lazy(() => import('./components/MarkdownTool').then(m => ({ default: m.MarkdownTool })));
const Dashboard       = lazy(() => import('./components/Dashboard').then(m => ({ default: m.Dashboard })));

function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/" element={<Layout />}>
          <Route index element={<Suspense><Dashboard /></Suspense>} />
          <Route path="diff"      element={<Suspense><DiffTool /></Suspense>} />
          <Route path="base64"    element={<Suspense><Base64Tool /></Suspense>} />
          <Route path="json"      element={<Suspense><JsonTool /></Suspense>} />
          <Route path="convert"   element={<Suspense><ConverterTool /></Suspense>} />
          <Route path="base"      element={<Suspense><BaseTool /></Suspense>} />
          <Route path="generators" element={<Suspense><GeneratorTool /></Suspense>} />
          <Route path="text"      element={<Suspense><TextTool /></Suspense>} />
          <Route path="encoder"   element={<Suspense><EncoderTool /></Suspense>} />
          <Route path="time"      element={<Suspense><TimeTool /></Suspense>} />
          <Route path="dev"       element={<Suspense><DevTools /></Suspense>} />
          <Route path="cron"      element={<Suspense><CronTool /></Suspense>} />
          <Route path="cert"      element={<Suspense><CertTool /></Suspense>} />
          <Route path="color"     element={<Suspense><ColorTool /></Suspense>} />
          <Route path="string"    element={<Suspense><StringTool /></Suspense>} />
          <Route path="diff-text" element={<Suspense><SimilarityTool /></Suspense>} />
          <Route path="sql"       element={<Suspense><SqlTool /></Suspense>} />
          <Route path="ip"        element={<Suspense><IpTool /></Suspense>} />
          <Route path="password"  element={<Suspense><PasswordTool /></Suspense>} />
          <Route path="markdown"  element={<Suspense><MarkdownTool /></Suspense>} />
          <Route path="*"         element={<Navigate to="/" replace />} />
        </Route>
      </Routes>
    </BrowserRouter>
  );
}

export default App;

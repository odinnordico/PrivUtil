import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { Layout } from './components/Layout';
import { DiffTool } from './components/DiffTool';
import { Base64Tool } from './components/Base64Tool';
import { JsonTool } from './components/JsonTool';
import { ConverterTool } from './components/ConverterTool';
import { GeneratorTool } from './components/GeneratorTool';
import { TextTool } from './components/TextTool';
import { EncoderTool } from './components/EncoderTool';
import { TimeTool } from './components/TimeTool';
import { DevTools } from './components/DevTools';
import { CronTool } from './components/CronTool';
import { CertTool } from './components/CertTool';
import { ColorTool } from './components/ColorTool';
import { StringTool } from './components/StringTool';
import { SimilarityTool } from './components/SimilarityTool';
import { SqlTool } from './components/SqlTool';
import { IpTool } from './components/IpTool';
import { PasswordTool } from './components/PasswordTool';
import { Dashboard } from './components/Dashboard';

function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/" element={<Layout />}>
          <Route index element={<Dashboard />} />
          <Route path="diff" element={<DiffTool />} />
          <Route path="base64" element={<Base64Tool />} />
          <Route path="json" element={<JsonTool />} />
          <Route path="convert" element={<ConverterTool />} />
          <Route path="generators" element={<GeneratorTool />} />
          <Route path="text" element={<TextTool />} />
          <Route path="encoder" element={<EncoderTool />} />
          <Route path="time" element={<TimeTool />} />
          <Route path="dev" element={<DevTools />} />
          <Route path="cron" element={<CronTool />} />
          <Route path="cert" element={<CertTool />} />
          <Route path="color" element={<ColorTool />} />
          <Route path="string" element={<StringTool />} />
          <Route path="diff-text" element={<SimilarityTool />} />
          <Route path="sql" element={<SqlTool />} />
          <Route path="ip" element={<IpTool />} />
          <Route path="password" element={<PasswordTool />} />
          <Route path="*" element={<Navigate to="/" replace />} />
        </Route>
      </Routes>
    </BrowserRouter>
  );
}

export default App;

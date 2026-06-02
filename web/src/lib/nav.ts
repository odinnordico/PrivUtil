import {
  FileDiff,
  Binary,
  LayoutDashboard,
  FileJson,
  ArrowLeftRight,
  ShieldCheck,
  Zap,
  AlignLeft,
  Link as LinkIcon,
  Clock,
  CalendarClock,
  Terminal,
  Palette,
  FileBadge,
  CaseSensitive,
  GitCompareArrows,
  Database,
  Network,
  Wifi,
  Key,
  Hash,
  FileCode,
  Lock,
  Shuffle,
  Calculator,
  Calendar,
  Globe,
  ImageIcon,
  Eye,
  Coins,
  SpellCheck,
  type LucideIcon
} from 'lucide-react';

export interface ToolItem {
  name: string;
  path: string;
  icon: LucideIcon;
  description: string;
}

export const navItems: ToolItem[] = [
  { 
    name: 'Dashboard', 
    icon: LayoutDashboard, 
    path: '/',
    description: 'Overview of all available tools'
  },
  { 
    name: 'Diff Utility', 
    icon: FileDiff, 
    path: '/diff',
    description: 'Compare text files and visualize differences'
  },
  { 
    name: 'Base64 Tool', 
    icon: Binary, 
    path: '/base64',
    description: 'Encode and decode Base64 strings'
  },
  { 
    name: 'JSON Formatter', 
    icon: FileJson, 
    path: '/json',
    description: 'Format, minify and validate JSON data'
  },
  {
    name: 'Universal Converter',
    icon: ArrowLeftRight,
    path: '/convert',
    description: 'Convert between JSON, YAML, XML, TOML, and CSV formats'
  },
  {
    name: 'Data Validator',
    icon: ShieldCheck,
    path: '/validate',
    description: 'Validate JSON, YAML, XML, and TOML with error highlighting'
  },
  { 
    name: 'Number Base', 
    icon: Hash, 
    path: '/base',
    description: 'Convert numbers across Decimal, Hexadecimal, Binary, etc.'
  },
  {
    name: 'Generators',
    icon: Zap,
    path: '/generators',
    description: 'Generate UUIDs, Lorem Ipsum, and Hashes'
  },
  {
    name: 'Encoding & Crypto',
    icon: Lock,
    path: '/crypto',
    description: 'HMAC, OTP/TOTP, ULID, Caesar cipher, text encoding, Morse code, Basic Auth'
  },
  { 
    name: 'Text Tools', 
    icon: AlignLeft, 
    path: '/text',
    description: 'Sort, manipulate, and inspect text content'
  },
  { 
    name: 'Encoders', 
    icon: LinkIcon, 
    path: '/encoder',
    description: 'URL and HTML entity encoding/decoding'
  },
  { 
    name: 'Time Converter', 
    icon: Clock, 
    path: '/time',
    description: 'Convert timestamps and dates'
  },
  { 
    name: 'Dev Utils', 
    icon: Terminal, 
    path: '/dev',
    description: 'JWT Debugger, Regex Tester, JSON to Go'
  },
  {
    name: 'Cron Tools',
    icon: CalendarClock,
    path: '/cron',
    description: 'Explain and test cron expressions'
  },
  { 
    name: 'Certificate', 
    icon: FileBadge, 
    path: '/cert',
    description: 'Parse and inspect X.509 certificates'
  },
  { 
    name: 'Color Converter', 
    icon: Palette, 
    path: '/color',
    description: 'Convert HEX, RGB, and HSL colors'
  },
  { 
    name: 'String Utils', 
    icon: CaseSensitive, 
    path: '/string',
    description: 'Case conversion and string escaping'
  },
  { 
    name: 'Similarity', 
    icon: GitCompareArrows, 
    path: '/diff-text',
    description: 'Calculate string similarity (Levenshtein)'
  },
  { 
    name: 'SQL Formatter', 
    icon: Database, 
    path: '/sql',
    description: 'Beautify and format SQL queries'
  },
  {
    name: 'IP Calc',
    icon: Network,
    path: '/ip',
    description: 'IPv4/IPv6 subnet calculator'
  },
  {
    name: 'Network Tools',
    icon: Wifi,
    path: '/network',
    description: 'chmod, IPv4 converter, range expander, port & MAC generator'
  },
  { 
    name: 'Password Generator', 
    icon: Key, 
    path: '/password',
    description: 'Generate secure random passwords'
  },
  {
    name: 'Text & String',
    icon: Shuffle,
    path: '/text-string',
    description: 'Slugify, hidden chars, find/replace, obfuscator, numeronym, NATO alphabet, list tools'
  },
  {
    name: 'Math & Units',
    icon: Calculator,
    path: '/math',
    description: 'Expression evaluator, percentage calculator, temperature & unit converter'
  },
  {
    name: 'Date & Time',
    icon: Calendar,
    path: '/datetime',
    description: 'Date diff, leap year checker, date arithmetic, formatters, and date info'
  },
  {
    name: 'Web & DevOps',
    icon: Globe,
    path: '/webdevops',
    description: 'URL parser, User-Agent, HTTP status codes, MIME types, Docker→Compose, Git cheat sheet'
  },
  {
    name: 'Media Tools',
    icon: ImageIcon,
    path: '/media',
    description: 'SVG optimizer, EXIF/image metadata reader, Base64 ↔ file converter'
  },
  {
    name: 'Markdown',
    icon: FileCode,
    path: '/markdown',
    description: 'Convert between Markdown and HTML'
  },
  {
    name: 'HTML/MD Viewer',
    icon: Eye,
    path: '/viewer',
    description: 'Safely render HTML or Markdown in a sandboxed iframe with file upload and CSP guardrails'
  },
  {
    name: 'Token Counter',
    icon: Coins,
    path: '/tokens',
    description: 'Count tokens for GPT-4o, Claude, Llama, Gemini, Mistral and classic tokenizers with side-by-side comparison'
  },
  {
    name: 'Spell & Grammar',
    icon: SpellCheck,
    path: '/spell',
    description: 'Offline spell- and grammar-checking for English and Spanish with inline fixes and a suggestions panel'
  },
];

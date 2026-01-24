import { 
  FileDiff, 
  Binary, 
  LayoutDashboard,
  FileJson,
  ArrowLeftRight,
  Zap,
  AlignLeft,
  Link as LinkIcon,
  Clock,
  Terminal,
  Palette,
  FileBadge,
  CaseSensitive,
  GitCompareArrows,
  Database,
  Network,
  Key,
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
    description: 'Convert between JSON, YAML, and XML formats'
  },
  { 
    name: 'Generators', 
    icon: Zap, 
    path: '/generators',
    description: 'Generate UUIDs, Lorem Ipsum, and Hashes'
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
    icon: Clock, 
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
    name: 'Password Generator', 
    icon: Key, 
    path: '/password',
    description: 'Generate secure random passwords'
  },
];

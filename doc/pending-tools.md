# Pending Tools

Tools identified from [omni-tools](https://github.com/iib0011/omni-tools) and [it-tools](https://github.com/corentinth/it-tools) that are suitable for privutil.

**Criteria applied:** developer-focused, runs fully server-side in Go, no external API calls, no heavy media processing (video/audio/PDF), no browser-only APIs. Lightweight media-adjacent tools that stay text/data-only are included (SVG, EXIF, Base64↔file).

Tools already covered by privutil are excluded (Base64, JSON format, JSON↔YAML↔XML, UUID, Lorem, Hash, RSA, Text manipulation, URL/HTML encode, Time convert, JWT, Regex, Cert parse, Color convert, Case convert, Levenshtein, SQL format, IP calc, Password gen, Markdown↔HTML).

---

## Data Conversion

| Tool | Description | Source |
|------|-------------|--------|
| ~~**TOML support in Converter**~~ | ~~Extend the existing JSON↔YAML↔XML converter to also handle TOML (json↔toml, yaml↔toml)~~ | ~~it-tools~~ |
| ~~**CSV Converter**~~ | ~~Convert CSV ↔ JSON, CSV ↔ YAML, CSV ↔ XML, CSV ↔ TSV~~ | ~~omni-tools / it-tools~~ |
| ~~**XML Validator**~~ | ~~Validate XML and report errors with line numbers~~ | ~~omni-tools~~ |
| ~~**YAML Validator**~~ | ~~Validate YAML and report errors~~ | ~~—~~ |
| **JSON to Go (struct tags)** | Already exists but extend: add JSON→TypeScript, JSON→Rust struct, JSON→Zod schema | it-tools |

---

## Network & System

| Tool | Description | Source |
|------|-------------|--------|
| ~~**chmod Calculator**~~ | ~~Interactive Unix permission calculator — click checkboxes, get octal and symbolic notation~~ | ~~it-tools~~ |
| ~~**IPv4 Address Converter**~~ | ~~Convert IPv4 to/from decimal integer, hex, and binary representations~~ | ~~it-tools~~ |
| ~~**IPv4 Range Expander**~~ | ~~Given a start and end IP, list all addresses or summarise as CIDR blocks~~ | ~~it-tools~~ |
| ~~**Random Port Generator**~~ | ~~Generate one or more random ports in a given range (optionally excluding well-known ports)~~ | ~~it-tools / omni-tools~~ |
| ~~**MAC Address Generator**~~ | ~~Generate random or OUI-specific MAC addresses~~ | ~~it-tools~~ |

---

## Encoding & Crypto

| Tool | Description | Source |
|------|-------------|--------|
| ~~**HMAC Generator**~~ | ~~Compute HMAC with SHA-256/SHA-512/SHA-1/MD5 from a message and secret key~~ | ~~it-tools~~ |
| ~~**OTP Generator / Validator**~~ | ~~Generate and validate TOTP/HOTP codes (RFC 6238/4226) — useful for debugging auth flows~~ | ~~it-tools~~ |
| ~~**ULID Generator**~~ | ~~Generate Universally Unique Lexicographically Sortable Identifiers~~ | ~~it-tools~~ |
| ~~**ROT13 / Caesar Cipher**~~ | ~~Encode/decode text with ROT13 or arbitrary Caesar shift~~ | ~~omni-tools~~ |
| ~~**Text to Binary / Hex**~~ | ~~Convert text to binary, hex, octal, and decimal codepoint representations~~ | ~~it-tools~~ |
| ~~**Text to Morse Code**~~ | ~~Encode/decode text to and from Morse code~~ | ~~omni-tools~~ |
| ~~**Basic Auth Generator**~~ | ~~Encode `user:password` as a Base64 `Authorization: Basic …` header value~~ | ~~it-tools~~ |

---

## Text & String

| Tool | Description | Source |
|------|-------------|--------|
| ~~**Slugify**~~ | ~~Convert any string to a URL-safe slug (lowercase, dashes, unicode-aware)~~ | ~~it-tools / omni-tools~~ |
| ~~**Hidden Character Detector**~~ | ~~Reveal zero-width spaces, non-breaking spaces, and other invisible Unicode characters~~ | ~~omni-tools~~ |
| ~~**Text Replacer**~~ | ~~Find-and-replace across multi-line text with plain or regex patterns~~ | ~~omni-tools~~ |
| ~~**String Obfuscator**~~ | ~~Partially mask a string (e.g. API keys, passwords) keeping only first/last N chars visible~~ | ~~it-tools~~ |
| ~~**Numeronym Generator**~~ | ~~Convert words to numeronyms: `internationalization` → `i18n`~~ | ~~it-tools~~ |
| ~~**NATO Alphabet**~~ | ~~Convert text to/from NATO phonetic alphabet~~ | ~~it-tools~~ |
| ~~**List Tools**~~ | ~~Deduplicate, sort, shuffle, find unique/most-frequent items in a newline-separated list~~ | ~~omni-tools / it-tools~~ |

---

## Math & Units

| Tool | Description | Source |
|------|-------------|--------|
| ~~**Math Evaluator**~~ | ~~Evaluate math expressions safely server-side (supports variables, functions)~~ | ~~it-tools~~ |
| ~~**Percentage Calculator**~~ | ~~`X% of Y`, `X is what % of Y`, `% change from X to Y`~~ | ~~it-tools~~ |
| ~~**Temperature Converter**~~ | ~~Convert between Celsius, Fahrenheit, and Kelvin~~ | ~~it-tools~~ |
| ~~**Byte / Unit Converter**~~ | ~~Convert between B, KB, MB, GB, TB (binary and SI) + length, mass, area, volume, speed~~ | ~~omni-tools~~ |

---

## Time & Date

| Tool | Description | Source |
|------|-------------|--------|
| ~~**Date Difference Calculator**~~ | ~~Calculate the duration between two dates in years, months, days, hours, minutes~~ | ~~omni-tools / it-tools~~ |
| ~~**Leap Year Checker**~~ | ~~Check whether one or more years are leap years~~ | ~~omni-tools~~ |
| ~~**Date Add/Subtract**~~ | ~~Add or subtract years/months/weeks/days/hours/minutes/seconds from a date~~ | ~~—~~ |
| ~~**Date Formatter**~~ | ~~Show a date in 20+ formats: ISO, RFC, Unix, ordinal, SQL, week date, etc.~~ | ~~—~~ |
| ~~**Date Info**~~ | ~~Week number, quarter, zodiac sign, season, day-of-year, days left in year/month~~ | ~~—~~ |

---

## Web & DevOps

| Tool | Description | Source |
|------|-------------|--------|
| ~~**URL Parser**~~ | ~~Break a URL into scheme, host, port, path, query params, fragment~~ | ~~it-tools~~ |
| ~~**User-Agent Parser**~~ | ~~Parse a `User-Agent` string into browser, OS, device fields~~ | ~~it-tools~~ |
| ~~**HTTP Status Codes Reference**~~ | ~~Searchable reference of all HTTP status codes with descriptions~~ | ~~it-tools~~ |
| ~~**MIME Type Lookup**~~ | ~~Look up MIME type by file extension and vice versa~~ | ~~it-tools~~ |
| ~~**Docker run → Compose**~~ | ~~Convert a `docker run` command string into an equivalent `docker-compose.yml` snippet~~ | ~~it-tools~~ |
| ~~**Git Cheat Sheet**~~ | ~~Quick searchable reference of common Git commands~~ | ~~it-tools~~ |

---

## Media-adjacent (lightweight, text/data only)

| Tool | Description | Notes |
|------|-------------|-------|
| ~~**SVG Optimizer**~~ | ~~Minify and clean SVG markup — remove comments, empty groups, redundant attributes, collapse whitespace~~ | ~~SVG is XML; pure text processing, no CGo~~ |
| ~~**Image Metadata (EXIF) Reader**~~ | ~~Extract EXIF/IPTC/XMP metadata from JPEG/PNG/WebP — camera model, GPS, dimensions, colour space~~ | ~~Small pure-Go lib (`rwcarlsen/goexif`), reads binary headers only, no image rendering~~ |
| ~~**Base64 ↔ File**~~ | ~~Encode an uploaded file to Base64 and decode Base64 back to a downloadable file~~ | ~~Extends existing Base64 handler to accept binary payloads; proto needs `bytes` field~~ |

---

## Implementation order (suggested)

Start with tools that extend existing proto/handler patterns — lowest friction to add:

1. ~~**TOML support in Converter**~~ ✓ Done
2. ~~**CSV Converter**~~ ✓ Done (JSON/YAML/XML/TOML/CSV multi-format converter with delimiter and header options)
3. ~~**XML Validator / YAML Validator**~~ ✓ Done (unified Data Validator for JSON, YAML, XML, TOML with line/col error reporting)
4. ~~**chmod Calculator**~~ ✓ Done
5. ~~**IPv4 Address Converter / Range Expander / Port Gen / MAC Gen**~~ ✓ Done (unified Network Tools tab)
6. ~~**HMAC / OTP / ULID / Caesar / Text Encode / Morse / Basic Auth**~~ ✓ Done (unified Encoding & Crypto tab)
7. ~~**Slugify / Hidden Chars / Text Replacer / Obfuscator / Numeronym / NATO / List Tools**~~ ✓ Done (unified Text & String tab)
8. ~~**Math Evaluator / Percentage / Temperature / Unit Converter**~~ ✓ Done (unified Math & Units tab)
9. ~~**Date Diff / Leap Year / Date Add / Date Format / Date Info**~~ ✓ Done (unified Date & Time tab)
10. ~~**URL Parser / User-Agent / HTTP Status / MIME Types / Docker→Compose / Git Cheat Sheet**~~ ✓ Done (unified Web & DevOps tab)
11. ~~**SVG Optimizer / Image Metadata (EXIF) / Base64↔File**~~ ✓ Done (unified Media Tools tab)

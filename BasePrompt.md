# PrivUtil

## Project Description

PrivUtil is a complete vive coding project using Gemini 3 Pro (High). The project is a stand allone app. The backend using golang. The front end using react and golang templating.
The comunication between the backend and the front end is done using grpc.

ITERATE OVER THE PROJECT UNTIL GET ALL THE FEATURES WORKING AND PRODUCTION READY.

Project go module: `github.com/odinnordico/privutil`

## Features

- Diff using the diff-match-patch algorithm. Consider using `github.com/google/go-cmp/cmp` for the diff if applies.
  - The diff should be displayed in a way that is easy to understand.
  - The diff should be displayed in a way that is easy to compare.
  - The diff should be displayed in a way that is easy to see the differences with colors.
- Base64 encode/decode.
- JSON format and unformat. The UI should support configurations for indentation and sorting.
- JSON to YAML convert and unconvert. The UI should support configurations for indentation and sorting.
- JSON to XML convert and unconvert. The UI should support configurations for indentation and sorting.
- JSON/YAML and XML Path or query. The UI should support configurations for JSON Path.
- YAML format and unformat. The UI should support configurations for indentation and sorting.
- XML format and unformat. The UI should support configurations for indentation and sorting.
- Sort lines. The UI should support configurations for sorting.
- Find and replace. The UI should support configurations for find and replace.
- Count lines, words, characters. The UI should support configurations for counting.
- Count similar lines. The UI should support configurations for counting similar lines line consider spaces and case sensitivity.
- JSON to GO struct convert and unconvert. The UI should support configurations for indentation and sorting.
- Regex tester. The UI should support configurations for regex tester.
- UUID/GUID Generator. The UI should support configurations for UUID/GUID Generator like version 1, 4, 5.
- Lorem Ipsum/Dummy Text Generator. The UI should support configurations for Lorem Ipsum/Dummy Text Generator like paragraphs, words, sentences.
- Hash/Checksum Calculator. The UI should support configurations for Hash/Checksum Calculator like MD5, SHA1, SHA256, SHA512.
- Cron Expression Generator. The UI should support configurations for Cron Expression Generator like seconds, minutes, hours, day of month, month, day of week.
- Cron Expression expander. The UI should support configurations for Cron Expression expander like seconds, minutes, hours, day of month, month, day of week.
- URL Encode/Decode. The UI should support configurations for URL Encode/Decode.
- HTML Entity Encode/Decode
- JWT (JSON Web Token) Debugger
- Certificate Parser (X.509)
- Unix Timestamp Converter
- Timezone Converter
- Color Converter (Hex, RGB, HSL, etc.)
- Case Converter. The UI should support configurations for case converter like uppercase, lowercase, title case, camel case, snake case, kebab case, etc.
- String Escaper/Unescaper. The UI should support configurations for string escaper/unescaper like HTML, URL, JSON, XML, etc.
- SQL Formatter. The UI should support configurations for SQL formatter like indentation and sorting.
- Markdown Preview. The UI should support configurations for markdown preview like indentation and sorting.
- IPv4 Subnet Calculator. The UI should support configurations for IPv4 Subnet Calculator like CIDR, netmask, etc.
- IPV4/IPV6 Converter. The UI should support configurations for IPV4/IPV6 Converter like CIDR, netmask, etc.
- Color Converter. The UI should support configurations for color converter like Hex, RGB, HSL, etc. With color picker

## Non Functional Requirements

- Iterate several times to improve the UI and UX.
- The UI should be responsive and work on different devices.
- The UI should be easy to use and understand.
- The UI should be fast and responsive.
- The UI should be accessible to users with disabilities.
- The UI should be secure.
- The UI should be reliable.
- The UI should be maintainable.
- The UI should be scalable.
- The UI should be testable.
- The UI should be documented.
- The UI should be versioned.
- The UI should be licensed.
- The UI should be supported.
- The UI should be maintained.
- The UI should be scalable.
- The UI should be testable.
- The UI should be documented.
- The UI should be versioned.
- The UI should be licensed.
- The UI should be supported.
- The GitHub actions should publish binaries for Linux, Windows and MacOS.

## Tech Stack

- Backend: Golang
- Frontend: React
- Communication: gRPC

## UI/UX

- The UI should be modern and user-friendly.
- The UI should be easy to navigate.
- The UI should be easy to understand.
- The UI should be easy to use.
- The UI should be easy to maintain.
- The UI should be easy to test.
- The UI should be easy to document.
- The UI should be easy to version.
- The UI should be easy to license.
- The UI should be easy to support.
- The UI should be easy to maintain.
- The UI should be easy to scale.
- The UI should be easy to test.
- The UI should be easy to document.
- The UI should be easy to version.
- The UI should be easy to license.
- The UI should be easy to support.

## Project Structure

- The project should be organized in a way that is easy to understand.
- The project should be organized in a way that is easy to maintain.
- The project should be organized in a way that is easy to test.
- The project should be organized in a way that is easy to document.
- The project should be organized in a way that is easy to version.
- The project should be organized in a way that is easy to license.
- The project should be organized in a way that is easy to support.

## Testing

- The project should be tested.
- The tests should be easy to understand.
- The tests should be easy to maintain.
- The tests should be easy to run.
- The tests should be easy to document.
- The tests should be easy to version.
- The tests should be easy to license.
- The tests should be easy to support.

## Coverage

- The coverage should be at least 80%.

## Documentation

- The project should be documented, the front and back end.
- The project should have a submodule for the wiki.

## License

- The project should be licensed under the MIT license.

## GitHub Files

- The project should have the following files with working and production ready content. And add others that considered required or necessary.
  - README.md
  - LICENSE
  - .gitignore
  - .github/FUNDING.yml
    - |
      github: odinnordico
  - .github/ISSUE_TEMPLATE.md
  - .github/dependabot.yml
    - |

      # To get started with Dependabot version updates, you'll need to specify which

      # package ecosystems to update and where the package manifests are located.

      # Please see the documentation for all configuration options:

      # https://docs.github.com/code-security/dependabot/dependabot-version-updates/configuration-options-for-the-dependabot.yml-file

      version: 2
      updates:
      - package-ecosystem: "" # See documentation for possible values
        directory: "/" # Location of package manifests
        schedule:
        interval: "monthly"

  - wiki/
  - .gitmodules
    - |
      [submodule "wiki"]
      path = wiki
      url = https://github.com/odinnordico/privutil.wiki.git
  - .github/workflows/build.yml
  - .github/workflows/test.yml
  - .github/workflows/lint.yml
  - .github/workflows/security.yml
  - .github/workflows/publis.yml

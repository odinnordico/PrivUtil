# Dictionary Attribution

The frequency word lists under `dict/` (`en.txt`, `es.txt`) are derived from
the **FrequencyWords** project by hermitdave
(https://github.com/hermitdave/FrequencyWords), which is itself built from the
**OpenSubtitles** corpus.

- FrequencyWords is distributed under the MIT License.
- Each list is the top ~50,000 words with their occurrence counts. The counts
  are used both for membership testing and for ranking spelling suggestions.
- Contractions appear in de-apostrophe'd form in these lists (e.g. `dont`,
  `youre`); the engine accounts for this when validating apostrophe-bearing
  tokens (see `acceptedContraction`).

To add a language, drop a `<code>.txt` frequency list here and register a
language pack in the package's `init()`.

# Writing Style Guide

Reference for how to write NVIDIA Run:ai documentation. `docs/AGENTS.md`
covers workflow — what to do and where content lives. This file covers the
actual writing rules.

## Claim Accuracy

- Verify every field name, default value, behavior, CLI flag, API endpoint,
  and UI label against the source code or an API spec before documenting it.
  Do not infer or extrapolate from indirect signals — if you cannot point to
  something concrete supporting a claim, do not make it.

## Style

Apply these rules to documentation, examples, headings, UI text, and release
notes that you create or edit.

- Capitalize a Run:ai resource-type noun (Cluster, Department, Project,
  Scope, Tenant, Node Pool, Training, Inference, and similar constructs) when
  naming a specific UI page, table, or element. Use lowercase when the word
  is an ordinary descriptive noun in running prose — for example, "inference
  workloads" (generic, lowercase) versus "Inference" as it appears in the
  workload grid (naming the UI element, capitalized).
- Write in a professional, active, conversational, and engaging voice.
- Use active voice whenever possible. Use present tense for product behavior.
  Address the reader in second person as "you."
- Keep sentences concise. Prefer sentences with fewer than 30 words.
- Use plain English and precise technical terms. Avoid jargon, filler,
  colloquialisms, and flowery marketing claims.
- Avoid contractions in technical documentation. Write "do not," "cannot,"
  and "it is."
- Write "NVIDIA" in all caps and use "an NVIDIA," not "a NVIDIA."
- Spell out uncommon abbreviations on first use. Spell out LLM, RAG, SLM, VLM,
  and MoE on first use.
- Use NVIDIA spellings such as data center, dataset, open source, pretrained,
  startup, webpage, website, and Wi-Fi.
- Replace Latinisms with plain English. Use "for example," "that is," "and so
  on," "through," and "compared to."
- Use "refer to" instead of "see," and "after" instead of "once" for time.
  Use "can" for capability or permission ("you can configure..."). Use
  "may" only for an uncertain or conditional outcome ("the workload may be
  preempted"). Do not use "may" to express capability.
- Do not use "please" in technical instructions.
- Use numerals for specific values, parameters, measurements, and values of 10
  or more. Spell out zero through nine in general prose.
- Include a space between a number and its unit. Use a comma in numbers with
  four or more digits.
- Use title case for headings. Do not style headings with code, bold, italics,
  quotation marks, ampersands, or exclamation marks.
- Use the Oxford comma. Put periods inside quotation marks in U.S. style.
- Use hyphens only for compound modifiers before nouns, except for terms
  approved in `docs/terminology.md` that are hyphenated regardless of
  position — for example `non-preemptible`, `over-quota`, `single-GPU
  fraction`. Do not hyphenate an adverb that ends in "ly."
- Format commands, code, filenames, paths, and API identifiers as code. Use
  bold for UI elements and the greater-than sign for UI navigation.
- Introduce lists, tables, code examples, and images with a complete sentence.
  Use parallel construction in lists.
- Write dates as Month DD, YYYY. Omit the year when it matches the publication
  year. Write time with a 12-hour clock and include minutes only when needed.
- Do not rewrite quoted UI labels, API field names, or audience role labels in
  tables to enforce second person.
- Provide useful alt text and preserve a logical heading hierarchy.
- Do not rewrite literal code, identifiers, commands, URLs, or quoted terminal
  and API output to satisfy prose rules.
- Apply rules to improve clarity. Do not make mechanical changes that reduce
  technical accuracy or readability.
- Put the condition before the instruction: "To enable GPU fractions, set
  `gpuFractionEnabled: true`" — not the reverse.
- One action per numbered step.
- Remove filler words: "simply," "just," "in order to," "it is important to
  note."
- Remove marketing language: "seamless," "powerful," "cutting-edge," "unlock,"
  "leverage," "comprehensive."
- Name the thing. Write the flag, field, or menu path — not "configure the
  appropriate settings."

## Fern Components

Use Fern MDX components. Do not use raw HTML. Always leave a blank line before
and after opening and closing tags.

Never use plain text for callouts or warnings. Every note, tip, warning, or
important callout must use a Fern component — not bold text, blockquotes, or
any other markdown construct.

- `<Note>` — notes, supplemental information, and short informational
  callouts; open body with **Note**; typically a single sentence, but a
  brief bulleted list is acceptable for grouped short facts
- `<Warning>` — caveats, prerequisites, breaking changes; use **Important**,
  **Prerequisite**, **Warning**, or **Breaking change from version X.Y:**
- `<Accordion title="...">` — optional detail most readers can skip
- `<Tabs>` / `<Tab title="...">` — steps that differ by interface; standard
  titles are `UI`, `CLI v2`, `API`
- `<Cards>` / `<Card>` — landing and index pages only; not in content pages
- `<Badge>` — inline feature lifecycle: `Experimental` (`success`), `Beta`
  (`warning`), `Legacy` (`note`)

## Links

- Within `docs/`: use relative paths with the file extension.
- To the REST API reference: use the versioned API docs URL.
- To external content: use the full URL.
- Use descriptive link text that names the destination. Do not use raw URLs
  in running text or generic link text such as "click here" or "read more."
- UI navigation paths: bold element names and use → as the separator.
- Linking to installation content (a prerequisite, a feature, a
  configuration step) from outside
  `docs/installation-and-configuration/installation/`: default to the Guided
  Installation page (`guided-installation/prerequisites-and-features.md`,
  etc.). Only link to `install-using-helm/` when the content is genuinely
  Helm-specific with no Guided equivalent — see
  `docs/.agents/skills/runai-docs-installation/SKILL.md` for how the two
  install paths differ.

# Torvecode brand and skills

The terminal identity follows the supplied Torve AI cube robot: lime #C7FF24,
mint #40F887, cyan #00DCC4 and white #F7FAFC. Terminal glyphs replace bitmap
rendering so the mark works without a special image protocol. Compact terminals
use a text signature. Authentication and workspace share that signature.

## Write and add skills

```sh
torve skills create code-review
# Edit the printed SKILL.md path in your editor.
torve skills add testing ./my-testing-SKILL.md
torve skills create writing --global
torve skills list
```

Project skills live at `.torvecode/skills/<name>/SKILL.md`. Personal skills live
under the OS config directory at `torvecode/skills/<name>/SKILL.md`. Project skills
win when names match. Skill creation never overwrites an existing directory.
Use lowercase letters, digits and hyphens in names. Documents are limited to 64 KiB.
The add command copies the document only; supporting scripts/assets are not imported.

In chat, say `Use the code-review skill to review these changes`. The agent gets
an inventory of available skills and reads the selected document using its file
read tool. Skills do not grant tool permissions. After adding a skill, start a
new session so the agent receives the updated inventory. Edit SKILL.md directly
to change it; remove its directory to uninstall it.

## Slash command palette

Press `/` in an empty chat prompt to open the Torve command palette. Continue
typing to filter commands, use the arrow keys to move, and press Enter to run.
The palette includes account connection, models, sessions, new session, themes,
external editor, help, skills, MCP status, workspace status, code review, logs,
project initialization and context compaction. `Ctrl+K` remains available as a
second way to open the same palette. A slash typed inside a message remains text.

## Images and documents

Use `Ctrl+F` or `/image` to attach PNG, JPG or WebP files. `Ctrl+V` detects the
clipboard content: it inserts text into the prompt or attaches a copied image.
Windows works through its built-in clipboard API; macOS requires `pngpaste`,
and Linux uses `wl-paste` or `xclip` for images.

Document attachments use Microsoft's MIT-licensed MarkItDown locally. Torvecode
does not bundle Python because that would break the lightweight binary target.
Install Python 3.10+ and run:

```sh
pip install 'markitdown[pdf,docx,pptx,xlsx,xls,outlook]'
torve documents status
torve documents convert report.pdf -o report.md
```

PDF, Word, PowerPoint, Excel, HTML, CSV, JSON, XML, EPUB, ZIP and Outlook files
can then be selected with `Ctrl+F`. Conversion has a 60-second, 5 MB input and
512 KiB Markdown-output limit to keep request context bounded.

Torve's gateway remains authoritative for plan limits. RPM, TPM, concurrency,
five-hour and monthly limits are not silently retried by the CLI. The CLI shows
the reset delay returned by the gateway. Insufficient balance tells the user to
add balance or purchase a plan in the Torve dashboard.

## Feature inventory

Implemented in source: sessions/history, streaming chat, file editing, diff,
shell tools and permissions, model picker, custom commands, project memory,
optional MCP and LSP, website/API-key login, and local skill creation/import.

A built-in visual skill editor, remote skill marketplace, and bundled skills
are not implemented. Model compatibility does not by itself verify the quality
of every model's tool calls. End-to-end skill invocation needs a live model test.

## Validation

TUI tests cover onboarding navigation and widths 40/80/120. Skills tests cover
creation, overwrite protection and invalid path names. Native Windows build is
produced separately from the published beta.4 model fix; this design and skills
change requires a new npm release.

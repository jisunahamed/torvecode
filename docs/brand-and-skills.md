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

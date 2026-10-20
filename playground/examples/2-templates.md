---
slides: true
paginate: true
title: Quarterly review
data:
  team: Platform
  quarters: [Q1, Q2, Q3, Q4]
  goals:
    - Ship the playground
    - Halve the build time
    - Write the docs
---

# {{ .title }} — {{ .team | upper }}

{{ len .goals }} goals, {{ len .quarters }} quarters.

{{ range .quarters }}
---

## {{ . }}

{{ range $.goals }}- {{ . }}
{{ end }}
{{ end }}

---

## Generated, then numbered

This deck has {{ total }} slides: four of them come from one
`{{ "{{ range .quarters }}" }}` block.

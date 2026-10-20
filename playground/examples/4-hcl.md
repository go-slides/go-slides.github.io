---hcl
slides   = true
paginate = true
data = {
  company = "acme"
  year    = 2026
}
footer = "${upper(data.company)} · ${data.year}"

kind "big" {
  class = "lead"
}
---

<!-- _kind: big -->

# HCL front matter

Expressions and functions where YAML only has values.

---

## The footer is computed

`footer = "${upper(data.company)} · ${data.year}"`

Look at the bottom of this slide.

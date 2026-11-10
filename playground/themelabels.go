package playground

import (
	"path"
	"strings"
)

// The words of the theme editor's controls, in the language in use: a property
// or a field of theme.hcl by its id, an option of an enumeration by its value.

var propLabelKeys = map[string]Key{
	"background": KThemePBackground, "text": KThemePText, "heading": KThemePHeading, "accent": KThemePAccent, "link": KThemePLink,
	"font-body": KThemePFontBody, "font-heading": KThemePFontHeading, "font-code": KThemePFontCode,
	"size-body": KThemePSizeBody, "size-heading": KThemePSizeHeading, "size-code": KThemePSizeCode,
	"weight-body": KThemePWeightBody, "weight-heading": KThemePWeightHeading,
	"logo-placement": KThemePLogoPlace, "logo-size": KThemePLogoSize,
	"pagination-placement": KThemePPagePlace, "pagination-size": KThemePPageSize,
	"footer-align": KThemePFooterAlign, "footer-size": KThemePFooterSize,
}

var metaLabelKeys = map[string]Key{
	"name": KThemeMName, "version": KThemeMVersion, "description": KThemeMDescription, "authors": KThemeMAuthors,
	"license": KThemeMLicense, "homepage": KThemeMHomepage, "extends": KThemeMExtends, "min_slides": KThemeMMinSlides,
}

var optionKeys = map[string]Key{
	"none": KThemeOptNone, "top-left": KThemeOptTopLeft, "top-right": KThemeOptTopRight, "top-center": KThemeOptTopCenter,
	"bottom-left": KThemeOptBottomLeft, "bottom-right": KThemeOptBottomRight, "bottom-center": KThemeOptBottomCenter,
	"normal": KThemeOptNormal, "italic": KThemeOptItalic, "oblique": KThemeOptOblique,
	"background": KThemeOptBackground, "left": KThemeOptLeft, "center": KThemeOptCenter, "right": KThemeOptRight,
}

// themePropLabel is a property's caption.
func themePropLabel(id string) string {
	if k, ok := propLabelKeys[id]; ok {
		return T(k)
	}
	return id
}

// themeMetaLabel is a field of theme.hcl's caption (a font's licence by its
// "font-license:Family" id).
func themeMetaLabel(field string) string {
	if k, ok := metaLabelKeys[field]; ok {
		return T(k)
	}
	if len(field) > len("font-license:") && field[:len("font-license:")] == "font-license:" {
		return T(KThemeMFontLicense, field[len("font-license:"):])
	}
	if f, ok := strings.CutPrefix(field, "font-weight:"); ok {
		return T(KThemeMFontWeight, path.Base(f))
	}
	if f, ok := strings.CutPrefix(field, "font-style:"); ok {
		return T(KThemeMFontStyle, path.Base(f))
	}
	return field
}

// ThemeOptionLabel is an enumeration option said in the language: the value
// "top-left" as "en haut à gauche".
func ThemeOptionLabel(value string) string {
	if k, ok := optionKeys[value]; ok {
		return T(k)
	}
	return value
}

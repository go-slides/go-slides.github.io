package playground

// enMessages is the reference catalogue: every other language is checked
// against its keys and its placeholders.
var enMessages = map[Key]string{
	KTabHome: "Home", KTabInsert: "Insert", KTabFormat: "Format", KTabView: "View", KTabExport: "Export",

	KGrpDeck: "Deck", KGrpLook: "Look", KGrpPresent: "Present", KGrpDraft: "Draft", KGrpSlides: "Slides",
	KGrpPictures: "Pictures", KGrpText: "Text", KGrpPreview: "Preview", KGrpEditor: "Editor", KGrpSaveAs: "Save as",
	KGrpParagraph: "Paragraph", KGrpTable: "Table", KGrpRowsCols: "Rows and columns", KGrpColumn: "Column",
	KGrpSlideText: "Slide text", KGrpLink: "Link",

	KLblExample: "Example", KLblReset: "Reset", KLblTheme: "Theme", KLblThemeOption: "Theme: %s", KLblDraft: "Draft",
	KLblShare: "Share link", KLblPresent: "Present", KLblPreview: "Preview", KLblMinimap: "Minimap",
	KLblSlides: "Slides", KLblFiles: "Files", KLblHelp: "Help", KLblNewSlide: "New slide", KLblImage: "Image",
	KLblBackground: "Background", KLblBgLeft: "Left split background", KLblBgRight: "Right split background",
	KLblMath: "Math", KLblCode: "Code block", KLblKind: "Kind", KLblHTML: "HTML page", KLblPDF: "PDF",
	KLblPPTX: "PowerPoint", KLblODP: "OpenDocument",

	KTipExample:    "Start from one of the examples. Your text stays in this browser: Draft › Restore replaced brings it back.",
	KTipTheme:      "The deck's theme: choosing one writes it into the deck's front matter.",
	KTipDraft:      "Restore, forget or reset the deck this browser keeps for you.",
	KTipReset:      "Put the example's text back; your deck is kept as a draft.",
	KTipPresent:    "Show the slideshow here and open the presenter view in its own window.",
	KTipShare:      "Copy a link that carries the deck itself: nothing is uploaded.",
	KTipKind:       "Give the slide a kind: a template with its own look, set by a <!-- _kind: … --> comment.",
	KTipNewSlide:   "Add an empty slide after the current one.",
	KTipImage:      "Choose an image file: it goes into the deck as a data: URI, nothing is uploaded. Selected text becomes its description.",
	KTipBackground: "Choose an image for the slide's background; the arrow fills only the left or the right 40%.",
	KTipMath:       "A display formula between $$ lines; the selection goes inside it.",
	KTipCode:       "A fenced code block with line numbers ({lineNumbers}); the selection goes inside it.",
	KTipMode:       "Overview stacks every slide and scrolls with the editor; Slideshow is the presentation itself.",
	KTipMinimap:    "A map of the deck beside the editor's scrollbar.",
	KTipThumbs:     "A column of slide thumbnails; click one to go to it.",
	KTipFiles:      "The project's files: themes, kinds, data and media. A narrow window starts without it.",
	KTipHelp:       "The help for what the caret is on, and the documentation links. A narrow window starts without it.",
	KTipHTML:       "One HTML file that presents the deck in any browser.",
	KTipPDF:        "A PDF, one page per slide, made here in your browser.",
	KTipPPTX:       "An editable PowerPoint file, made here in your browser.",
	KTipODP:        "An editable OpenDocument presentation, made here in your browser.",

	KModeOverview: "Overview", KModeSlideshow: "Slideshow",

	KExWelcome: "Welcome", KExTemplates: "Templates: generated slides", KExKinds: "Slide kinds",
	KExHCL: "HCL front matter", KExGaia: "Themes and spot directives", KExMedia: "Math, images and backgrounds",
	KExGalleryDefault: "Gallery: the default theme", KExGalleryGaia: "Gallery: the gaia theme",
	KExGalleryUncover: "Gallery: the uncover theme", KExGalleryMath: "Gallery: math",
	KExGalleryCode:       "Gallery: code, line numbers and callouts",
	KExGalleryEmoji:      "Gallery: emoji and fitting headings",
	KExGalleryPresenting: "Gallery: transitions and speaker notes",
	KExGalleryHCL:        "Gallery: HCL expressions and typed data",

	KStatusStarting:           "Starting…",
	KStatusOpenedShared:       "Opened the deck shared in the link.",
	KStatusSharedUnreadable:   "The link's deck could not be read: %s",
	"status.rendered.one":     "%d slide · rendered in %d ms",
	"status.rendered.other":   "%d slides · rendered in %d ms",
	KStatusStale:              "the preview shows the last deck that rendered",
	"status.problems.one":     "%d problem · %s",
	"status.problems.other":   "%d problems · %s",
	KStatusDraftNotKept:       " · ⚠ draft not kept in this browser",
	KStatusPresenting:         "Presenting: the presenter view is in its own window; arrow keys in either move both.",
	KStatusLinkCopied:         "Link copied: the deck travels in the link itself (%d characters), nothing is uploaded.",
	KStatusSavedHTML:          "Saved deck.html: open it in any browser, or present from it.",
	KStatusShareDeckOnly:      " Only the deck is in the link: Project › Download project (.zip) keeps its files.",
	KStatusRestoredDraft:      "Restored your draft of %s.",
	KStatusSharedKeptDraft:    "Opened the deck shared in the link. Your draft of %s is kept: Draft › Restore my draft.",
	KStatusRestoredDraftReset: "Restored your draft of %s. Reset goes back to the example.",
	KStatusDraftSavedAgain:    "Your draft is saved in this browser again.",
	KStatusDraftNotSaved:      "⚠ This browser did not keep your draft (%s): Share link or Download HTML keeps the deck.",
	KStatusOtherTab:           "Another tab saved its draft %s. Your edits here will not erase it: Draft › Restore replaced.",
	KStatusRestoredTheDraft:   "Restored the draft of %s.",
	KStatusForgot:             "Forgot your drafts: nothing of this deck is stored in this browser now (until you edit it again).",
	KStatusNoExport:           "⚠ This page cannot export: use the slides command.",
	KStatusExportRunning:      "An export is running: %s.",
	KStatusTooLargeExport:     "⚠ The deck is too large to export here: use the slides command.",
	KStatusPreparing:          "Preparing the %s…",
	KStatusLoadingEnginePct:   "Loading the export engine (once; your browser keeps it): %d%% of %.1f MiB",
	KStatusLoadingEngine:      "Loading the export engine (once; your browser keeps it): %.1f MiB",
	KStatusExportProgress:     "%s: slide %d of %d…",
	KStatusExportFailed:       "⚠ The %s export failed: %s",
	"status.saved.one":        "Saved %s: %s, %d slide in %.1f s, made in your browser.",
	"status.saved.other":      "Saved %s: %s, %d slides in %.1f s, made in your browser.",
	KStatusIsMedium:           "%s is a medium: the deck shows it with %s (drop it on the editor, or its menu: Insert into the deck).",
	KStatusMadeTheme:          "Made %s, built on %s: choose %s in the Theme list to use it.",
	KStatusMadeKind:           "Made %s: a slide takes the kind with <!-- _kind: %s -->.",
	KStatusMadeData:           "Made %s: the deck reads it as {{ .title }}.",
	KStatusCannotSaveFile:     "⚠ this page cannot save a file",
	KStatusSavedZip:           "Saved %s: unzipped, `slides build deck.md` builds it, and the VS Code extension opens it.",
	KStatusOpenedZip:          "Opened the .zip: %d files.",
	KStatusOpenedFolder:       "Opened the folder: %d files.",
	KStatusAddedSlide:         "Added a slide after the current one.",
	KStatusInsertedFormula:    "Inserted a display formula: type inside it.",
	KStatusInsertedCode:       "Inserted a Go code block with line numbers: type inside it.",
	KStatusPutFormula:         "Put the selection in a display formula.",
	KStatusPutCode:            "Put the selection in a Go code block with line numbers.",
	KStatusNoFilePicker:       "⚠ This page cannot ask for a file: drop or paste an image file into the editor.",
	KStatusKindSet:            "The slide is now of kind %s.",
	KStatusInsertedRule:       "Inserted a horizontal rule (***: three dashes on a line would start a new slide).",
	KStatusInsertedTable:      "Inserted a table: type in its cells; the table-aligning commands keep it tidy.",
	KStatusJournalLarge:       "This project is too large for the quick save: the edits of the last half-second before you close the page may not be kept.",
	KStatusJournalRefused:     "⚠ The quick save is refused (%s): the edits of the last half-second before you close the page may not be kept.",
	KStatusAdded:              "Added %s to the project (%d KiB): it is kept in this browser and goes in the project's .zip.",
	KStatusAddedPut:           "Added %s to the project (%d KiB) and put it in the deck: it is kept in this browser and goes in the project's .zip.",
	KStatusPopupBlocked:       "⚠ The browser blocked the presenter window: allow pop-ups for this page, then Present again.",

	KWhyHeading:    "A heading goes on a line of the deck's text, not in the front matter or a code block.",
	KWhyList:       "A list goes on a line of the deck's text, not in the front matter or a code block.",
	KWhyQuote:      "A quote goes on a line of the deck's text, not in the front matter or a code block.",
	KWhyNotInTable: "Put the caret in a table first.",
	KWhyNotInBody:  "Put the caret in a row of the table's body: the header stays.",
	KWhyLastColumn: "A table keeps at least one column.",

	KProblemLine: "line %d", KProblemError: "error", KProblemWarning: "warning",

	KConfirmReset:      "Replace your deck with the example? Your text stays in this browser: Draft › Restore replaced.",
	KConfirmShareFiles: "A link carries the deck only, not its other files (themes, kinds, media): download the project as a .zip instead?",
	KConfirmOpenZip:    "Replace your deck and its files with the .zip's? What you have now stays in this browser: Draft › Restore replaced.",
	KConfirmOpenFolder: "Replace your deck and its files with the folder's? What you have now stays in this browser: Draft › Restore replaced.",

	KDateToday: "today at %s", KDateAt: "%s at %s", KDateShort: "%[1]s %[2]d",
	KDateMonths: "Jan,Feb,Mar,Apr,May,Jun,Jul,Aug,Sep,Oct,Nov,Dec",
	KBytes:      "%d bytes", KWarn: "⚠ %s",

	KDraftRestoreMine: "Restore my draft (%s)", KDraftRestoreReplaced: "Restore replaced (%s)",
	KDraftForget: "Forget my drafts", KDraftResetExample: "Reset to the example", KDraftNothing: "Nothing saved yet",

	KSaveAs: "Save as…", KKindTitle: "Kind…",

	KProject: "Project", KFilesTreeLabel: "Project files", KFilesNewThemeFrom: "New theme (from %s)",
	KFilesNewKind: "New kind", KFilesNewData: "New data file", KFilesNewFile: "New file…", KFilesNewFolder: "New folder…",
	KFilesAddMedia: "Add media…", KFilesOpenZip: "Open project (.zip)…", KFilesOpenFolder: "Open project folder…",
	KFilesDownloadZip: "Download project (.zip)", KFilesInsertIntoDeck: "Insert into the deck", KFilesDownload: "Download",
	"files.usage.one": "%d file · %s of %d MiB", "files.usage.other": "%d files · %s of %d MiB",

	KPaneProblems: "Problems", KPaneDocumentation: "Documentation",
	KFrameTitlePreview: "Slide preview", KFrameTitleRail: "Slide thumbnails",

	KDocGettingStarted: "Getting started", KDocMarp: "Marp compatibility", KDocTemplates: "Templates",
	KDocKinds: "Kinds", KDocHCL: "HCL", KDocCLI: "Command line", KDocVSCode: "VS Code", KDocPlayground: "Playground",

	KTableColumn: "Column %d",

	KErrExists: "%s exists already", KErrDeckName: "the deck stays deck.md: the slides command and the playground read it by that name",
	KErrNoSuchFile: "%s: no such file", KErrNotFileName: "%q is not a file name",
	KErrDeckTooLarge: "the deck is %d KiB; the playground renders decks up to %d KiB",
	KErrMaxFiles:     "a project holds at most %d files", KErrPathReason: "%s: %v",
	KErrProjectOver: "the project would be over %d MiB", KErrFileSize: "%s is %d KiB: a file is at most %d KiB",
	KErrShareTooLargeDeck: "the deck is larger than the playground takes; download it instead",
	KErrShareTooLargeLink: "the deck is too large to travel in a link; download it instead",
	KErrShareNoDeck:       "the link carries no deck", KErrShareTooLong: "the link's deck is too long",
	KErrShareDamaged: "the link's deck is damaged (it may have been cut)",
	KErrShareLarger:  "the link's deck is larger than the playground takes", KErrShareNotText: "the link's deck is not text",
	KErrZipNoDeck: "the archive holds no deck: put deck.md at its top (or one .md file)",
	KErrZipSize:   "the archive is %d MiB: at most %d MiB", KErrZipNot: "not a .zip archive: %v",
	KErrZipEntries: "the archive lists %d entries: at most %d", KErrZipLinks: "%s: links are not taken",
	KErrZipFileOver: "%s is over %d KiB: a file is at most that",
	KErrZipBomb:     "the archive inflates more than %d times: refused as a possible zip bomb",
	KErrZipTwice:    "%s is in the archive twice", KErrZipMoreThan: "more than %d KiB",
	KErrFileOver: "%s is over %s", KErrFailed: "failed",
	KErrFolderCount: "the folder holds %d files: a project holds at most %d", KErrFolderOver: "the folder is over %s",
	KErrFolderMore: "the folder holds more than %d files", KErrNotImage: "%s is not an image.",

	// the errors of the store, media and templates packages, by their code (coded.Error)
	"err.store.pathlen":     "%q: a path is 1 to %d bytes",
	"err.store.pathrel":     "%q: a project path is relative and uses /",
	"err.store.pathctl":     "%q: a control character in a path",
	"err.store.pathleaves":  "%q: the path leaves the project",
	"err.store.filesize":    "%s is %d KiB: a file is at most %d KiB",
	"err.store.maxfiles":    "a project holds at most %d files",
	"err.store.projectsize": "the project would be %d MiB: at most %d MiB",
	"err.store.nostore":     "this browser keeps no storage for the page",
	"err.media.notfile":     "%q is not a file of the project",
	"err.media.textover":    "the project's text files are over %d KiB",
	"err.media.notmedium":   "%q is not a medium of the project",
	"err.media.mediaover":   "the project's media are over %d MiB",
	"err.tpl.mapping":       "%s: %s is a mapping; list %s in it by hand",
	"err.tpl.config":        "%s: %v",
	"err.tpl.notmapping":    "%s: the configuration is not a mapping",

	// what a diagnostic's code means (editor.Codes): said after the compiler's own English message
	"code.kind.unknown":      "The slide names a kind that is not defined.",
	"code.directive.unknown": "The directive is not one go-slides knows.",
	"code.template.function": "The template calls a function that does not exist.",
	"code.template.exec":     "The template failed while it was filled in.",
	"code.template.syntax":   "The template is not well formed ({{ … }} unbalanced or unexpected).",
	"code.yaml.syntax":       "The YAML is not well formed.",
	"code.hcl.syntax":        "The HCL is not well formed.",
	"code.hcl.unsupported":   "The HCL file uses something go-slides does not read.",
	"code.theme.unknown":     "The deck's theme is not one that exists.",
	"code.file.missing":      "A file the deck reads cannot be found.",

	// the toolkit's own words (toolkit.MessageKeys)
	"tk.filetree.new_file":    "New File…",
	"tk.filetree.new_folder":  "New Folder…",
	"tk.filetree.rename":      "Rename…",
	"tk.filetree.delete":      "Delete",
	"tk.filetree.cancel":      "Cancel",
	"tk.filetree.err.slash":   "A file or folder name cannot start with a slash.",
	"tk.filetree.err.invalid": "The name '%s' is not valid as a file or folder name. Please choose a different name.",
	"tk.filetree.err.exists":  "A file or folder '%s' already exists at this location. Please choose a different name.",
	"tk.filetree.warn.space":  "Leading or trailing whitespace detected in file or folder name.",
	"tk.find.replace":         "Replace",
	"tk.find.replace_all":     "All",
	"tk.find.bad_pattern":     "Bad pattern",
	"tk.find.no_results":      "No results",
	"tk.find.count":           "%d of %d",
	"tk.find.a11y":            "Find and replace",
	"tk.ribbon.a11y":          "Ribbon",
	"tk.ribbon.a11y_tabs":     "Ribbon tabs",
	"tk.filechooser.open":     "Open",
	"tk.filechooser.cancel":   "Cancel",
	"tk.dialog.ok":            "OK",

	"fmt.label.bold": "Bold", "fmt.label.italic": "Italic", "fmt.label.underline": "Underline",
	"fmt.label.strike": "Strikethrough", "fmt.label.code": "Inline code",
	"fmt.label.h1": "Heading 1", "fmt.label.h2": "Heading 2", "fmt.label.h3": "Heading 3",
	"fmt.label.bullets": "Bullets", "fmt.label.numbered": "Numbering", "fmt.label.quote": "Quote", "fmt.label.rule": "Rule",
	"fmt.label.table": "Table", "fmt.label.table-format": "Tidy",
	"fmt.label.row-above": "Row above", "fmt.label.row-below": "Row below", "fmt.label.row-delete": "Delete row",
	"fmt.label.col-left": "Column left", "fmt.label.col-right": "Column right", "fmt.label.col-delete": "Delete column",
	"fmt.label.colalign-left": "Column left-aligned", "fmt.label.colalign-center": "Column centred",
	"fmt.label.colalign-right": "Column right-aligned",
	"fmt.label.slide-left":     "Slide text left", "fmt.label.slide-center": "Slide text centre", "fmt.label.slide-right": "Slide text right",
	"fmt.label.link": "Link",

	"fmt.tip.bold":            "Make the selected text, or the word at the caret, bold: **text**. Press again to undo.",
	"fmt.tip.italic":          "Italic: _text_. Press again to undo.",
	"fmt.tip.underline":       "Markdown has no underline: this wraps the text in <u>…</u>, a tag Marp's default HTML allow-list keeps, so it works with raw HTML off. Press again to undo.",
	"fmt.tip.strike":          "Strikethrough: ~~text~~. Press again to undo.",
	"fmt.tip.code":            "Inline code: `text`. Press again to undo.",
	"fmt.tip.h1":              "Make the line (or each selected line) a heading 1: # text. Press again for plain text.",
	"fmt.tip.h2":              "Heading 2: ## text.",
	"fmt.tip.h3":              "Heading 3: ### text.",
	"fmt.tip.bullets":         "Bulleted list: - item, one per selected line. Press again to undo.",
	"fmt.tip.numbered":        "Numbered list: 1. item, 2. item… Press again to undo.",
	"fmt.tip.quote":           "Quote: > text. Press again to take one level off. Ctrl+Shift+.",
	"fmt.tip.rule":            "A horizontal rule. Marp reads ---, *** and ___ as a new slide, so this writes the <hr> tag, which the default HTML allow-list keeps.",
	"fmt.tip.table":           "Insert a table after the current line: choose its columns and rows.",
	"fmt.tip.table-format":    "Align the table around the caret: every column padded to its widest cell, the alignment marks kept.",
	"fmt.tip.row-above":       "Add an empty row above the caret's row (under the header from the header).",
	"fmt.tip.row-below":       "Add an empty row below the caret's row.",
	"fmt.tip.row-delete":      "Remove the caret's row; the header stays.",
	"fmt.tip.col-left":        "Add a column left of the caret's column.",
	"fmt.tip.col-right":       "Add a column right of the caret's column.",
	"fmt.tip.col-delete":      "Remove the caret's column; a table keeps one.",
	"fmt.tip.colalign-left":   "Align the caret's table column left (:--). Press again to clear it.",
	"fmt.tip.colalign-center": "Centre the caret's table column (:-:). Press again to clear it.",
	"fmt.tip.colalign-right":  "Align the caret's table column right (--:). Press again to clear it.",
	"fmt.tip.slide-left":      "Put the slide's text back at the left, the theme's own: removes the style block this tab wrote.",
	"fmt.tip.slide-center":    "Centre the text of the slide the caret is in, with a <style scoped> block at the top of the slide. It works with raw HTML off and shows in the preview, the HTML page and the PDF.",
	"fmt.tip.slide-right":     "Right-align the text of the slide the caret is in (a <style scoped> block).",
	"fmt.tip.link":            "Make the selection, or the word, a link: [text](url) with the url selected for typing. In a link, select its url.",

	KGrpHistory: "History", KLblUndo: "Undo", KLblRedo: "Redo",
	KTipUndo:     "Take back the last change: an edit in the file shown, else an operation on the whole project.",
	KTipUndoStep: "Undo: %s.",
	KTipRedo:     "Do again what Undo took back.",
	KTipRedoStep: "Redo: %s.",
	KStepExample: "the example “%s”", KStepReset: "Reset", KStepOpenZip: "Open a .zip", KStepOpenFolder: "Open a folder",
	KStepRename: "Rename %s", KStepDelete: "Delete %s", KStepRestoreDraft: "Restore the draft", KStepRestoreVersion: "Restore %s",
	KStatusUndone: "Undone: %s.", KStatusRedone: "Redone: %s.",

	KLblSaveVersion: "Save version…", KLblVersions: "Versions",
	KTipSaveVersion: "Keep the whole project as it is now, under a name, in this browser.",
	KTipVersions:    "The project's versions, kept in this browser: preview one, compare it with now, restore it.",
	KAskVersionName: "Name this version:",
	KVersionAuto:    "Automatic", KVersionBefore: "Before: %s", KVersionExport: "Exported: %s", KVersionShared: "Shared link",
	KVersionLine:   "%s · %s · %s",
	KVersionsEmpty: "No version yet. One is kept a little after you stop typing, before an operation on the whole project, at each export or share, and when you save one.",
	KVersionsOff:   "Versions are not kept: this browser keeps no storage for the page.",
	KLblOnlyNamed:  "Only named", KLblVersionPreview: "Preview", KLblVersionCompare: "Compare with now",
	KLblVersionRestore: "Restore", KLblVersionDelete: "Delete", KLblBackToNow: "Back to now", KLblSideBySide: "Side by side",
	KCompareTitle:    "%s compared with now",
	KCompareFileLine: "%s (+%d −%d)",
	KCompareSame:     "No difference with now.",
	KCompareBinary:   "%s is not text: %s then, %s now.",
	KCompareNow:      "Now",

	KStatusVersionSaved:    "Saved the version “%s”.",
	KStatusVersionRestored: "Restored %s; Undo takes it back.",
	KStatusVersionDeleted:  "Deleted the version %s.",
	KStatusPreviewing:      "The preview shows %s; Back to now, or an edit, shows the deck again.",
	KStatusVersionsFull:    "This page's storage is full: the oldest automatic versions were dropped.",
	KStatusVersionsFailed:  "A version could not be kept: %s",
	KStatusVersionLimit:    "At most %d named versions are kept: delete one to save another.",
	KConfirmDeleteVersion:  "Delete the version %s? This cannot be undone.",

	KHelpGeneral: `Type a deck on the left; the slides follow on the right.

- Slides are separated by a line of three dashes: ---
- The front matter, between the first two --- lines, holds the global directives (theme, paginate, header, footer, size) and go-slides' own keys: kinds, data, template.
- <!-- _class: lead --> in a slide is a directive for that slide alone.
- {{ .title }}, {{ page }}/{{ total }}, {{ range seq 3 }}: Go templates.
- kind: title applies a kind, a named set of directives.
- Math: $e^{i\pi}$ inline, $$ ... $$ on its own lines.
- Images: ![w:300](url) sizes one, ![sepia](url) filters it; ![bg](url) is a background, ![bg left:40%](url) a split one beside the text.

Files, left of the editor, is the deck's project, as the slides command reads it: deck.md, slides.yaml, themes/, kinds/, media/. Click a file to edit it. Right-click (or Shift+F10) for New File, New Folder, Rename (F2), Delete, New theme, New kind; drag a file onto a folder to move it. The Project menu makes a theme built on default, gaia or uncover (it joins the Theme list), a kind, a data file, and opens or downloads the whole project as a .zip (unzipped, slides build deck.md builds it).

Panes: drag the bar between two panes to resize them (Files, the editor, Slides, the preview, Help); double-click a bar, or press Enter on it, to pull the pane shut and bring it back; F6 moves between the bars, the arrow keys move one by 16 px, Home and End go to the smallest and the largest, Escape returns to the editor. The sizes you set are kept in this browser; the View tab still switches panes off.

Format tab: Bold, Italic, Strikethrough and Inline code work on the selection or the word at the caret and toggle (Ctrl+B, Ctrl+I, Ctrl+Shift+X, Ctrl+E; Cmd on a Mac): each button shows pressed inside its marks, and one Undo takes one action back. Underline (Ctrl+U) wraps the text in <u>…</u>, because Markdown has no underline and Marp's default HTML list keeps that tag. Headings 1 to 3 (Ctrl+Alt+1..3), bullets (Ctrl+Shift+8), numbering (Ctrl+Shift+7) and quote work on the selected lines; the rule is an <hr> tag, because ---, *** and ___ would start a new slide. Table opens a grid of columns by rows; at the caret in a table you add or remove a row or a column, align the column (:--, :-:, --:) and Tidy keeps the cells lined up. Slide text left, centre or right writes a <style scoped> block into the current slide: it works with raw HTML off and shows in the preview, the HTML page and the PDF. Link (Ctrl+K) writes [text](url) with the url selected.

Drop or paste an image or a video on the editor: it goes into media/ and the deck names it, ![](media/photo.png); drop files on a folder of the panel to put them there. A file is at most 8 MiB and a project 64 MiB (the line under the files says how much is used). Nothing is uploaded. An https: image works too, but your browser fetches it from its server, which then sees your address (no referrer is sent).

Theme picks default, gaia, uncover or a theme your deck defines (/* @theme name */ in a <style>): it writes theme: into the front matter, and follows the front matter when you edit it.

Completion opens as you type a directive, a kind, or inside {{ }}. Put the caret on a directive, a kind or a template function and this pane explains it. Click a slide in the overview to go to its line.

Present (or p in the slideshow) opens the presenter view in a second window: the current slide, the next step, the speaker notes (<!-- comments -->), a clock and a timer. Arrow keys, Space, Home and End move both windows; the slideshow and the presenter follow each other, also across edits. Allow pop-ups for this site if the window does not open.

Your deck stays in this browser: it is saved as you type in this browser's own storage, for this site only, and comes back when you reload; nothing is sent anywhere. Draft restores the draft a shared link, an example or another tab replaced, resets to the example, and Forget my drafts erases them. Share link puts the deck in the link itself (not its other files: download the project's .zip for those).

Undo and Redo (Home tab; Ctrl+Z, and Ctrl+Shift+Z or Ctrl+Y; Cmd on a Mac) take back the last edit of the file shown, each file keeping its own, and then the last operation on the whole project: an example, Reset, a .zip or a folder opened, a draft restored, a file deleted or renamed, a version restored. Redo brings back everything as it was when you pressed Undo.

Versions keep the whole project in this browser: one is saved a little after you stop typing, before an operation on the whole project, at each export and share, and Save version… keeps one under a name. The Versions pane lists them: Preview shows one in the preview, Compare with now shows each file's changes line by line (F7 goes to the next change), Restore puts it back, and Undo takes the restore back. The newest automatic versions are kept, then one an hour for a day and one a day for a month; named ones stay until you delete them.

The ribbon's tabs: Home (Undo, Redo, versions, example, theme, Present, Share link, Draft), Insert (a new slide, an image or background from a file, math, a code block with line numbers, a kind: each is typed into the deck at the caret, around the selection if there is one, and Undo takes it back), View (the preview, the minimap, the slide thumbnails) and Export. Alt shows a letter over each tab; type it, then the letter over a control. Ctrl+F1 collapses the ribbon.

Export tab: the HTML page of the presentation, its media inside; a PDF (vector, its text selectable), PowerPoint or OpenDocument deck (editable), made in your browser exactly as the slides command makes them. The first export downloads the export engine (about 5 MiB, once: your browser keeps it), compiling it as it arrives, and then only the fonts your slides use; https: images are fetched only if their server allows it (CORS), data: images always.`,
}

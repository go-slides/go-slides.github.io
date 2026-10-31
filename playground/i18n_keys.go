package playground

// The messages of the playground, by key. A key is a typed constant, so a
// view cannot name a message that does not exist; i18n_test.go checks that
// every key is in every language with the same placeholders.
//
// Counted messages (K...Rendered, Problems, Saved, FilesUsage) are bases: the
// catalogue holds base+".one", ".few", ".many" and ".other" as the language's
// CLDR plural rules ask (Tn).
const (
	// the ribbon's tabs
	KTabHome   Key = "tab.home"
	KTabInsert Key = "tab.insert"
	KTabFormat Key = "tab.format"
	KTabView   Key = "tab.view"
	KTabExport Key = "tab.export"

	// the ribbon's groups
	KGrpDeck      Key = "grp.deck"
	KGrpLook      Key = "grp.look"
	KGrpPresent   Key = "grp.present"
	KGrpDraft     Key = "grp.draft"
	KGrpSlides    Key = "grp.slides"
	KGrpPictures  Key = "grp.pictures"
	KGrpText      Key = "grp.text"
	KGrpPreview   Key = "grp.preview"
	KGrpEditor    Key = "grp.editor"
	KGrpSaveAs    Key = "grp.saveas"
	KGrpParagraph Key = "grp.paragraph"
	KGrpTable     Key = "grp.table"
	KGrpRowsCols  Key = "grp.rowscols"
	KGrpColumn    Key = "grp.column"
	KGrpSlideText Key = "grp.slidetext"
	KGrpLink      Key = "grp.link"

	// the controls' labels
	KLblExample     Key = "lbl.example"
	KLblReset       Key = "lbl.reset"
	KLblTheme       Key = "lbl.theme"
	KLblThemeOption Key = "lbl.themeoption"
	KLblDraft       Key = "lbl.draft"
	KLblShare       Key = "lbl.share"
	KLblPresent     Key = "lbl.present"
	KLblPreview     Key = "lbl.preview"
	KLblMinimap     Key = "lbl.minimap"
	KLblSlides      Key = "lbl.slides"
	KLblFiles       Key = "lbl.files"
	KLblHelp        Key = "lbl.help"
	KLblNewSlide    Key = "lbl.newslide"
	KLblImage       Key = "lbl.image"
	KLblBackground  Key = "lbl.background"
	KLblBgLeft      Key = "lbl.bgleft"
	KLblBgRight     Key = "lbl.bgright"
	KLblMath        Key = "lbl.math"
	KLblCode        Key = "lbl.code"
	KLblKind        Key = "lbl.kind"
	KLblHTML        Key = "lbl.html"
	KLblPDF         Key = "lbl.pdf"
	KLblPPTX        Key = "lbl.pptx"
	KLblODP         Key = "lbl.odp"

	// the controls' tooltips
	KTipExample    Key = "tip.example"
	KTipTheme      Key = "tip.theme"
	KTipDraft      Key = "tip.draft"
	KTipReset      Key = "tip.reset"
	KTipPresent    Key = "tip.present"
	KTipShare      Key = "tip.share"
	KTipKind       Key = "tip.kind"
	KTipNewSlide   Key = "tip.newslide"
	KTipImage      Key = "tip.image"
	KTipBackground Key = "tip.background"
	KTipMath       Key = "tip.math"
	KTipCode       Key = "tip.code"
	KTipMode       Key = "tip.mode"
	KTipMinimap    Key = "tip.minimap"
	KTipThumbs     Key = "tip.thumbs"
	KTipFiles      Key = "tip.files"
	KTipHelp       Key = "tip.help"
	KTipHTML       Key = "tip.html"
	KTipPDF        Key = "tip.pdf"
	KTipPPTX       Key = "tip.pptx"
	KTipODP        Key = "tip.odp"

	// the preview's modes
	KModeOverview  Key = "mode.overview"
	KModeSlideshow Key = "mode.slideshow"

	// the examples' titles (the decks stay in English)
	KExWelcome           Key = "ex.welcome"
	KExTemplates         Key = "ex.templates"
	KExKinds             Key = "ex.kinds"
	KExHCL               Key = "ex.hcl"
	KExGaia              Key = "ex.gaia"
	KExMedia             Key = "ex.media"
	KExGalleryDefault    Key = "ex.gallery.default"
	KExGalleryGaia       Key = "ex.gallery.gaia"
	KExGalleryUncover    Key = "ex.gallery.uncover"
	KExGalleryMath       Key = "ex.gallery.math"
	KExGalleryCode       Key = "ex.gallery.code"
	KExGalleryEmoji      Key = "ex.gallery.emoji"
	KExGalleryPresenting Key = "ex.gallery.presenting"
	KExGalleryHCL        Key = "ex.gallery.hcl"

	// the status line
	KStatusStarting           Key = "status.starting"
	KStatusOpenedShared       Key = "status.openedshared"
	KStatusSharedUnreadable   Key = "status.sharedunreadable"
	KStatusRendered           Key = "status.rendered"
	KStatusStale              Key = "status.stale"
	KStatusProblems           Key = "status.problems"
	KStatusDraftNotKept       Key = "status.draftnotkept"
	KStatusPresenting         Key = "status.presenting"
	KStatusLinkCopied         Key = "status.linkcopied"
	KStatusSavedHTML          Key = "status.savedhtml"
	KStatusShareDeckOnly      Key = "status.sharedeckonly"
	KStatusRestoredDraft      Key = "status.restoreddraft"
	KStatusSharedKeptDraft    Key = "status.sharedkeptdraft"
	KStatusRestoredDraftReset Key = "status.restoreddraftreset"
	KStatusDraftSavedAgain    Key = "status.draftsavedagain"
	KStatusDraftNotSaved      Key = "status.draftnotsaved"
	KStatusOtherTab           Key = "status.othertab"
	KStatusRestoredTheDraft   Key = "status.restoredthedraft"
	KStatusForgot             Key = "status.forgot"
	KStatusNoExport           Key = "status.noexport"
	KStatusExportRunning      Key = "status.exportrunning"
	KStatusTooLargeExport     Key = "status.toolargeexport"
	KStatusPreparing          Key = "status.preparing"
	KStatusLoadingEnginePct   Key = "status.loadingenginepct"
	KStatusLoadingEngine      Key = "status.loadingengine"
	KStatusExportProgress     Key = "status.exportprogress"
	KStatusExportFailed       Key = "status.exportfailed"
	KStatusSaved              Key = "status.saved"
	KStatusIsMedium           Key = "status.ismedium"
	KStatusMadeTheme          Key = "status.madetheme"
	KStatusMadeKind           Key = "status.madekind"
	KStatusMadeData           Key = "status.madedata"
	KStatusCannotSaveFile     Key = "status.cannotsavefile"
	KStatusSavedZip           Key = "status.savedzip"
	KStatusOpenedZip          Key = "status.openedzip"
	KStatusOpenedFolder       Key = "status.openedfolder"
	KStatusAddedSlide         Key = "status.addedslide"
	KStatusNoFilePicker       Key = "status.nofilepicker"
	KStatusKindSet            Key = "status.kindset"
	KStatusInsertedRule       Key = "status.insertedrule"
	KStatusInsertedTable      Key = "status.insertedtable"
	KStatusJournalLarge       Key = "status.journallarge"
	KStatusJournalRefused     Key = "status.journalrefused"
	KStatusAdded              Key = "status.added"
	KStatusAddedPut           Key = "status.addedput"
	KStatusPopupBlocked       Key = "status.popupblocked"

	// what an Insert command said it wrote
	KStatusInsertedFormula Key = "status.insertedformula"
	KStatusInsertedCode    Key = "status.insertedcode"
	KStatusPutFormula      Key = "status.putformula"
	KStatusPutCode         Key = "status.putcode"

	// why a Format command did nothing
	KWhyHeading    Key = "why.heading"
	KWhyList       Key = "why.list"
	KWhyQuote      Key = "why.quote"
	KWhyNotInTable Key = "why.notintable"
	KWhyNotInBody  Key = "why.notinbody"
	KWhyLastColumn Key = "why.lastcolumn"

	// the problems list
	KProblemLine    Key = "problem.line"
	KProblemError   Key = "problem.error"
	KProblemWarning Key = "problem.warning"

	// confirmations
	KConfirmReset      Key = "confirm.reset"
	KConfirmShareFiles Key = "confirm.sharefiles"
	KConfirmOpenZip    Key = "confirm.openzip"
	KConfirmOpenFolder Key = "confirm.openfolder"

	// dates and sizes
	KDateToday  Key = "date.today"
	KDateAt     Key = "date.at"
	KDateShort  Key = "date.short"
	KDateMonths Key = "date.months"
	KBytes      Key = "bytes"
	KWarn       Key = "warn"

	// the Draft menu
	KDraftRestoreMine     Key = "draft.restoremine"
	KDraftRestoreReplaced Key = "draft.restorereplaced"
	KDraftForget          Key = "draft.forget"
	KDraftResetExample    Key = "draft.resetexample"
	KDraftNothing         Key = "draft.nothing"

	// the Save as menu, the kind picker
	KSaveAs    Key = "saveas"
	KKindTitle Key = "kindtitle"

	// the file panel and the Project menu
	KProject             Key = "project"
	KFilesTreeLabel      Key = "files.treelabel"
	KFilesNewThemeFrom   Key = "files.newthemefrom"
	KFilesNewKind        Key = "files.newkind"
	KFilesNewData        Key = "files.newdata"
	KFilesNewFile        Key = "files.newfile"
	KFilesNewFolder      Key = "files.newfolder"
	KFilesAddMedia       Key = "files.addmedia"
	KFilesOpenZip        Key = "files.openzip"
	KFilesOpenFolder     Key = "files.openfolder"
	KFilesDownloadZip    Key = "files.downloadzip"
	KFilesInsertIntoDeck Key = "files.insertintodeck"
	KFilesDownload       Key = "files.download"
	KFilesUsage          Key = "files.usage"

	// the panes
	KPaneProblems      Key = "pane.problems"
	KPaneDocumentation Key = "pane.documentation"

	// the frames' accessible titles
	KFrameTitlePreview Key = "frame.preview"
	KFrameTitleRail    Key = "frame.rail"

	// the documentation links
	KDocGettingStarted Key = "doc.gettingstarted"
	KDocMarp           Key = "doc.marp"
	KDocTemplates      Key = "doc.templates"
	KDocKinds          Key = "doc.kinds"
	KDocHCL            Key = "doc.hcl"
	KDocCLI            Key = "doc.cli"
	KDocVSCode         Key = "doc.vscode"
	KDocPlayground     Key = "doc.playground"

	// the help pane
	KHelpGeneral Key = "help.general"

	// what a new table's header says: "Column 1"
	KTableColumn Key = "table.column"

	// errors
	KErrExists            Key = "err.exists"
	KErrDeckName          Key = "err.deckname"
	KErrNoSuchFile        Key = "err.nosuchfile"
	KErrNotFileName       Key = "err.notfilename"
	KErrDeckTooLarge      Key = "err.decktoolarge"
	KErrMaxFiles          Key = "err.maxfiles"
	KErrPathReason        Key = "err.pathreason"
	KErrProjectOver       Key = "err.projectover"
	KErrFileSize          Key = "err.filesize"
	KErrShareTooLargeDeck Key = "err.share.toolargedeck"
	KErrShareTooLargeLink Key = "err.share.toolargelink"
	KErrShareNoDeck       Key = "err.share.nodeck"
	KErrShareTooLong      Key = "err.share.toolong"
	KErrShareDamaged      Key = "err.share.damaged"
	KErrShareLarger       Key = "err.share.larger"
	KErrShareNotText      Key = "err.share.nottext"
	KErrZipNoDeck         Key = "err.zip.nodeck"
	KErrZipSize           Key = "err.zip.size"
	KErrZipNot            Key = "err.zip.not"
	KErrZipEntries        Key = "err.zip.entries"
	KErrZipLinks          Key = "err.zip.links"
	KErrZipFileOver       Key = "err.zip.fileover"
	KErrZipBomb           Key = "err.zip.bomb"
	KErrZipTwice          Key = "err.zip.twice"
	KErrZipMoreThan       Key = "err.zip.morethan"
	KErrFileOver          Key = "err.fileover"
	KErrFailed            Key = "err.failed"
	KErrFolderCount       Key = "err.foldercount"
	KErrFolderOver        Key = "err.folderover"
	KErrFolderMore        Key = "err.foldermore"
	KErrNotImage          Key = "err.notimage"
)

// counted lists the bases of the counted messages: each needs the forms of the
// language's plural rules (PluralForms).
var counted = []Key{KStatusRendered, KStatusProblems, KStatusSaved, KFilesUsage}

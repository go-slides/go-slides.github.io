package playground

// deMessages is the German catalogue.
var deMessages = map[Key]string{
	KTabHome: "Start", KTabInsert: "Einfügen", KTabFormat: "Format", KTabView: "Ansicht", KTabExport: "Export",

	KGrpDeck: "Dokument", KGrpLook: "Design", KGrpPresent: "Vortrag", KGrpDraft: "Entwurf", KGrpSlides: "Folien",
	KGrpPictures: "Bilder", KGrpText: "Text", KGrpPreview: "Vorschau", KGrpEditor: "Editor", KGrpSaveAs: "Speichern als",
	KGrpParagraph: "Absatz", KGrpTable: "Tabelle", KGrpRowsCols: "Zeilen, Spalten", KGrpColumn: "Spalte",
	KGrpSlideText: "Folientext", KGrpLink: "Link",

	KLblExample: "Beispiel", KLblReset: "Zurück", KLblTheme: "Design", KLblThemeOption: "Design: %s", KLblDraft: "Entwurf",
	KLblShare: "Link teilen", KLblPresent: "Vortrag", KLblPreview: "Vorschau", KLblMinimap: "Minimap",
	KLblSlides: "Folien", KLblFiles: "Dateien", KLblHelp: "Hilfe", KLblNewSlide: "Neue Folie", KLblImage: "Bild",
	KLblBackground: "Hintergrund", KLblBgLeft: "Hintergrund links", KLblBgRight: "Hintergrund rechts",
	KLblMath: "Formel", KLblCode: "Codeblock", KLblKind: "Kind", KLblHTML: "HTML-Seite", KLblPDF: "PDF",
	KLblPPTX: "PowerPoint", KLblODP: "OpenDocument",

	KTipExample:    "Mit einem der Beispiele beginnen. Dein Text bleibt in diesem Browser: Entwurf ▸ Ersetzten wiederherstellen holt ihn zurück.",
	KTipTheme:      "Das Design der Präsentation: Die Auswahl wird in den Front Matter der Präsentation geschrieben.",
	KTipDraft:      "Den Entwurf, den dieser Browser für dich speichert, wiederherstellen, vergessen oder zurücksetzen.",
	KTipReset:      "Den Text des Beispiels wiederherstellen; deine Präsentation bleibt als Entwurf erhalten.",
	KTipPresent:    "Die Bildschirmpräsentation hier zeigen und die Referentenansicht in einem eigenen Fenster öffnen.",
	KTipShare:      "Einen Link kopieren, der die Präsentation selbst enthält: Es wird nichts hochgeladen.",
	KTipKind:       "Der Folie ein Kind geben: eine Vorlage mit eigenem Aussehen, festgelegt durch einen Kommentar <!-- _kind: … -->.",
	KTipNewSlide:   "Eine leere Folie nach der aktuellen einfügen.",
	KTipImage:      "Eine Bilddatei auswählen: Sie wird als data:-URI in die Präsentation eingefügt, nichts wird hochgeladen. Markierter Text wird zur Beschreibung.",
	KTipBackground: "Ein Bild als Folienhintergrund auswählen; der Pfeil füllt nur die linken oder rechten 40 %.",
	KTipMath:       "Eine abgesetzte Formel zwischen $$-Zeilen; die Auswahl kommt hinein.",
	KTipCode:       "Ein umzäunter Codeblock mit Zeilennummern ({lineNumbers}); die Auswahl kommt hinein.",
	KTipMode:       "Die Übersicht zeigt alle Folien untereinander und scrollt mit dem Editor; die Bildschirmpräsentation ist der Vortrag selbst.",
	KTipMinimap:    "Eine Übersichtskarte der Präsentation neben der Bildlaufleiste des Editors.",
	KTipThumbs:     "Eine Spalte mit Folienminiaturen; ein Klick springt zur Folie.",
	KTipFiles:      "Die Dateien des Projekts: Designs, Kinds, Daten und Medien. Ein schmales Fenster startet ohne sie.",
	KTipHelp:       "Die Hilfe zur Stelle des Cursors und die Links zur Dokumentation. Ein schmales Fenster startet ohne sie.",
	KTipHTML:       "Eine HTML-Datei, die die Präsentation in jedem Browser zeigt.",
	KTipPDF:        "Ein PDF mit einer Seite pro Folie, hier in deinem Browser erstellt.",
	KTipPPTX:       "Eine bearbeitbare PowerPoint-Datei, hier in deinem Browser erstellt.",
	KTipODP:        "Eine bearbeitbare OpenDocument-Präsentation, hier in deinem Browser erstellt.",

	KModeOverview: "Übersicht", KModeSlideshow: "Bildschirmpräsentation",

	KExWelcome: "Willkommen", KExTemplates: "Vorlagen: erzeugte Folien", KExKinds: "Folien-Kinds",
	KExHCL: "HCL-Front-Matter", KExGaia: "Designs und Spot-Direktiven", KExMedia: "Formeln, Bilder und Hintergründe",
	KExGalleryDefault: "Galerie: das Standarddesign", KExGalleryGaia: "Galerie: das Design gaia",
	KExGalleryUncover: "Galerie: das Design uncover", KExGalleryMath: "Galerie: Formeln",
	KExGalleryCode:       "Galerie: Code, Zeilennummern und Hinweise",
	KExGalleryEmoji:      "Galerie: Emoji und angepasste Überschriften",
	KExGalleryPresenting: "Galerie: Übergänge und Sprechernotizen",
	KExGalleryHCL:        "Galerie: HCL-Ausdrücke und typisierte Daten",

	KStatusStarting:           "Wird gestartet …",
	KStatusOpenedShared:       "Die im Link geteilte Präsentation wurde geöffnet.",
	KStatusSharedUnreadable:   "Die Präsentation aus dem Link konnte nicht gelesen werden: %s",
	"status.rendered.one":     "%d Folie · in %d ms gerendert",
	"status.rendered.other":   "%d Folien · in %d ms gerendert",
	KStatusStale:              "die Vorschau zeigt die zuletzt gerenderte Präsentation",
	"status.problems.one":     "%d Problem · %s",
	"status.problems.other":   "%d Probleme · %s",
	KStatusDraftNotKept:       " · ⚠ Entwurf nicht in diesem Browser gespeichert",
	KStatusPresenting:         "Vortrag läuft: Die Referentenansicht ist in einem eigenen Fenster; die Pfeiltasten in beiden Fenstern steuern beide.",
	KStatusLinkCopied:         "Link kopiert: Die Präsentation steckt im Link selbst (%d Zeichen), nichts wird hochgeladen.",
	KStatusSavedHTML:          "deck.html gespeichert: In jedem Browser öffnen oder daraus präsentieren.",
	KStatusShareDeckOnly:      " Der Link enthält nur die Präsentation: Projekt ▸ Projekt herunterladen (.zip) sichert auch die Dateien.",
	KStatusRestoredDraft:      "Dein Entwurf von %s wurde wiederhergestellt.",
	KStatusSharedKeptDraft:    "Die im Link geteilte Präsentation wurde geöffnet. Dein Entwurf von %s bleibt erhalten: Entwurf ▸ Meinen Entwurf wiederherstellen.",
	KStatusRestoredDraftReset: "Dein Entwurf von %s wurde wiederhergestellt. Mit Zurück gelangst du zum Beispiel.",
	KStatusDraftSavedAgain:    "Dein Entwurf wird wieder in diesem Browser gespeichert.",
	KStatusDraftNotSaved:      "⚠ Dieser Browser hat deinen Entwurf nicht gespeichert (%s): Link teilen oder HTML herunterladen sichert die Präsentation.",
	KStatusOtherTab:           "Ein anderer Tab hat seinen Entwurf %s gespeichert. Deine Änderungen hier überschreiben ihn nicht: Entwurf ▸ Ersetzten wiederherstellen.",
	KStatusRestoredTheDraft:   "Der Entwurf von %s wurde wiederhergestellt.",
	KStatusForgot:             "Deine Entwürfe wurden verworfen: Von dieser Präsentation ist nichts mehr in diesem Browser gespeichert (bis zur nächsten Änderung).",
	KStatusNoExport:           "⚠ Diese Seite kann nicht exportieren: Verwende den Befehl slides.",
	KStatusExportRunning:      "Ein Export läuft: %s.",
	KStatusTooLargeExport:     "⚠ Die Präsentation ist zu groß für den Export hier: Verwende den Befehl slides.",
	KStatusPreparing:          "%s wird vorbereitet …",
	KStatusLoadingEnginePct:   "Export-Engine wird geladen (einmalig; dein Browser behält sie): %d %% von %.1f MiB",
	KStatusLoadingEngine:      "Export-Engine wird geladen (einmalig; dein Browser behält sie): %.1f MiB",
	KStatusExportProgress:     "%s: Folie %d von %d …",
	KStatusExportFailed:       "⚠ Der Export als %s ist fehlgeschlagen: %s",
	"status.saved.one":        "%s gespeichert: %s, %d Folie in %.1f s, in deinem Browser erstellt.",
	"status.saved.other":      "%s gespeichert: %s, %d Folien in %.1f s, in deinem Browser erstellt.",
	KStatusIsMedium:           "%s ist ein Medium: Die Präsentation zeigt es mit %s (auf den Editor ziehen oder über sein Menü: In die Präsentation einfügen).",
	KStatusMadeTheme:          "%s erstellt, basierend auf %s: Wähle %s in der Designliste, um es zu verwenden.",
	KStatusMadeKind:           "%s erstellt: Eine Folie übernimmt das Kind mit <!-- _kind: %s -->.",
	KStatusMadeData:           "%s erstellt: Die Präsentation liest es als {{ .title }}.",
	KStatusCannotSaveFile:     "⚠ diese Seite kann keine Datei speichern",
	KStatusSavedZip:           "%s gespeichert: Entpackt wird es mit `slides build deck.md` erstellt, und die VS-Code-Erweiterung öffnet es.",
	KStatusOpenedZip:          "Die .zip-Datei wurde geöffnet: %d Dateien.",
	KStatusOpenedFolder:       "Der Ordner wurde geöffnet: %d Dateien.",
	KStatusAddedSlide:         "Eine Folie nach der aktuellen eingefügt.",
	KStatusInsertedFormula:    "Eine abgesetzte Formel eingefügt: Tippe darin.",
	KStatusInsertedCode:       "Einen Go-Codeblock mit Zeilennummern eingefügt: Tippe darin.",
	KStatusPutFormula:         "Die Auswahl in eine abgesetzte Formel gesetzt.",
	KStatusPutCode:            "Die Auswahl in einen Go-Codeblock mit Zeilennummern gesetzt.",
	KStatusNoFilePicker:       "⚠ Diese Seite kann keine Datei abfragen: Ziehe eine Bilddatei in den Editor oder füge sie dort ein.",
	KStatusKindSet:            "Die Folie hat jetzt das Kind %s.",
	KStatusInsertedRule:       "Eine horizontale Linie eingefügt (***: drei Bindestriche in einer Zeile würden eine neue Folie beginnen).",
	KStatusInsertedTable:      "Eine Tabelle eingefügt: Tippe in ihre Zellen; die Befehle zum Ausrichten der Tabelle halten sie ordentlich.",
	KStatusJournalLarge:       "Dieses Projekt ist zu groß für die Schnellsicherung: Änderungen der letzten halben Sekunde vor dem Schließen der Seite gehen eventuell verloren.",
	KStatusJournalRefused:     "⚠ Die Schnellsicherung wurde abgelehnt (%s): Änderungen der letzten halben Sekunde vor dem Schließen der Seite gehen eventuell verloren.",
	KStatusAdded:              "%s zum Projekt hinzugefügt (%d KiB): Es bleibt in diesem Browser erhalten und kommt in die .zip-Datei des Projekts.",
	KStatusAddedPut:           "%s zum Projekt hinzugefügt (%d KiB) und in die Präsentation eingefügt: Es bleibt in diesem Browser erhalten und kommt in die .zip-Datei des Projekts.",
	KStatusPopupBlocked:       "⚠ Der Browser hat das Referentenfenster blockiert: Erlaube Pop-ups für diese Seite und starte den Vortrag erneut.",

	KWhyHeading:    "Eine Überschrift gehört in eine Zeile des Präsentationstexts, nicht in den Front Matter oder einen Codeblock.",
	KWhyList:       "Eine Liste gehört in eine Zeile des Präsentationstexts, nicht in den Front Matter oder einen Codeblock.",
	KWhyQuote:      "Ein Zitat gehört in eine Zeile des Präsentationstexts, nicht in den Front Matter oder einen Codeblock.",
	KWhyNotInTable: "Setze den Cursor zuerst in eine Tabelle.",
	KWhyNotInBody:  "Setze den Cursor in eine Zeile des Tabellenkörpers: Die Kopfzeile bleibt.",
	KWhyLastColumn: "Eine Tabelle behält mindestens eine Spalte.",

	KProblemLine: "Zeile %d", KProblemError: "Fehler", KProblemWarning: "Warnung",

	KConfirmReset:      "Deine Präsentation durch das Beispiel ersetzen? Dein Text bleibt in diesem Browser: Entwurf ▸ Ersetzten wiederherstellen.",
	KConfirmShareFiles: "Ein Link enthält nur die Präsentation, nicht die übrigen Dateien (Designs, Kinds, Medien): Stattdessen das Projekt als .zip herunterladen?",
	KConfirmOpenZip:    "Deine Präsentation und ihre Dateien durch die der .zip-Datei ersetzen? Der aktuelle Stand bleibt in diesem Browser: Entwurf ▸ Ersetzten wiederherstellen.",
	KConfirmOpenFolder: "Deine Präsentation und ihre Dateien durch die des Ordners ersetzen? Der aktuelle Stand bleibt in diesem Browser: Entwurf ▸ Ersetzten wiederherstellen.",

	KDateToday: "heute um %s", KDateAt: "%s um %s", KDateShort: "%[2]d. %[1]s",
	KDateMonths: "Jan,Feb,Mär,Apr,Mai,Jun,Jul,Aug,Sep,Okt,Nov,Dez",
	KBytes:      "%d Bytes", KWarn: "⚠ %s",

	KDraftRestoreMine: "Meinen Entwurf wiederherstellen (%s)", KDraftRestoreReplaced: "Ersetzten wiederherstellen (%s)",
	KDraftForget: "Meine Entwürfe verwerfen", KDraftResetExample: "Auf das Beispiel zurücksetzen", KDraftNothing: "Noch nichts gespeichert",

	KSaveAs: "Speichern unter …", KKindTitle: "Kind …",

	KProject: "Projekt", KFilesTreeLabel: "Projektdateien", KFilesNewThemeFrom: "Neues Design (von %s)",
	KFilesNewKind: "Neues Kind", KFilesNewData: "Neue Datendatei", KFilesNewFile: "Neue Datei …", KFilesNewFolder: "Neuer Ordner …",
	KFilesAddMedia: "Medien hinzufügen …", KFilesOpenZip: "Projekt öffnen (.zip) …", KFilesOpenFolder: "Projektordner öffnen …",
	KFilesDownloadZip: "Projekt herunterladen (.zip)", KFilesInsertIntoDeck: "In die Präsentation einfügen", KFilesDownload: "Herunterladen",
	"files.usage.one": "%d Datei · %s von %d MiB", "files.usage.other": "%d Dateien · %s von %d MiB",

	KPaneProblems: "Probleme", KPaneDocumentation: "Dokumentation",
	KFrameTitlePreview: "Folienvorschau", KFrameTitleRail: "Folienminiaturen",

	KDocGettingStarted: "Erste Schritte", KDocMarp: "Marp-Kompatibilität", KDocTemplates: "Vorlagen",
	KDocKinds: "Kinds", KDocHCL: "HCL", KDocCLI: "Befehlszeile", KDocVSCode: "VS Code", KDocPlayground: "Playground",

	KTableColumn: "Spalte %d",

	KErrExists: "%s existiert bereits", KErrDeckName: "die Präsentation bleibt deck.md: Der Befehl slides und der Playground lesen sie unter diesem Namen",
	KErrNoSuchFile: "%s: Datei nicht gefunden", KErrNotFileName: "%q ist kein Dateiname",
	KErrDeckTooLarge: "die Präsentation ist %d KiB groß; der Playground rendert Präsentationen bis %d KiB",
	KErrMaxFiles:     "ein Projekt enthält höchstens %d Dateien", KErrPathReason: "%s: %v",
	KErrProjectOver: "das Projekt würde %d MiB überschreiten", KErrFileSize: "%s ist %d KiB groß: Eine Datei darf höchstens %d KiB groß sein",
	KErrShareTooLargeDeck: "die Präsentation ist größer, als der Playground annimmt; lade sie stattdessen herunter",
	KErrShareTooLargeLink: "die Präsentation ist zu groß für einen Link; lade sie stattdessen herunter",
	KErrShareNoDeck:       "der Link enthält keine Präsentation", KErrShareTooLong: "die Präsentation im Link ist zu lang",
	KErrShareDamaged: "die Präsentation im Link ist beschädigt (möglicherweise abgeschnitten)",
	KErrShareLarger:  "die Präsentation im Link ist größer, als der Playground annimmt", KErrShareNotText: "die Präsentation im Link ist kein Text",
	KErrZipNoDeck: "das Archiv enthält keine Präsentation: Lege deck.md auf oberster Ebene ab (oder eine .md-Datei)",
	KErrZipSize:   "das Archiv ist %d MiB groß: höchstens %d MiB", KErrZipNot: "kein .zip-Archiv: %v",
	KErrZipEntries: "das Archiv enthält %d Einträge: höchstens %d", KErrZipLinks: "%s: Links werden nicht übernommen",
	KErrZipFileOver: "%s ist größer als %d KiB: Mehr ist für eine Datei nicht erlaubt",
	KErrZipBomb:     "das Archiv wird um mehr als das %d-Fache größer: als mögliche Zip-Bombe abgelehnt",
	KErrZipTwice:    "%s ist zweimal im Archiv enthalten", KErrZipMoreThan: "mehr als %d KiB",
	KErrFileOver: "%s ist größer als %s", KErrFailed: "fehlgeschlagen",
	KErrFolderCount: "der Ordner enthält %d Dateien: ein Projekt enthält höchstens %d", KErrFolderOver: "der Ordner ist größer als %s",
	KErrFolderMore: "der Ordner enthält mehr als %d Dateien", KErrNotImage: "%s ist kein Bild.",

	"fmt.label.bold": "Fett", "fmt.label.italic": "Kursiv", "fmt.label.underline": "Unterstr.",
	"fmt.label.strike": "Durchgestr.", "fmt.label.code": "Code",
	"fmt.label.h1": "Titel 1", "fmt.label.h2": "Titel 2", "fmt.label.h3": "Titel 3",
	"fmt.label.bullets": "Liste", "fmt.label.numbered": "Nummern", "fmt.label.quote": "Zitat", "fmt.label.rule": "Linie",
	"fmt.label.table": "Tabelle", "fmt.label.table-format": "Ordnen",
	"fmt.label.row-above": "Zeile oben", "fmt.label.row-below": "Zeile unten", "fmt.label.row-delete": "Zeile löschen",
	"fmt.label.col-left": "Spalte links", "fmt.label.col-right": "Spalte rechts", "fmt.label.col-delete": "Spalte löschen",
	"fmt.label.colalign-left": "Spalte links", "fmt.label.colalign-center": "Spalte zentr.",
	"fmt.label.colalign-right": "Spalte rechts",
	"fmt.label.slide-left":     "Text links", "fmt.label.slide-center": "Text zentr.", "fmt.label.slide-right": "Text rechts",
	"fmt.label.link": "Link",

	"fmt.tip.bold":            "Den markierten Text oder das Wort am Cursor fett setzen: **Text**. Erneut drücken zum Rückgängigmachen.",
	"fmt.tip.italic":          "Kursiv: _Text_. Erneut drücken zum Rückgängigmachen.",
	"fmt.tip.underline":       "Markdown kennt kein Unterstreichen: Dies umgibt den Text mit <u>…</u>, einem Tag, den die Standard-HTML-Zulassungsliste von Marp behält, es funktioniert also auch mit deaktiviertem rohem HTML. Erneut drücken zum Rückgängigmachen.",
	"fmt.tip.strike":          "Durchgestrichen: ~~Text~~. Erneut drücken zum Rückgängigmachen.",
	"fmt.tip.code":            "Inline-Code: `Text`. Erneut drücken zum Rückgängigmachen.",
	"fmt.tip.h1":              "Die Zeile (oder jede markierte Zeile) zur Überschrift 1 machen: # Text. Erneut drücken für normalen Text.",
	"fmt.tip.h2":              "Überschrift 2: ## Text.",
	"fmt.tip.h3":              "Überschrift 3: ### Text.",
	"fmt.tip.bullets":         "Aufzählung: - Eintrag, einer pro markierter Zeile. Erneut drücken zum Rückgängigmachen.",
	"fmt.tip.numbered":        "Nummerierte Liste: 1. Eintrag, 2. Eintrag … Erneut drücken zum Rückgängigmachen.",
	"fmt.tip.quote":           "Zitat: > Text. Erneut drücken, um eine Ebene zu entfernen. Ctrl+Shift+.",
	"fmt.tip.rule":            "Eine horizontale Linie. Marp liest ---, *** und ___ als neue Folie; daher wird der Tag <hr> geschrieben, den die Standard-HTML-Zulassungsliste behält.",
	"fmt.tip.table":           "Eine Tabelle nach der aktuellen Zeile einfügen: Wähle Spalten und Zeilen.",
	"fmt.tip.table-format":    "Die Tabelle am Cursor ausrichten: Jede Spalte wird auf ihre breiteste Zelle aufgefüllt, die Ausrichtungsmarken bleiben erhalten.",
	"fmt.tip.row-above":       "Eine leere Zeile über der Zeile am Cursor einfügen (in der Kopfzeile: unter der Kopfzeile).",
	"fmt.tip.row-below":       "Eine leere Zeile unter der Zeile am Cursor einfügen.",
	"fmt.tip.row-delete":      "Die Zeile am Cursor entfernen; die Kopfzeile bleibt.",
	"fmt.tip.col-left":        "Eine Spalte links von der Spalte am Cursor einfügen.",
	"fmt.tip.col-right":       "Eine Spalte rechts von der Spalte am Cursor einfügen.",
	"fmt.tip.col-delete":      "Die Spalte am Cursor entfernen; eine Spalte bleibt immer.",
	"fmt.tip.colalign-left":   "Die Tabellenspalte am Cursor linksbündig ausrichten (:--). Erneut drücken, um die Ausrichtung zu entfernen.",
	"fmt.tip.colalign-center": "Die Tabellenspalte am Cursor zentrieren (:-:). Erneut drücken, um die Ausrichtung zu entfernen.",
	"fmt.tip.colalign-right":  "Die Tabellenspalte am Cursor rechtsbündig ausrichten (--:). Erneut drücken, um die Ausrichtung zu entfernen.",
	"fmt.tip.slide-left":      "Den Folientext wieder linksbündig setzen, wie im Design vorgesehen: entfernt den Style-Block, den dieser Tab geschrieben hat.",
	"fmt.tip.slide-center":    "Den Text der Folie am Cursor zentrieren, mit einem <style scoped>-Block am Anfang der Folie. Es funktioniert mit deaktiviertem rohem HTML und erscheint in der Vorschau, der HTML-Seite und im PDF.",
	"fmt.tip.slide-right":     "Den Text der Folie am Cursor rechtsbündig ausrichten (ein <style scoped>-Block).",
	"fmt.tip.link":            "Aus der Auswahl oder dem Wort einen Link machen: [Text](url), die url ist zum Überschreiben markiert. In einem Link wird seine url markiert.",

	KGrpHistory: "Verlauf", KLblUndo: "Rückgängig", KLblRedo: "Wiederholen",
	KTipUndo:     "Nimmt die letzte Änderung zurück: eine Bearbeitung in der angezeigten Datei, sonst einen Vorgang am ganzen Projekt.",
	KTipUndoStep: "Rückgängig: %s.",
	KTipRedo:     "Macht noch einmal, was Rückgängig zurückgenommen hat.",
	KTipRedoStep: "Wiederholen: %s.",
	KStepExample: "das Beispiel „%s“", KStepReset: "Zurück", KStepOpenZip: "Eine .zip-Datei öffnen", KStepOpenFolder: "Einen Ordner öffnen",
	KStepRename: "%s umbenennen", KStepDelete: "%s löschen", KStepRestoreDraft: "Den Entwurf wiederherstellen", KStepRestoreVersion: "%s wiederherstellen",
	KStatusUndone: "Rückgängig gemacht: %s.", KStatusRedone: "Wiederholt: %s.",

	KLblSaveVersion: "Version speichern…", KLblVersions: "Versionen",
	KTipSaveVersion: "Das ganze Projekt, wie es jetzt ist, unter einem Namen in diesem Browser aufbewahren.",
	KTipVersions:    "Die Versionen des Projekts, in diesem Browser aufbewahrt: eine ansehen, mit dem Jetzt vergleichen, wiederherstellen.",
	KAskVersionName: "Name dieser Version:",
	KVersionAuto:    "Automatisch", KVersionBefore: "Vor: %s", KVersionExport: "Exportiert: %s", KVersionShared: "Geteilter Link",
	KVersionLine:   "%s · %s · %s",
	KVersionsEmpty: "Noch keine Version. Eine wird kurz nach dem Tippen aufbewahrt, vor einem Vorgang am ganzen Projekt, bei jedem Export oder Teilen und wenn du eine speicherst.",
	KVersionsOff:   "Versionen werden nicht aufbewahrt: Dieser Browser hat keinen Speicher für die Seite.",
	KLblOnlyNamed:  "Nur benannte", KLblVersionPreview: "Vorschau", KLblVersionCompare: "Mit jetzt vergleichen",
	KLblVersionRestore: "Wiederherstellen", KLblVersionDelete: "Löschen", KLblBackToNow: "Zurück zu jetzt", KLblSideBySide: "Nebeneinander",
	KCompareTitle:    "%s im Vergleich mit jetzt",
	KCompareFileLine: "%s (+%d −%d)",
	KCompareSame:     "Kein Unterschied zu jetzt.",
	KCompareBinary:   "%s ist kein Text: damals %s, jetzt %s.",
	KCompareNow:      "Jetzt",

	KStatusVersionSaved:    "Version „%s“ gespeichert.",
	KStatusVersionRestored: "%s wiederhergestellt; Rückgängig nimmt es zurück.",
	KStatusVersionDeleted:  "Version %s gelöscht.",
	KStatusPreviewing:      "Die Vorschau zeigt %s; Zurück zu jetzt oder eine Bearbeitung zeigt wieder die Präsentation.",
	KStatusVersionsFull:    "Der Speicher dieser Seite ist voll: Die ältesten automatischen Versionen wurden entfernt.",
	KStatusVersionsFailed:  "Eine Version konnte nicht aufbewahrt werden: %s",
	KStatusVersionLimit:    "Es werden höchstens %d benannte Versionen aufbewahrt: Lösche eine, um eine weitere zu speichern.",
	KConfirmDeleteVersion:  "Die Version %s löschen? Das lässt sich nicht rückgängig machen.",

	KHelpGeneral: `Tippe links eine Präsentation; die Folien erscheinen rechts.

- Folien werden durch eine Zeile mit drei Bindestrichen getrennt: ---
- Der Front Matter zwischen den ersten beiden ---Zeilen enthält die globalen Direktiven (theme, paginate, header, footer, size) und die eigenen Schlüssel von go-slides: kinds, data, template.
- <!-- _class: lead --> in einer Folie ist eine Direktive nur für diese Folie.
- {{ .title }}, {{ page }}/{{ total }}, {{ range seq 3 }}: Go-Templates.
- kind: title wendet ein Kind an (einen benannten Satz von Direktiven).
- Formeln: $e^{i\pi}$ im Text, $$ ... $$ in eigenen Zeilen.
- Bilder: ![w:300](url) legt die Größe fest, ![sepia](url) wendet einen Filter an; ![bg](url) ist ein Hintergrund, ![bg left:40%](url) ein geteilter neben dem Text.

Dateien, links vom Editor, ist das Projekt der Präsentation, so wie der Befehl slides es liest: deck.md, slides.yaml, themes/, kinds/, media/. Klicke auf eine Datei, um sie zu bearbeiten. Rechtsklick (oder Shift+F10) für Neue Datei, Neuer Ordner, Umbenennen (F2), Löschen, Neues Design, Neues Kind; ziehe eine Datei auf einen Ordner, um sie zu verschieben. Das Menü Projekt erstellt ein Design auf Basis von default, gaia oder uncover (es erscheint in der Designliste), ein Kind, eine Datendatei und öffnet oder lädt das ganze Projekt als .zip herunter (entpackt wird es mit slides build deck.md erstellt).

Bereiche: Ziehe die Leiste zwischen zwei Bereichen, um sie in der Größe zu ändern (Dateien, Editor, Folien, Vorschau, Hilfe); doppelklicke auf eine Leiste oder drücke darauf die Eingabetaste, um den Bereich zu schließen und wieder zu öffnen; F6 wechselt zwischen den Leisten, die Pfeiltasten verschieben eine um 16 px, Pos1 und Ende gehen zur kleinsten und zur größten Größe, Escape kehrt zum Editor zurück. Die eingestellten Größen bleiben in diesem Browser erhalten; im Tab Ansicht lassen sich Bereiche weiterhin ausblenden.

Tab Format: Fett, Kursiv, Durchgestrichen und Inline-Code wirken auf die Auswahl oder das Wort am Cursor und schalten um (Ctrl+B, Ctrl+I, Ctrl+Shift+X, Ctrl+E; Cmd auf dem Mac): Jede Schaltfläche erscheint innerhalb ihrer Markierungen gedrückt, und ein Rückgängig nimmt eine Aktion zurück. Unterstrichen (Ctrl+U) umgibt den Text mit <u>…</u>, weil Markdown kein Unterstreichen kennt und die Standard-HTML-Liste von Marp diesen Tag behält. Überschriften 1 bis 3 (Ctrl+Alt+1..3), Aufzählung (Ctrl+Shift+8), Nummerierung (Ctrl+Shift+7) und Zitat wirken auf die markierten Zeilen; die Linie ist ein <hr>-Tag, weil ---, *** und ___ eine neue Folie beginnen würden. Tabelle öffnet ein Raster aus Spalten und Zeilen; am Cursor in einer Tabelle fügst du eine Zeile oder Spalte hinzu oder entfernst sie, richtest die Spalte aus (:--, :-:, --:) und Ordnen hält die Zellen bündig. Text links, zentriert oder rechts schreibt einen <style scoped>-Block in die aktuelle Folie: Er funktioniert mit deaktiviertem rohem HTML und erscheint in der Vorschau, der HTML-Seite und im PDF. Link (Ctrl+K) schreibt [Text](url) mit markierter url.

Ziehe ein Bild oder ein Video auf den Editor oder füge es ein: Es kommt nach media/ und die Präsentation verweist darauf, ![](media/photo.png); ziehe Dateien auf einen Ordner des Bereichs, um sie dort abzulegen. Eine Datei ist höchstens 8 MiB groß, ein Projekt höchstens 64 MiB (die Zeile unter den Dateien zeigt den belegten Platz). Es wird nichts hochgeladen. Ein https:-Bild funktioniert auch, aber dein Browser lädt es von dessen Server, der dann deine Adresse sieht (es wird kein Referrer gesendet).

Design wählt default, gaia, uncover oder ein Design, das deine Präsentation definiert (/* @theme name */ in einem <style>): Es schreibt theme: in den Front Matter und folgt dem Front Matter, wenn du ihn bearbeitest.

Die Vervollständigung öffnet sich beim Tippen einer Direktive, eines Kinds oder innerhalb von {{ }}. Setze den Cursor auf eine Direktive, ein Kind oder eine Template-Funktion, und dieser Bereich erklärt sie. Klicke in der Übersicht auf eine Folie, um zu ihrer Zeile zu springen.

Vortrag (oder p in der Bildschirmpräsentation) öffnet die Referentenansicht in einem zweiten Fenster: die aktuelle Folie, den nächsten Schritt, die Sprechernotizen (<!-- Kommentare -->), eine Uhr und einen Timer. Pfeiltasten, Leertaste, Pos1 und Ende steuern beide Fenster; Bildschirmpräsentation und Referentenansicht folgen einander, auch über Änderungen hinweg. Erlaube Pop-ups für diese Seite, wenn sich das Fenster nicht öffnet.

Deine Präsentation bleibt in diesem Browser: Sie wird beim Tippen im eigenen Speicher dieses Browsers gesichert, nur für diese Seite, und kehrt nach dem Neuladen zurück; nichts wird irgendwohin gesendet. Entwurf stellt den Entwurf wieder her, den ein geteilter Link, ein Beispiel oder ein anderer Tab ersetzt hat, setzt auf das Beispiel zurück, und Meine Entwürfe verwerfen löscht sie. Link teilen legt die Präsentation in den Link selbst (nicht ihre übrigen Dateien: Lade dafür die .zip des Projekts herunter).

Rückgängig und Wiederholen (Tab Start; Ctrl+Z sowie Ctrl+Shift+Z oder Ctrl+Y; Cmd auf dem Mac) nehmen die letzte Bearbeitung der angezeigten Datei zurück, wobei jede Datei ihre eigene hat, und danach den letzten Vorgang am ganzen Projekt: ein Beispiel, Zurück, eine geöffnete .zip-Datei oder ein geöffneter Ordner, ein wiederhergestellter Entwurf, eine gelöschte oder umbenannte Datei, eine wiederhergestellte Version. Wiederholen bringt alles so zurück, wie es war, als du Rückgängig gedrückt hast.

Versionen bewahren das ganze Projekt in diesem Browser auf: Eine wird kurz nach dem Tippen gespeichert, vor einem Vorgang am ganzen Projekt, bei jedem Export und Teilen, und Version speichern… bewahrt eine unter einem Namen auf. Der Bereich Versionen listet sie auf: Vorschau zeigt eine in der Vorschau, Mit jetzt vergleichen zeigt die Änderungen jeder Datei Zeile für Zeile (F7 springt zur nächsten Änderung), Wiederherstellen stellt sie wieder her, und Rückgängig nimmt das Wiederherstellen zurück. Die neuesten automatischen Versionen bleiben erhalten, danach eine pro Stunde für einen Tag und eine pro Tag für einen Monat; benannte bleiben, bis du sie löschst.

Die Tabs des Menübands: Start (Rückgängig, Wiederholen, Versionen, Beispiel, Design, Vortrag, Link teilen, Entwurf), Einfügen (eine neue Folie, ein Bild oder Hintergrund aus einer Datei, Formel, ein Codeblock mit Zeilennummern, ein Kind: Jedes wird am Cursor in die Präsentation geschrieben, um die Auswahl herum, falls es eine gibt, und Rückgängig nimmt es zurück), Ansicht (die Vorschau, die Minimap, die Folienminiaturen) und Export. Alt zeigt einen Buchstaben über jedem Tab; tippe ihn, dann den Buchstaben über einem Steuerelement. Ctrl+F1 klappt das Menüband ein.

Tab Export: die HTML-Seite der Präsentation mit ihren Medien; ein PDF (Vektor, Text auswählbar), eine PowerPoint- oder OpenDocument-Präsentation (bearbeitbar), in deinem Browser genau so erstellt, wie der Befehl slides sie erstellt. Der erste Export lädt die Export-Engine herunter (etwa 5 MiB, einmalig: dein Browser behält sie) und kompiliert sie beim Eintreffen, danach nur noch die Schriften, die deine Folien verwenden; https:-Bilder werden nur geladen, wenn ihr Server es erlaubt (CORS), data:-Bilder immer.`,
}

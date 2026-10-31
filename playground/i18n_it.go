package playground

// itMessages is the Italian catalogue.
var itMessages = map[Key]string{
	KTabHome: "Home", KTabInsert: "Inserisci", KTabFormat: "Formato", KTabView: "Visualizza", KTabExport: "Esporta",

	KGrpDeck: "Presentazione", KGrpLook: "Aspetto", KGrpPresent: "Presenta", KGrpDraft: "Bozza", KGrpSlides: "Diapositive",
	KGrpPictures: "Immagini", KGrpText: "Testo", KGrpPreview: "Anteprima", KGrpEditor: "Editor", KGrpSaveAs: "Salva con nome",
	KGrpParagraph: "Paragrafo", KGrpTable: "Tabella", KGrpRowsCols: "Righe e colonne", KGrpColumn: "Colonna",
	KGrpSlideText: "Testo diapositiva", KGrpLink: "Collegamento",

	KLblExample: "Esempio", KLblReset: "Ripristina", KLblTheme: "Tema", KLblThemeOption: "Tema: %s", KLblDraft: "Bozza",
	KLblShare: "Condividi link", KLblPresent: "Presenta", KLblPreview: "Anteprima", KLblMinimap: "Minimappa",
	KLblSlides: "Diapositive", KLblFiles: "File", KLblHelp: "Guida", KLblNewSlide: "Nuova diapositiva", KLblImage: "Immagine",
	KLblBackground: "Sfondo", KLblBgLeft: "Sfondo a sinistra", KLblBgRight: "Sfondo a destra",
	KLblMath: "Formula", KLblCode: "Blocco di codice", KLblKind: "Kind", KLblHTML: "Pagina HTML", KLblPDF: "PDF",
	KLblPPTX: "PowerPoint", KLblODP: "OpenDocument",

	KTipExample:    "Parti da uno degli esempi. Il tuo testo resta in questo browser: Bozza › Ripristina la bozza sostituita lo recupera.",
	KTipTheme:      "Il tema della presentazione: sceglierne uno lo scrive nel front matter della presentazione.",
	KTipDraft:      "Ripristina, dimentica o reimposta la presentazione che questo browser conserva per te.",
	KTipReset:      "Rimette il testo dell’esempio; la tua presentazione viene conservata come bozza.",
	KTipPresent:    "Mostra la presentazione qui e apre la vista relatore in una finestra a parte.",
	KTipShare:      "Copia un link che contiene la presentazione stessa: non viene caricato nulla.",
	KTipKind:       "Assegna un kind alla diapositiva: un modello con un aspetto proprio, impostato da un commento <!-- _kind: … -->.",
	KTipNewSlide:   "Aggiunge una diapositiva vuota dopo quella corrente.",
	KTipImage:      "Scegli un file immagine: entra nella presentazione come URI data:, non viene caricato nulla. Il testo selezionato diventa la sua descrizione.",
	KTipBackground: "Scegli un’immagine per lo sfondo della diapositiva; la freccia riempie solo il 40% di sinistra o di destra.",
	KTipMath:       "Una formula in evidenza tra due righe $$; la selezione viene inserita al suo interno.",
	KTipCode:       "Un blocco di codice delimitato con numeri di riga ({lineNumbers}); la selezione viene inserita al suo interno.",
	KTipMode:       "Panoramica impila tutte le diapositive e scorre insieme all’editor; Presentazione è la presentazione vera e propria.",
	KTipMinimap:    "Una mappa della presentazione accanto alla barra di scorrimento dell’editor.",
	KTipThumbs:     "Una colonna di miniature delle diapositive; fai clic su una per raggiungerla.",
	KTipFiles:      "I file del progetto: temi, kind, dati e contenuti multimediali. Una finestra stretta parte senza questo riquadro.",
	KTipHelp:       "La guida su ciò che si trova sotto il cursore e i link alla documentazione. Una finestra stretta parte senza questo riquadro.",
	KTipHTML:       "Un unico file HTML che presenta la presentazione in qualsiasi browser.",
	KTipPDF:        "Un PDF, una pagina per diapositiva, creato qui nel tuo browser.",
	KTipPPTX:       "Un file PowerPoint modificabile, creato qui nel tuo browser.",
	KTipODP:        "Una presentazione OpenDocument modificabile, creata qui nel tuo browser.",

	KModeOverview: "Panoramica", KModeSlideshow: "Presentazione",

	KExWelcome: "Benvenuto", KExTemplates: "Modelli: diapositive generate", KExKinds: "Kind di diapositiva",
	KExHCL: "Front matter HCL", KExGaia: "Temi e direttive locali", KExMedia: "Formule, immagini e sfondi",
	KExGalleryDefault: "Galleria: il tema default", KExGalleryGaia: "Galleria: il tema gaia",
	KExGalleryUncover: "Galleria: il tema uncover", KExGalleryMath: "Galleria: formule matematiche",
	KExGalleryCode:       "Galleria: codice, numeri di riga e riquadri",
	KExGalleryEmoji:      "Galleria: emoji e titoli adattati",
	KExGalleryPresenting: "Galleria: transizioni e note del relatore",
	KExGalleryHCL:        "Galleria: espressioni HCL e dati tipizzati",

	KStatusStarting:           "Avvio in corso…",
	KStatusOpenedShared:       "Aperta la presentazione condivisa nel link.",
	KStatusSharedUnreadable:   "Non è stato possibile leggere la presentazione del link: %s",
	"status.rendered.one":     "%d diapositiva · generata in %d ms",
	"status.rendered.other":   "%d diapositive · generate in %d ms",
	KStatusStale:              "l’anteprima mostra l’ultima presentazione generata",
	"status.problems.one":     "%d problema · %s",
	"status.problems.other":   "%d problemi · %s",
	KStatusDraftNotKept:       " · ⚠ bozza non conservata in questo browser",
	KStatusPresenting:         "Presentazione in corso: la vista relatore è in una finestra a parte; i tasti freccia in una qualsiasi delle due le fanno avanzare entrambe.",
	KStatusLinkCopied:         "Link copiato: la presentazione viaggia nel link stesso (%d caratteri), non viene caricato nulla.",
	KStatusSavedHTML:          "deck.html salvato: aprilo in qualsiasi browser, o presenta da lì.",
	KStatusShareDeckOnly:      " Nel link c’è solo la presentazione: Progetto › Scarica il progetto (.zip) ne conserva i file.",
	KStatusRestoredDraft:      "Ripristinata la tua bozza di %s.",
	KStatusSharedKeptDraft:    "Aperta la presentazione condivisa nel link. La tua bozza di %s è conservata: Bozza › Ripristina la mia bozza.",
	KStatusRestoredDraftReset: "Ripristinata la tua bozza di %s. Ripristina torna all’esempio.",
	KStatusDraftSavedAgain:    "La tua bozza è di nuovo salvata in questo browser.",
	KStatusDraftNotSaved:      "⚠ Questo browser non ha conservato la tua bozza (%s): Condividi link o Scarica HTML conserva la presentazione.",
	KStatusOtherTab:           "Un’altra scheda ha salvato la sua bozza %s. Le tue modifiche qui non la cancelleranno: Bozza › Ripristina la bozza sostituita.",
	KStatusRestoredTheDraft:   "Ripristinata la bozza di %s.",
	KStatusForgot:             "Le tue bozze sono state dimenticate: ora nulla di questa presentazione è memorizzato in questo browser (finché non la modifichi di nuovo).",
	KStatusNoExport:           "⚠ Questa pagina non può esportare: usa il comando slides.",
	KStatusExportRunning:      "Un’esportazione è in corso: %s.",
	KStatusTooLargeExport:     "⚠ La presentazione è troppo grande per essere esportata qui: usa il comando slides.",
	KStatusPreparing:          "Preparazione di %s…",
	KStatusLoadingEnginePct:   "Caricamento del motore di esportazione (una sola volta; il browser lo conserva): %d%% di %.1f MiB",
	KStatusLoadingEngine:      "Caricamento del motore di esportazione (una sola volta; il browser lo conserva): %.1f MiB",
	KStatusExportProgress:     "%s: diapositiva %d di %d…",
	KStatusExportFailed:       "⚠ Esportazione %s non riuscita: %s",
	"status.saved.one":        "%s salvato: %s, %d diapositiva in %.1f s, creato nel tuo browser.",
	"status.saved.other":      "%s salvato: %s, %d diapositive in %.1f s, creato nel tuo browser.",
	KStatusIsMedium:           "%s è un contenuto multimediale: la presentazione lo mostra con %s (trascinalo sull’editor, o dal suo menu: Inserisci nella presentazione).",
	KStatusMadeTheme:          "Creato %s, basato su %s: scegli %s nell’elenco Tema per usarlo.",
	KStatusMadeKind:           "Creato %s: una diapositiva assume il kind con <!-- _kind: %s -->.",
	KStatusMadeData:           "Creato %s: la presentazione lo legge come {{ .title }}.",
	KStatusCannotSaveFile:     "⚠ questa pagina non può salvare un file",
	KStatusSavedZip:           "%s salvato: una volta decompresso, `slides build deck.md` lo compila e l’estensione di VS Code lo apre.",
	KStatusOpenedZip:          "Aperto il .zip: %d file.",
	KStatusOpenedFolder:       "Aperta la cartella: %d file.",
	KStatusAddedSlide:         "Aggiunta una diapositiva dopo quella corrente.",
	KStatusInsertedFormula:    "Inserita una formula in evidenza: scrivi al suo interno.",
	KStatusInsertedCode:       "Inserito un blocco di codice Go con numeri di riga: scrivi al suo interno.",
	KStatusPutFormula:         "Selezione inserita in una formula in evidenza.",
	KStatusPutCode:            "Selezione inserita in un blocco di codice Go con numeri di riga.",
	KStatusNoFilePicker:       "⚠ Questa pagina non può richiedere un file: trascina o incolla un file immagine nell’editor.",
	KStatusKindSet:            "La diapositiva ora è del kind %s.",
	KStatusInsertedRule:       "Inserita una linea orizzontale (***: tre trattini su una riga inizierebbero una nuova diapositiva).",
	KStatusInsertedTable:      "Inserita una tabella: scrivi nelle sue celle; i comandi di allineamento della tabella la mantengono in ordine.",
	KStatusJournalLarge:       "Questo progetto è troppo grande per il salvataggio rapido: le modifiche dell’ultimo mezzo secondo prima della chiusura della pagina potrebbero non essere conservate.",
	KStatusJournalRefused:     "⚠ Il salvataggio rapido è stato rifiutato (%s): le modifiche dell’ultimo mezzo secondo prima della chiusura della pagina potrebbero non essere conservate.",
	KStatusAdded:              "Aggiunto %s al progetto (%d KiB): è conservato in questo browser e finisce nel .zip del progetto.",
	KStatusAddedPut:           "Aggiunto %s al progetto (%d KiB) e inserito nella presentazione: è conservato in questo browser e finisce nel .zip del progetto.",
	KStatusPopupBlocked:       "⚠ Il browser ha bloccato la finestra del relatore: consenti i pop-up per questa pagina, poi Presenta di nuovo.",

	KWhyHeading:    "Un titolo va su una riga del testo della presentazione, non nel front matter né in un blocco di codice.",
	KWhyList:       "Un elenco va su una riga del testo della presentazione, non nel front matter né in un blocco di codice.",
	KWhyQuote:      "Una citazione va su una riga del testo della presentazione, non nel front matter né in un blocco di codice.",
	KWhyNotInTable: "Posiziona prima il cursore in una tabella.",
	KWhyNotInBody:  "Posiziona il cursore in una riga del corpo della tabella: l’intestazione resta.",
	KWhyLastColumn: "Una tabella conserva almeno una colonna.",

	KProblemLine: "riga %d", KProblemError: "errore", KProblemWarning: "avviso",

	KConfirmReset:      "Sostituire la tua presentazione con l’esempio? Il tuo testo resta in questo browser: Bozza › Ripristina la bozza sostituita.",
	KConfirmShareFiles: "Un link contiene solo la presentazione, non gli altri file (temi, kind, contenuti multimediali): scaricare invece il progetto come .zip?",
	KConfirmOpenZip:    "Sostituire la tua presentazione e i suoi file con quelli del .zip? Quello che hai ora resta in questo browser: Bozza › Ripristina la bozza sostituita.",
	KConfirmOpenFolder: "Sostituire la tua presentazione e i suoi file con quelli della cartella? Quello che hai ora resta in questo browser: Bozza › Ripristina la bozza sostituita.",

	KDateToday: "oggi alle %s", KDateAt: "%s alle %s", KDateShort: "%[2]d %[1]s",
	KDateMonths: "gen,feb,mar,apr,mag,giu,lug,ago,set,ott,nov,dic",
	KBytes:      "%d byte", KWarn: "⚠ %s",

	KDraftRestoreMine: "Ripristina la mia bozza (%s)", KDraftRestoreReplaced: "Ripristina la bozza sostituita (%s)",
	KDraftForget: "Dimentica le mie bozze", KDraftResetExample: "Torna all’esempio", KDraftNothing: "Ancora nulla di salvato",

	KSaveAs: "Salva con nome…", KKindTitle: "Kind…",

	KProject: "Progetto", KFilesTreeLabel: "File del progetto", KFilesNewThemeFrom: "Nuovo tema (da %s)",
	KFilesNewKind: "Nuovo kind", KFilesNewData: "Nuovo file di dati", KFilesNewFile: "Nuovo file…", KFilesNewFolder: "Nuova cartella…",
	KFilesAddMedia: "Aggiungi contenuto multimediale…", KFilesOpenZip: "Apri progetto (.zip)…", KFilesOpenFolder: "Apri cartella del progetto…",
	KFilesDownloadZip: "Scarica il progetto (.zip)", KFilesInsertIntoDeck: "Inserisci nella presentazione", KFilesDownload: "Scarica",
	"files.usage.one": "%d file · %s di %d MiB", "files.usage.other": "%d file · %s di %d MiB",

	KPaneProblems: "Problemi", KPaneDocumentation: "Documentazione",
	KFrameTitlePreview: "Anteprima della diapositiva", KFrameTitleRail: "Miniature delle diapositive",

	KDocGettingStarted: "Per iniziare", KDocMarp: "Compatibilità con Marp", KDocTemplates: "Modelli",
	KDocKinds: "Kind", KDocHCL: "HCL", KDocCLI: "Riga di comando", KDocVSCode: "VS Code", KDocPlayground: "Playground",

	KTableColumn: "Colonna %d",

	KErrExists: "%s esiste già", KErrDeckName: "la presentazione resta deck.md: il comando slides e il playground la leggono con questo nome",
	KErrNoSuchFile: "%s: file inesistente", KErrNotFileName: "%q non è un nome di file",
	KErrDeckTooLarge: "la presentazione è di %d KiB; il playground elabora presentazioni fino a %d KiB",
	KErrMaxFiles:     "un progetto contiene al massimo %d file", KErrPathReason: "%s: %v",
	KErrProjectOver: "il progetto supererebbe %d MiB", KErrFileSize: "%s è di %d KiB: un file è al massimo di %d KiB",
	KErrShareTooLargeDeck: "la presentazione supera quanto il playground accetta; scaricala invece",
	KErrShareTooLargeLink: "la presentazione è troppo grande per viaggiare in un link; scaricala invece",
	KErrShareNoDeck:       "il link non contiene alcuna presentazione", KErrShareTooLong: "la presentazione del link è troppo lunga",
	KErrShareDamaged: "la presentazione del link è danneggiata (potrebbe essere stata troncata)",
	KErrShareLarger:  "la presentazione del link supera quanto il playground accetta", KErrShareNotText: "la presentazione del link non è testo",
	KErrZipNoDeck: "l’archivio non contiene alcuna presentazione: metti deck.md nella radice (oppure un solo file .md)",
	KErrZipSize:   "l’archivio è di %d MiB: al massimo %d MiB", KErrZipNot: "non è un archivio .zip: %v",
	KErrZipEntries: "l’archivio elenca %d voci: al massimo %d", KErrZipLinks: "%s: i collegamenti non sono accettati",
	KErrZipFileOver: "%s supera %d KiB: un file non può superare questa dimensione",
	KErrZipBomb:     "l’archivio si espande più di %d volte: rifiutato come possibile zip bomb",
	KErrZipTwice:    "%s è presente due volte nell’archivio", KErrZipMoreThan: "più di %d KiB",
	KErrFileOver: "%s supera %s", KErrFailed: "non riuscito",
	KErrFolderCount: "la cartella contiene %d file: un progetto ne contiene al massimo %d", KErrFolderOver: "la cartella supera %s",
	KErrFolderMore: "la cartella contiene più di %d file", KErrNotImage: "%s non è un’immagine.",

	"fmt.label.bold": "Grassetto", "fmt.label.italic": "Corsivo", "fmt.label.underline": "Sottolineato",
	"fmt.label.strike": "Barrato", "fmt.label.code": "Codice inline",
	"fmt.label.h1": "Titolo 1", "fmt.label.h2": "Titolo 2", "fmt.label.h3": "Titolo 3",
	"fmt.label.bullets": "Elenco puntato", "fmt.label.numbered": "Numerazione", "fmt.label.quote": "Citazione", "fmt.label.rule": "Linea",
	"fmt.label.table": "Tabella", "fmt.label.table-format": "Allinea",
	"fmt.label.row-above": "Riga sopra", "fmt.label.row-below": "Riga sotto", "fmt.label.row-delete": "Elimina riga",
	"fmt.label.col-left": "Colonna a sinistra", "fmt.label.col-right": "Colonna a destra", "fmt.label.col-delete": "Elimina colonna",
	"fmt.label.colalign-left": "Colonna a sinistra", "fmt.label.colalign-center": "Colonna al centro",
	"fmt.label.colalign-right": "Colonna a destra",
	"fmt.label.slide-left":     "Testo a sinistra", "fmt.label.slide-center": "Testo al centro", "fmt.label.slide-right": "Testo a destra",
	"fmt.label.link": "Collegamento",

	"fmt.tip.bold":            "Mette in grassetto il testo selezionato, o la parola sotto il cursore: **testo**. Premi di nuovo per annullare.",
	"fmt.tip.italic":          "Corsivo: _testo_. Premi di nuovo per annullare.",
	"fmt.tip.underline":       "Markdown non ha la sottolineatura: questo comando racchiude il testo in <u>…</u>, un tag che l’elenco HTML consentito predefinito di Marp mantiene, quindi funziona anche con l’HTML grezzo disattivato. Premi di nuovo per annullare.",
	"fmt.tip.strike":          "Barrato: ~~testo~~. Premi di nuovo per annullare.",
	"fmt.tip.code":            "Codice inline: `testo`. Premi di nuovo per annullare.",
	"fmt.tip.h1":              "Trasforma la riga (o ogni riga selezionata) in un titolo di livello 1: # testo. Premi di nuovo per tornare al testo normale.",
	"fmt.tip.h2":              "Titolo di livello 2: ## testo.",
	"fmt.tip.h3":              "Titolo di livello 3: ### testo.",
	"fmt.tip.bullets":         "Elenco puntato: - voce, una per riga selezionata. Premi di nuovo per annullare.",
	"fmt.tip.numbered":        "Elenco numerato: 1. voce, 2. voce… Premi di nuovo per annullare.",
	"fmt.tip.quote":           "Citazione: > testo. Premi di nuovo per togliere un livello. Ctrl+Shift+.",
	"fmt.tip.rule":            "Una linea orizzontale. Marp legge ---, *** e ___ come una nuova diapositiva, quindi questo comando scrive il tag <hr>, che l’elenco HTML consentito predefinito mantiene.",
	"fmt.tip.table":           "Inserisce una tabella dopo la riga corrente: scegli le colonne e le righe.",
	"fmt.tip.table-format":    "Allinea la tabella intorno al cursore: ogni colonna viene riempita fino alla larghezza della cella più larga, i segni di allineamento restano.",
	"fmt.tip.row-above":       "Aggiunge una riga vuota sopra la riga del cursore (sotto l’intestazione se si parte dall’intestazione).",
	"fmt.tip.row-below":       "Aggiunge una riga vuota sotto la riga del cursore.",
	"fmt.tip.row-delete":      "Rimuove la riga del cursore; l’intestazione resta.",
	"fmt.tip.col-left":        "Aggiunge una colonna a sinistra di quella del cursore.",
	"fmt.tip.col-right":       "Aggiunge una colonna a destra di quella del cursore.",
	"fmt.tip.col-delete":      "Rimuove la colonna del cursore; una tabella ne conserva sempre una.",
	"fmt.tip.colalign-left":   "Allinea a sinistra la colonna della tabella sotto il cursore (:--). Premi di nuovo per rimuoverlo.",
	"fmt.tip.colalign-center": "Centra la colonna della tabella sotto il cursore (:-:). Premi di nuovo per rimuoverlo.",
	"fmt.tip.colalign-right":  "Allinea a destra la colonna della tabella sotto il cursore (--:). Premi di nuovo per rimuoverlo.",
	"fmt.tip.slide-left":      "Rimette il testo della diapositiva a sinistra, come nel tema: rimuove il blocco di stile scritto da questa scheda.",
	"fmt.tip.slide-center":    "Centra il testo della diapositiva in cui si trova il cursore, con un blocco <style scoped> all’inizio della diapositiva. Funziona con l’HTML grezzo disattivato e compare nell’anteprima, nella pagina HTML e nel PDF.",
	"fmt.tip.slide-right":     "Allinea a destra il testo della diapositiva in cui si trova il cursore (un blocco <style scoped>).",
	"fmt.tip.link":            "Trasforma la selezione, o la parola, in un collegamento: [testo](url) con l’url selezionato per scriverlo. In un collegamento, seleziona il suo url.",

	KGrpHistory: "Cronologia", KLblUndo: "Annulla", KLblRedo: "Ripeti",
	KTipUndo:     "Ritira l’ultima modifica: una modifica del file mostrato, altrimenti un’operazione sull’intero progetto.",
	KTipUndoStep: "Annulla: %s.",
	KTipRedo:     "Rifà ciò che Annulla ha ritirato.",
	KTipRedoStep: "Ripeti: %s.",
	KStepExample: "l’esempio «%s»", KStepReset: "Ripristina", KStepOpenZip: "Apri uno .zip", KStepOpenFolder: "Apri una cartella",
	KStepRename: "Rinomina %s", KStepDelete: "Elimina %s", KStepRestoreDraft: "Ripristina la bozza", KStepRestoreVersion: "Recupera %s",
	KStatusUndone: "Annullato: %s.", KStatusRedone: "Ripetuto: %s.",

	KLblSaveVersion: "Salva versione…", KLblVersions: "Versioni",
	KTipSaveVersion: "Conserva l’intero progetto così com’è ora, con un nome, in questo browser.",
	KTipVersions:    "Le versioni del progetto, conservate in questo browser: visualizzane una in anteprima, confrontala con ora, recuperala.",
	KAskVersionName: "Nome di questa versione:",
	KVersionAuto:    "Automatica", KVersionBefore: "Prima: %s", KVersionExport: "Esportata: %s", KVersionShared: "Link condiviso",
	KVersionLine:   "%s · %s · %s",
	KVersionsEmpty: "Nessuna versione per ora. Ne viene conservata una poco dopo che smetti di scrivere, prima di un’operazione sull’intero progetto, a ogni esportazione o condivisione, e quando ne salvi una.",
	KVersionsOff:   "Le versioni non vengono conservate: questo browser non mette a disposizione spazio per la pagina.",
	KLblOnlyNamed:  "Solo con nome", KLblVersionPreview: "Anteprima", KLblVersionCompare: "Confronta con ora",
	KLblVersionRestore: "Recupera", KLblVersionDelete: "Elimina", KLblBackToNow: "Torna a ora", KLblSideBySide: "Affiancate",
	KCompareTitle:    "%s confrontata con ora",
	KCompareFileLine: "%s (+%d −%d)",
	KCompareSame:     "Nessuna differenza con ora.",
	KCompareBinary:   "%s non è testo: %s allora, %s ora.",
	KCompareNow:      "Ora",

	KStatusVersionSaved:    "Versione «%s» salvata.",
	KStatusVersionRestored: "%s recuperata; Annulla la ritira.",
	KStatusVersionDeleted:  "Versione %s eliminata.",
	KStatusPreviewing:      "L’anteprima mostra %s; Torna a ora, o una modifica, mostra di nuovo la presentazione.",
	KStatusVersionsFull:    "Lo spazio di questa pagina è esaurito: le versioni automatiche più vecchie sono state eliminate.",
	KStatusVersionsFailed:  "Non è stato possibile conservare una versione: %s",
	KStatusVersionLimit:    "Vengono conservate al massimo %d versioni con nome: eliminane una per salvarne un’altra.",
	KConfirmDeleteVersion:  "Eliminare la versione %s? L’operazione non può essere annullata.",

	KHelpGeneral: `Scrivi una presentazione a sinistra; le diapositive seguono a destra.

- Le diapositive sono separate da una riga di tre trattini: ---
- Il front matter, tra le prime due righe ---, contiene le direttive globali (theme, paginate, header, footer, size) e le chiavi proprie di go-slides: kinds, data, template.
- <!-- _class: lead --> in una diapositiva è una direttiva per quella sola diapositiva.
- {{ .title }}, {{ page }}/{{ total }}, {{ range seq 3 }}: template Go.
- kind: title applica un kind (un insieme di direttive con un nome).
- Formule: $e^{i\pi}$ nel testo, $$ ... $$ su righe a sé.
- Immagini: ![w:300](url) ne imposta la dimensione, ![sepia](url) le applica un filtro; ![bg](url) è uno sfondo, ![bg left:40%](url) uno sfondo diviso accanto al testo.

File, a sinistra dell’editor, è il progetto della presentazione, come lo legge il comando slides: deck.md, slides.yaml, themes/, kinds/, media/. Fai clic su un file per modificarlo. Clic destro (o Shift+F10) per Nuovo file, Nuova cartella, Rinomina (F2), Elimina, Nuovo tema, Nuovo kind; trascina un file su una cartella per spostarlo. Il menu Progetto crea un tema basato su default, gaia o uncover (entra nell’elenco Tema), un kind, un file di dati, e apre o scarica l’intero progetto come .zip (una volta decompresso, slides build deck.md lo compila).

Riquadri: trascina la barra tra due riquadri per ridimensionarli (File, l’editor, Diapositive, l’anteprima, Guida); fai doppio clic su una barra, o premi Invio su di essa, per chiudere il riquadro e riaprirlo; F6 passa da una barra all’altra, i tasti freccia la spostano di 16 px, Home e Fine vanno al minimo e al massimo, Esc torna all’editor. Le dimensioni impostate sono conservate in questo browser; la scheda Visualizza permette comunque di disattivare i riquadri.

Scheda Formato: Grassetto, Corsivo, Barrato e Codice inline agiscono sulla selezione o sulla parola sotto il cursore e si alternano (Ctrl+B, Ctrl+I, Ctrl+Shift+X, Ctrl+E; Cmd su Mac): ogni pulsante appare premuto all’interno dei suoi segni, e un solo Annulla ritira una sola azione. Sottolineato (Ctrl+U) racchiude il testo in <u>…</u>, perché Markdown non ha la sottolineatura e l’elenco HTML predefinito di Marp mantiene quel tag. I titoli da 1 a 3 (Ctrl+Alt+1..3), l’elenco puntato (Ctrl+Shift+8), la numerazione (Ctrl+Shift+7) e la citazione agiscono sulle righe selezionate; la linea è un tag <hr>, perché ---, *** e ___ inizierebbero una nuova diapositiva. Tabella apre una griglia di colonne per righe; con il cursore in una tabella aggiungi o rimuovi una riga o una colonna, allinei la colonna (:--, :-:, --:) e Allinea mantiene le celle incolonnate. Testo a sinistra, al centro o a destra scrive un blocco <style scoped> nella diapositiva corrente: funziona con l’HTML grezzo disattivato e compare nell’anteprima, nella pagina HTML e nel PDF. Collegamento (Ctrl+K) scrive [testo](url) con l’url selezionato.

Trascina o incolla un’immagine o un video nell’editor: finisce in media/ e la presentazione lo nomina, ![](media/photo.png); trascina i file su una cartella del pannello per metterli lì. Un file è al massimo di 8 MiB e un progetto di 64 MiB (la riga sotto i file indica quanto è usato). Non viene caricato nulla. Funziona anche un’immagine https:, ma il tuo browser la recupera dal suo server, che vede così il tuo indirizzo (non viene inviato alcun referrer).

Tema propone default, gaia, uncover o un tema definito dalla tua presentazione (/* @theme name */ in un <style>): scrive theme: nel front matter e segue il front matter quando lo modifichi.

Il completamento si apre mentre scrivi una direttiva, un kind, o dentro {{ }}. Posiziona il cursore su una direttiva, un kind o una funzione dei template e questo riquadro la spiega. Fai clic su una diapositiva nella panoramica per andare alla sua riga.

Presenta (o p nella presentazione) apre la vista relatore in una seconda finestra: la diapositiva corrente, il passaggio successivo, le note del relatore (<!-- commenti -->), un orologio e un timer. I tasti freccia, Spazio, Home e Fine fanno avanzare entrambe le finestre; la presentazione e la vista relatore si seguono, anche durante le modifiche. Consenti i pop-up per questo sito se la finestra non si apre.

La tua presentazione resta in questo browser: viene salvata mentre scrivi nella memoria propria del browser, solo per questo sito, e ricompare quando ricarichi la pagina; non viene inviato nulla da nessuna parte. Bozza ripristina la bozza che un link condiviso, un esempio o un’altra scheda ha sostituito, torna all’esempio, e Dimentica le mie bozze le cancella. Condividi link mette la presentazione nel link stesso (non gli altri file: scarica il .zip del progetto per quelli).

Annulla e Ripeti (scheda Home; Ctrl+Z, e Ctrl+Shift+Z o Ctrl+Y; Cmd su un Mac) ritirano l’ultima modifica del file mostrato, ogni file con la propria, e poi l’ultima operazione sull’intero progetto: un esempio, Ripristina, uno .zip o una cartella aperti, una bozza ripristinata, un file eliminato o rinominato, una versione recuperata. Ripeti riporta tutto com’era quando hai premuto Annulla.

Le versioni conservano l’intero progetto in questo browser: ne viene salvata una poco dopo che smetti di scrivere, prima di un’operazione sull’intero progetto, a ogni esportazione e condivisione, e Salva versione… ne conserva una con un nome. Il pannello Versioni le elenca: Anteprima ne mostra una nell’anteprima, Confronta con ora mostra le modifiche di ogni file riga per riga (F7 passa alla modifica successiva), Recupera la rimette al suo posto, e Annulla ritira il recupero. Si conservano le versioni automatiche più recenti, poi una all’ora per un giorno e una al giorno per un mese; quelle con nome restano finché non le elimini.

Le schede della barra multifunzione: Home (Annulla, Ripeti, versioni, esempio, tema, Presenta, Condividi link, Bozza), Inserisci (una nuova diapositiva, un’immagine o uno sfondo da un file, una formula, un blocco di codice con numeri di riga, un kind: ognuno viene scritto nella presentazione alla posizione del cursore, intorno alla selezione se c’è, e Annulla lo ritira), Visualizza (l’anteprima, la minimappa, le miniature delle diapositive) ed Esporta. Alt mostra una lettera sopra ogni scheda; digitala, poi la lettera sopra un controllo. Ctrl+F1 comprime la barra multifunzione.

Scheda Esporta: la pagina HTML della presentazione, con i contenuti multimediali dentro; una presentazione PDF (vettoriale, testo selezionabile), PowerPoint o OpenDocument (modificabile), creata nel tuo browser esattamente come le crea il comando slides. La prima esportazione scarica il motore di esportazione (circa 5 MiB, una sola volta: il browser lo conserva), compilandolo man mano che arriva, e poi solo i tipi di carattere usati dalle tue diapositive; le immagini https: vengono recuperate solo se il loro server lo consente (CORS), quelle data: sempre.`,

	"err.store.pathlen":     "%q: un percorso è lungo da 1 a %d byte",
	"err.store.pathrel":     "%q: un percorso di progetto è relativo e usa /",
	"err.store.pathctl":     "%q: carattere di controllo in un percorso",
	"err.store.pathleaves":  "%q: il percorso esce dal progetto",
	"err.store.filesize":    "%s è di %d KiB: un file è al massimo di %d KiB",
	"err.store.maxfiles":    "un progetto contiene al massimo %d file",
	"err.store.projectsize": "il progetto sarebbe di %d MiB: al massimo %d MiB",
	"err.store.nostore":     "questo browser non conserva alcuna memoria per la pagina",
	"err.media.notfile":     "%q non è un file del progetto",
	"err.media.textover":    "i file di testo del progetto superano %d KiB",
	"err.media.notmedium":   "%q non è un media del progetto",
	"err.media.mediaover":   "i media del progetto superano %d MiB",
	"err.tpl.mapping":       "%s: %s è un mapping; elenca %s al suo interno a mano",
	"err.tpl.config":        "%s: %v",
	"err.tpl.notmapping":    "%s: la configurazione non è un mapping",

	"code.kind.unknown":      "La diapositiva indica un kind che non è definito.",
	"code.directive.unknown": "La direttiva non è nota a go-slides.",
	"code.template.function": "Il modello chiama una funzione che non esiste.",
	"code.template.exec":     "Il modello è fallito durante la compilazione.",
	"code.template.syntax":   "Il modello non è ben formato ({{ … }} sbilanciato o inatteso).",
	"code.yaml.syntax":       "Lo YAML non è ben formato.",
	"code.hcl.unsupported":   "Il file HCL usa qualcosa che go-slides non legge.",
	"code.theme.unknown":     "Il tema della presentazione non esiste.",
	"code.file.missing":      "Non si trova un file letto dalla presentazione.",

	"tk.filetree.new_file":    "Nuovo file…",
	"tk.filetree.new_folder":  "Nuova cartella…",
	"tk.filetree.rename":      "Rinomina…",
	"tk.filetree.delete":      "Elimina",
	"tk.filetree.cancel":      "Annulla",
	"tk.filetree.err.slash":   "Il nome di un file o di una cartella non può iniziare con una barra.",
	"tk.filetree.err.invalid": "Il nome '%s' non è valido come nome di file o cartella. Scegliere un nome diverso.",
	"tk.filetree.err.exists":  "In questo percorso esiste già un file o una cartella '%s'. Scegliere un nome diverso.",
	"tk.filetree.warn.space":  "Rilevato spazio iniziale o finale nel nome del file o della cartella.",
	"tk.find.replace":         "Sostituisci",
	"tk.find.replace_all":     "Tutto",
	"tk.find.bad_pattern":     "Criterio non valido",
	"tk.find.no_results":      "Nessun risultato",
	"tk.find.count":           "%d di %d",
	"tk.find.a11y":            "Trova e sostituisci",
	"tk.ribbon.a11y":          "Barra multifunzione",
	"tk.ribbon.a11y_tabs":     "Schede della barra multifunzione",
	"tk.filechooser.open":     "Apri",
	"tk.filechooser.cancel":   "Annulla",
	"tk.dialog.ok":            "OK",
}

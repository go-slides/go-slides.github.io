package playground

// plMessages is the Polish catalogue.
var plMessages = map[Key]string{
	KTabHome: "Start", KTabInsert: "Wstawianie", KTabFormat: "Format", KTabView: "Widok", KTabExport: "Eksport",

	KGrpDeck: "Prezentacja", KGrpLook: "Wygląd", KGrpPresent: "Pokaz", KGrpDraft: "Wersja robocza", KGrpSlides: "Slajdy",
	KGrpPictures: "Obrazy", KGrpText: "Tekst", KGrpPreview: "Podgląd", KGrpEditor: "Edytor", KGrpSaveAs: "Zapisz jako",
	KGrpParagraph: "Akapit", KGrpTable: "Tabela", KGrpRowsCols: "Wiersze i kolumny", KGrpColumn: "Kolumna",
	KGrpSlideText: "Tekst slajdu", KGrpLink: "Łącze",

	KLblExample: "Przykład", KLblReset: "Resetuj", KLblTheme: "Motyw", KLblThemeOption: "Motyw: %s", KLblDraft: "Wersja robocza",
	KLblShare: "Udostępnij", KLblPresent: "Pokaz", KLblPreview: "Podgląd", KLblMinimap: "Minimapa",
	KLblSlides: "Slajdy", KLblFiles: "Pliki", KLblHelp: "Pomoc", KLblNewSlide: "Nowy slajd", KLblImage: "Obraz",
	KLblBackground: "Tło", KLblBgLeft: "Tło po lewej", KLblBgRight: "Tło po prawej",
	KLblMath: "Matematyka", KLblCode: "Blok kodu", KLblKind: "Kind", KLblHTML: "Strona HTML", KLblPDF: "PDF",
	KLblPPTX: "PowerPoint", KLblODP: "OpenDocument",

	KTipExample:    "Zacznij od jednego z przykładów. Twój tekst zostaje w tej przeglądarce: Wersja robocza ▸ Przywróć zastąpioną go przywraca.",
	KTipTheme:      "Motyw prezentacji: wybranie go zapisuje go we front matter prezentacji.",
	KTipDraft:      "Przywróć, zapomnij lub zresetuj prezentację, którą ta przeglądarka dla Ciebie przechowuje.",
	KTipReset:      "Przywraca tekst przykładu; Twoja prezentacja zostaje jako wersja robocza.",
	KTipPresent:    "Pokazuje tu pokaz slajdów i otwiera widok prezentera we własnym oknie.",
	KTipShare:      "Kopiuje łącze zawierające samą prezentację: nic nie jest przesyłane.",
	KTipKind:       "Nadaje slajdowi kind: szablon z własnym wyglądem, ustawiany komentarzem <!-- _kind: … -->.",
	KTipNewSlide:   "Dodaje pusty slajd po bieżącym.",
	KTipImage:      "Wybierz plik obrazu: trafia do prezentacji jako identyfikator URI data:, nic nie jest przesyłane. Zaznaczony tekst staje się jego opisem.",
	KTipBackground: "Wybierz obraz tła slajdu; strzałka wypełnia tylko lewe lub prawe 40%.",
	KTipMath:       "Wzór wyświetlany między wierszami $$; zaznaczenie trafia do środka.",
	KTipCode:       "Ogrodzony blok kodu z numerami wierszy ({lineNumbers}); zaznaczenie trafia do środka.",
	KTipMode:       "Przegląd układa wszystkie slajdy w stos i przewija się razem z edytorem; Pokaz slajdów to sama prezentacja.",
	KTipMinimap:    "Mapa prezentacji obok paska przewijania edytora.",
	KTipThumbs:     "Kolumna miniatur slajdów; kliknij jedną, aby do niej przejść.",
	KTipFiles:      "Pliki projektu: motywy, kinds, dane i multimedia. W wąskim oknie panel jest domyślnie ukryty.",
	KTipHelp:       "Pomoc dotycząca miejsca, w którym jest kursor, i łącza do dokumentacji. W wąskim oknie panel jest domyślnie ukryty.",
	KTipHTML:       "Jeden plik HTML, który pokazuje prezentację w dowolnej przeglądarce.",
	KTipPDF:        "Plik PDF, jedna strona na slajd, utworzony tu, w Twojej przeglądarce.",
	KTipPPTX:       "Edytowalny plik PowerPoint, utworzony tu, w Twojej przeglądarce.",
	KTipODP:        "Edytowalna prezentacja OpenDocument, utworzona tu, w Twojej przeglądarce.",

	KModeOverview: "Przegląd", KModeSlideshow: "Pokaz slajdów",

	KExWelcome: "Witamy", KExTemplates: "Szablony: generowane slajdy", KExKinds: "Kinds slajdów",
	KExHCL: "Front matter w HCL", KExGaia: "Motywy i dyrektywy lokalne", KExMedia: "Matematyka, obrazy i tła",
	KExGalleryDefault: "Galeria: motyw default", KExGalleryGaia: "Galeria: motyw gaia",
	KExGalleryUncover: "Galeria: motyw uncover", KExGalleryMath: "Galeria: matematyka",
	KExGalleryCode:       "Galeria: kod, numery wierszy i objaśnienia",
	KExGalleryEmoji:      "Galeria: emoji i dopasowane nagłówki",
	KExGalleryPresenting: "Galeria: przejścia i notatki prelegenta",
	KExGalleryHCL:        "Galeria: wyrażenia HCL i dane typowane",

	KStatusStarting:           "Uruchamianie…",
	KStatusOpenedShared:       "Otwarto prezentację udostępnioną w łączu.",
	KStatusSharedUnreadable:   "Nie można odczytać prezentacji z łącza: %s",
	"status.rendered.one":     "%d slajd · wyrenderowano w %d ms",
	"status.rendered.few":     "%d slajdy · wyrenderowano w %d ms",
	"status.rendered.many":    "%d slajdów · wyrenderowano w %d ms",
	"status.rendered.other":   "%d slajdów · wyrenderowano w %d ms",
	KStatusStale:              "podgląd pokazuje ostatnią prezentację, która została wyrenderowana",
	"status.problems.one":     "%d problem · %s",
	"status.problems.few":     "%d problemy · %s",
	"status.problems.many":    "%d problemów · %s",
	"status.problems.other":   "%d problemów · %s",
	KStatusDraftNotKept:       " · ⚠ wersja robocza nie została zachowana w tej przeglądarce",
	KStatusPresenting:         "Prezentowanie: widok prezentera jest we własnym oknie; strzałki w każdym z nich poruszają oba.",
	KStatusLinkCopied:         "Skopiowano łącze: prezentacja podróżuje w samym łączu (%d znaków), nic nie jest przesyłane.",
	KStatusSavedHTML:          "Zapisano deck.html: otwórz go w dowolnej przeglądarce lub prezentuj z niego.",
	KStatusShareDeckOnly:      " W łączu jest tylko prezentacja: Projekt ▸ Pobierz projekt (.zip) zachowuje jej pliki.",
	KStatusRestoredDraft:      "Przywrócono Twoją wersję roboczą (%s).",
	KStatusSharedKeptDraft:    "Otwarto prezentację udostępnioną w łączu. Twoja wersja robocza (%s) została zachowana: Wersja robocza ▸ Przywróć moją wersję.",
	KStatusRestoredDraftReset: "Przywrócono Twoją wersję roboczą (%s). Resetuj wraca do przykładu.",
	KStatusDraftSavedAgain:    "Twoja wersja robocza jest znów zapisywana w tej przeglądarce.",
	KStatusDraftNotSaved:      "⚠ Ta przeglądarka nie zachowała Twojej wersji roboczej (%s): Udostępnij lub Pobierz HTML zachowuje prezentację.",
	KStatusOtherTab:           "Inna karta zapisała swoją wersję roboczą (%s). Twoje zmiany tutaj jej nie usuną: Wersja robocza ▸ Przywróć zastąpioną.",
	KStatusRestoredTheDraft:   "Przywrócono wersję roboczą (%s).",
	KStatusForgot:             "Zapomniano Twoje wersje robocze: nic z tej prezentacji nie jest już przechowywane w tej przeglądarce (do czasu ponownej edycji).",
	KStatusNoExport:           "⚠ Ta strona nie może eksportować: użyj polecenia slides.",
	KStatusExportRunning:      "Trwa eksport: %s.",
	KStatusTooLargeExport:     "⚠ Prezentacja jest zbyt duża, aby ją tu wyeksportować: użyj polecenia slides.",
	KStatusPreparing:          "Przygotowywanie: %s…",
	KStatusLoadingEnginePct:   "Ładowanie silnika eksportu (jednorazowo; przeglądarka go zachowuje): %d%% z %.1f MiB",
	KStatusLoadingEngine:      "Ładowanie silnika eksportu (jednorazowo; przeglądarka go zachowuje): %.1f MiB",
	KStatusExportProgress:     "%s: slajd %d z %d…",
	KStatusExportFailed:       "⚠ Eksport do %s nie powiódł się: %s",
	"status.saved.one":        "Zapisano %s: %s, %d slajd w %.1f s, utworzony w Twojej przeglądarce.",
	"status.saved.few":        "Zapisano %s: %s, %d slajdy w %.1f s, utworzone w Twojej przeglądarce.",
	"status.saved.many":       "Zapisano %s: %s, %d slajdów w %.1f s, utworzone w Twojej przeglądarce.",
	"status.saved.other":      "Zapisano %s: %s, %d slajdów w %.1f s, utworzone w Twojej przeglądarce.",
	KStatusIsMedium:           "%s to multimedium: prezentacja pokazuje je za pomocą %s (upuść je na edytor albo użyj menu pliku: Wstaw do prezentacji).",
	KStatusMadeTheme:          "Utworzono %s na bazie %s: wybierz %s na liście Motyw, aby go użyć.",
	KStatusMadeKind:           "Utworzono %s: slajd przyjmuje ten kind za pomocą <!-- _kind: %s -->.",
	KStatusMadeData:           "Utworzono %s: prezentacja odczytuje go jako {{ .title }}.",
	KStatusCannotSaveFile:     "⚠ ta strona nie może zapisać pliku",
	KStatusSavedZip:           "Zapisano %s: po rozpakowaniu `slides build deck.md` go buduje, a rozszerzenie VS Code go otwiera.",
	KStatusOpenedZip:          "Otwarto plik .zip: %d plików.",
	KStatusOpenedFolder:       "Otwarto folder: %d plików.",
	KStatusAddedSlide:         "Dodano slajd po bieżącym.",
	KStatusInsertedFormula:    "Wstawiono wzór wyświetlany: pisz w jego wnętrzu.",
	KStatusInsertedCode:       "Wstawiono blok kodu Go z numerami wierszy: pisz w jego wnętrzu.",
	KStatusPutFormula:         "Umieszczono zaznaczenie we wzorze wyświetlanym.",
	KStatusPutCode:            "Umieszczono zaznaczenie w bloku kodu Go z numerami wierszy.",
	KStatusNoFilePicker:       "⚠ Ta strona nie może poprosić o plik: upuść lub wklej plik obrazu do edytora.",
	KStatusKindSet:            "Kind slajdu to teraz %s.",
	KStatusInsertedRule:       "Wstawiono linię poziomą (***: trzy myślniki w wierszu rozpoczęłyby nowy slajd).",
	KStatusInsertedTable:      "Wstawiono tabelę: pisz w jej komórkach; polecenia wyrównywania tabeli utrzymują w niej porządek.",
	KStatusJournalLarge:       "Ten projekt jest zbyt duży na szybki zapis: zmiany z ostatniej pół sekundy przed zamknięciem strony mogą nie zostać zachowane.",
	KStatusJournalRefused:     "⚠ Szybki zapis został odrzucony (%s): zmiany z ostatniej pół sekundy przed zamknięciem strony mogą nie zostać zachowane.",
	KStatusAdded:              "Dodano %s do projektu (%d KiB): jest przechowywany w tej przeglądarce i trafia do pliku .zip projektu.",
	KStatusAddedPut:           "Dodano %s do projektu (%d KiB) i umieszczono w prezentacji: jest przechowywany w tej przeglądarce i trafia do pliku .zip projektu.",
	KStatusPopupBlocked:       "⚠ Przeglądarka zablokowała okno prezentera: zezwól na wyskakujące okna dla tej strony, a potem ponów Pokaz.",

	KWhyHeading:    "Nagłówek idzie do wiersza tekstu prezentacji, a nie do front matter ani bloku kodu.",
	KWhyList:       "Lista idzie do wiersza tekstu prezentacji, a nie do front matter ani bloku kodu.",
	KWhyQuote:      "Cytat idzie do wiersza tekstu prezentacji, a nie do front matter ani bloku kodu.",
	KWhyNotInTable: "Najpierw umieść kursor w tabeli.",
	KWhyNotInBody:  "Umieść kursor w wierszu treści tabeli: nagłówek zostaje.",
	KWhyLastColumn: "Tabela zachowuje co najmniej jedną kolumnę.",

	KProblemLine: "wiersz %d", KProblemError: "błąd", KProblemWarning: "ostrzeżenie",

	KConfirmReset:      "Zastąpić Twoją prezentację przykładem? Twój tekst zostaje w tej przeglądarce: Wersja robocza ▸ Przywróć zastąpioną.",
	KConfirmShareFiles: "Łącze zawiera tylko prezentację, bez jej pozostałych plików (motywów, kinds, multimediów): pobrać zamiast tego projekt jako plik .zip?",
	KConfirmOpenZip:    "Zastąpić Twoją prezentację i jej pliki zawartością pliku .zip? To, co masz teraz, zostaje w tej przeglądarce: Wersja robocza ▸ Przywróć zastąpioną.",
	KConfirmOpenFolder: "Zastąpić Twoją prezentację i jej pliki zawartością folderu? To, co masz teraz, zostaje w tej przeglądarce: Wersja robocza ▸ Przywróć zastąpioną.",

	KDateToday: "dziś o %s", KDateAt: "%s o %s", KDateShort: "%[2]d %[1]s",
	KDateMonths: "sty,lut,mar,kwi,maj,cze,lip,sie,wrz,paź,lis,gru",
	KBytes:      "%d B", KWarn: "⚠ %s",

	KDraftRestoreMine: "Przywróć moją wersję (%s)", KDraftRestoreReplaced: "Przywróć zastąpioną (%s)",
	KDraftForget: "Zapomnij moje wersje robocze", KDraftResetExample: "Resetuj do przykładu", KDraftNothing: "Nic jeszcze nie zapisano",

	KSaveAs: "Zapisz jako…", KKindTitle: "Kind…",

	KProject: "Projekt", KFilesTreeLabel: "Pliki projektu", KFilesNewThemeFrom: "Nowy motyw (na bazie %s)",
	KFilesNewKind: "Nowy kind", KFilesNewData: "Nowy plik danych", KFilesNewFile: "Nowy plik…", KFilesNewFolder: "Nowy folder…",
	KFilesAddMedia: "Dodaj multimedia…", KFilesOpenZip: "Otwórz projekt (.zip)…", KFilesOpenFolder: "Otwórz folder projektu…",
	KFilesDownloadZip: "Pobierz projekt (.zip)", KFilesInsertIntoDeck: "Wstaw do prezentacji", KFilesDownload: "Pobierz",
	"files.usage.one": "%d plik · %s z %d MiB", "files.usage.few": "%d pliki · %s z %d MiB",
	"files.usage.many": "%d plików · %s z %d MiB", "files.usage.other": "%d plików · %s z %d MiB",

	KPaneProblems: "Problemy", KPaneDocumentation: "Dokumentacja",
	KFrameTitlePreview: "Podgląd slajdu", KFrameTitleRail: "Miniatury slajdów",

	KDocGettingStarted: "Pierwsze kroki", KDocMarp: "Zgodność z Marp", KDocTemplates: "Szablony",
	KDocKinds: "Kinds", KDocHCL: "HCL", KDocCLI: "Wiersz poleceń", KDocVSCode: "VS Code", KDocPlayground: "Playground",

	KTableColumn: "Kolumna %d",

	KErrExists: "%s już istnieje", KErrDeckName: "prezentacja pozostaje deck.md: polecenie slides i playground odczytują ją pod tą nazwą",
	KErrNoSuchFile: "%s: nie ma takiego pliku", KErrNotFileName: "%q nie jest nazwą pliku",
	KErrDeckTooLarge: "prezentacja ma %d KiB; playground renderuje prezentacje do %d KiB",
	KErrMaxFiles:     "projekt zawiera co najwyżej %d plików", KErrPathReason: "%s: %v",
	KErrProjectOver: "projekt przekroczyłby %d MiB", KErrFileSize: "%s ma %d KiB: plik ma co najwyżej %d KiB",
	KErrShareTooLargeDeck: "prezentacja jest większa, niż playground przyjmuje; pobierz ją zamiast tego",
	KErrShareTooLargeLink: "prezentacja jest zbyt duża, aby podróżować w łączu; pobierz ją zamiast tego",
	KErrShareNoDeck:       "łącze nie zawiera prezentacji", KErrShareTooLong: "prezentacja z łącza jest zbyt długa",
	KErrShareDamaged: "prezentacja z łącza jest uszkodzona (mogła zostać ucięta)",
	KErrShareLarger:  "prezentacja z łącza jest większa, niż playground przyjmuje", KErrShareNotText: "prezentacja z łącza nie jest tekstem",
	KErrZipNoDeck: "archiwum nie zawiera prezentacji: umieść deck.md na jego najwyższym poziomie (lub jeden plik .md)",
	KErrZipSize:   "archiwum ma %d MiB: co najwyżej %d MiB", KErrZipNot: "to nie jest archiwum .zip: %v",
	KErrZipEntries: "archiwum zawiera %d wpisów: co najwyżej %d", KErrZipLinks: "%s: łącza symboliczne nie są przyjmowane",
	KErrZipFileOver: "%s przekracza %d KiB: plik ma co najwyżej tyle",
	KErrZipBomb:     "archiwum rozpakowuje się ponad %d razy: odrzucone jako możliwa bomba zip",
	KErrZipTwice:    "%s występuje w archiwum dwa razy", KErrZipMoreThan: "ponad %d KiB",
	KErrFileOver: "%s przekracza %s", KErrFailed: "niepowodzenie",
	KErrFolderCount: "folder zawiera %d plików: projekt zawiera co najwyżej %d", KErrFolderOver: "folder przekracza %s",
	KErrFolderMore: "folder zawiera ponad %d plików", KErrNotImage: "%s nie jest obrazem.",

	"fmt.label.bold": "Pogrubienie", "fmt.label.italic": "Kursywa", "fmt.label.underline": "Podkreślenie",
	"fmt.label.strike": "Przekreślenie", "fmt.label.code": "Kod w wierszu",
	"fmt.label.h1": "Nagłówek 1", "fmt.label.h2": "Nagłówek 2", "fmt.label.h3": "Nagłówek 3",
	"fmt.label.bullets": "Punktory", "fmt.label.numbered": "Numeracja", "fmt.label.quote": "Cytat", "fmt.label.rule": "Linia",
	"fmt.label.table": "Tabela", "fmt.label.table-format": "Wyrównaj",
	"fmt.label.row-above": "Wiersz nad", "fmt.label.row-below": "Wiersz pod", "fmt.label.row-delete": "Usuń wiersz",
	"fmt.label.col-left": "Kolumna z lewej", "fmt.label.col-right": "Kolumna z prawej", "fmt.label.col-delete": "Usuń kolumnę",
	"fmt.label.colalign-left": "Kolumna do lewej", "fmt.label.colalign-center": "Kolumna wyśrodk.",
	"fmt.label.colalign-right": "Kolumna do prawej",
	"fmt.label.slide-left":     "Tekst do lewej", "fmt.label.slide-center": "Tekst wyśrodk.", "fmt.label.slide-right": "Tekst do prawej",
	"fmt.label.link": "Łącze",

	"fmt.tip.bold":            "Pogrubia zaznaczony tekst lub słowo przy kursorze: **tekst**. Naciśnij ponownie, aby cofnąć.",
	"fmt.tip.italic":          "Kursywa: _tekst_. Naciśnij ponownie, aby cofnąć.",
	"fmt.tip.underline":       "Markdown nie ma podkreślenia: to ujmuje tekst w <u>…</u>, znacznik, który domyślna lista dozwolonego HTML w Marp zachowuje, więc działa przy wyłączonym surowym HTML. Naciśnij ponownie, aby cofnąć.",
	"fmt.tip.strike":          "Przekreślenie: ~~tekst~~. Naciśnij ponownie, aby cofnąć.",
	"fmt.tip.code":            "Kod w wierszu: `tekst`. Naciśnij ponownie, aby cofnąć.",
	"fmt.tip.h1":              "Zamienia wiersz (lub każdy zaznaczony wiersz) w nagłówek 1: # tekst. Naciśnij ponownie, aby wrócić do zwykłego tekstu.",
	"fmt.tip.h2":              "Nagłówek 2: ## tekst.",
	"fmt.tip.h3":              "Nagłówek 3: ### tekst.",
	"fmt.tip.bullets":         "Lista punktowana: - element, po jednym w każdym zaznaczonym wierszu. Naciśnij ponownie, aby cofnąć.",
	"fmt.tip.numbered":        "Lista numerowana: 1. element, 2. element… Naciśnij ponownie, aby cofnąć.",
	"fmt.tip.quote":           "Cytat: > tekst. Naciśnij ponownie, aby zdjąć jeden poziom. Ctrl+Shift+.",
	"fmt.tip.rule":            "Linia pozioma. Marp odczytuje ---, *** i ___ jako nowy slajd, więc to zapisuje znacznik <hr>, który domyślna lista dozwolonego HTML zachowuje.",
	"fmt.tip.table":           "Wstawia tabelę po bieżącym wierszu: wybierz liczbę kolumn i wierszy.",
	"fmt.tip.table-format":    "Wyrównuje tabelę wokół kursora: każda kolumna jest dopełniona do najszerszej komórki, znaczniki wyrównania zostają.",
	"fmt.tip.row-above":       "Dodaje pusty wiersz nad wierszem kursora (pod nagłówkiem, gdy kursor jest w nagłówku).",
	"fmt.tip.row-below":       "Dodaje pusty wiersz pod wierszem kursora.",
	"fmt.tip.row-delete":      "Usuwa wiersz kursora; nagłówek zostaje.",
	"fmt.tip.col-left":        "Dodaje kolumnę z lewej strony kolumny kursora.",
	"fmt.tip.col-right":       "Dodaje kolumnę z prawej strony kolumny kursora.",
	"fmt.tip.col-delete":      "Usuwa kolumnę kursora; tabela zachowuje jedną.",
	"fmt.tip.colalign-left":   "Wyrównuje kolumnę tabeli przy kursorze do lewej (:--). Naciśnij ponownie, aby usunąć.",
	"fmt.tip.colalign-center": "Wyśrodkowuje kolumnę tabeli przy kursorze (:-:). Naciśnij ponownie, aby usunąć.",
	"fmt.tip.colalign-right":  "Wyrównuje kolumnę tabeli przy kursorze do prawej (--:). Naciśnij ponownie, aby usunąć.",
	"fmt.tip.slide-left":      "Przywraca tekst slajdu do lewej, jak w motywie: usuwa blok stylu zapisany przez tę kartę.",
	"fmt.tip.slide-center":    "Wyśrodkowuje tekst slajdu, w którym jest kursor, za pomocą bloku <style scoped> na początku slajdu. Działa przy wyłączonym surowym HTML i jest widoczne w podglądzie, na stronie HTML i w pliku PDF.",
	"fmt.tip.slide-right":     "Wyrównuje do prawej tekst slajdu, w którym jest kursor (blok <style scoped>).",
	"fmt.tip.link":            "Zamienia zaznaczenie lub słowo w łącze: [tekst](url) z zaznaczonym url do wpisania. W łączu zaznacza jego url.",

	KGrpHistory: "Historia", KLblUndo: "Cofnij", KLblRedo: "Ponów",
	KTipUndo:     "Cofa ostatnią zmianę: edycję w pokazanym pliku, a jeśli jej nie ma, operację na całym projekcie.",
	KTipUndoStep: "Cofnij: %s.",
	KTipRedo:     "Robi jeszcze raz to, co cofnęło Cofnij.",
	KTipRedoStep: "Ponów: %s.",
	KStepExample: "przykład „%s”", KStepReset: "Resetuj", KStepOpenZip: "Otwórz plik .zip", KStepOpenFolder: "Otwórz folder",
	KStepRename: "Zmień nazwę %s", KStepDelete: "Usuń %s", KStepRestoreDraft: "Przywróć wersję roboczą", KStepRestoreVersion: "Przywróć %s",
	KStatusUndone: "Cofnięto: %s.", KStatusRedone: "Ponowiono: %s.",

	KLblSaveVersion: "Zapisz wersję…", KLblVersions: "Wersje",
	KTipSaveVersion: "Zachowuje cały projekt w obecnym stanie, pod nazwą, w tej przeglądarce.",
	KTipVersions:    "Wersje projektu, przechowywane w tej przeglądarce: obejrzyj jedną w podglądzie, porównaj ją z teraźniejszością, przywróć ją.",
	KAskVersionName: "Nazwa tej wersji:",
	KVersionAuto:    "Automatyczna", KVersionBefore: "Przed: %s", KVersionExport: "Wyeksportowana: %s", KVersionShared: "Udostępnione łącze",
	KVersionLine:   "%s · %s · %s",
	KVersionsEmpty: "Nie ma jeszcze żadnej wersji. Jedna jest zachowywana chwilę po tym, jak przestaniesz pisać, przed operacją na całym projekcie, przy każdym eksporcie i udostępnieniu oraz gdy zapiszesz ją sam.",
	KVersionsOff:   "Wersje nie są zachowywane: ta przeglądarka nie udostępnia stronie pamięci.",
	KLblOnlyNamed:  "Tylko nazwane", KLblVersionPreview: "Podgląd", KLblVersionCompare: "Porównaj z teraz",
	KLblVersionRestore: "Przywróć", KLblVersionDelete: "Usuń", KLblBackToNow: "Wróć do teraz", KLblSideBySide: "Obok siebie",
	KCompareTitle:    "%s w porównaniu z teraz",
	KCompareFileLine: "%s (+%d −%d)",
	KCompareSame:     "Brak różnic w stosunku do teraz.",
	KCompareBinary:   "%s nie jest tekstem: wtedy %s, teraz %s.",
	KCompareNow:      "Teraz",

	KStatusVersionSaved:    "Zapisano wersję „%s”.",
	KStatusVersionRestored: "Przywrócono %s; Cofnij to cofa.",
	KStatusVersionDeleted:  "Usunięto wersję %s.",
	KStatusPreviewing:      "Podgląd pokazuje %s; Wróć do teraz lub edycja znów pokaże prezentację.",
	KStatusVersionsFull:    "Pamięć tej strony jest pełna: najstarsze wersje automatyczne zostały usunięte.",
	KStatusVersionsFailed:  "Nie udało się zachować wersji: %s",
	KStatusVersionLimit:    "Zachowywanych jest co najwyżej %d nazwanych wersji: usuń jedną, aby zapisać kolejną.",
	KConfirmDeleteVersion:  "Usunąć wersję %s? Tego nie można cofnąć.",

	KHelpGeneral: `Wpisuj prezentację po lewej; slajdy pojawiają się po prawej.

- Slajdy oddziela wiersz z trzema myślnikami: ---
- Front matter, między dwoma pierwszymi wierszami ---, zawiera dyrektywy globalne (theme, paginate, header, footer, size) i własne klucze go-slides: kinds, data, template.
- <!-- _class: lead --> w slajdzie to dyrektywa tylko dla tego slajdu.
- {{ .title }}, {{ page }}/{{ total }}, {{ range seq 3 }}: szablony Go.
- kind: title stosuje kind (czyli nazwany zestaw dyrektyw).
- Matematyka: $e^{i\pi}$ w wierszu, $$ ... $$ w osobnych wierszach.
- Obrazy: ![w:300](url) zmienia rozmiar obrazu, ![sepia](url) nakłada filtr; ![bg](url) to tło, ![bg left:40%](url) tło dzielone obok tekstu.

Pliki, po lewej od edytora, to projekt prezentacji, tak jak czyta go polecenie slides: deck.md, slides.yaml, themes/, kinds/, media/. Kliknij plik, aby go edytować. Kliknij prawym przyciskiem (lub Shift+F10), aby użyć poleceń Nowy plik, Nowy folder, Zmień nazwę (F2), Usuń, Nowy motyw, Nowy kind; przeciągnij plik na folder, aby go przenieść. Menu Projekt tworzy motyw na bazie default, gaia lub uncover (dołącza do listy Motyw), kind, plik danych oraz otwiera lub pobiera cały projekt jako plik .zip (po rozpakowaniu buduje go slides build deck.md).

Panele: przeciągnij pasek między dwoma panelami, aby zmienić ich rozmiar (Pliki, edytor, Slajdy, podgląd, Pomoc); kliknij pasek dwukrotnie lub naciśnij na nim Enter, aby zwinąć panel i go przywrócić; F6 przechodzi między paskami, strzałki przesuwają pasek o 16 px, Home i End przenoszą do najmniejszego i największego rozmiaru, Escape wraca do edytora. Ustawione rozmiary są zapamiętywane w tej przeglądarce; karta Widok nadal pozwala ukrywać panele.

Karta Format: Pogrubienie, Kursywa, Przekreślenie i Kod w wierszu działają na zaznaczeniu lub słowie przy kursorze i przełączają się (Ctrl+B, Ctrl+I, Ctrl+Shift+X, Ctrl+E; Cmd na Macu): każdy przycisk jest wciśnięty wewnątrz swoich znaczników, a jedno Cofnij cofa jedną czynność. Podkreślenie (Ctrl+U) ujmuje tekst w <u>…</u>, ponieważ Markdown nie ma podkreślenia, a domyślna lista HTML w Marp zachowuje ten znacznik. Nagłówki od 1 do 3 (Ctrl+Alt+1..3), punktory (Ctrl+Shift+8), numeracja (Ctrl+Shift+7) i cytat działają na zaznaczonych wierszach; linia pozioma to znacznik <hr>, ponieważ ---, *** i ___ rozpoczęłyby nowy slajd. Tabela otwiera siatkę kolumn i wierszy; przy kursorze w tabeli dodajesz lub usuwasz wiersz albo kolumnę, wyrównujesz kolumnę (:--, :-:, --:), a Wyrównaj utrzymuje komórki w jednej linii. Tekst slajdu do lewej, wyśrodkowany lub do prawej zapisuje blok <style scoped> w bieżącym slajdzie: działa przy wyłączonym surowym HTML i jest widoczny w podglądzie, na stronie HTML i w pliku PDF. Łącze (Ctrl+K) zapisuje [tekst](url) z zaznaczonym url.

Upuść lub wklej obraz albo wideo na edytor: trafia do media/, a prezentacja go nazywa, ![](media/photo.png); upuść pliki na folder w panelu, aby umieścić je tam. Plik ma co najwyżej 8 MiB, a projekt 64 MiB (wiersz pod plikami pokazuje, ile jest zajęte). Nic nie jest przesyłane. Obraz https: też działa, ale przeglądarka pobiera go z jego serwera, który widzi wtedy Twój adres (nie jest wysyłany referer).

Motyw wybiera default, gaia, uncover lub motyw zdefiniowany w prezentacji (/* @theme name */ w <style>): zapisuje theme: we front matter i podąża za front matter, gdy go edytujesz.

Uzupełnianie otwiera się podczas pisania dyrektywy, kind lub wewnątrz {{ }}. Umieść kursor na dyrektywie, kind lub funkcji szablonu, a ten panel ją wyjaśni. Kliknij slajd w przeglądzie, aby przejść do jego wiersza.

Pokaz (lub p w pokazie slajdów) otwiera widok prezentera w drugim oknie: bieżący slajd, następny krok, notatki prelegenta (<!-- komentarze -->), zegar i stoper. Strzałki, Spacja, Home i End poruszają oba okna; pokaz slajdów i prezenter podążają za sobą, także po zmianach w edycji. Zezwól na wyskakujące okna dla tej witryny, jeśli okno się nie otwiera.

Twoja prezentacja zostaje w tej przeglądarce: jest zapisywana podczas pisania we własnym magazynie przeglądarki, tylko dla tej witryny, i wraca po odświeżeniu; nic nie jest nigdzie wysyłane. Wersja robocza przywraca wersję zastąpioną przez udostępnione łącze, przykład lub inną kartę, resetuje do przykładu, a Zapomnij moje wersje robocze je usuwa. Udostępnij umieszcza prezentację w samym łączu (bez jej pozostałych plików: pobierz plik .zip projektu, aby je mieć).

Cofnij i Ponów (karta Start; Ctrl+Z oraz Ctrl+Shift+Z lub Ctrl+Y; Cmd na Macu) cofają ostatnią edycję pokazanego pliku, przy czym każdy plik ma własną, a potem ostatnią operację na całym projekcie: przykład, Resetuj, otwarty plik .zip lub folder, przywróconą wersję roboczą, usunięty lub przemianowany plik, przywróconą wersję. Ponów przywraca wszystko tak, jak było w chwili, gdy nacisnąłeś Cofnij.

Wersje zachowują cały projekt w tej przeglądarce: jedna jest zapisywana chwilę po tym, jak przestaniesz pisać, przed operacją na całym projekcie, przy każdym eksporcie i udostępnieniu, a Zapisz wersję… zachowuje jedną pod nazwą. Panel Wersje je wylicza: Podgląd pokazuje jedną w podglądzie, Porównaj z teraz pokazuje zmiany każdego pliku wiersz po wierszu (F7 przechodzi do następnej zmiany), Przywróć wstawia ją z powrotem, a Cofnij cofa przywrócenie. Zachowywane są najnowsze wersje automatyczne, potem jedna na godzinę przez dobę i jedna na dzień przez miesiąc; nazwane zostają, dopóki ich nie usuniesz.

Karty Wstążki: Start (Cofnij, Ponów, wersje, przykład, motyw, Pokaz, Udostępnij, Wersja robocza), Wstawianie (nowy slajd, obraz lub tło z pliku, matematyka, blok kodu z numerami wierszy, kind: każde jest wpisywane do prezentacji przy kursorze, wokół zaznaczenia, jeśli jest, a Cofnij je cofa), Widok (podgląd, minimapa, miniatury slajdów) i Eksport. Alt pokazuje literę nad każdą kartą; wpisz ją, a potem literę nad kontrolką. Ctrl+F1 zwija Wstążkę.

Karta Eksport: strona HTML prezentacji z multimediami w środku; plik PDF (wektorowy, z zaznaczalnym tekstem), prezentacja PowerPoint lub OpenDocument (edytowalne), tworzone w Twojej przeglądarce dokładnie tak, jak tworzy je polecenie slides. Pierwszy eksport pobiera silnik eksportu (około 5 MiB, jednorazowo: przeglądarka go zachowuje), kompilując go w miarę nadchodzenia, a potem tylko czcionki używane przez Twoje slajdy; obrazy https: są pobierane tylko wtedy, gdy ich serwer na to pozwala (CORS), obrazy data: zawsze.`,
}

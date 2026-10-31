package playground

// frMessages is the French catalogue.
var frMessages = map[Key]string{
	KTabHome: "Accueil", KTabInsert: "Insertion", KTabFormat: "Format", KTabView: "Affichage", KTabExport: "Exporter",

	KGrpDeck: "Présentation", KGrpLook: "Apparence", KGrpPresent: "Présenter", KGrpDraft: "Brouillon", KGrpSlides: "Diapositives",
	KGrpPictures: "Images", KGrpText: "Texte", KGrpPreview: "Aperçu", KGrpEditor: "Éditeur", KGrpSaveAs: "Enregistrer sous",
	KGrpParagraph: "Paragraphe", KGrpTable: "Tableau", KGrpRowsCols: "Lignes et colonnes", KGrpColumn: "Colonne",
	KGrpSlideText: "Texte de la diapo", KGrpLink: "Lien",

	KLblExample: "Exemple", KLblReset: "Réinitialiser", KLblTheme: "Thème", KLblThemeOption: "Thème : %s", KLblDraft: "Brouillon",
	KLblShare: "Partager le lien", KLblPresent: "Présenter", KLblPreview: "Aperçu", KLblMinimap: "Minicarte",
	KLblSlides: "Diapositives", KLblFiles: "Fichiers", KLblHelp: "Aide", KLblNewSlide: "Nouvelle diapositive", KLblImage: "Image",
	KLblBackground: "Arrière-plan", KLblBgLeft: "Fond de la moitié gauche", KLblBgRight: "Fond de la moitié droite",
	KLblMath: "Formule", KLblCode: "Bloc de code", KLblKind: "Kind", KLblHTML: "Page HTML", KLblPDF: "PDF",
	KLblPPTX: "PowerPoint", KLblODP: "OpenDocument",

	KTipExample:    "Partir d’un des exemples. Votre texte reste dans ce navigateur : Brouillon ▸ Restaurer le brouillon remplacé le ramène.",
	KTipTheme:      "Le thème de la présentation : en choisir un l’écrit dans l’en-tête (front matter) de la présentation.",
	KTipDraft:      "Restaurer, oublier ou réinitialiser la présentation que ce navigateur conserve pour vous.",
	KTipReset:      "Remettre le texte de l’exemple ; votre présentation est conservée comme brouillon.",
	KTipPresent:    "Afficher le diaporama ici et ouvrir la vue de l’orateur dans sa propre fenêtre.",
	KTipShare:      "Copier un lien qui contient la présentation elle-même : rien n’est envoyé sur un serveur.",
	KTipKind:       "Donner un kind à la diapositive : un modèle avec son propre style, défini par un commentaire <!-- _kind: … -->.",
	KTipNewSlide:   "Ajouter une diapositive vide après la diapositive courante.",
	KTipImage:      "Choisir un fichier image : il entre dans la présentation sous forme d’URI data:, rien n’est envoyé. Le texte sélectionné devient sa description.",
	KTipBackground: "Choisir une image pour l’arrière-plan de la diapositive ; la flèche ne remplit que les 40 % de gauche ou de droite.",
	KTipMath:       "Une formule en mode affiché entre deux lignes $$ ; la sélection est placée à l’intérieur.",
	KTipCode:       "Un bloc de code délimité avec numéros de ligne ({lineNumbers}) ; la sélection est placée à l’intérieur.",
	KTipMode:       "Aperçu empile toutes les diapositives et défile avec l’éditeur ; Diaporama est la présentation elle-même.",
	KTipMinimap:    "Une carte de la présentation à côté de la barre de défilement de l’éditeur.",
	KTipThumbs:     "Une colonne de miniatures de diapositives ; cliquez sur l’une d’elles pour y aller.",
	KTipFiles:      "Les fichiers du projet : thèmes, kinds, données et médias. Une fenêtre étroite démarre sans ce volet.",
	KTipHelp:       "L’aide sur l’élément à la position du curseur, et les liens vers la documentation. Une fenêtre étroite démarre sans ce volet.",
	KTipHTML:       "Un seul fichier HTML qui présente la présentation dans n’importe quel navigateur.",
	KTipPDF:        "Un PDF, une page par diapositive, produit ici dans votre navigateur.",
	KTipPPTX:       "Un fichier PowerPoint modifiable, produit ici dans votre navigateur.",
	KTipODP:        "Une présentation OpenDocument modifiable, produite ici dans votre navigateur.",

	KModeOverview: "Aperçu", KModeSlideshow: "Diaporama",

	KExWelcome: "Bienvenue", KExTemplates: "Modèles : diapositives générées", KExKinds: "Kinds de diapositives",
	KExHCL: "En-tête HCL", KExGaia: "Thèmes et directives locales", KExMedia: "Formules, images et arrière-plans",
	KExGalleryDefault: "Galerie : le thème default", KExGalleryGaia: "Galerie : le thème gaia",
	KExGalleryUncover: "Galerie : le thème uncover", KExGalleryMath: "Galerie : formules mathématiques",
	KExGalleryCode:       "Galerie : code, numéros de ligne et encadrés",
	KExGalleryEmoji:      "Galerie : émojis et titres ajustés",
	KExGalleryPresenting: "Galerie : transitions et notes de l’orateur",
	KExGalleryHCL:        "Galerie : expressions HCL et données typées",

	KStatusStarting:           "Démarrage…",
	KStatusOpenedShared:       "Présentation partagée dans le lien ouverte.",
	KStatusSharedUnreadable:   "La présentation du lien n’a pas pu être lue : %s",
	"status.rendered.one":     "%d diapositive · rendue en %d ms",
	"status.rendered.other":   "%d diapositives · rendues en %d ms",
	KStatusStale:              "l’aperçu affiche la dernière présentation rendue",
	"status.problems.one":     "%d problème · %s",
	"status.problems.other":   "%d problèmes · %s",
	KStatusDraftNotKept:       " · ⚠ brouillon non conservé dans ce navigateur",
	KStatusPresenting:         "Présentation en cours : la vue de l’orateur est dans sa propre fenêtre ; les touches fléchées dans l’une ou l’autre font avancer les deux.",
	KStatusLinkCopied:         "Lien copié : la présentation voyage dans le lien lui-même (%d caractères), rien n’est envoyé.",
	KStatusSavedHTML:          "deck.html enregistré : ouvrez-le dans n’importe quel navigateur, ou présentez depuis ce fichier.",
	KStatusShareDeckOnly:      " Seule la présentation est dans le lien : Projet ▸ Télécharger le projet (.zip) conserve ses fichiers.",
	KStatusRestoredDraft:      "Votre brouillon de %s a été restauré.",
	KStatusSharedKeptDraft:    "Présentation partagée dans le lien ouverte. Votre brouillon de %s est conservé : Brouillon ▸ Restaurer mon brouillon.",
	KStatusRestoredDraftReset: "Votre brouillon de %s a été restauré. Réinitialiser revient à l’exemple.",
	KStatusDraftSavedAgain:    "Votre brouillon est de nouveau enregistré dans ce navigateur.",
	KStatusDraftNotSaved:      "⚠ Ce navigateur n’a pas conservé votre brouillon (%s) : Partager le lien ou Télécharger le HTML conserve la présentation.",
	KStatusOtherTab:           "Un autre onglet a enregistré son brouillon %s. Vos modifications ici ne l’effaceront pas : Brouillon ▸ Restaurer le brouillon remplacé.",
	KStatusRestoredTheDraft:   "Le brouillon de %s a été restauré.",
	KStatusForgot:             "Vos brouillons ont été oubliés : rien de cette présentation n’est plus stocké dans ce navigateur (jusqu’à votre prochaine modification).",
	KStatusNoExport:           "⚠ Cette page ne peut pas exporter : utilisez la commande slides.",
	KStatusExportRunning:      "Un export est en cours : %s.",
	KStatusTooLargeExport:     "⚠ La présentation est trop volumineuse pour être exportée ici : utilisez la commande slides.",
	KStatusPreparing:          "Préparation de %s…",
	KStatusLoadingEnginePct:   "Chargement du moteur d’export (une seule fois ; votre navigateur le conserve) : %d %% de %.1f MiB",
	KStatusLoadingEngine:      "Chargement du moteur d’export (une seule fois ; votre navigateur le conserve) : %.1f MiB",
	KStatusExportProgress:     "%s : diapositive %d sur %d…",
	KStatusExportFailed:       "⚠ L’export %s a échoué : %s",
	"status.saved.one":        "%s enregistré : %s, %d diapositive en %.1f s, produit dans votre navigateur.",
	"status.saved.other":      "%s enregistré : %s, %d diapositives en %.1f s, produit dans votre navigateur.",
	KStatusIsMedium:           "%s est un média : la présentation l’affiche avec %s (déposez-le sur l’éditeur, ou via son menu : Insérer dans la présentation).",
	KStatusMadeTheme:          "%s créé, fondé sur %s : choisissez %s dans la liste Thème pour l’utiliser.",
	KStatusMadeKind:           "%s créé : une diapositive prend ce kind avec <!-- _kind: %s -->.",
	KStatusMadeData:           "%s créé : la présentation le lit sous la forme {{ .title }}.",
	KStatusCannotSaveFile:     "⚠ cette page ne peut pas enregistrer de fichier",
	KStatusSavedZip:           "%s enregistré : une fois décompressé, `slides build deck.md` le construit, et l’extension VS Code l’ouvre.",
	KStatusOpenedZip:          "Archive .zip ouverte : %d fichiers.",
	KStatusOpenedFolder:       "Dossier ouvert : %d fichiers.",
	KStatusAddedSlide:         "Diapositive ajoutée après la diapositive courante.",
	KStatusInsertedFormula:    "Formule en mode affiché insérée : saisissez à l’intérieur.",
	KStatusInsertedCode:       "Bloc de code Go avec numéros de ligne inséré : saisissez à l’intérieur.",
	KStatusPutFormula:         "Sélection placée dans une formule en mode affiché.",
	KStatusPutCode:            "Sélection placée dans un bloc de code Go avec numéros de ligne.",
	KStatusNoFilePicker:       "⚠ Cette page ne peut pas demander de fichier : déposez ou collez un fichier image dans l’éditeur.",
	KStatusKindSet:            "La diapositive est maintenant du kind %s.",
	KStatusInsertedRule:       "Filet horizontal inséré (***: trois tirets sur une ligne commenceraient une nouvelle diapositive).",
	KStatusInsertedTable:      "Tableau inséré : saisissez dans ses cellules ; les commandes d’alignement du tableau le gardent propre.",
	KStatusJournalLarge:       "Ce projet est trop volumineux pour l’enregistrement rapide : les modifications de la dernière demi-seconde avant la fermeture de la page peuvent ne pas être conservées.",
	KStatusJournalRefused:     "⚠ L’enregistrement rapide est refusé (%s) : les modifications de la dernière demi-seconde avant la fermeture de la page peuvent ne pas être conservées.",
	KStatusAdded:              "%s ajouté au projet (%d KiB) : il est conservé dans ce navigateur et figure dans le .zip du projet.",
	KStatusAddedPut:           "%s ajouté au projet (%d KiB) et placé dans la présentation : il est conservé dans ce navigateur et figure dans le .zip du projet.",
	KStatusPopupBlocked:       "⚠ Le navigateur a bloqué la fenêtre de l’orateur : autorisez les fenêtres contextuelles pour cette page, puis Présentez à nouveau.",

	KWhyHeading:    "Un titre se place sur une ligne du texte de la présentation, pas dans l’en-tête ni dans un bloc de code.",
	KWhyList:       "Une liste se place sur une ligne du texte de la présentation, pas dans l’en-tête ni dans un bloc de code.",
	KWhyQuote:      "Une citation se place sur une ligne du texte de la présentation, pas dans l’en-tête ni dans un bloc de code.",
	KWhyNotInTable: "Placez d’abord le curseur dans un tableau.",
	KWhyNotInBody:  "Placez le curseur dans une ligne du corps du tableau : l’en-tête reste.",
	KWhyLastColumn: "Un tableau garde au moins une colonne.",

	KProblemLine: "ligne %d", KProblemError: "erreur", KProblemWarning: "avertissement",

	KConfirmReset:      "Remplacer votre présentation par l’exemple ? Votre texte reste dans ce navigateur : Brouillon ▸ Restaurer le brouillon remplacé.",
	KConfirmShareFiles: "Un lien ne contient que la présentation, pas ses autres fichiers (thèmes, kinds, médias) : télécharger plutôt le projet sous forme de .zip ?",
	KConfirmOpenZip:    "Remplacer votre présentation et ses fichiers par ceux du .zip ? Ce que vous avez maintenant reste dans ce navigateur : Brouillon ▸ Restaurer le brouillon remplacé.",
	KConfirmOpenFolder: "Remplacer votre présentation et ses fichiers par ceux du dossier ? Ce que vous avez maintenant reste dans ce navigateur : Brouillon ▸ Restaurer le brouillon remplacé.",

	KDateToday: "aujourd’hui à %s", KDateAt: "%s à %s", KDateShort: "%[2]d %[1]s",
	KDateMonths: "janv.,févr.,mars,avr.,mai,juin,juil.,août,sept.,oct.,nov.,déc.",
	KBytes:      "%d octets", KWarn: "⚠ %s",

	KDraftRestoreMine: "Restaurer mon brouillon (%s)", KDraftRestoreReplaced: "Restaurer le brouillon remplacé (%s)",
	KDraftForget: "Oublier mes brouillons", KDraftResetExample: "Revenir à l’exemple", KDraftNothing: "Rien d’enregistré pour l’instant",

	KSaveAs: "Enregistrer sous…", KKindTitle: "Kind…",

	KProject: "Projet", KFilesTreeLabel: "Fichiers du projet", KFilesNewThemeFrom: "Nouveau thème (d’après %s)",
	KFilesNewKind: "Nouveau kind", KFilesNewData: "Nouveau fichier de données", KFilesNewFile: "Nouveau fichier…", KFilesNewFolder: "Nouveau dossier…",
	KFilesAddMedia: "Ajouter un média…", KFilesOpenZip: "Ouvrir un projet (.zip)…", KFilesOpenFolder: "Ouvrir un dossier de projet…",
	KFilesDownloadZip: "Télécharger le projet (.zip)", KFilesInsertIntoDeck: "Insérer dans la présentation", KFilesDownload: "Télécharger",
	"files.usage.one": "%d fichier · %s sur %d MiB", "files.usage.other": "%d fichiers · %s sur %d MiB",

	KPaneProblems: "Problèmes", KPaneDocumentation: "Documentation",
	KFrameTitlePreview: "Aperçu de la diapositive", KFrameTitleRail: "Miniatures des diapositives",

	KDocGettingStarted: "Prise en main", KDocMarp: "Compatibilité Marp", KDocTemplates: "Modèles",
	KDocKinds: "Kinds", KDocHCL: "HCL", KDocCLI: "Ligne de commande", KDocVSCode: "VS Code", KDocPlayground: "Playground",

	KTableColumn: "Colonne %d",

	KErrExists: "%s existe déjà", KErrDeckName: "la présentation reste deck.md : la commande slides et le Playground la lisent sous ce nom",
	KErrNoSuchFile: "%s : fichier introuvable", KErrNotFileName: "%q n’est pas un nom de fichier",
	KErrDeckTooLarge: "la présentation fait %d KiB ; le Playground traite des présentations jusqu’à %d KiB",
	KErrMaxFiles:     "un projet contient au plus %d fichiers", KErrPathReason: "%s : %v",
	KErrProjectOver: "le projet dépasserait %d MiB", KErrFileSize: "%s fait %d KiB : un fichier fait au plus %d KiB",
	KErrShareTooLargeDeck: "la présentation dépasse ce que le Playground accepte ; téléchargez-la plutôt",
	KErrShareTooLargeLink: "la présentation est trop volumineuse pour voyager dans un lien ; téléchargez-la plutôt",
	KErrShareNoDeck:       "le lien ne contient pas de présentation", KErrShareTooLong: "la présentation du lien est trop longue",
	KErrShareDamaged: "la présentation du lien est endommagée (le lien a peut-être été coupé)",
	KErrShareLarger:  "la présentation du lien dépasse ce que le Playground accepte", KErrShareNotText: "la présentation du lien n’est pas du texte",
	KErrZipNoDeck: "l’archive ne contient pas de présentation : placez deck.md à sa racine (ou un seul fichier .md)",
	KErrZipSize:   "l’archive fait %d MiB : au plus %d MiB", KErrZipNot: "n’est pas une archive .zip : %v",
	KErrZipEntries: "l’archive compte %d entrées : au plus %d", KErrZipLinks: "%s : les liens ne sont pas acceptés",
	KErrZipFileOver: "%s dépasse %d KiB : un fichier ne peut pas dépasser cette taille",
	KErrZipBomb:     "l’archive se décompresse plus de %d fois : refusée comme bombe de décompression possible",
	KErrZipTwice:    "%s figure deux fois dans l’archive", KErrZipMoreThan: "plus de %d KiB",
	KErrFileOver: "%s dépasse %s", KErrFailed: "échec",
	KErrFolderCount: "le dossier contient %d fichiers : un projet en contient au plus %d", KErrFolderOver: "le dossier dépasse %s",
	KErrFolderMore: "le dossier contient plus de %d fichiers", KErrNotImage: "%s n’est pas une image.",

	"fmt.label.bold": "Gras", "fmt.label.italic": "Italique", "fmt.label.underline": "Souligné",
	"fmt.label.strike": "Barré", "fmt.label.code": "Code en ligne",
	"fmt.label.h1": "Titre 1", "fmt.label.h2": "Titre 2", "fmt.label.h3": "Titre 3",
	"fmt.label.bullets": "Puces", "fmt.label.numbered": "Numérotation", "fmt.label.quote": "Citation", "fmt.label.rule": "Filet",
	"fmt.label.table": "Tableau", "fmt.label.table-format": "Aligner",
	"fmt.label.row-above": "Ligne au-dessus", "fmt.label.row-below": "Ligne en dessous", "fmt.label.row-delete": "Suppr. la ligne",
	"fmt.label.col-left": "Colonne à gauche", "fmt.label.col-right": "Colonne à droite", "fmt.label.col-delete": "Suppr. la colonne",
	"fmt.label.colalign-left": "Colonne à gauche", "fmt.label.colalign-center": "Colonne centrée",
	"fmt.label.colalign-right": "Colonne à droite",
	"fmt.label.slide-left":     "Texte à gauche", "fmt.label.slide-center": "Texte centré", "fmt.label.slide-right": "Texte à droite",
	"fmt.label.link": "Lien",

	"fmt.tip.bold":            "Met en gras le texte sélectionné, ou le mot sous le curseur : **texte**. Appuyez de nouveau pour annuler.",
	"fmt.tip.italic":          "Italique : _texte_. Appuyez de nouveau pour annuler.",
	"fmt.tip.underline":       "Markdown n’a pas de soulignement : cette commande entoure le texte de <u>…</u>, une balise que la liste d’autorisation HTML par défaut de Marp conserve, donc cela fonctionne HTML brut désactivé. Appuyez de nouveau pour annuler.",
	"fmt.tip.strike":          "Barré : ~~texte~~. Appuyez de nouveau pour annuler.",
	"fmt.tip.code":            "Code en ligne : `texte`. Appuyez de nouveau pour annuler.",
	"fmt.tip.h1":              "Fait de la ligne (ou de chaque ligne sélectionnée) un titre de niveau 1 : # texte. Appuyez de nouveau pour revenir au texte simple.",
	"fmt.tip.h2":              "Titre de niveau 2 : ## texte.",
	"fmt.tip.h3":              "Titre de niveau 3 : ### texte.",
	"fmt.tip.bullets":         "Liste à puces : - élément, un par ligne sélectionnée. Appuyez de nouveau pour annuler.",
	"fmt.tip.numbered":        "Liste numérotée : 1. élément, 2. élément… Appuyez de nouveau pour annuler.",
	"fmt.tip.quote":           "Citation : > texte. Appuyez de nouveau pour retirer un niveau. Ctrl+Shift+.",
	"fmt.tip.rule":            "Un filet horizontal. Marp lit ---, *** et ___ comme une nouvelle diapositive ; cette commande écrit donc la balise <hr>, que la liste d’autorisation HTML par défaut conserve.",
	"fmt.tip.table":           "Insère un tableau après la ligne courante : choisissez ses colonnes et ses lignes.",
	"fmt.tip.table-format":    "Aligne le tableau autour du curseur : chaque colonne est complétée à la largeur de sa cellule la plus large, les marques d’alignement sont conservées.",
	"fmt.tip.row-above":       "Ajoute une ligne vide au-dessus de la ligne du curseur (sous l’en-tête depuis l’en-tête).",
	"fmt.tip.row-below":       "Ajoute une ligne vide sous la ligne du curseur.",
	"fmt.tip.row-delete":      "Supprime la ligne du curseur ; l’en-tête reste.",
	"fmt.tip.col-left":        "Ajoute une colonne à gauche de la colonne du curseur.",
	"fmt.tip.col-right":       "Ajoute une colonne à droite de la colonne du curseur.",
	"fmt.tip.col-delete":      "Supprime la colonne du curseur ; un tableau en garde toujours une.",
	"fmt.tip.colalign-left":   "Aligne à gauche la colonne du tableau sous le curseur (:--). Appuyez de nouveau pour l’effacer.",
	"fmt.tip.colalign-center": "Centre la colonne du tableau sous le curseur (:-:). Appuyez de nouveau pour l’effacer.",
	"fmt.tip.colalign-right":  "Aligne à droite la colonne du tableau sous le curseur (--:). Appuyez de nouveau pour l’effacer.",
	"fmt.tip.slide-left":      "Remet le texte de la diapositive à gauche, comme le thème : supprime le bloc de style écrit par cet onglet.",
	"fmt.tip.slide-center":    "Centre le texte de la diapositive où se trouve le curseur, avec un bloc <style scoped> en haut de la diapositive. Cela fonctionne HTML brut désactivé et s’affiche dans l’aperçu, la page HTML et le PDF.",
	"fmt.tip.slide-right":     "Aligne à droite le texte de la diapositive où se trouve le curseur (un bloc <style scoped>).",
	"fmt.tip.link":            "Transforme la sélection, ou le mot, en lien : [texte](url) avec l’url sélectionnée pour la saisie. Dans un lien, sélectionne son url.",

	KGrpHistory: "Historique", KLblUndo: "Annuler", KLblRedo: "Rétablir",
	KTipUndo:     "Reprend la dernière modification : une modification du fichier affiché, sinon une opération sur tout le projet.",
	KTipUndoStep: "Annuler : %s.",
	KTipRedo:     "Refait ce qu’Annuler a repris.",
	KTipRedoStep: "Rétablir : %s.",
	KStepExample: "l’exemple « %s »", KStepReset: "Réinitialiser", KStepOpenZip: "Ouvrir un .zip", KStepOpenFolder: "Ouvrir un dossier",
	KStepRename: "Renommer %s", KStepDelete: "Supprimer %s", KStepRestoreDraft: "Restaurer le brouillon", KStepRestoreVersion: "Restaurer %s",
	KStatusUndone: "Annulé : %s.", KStatusRedone: "Rétabli : %s.",

	KLblSaveVersion: "Enregistrer la version…", KLblVersions: "Versions",
	KTipSaveVersion: "Conserve tout le projet tel qu’il est maintenant, sous un nom, dans ce navigateur.",
	KTipVersions:    "Les versions du projet, conservées dans ce navigateur : en afficher une, la comparer à l’état actuel, la restaurer.",
	KAskVersionName: "Nom de cette version :",
	KVersionAuto:    "Automatique", KVersionBefore: "Avant : %s", KVersionExport: "Exportée : %s", KVersionShared: "Lien partagé",
	KVersionLine:   "%s · %s · %s",
	KVersionsEmpty: "Aucune version pour l’instant. Une version est conservée peu après que vous cessez de saisir, avant une opération sur tout le projet, à chaque export ou partage, et quand vous en enregistrez une.",
	KVersionsOff:   "Les versions ne sont pas conservées : ce navigateur ne garde aucun stockage pour la page.",
	KLblOnlyNamed:  "Nommées seulement", KLblVersionPreview: "Aperçu", KLblVersionCompare: "Comparer à maintenant",
	KLblVersionRestore: "Restaurer", KLblVersionDelete: "Supprimer", KLblBackToNow: "Revenir à maintenant", KLblSideBySide: "Côte à côte",
	KCompareTitle:    "%s comparée à maintenant",
	KCompareFileLine: "%s (+%d −%d)",
	KCompareSame:     "Aucune différence avec maintenant.",
	KCompareBinary:   "%s n’est pas du texte : %s avant, %s maintenant.",
	KCompareNow:      "Maintenant",

	KStatusVersionSaved:    "Version « %s » enregistrée.",
	KStatusVersionRestored: "%s restaurée ; Annuler la reprend.",
	KStatusVersionDeleted:  "Version %s supprimée.",
	KStatusPreviewing:      "L’aperçu affiche %s ; Revenir à maintenant, ou une modification, réaffiche la présentation.",
	KStatusVersionsFull:    "Le stockage de cette page est plein : les plus anciennes versions automatiques ont été supprimées.",
	KStatusVersionsFailed:  "Une version n’a pas pu être conservée : %s",
	KStatusVersionLimit:    "Au plus %d versions nommées sont conservées : supprimez-en une pour en enregistrer une autre.",
	KConfirmDeleteVersion:  "Supprimer la version %s ? Cette action est irréversible.",

	KHelpGeneral: `Saisissez une présentation à gauche ; les diapositives suivent à droite.

- Les diapositives sont séparées par une ligne de trois tirets : ---
- L’en-tête (front matter), entre les deux premières lignes ---, contient les directives globales (theme, paginate, header, footer, size) et les clés propres à go-slides : kinds, data, template.
- <!-- _class: lead --> dans une diapositive est une directive pour cette seule diapositive.
- {{ .title }}, {{ page }}/{{ total }}, {{ range seq 3 }} : modèles Go.
- kind: title applique un kind (un jeu nommé de directives).
- Maths : $e^{i\pi}$ en ligne, $$ ... $$ sur des lignes à part.
- Images : ![w:300](url) en fixe la taille, ![sepia](url) lui applique un filtre ; ![bg](url) est un arrière-plan, ![bg left:40%](url) un arrière-plan partagé à côté du texte.

Fichiers, à gauche de l’éditeur, est le projet de la présentation, tel que la commande slides le lit : deck.md, slides.yaml, themes/, kinds/, media/. Cliquez sur un fichier pour le modifier. Clic droit (ou Shift+F10) pour Nouveau fichier, Nouveau dossier, Renommer (F2), Supprimer, Nouveau thème, Nouveau kind ; glissez un fichier sur un dossier pour le déplacer. Le menu Projet crée un thème fondé sur default, gaia ou uncover (il rejoint la liste Thème), un kind, un fichier de données, et ouvre ou télécharge tout le projet sous forme de .zip (décompressé, slides build deck.md le construit).

Volets : faites glisser la barre entre deux volets pour les redimensionner (Fichiers, l’éditeur, Diapositives, l’aperçu, Aide) ; double-cliquez sur une barre, ou appuyez sur Entrée dessus, pour refermer le volet et le rouvrir ; F6 passe d’une barre à l’autre, les touches fléchées la déplacent de 16 px, Début et Fin vont au plus petit et au plus grand, Échap revient à l’éditeur. Les tailles que vous réglez sont conservées dans ce navigateur ; l’onglet Affichage permet toujours de désactiver des volets.

Onglet Format : Gras, Italique, Barré et Code en ligne agissent sur la sélection ou le mot sous le curseur et basculent (Ctrl+B, Ctrl+I, Ctrl+Shift+X, Ctrl+E ; Cmd sur un Mac) : chaque bouton apparaît enfoncé à l’intérieur de ses marques, et un seul Annuler défait une action. Souligné (Ctrl+U) entoure le texte de <u>…</u>, car Markdown n’a pas de soulignement et la liste HTML par défaut de Marp conserve cette balise. Les titres 1 à 3 (Ctrl+Alt+1..3), les puces (Ctrl+Shift+8), la numérotation (Ctrl+Shift+7) et la citation agissent sur les lignes sélectionnées ; le filet est une balise <hr>, car ---, *** et ___ commenceraient une nouvelle diapositive. Tableau ouvre une grille de colonnes par lignes ; avec le curseur dans un tableau, vous ajoutez ou supprimez une ligne ou une colonne, alignez la colonne (:--, :-:, --:) et Aligner garde les cellules en face les unes des autres. Texte à gauche, centré ou à droite écrit un bloc <style scoped> dans la diapositive courante : cela fonctionne HTML brut désactivé et s’affiche dans l’aperçu, la page HTML et le PDF. Lien (Ctrl+K) écrit [texte](url) avec l’url sélectionnée.

Déposez ou collez une image ou une vidéo dans l’éditeur : elle va dans media/ et la présentation la nomme, ![](media/photo.png) ; déposez des fichiers sur un dossier du panneau pour les y placer. Un fichier fait au plus 8 MiB et un projet 64 MiB (la ligne sous les fichiers indique la place utilisée). Rien n’est envoyé. Une image https: fonctionne aussi, mais votre navigateur la récupère auprès de son serveur, qui voit alors votre adresse (aucun référent n’est envoyé).

Thème propose default, gaia, uncover ou un thème défini par votre présentation (/* @theme name */ dans un <style>) : il écrit theme: dans l’en-tête, et suit l’en-tête quand vous le modifiez.

La complétion s’ouvre quand vous saisissez une directive, un kind, ou à l’intérieur de {{ }}. Placez le curseur sur une directive, un kind ou une fonction de modèle et ce volet l’explique. Cliquez sur une diapositive dans l’aperçu pour aller à sa ligne.

Présenter (ou p dans le diaporama) ouvre la vue de l’orateur dans une seconde fenêtre : la diapositive courante, l’étape suivante, les notes de l’orateur (<!-- commentaires -->), une horloge et un chronomètre. Les touches fléchées, Espace, Début et Fin font avancer les deux fenêtres ; le diaporama et la vue de l’orateur se suivent, y compris pendant les modifications. Autorisez les fenêtres contextuelles pour ce site si la fenêtre ne s’ouvre pas.

Votre présentation reste dans ce navigateur : elle est enregistrée au fil de la saisie dans le stockage propre à ce navigateur, pour ce site seulement, et revient quand vous rechargez la page ; rien n’est envoyé nulle part. Brouillon restaure le brouillon qu’un lien partagé, un exemple ou un autre onglet a remplacé, revient à l’exemple, et Oublier mes brouillons les efface. Partager le lien place la présentation dans le lien lui-même (pas ses autres fichiers : téléchargez le .zip du projet pour ceux-là).

Annuler et Rétablir (onglet Accueil ; Ctrl+Z, et Ctrl+Shift+Z ou Ctrl+Y ; Cmd sur un Mac) reprennent la dernière modification du fichier affiché, chaque fichier gardant la sienne, puis la dernière opération sur tout le projet : un exemple, Réinitialiser, un .zip ou un dossier ouvert, un brouillon restauré, un fichier supprimé ou renommé, une version restaurée. Rétablir ramène tout comme c’était quand vous avez appuyé sur Annuler.

Les versions conservent tout le projet dans ce navigateur : une est enregistrée peu après que vous cessez de saisir, avant une opération sur tout le projet, à chaque export et partage, et Enregistrer la version… en conserve une sous un nom. Le volet Versions les liste : Aperçu en affiche une dans l’aperçu, Comparer à maintenant montre les changements de chaque fichier ligne à ligne (F7 va au changement suivant), Restaurer la remet en place, et Annuler reprend la restauration. Les plus récentes versions automatiques sont conservées, puis une par heure pendant un jour et une par jour pendant un mois ; les versions nommées restent jusqu’à ce que vous les supprimiez.

Les onglets du ruban : Accueil (Annuler, Rétablir, versions, exemple, thème, Présenter, Partager le lien, Brouillon), Insertion (une nouvelle diapositive, une image ou un arrière-plan depuis un fichier, une formule, un bloc de code avec numéros de ligne, un kind : chacun est saisi dans la présentation à la position du curseur, autour de la sélection s’il y en a une, et Annuler le retire), Affichage (l’aperçu, la minicarte, les miniatures des diapositives) et Exporter. Alt affiche une lettre au-dessus de chaque onglet ; tapez-la, puis la lettre au-dessus d’une commande. Ctrl+F1 réduit le ruban.

Onglet Exporter : la page HTML de la présentation, médias inclus ; une présentation PDF (vectorielle, texte sélectionnable), PowerPoint ou OpenDocument (modifiable), produite dans votre navigateur exactement comme la commande slides les produit. Le premier export télécharge le moteur d’export (environ 5 MiB, une seule fois : votre navigateur le conserve), en le compilant à mesure qu’il arrive, puis seulement les polices que vos diapositives utilisent ; les images https: ne sont récupérées que si leur serveur l’autorise (CORS), les images data: toujours.`,
}

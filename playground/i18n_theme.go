package playground

import "maps"

// The theme editor's words, in the twelve languages. They are merged into the
// catalogues at start-up so that the editor's catalogue reads as one piece.
func init() {
	for l, m := range map[Lang]map[Key]string{
		LangEN: themeEN, LangFR: themeFR, LangDE: themeDE, LangES: themeES, LangIT: themeIT, LangPT: themePT,
		LangPL: themePL, LangUK: themeUK, LangRU: themeRU, LangZH: themeZH, LangKO: themeKO, LangJA: themeJA,
	} {
		maps.Copy(catalogs[l], m)
	}
}

var themeEN = map[Key]string{
	KLblThemeEditor: "Theme editor",
	KTipThemeEditor: "Edit the deck's theme with controls: colours, fonts, logo, page number. The preview follows; the theme is kept as a package you can export.",

	KThemeSecPalette: "Palette", KThemeSecType: "Typography", KThemeSecLogo: "Logo", KThemeSecFooter: "Page number and footer",
	KThemeSecMeta: "Details", KThemeSecContrast: "Contrast (WCAG)",

	KThemePBackground: "Background", KThemePText: "Text", KThemePHeading: "Headings", KThemePAccent: "Accent", KThemePLink: "Links",
	KThemePFontBody: "Body font", KThemePFontHeading: "Heading font", KThemePFontCode: "Code font",
	KThemePSizeBody: "Body size", KThemePSizeHeading: "Heading size", KThemePSizeCode: "Code size",
	KThemePWeightBody: "Body weight", KThemePWeightHeading: "Heading weight",
	KThemePLogoPlace: "Placement", KThemePLogoSize: "Logo size", KThemePPagePlace: "Page number", KThemePPageSize: "Page number size",
	KThemePFooterAlign: "Footer alignment", KThemePFooterSize: "Footer size",

	KThemeMName: "Name", KThemeMVersion: "Version", KThemeMDescription: "Description", KThemeMAuthors: "Authors (separated by ;)",
	KThemeMLicense: "License (SPDX)", KThemeMHomepage: "Homepage", KThemeMExtends: "Based on", KThemeMMinSlides: "Needs go-slides",
	KThemeMFontLicense: "Font license: %s",

	KThemeBtnNew: "New theme", KThemeBtnImport: "Import…", KThemeBtnPackage: "Export .gsta.xz", KThemeBtnCSS: "Export CSS",
	KThemeBtnVersion: "Save version", KThemeBtnFont: "Add font…", KThemeBtnLogo: "Logo…", KThemeBtnNoLogo: "Remove logo",
	KThemeBtnOpenCSS: "Edit CSS", KThemeBtnRestore: "Restore the block",

	KThemeContrastPair: "%s on %s", KThemeContrastFail: "Fail", KThemeContrastNeed: "Set both colours to check",

	KThemeNoteImport:   "The sheet's @import of “%s” became the theme's base (extends): a package names its base there.",
	KThemeNoteDiverged: "The block between the gsta:visual comments was edited by hand, so the controls do not write to it. “Restore the block” rewrites it from the values it holds; your CSS outside the block is never touched.",
	KThemeNoteNone:     "No theme is being edited. Make a new one, import a package or a style sheet, or pick one of the project's themes.",

	KStatusThemeMade:        "Made the theme “%s” on “%s”: the deck uses it.",
	KStatusThemeOpened:      "Opened %s as the theme “%s”.",
	KStatusThemeExported:    "Exported %s. It is checked by the same code as slides theme verify.",
	KStatusThemeCSSExported: "Exported %s.",
	KStatusThemeSaved:       "Kept the version “%s”.",

	KStepThemeEdit: "Edit the theme: %s", KStepThemeNew: "New theme “%s”", KStepThemeFont: "Add the font %s", KStepThemeLogo: "Set the logo",

	KErrThemeDiverged: "the CSS block between the gsta:visual comments was edited by hand: restore it first",
	KErrThemeValue:    "“%s” is not a value of %s",
	KErrThemeNotSheet: "%s is not a style sheet (.css) or a theme package (.gsta.xz)",

	KThemeOptNone: "none", KThemeOptTopLeft: "top left", KThemeOptTopRight: "top right", KThemeOptTopCenter: "top centre",
	KThemeOptBottomLeft: "bottom left", KThemeOptBottomRight: "bottom right", KThemeOptBottomCenter: "bottom centre",
	KThemeOptBackground: "full background", KThemeOptLeft: "left", KThemeOptCenter: "centre", KThemeOptRight: "right",
}

var themeFR = map[Key]string{
	KLblThemeEditor: "Éditeur de thème",
	KTipThemeEditor: "Modifier le thème du diaporama avec des réglages : couleurs, polices, logo, numéro de page. L’aperçu suit ; le thème est conservé sous forme de paquet exportable.",

	KThemeSecPalette: "Palette", KThemeSecType: "Typographie", KThemeSecLogo: "Logo", KThemeSecFooter: "Numéro de page et pied de page",
	KThemeSecMeta: "Détails", KThemeSecContrast: "Contraste (WCAG)",

	KThemePBackground: "Fond", KThemePText: "Texte", KThemePHeading: "Titres", KThemePAccent: "Accent", KThemePLink: "Liens",
	KThemePFontBody: "Police du texte", KThemePFontHeading: "Police des titres", KThemePFontCode: "Police du code",
	KThemePSizeBody: "Taille du texte", KThemePSizeHeading: "Taille des titres", KThemePSizeCode: "Taille du code",
	KThemePWeightBody: "Graisse du texte", KThemePWeightHeading: "Graisse des titres",
	KThemePLogoPlace: "Emplacement", KThemePLogoSize: "Taille du logo", KThemePPagePlace: "Numéro de page", KThemePPageSize: "Taille du numéro de page",
	KThemePFooterAlign: "Alignement du pied de page", KThemePFooterSize: "Taille du pied de page",

	KThemeMName: "Nom", KThemeMVersion: "Version", KThemeMDescription: "Description", KThemeMAuthors: "Auteurs (séparés par ;)",
	KThemeMLicense: "Licence (SPDX)", KThemeMHomepage: "Page d’accueil", KThemeMExtends: "Basé sur", KThemeMMinSlides: "Requiert go-slides",
	KThemeMFontLicense: "Licence de la police : %s",

	KThemeBtnNew: "Nouveau thème", KThemeBtnImport: "Importer…", KThemeBtnPackage: "Exporter .gsta.xz", KThemeBtnCSS: "Exporter le CSS",
	KThemeBtnVersion: "Enregistrer une version", KThemeBtnFont: "Ajouter une police…", KThemeBtnLogo: "Logo…", KThemeBtnNoLogo: "Retirer le logo",
	KThemeBtnOpenCSS: "Modifier le CSS", KThemeBtnRestore: "Rétablir le bloc",

	KThemeContrastPair: "%s sur %s", KThemeContrastFail: "Échec", KThemeContrastNeed: "Réglez les deux couleurs pour vérifier",

	KThemeNoteImport:   "L’@import de « %s » dans la feuille est devenu la base du thème (extends) : un paquet y nomme sa base.",
	KThemeNoteDiverged: "Le bloc entre les commentaires gsta:visual a été modifié à la main : les réglages n’y écrivent plus. « Rétablir le bloc » le réécrit à partir des valeurs qu’il contient ; votre CSS en dehors du bloc n’est jamais touché.",
	KThemeNoteNone:     "Aucun thème n’est en cours de modification. Créez-en un, importez un paquet ou une feuille de style, ou choisissez un thème du projet.",

	KStatusThemeMade:        "Thème « %s » créé sur « %s » : le diaporama l’utilise.",
	KStatusThemeOpened:      "%s ouvert comme thème « %s ».",
	KStatusThemeExported:    "%s exporté. Il est vérifié par le même code que slides theme verify.",
	KStatusThemeCSSExported: "%s exporté.",
	KStatusThemeSaved:       "Version « %s » conservée.",

	KStepThemeEdit: "Modifier le thème : %s", KStepThemeNew: "Nouveau thème « %s »", KStepThemeFont: "Ajouter la police %s", KStepThemeLogo: "Définir le logo",

	KErrThemeDiverged: "le bloc CSS entre les commentaires gsta:visual a été modifié à la main : rétablissez-le d’abord",
	KErrThemeValue:    "« %s » n’est pas une valeur de %s",
	KErrThemeNotSheet: "%s n’est ni une feuille de style (.css) ni un paquet de thème (.gsta.xz)",

	KThemeOptNone: "aucun", KThemeOptTopLeft: "en haut à gauche", KThemeOptTopRight: "en haut à droite", KThemeOptTopCenter: "en haut au centre",
	KThemeOptBottomLeft: "en bas à gauche", KThemeOptBottomRight: "en bas à droite", KThemeOptBottomCenter: "en bas au centre",
	KThemeOptBackground: "fond entier", KThemeOptLeft: "à gauche", KThemeOptCenter: "au centre", KThemeOptRight: "à droite",
}

var themeDE = map[Key]string{
	KLblThemeEditor: "Theme-Editor",
	KTipThemeEditor: "Das Theme der Präsentation mit Reglern bearbeiten: Farben, Schriften, Logo, Seitenzahl. Die Vorschau folgt; das Theme wird als exportierbares Paket gehalten.",

	KThemeSecPalette: "Palette", KThemeSecType: "Typografie", KThemeSecLogo: "Logo", KThemeSecFooter: "Seitenzahl und Fußzeile",
	KThemeSecMeta: "Angaben", KThemeSecContrast: "Kontrast (WCAG)",

	KThemePBackground: "Hintergrund", KThemePText: "Text", KThemePHeading: "Überschriften", KThemePAccent: "Akzent", KThemePLink: "Links",
	KThemePFontBody: "Textschrift", KThemePFontHeading: "Überschriftschrift", KThemePFontCode: "Codeschrift",
	KThemePSizeBody: "Textgröße", KThemePSizeHeading: "Überschriftgröße", KThemePSizeCode: "Codegröße",
	KThemePWeightBody: "Textstärke", KThemePWeightHeading: "Überschriftstärke",
	KThemePLogoPlace: "Platzierung", KThemePLogoSize: "Logogröße", KThemePPagePlace: "Seitenzahl", KThemePPageSize: "Größe der Seitenzahl",
	KThemePFooterAlign: "Ausrichtung der Fußzeile", KThemePFooterSize: "Größe der Fußzeile",

	KThemeMName: "Name", KThemeMVersion: "Version", KThemeMDescription: "Beschreibung", KThemeMAuthors: "Autoren (mit ; getrennt)",
	KThemeMLicense: "Lizenz (SPDX)", KThemeMHomepage: "Webseite", KThemeMExtends: "Basiert auf", KThemeMMinSlides: "Benötigt go-slides",
	KThemeMFontLicense: "Schriftlizenz: %s",

	KThemeBtnNew: "Neues Theme", KThemeBtnImport: "Importieren …", KThemeBtnPackage: ".gsta.xz exportieren", KThemeBtnCSS: "CSS exportieren",
	KThemeBtnVersion: "Version speichern", KThemeBtnFont: "Schrift hinzufügen …", KThemeBtnLogo: "Logo …", KThemeBtnNoLogo: "Logo entfernen",
	KThemeBtnOpenCSS: "CSS bearbeiten", KThemeBtnRestore: "Block wiederherstellen",

	KThemeContrastPair: "%s auf %s", KThemeContrastFail: "Nicht bestanden", KThemeContrastNeed: "Beide Farben festlegen, um zu prüfen",

	KThemeNoteImport:   "Das @import von „%s“ im Stylesheet ist zur Basis des Themes geworden (extends): Ein Paket nennt seine Basis dort.",
	KThemeNoteDiverged: "Der Block zwischen den gsta:visual-Kommentaren wurde von Hand geändert, daher schreiben die Regler nicht hinein. „Block wiederherstellen“ schreibt ihn aus den enthaltenen Werten neu; Ihr CSS außerhalb des Blocks bleibt unberührt.",
	KThemeNoteNone:     "Es wird kein Theme bearbeitet. Legen Sie ein neues an, importieren Sie ein Paket oder ein Stylesheet oder wählen Sie ein Theme des Projekts.",

	KStatusThemeMade:        "Theme „%s“ auf Basis von „%s“ erstellt: Die Präsentation verwendet es.",
	KStatusThemeOpened:      "%s als Theme „%s“ geöffnet.",
	KStatusThemeExported:    "%s exportiert. Es wird vom selben Code geprüft wie slides theme verify.",
	KStatusThemeCSSExported: "%s exportiert.",
	KStatusThemeSaved:       "Version „%s“ gespeichert.",

	KStepThemeEdit: "Theme bearbeiten: %s", KStepThemeNew: "Neues Theme „%s“", KStepThemeFont: "Schrift %s hinzufügen", KStepThemeLogo: "Logo festlegen",

	KErrThemeDiverged: "der CSS-Block zwischen den gsta:visual-Kommentaren wurde von Hand geändert: erst wiederherstellen",
	KErrThemeValue:    "„%s“ ist kein Wert für %s",
	KErrThemeNotSheet: "%s ist weder ein Stylesheet (.css) noch ein Theme-Paket (.gsta.xz)",

	KThemeOptNone: "keine", KThemeOptTopLeft: "oben links", KThemeOptTopRight: "oben rechts", KThemeOptTopCenter: "oben mittig",
	KThemeOptBottomLeft: "unten links", KThemeOptBottomRight: "unten rechts", KThemeOptBottomCenter: "unten mittig",
	KThemeOptBackground: "ganzer Hintergrund", KThemeOptLeft: "links", KThemeOptCenter: "mittig", KThemeOptRight: "rechts",
}

var themeES = map[Key]string{
	KLblThemeEditor: "Editor de temas",
	KTipThemeEditor: "Edite el tema de la presentación con controles: colores, fuentes, logotipo, número de página. La vista previa sigue; el tema se guarda como un paquete que puede exportar.",

	KThemeSecPalette: "Paleta", KThemeSecType: "Tipografía", KThemeSecLogo: "Logotipo", KThemeSecFooter: "Número de página y pie",
	KThemeSecMeta: "Detalles", KThemeSecContrast: "Contraste (WCAG)",

	KThemePBackground: "Fondo", KThemePText: "Texto", KThemePHeading: "Títulos", KThemePAccent: "Acento", KThemePLink: "Enlaces",
	KThemePFontBody: "Fuente del texto", KThemePFontHeading: "Fuente de los títulos", KThemePFontCode: "Fuente del código",
	KThemePSizeBody: "Tamaño del texto", KThemePSizeHeading: "Tamaño de los títulos", KThemePSizeCode: "Tamaño del código",
	KThemePWeightBody: "Grosor del texto", KThemePWeightHeading: "Grosor de los títulos",
	KThemePLogoPlace: "Ubicación", KThemePLogoSize: "Tamaño del logotipo", KThemePPagePlace: "Número de página", KThemePPageSize: "Tamaño del número de página",
	KThemePFooterAlign: "Alineación del pie", KThemePFooterSize: "Tamaño del pie",

	KThemeMName: "Nombre", KThemeMVersion: "Versión", KThemeMDescription: "Descripción", KThemeMAuthors: "Autores (separados por ;)",
	KThemeMLicense: "Licencia (SPDX)", KThemeMHomepage: "Página web", KThemeMExtends: "Basado en", KThemeMMinSlides: "Requiere go-slides",
	KThemeMFontLicense: "Licencia de la fuente: %s",

	KThemeBtnNew: "Tema nuevo", KThemeBtnImport: "Importar…", KThemeBtnPackage: "Exportar .gsta.xz", KThemeBtnCSS: "Exportar CSS",
	KThemeBtnVersion: "Guardar versión", KThemeBtnFont: "Añadir fuente…", KThemeBtnLogo: "Logotipo…", KThemeBtnNoLogo: "Quitar logotipo",
	KThemeBtnOpenCSS: "Editar CSS", KThemeBtnRestore: "Restaurar el bloque",

	KThemeContrastPair: "%s sobre %s", KThemeContrastFail: "No cumple", KThemeContrastNeed: "Defina ambos colores para comprobar",

	KThemeNoteImport:   "El @import de «%s» de la hoja pasó a ser la base del tema (extends): un paquete nombra allí su base.",
	KThemeNoteDiverged: "El bloque entre los comentarios gsta:visual se editó a mano, así que los controles no escriben en él. «Restaurar el bloque» lo reescribe a partir de los valores que contiene; su CSS fuera del bloque nunca se toca.",
	KThemeNoteNone:     "No se está editando ningún tema. Cree uno nuevo, importe un paquete o una hoja de estilo, o elija un tema del proyecto.",

	KStatusThemeMade:        "Tema «%s» creado sobre «%s»: la presentación lo usa.",
	KStatusThemeOpened:      "%s abierto como el tema «%s».",
	KStatusThemeExported:    "%s exportado. Lo comprueba el mismo código que slides theme verify.",
	KStatusThemeCSSExported: "%s exportado.",
	KStatusThemeSaved:       "Versión «%s» guardada.",

	KStepThemeEdit: "Editar el tema: %s", KStepThemeNew: "Tema nuevo «%s»", KStepThemeFont: "Añadir la fuente %s", KStepThemeLogo: "Definir el logotipo",

	KErrThemeDiverged: "el bloque CSS entre los comentarios gsta:visual se editó a mano: restáurelo primero",
	KErrThemeValue:    "«%s» no es un valor de %s",
	KErrThemeNotSheet: "%s no es una hoja de estilo (.css) ni un paquete de tema (.gsta.xz)",

	KThemeOptNone: "ninguno", KThemeOptTopLeft: "arriba a la izquierda", KThemeOptTopRight: "arriba a la derecha", KThemeOptTopCenter: "arriba al centro",
	KThemeOptBottomLeft: "abajo a la izquierda", KThemeOptBottomRight: "abajo a la derecha", KThemeOptBottomCenter: "abajo al centro",
	KThemeOptBackground: "fondo completo", KThemeOptLeft: "izquierda", KThemeOptCenter: "centro", KThemeOptRight: "derecha",
}

var themeIT = map[Key]string{
	KLblThemeEditor: "Editor di temi",
	KTipThemeEditor: "Modifica il tema della presentazione con i controlli: colori, caratteri, logo, numero di pagina. L’anteprima segue; il tema è conservato come un pacchetto esportabile.",

	KThemeSecPalette: "Tavolozza", KThemeSecType: "Tipografia", KThemeSecLogo: "Logo", KThemeSecFooter: "Numero di pagina e piè di pagina",
	KThemeSecMeta: "Dettagli", KThemeSecContrast: "Contrasto (WCAG)",

	KThemePBackground: "Sfondo", KThemePText: "Testo", KThemePHeading: "Titoli", KThemePAccent: "Accento", KThemePLink: "Link",
	KThemePFontBody: "Carattere del testo", KThemePFontHeading: "Carattere dei titoli", KThemePFontCode: "Carattere del codice",
	KThemePSizeBody: "Dimensione del testo", KThemePSizeHeading: "Dimensione dei titoli", KThemePSizeCode: "Dimensione del codice",
	KThemePWeightBody: "Spessore del testo", KThemePWeightHeading: "Spessore dei titoli",
	KThemePLogoPlace: "Posizione", KThemePLogoSize: "Dimensione del logo", KThemePPagePlace: "Numero di pagina", KThemePPageSize: "Dimensione del numero di pagina",
	KThemePFooterAlign: "Allineamento del piè di pagina", KThemePFooterSize: "Dimensione del piè di pagina",

	KThemeMName: "Nome", KThemeMVersion: "Versione", KThemeMDescription: "Descrizione", KThemeMAuthors: "Autori (separati da ;)",
	KThemeMLicense: "Licenza (SPDX)", KThemeMHomepage: "Pagina web", KThemeMExtends: "Basato su", KThemeMMinSlides: "Richiede go-slides",
	KThemeMFontLicense: "Licenza del carattere: %s",

	KThemeBtnNew: "Nuovo tema", KThemeBtnImport: "Importa…", KThemeBtnPackage: "Esporta .gsta.xz", KThemeBtnCSS: "Esporta CSS",
	KThemeBtnVersion: "Salva versione", KThemeBtnFont: "Aggiungi carattere…", KThemeBtnLogo: "Logo…", KThemeBtnNoLogo: "Rimuovi logo",
	KThemeBtnOpenCSS: "Modifica CSS", KThemeBtnRestore: "Ripristina il blocco",

	KThemeContrastPair: "%s su %s", KThemeContrastFail: "Non conforme", KThemeContrastNeed: "Imposta entrambi i colori per verificare",

	KThemeNoteImport:   "L’@import di «%s» nel foglio è diventato la base del tema (extends): un pacchetto indica lì la sua base.",
	KThemeNoteDiverged: "Il blocco tra i commenti gsta:visual è stato modificato a mano, quindi i controlli non vi scrivono. «Ripristina il blocco» lo riscrive dai valori che contiene; il tuo CSS fuori dal blocco non viene mai toccato.",
	KThemeNoteNone:     "Nessun tema in modifica. Creane uno nuovo, importa un pacchetto o un foglio di stile, oppure scegli un tema del progetto.",

	KStatusThemeMade:        "Tema «%s» creato su «%s»: la presentazione lo usa.",
	KStatusThemeOpened:      "%s aperto come tema «%s».",
	KStatusThemeExported:    "%s esportato. È verificato dallo stesso codice di slides theme verify.",
	KStatusThemeCSSExported: "%s esportato.",
	KStatusThemeSaved:       "Versione «%s» conservata.",

	KStepThemeEdit: "Modifica il tema: %s", KStepThemeNew: "Nuovo tema «%s»", KStepThemeFont: "Aggiungi il carattere %s", KStepThemeLogo: "Imposta il logo",

	KErrThemeDiverged: "il blocco CSS tra i commenti gsta:visual è stato modificato a mano: ripristinalo prima",
	KErrThemeValue:    "«%s» non è un valore di %s",
	KErrThemeNotSheet: "%s non è un foglio di stile (.css) né un pacchetto di tema (.gsta.xz)",

	KThemeOptNone: "nessuno", KThemeOptTopLeft: "in alto a sinistra", KThemeOptTopRight: "in alto a destra", KThemeOptTopCenter: "in alto al centro",
	KThemeOptBottomLeft: "in basso a sinistra", KThemeOptBottomRight: "in basso a destra", KThemeOptBottomCenter: "in basso al centro",
	KThemeOptBackground: "sfondo intero", KThemeOptLeft: "a sinistra", KThemeOptCenter: "al centro", KThemeOptRight: "a destra",
}

var themePT = map[Key]string{
	KLblThemeEditor: "Editor de temas",
	KTipThemeEditor: "Edite o tema da apresentação com controles: cores, fontes, logotipo, número de página. A pré-visualização acompanha; o tema é guardado como um pacote que você pode exportar.",

	KThemeSecPalette: "Paleta", KThemeSecType: "Tipografia", KThemeSecLogo: "Logotipo", KThemeSecFooter: "Número de página e rodapé",
	KThemeSecMeta: "Detalhes", KThemeSecContrast: "Contraste (WCAG)",

	KThemePBackground: "Fundo", KThemePText: "Texto", KThemePHeading: "Títulos", KThemePAccent: "Destaque", KThemePLink: "Links",
	KThemePFontBody: "Fonte do texto", KThemePFontHeading: "Fonte dos títulos", KThemePFontCode: "Fonte do código",
	KThemePSizeBody: "Tamanho do texto", KThemePSizeHeading: "Tamanho dos títulos", KThemePSizeCode: "Tamanho do código",
	KThemePWeightBody: "Peso do texto", KThemePWeightHeading: "Peso dos títulos",
	KThemePLogoPlace: "Posição", KThemePLogoSize: "Tamanho do logotipo", KThemePPagePlace: "Número de página", KThemePPageSize: "Tamanho do número de página",
	KThemePFooterAlign: "Alinhamento do rodapé", KThemePFooterSize: "Tamanho do rodapé",

	KThemeMName: "Nome", KThemeMVersion: "Versão", KThemeMDescription: "Descrição", KThemeMAuthors: "Autores (separados por ;)",
	KThemeMLicense: "Licença (SPDX)", KThemeMHomepage: "Página inicial", KThemeMExtends: "Baseado em", KThemeMMinSlides: "Requer go-slides",
	KThemeMFontLicense: "Licença da fonte: %s",

	KThemeBtnNew: "Novo tema", KThemeBtnImport: "Importar…", KThemeBtnPackage: "Exportar .gsta.xz", KThemeBtnCSS: "Exportar CSS",
	KThemeBtnVersion: "Salvar versão", KThemeBtnFont: "Adicionar fonte…", KThemeBtnLogo: "Logotipo…", KThemeBtnNoLogo: "Remover logotipo",
	KThemeBtnOpenCSS: "Editar CSS", KThemeBtnRestore: "Restaurar o bloco",

	KThemeContrastPair: "%s sobre %s", KThemeContrastFail: "Reprovado", KThemeContrastNeed: "Defina as duas cores para verificar",

	KThemeNoteImport:   "O @import de “%s” na folha virou a base do tema (extends): um pacote indica ali a sua base.",
	KThemeNoteDiverged: "O bloco entre os comentários gsta:visual foi editado à mão, então os controles não escrevem nele. “Restaurar o bloco” o reescreve a partir dos valores que ele contém; seu CSS fora do bloco nunca é tocado.",
	KThemeNoteNone:     "Nenhum tema está sendo editado. Crie um novo, importe um pacote ou uma folha de estilo, ou escolha um tema do projeto.",

	KStatusThemeMade:        "Tema “%s” criado sobre “%s”: a apresentação o usa.",
	KStatusThemeOpened:      "%s aberto como o tema “%s”.",
	KStatusThemeExported:    "%s exportado. Ele é verificado pelo mesmo código de slides theme verify.",
	KStatusThemeCSSExported: "%s exportado.",
	KStatusThemeSaved:       "Versão “%s” guardada.",

	KStepThemeEdit: "Editar o tema: %s", KStepThemeNew: "Novo tema “%s”", KStepThemeFont: "Adicionar a fonte %s", KStepThemeLogo: "Definir o logotipo",

	KErrThemeDiverged: "o bloco CSS entre os comentários gsta:visual foi editado à mão: restaure-o primeiro",
	KErrThemeValue:    "“%s” não é um valor de %s",
	KErrThemeNotSheet: "%s não é uma folha de estilo (.css) nem um pacote de tema (.gsta.xz)",

	KThemeOptNone: "nenhum", KThemeOptTopLeft: "no alto à esquerda", KThemeOptTopRight: "no alto à direita", KThemeOptTopCenter: "no alto ao centro",
	KThemeOptBottomLeft: "embaixo à esquerda", KThemeOptBottomRight: "embaixo à direita", KThemeOptBottomCenter: "embaixo ao centro",
	KThemeOptBackground: "fundo inteiro", KThemeOptLeft: "à esquerda", KThemeOptCenter: "ao centro", KThemeOptRight: "à direita",
}

var themePL = map[Key]string{
	KLblThemeEditor: "Edytor motywów",
	KTipThemeEditor: "Edytuj motyw prezentacji za pomocą kontrolek: kolory, czcionki, logo, numer strony. Podgląd nadąża; motyw jest przechowywany jako pakiet, który można wyeksportować.",

	KThemeSecPalette: "Paleta", KThemeSecType: "Typografia", KThemeSecLogo: "Logo", KThemeSecFooter: "Numer strony i stopka",
	KThemeSecMeta: "Szczegóły", KThemeSecContrast: "Kontrast (WCAG)",

	KThemePBackground: "Tło", KThemePText: "Tekst", KThemePHeading: "Nagłówki", KThemePAccent: "Akcent", KThemePLink: "Odnośniki",
	KThemePFontBody: "Czcionka tekstu", KThemePFontHeading: "Czcionka nagłówków", KThemePFontCode: "Czcionka kodu",
	KThemePSizeBody: "Rozmiar tekstu", KThemePSizeHeading: "Rozmiar nagłówków", KThemePSizeCode: "Rozmiar kodu",
	KThemePWeightBody: "Grubość tekstu", KThemePWeightHeading: "Grubość nagłówków",
	KThemePLogoPlace: "Położenie", KThemePLogoSize: "Rozmiar logo", KThemePPagePlace: "Numer strony", KThemePPageSize: "Rozmiar numeru strony",
	KThemePFooterAlign: "Wyrównanie stopki", KThemePFooterSize: "Rozmiar stopki",

	KThemeMName: "Nazwa", KThemeMVersion: "Wersja", KThemeMDescription: "Opis", KThemeMAuthors: "Autorzy (rozdzieleni średnikiem)",
	KThemeMLicense: "Licencja (SPDX)", KThemeMHomepage: "Strona domowa", KThemeMExtends: "Oparty na", KThemeMMinSlides: "Wymaga go-slides",
	KThemeMFontLicense: "Licencja czcionki: %s",

	KThemeBtnNew: "Nowy motyw", KThemeBtnImport: "Importuj…", KThemeBtnPackage: "Eksportuj .gsta.xz", KThemeBtnCSS: "Eksportuj CSS",
	KThemeBtnVersion: "Zapisz wersję", KThemeBtnFont: "Dodaj czcionkę…", KThemeBtnLogo: "Logo…", KThemeBtnNoLogo: "Usuń logo",
	KThemeBtnOpenCSS: "Edytuj CSS", KThemeBtnRestore: "Przywróć blok",

	KThemeContrastPair: "%s na %s", KThemeContrastFail: "Niezgodne", KThemeContrastNeed: "Ustaw oba kolory, aby sprawdzić",

	KThemeNoteImport:   "@import „%s” z arkusza stał się bazą motywu (extends): pakiet wskazuje tam swoją bazę.",
	KThemeNoteDiverged: "Blok między komentarzami gsta:visual został zmieniony ręcznie, więc kontrolki nie zapisują w nim zmian. „Przywróć blok” zapisuje go ponownie z zawartych w nim wartości; Twój CSS poza blokiem nigdy nie jest ruszany.",
	KThemeNoteNone:     "Żaden motyw nie jest edytowany. Utwórz nowy, zaimportuj pakiet lub arkusz stylów albo wybierz motyw projektu.",

	KStatusThemeMade:        "Utworzono motyw „%s” na bazie „%s”: prezentacja go używa.",
	KStatusThemeOpened:      "Otwarto %s jako motyw „%s”.",
	KStatusThemeExported:    "Wyeksportowano %s. Sprawdza go ten sam kod co slides theme verify.",
	KStatusThemeCSSExported: "Wyeksportowano %s.",
	KStatusThemeSaved:       "Zachowano wersję „%s”.",

	KStepThemeEdit: "Edytuj motyw: %s", KStepThemeNew: "Nowy motyw „%s”", KStepThemeFont: "Dodaj czcionkę %s", KStepThemeLogo: "Ustaw logo",

	KErrThemeDiverged: "blok CSS między komentarzami gsta:visual został zmieniony ręcznie: najpierw go przywróć",
	KErrThemeValue:    "„%s” nie jest wartością dla %s",
	KErrThemeNotSheet: "%s nie jest arkuszem stylów (.css) ani pakietem motywu (.gsta.xz)",

	KThemeOptNone: "brak", KThemeOptTopLeft: "u góry po lewej", KThemeOptTopRight: "u góry po prawej", KThemeOptTopCenter: "u góry na środku",
	KThemeOptBottomLeft: "u dołu po lewej", KThemeOptBottomRight: "u dołu po prawej", KThemeOptBottomCenter: "u dołu na środku",
	KThemeOptBackground: "całe tło", KThemeOptLeft: "do lewej", KThemeOptCenter: "do środka", KThemeOptRight: "do prawej",
}

var themeUK = map[Key]string{
	KLblThemeEditor: "Редактор тем",
	KTipThemeEditor: "Редагуйте тему презентації за допомогою елементів керування: кольори, шрифти, логотип, номер сторінки. Попередній перегляд слідує за змінами; тема зберігається як пакет, який можна експортувати.",

	KThemeSecPalette: "Палітра", KThemeSecType: "Типографіка", KThemeSecLogo: "Логотип", KThemeSecFooter: "Номер сторінки та колонтитул",
	KThemeSecMeta: "Відомості", KThemeSecContrast: "Контраст (WCAG)",

	KThemePBackground: "Тло", KThemePText: "Текст", KThemePHeading: "Заголовки", KThemePAccent: "Акцент", KThemePLink: "Посилання",
	KThemePFontBody: "Шрифт тексту", KThemePFontHeading: "Шрифт заголовків", KThemePFontCode: "Шрифт коду",
	KThemePSizeBody: "Розмір тексту", KThemePSizeHeading: "Розмір заголовків", KThemePSizeCode: "Розмір коду",
	KThemePWeightBody: "Насиченість тексту", KThemePWeightHeading: "Насиченість заголовків",
	KThemePLogoPlace: "Розташування", KThemePLogoSize: "Розмір логотипа", KThemePPagePlace: "Номер сторінки", KThemePPageSize: "Розмір номера сторінки",
	KThemePFooterAlign: "Вирівнювання колонтитула", KThemePFooterSize: "Розмір колонтитула",

	KThemeMName: "Назва", KThemeMVersion: "Версія", KThemeMDescription: "Опис", KThemeMAuthors: "Автори (через ;)",
	KThemeMLicense: "Ліцензія (SPDX)", KThemeMHomepage: "Вебсторінка", KThemeMExtends: "Заснована на", KThemeMMinSlides: "Потребує go-slides",
	KThemeMFontLicense: "Ліцензія шрифту: %s",

	KThemeBtnNew: "Нова тема", KThemeBtnImport: "Імпортувати…", KThemeBtnPackage: "Експортувати .gsta.xz", KThemeBtnCSS: "Експортувати CSS",
	KThemeBtnVersion: "Зберегти версію", KThemeBtnFont: "Додати шрифт…", KThemeBtnLogo: "Логотип…", KThemeBtnNoLogo: "Прибрати логотип",
	KThemeBtnOpenCSS: "Редагувати CSS", KThemeBtnRestore: "Відновити блок",

	KThemeContrastPair: "%s на %s", KThemeContrastFail: "Не відповідає", KThemeContrastNeed: "Задайте обидва кольори для перевірки",

	KThemeNoteImport:   "@import «%s» в аркуші став основою теми (extends): пакет указує свою основу саме там.",
	KThemeNoteDiverged: "Блок між коментарями gsta:visual змінено вручну, тому елементи керування в нього не пишуть. «Відновити блок» переписує його зі значень, що в ньому є; ваш CSS поза блоком ніколи не змінюється.",
	KThemeNoteNone:     "Жодну тему не редагується. Створіть нову, імпортуйте пакет або аркуш стилів чи виберіть тему проєкту.",

	KStatusThemeMade:        "Створено тему «%s» на основі «%s»: презентація її використовує.",
	KStatusThemeOpened:      "%s відкрито як тему «%s».",
	KStatusThemeExported:    "Експортовано %s. Його перевіряє той самий код, що й slides theme verify.",
	KStatusThemeCSSExported: "Експортовано %s.",
	KStatusThemeSaved:       "Збережено версію «%s».",

	KStepThemeEdit: "Змінити тему: %s", KStepThemeNew: "Нова тема «%s»", KStepThemeFont: "Додати шрифт %s", KStepThemeLogo: "Задати логотип",

	KErrThemeDiverged: "блок CSS між коментарями gsta:visual змінено вручну: спершу відновіть його",
	KErrThemeValue:    "«%s» не є значенням для %s",
	KErrThemeNotSheet: "%s — це не аркуш стилів (.css) і не пакет теми (.gsta.xz)",

	KThemeOptNone: "немає", KThemeOptTopLeft: "угорі ліворуч", KThemeOptTopRight: "угорі праворуч", KThemeOptTopCenter: "угорі посередині",
	KThemeOptBottomLeft: "унизу ліворуч", KThemeOptBottomRight: "унизу праворуч", KThemeOptBottomCenter: "унизу посередині",
	KThemeOptBackground: "усе тло", KThemeOptLeft: "ліворуч", KThemeOptCenter: "посередині", KThemeOptRight: "праворуч",
}

var themeRU = map[Key]string{
	KLblThemeEditor: "Редактор тем",
	KTipThemeEditor: "Редактируйте тему презентации с помощью элементов управления: цвета, шрифты, логотип, номер страницы. Предпросмотр следует за изменениями; тема хранится как пакет, который можно экспортировать.",

	KThemeSecPalette: "Палитра", KThemeSecType: "Типографика", KThemeSecLogo: "Логотип", KThemeSecFooter: "Номер страницы и колонтитул",
	KThemeSecMeta: "Сведения", KThemeSecContrast: "Контраст (WCAG)",

	KThemePBackground: "Фон", KThemePText: "Текст", KThemePHeading: "Заголовки", KThemePAccent: "Акцент", KThemePLink: "Ссылки",
	KThemePFontBody: "Шрифт текста", KThemePFontHeading: "Шрифт заголовков", KThemePFontCode: "Шрифт кода",
	KThemePSizeBody: "Размер текста", KThemePSizeHeading: "Размер заголовков", KThemePSizeCode: "Размер кода",
	KThemePWeightBody: "Насыщенность текста", KThemePWeightHeading: "Насыщенность заголовков",
	KThemePLogoPlace: "Расположение", KThemePLogoSize: "Размер логотипа", KThemePPagePlace: "Номер страницы", KThemePPageSize: "Размер номера страницы",
	KThemePFooterAlign: "Выравнивание колонтитула", KThemePFooterSize: "Размер колонтитула",

	KThemeMName: "Название", KThemeMVersion: "Версия", KThemeMDescription: "Описание", KThemeMAuthors: "Авторы (через ;)",
	KThemeMLicense: "Лицензия (SPDX)", KThemeMHomepage: "Сайт", KThemeMExtends: "Основана на", KThemeMMinSlides: "Требует go-slides",
	KThemeMFontLicense: "Лицензия шрифта: %s",

	KThemeBtnNew: "Новая тема", KThemeBtnImport: "Импортировать…", KThemeBtnPackage: "Экспортировать .gsta.xz", KThemeBtnCSS: "Экспортировать CSS",
	KThemeBtnVersion: "Сохранить версию", KThemeBtnFont: "Добавить шрифт…", KThemeBtnLogo: "Логотип…", KThemeBtnNoLogo: "Убрать логотип",
	KThemeBtnOpenCSS: "Править CSS", KThemeBtnRestore: "Восстановить блок",

	KThemeContrastPair: "%s на %s", KThemeContrastFail: "Не проходит", KThemeContrastNeed: "Задайте оба цвета для проверки",

	KThemeNoteImport:   "@import «%s» в таблице стал основой темы (extends): пакет указывает свою основу именно там.",
	KThemeNoteDiverged: "Блок между комментариями gsta:visual изменён вручную, поэтому элементы управления в него не пишут. «Восстановить блок» переписывает его из имеющихся в нём значений; ваш CSS вне блока не затрагивается никогда.",
	KThemeNoteNone:     "Ни одна тема не редактируется. Создайте новую, импортируйте пакет или таблицу стилей либо выберите тему проекта.",

	KStatusThemeMade:        "Создана тема «%s» на основе «%s»: презентация её использует.",
	KStatusThemeOpened:      "%s открыт как тема «%s».",
	KStatusThemeExported:    "Экспортировано: %s. Его проверяет тот же код, что и slides theme verify.",
	KStatusThemeCSSExported: "Экспортировано: %s.",
	KStatusThemeSaved:       "Версия «%s» сохранена.",

	KStepThemeEdit: "Изменить тему: %s", KStepThemeNew: "Новая тема «%s»", KStepThemeFont: "Добавить шрифт %s", KStepThemeLogo: "Задать логотип",

	KErrThemeDiverged: "блок CSS между комментариями gsta:visual изменён вручную: сначала восстановите его",
	KErrThemeValue:    "«%s» не является значением для %s",
	KErrThemeNotSheet: "%s — не таблица стилей (.css) и не пакет темы (.gsta.xz)",

	KThemeOptNone: "нет", KThemeOptTopLeft: "вверху слева", KThemeOptTopRight: "вверху справа", KThemeOptTopCenter: "вверху по центру",
	KThemeOptBottomLeft: "внизу слева", KThemeOptBottomRight: "внизу справа", KThemeOptBottomCenter: "внизу по центру",
	KThemeOptBackground: "весь фон", KThemeOptLeft: "слева", KThemeOptCenter: "по центру", KThemeOptRight: "справа",
}

var themeZH = map[Key]string{
	KLblThemeEditor: "主题编辑器",
	KTipThemeEditor: "用控件编辑演示文稿的主题：颜色、字体、徽标、页码。预览会同步更新；主题以可导出的主题包形式保存。",

	KThemeSecPalette: "调色板", KThemeSecType: "字体排印", KThemeSecLogo: "徽标", KThemeSecFooter: "页码和页脚",
	KThemeSecMeta: "详细信息", KThemeSecContrast: "对比度（WCAG）",

	KThemePBackground: "背景", KThemePText: "文字", KThemePHeading: "标题", KThemePAccent: "强调色", KThemePLink: "链接",
	KThemePFontBody: "正文字体", KThemePFontHeading: "标题字体", KThemePFontCode: "代码字体",
	KThemePSizeBody: "正文字号", KThemePSizeHeading: "标题字号", KThemePSizeCode: "代码字号",
	KThemePWeightBody: "正文字重", KThemePWeightHeading: "标题字重",
	KThemePLogoPlace: "位置", KThemePLogoSize: "徽标大小", KThemePPagePlace: "页码", KThemePPageSize: "页码大小",
	KThemePFooterAlign: "页脚对齐", KThemePFooterSize: "页脚大小",

	KThemeMName: "名称", KThemeMVersion: "版本", KThemeMDescription: "说明", KThemeMAuthors: "作者（用 ; 分隔）",
	KThemeMLicense: "许可证（SPDX）", KThemeMHomepage: "主页", KThemeMExtends: "基于", KThemeMMinSlides: "需要 go-slides",
	KThemeMFontLicense: "字体许可证：%s",

	KThemeBtnNew: "新建主题", KThemeBtnImport: "导入…", KThemeBtnPackage: "导出 .gsta.xz", KThemeBtnCSS: "导出 CSS",
	KThemeBtnVersion: "保存版本", KThemeBtnFont: "添加字体…", KThemeBtnLogo: "徽标…", KThemeBtnNoLogo: "移除徽标",
	KThemeBtnOpenCSS: "编辑 CSS", KThemeBtnRestore: "恢复该代码块",

	KThemeContrastPair: "%s 对 %s", KThemeContrastFail: "不达标", KThemeContrastNeed: "请设置两种颜色以检查",

	KThemeNoteImport:   "样式表中对“%s”的 @import 已成为主题的基础（extends）：主题包在那里声明其基础主题。",
	KThemeNoteDiverged: "gsta:visual 注释之间的代码块已被手动修改，因此控件不会写入其中。“恢复该代码块”会根据其中的值重新写入；代码块之外的 CSS 永远不会被改动。",
	KThemeNoteNone:     "当前没有正在编辑的主题。请新建一个、导入主题包或样式表，或选择项目中的某个主题。",

	KStatusThemeMade:        "已基于“%s”创建主题“%s”：演示文稿正在使用它。",
	KStatusThemeOpened:      "已将 %s 打开为主题“%s”。",
	KStatusThemeExported:    "已导出 %s。它由与 slides theme verify 相同的代码检查。",
	KStatusThemeCSSExported: "已导出 %s。",
	KStatusThemeSaved:       "已保留版本“%s”。",

	KStepThemeEdit: "编辑主题：%s", KStepThemeNew: "新建主题“%s”", KStepThemeFont: "添加字体 %s", KStepThemeLogo: "设置徽标",

	KErrThemeDiverged: "gsta:visual 注释之间的 CSS 代码块已被手动修改：请先恢复它",
	KErrThemeValue:    "“%s”不是 %s 的有效值",
	KErrThemeNotSheet: "%s 既不是样式表（.css）也不是主题包（.gsta.xz）",

	KThemeOptNone: "无", KThemeOptTopLeft: "左上", KThemeOptTopRight: "右上", KThemeOptTopCenter: "上方居中",
	KThemeOptBottomLeft: "左下", KThemeOptBottomRight: "右下", KThemeOptBottomCenter: "下方居中",
	KThemeOptBackground: "整个背景", KThemeOptLeft: "靠左", KThemeOptCenter: "居中", KThemeOptRight: "靠右",
}

var themeKO = map[Key]string{
	KLblThemeEditor: "테마 편집기",
	KTipThemeEditor: "컨트롤로 프레젠테이션의 테마를 편집합니다: 색, 글꼴, 로고, 쪽 번호. 미리보기가 따라가며, 테마는 내보낼 수 있는 패키지로 보관됩니다.",

	KThemeSecPalette: "팔레트", KThemeSecType: "타이포그래피", KThemeSecLogo: "로고", KThemeSecFooter: "쪽 번호와 바닥글",
	KThemeSecMeta: "세부 정보", KThemeSecContrast: "대비(WCAG)",

	KThemePBackground: "배경", KThemePText: "텍스트", KThemePHeading: "제목", KThemePAccent: "강조색", KThemePLink: "링크",
	KThemePFontBody: "본문 글꼴", KThemePFontHeading: "제목 글꼴", KThemePFontCode: "코드 글꼴",
	KThemePSizeBody: "본문 크기", KThemePSizeHeading: "제목 크기", KThemePSizeCode: "코드 크기",
	KThemePWeightBody: "본문 굵기", KThemePWeightHeading: "제목 굵기",
	KThemePLogoPlace: "위치", KThemePLogoSize: "로고 크기", KThemePPagePlace: "쪽 번호", KThemePPageSize: "쪽 번호 크기",
	KThemePFooterAlign: "바닥글 정렬", KThemePFooterSize: "바닥글 크기",

	KThemeMName: "이름", KThemeMVersion: "버전", KThemeMDescription: "설명", KThemeMAuthors: "작성자(;로 구분)",
	KThemeMLicense: "라이선스(SPDX)", KThemeMHomepage: "홈페이지", KThemeMExtends: "기반 테마", KThemeMMinSlides: "필요한 go-slides",
	KThemeMFontLicense: "글꼴 라이선스: %s",

	KThemeBtnNew: "새 테마", KThemeBtnImport: "가져오기…", KThemeBtnPackage: ".gsta.xz 내보내기", KThemeBtnCSS: "CSS 내보내기",
	KThemeBtnVersion: "버전 저장", KThemeBtnFont: "글꼴 추가…", KThemeBtnLogo: "로고…", KThemeBtnNoLogo: "로고 제거",
	KThemeBtnOpenCSS: "CSS 편집", KThemeBtnRestore: "블록 복원",

	KThemeContrastPair: "%s / %s", KThemeContrastFail: "미달", KThemeContrastNeed: "확인하려면 두 색을 모두 지정하세요",

	KThemeNoteImport:   "스타일시트의 “%s” @import가 테마의 기반(extends)이 되었습니다: 패키지는 그곳에 기반 테마를 지정합니다.",
	KThemeNoteDiverged: "gsta:visual 주석 사이의 블록이 직접 편집되어 컨트롤이 그곳에 쓰지 않습니다. “블록 복원”은 블록에 담긴 값으로 다시 작성하며, 블록 밖의 CSS는 절대 건드리지 않습니다.",
	KThemeNoteNone:     "편집 중인 테마가 없습니다. 새로 만들거나, 패키지 또는 스타일시트를 가져오거나, 프로젝트의 테마를 고르세요.",

	KStatusThemeMade:        "“%s” 기반의 테마 “%s”을(를) 만들었습니다: 프레젠테이션이 이 테마를 사용합니다.",
	KStatusThemeOpened:      "%s을(를) 테마 “%s”(으)로 열었습니다.",
	KStatusThemeExported:    "%s을(를) 내보냈습니다. slides theme verify와 같은 코드로 검사됩니다.",
	KStatusThemeCSSExported: "%s을(를) 내보냈습니다.",
	KStatusThemeSaved:       "버전 “%s”을(를) 보관했습니다.",

	KStepThemeEdit: "테마 편집: %s", KStepThemeNew: "새 테마 “%s”", KStepThemeFont: "글꼴 %s 추가", KStepThemeLogo: "로고 설정",

	KErrThemeDiverged: "gsta:visual 주석 사이의 CSS 블록이 직접 편집되었습니다: 먼저 복원하세요",
	KErrThemeValue:    "“%s”은(는) %s의 값이 아닙니다",
	KErrThemeNotSheet: "%s은(는) 스타일시트(.css)도 테마 패키지(.gsta.xz)도 아닙니다",

	KThemeOptNone: "없음", KThemeOptTopLeft: "왼쪽 위", KThemeOptTopRight: "오른쪽 위", KThemeOptTopCenter: "위 가운데",
	KThemeOptBottomLeft: "왼쪽 아래", KThemeOptBottomRight: "오른쪽 아래", KThemeOptBottomCenter: "아래 가운데",
	KThemeOptBackground: "배경 전체", KThemeOptLeft: "왼쪽", KThemeOptCenter: "가운데", KThemeOptRight: "오른쪽",
}

var themeJA = map[Key]string{
	KLblThemeEditor: "テーマエディター",
	KTipThemeEditor: "コントロールでスライドのテーマを編集します（色、フォント、ロゴ、ページ番号）。プレビューが追従し、テーマはエクスポートできるパッケージとして保持されます。",

	KThemeSecPalette: "パレット", KThemeSecType: "タイポグラフィ", KThemeSecLogo: "ロゴ", KThemeSecFooter: "ページ番号とフッター",
	KThemeSecMeta: "詳細", KThemeSecContrast: "コントラスト（WCAG）",

	KThemePBackground: "背景", KThemePText: "テキスト", KThemePHeading: "見出し", KThemePAccent: "アクセント", KThemePLink: "リンク",
	KThemePFontBody: "本文のフォント", KThemePFontHeading: "見出しのフォント", KThemePFontCode: "コードのフォント",
	KThemePSizeBody: "本文のサイズ", KThemePSizeHeading: "見出しのサイズ", KThemePSizeCode: "コードのサイズ",
	KThemePWeightBody: "本文の太さ", KThemePWeightHeading: "見出しの太さ",
	KThemePLogoPlace: "配置", KThemePLogoSize: "ロゴのサイズ", KThemePPagePlace: "ページ番号", KThemePPageSize: "ページ番号のサイズ",
	KThemePFooterAlign: "フッターの配置", KThemePFooterSize: "フッターのサイズ",

	KThemeMName: "名前", KThemeMVersion: "バージョン", KThemeMDescription: "説明", KThemeMAuthors: "作者（; 区切り）",
	KThemeMLicense: "ライセンス（SPDX）", KThemeMHomepage: "ホームページ", KThemeMExtends: "ベース", KThemeMMinSlides: "必要な go-slides",
	KThemeMFontLicense: "フォントのライセンス: %s",

	KThemeBtnNew: "新しいテーマ", KThemeBtnImport: "インポート…", KThemeBtnPackage: ".gsta.xz をエクスポート", KThemeBtnCSS: "CSS をエクスポート",
	KThemeBtnVersion: "バージョンを保存", KThemeBtnFont: "フォントを追加…", KThemeBtnLogo: "ロゴ…", KThemeBtnNoLogo: "ロゴを削除",
	KThemeBtnOpenCSS: "CSS を編集", KThemeBtnRestore: "ブロックを復元",

	KThemeContrastPair: "%s／%s", KThemeContrastFail: "不適合", KThemeContrastNeed: "確認するには両方の色を設定してください",

	KThemeNoteImport:   "シートにあった「%s」の @import はテーマのベース（extends）になりました。パッケージではそこでベースを指定します。",
	KThemeNoteDiverged: "gsta:visual コメントの間のブロックが手で編集されたため、コントロールはそこに書き込みません。「ブロックを復元」はブロック内の値から書き直します。ブロックの外の CSS には一切触れません。",
	KThemeNoteNone:     "編集中のテーマはありません。新しく作成するか、パッケージまたはスタイルシートをインポートするか、プロジェクトのテーマを選んでください。",

	KStatusThemeMade:        "「%s」をベースにテーマ「%s」を作成しました。スライドがこれを使います。",
	KStatusThemeOpened:      "%s をテーマ「%s」として開きました。",
	KStatusThemeExported:    "%s をエクスポートしました。slides theme verify と同じコードで検査されます。",
	KStatusThemeCSSExported: "%s をエクスポートしました。",
	KStatusThemeSaved:       "バージョン「%s」を保存しました。",

	KStepThemeEdit: "テーマを編集: %s", KStepThemeNew: "新しいテーマ「%s」", KStepThemeFont: "フォント %s を追加", KStepThemeLogo: "ロゴを設定",

	KErrThemeDiverged: "gsta:visual コメントの間の CSS ブロックが手で編集されています。先に復元してください",
	KErrThemeValue:    "「%s」は %s の値ではありません",
	KErrThemeNotSheet: "%s はスタイルシート（.css）でもテーマパッケージ（.gsta.xz）でもありません",

	KThemeOptNone: "なし", KThemeOptTopLeft: "左上", KThemeOptTopRight: "右上", KThemeOptTopCenter: "上中央",
	KThemeOptBottomLeft: "左下", KThemeOptBottomRight: "右下", KThemeOptBottomCenter: "下中央",
	KThemeOptBackground: "背景全体", KThemeOptLeft: "左", KThemeOptCenter: "中央", KThemeOptRight: "右",
}

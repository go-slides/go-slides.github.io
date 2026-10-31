package playground

// esMessages is the Spanish catalogue.
var esMessages = map[Key]string{
	KTabHome: "Inicio", KTabInsert: "Insertar", KTabFormat: "Formato", KTabView: "Vista", KTabExport: "Exportar",

	KGrpDeck: "Presentación", KGrpLook: "Aspecto", KGrpPresent: "Presentar", KGrpDraft: "Borrador", KGrpSlides: "Diapositivas",
	KGrpPictures: "Imágenes", KGrpText: "Texto", KGrpPreview: "Vista previa", KGrpEditor: "Editor", KGrpSaveAs: "Guardar como",
	KGrpParagraph: "Párrafo", KGrpTable: "Tabla", KGrpRowsCols: "Filas y columnas", KGrpColumn: "Columna",
	KGrpSlideText: "Texto de diapositiva", KGrpLink: "Vínculo",

	KLblExample: "Ejemplo", KLblReset: "Restablecer", KLblTheme: "Tema", KLblThemeOption: "Tema: %s", KLblDraft: "Borrador",
	KLblShare: "Compartir vínculo", KLblPresent: "Presentar", KLblPreview: "Vista previa", KLblMinimap: "Minimapa",
	KLblSlides: "Diapositivas", KLblFiles: "Archivos", KLblHelp: "Ayuda", KLblNewSlide: "Nueva diapositiva", KLblImage: "Imagen",
	KLblBackground: "Fondo", KLblBgLeft: "Fondo dividido izq.", KLblBgRight: "Fondo dividido der.",
	KLblMath: "Matemáticas", KLblCode: "Bloque de código", KLblKind: "Kind", KLblHTML: "Página HTML", KLblPDF: "PDF",
	KLblPPTX: "PowerPoint", KLblODP: "OpenDocument",

	KTipExample:    "Empiece con uno de los ejemplos. Su texto se conserva en este navegador: Borrador ▸ Restaurar reemplazado lo recupera.",
	KTipTheme:      "El tema de la presentación: al elegir uno, se escribe en el front matter de la presentación.",
	KTipDraft:      "Restaure, olvide o restablezca la presentación que este navegador guarda para usted.",
	KTipReset:      "Recupera el texto del ejemplo; su presentación se conserva como borrador.",
	KTipPresent:    "Muestra aquí la presentación y abre la vista del presentador en su propia ventana.",
	KTipShare:      "Copia un vínculo que contiene la propia presentación: no se carga nada.",
	KTipKind:       "Asigna un kind a la diapositiva: una plantilla con su propio aspecto, definida por un comentario <!-- _kind: … -->.",
	KTipNewSlide:   "Añade una diapositiva vacía después de la actual.",
	KTipImage:      "Elija un archivo de imagen: se incluye en la presentación como un URI data:, no se carga nada. El texto seleccionado se convierte en su descripción.",
	KTipBackground: "Elija una imagen para el fondo de la diapositiva; la flecha rellena solo el 40 % izquierdo o derecho.",
	KTipMath:       "Una fórmula en bloque entre líneas $$; la selección se coloca dentro.",
	KTipCode:       "Un bloque de código delimitado con números de línea ({lineNumbers}); la selección se coloca dentro.",
	KTipMode:       "Vista general apila todas las diapositivas y se desplaza con el editor; Presentación es la propia presentación.",
	KTipMinimap:    "Un mapa de la presentación junto a la barra de desplazamiento del editor.",
	KTipThumbs:     "Una columna de miniaturas de diapositivas; haga clic en una para ir a ella.",
	KTipFiles:      "Los archivos del proyecto: temas, kinds, datos y medios. Una ventana estrecha se abre sin él.",
	KTipHelp:       "La ayuda sobre lo que hay bajo el cursor y los vínculos a la documentación. Una ventana estrecha se abre sin ella.",
	KTipHTML:       "Un único archivo HTML que presenta la presentación en cualquier navegador.",
	KTipPDF:        "Un PDF, una página por diapositiva, generado aquí en su navegador.",
	KTipPPTX:       "Un archivo PowerPoint editable, generado aquí en su navegador.",
	KTipODP:        "Una presentación OpenDocument editable, generada aquí en su navegador.",

	KModeOverview: "Vista general", KModeSlideshow: "Presentación",

	KExWelcome: "Bienvenida", KExTemplates: "Plantillas: diapositivas generadas", KExKinds: "Kinds de diapositiva",
	KExHCL: "Front matter HCL", KExGaia: "Temas y directivas locales", KExMedia: "Matemáticas, imágenes y fondos",
	KExGalleryDefault: "Galería: el tema default", KExGalleryGaia: "Galería: el tema gaia",
	KExGalleryUncover: "Galería: el tema uncover", KExGalleryMath: "Galería: matemáticas",
	KExGalleryCode:       "Galería: código, números de línea y llamadas",
	KExGalleryEmoji:      "Galería: emoji y títulos ajustados",
	KExGalleryPresenting: "Galería: transiciones y notas del orador",
	KExGalleryHCL:        "Galería: expresiones HCL y datos tipados",

	KStatusStarting:           "Iniciando…",
	KStatusOpenedShared:       "Se abrió la presentación compartida en el vínculo.",
	KStatusSharedUnreadable:   "No se pudo leer la presentación del vínculo: %s",
	"status.rendered.one":     "%d diapositiva · procesada en %d ms",
	"status.rendered.other":   "%d diapositivas · procesadas en %d ms",
	KStatusStale:              "la vista previa muestra la última presentación procesada",
	"status.problems.one":     "%d problema · %s",
	"status.problems.other":   "%d problemas · %s",
	KStatusDraftNotKept:       " · ⚠ borrador no guardado en este navegador",
	KStatusPresenting:         "Presentando: la vista del presentador está en su propia ventana; las teclas de flecha de cualquiera mueven ambas.",
	KStatusLinkCopied:         "Vínculo copiado: la presentación viaja en el propio vínculo (%d caracteres), no se carga nada.",
	KStatusSavedHTML:          "Se guardó deck.html: ábralo en cualquier navegador o presente desde él.",
	KStatusShareDeckOnly:      " Solo la presentación va en el vínculo: Proyecto ▸ Descargar proyecto (.zip) conserva sus archivos.",
	KStatusRestoredDraft:      "Se restauró su borrador de %s.",
	KStatusSharedKeptDraft:    "Se abrió la presentación compartida en el vínculo. Su borrador de %s se conserva: Borrador ▸ Restaurar mi borrador.",
	KStatusRestoredDraftReset: "Se restauró su borrador de %s. Restablecer vuelve al ejemplo.",
	KStatusDraftSavedAgain:    "Su borrador vuelve a guardarse en este navegador.",
	KStatusDraftNotSaved:      "⚠ Este navegador no guardó su borrador (%s): Compartir vínculo o Descargar HTML conserva la presentación.",
	KStatusOtherTab:           "Otra pestaña guardó su borrador %s. Sus cambios aquí no lo borrarán: Borrador ▸ Restaurar reemplazado.",
	KStatusRestoredTheDraft:   "Se restauró el borrador de %s.",
	KStatusForgot:             "Se olvidaron sus borradores: ya no se guarda nada de esta presentación en este navegador (hasta que la edite de nuevo).",
	KStatusNoExport:           "⚠ Esta página no puede exportar: use el comando slides.",
	KStatusExportRunning:      "Hay una exportación en curso: %s.",
	KStatusTooLargeExport:     "⚠ La presentación es demasiado grande para exportarla aquí: use el comando slides.",
	KStatusPreparing:          "Preparando %s…",
	KStatusLoadingEnginePct:   "Cargando el motor de exportación (una sola vez; su navegador lo conserva): %d %% de %.1f MiB",
	KStatusLoadingEngine:      "Cargando el motor de exportación (una sola vez; su navegador lo conserva): %.1f MiB",
	KStatusExportProgress:     "%s: diapositiva %d de %d…",
	KStatusExportFailed:       "⚠ Falló la exportación a %s: %s",
	"status.saved.one":        "Se guardó %s: %s, %d diapositiva en %.1f s, generada en su navegador.",
	"status.saved.other":      "Se guardó %s: %s, %d diapositivas en %.1f s, generadas en su navegador.",
	KStatusIsMedium:           "%s es un medio: la presentación lo muestra con %s (suéltelo en el editor o use su menú: Insertar en la presentación).",
	KStatusMadeTheme:          "Se creó %s, basado en %s: elija %s en la lista Tema para usarlo.",
	KStatusMadeKind:           "Se creó %s: una diapositiva toma el kind con <!-- _kind: %s -->.",
	KStatusMadeData:           "Se creó %s: la presentación lo lee como {{ .title }}.",
	KStatusCannotSaveFile:     "⚠ esta página no puede guardar un archivo",
	KStatusSavedZip:           "Se guardó %s: descomprimido, `slides build deck.md` lo compila y la extensión de VS Code lo abre.",
	KStatusOpenedZip:          "Se abrió el .zip: %d archivos.",
	KStatusOpenedFolder:       "Se abrió la carpeta: %d archivos.",
	KStatusAddedSlide:         "Se añadió una diapositiva después de la actual.",
	KStatusInsertedFormula:    "Se insertó una fórmula en bloque: escriba dentro.",
	KStatusInsertedCode:       "Se insertó un bloque de código Go con números de línea: escriba dentro.",
	KStatusPutFormula:         "La selección se colocó en una fórmula en bloque.",
	KStatusPutCode:            "La selección se colocó en un bloque de código Go con números de línea.",
	KStatusNoFilePicker:       "⚠ Esta página no puede pedir un archivo: suelte o pegue un archivo de imagen en el editor.",
	KStatusKindSet:            "La diapositiva ahora es del kind %s.",
	KStatusInsertedRule:       "Se insertó una línea horizontal (***: tres guiones en una línea iniciarían una nueva diapositiva).",
	KStatusInsertedTable:      "Se insertó una tabla: escriba en sus celdas; los comandos de alineación de tablas la mantienen ordenada.",
	KStatusJournalLarge:       "Este proyecto es demasiado grande para el guardado rápido: los cambios del último medio segundo antes de cerrar la página podrían no conservarse.",
	KStatusJournalRefused:     "⚠ Se rechazó el guardado rápido (%s): los cambios del último medio segundo antes de cerrar la página podrían no conservarse.",
	KStatusAdded:              "Se añadió %s al proyecto (%d KiB): se conserva en este navegador y va en el .zip del proyecto.",
	KStatusAddedPut:           "Se añadió %s al proyecto (%d KiB) y se colocó en la presentación: se conserva en este navegador y va en el .zip del proyecto.",
	KStatusPopupBlocked:       "⚠ El navegador bloqueó la ventana del presentador: permita las ventanas emergentes para esta página y vuelva a Presentar.",

	KWhyHeading:    "Un título va en una línea del texto de la presentación, no en el front matter ni en un bloque de código.",
	KWhyList:       "Una lista va en una línea del texto de la presentación, no en el front matter ni en un bloque de código.",
	KWhyQuote:      "Una cita va en una línea del texto de la presentación, no en el front matter ni en un bloque de código.",
	KWhyNotInTable: "Coloque primero el cursor en una tabla.",
	KWhyNotInBody:  "Coloque el cursor en una fila del cuerpo de la tabla: el encabezado se conserva.",
	KWhyLastColumn: "Una tabla conserva al menos una columna.",

	KProblemLine: "línea %d", KProblemError: "error", KProblemWarning: "advertencia",

	KConfirmReset:      "¿Reemplazar su presentación por el ejemplo? Su texto se conserva en este navegador: Borrador ▸ Restaurar reemplazado.",
	KConfirmShareFiles: "Un vínculo contiene solo la presentación, no sus otros archivos (temas, kinds, medios): ¿descargar el proyecto como .zip en su lugar?",
	KConfirmOpenZip:    "¿Reemplazar su presentación y sus archivos por los del .zip? Lo que tiene ahora se conserva en este navegador: Borrador ▸ Restaurar reemplazado.",
	KConfirmOpenFolder: "¿Reemplazar su presentación y sus archivos por los de la carpeta? Lo que tiene ahora se conserva en este navegador: Borrador ▸ Restaurar reemplazado.",

	KDateToday: "hoy a las %s", KDateAt: "%s a las %s", KDateShort: "%[2]d %[1]s",
	KDateMonths: "ene,feb,mar,abr,may,jun,jul,ago,sep,oct,nov,dic",
	KBytes:      "%d bytes", KWarn: "⚠ %s",

	KDraftRestoreMine: "Restaurar mi borrador (%s)", KDraftRestoreReplaced: "Restaurar reemplazado (%s)",
	KDraftForget: "Olvidar mis borradores", KDraftResetExample: "Restablecer el ejemplo", KDraftNothing: "Nada guardado aún",

	KSaveAs: "Guardar como…", KKindTitle: "Kind…",

	KProject: "Proyecto", KFilesTreeLabel: "Archivos del proyecto", KFilesNewThemeFrom: "Nuevo tema (a partir de %s)",
	KFilesNewKind: "Nuevo kind", KFilesNewData: "Nuevo archivo de datos", KFilesNewFile: "Nuevo archivo…", KFilesNewFolder: "Nueva carpeta…",
	KFilesAddMedia: "Añadir medios…", KFilesOpenZip: "Abrir proyecto (.zip)…", KFilesOpenFolder: "Abrir carpeta de proyecto…",
	KFilesDownloadZip: "Descargar proyecto (.zip)", KFilesInsertIntoDeck: "Insertar en la presentación", KFilesDownload: "Descargar",
	"files.usage.one": "%d archivo · %s de %d MiB", "files.usage.other": "%d archivos · %s de %d MiB",

	KPaneProblems: "Problemas", KPaneDocumentation: "Documentación",
	KFrameTitlePreview: "Vista previa de la diapositiva", KFrameTitleRail: "Miniaturas de diapositivas",

	KDocGettingStarted: "Primeros pasos", KDocMarp: "Compatibilidad con Marp", KDocTemplates: "Plantillas",
	KDocKinds: "Kinds", KDocHCL: "HCL", KDocCLI: "Línea de comandos", KDocVSCode: "VS Code", KDocPlayground: "Playground",

	KTableColumn: "Columna %d",

	KErrExists: "%s ya existe", KErrDeckName: "la presentación sigue siendo deck.md: el comando slides y el playground la leen con ese nombre",
	KErrNoSuchFile: "%s: no existe el archivo", KErrNotFileName: "%q no es un nombre de archivo",
	KErrDeckTooLarge: "la presentación pesa %d KiB; el playground procesa presentaciones de hasta %d KiB",
	KErrMaxFiles:     "un proyecto contiene como máximo %d archivos", KErrPathReason: "%s: %v",
	KErrProjectOver: "el proyecto superaría los %d MiB", KErrFileSize: "%s pesa %d KiB: un archivo pesa como máximo %d KiB",
	KErrShareTooLargeDeck: "la presentación es mayor de lo que admite el playground; descárguela en su lugar",
	KErrShareTooLargeLink: "la presentación es demasiado grande para viajar en un vínculo; descárguela en su lugar",
	KErrShareNoDeck:       "el vínculo no contiene ninguna presentación", KErrShareTooLong: "la presentación del vínculo es demasiado larga",
	KErrShareDamaged: "la presentación del vínculo está dañada (puede que se haya cortado)",
	KErrShareLarger:  "la presentación del vínculo es mayor de lo que admite el playground", KErrShareNotText: "la presentación del vínculo no es texto",
	KErrZipNoDeck: "el archivo no contiene ninguna presentación: ponga deck.md en su raíz (o un único archivo .md)",
	KErrZipSize:   "el archivo pesa %d MiB: como máximo %d MiB", KErrZipNot: "no es un archivo .zip: %v",
	KErrZipEntries: "el archivo enumera %d entradas: como máximo %d", KErrZipLinks: "%s: no se admiten los vínculos simbólicos",
	KErrZipFileOver: "%s supera los %d KiB: un archivo pesa como máximo eso",
	KErrZipBomb:     "el archivo se expande más de %d veces: rechazado como posible bomba zip",
	KErrZipTwice:    "%s aparece dos veces en el archivo", KErrZipMoreThan: "más de %d KiB",
	KErrFileOver: "%s supera %s", KErrFailed: "falló",
	KErrFolderCount: "la carpeta contiene %d archivos: un proyecto contiene como máximo %d", KErrFolderOver: "la carpeta supera %s",
	KErrFolderMore: "la carpeta contiene más de %d archivos", KErrNotImage: "%s no es una imagen.",

	"fmt.label.bold": "Negrita", "fmt.label.italic": "Cursiva", "fmt.label.underline": "Subrayado",
	"fmt.label.strike": "Tachado", "fmt.label.code": "Código en línea",
	"fmt.label.h1": "Título 1", "fmt.label.h2": "Título 2", "fmt.label.h3": "Título 3",
	"fmt.label.bullets": "Viñetas", "fmt.label.numbered": "Numeración", "fmt.label.quote": "Cita", "fmt.label.rule": "Línea",
	"fmt.label.table": "Tabla", "fmt.label.table-format": "Ordenar",
	"fmt.label.row-above": "Fila arriba", "fmt.label.row-below": "Fila abajo", "fmt.label.row-delete": "Eliminar fila",
	"fmt.label.col-left": "Columna izq.", "fmt.label.col-right": "Columna der.", "fmt.label.col-delete": "Eliminar columna",
	"fmt.label.colalign-left": "Columna a la izq.", "fmt.label.colalign-center": "Columna centrada",
	"fmt.label.colalign-right": "Columna a la der.",
	"fmt.label.slide-left":     "Texto a la izq.", "fmt.label.slide-center": "Texto centrado", "fmt.label.slide-right": "Texto a la der.",
	"fmt.label.link": "Vínculo",

	"fmt.tip.bold":            "Pone en negrita el texto seleccionado, o la palabra bajo el cursor: **texto**. Púlselo de nuevo para deshacerlo.",
	"fmt.tip.italic":          "Cursiva: _texto_. Púlselo de nuevo para deshacerlo.",
	"fmt.tip.underline":       "Markdown no tiene subrayado: esto encierra el texto en <u>…</u>, una etiqueta que la lista de HTML permitido de Marp conserva por defecto, así que funciona con el HTML sin procesar desactivado. Púlselo de nuevo para deshacerlo.",
	"fmt.tip.strike":          "Tachado: ~~texto~~. Púlselo de nuevo para deshacerlo.",
	"fmt.tip.code":            "Código en línea: `texto`. Púlselo de nuevo para deshacerlo.",
	"fmt.tip.h1":              "Convierte la línea (o cada línea seleccionada) en un título 1: # texto. Púlselo de nuevo para volver a texto normal.",
	"fmt.tip.h2":              "Título 2: ## texto.",
	"fmt.tip.h3":              "Título 3: ### texto.",
	"fmt.tip.bullets":         "Lista con viñetas: - elemento, uno por línea seleccionada. Púlselo de nuevo para deshacerlo.",
	"fmt.tip.numbered":        "Lista numerada: 1. elemento, 2. elemento… Púlselo de nuevo para deshacerlo.",
	"fmt.tip.quote":           "Cita: > texto. Púlselo de nuevo para quitar un nivel. Ctrl+Shift+.",
	"fmt.tip.rule":            "Una línea horizontal. Marp interpreta ---, *** y ___ como una nueva diapositiva, así que esto escribe la etiqueta <hr>, que la lista de HTML permitido conserva por defecto.",
	"fmt.tip.table":           "Inserta una tabla después de la línea actual: elija sus columnas y filas.",
	"fmt.tip.table-format":    "Alinea la tabla alrededor del cursor: cada columna se rellena hasta su celda más ancha, conservando las marcas de alineación.",
	"fmt.tip.row-above":       "Añade una fila vacía encima de la fila del cursor (bajo el encabezado, si está en el encabezado).",
	"fmt.tip.row-below":       "Añade una fila vacía debajo de la fila del cursor.",
	"fmt.tip.row-delete":      "Elimina la fila del cursor; el encabezado se conserva.",
	"fmt.tip.col-left":        "Añade una columna a la izquierda de la columna del cursor.",
	"fmt.tip.col-right":       "Añade una columna a la derecha de la columna del cursor.",
	"fmt.tip.col-delete":      "Elimina la columna del cursor; una tabla conserva una.",
	"fmt.tip.colalign-left":   "Alinea a la izquierda la columna de la tabla bajo el cursor (:--). Púlselo de nuevo para quitarlo.",
	"fmt.tip.colalign-center": "Centra la columna de la tabla bajo el cursor (:-:). Púlselo de nuevo para quitarlo.",
	"fmt.tip.colalign-right":  "Alinea a la derecha la columna de la tabla bajo el cursor (--:). Púlselo de nuevo para quitarlo.",
	"fmt.tip.slide-left":      "Devuelve el texto de la diapositiva a la izquierda, la alineación propia del tema: elimina el bloque de estilo que escribió esta pestaña.",
	"fmt.tip.slide-center":    "Centra el texto de la diapositiva donde está el cursor, con un bloque <style scoped> al principio de la diapositiva. Funciona con el HTML sin procesar desactivado y se ve en la vista previa, la página HTML y el PDF.",
	"fmt.tip.slide-right":     "Alinea a la derecha el texto de la diapositiva donde está el cursor (un bloque <style scoped>).",
	"fmt.tip.link":            "Convierte la selección, o la palabra, en un vínculo: [texto](url) con la url seleccionada para escribirla. En un vínculo, selecciona su url.",

	KGrpHistory: "Historial", KLblUndo: "Deshacer", KLblRedo: "Rehacer",
	KTipUndo:     "Revierte el último cambio: una edición del archivo mostrado o, si no, una operación sobre todo el proyecto.",
	KTipUndoStep: "Deshacer: %s.",
	KTipRedo:     "Vuelve a hacer lo que Deshacer revirtió.",
	KTipRedoStep: "Rehacer: %s.",
	KStepExample: "el ejemplo «%s»", KStepReset: "Restablecer", KStepOpenZip: "Abrir un .zip", KStepOpenFolder: "Abrir una carpeta",
	KStepRename: "Renombrar %s", KStepDelete: "Eliminar %s", KStepRestoreDraft: "Restaurar el borrador", KStepRestoreVersion: "Restaurar %s",
	KStatusUndone: "Deshecho: %s.", KStatusRedone: "Rehecho: %s.",

	KLblSaveVersion: "Guardar versión…", KLblVersions: "Versiones",
	KTipSaveVersion: "Conserva todo el proyecto tal como está ahora, con un nombre, en este navegador.",
	KTipVersions:    "Las versiones del proyecto, conservadas en este navegador: previsualizar una, compararla con el estado actual, restaurarla.",
	KAskVersionName: "Nombre de esta versión:",
	KVersionAuto:    "Automática", KVersionBefore: "Antes: %s", KVersionExport: "Exportada: %s", KVersionShared: "Vínculo compartido",
	KVersionLine:   "%s · %s · %s",
	KVersionsEmpty: "Aún no hay versiones. Se conserva una poco después de que deje de escribir, antes de una operación sobre todo el proyecto, en cada exportación o al compartir, y cuando guarda una.",
	KVersionsOff:   "Las versiones no se conservan: este navegador no guarda almacenamiento para la página.",
	KLblOnlyNamed:  "Solo con nombre", KLblVersionPreview: "Vista previa", KLblVersionCompare: "Comparar con ahora",
	KLblVersionRestore: "Restaurar", KLblVersionDelete: "Eliminar", KLblBackToNow: "Volver a ahora", KLblSideBySide: "En paralelo",
	KCompareTitle:    "%s comparada con ahora",
	KCompareFileLine: "%s (+%d −%d)",
	KCompareSame:     "Sin diferencias con ahora.",
	KCompareBinary:   "%s no es texto: %s entonces, %s ahora.",
	KCompareNow:      "Ahora",

	KStatusVersionSaved:    "Versión «%s» guardada.",
	KStatusVersionRestored: "%s restaurada; Deshacer la revierte.",
	KStatusVersionDeleted:  "Versión %s eliminada.",
	KStatusPreviewing:      "La vista previa muestra %s; Volver a ahora, o una edición, vuelve a mostrar la presentación.",
	KStatusVersionsFull:    "El almacenamiento de esta página está lleno: se descartaron las versiones automáticas más antiguas.",
	KStatusVersionsFailed:  "No se pudo conservar una versión: %s",
	KStatusVersionLimit:    "Se conservan como máximo %d versiones con nombre: elimine una para guardar otra.",
	KConfirmDeleteVersion:  "¿Eliminar la versión %s? Esta acción no se puede deshacer.",

	KHelpGeneral: `Escriba una presentación a la izquierda; las diapositivas aparecen a la derecha.

- Las diapositivas se separan con una línea de tres guiones: ---
- El front matter, entre las dos primeras líneas ---, contiene las directivas globales (theme, paginate, header, footer, size) y las claves propias de go-slides: kinds, data, template.
- <!-- _class: lead --> en una diapositiva es una directiva solo para esa diapositiva.
- {{ .title }}, {{ page }}/{{ total }}, {{ range seq 3 }}: plantillas de Go.
- kind: title aplica un kind (un conjunto de directivas con nombre).
- Matemáticas: $e^{i\pi}$ en línea, $$ ... $$ en líneas propias.
- Imágenes: ![w:300](url) cambia el tamaño de una, ![sepia](url) le aplica un filtro; ![bg](url) es un fondo, ![bg left:40%](url) uno dividido junto al texto.

Archivos, a la izquierda del editor, es el proyecto de la presentación, tal como lo lee el comando slides: deck.md, slides.yaml, themes/, kinds/, media/. Haga clic en un archivo para editarlo. Clic derecho (o Shift+F10) para Nuevo archivo, Nueva carpeta, Cambiar nombre (F2), Eliminar, Nuevo tema, Nuevo kind; arrastre un archivo a una carpeta para moverlo. El menú Proyecto crea un tema basado en default, gaia o uncover (se une a la lista Tema), un kind, un archivo de datos, y abre o descarga todo el proyecto como .zip (descomprimido, slides build deck.md lo compila).

Paneles: arrastre la barra entre dos paneles para cambiar su tamaño (Archivos, el editor, Diapositivas, la vista previa, Ayuda); haga doble clic en una barra, o pulse Intro sobre ella, para cerrar el panel y volver a abrirlo; F6 se mueve entre las barras, las teclas de flecha mueven una 16 px, Inicio y Fin van al mínimo y al máximo, Escape vuelve al editor. Los tamaños que establezca se conservan en este navegador; la pestaña Vista sigue permitiendo ocultar paneles.

Pestaña Formato: Negrita, Cursiva, Tachado y Código en línea actúan sobre la selección o la palabra bajo el cursor y alternan (Ctrl+B, Ctrl+I, Ctrl+Shift+X, Ctrl+E; Cmd en un Mac): cada botón aparece pulsado dentro de sus marcas, y un solo Deshacer revierte una acción. Subrayado (Ctrl+U) encierra el texto en <u>…</u>, porque Markdown no tiene subrayado y la lista de HTML de Marp por defecto conserva esa etiqueta. Los títulos 1 a 3 (Ctrl+Alt+1..3), las viñetas (Ctrl+Shift+8), la numeración (Ctrl+Shift+7) y la cita actúan sobre las líneas seleccionadas; la línea horizontal es una etiqueta <hr>, porque ---, *** y ___ iniciarían una nueva diapositiva. Tabla abre una cuadrícula de columnas por filas; con el cursor en una tabla puede añadir o quitar una fila o una columna, alinear la columna (:--, :-:, --:) y Ordenar mantiene las celdas alineadas. Texto de diapositiva a la izquierda, centrado o a la derecha escribe un bloque <style scoped> en la diapositiva actual: funciona con el HTML sin procesar desactivado y se ve en la vista previa, la página HTML y el PDF. Vínculo (Ctrl+K) escribe [texto](url) con la url seleccionada.

Suelte o pegue una imagen o un vídeo en el editor: va a media/ y la presentación lo nombra, ![](media/photo.png); suelte archivos en una carpeta del panel para ponerlos allí. Un archivo pesa como máximo 8 MiB y un proyecto 64 MiB (la línea bajo los archivos indica cuánto se usa). No se carga nada. Una imagen https: también funciona, pero su navegador la obtiene de su servidor, que entonces ve su dirección (no se envía referente).

Tema elige default, gaia, uncover o un tema que defina su presentación (/* @theme name */ en un <style>): escribe theme: en el front matter y sigue al front matter cuando usted lo edita.

La finalización se abre al escribir una directiva, un kind o dentro de {{ }}. Coloque el cursor sobre una directiva, un kind o una función de plantilla y este panel la explica. Haga clic en una diapositiva de la vista general para ir a su línea.

Presentar (o p en la presentación) abre la vista del presentador en una segunda ventana: la diapositiva actual, el paso siguiente, las notas del orador (<!-- comentarios -->), un reloj y un temporizador. Las teclas de flecha, Espacio, Inicio y Fin mueven ambas ventanas; la presentación y el presentador se siguen mutuamente, también tras ediciones. Permita las ventanas emergentes para este sitio si la ventana no se abre.

Su presentación se queda en este navegador: se guarda mientras escribe en el almacenamiento propio del navegador, solo para este sitio, y vuelve al recargar; no se envía nada a ninguna parte. Borrador restaura el borrador que un vínculo compartido, un ejemplo u otra pestaña reemplazó, restablece el ejemplo, y Olvidar mis borradores los borra. Compartir vínculo pone la presentación en el propio vínculo (no sus otros archivos: descargue el .zip del proyecto para ellos).

Deshacer y Rehacer (pestaña Inicio; Ctrl+Z, y Ctrl+Shift+Z o Ctrl+Y; Cmd en un Mac) revierten la última edición del archivo mostrado, cada archivo con la suya, y después la última operación sobre todo el proyecto: un ejemplo, Restablecer, un .zip o una carpeta abiertos, un borrador restaurado, un archivo eliminado o renombrado, una versión restaurada. Rehacer devuelve todo como estaba cuando pulsó Deshacer.

Las versiones conservan todo el proyecto en este navegador: se guarda una poco después de que deje de escribir, antes de una operación sobre todo el proyecto, en cada exportación y al compartir, y Guardar versión… conserva una con un nombre. El panel Versiones las lista: Vista previa muestra una en la vista previa, Comparar con ahora muestra los cambios de cada archivo línea por línea (F7 va al siguiente cambio), Restaurar la devuelve a su lugar y Deshacer revierte la restauración. Se conservan las versiones automáticas más recientes, luego una por hora durante un día y una por día durante un mes; las que tienen nombre permanecen hasta que las elimine.

Las pestañas de la cinta de opciones: Inicio (Deshacer, Rehacer, versiones, ejemplo, tema, Presentar, Compartir vínculo, Borrador), Insertar (una nueva diapositiva, una imagen o un fondo desde un archivo, matemáticas, un bloque de código con números de línea, un kind: cada uno se escribe en la presentación en el cursor, alrededor de la selección si la hay, y Deshacer lo revierte), Vista (la vista previa, el minimapa, las miniaturas de diapositivas) y Exportar. Alt muestra una letra sobre cada pestaña; escríbala y luego la letra sobre un control. Ctrl+F1 contrae la cinta de opciones.

Pestaña Exportar: la página HTML de la presentación, con sus medios dentro; un PDF (vectorial, con texto seleccionable), una presentación PowerPoint u OpenDocument (editables), generados en su navegador exactamente como los genera el comando slides. La primera exportación descarga el motor de exportación (unos 5 MiB, una sola vez: su navegador lo conserva), compilándolo a medida que llega, y después solo las fuentes que usan sus diapositivas; las imágenes https: se obtienen solo si su servidor lo permite (CORS), las imágenes data: siempre.`,
}

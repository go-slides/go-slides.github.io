package playground

// ptMessages is the Brazilian Portuguese catalogue.
var ptMessages = map[Key]string{
	KTabHome: "Página Inicial", KTabInsert: "Inserir", KTabFormat: "Formatar", KTabView: "Exibir", KTabExport: "Exportar",

	KGrpDeck: "Apresentação", KGrpLook: "Aparência", KGrpPresent: "Apresentar", KGrpDraft: "Rascunho", KGrpSlides: "Slides",
	KGrpPictures: "Imagens", KGrpText: "Texto", KGrpPreview: "Visualização", KGrpEditor: "Editor", KGrpSaveAs: "Salvar como",
	KGrpParagraph: "Parágrafo", KGrpTable: "Tabela", KGrpRowsCols: "Linhas e colunas", KGrpColumn: "Coluna",
	KGrpSlideText: "Texto do slide", KGrpLink: "Link",

	KLblExample: "Exemplo", KLblReset: "Redefinir", KLblTheme: "Tema", KLblThemeOption: "Tema: %s", KLblDraft: "Rascunho",
	KLblShare: "Compartilhar", KLblPresent: "Apresentar", KLblPreview: "Visualizar", KLblMinimap: "Minimapa",
	KLblSlides: "Slides", KLblFiles: "Arquivos", KLblHelp: "Ajuda", KLblNewSlide: "Novo slide", KLblImage: "Imagem",
	KLblBackground: "Plano de fundo", KLblBgLeft: "Fundo à esquerda", KLblBgRight: "Fundo à direita",
	KLblMath: "Equação", KLblCode: "Bloco de código", KLblKind: "Kind", KLblHTML: "Página HTML", KLblPDF: "PDF",
	KLblPPTX: "PowerPoint", KLblODP: "OpenDocument",

	KTipExample:    "Comece por um dos exemplos. Seu texto permanece neste navegador: Rascunho › Restaurar substituído o traz de volta.",
	KTipTheme:      "O tema da apresentação: ao escolher um, ele é gravado no front matter da apresentação.",
	KTipDraft:      "Restaurar, esquecer ou redefinir a apresentação que este navegador guarda para você.",
	KTipReset:      "Restaura o texto do exemplo; sua apresentação é mantida como rascunho.",
	KTipPresent:    "Exibe a apresentação aqui e abre o modo de apresentador em sua própria janela.",
	KTipShare:      "Copia um link que contém a própria apresentação: nada é enviado.",
	KTipKind:       "Dá um kind ao slide: um modelo com visual próprio, definido por um comentário <!-- _kind: … -->.",
	KTipNewSlide:   "Adiciona um slide vazio depois do atual.",
	KTipImage:      "Escolha um arquivo de imagem: ele entra na apresentação como um URI data:, nada é enviado. O texto selecionado vira a descrição.",
	KTipBackground: "Escolha uma imagem para o plano de fundo do slide; a seta preenche apenas 40% da esquerda ou da direita.",
	KTipMath:       "Uma fórmula em destaque entre linhas $$; a seleção vai dentro dela.",
	KTipCode:       "Um bloco de código delimitado com números de linha ({lineNumbers}); a seleção vai dentro dele.",
	KTipMode:       "A visão geral empilha todos os slides e rola junto com o editor; a apresentação de slides é a apresentação em si.",
	KTipMinimap:    "Um mapa da apresentação ao lado da barra de rolagem do editor.",
	KTipThumbs:     "Uma coluna de miniaturas dos slides; clique em uma para ir até ela.",
	KTipFiles:      "Os arquivos do projeto: temas, kinds, dados e mídia. Uma janela estreita começa sem ele.",
	KTipHelp:       "A ajuda sobre o que está sob o cursor e os links da documentação. Uma janela estreita começa sem ela.",
	KTipHTML:       "Um arquivo HTML que apresenta a apresentação em qualquer navegador.",
	KTipPDF:        "Um PDF com uma página por slide, gerado aqui no seu navegador.",
	KTipPPTX:       "Um arquivo PowerPoint editável, gerado aqui no seu navegador.",
	KTipODP:        "Uma apresentação OpenDocument editável, gerada aqui no seu navegador.",

	KModeOverview: "Visão geral", KModeSlideshow: "Apresentação",

	KExWelcome: "Boas-vindas", KExTemplates: "Modelos: slides gerados", KExKinds: "Kinds de slide",
	KExHCL: "Front matter HCL", KExGaia: "Temas e diretivas locais", KExMedia: "Equações, imagens e planos de fundo",
	KExGalleryDefault: "Galeria: o tema padrão", KExGalleryGaia: "Galeria: o tema gaia",
	KExGalleryUncover: "Galeria: o tema uncover", KExGalleryMath: "Galeria: equações",
	KExGalleryCode:       "Galeria: código, números de linha e destaques",
	KExGalleryEmoji:      "Galeria: emojis e títulos ajustados",
	KExGalleryPresenting: "Galeria: transições e anotações do apresentador",
	KExGalleryHCL:        "Galeria: expressões HCL e dados tipados",

	KStatusStarting:           "Iniciando…",
	KStatusOpenedShared:       "A apresentação compartilhada no link foi aberta.",
	KStatusSharedUnreadable:   "Não foi possível ler a apresentação do link: %s",
	"status.rendered.one":     "%d slide · renderizado em %d ms",
	"status.rendered.other":   "%d slides · renderizados em %d ms",
	KStatusStale:              "a visualização mostra a última apresentação renderizada",
	"status.problems.one":     "%d problema · %s",
	"status.problems.other":   "%d problemas · %s",
	KStatusDraftNotKept:       " · ⚠ rascunho não guardado neste navegador",
	KStatusPresenting:         "Apresentando: o modo de apresentador está em sua própria janela; as teclas de seta em qualquer uma movem as duas.",
	KStatusLinkCopied:         "Link copiado: a apresentação viaja no próprio link (%d caracteres), nada é enviado.",
	KStatusSavedHTML:          "deck.html salvo: abra-o em qualquer navegador ou apresente a partir dele.",
	KStatusShareDeckOnly:      " Apenas a apresentação está no link: Projeto › Baixar projeto (.zip) mantém os arquivos.",
	KStatusRestoredDraft:      "Seu rascunho de %s foi restaurado.",
	KStatusSharedKeptDraft:    "A apresentação compartilhada no link foi aberta. Seu rascunho de %s foi mantido: Rascunho › Restaurar meu rascunho.",
	KStatusRestoredDraftReset: "Seu rascunho de %s foi restaurado. Redefinir volta ao exemplo.",
	KStatusDraftSavedAgain:    "Seu rascunho voltou a ser salvo neste navegador.",
	KStatusDraftNotSaved:      "⚠ Este navegador não guardou seu rascunho (%s): Compartilhar ou Baixar HTML mantém a apresentação.",
	KStatusOtherTab:           "Outra guia salvou seu rascunho %s. Suas edições aqui não o apagarão: Rascunho › Restaurar substituído.",
	KStatusRestoredTheDraft:   "O rascunho de %s foi restaurado.",
	KStatusForgot:             "Seus rascunhos foram esquecidos: nada desta apresentação está guardado neste navegador agora (até você editá-la de novo).",
	KStatusNoExport:           "⚠ Esta página não pode exportar: use o comando slides.",
	KStatusExportRunning:      "Uma exportação está em andamento: %s.",
	KStatusTooLargeExport:     "⚠ A apresentação é grande demais para exportar aqui: use o comando slides.",
	KStatusPreparing:          "Preparando %s…",
	KStatusLoadingEnginePct:   "Carregando o mecanismo de exportação (uma vez; seu navegador o guarda): %d%% de %.1f MiB",
	KStatusLoadingEngine:      "Carregando o mecanismo de exportação (uma vez; seu navegador o guarda): %.1f MiB",
	KStatusExportProgress:     "%s: slide %d de %d…",
	KStatusExportFailed:       "⚠ A exportação de %s falhou: %s",
	"status.saved.one":        "%s salvo: %s, %d slide em %.1f s, gerado no seu navegador.",
	"status.saved.other":      "%s salvo: %s, %d slides em %.1f s, gerado no seu navegador.",
	KStatusIsMedium:           "%s é uma mídia: a apresentação a exibe com %s (solte-a no editor ou use o menu dela: Inserir na apresentação).",
	KStatusMadeTheme:          "%s criado, baseado em %s: escolha %s na lista de temas para usá-lo.",
	KStatusMadeKind:           "%s criado: um slide assume o kind com <!-- _kind: %s -->.",
	KStatusMadeData:           "%s criado: a apresentação o lê como {{ .title }}.",
	KStatusCannotSaveFile:     "⚠ esta página não pode salvar um arquivo",
	KStatusSavedZip:           "%s salvo: descompactado, `slides build deck.md` o gera, e a extensão do VS Code o abre.",
	KStatusOpenedZip:          "O .zip foi aberto: %d arquivos.",
	KStatusOpenedFolder:       "A pasta foi aberta: %d arquivos.",
	KStatusAddedSlide:         "Slide adicionado depois do atual.",
	KStatusInsertedFormula:    "Fórmula em destaque inserida: digite dentro dela.",
	KStatusInsertedCode:       "Bloco de código Go com números de linha inserido: digite dentro dele.",
	KStatusPutFormula:         "A seleção foi colocada em uma fórmula em destaque.",
	KStatusPutCode:            "A seleção foi colocada em um bloco de código Go com números de linha.",
	KStatusNoFilePicker:       "⚠ Esta página não pode pedir um arquivo: solte ou cole um arquivo de imagem no editor.",
	KStatusKindSet:            "O kind do slide agora é %s.",
	KStatusInsertedRule:       "Linha horizontal inserida (***: três hifens em uma linha iniciariam um novo slide).",
	KStatusInsertedTable:      "Tabela inserida: digite nas células; os comandos de alinhamento da tabela a mantêm organizada.",
	KStatusJournalLarge:       "Este projeto é grande demais para o salvamento rápido: as edições dos últimos meio segundo antes de fechar a página podem não ser guardadas.",
	KStatusJournalRefused:     "⚠ O salvamento rápido foi recusado (%s): as edições dos últimos meio segundo antes de fechar a página podem não ser guardadas.",
	KStatusAdded:              "%s adicionado ao projeto (%d KiB): ele fica neste navegador e entra no .zip do projeto.",
	KStatusAddedPut:           "%s adicionado ao projeto (%d KiB) e colocado na apresentação: ele fica neste navegador e entra no .zip do projeto.",
	KStatusPopupBlocked:       "⚠ O navegador bloqueou a janela do apresentador: permita pop-ups nesta página e clique em Apresentar de novo.",

	KWhyHeading:    "Um título vai em uma linha do texto da apresentação, não no front matter nem em um bloco de código.",
	KWhyList:       "Uma lista vai em uma linha do texto da apresentação, não no front matter nem em um bloco de código.",
	KWhyQuote:      "Uma citação vai em uma linha do texto da apresentação, não no front matter nem em um bloco de código.",
	KWhyNotInTable: "Coloque o cursor em uma tabela primeiro.",
	KWhyNotInBody:  "Coloque o cursor em uma linha do corpo da tabela: o cabeçalho permanece.",
	KWhyLastColumn: "Uma tabela mantém pelo menos uma coluna.",

	KProblemLine: "linha %d", KProblemError: "erro", KProblemWarning: "aviso",

	KConfirmReset:      "Substituir sua apresentação pelo exemplo? Seu texto permanece neste navegador: Rascunho › Restaurar substituído.",
	KConfirmShareFiles: "Um link contém apenas a apresentação, não os outros arquivos (temas, kinds, mídia): baixar o projeto como .zip?",
	KConfirmOpenZip:    "Substituir sua apresentação e seus arquivos pelos do .zip? O que você tem agora permanece neste navegador: Rascunho › Restaurar substituído.",
	KConfirmOpenFolder: "Substituir sua apresentação e seus arquivos pelos da pasta? O que você tem agora permanece neste navegador: Rascunho › Restaurar substituído.",

	KDateToday: "hoje às %s", KDateAt: "%s às %s", KDateShort: "%[2]d %[1]s",
	KDateMonths: "jan,fev,mar,abr,mai,jun,jul,ago,set,out,nov,dez",
	KBytes:      "%d bytes", KWarn: "⚠ %s",

	KDraftRestoreMine: "Restaurar meu rascunho (%s)", KDraftRestoreReplaced: "Restaurar substituído (%s)",
	KDraftForget: "Esquecer meus rascunhos", KDraftResetExample: "Redefinir para o exemplo", KDraftNothing: "Nada salvo ainda",

	KSaveAs: "Salvar como…", KKindTitle: "Kind…",

	KProject: "Projeto", KFilesTreeLabel: "Arquivos do projeto", KFilesNewThemeFrom: "Novo tema (a partir de %s)",
	KFilesNewKind: "Novo kind", KFilesNewData: "Novo arquivo de dados", KFilesNewFile: "Novo arquivo…", KFilesNewFolder: "Nova pasta…",
	KFilesAddMedia: "Adicionar mídia…", KFilesOpenZip: "Abrir projeto (.zip)…", KFilesOpenFolder: "Abrir pasta do projeto…",
	KFilesDownloadZip: "Baixar projeto (.zip)", KFilesInsertIntoDeck: "Inserir na apresentação", KFilesDownload: "Baixar",
	"files.usage.one": "%d arquivo · %s de %d MiB", "files.usage.other": "%d arquivos · %s de %d MiB",

	KPaneProblems: "Problemas", KPaneDocumentation: "Documentação",
	KFrameTitlePreview: "Visualização do slide", KFrameTitleRail: "Miniaturas dos slides",

	KDocGettingStarted: "Primeiros passos", KDocMarp: "Compatibilidade com Marp", KDocTemplates: "Modelos",
	KDocKinds: "Kinds", KDocHCL: "HCL", KDocCLI: "Linha de comando", KDocVSCode: "VS Code", KDocPlayground: "Playground",

	KTableColumn: "Coluna %d",

	KErrExists: "%s já existe", KErrDeckName: "a apresentação continua sendo deck.md: o comando slides e o playground a leem com esse nome",
	KErrNoSuchFile: "%s: arquivo não encontrado", KErrNotFileName: "%q não é um nome de arquivo",
	KErrDeckTooLarge: "a apresentação tem %d KiB; o playground renderiza apresentações de até %d KiB",
	KErrMaxFiles:     "um projeto contém no máximo %d arquivos", KErrPathReason: "%s: %v",
	KErrProjectOver: "o projeto passaria de %d MiB", KErrFileSize: "%s tem %d KiB: um arquivo tem no máximo %d KiB",
	KErrShareTooLargeDeck: "a apresentação é maior do que o playground aceita; baixe-a em vez disso",
	KErrShareTooLargeLink: "a apresentação é grande demais para viajar em um link; baixe-a em vez disso",
	KErrShareNoDeck:       "o link não contém apresentação", KErrShareTooLong: "a apresentação do link é longa demais",
	KErrShareDamaged: "a apresentação do link está danificada (pode ter sido cortada)",
	KErrShareLarger:  "a apresentação do link é maior do que o playground aceita", KErrShareNotText: "a apresentação do link não é texto",
	KErrZipNoDeck: "o arquivo compactado não contém apresentação: coloque deck.md na raiz (ou um arquivo .md)",
	KErrZipSize:   "o arquivo compactado tem %d MiB: no máximo %d MiB", KErrZipNot: "não é um arquivo .zip: %v",
	KErrZipEntries: "o arquivo compactado lista %d entradas: no máximo %d", KErrZipLinks: "%s: links não são aceitos",
	KErrZipFileOver: "%s passa de %d KiB: é o máximo para um arquivo",
	KErrZipBomb:     "o arquivo compactado se expande mais de %d vezes: recusado como possível zip bomb",
	KErrZipTwice:    "%s aparece duas vezes no arquivo compactado", KErrZipMoreThan: "mais de %d KiB",
	KErrFileOver: "%s passa de %s", KErrFailed: "falhou",
	KErrFolderCount: "a pasta contém %d arquivos: um projeto contém no máximo %d", KErrFolderOver: "a pasta passa de %s",
	KErrFolderMore: "a pasta contém mais de %d arquivos", KErrNotImage: "%s não é uma imagem.",

	"fmt.label.bold": "Negrito", "fmt.label.italic": "Itálico", "fmt.label.underline": "Sublinhado",
	"fmt.label.strike": "Tachado", "fmt.label.code": "Código",
	"fmt.label.h1": "Título 1", "fmt.label.h2": "Título 2", "fmt.label.h3": "Título 3",
	"fmt.label.bullets": "Marcadores", "fmt.label.numbered": "Numeração", "fmt.label.quote": "Citação", "fmt.label.rule": "Linha",
	"fmt.label.table": "Tabela", "fmt.label.table-format": "Organizar",
	"fmt.label.row-above": "Linha acima", "fmt.label.row-below": "Linha abaixo", "fmt.label.row-delete": "Excluir linha",
	"fmt.label.col-left": "Coluna à esq.", "fmt.label.col-right": "Coluna à dir.", "fmt.label.col-delete": "Excluir coluna",
	"fmt.label.colalign-left": "Coluna à esq.", "fmt.label.colalign-center": "Coluna centr.",
	"fmt.label.colalign-right": "Coluna à dir.",
	"fmt.label.slide-left":     "Texto à esq.", "fmt.label.slide-center": "Texto centr.", "fmt.label.slide-right": "Texto à dir.",
	"fmt.label.link": "Link",

	"fmt.tip.bold":            "Coloca em negrito o texto selecionado, ou a palavra sob o cursor: **texto**. Pressione de novo para desfazer.",
	"fmt.tip.italic":          "Itálico: _texto_. Pressione de novo para desfazer.",
	"fmt.tip.underline":       "O Markdown não tem sublinhado: isto envolve o texto em <u>…</u>, uma tag que a lista de HTML permitido padrão do Marp mantém, então funciona com o HTML bruto desativado. Pressione de novo para desfazer.",
	"fmt.tip.strike":          "Tachado: ~~texto~~. Pressione de novo para desfazer.",
	"fmt.tip.code":            "Código em linha: `texto`. Pressione de novo para desfazer.",
	"fmt.tip.h1":              "Transforma a linha (ou cada linha selecionada) em título 1: # texto. Pressione de novo para voltar ao texto simples.",
	"fmt.tip.h2":              "Título 2: ## texto.",
	"fmt.tip.h3":              "Título 3: ### texto.",
	"fmt.tip.bullets":         "Lista com marcadores: - item, um por linha selecionada. Pressione de novo para desfazer.",
	"fmt.tip.numbered":        "Lista numerada: 1. item, 2. item… Pressione de novo para desfazer.",
	"fmt.tip.quote":           "Citação: > texto. Pressione de novo para remover um nível. Ctrl+Shift+.",
	"fmt.tip.rule":            "Uma linha horizontal. O Marp lê ---, *** e ___ como um novo slide, então isto escreve a tag <hr>, que a lista de HTML permitido padrão mantém.",
	"fmt.tip.table":           "Insere uma tabela depois da linha atual: escolha suas colunas e linhas.",
	"fmt.tip.table-format":    "Alinha a tabela sob o cursor: cada coluna é preenchida até sua maior célula, e as marcas de alinhamento são mantidas.",
	"fmt.tip.row-above":       "Adiciona uma linha vazia acima da linha do cursor (abaixo do cabeçalho, se o cursor estiver nele).",
	"fmt.tip.row-below":       "Adiciona uma linha vazia abaixo da linha do cursor.",
	"fmt.tip.row-delete":      "Remove a linha do cursor; o cabeçalho permanece.",
	"fmt.tip.col-left":        "Adiciona uma coluna à esquerda da coluna do cursor.",
	"fmt.tip.col-right":       "Adiciona uma coluna à direita da coluna do cursor.",
	"fmt.tip.col-delete":      "Remove a coluna do cursor; a tabela mantém uma.",
	"fmt.tip.colalign-left":   "Alinha à esquerda a coluna da tabela sob o cursor (:--). Pressione de novo para limpar.",
	"fmt.tip.colalign-center": "Centraliza a coluna da tabela sob o cursor (:-:). Pressione de novo para limpar.",
	"fmt.tip.colalign-right":  "Alinha à direita a coluna da tabela sob o cursor (--:). Pressione de novo para limpar.",
	"fmt.tip.slide-left":      "Devolve o texto do slide à esquerda, como no tema: remove o bloco de estilo que esta guia escreveu.",
	"fmt.tip.slide-center":    "Centraliza o texto do slide em que o cursor está, com um bloco <style scoped> no início do slide. Funciona com o HTML bruto desativado e aparece na visualização, na página HTML e no PDF.",
	"fmt.tip.slide-right":     "Alinha à direita o texto do slide em que o cursor está (um bloco <style scoped>).",
	"fmt.tip.link":            "Transforma a seleção, ou a palavra, em um link: [texto](url) com a url selecionada para digitar. Em um link, seleciona a url dele.",

	KGrpHistory: "Histórico", KLblUndo: "Desfazer", KLblRedo: "Refazer",
	KTipUndo:     "Desfaz a última alteração: uma edição do arquivo exibido ou, se não houver, uma operação sobre todo o projeto.",
	KTipUndoStep: "Desfazer: %s.",
	KTipRedo:     "Refaz o que Desfazer desfez.",
	KTipRedoStep: "Refazer: %s.",
	KStepExample: "o exemplo “%s”", KStepReset: "Redefinir", KStepOpenZip: "Abrir um .zip", KStepOpenFolder: "Abrir uma pasta",
	KStepRename: "Renomear %s", KStepDelete: "Excluir %s", KStepRestoreDraft: "Restaurar o rascunho", KStepRestoreVersion: "Restaurar %s",
	KStatusUndone: "Desfeito: %s.", KStatusRedone: "Refeito: %s.",

	KLblSaveVersion: "Salvar versão…", KLblVersions: "Versões",
	KTipSaveVersion: "Guarda todo o projeto como está agora, com um nome, neste navegador.",
	KTipVersions:    "As versões do projeto, guardadas neste navegador: visualizar uma, compará-la com agora, restaurá-la.",
	KAskVersionName: "Nome desta versão:",
	KVersionAuto:    "Automática", KVersionBefore: "Antes: %s", KVersionExport: "Exportada: %s", KVersionShared: "Link compartilhado",
	KVersionLine:   "%s · %s · %s",
	KVersionsEmpty: "Nenhuma versão ainda. Uma é guardada pouco depois de você parar de digitar, antes de uma operação sobre todo o projeto, a cada exportação ou compartilhamento, e quando você salva uma.",
	KVersionsOff:   "As versões não são guardadas: este navegador não oferece armazenamento para a página.",
	KLblOnlyNamed:  "Só com nome", KLblVersionPreview: "Visualizar", KLblVersionCompare: "Comparar com agora",
	KLblVersionRestore: "Restaurar", KLblVersionDelete: "Excluir", KLblBackToNow: "Voltar para agora", KLblSideBySide: "Lado a lado",
	KCompareTitle:    "%s comparada com agora",
	KCompareFileLine: "%s (+%d −%d)",
	KCompareSame:     "Nenhuma diferença em relação a agora.",
	KCompareBinary:   "%s não é texto: %s então, %s agora.",
	KCompareNow:      "Agora",

	KStatusVersionSaved:    "Versão “%s” salva.",
	KStatusVersionRestored: "%s restaurada; Desfazer a desfaz.",
	KStatusVersionDeleted:  "Versão %s excluída.",
	KStatusPreviewing:      "A visualização mostra %s; Voltar para agora, ou uma edição, mostra a apresentação de novo.",
	KStatusVersionsFull:    "O armazenamento desta página está cheio: as versões automáticas mais antigas foram descartadas.",
	KStatusVersionsFailed:  "Não foi possível guardar uma versão: %s",
	KStatusVersionLimit:    "No máximo %d versões com nome são guardadas: exclua uma para salvar outra.",
	KConfirmDeleteVersion:  "Excluir a versão %s? Isso não pode ser desfeito.",

	KHelpGeneral: `Digite uma apresentação à esquerda; os slides aparecem à direita.

- Os slides são separados por uma linha com três hifens: ---
- O front matter, entre as duas primeiras linhas ---, contém as diretivas globais (theme, paginate, header, footer, size) e as chaves próprias do go-slides: kinds, data, template.
- <!-- _class: lead --> em um slide é uma diretiva só para aquele slide.
- {{ .title }}, {{ page }}/{{ total }}, {{ range seq 3 }}: templates Go.
- kind: title aplica um kind (um conjunto nomeado de diretivas).
- Equações: $e^{i\pi}$ em linha, $$ ... $$ em linhas próprias.
- Imagens: ![w:300](url) define o tamanho, ![sepia](url) aplica um filtro; ![bg](url) é um plano de fundo, ![bg left:40%](url) um dividido ao lado do texto.

Arquivos, à esquerda do editor, é o projeto da apresentação, como o comando slides o lê: deck.md, slides.yaml, themes/, kinds/, media/. Clique em um arquivo para editá-lo. Clique com o botão direito (ou Shift+F10) para Novo arquivo, Nova pasta, Renomear (F2), Excluir, Novo tema, Novo kind; arraste um arquivo para uma pasta para movê-lo. O menu Projeto cria um tema baseado em default, gaia ou uncover (ele entra na lista de temas), um kind, um arquivo de dados, e abre ou baixa o projeto inteiro como .zip (descompactado, slides build deck.md o gera).

Painéis: arraste a barra entre dois painéis para redimensioná-los (Arquivos, o editor, Slides, a visualização, Ajuda); clique duas vezes em uma barra, ou pressione Enter nela, para fechar o painel e reabri-lo; F6 alterna entre as barras, as teclas de seta movem uma em 16 px, Home e End vão para o menor e o maior tamanho, Escape volta ao editor. Os tamanhos definidos ficam guardados neste navegador; a guia Exibir ainda desativa painéis.

Guia Formatar: Negrito, Itálico, Tachado e Código funcionam na seleção ou na palavra sob o cursor e alternam (Ctrl+B, Ctrl+I, Ctrl+Shift+X, Ctrl+E; Cmd em um Mac): cada botão aparece pressionado dentro de suas marcas, e um Desfazer desfaz uma ação. Sublinhado (Ctrl+U) envolve o texto em <u>…</u>, porque o Markdown não tem sublinhado e a lista de HTML padrão do Marp mantém essa tag. Títulos 1 a 3 (Ctrl+Alt+1..3), marcadores (Ctrl+Shift+8), numeração (Ctrl+Shift+7) e citação funcionam nas linhas selecionadas; a linha é uma tag <hr>, porque ---, *** e ___ iniciariam um novo slide. Tabela abre uma grade de colunas por linhas; com o cursor em uma tabela, você adiciona ou remove uma linha ou uma coluna, alinha a coluna (:--, :-:, --:) e Organizar mantém as células alinhadas. Texto à esquerda, centralizado ou à direita escreve um bloco <style scoped> no slide atual: funciona com o HTML bruto desativado e aparece na visualização, na página HTML e no PDF. Link (Ctrl+K) escreve [texto](url) com a url selecionada.

Solte ou cole uma imagem ou um vídeo no editor: ele vai para media/ e a apresentação o referencia, ![](media/photo.png); solte arquivos em uma pasta do painel para colocá-los lá. Um arquivo tem no máximo 8 MiB e um projeto 64 MiB (a linha sob os arquivos mostra quanto está em uso). Nada é enviado. Uma imagem https: também funciona, mas seu navegador a busca no servidor dela, que então vê seu endereço (nenhum referenciador é enviado).

Tema escolhe default, gaia, uncover ou um tema que sua apresentação define (/* @theme name */ em um <style>): ele grava theme: no front matter e acompanha o front matter quando você o edita.

A conclusão abre enquanto você digita uma diretiva, um kind ou dentro de {{ }}. Coloque o cursor sobre uma diretiva, um kind ou uma função de template e este painel a explica. Clique em um slide na visão geral para ir até a linha dele.

Apresentar (ou p na apresentação) abre o modo de apresentador em uma segunda janela: o slide atual, a próxima etapa, as anotações do apresentador (<!-- comentários -->), um relógio e um cronômetro. As teclas de seta, Espaço, Home e End movem as duas janelas; a apresentação e o apresentador se acompanham, inclusive entre edições. Permita pop-ups neste site se a janela não abrir.

Sua apresentação permanece neste navegador: ela é salva enquanto você digita no armazenamento próprio do navegador, só para este site, e volta quando você recarrega; nada é enviado a lugar nenhum. Rascunho restaura o rascunho que um link compartilhado, um exemplo ou outra guia substituiu, redefine para o exemplo, e Esquecer meus rascunhos os apaga. Compartilhar coloca a apresentação no próprio link (não os outros arquivos: baixe o .zip do projeto para eles).

Desfazer e Refazer (guia Página Inicial; Ctrl+Z, e Ctrl+Shift+Z ou Ctrl+Y; Cmd em um Mac) desfazem a última edição do arquivo exibido, cada arquivo com a sua, e depois a última operação sobre todo o projeto: um exemplo, Redefinir, um .zip ou uma pasta aberta, um rascunho restaurado, um arquivo excluído ou renomeado, uma versão restaurada. Refazer traz tudo de volta como estava quando você pressionou Desfazer.

As versões guardam todo o projeto neste navegador: uma é salva pouco depois de você parar de digitar, antes de uma operação sobre todo o projeto, a cada exportação e compartilhamento, e Salvar versão… guarda uma com um nome. O painel Versões as lista: Visualizar mostra uma na visualização, Comparar com agora mostra as alterações de cada arquivo linha a linha (F7 vai para a próxima alteração), Restaurar a coloca de volta e Desfazer desfaz a restauração. As versões automáticas mais recentes são mantidas, depois uma por hora durante um dia e uma por dia durante um mês; as com nome ficam até você excluí-las.

As guias da faixa de opções: Página Inicial (Desfazer, Refazer, versões, exemplo, tema, Apresentar, Compartilhar, Rascunho), Inserir (um novo slide, uma imagem ou plano de fundo de um arquivo, equação, um bloco de código com números de linha, um kind: cada um é digitado na apresentação no cursor, ao redor da seleção se houver uma, e Desfazer o remove), Exibir (a visualização, o minimapa, as miniaturas dos slides) e Exportar. Alt mostra uma letra sobre cada guia; digite-a e depois a letra sobre um controle. Ctrl+F1 recolhe a faixa de opções.

Guia Exportar: a página HTML da apresentação, com a mídia dentro; um PDF (vetorial, texto selecionável), uma apresentação PowerPoint ou OpenDocument (editável), gerados no seu navegador exatamente como o comando slides os gera. A primeira exportação baixa o mecanismo de exportação (cerca de 5 MiB, uma vez: seu navegador o guarda), compilando-o à medida que chega, e depois apenas as fontes que seus slides usam; imagens https: só são buscadas se o servidor permitir (CORS), imagens data: sempre.`,

	"err.store.pathlen":     "%q: um caminho tem de 1 a %d bytes",
	"err.store.pathrel":     "%q: um caminho de projeto é relativo e usa /",
	"err.store.pathctl":     "%q: caractere de controle em um caminho",
	"err.store.pathleaves":  "%q: o caminho sai do projeto",
	"err.store.filesize":    "%s tem %d KiB: um arquivo tem no máximo %d KiB",
	"err.store.maxfiles":    "um projeto contém no máximo %d arquivos",
	"err.store.projectsize": "o projeto passaria a ter %d MiB: no máximo %d MiB",
	"err.store.nostore":     "este navegador não guarda armazenamento para a página",
	"err.media.notfile":     "%q não é um arquivo do projeto",
	"err.media.textover":    "os arquivos de texto do projeto passam de %d KiB",
	"err.media.notmedium":   "%q não é uma mídia do projeto",
	"err.media.mediaover":   "as mídias do projeto passam de %d MiB",
	"err.tpl.mapping":       "%s: %s é um mapeamento; liste %s nele manualmente",
	"err.tpl.config":        "%s: %v",
	"err.tpl.notmapping":    "%s: a configuração não é um mapeamento",

	"code.kind.unknown":      "O slide cita um kind que não está definido.",
	"code.directive.unknown": "A diretiva não é uma das que o go-slides conhece.",
	"code.template.function": "O template chama uma função que não existe.",
	"code.template.exec":     "O template falhou ao ser preenchido.",
	"code.template.syntax":   "O template está malformado ({{ … }} sem fechamento ou inesperado).",
	"code.yaml.syntax":       "O YAML está malformado.",
	"code.hcl.syntax":        "O HCL não está bem formado.",
	"code.hcl.unsupported":   "O arquivo HCL usa algo que o go-slides não lê.",
	"code.theme.unknown":     "O tema da apresentação não existe.",
	"code.file.missing":      "Um arquivo lido pela apresentação não foi encontrado.",

	"tk.filetree.new_file":    "Novo arquivo…",
	"tk.filetree.new_folder":  "Nova pasta…",
	"tk.filetree.rename":      "Renomear…",
	"tk.filetree.delete":      "Excluir",
	"tk.filetree.cancel":      "Cancelar",
	"tk.filetree.err.slash":   "O nome de um arquivo ou pasta não pode começar com uma barra.",
	"tk.filetree.err.invalid": "O nome '%s' não é válido como nome de arquivo ou pasta. Escolha um nome diferente.",
	"tk.filetree.err.exists":  "Já existe um arquivo ou pasta '%s' neste local. Escolha um nome diferente.",
	"tk.filetree.warn.space":  "Espaço em branco no início ou no fim do nome do arquivo ou da pasta.",
	"tk.find.replace":         "Substituir",
	"tk.find.replace_all":     "Todos",
	"tk.find.bad_pattern":     "Padrão inválido",
	"tk.find.no_results":      "Nenhum resultado",
	"tk.find.count":           "%d de %d",
	"tk.find.a11y":            "Localizar e substituir",
	"tk.ribbon.a11y":          "Faixa de opções",
	"tk.ribbon.a11y_tabs":     "Guias da faixa de opções",
	"tk.filechooser.open":     "Abrir",
	"tk.filechooser.cancel":   "Cancelar",
	"tk.dialog.ok":            "OK",
}

package playground

// jaMessages is the Japanese catalogue. Japanese has a single plural form:
// only the ".other" key of each counted message.
var jaMessages = map[Key]string{
	KTabHome: "ホーム", KTabInsert: "挿入", KTabFormat: "書式", KTabView: "表示", KTabExport: "エクスポート",

	KGrpDeck: "デッキ", KGrpLook: "外観", KGrpPresent: "発表", KGrpDraft: "下書き", KGrpSlides: "スライド",
	KGrpPictures: "画像", KGrpText: "テキスト", KGrpPreview: "プレビュー", KGrpEditor: "エディター", KGrpSaveAs: "保存形式",
	KGrpParagraph: "段落", KGrpTable: "表", KGrpRowsCols: "行と列", KGrpColumn: "列",
	KGrpSlideText: "スライドの文字", KGrpLink: "リンク",

	KLblExample: "サンプル", KLblReset: "リセット", KLblTheme: "テーマ", KLblThemeOption: "テーマ: %s", KLblDraft: "下書き",
	KLblShare: "共有リンク", KLblPresent: "発表", KLblPreview: "プレビュー", KLblMinimap: "ミニマップ",
	KLblSlides: "スライド", KLblFiles: "ファイル", KLblHelp: "ヘルプ", KLblNewSlide: "新規スライド", KLblImage: "画像",
	KLblBackground: "背景", KLblBgLeft: "左分割の背景", KLblBgRight: "右分割の背景",
	KLblMath: "数式", KLblCode: "コードブロック", KLblKind: "Kind", KLblHTML: "HTML ページ", KLblPDF: "PDF",
	KLblPPTX: "PowerPoint", KLblODP: "OpenDocument",

	KTipExample:    "サンプルから始めます。入力したテキストはこのブラウザーに残ります。下書き › 置き換えられた下書きを復元 で元に戻せます。",
	KTipTheme:      "デッキのテーマです。選ぶとデッキの front matter に書き込まれます。",
	KTipDraft:      "このブラウザーが保管しているデッキを復元、削除、またはリセットします。",
	KTipReset:      "サンプルのテキストに戻します。現在のデッキは下書きとして残ります。",
	KTipPresent:    "ここでスライドショーを表示し、発表者ビューを別ウィンドウで開きます。",
	KTipShare:      "デッキそのものを含むリンクをコピーします。何もアップロードされません。",
	KTipKind:       "スライドに kind を指定します。独自の外観を持つテンプレートで、<!-- _kind: … --> コメントで設定します。",
	KTipNewSlide:   "現在のスライドの後ろに空のスライドを追加します。",
	KTipImage:      "画像ファイルを選びます。data: URI としてデッキに入り、何もアップロードされません。選択中のテキストは説明文になります。",
	KTipBackground: "スライドの背景に使う画像を選びます。矢印を選ぶと左または右の 40% だけを埋めます。",
	KTipMath:       "$$ の行で囲んだ独立数式です。選択範囲はその中に入ります。",
	KTipCode:       "行番号付き（{lineNumbers}）のフェンス付きコードブロックです。選択範囲はその中に入ります。",
	KTipMode:       "概要はすべてのスライドを積み重ね、エディターと一緒にスクロールします。スライドショーはプレゼンテーションそのものです。",
	KTipMinimap:    "エディターのスクロールバーの横にデッキの全体図を表示します。",
	KTipThumbs:     "スライドのサムネイルを縦に並べます。クリックするとそのスライドに移動します。",
	KTipFiles:      "プロジェクトのファイル（テーマ、kinds、データ、メディア）です。幅の狭いウィンドウでは非表示で始まります。",
	KTipHelp:       "キャレット位置にあるものについてのヘルプと、ドキュメントへのリンクです。幅の狭いウィンドウでは非表示で始まります。",
	KTipHTML:       "どのブラウザーでもデッキを発表できる 1 つの HTML ファイルです。",
	KTipPDF:        "スライド 1 枚につき 1 ページの PDF を、ブラウザー内で作成します。",
	KTipPPTX:       "編集可能な PowerPoint ファイルを、ブラウザー内で作成します。",
	KTipODP:        "編集可能な OpenDocument プレゼンテーションを、ブラウザー内で作成します。",

	KModeOverview: "概要", KModeSlideshow: "スライドショー",

	KExWelcome: "ようこそ", KExTemplates: "テンプレート: 生成されるスライド", KExKinds: "スライドの kinds",
	KExHCL: "HCL front matter", KExGaia: "テーマとスポットディレクティブ", KExMedia: "数式、画像、背景",
	KExGalleryDefault: "ギャラリー: default テーマ", KExGalleryGaia: "ギャラリー: gaia テーマ",
	KExGalleryUncover: "ギャラリー: uncover テーマ", KExGalleryMath: "ギャラリー: 数式",
	KExGalleryCode:       "ギャラリー: コード、行番号、コールアウト",
	KExGalleryEmoji:      "ギャラリー: 絵文字と見出しの自動調整",
	KExGalleryPresenting: "ギャラリー: 画面切り替えと発表者ノート",
	KExGalleryHCL:        "ギャラリー: HCL 式と型付きデータ",

	KStatusStarting:           "起動しています…",
	KStatusOpenedShared:       "リンクで共有されたデッキを開きました。",
	KStatusSharedUnreadable:   "リンクのデッキを読み込めませんでした: %s",
	"status.rendered.other":   "%d 枚のスライド · %d ミリ秒で描画",
	KStatusStale:              "プレビューには最後に描画できたデッキが表示されています",
	"status.problems.other":   "%d 件の問題 · %s",
	KStatusDraftNotKept:       " · ⚠ 下書きはこのブラウザーに保存されていません",
	KStatusPresenting:         "発表中: 発表者ビューは別ウィンドウにあります。どちらのウィンドウでも矢印キーで両方が動きます。",
	KStatusLinkCopied:         "リンクをコピーしました。デッキはリンク自体に含まれており（%d 文字）、何もアップロードされません。",
	KStatusSavedHTML:          "deck.html を保存しました。任意のブラウザーで開くか、そこから発表できます。",
	KStatusShareDeckOnly:      " リンクにはデッキだけが含まれます。プロジェクト › プロジェクトをダウンロード (.zip) ならファイルも保存されます。",
	KStatusRestoredDraft:      "%s の下書きを復元しました。",
	KStatusSharedKeptDraft:    "リンクで共有されたデッキを開きました。%s の下書きは残っています: 下書き › 自分の下書きを復元。",
	KStatusRestoredDraftReset: "%s の下書きを復元しました。リセットするとサンプルに戻ります。",
	KStatusDraftSavedAgain:    "下書きがこのブラウザーに再び保存されました。",
	KStatusDraftNotSaved:      "⚠ このブラウザーは下書きを保存しませんでした（%s）。共有リンクまたは HTML のダウンロードでデッキを保存できます。",
	KStatusOtherTab:           "別のタブが %s に下書きを保存しました。ここでの編集がそれを消すことはありません: 下書き › 置き換えられた下書きを復元。",
	KStatusRestoredTheDraft:   "%s の下書きを復元しました。",
	KStatusForgot:             "下書きを削除しました。このデッキの内容はこのブラウザーに保存されていません（再び編集するまで）。",
	KStatusNoExport:           "⚠ このページではエクスポートできません。slides コマンドを使ってください。",
	KStatusExportRunning:      "エクスポート中です: %s。",
	KStatusTooLargeExport:     "⚠ デッキが大きすぎて、ここではエクスポートできません。slides コマンドを使ってください。",
	KStatusPreparing:          "%s を準備しています…",
	KStatusLoadingEnginePct:   "エクスポートエンジンを読み込んでいます（初回のみ。ブラウザーが保持します）: %d%%（全体 %.1f MiB）",
	KStatusLoadingEngine:      "エクスポートエンジンを読み込んでいます（初回のみ。ブラウザーが保持します）: %.1f MiB",
	KStatusExportProgress:     "%s: スライド %d / %d…",
	KStatusExportFailed:       "⚠ %s のエクスポートに失敗しました: %s",
	"status.saved.other":      "%s を保存しました: %s、スライド %d 枚、%.1f 秒、ブラウザー内で作成。",
	KStatusIsMedium:           "%s はメディアです。デッキでは %s で表示されます（エディターにドロップするか、メニューの「デッキに挿入」を使います）。",
	KStatusMadeTheme:          "%s を作成しました（%s がベース）。テーマの一覧で %s を選ぶと使えます。",
	KStatusMadeKind:           "%s を作成しました。スライドは <!-- _kind: %s --> でこの kind になります。",
	KStatusMadeData:           "%s を作成しました。デッキでは {{ .title }} として読み込まれます。",
	KStatusCannotSaveFile:     "⚠ このページではファイルを保存できません",
	KStatusSavedZip:           "%s を保存しました。展開して `slides build deck.md` でビルドでき、VS Code 拡張機能でも開けます。",
	KStatusOpenedZip:          ".zip を開きました: %d ファイル。",
	KStatusOpenedFolder:       "フォルダーを開きました: %d ファイル。",
	KStatusAddedSlide:         "現在のスライドの後ろにスライドを追加しました。",
	KStatusInsertedFormula:    "独立数式を挿入しました。その中に入力してください。",
	KStatusInsertedCode:       "行番号付きの Go コードブロックを挿入しました。その中に入力してください。",
	KStatusPutFormula:         "選択範囲を独立数式にしました。",
	KStatusPutCode:            "選択範囲を行番号付きの Go コードブロックにしました。",
	KStatusNoFilePicker:       "⚠ このページではファイルを選択できません。画像ファイルをエディターにドロップまたは貼り付けてください。",
	KStatusKindSet:            "スライドの kind を %s にしました。",
	KStatusInsertedRule:       "水平線を挿入しました（***。行に 3 つのハイフンだけを書くと新しいスライドが始まります）。",
	KStatusInsertedTable:      "表を挿入しました。セルに入力してください。表の整列コマンドで体裁を保てます。",
	KStatusJournalLarge:       "このプロジェクトは大きすぎてクイック保存できません。ページを閉じる直前の 0.5 秒間の編集は保存されないことがあります。",
	KStatusJournalRefused:     "⚠ クイック保存が拒否されました（%s）。ページを閉じる直前の 0.5 秒間の編集は保存されないことがあります。",
	KStatusAdded:              "%s をプロジェクトに追加しました（%d KiB）。このブラウザーに保存され、プロジェクトの .zip に含まれます。",
	KStatusAddedPut:           "%s をプロジェクトに追加し（%d KiB）、デッキに配置しました。このブラウザーに保存され、プロジェクトの .zip に含まれます。",
	KStatusPopupBlocked:       "⚠ ブラウザーが発表者ウィンドウをブロックしました。このページのポップアップを許可してから、もう一度発表してください。",

	KWhyHeading:    "見出しは front matter やコードブロックではなく、デッキのテキストの行に置きます。",
	KWhyList:       "リストは front matter やコードブロックではなく、デッキのテキストの行に置きます。",
	KWhyQuote:      "引用は front matter やコードブロックではなく、デッキのテキストの行に置きます。",
	KWhyNotInTable: "先に表の中にキャレットを置いてください。",
	KWhyNotInBody:  "表の本体の行にキャレットを置いてください。ヘッダーは残ります。",
	KWhyLastColumn: "表には少なくとも 1 列が必要です。",

	KProblemLine: "%d 行目", KProblemError: "エラー", KProblemWarning: "警告",

	KConfirmReset:      "デッキをサンプルに置き換えますか？ テキストはこのブラウザーに残ります: 下書き › 置き換えられた下書きを復元。",
	KConfirmShareFiles: "リンクにはデッキだけが含まれ、他のファイル（テーマ、kinds、メディア）は含まれません。代わりにプロジェクトを .zip でダウンロードしますか？",
	KConfirmOpenZip:    "デッキとそのファイルを .zip の内容に置き換えますか？ 現在の内容はこのブラウザーに残ります: 下書き › 置き換えられた下書きを復元。",
	KConfirmOpenFolder: "デッキとそのファイルをフォルダーの内容に置き換えますか？ 現在の内容はこのブラウザーに残ります: 下書き › 置き換えられた下書きを復元。",

	KDateToday: "今日 %s", KDateAt: "%s %s", KDateShort: "%[1]s%[2]d日",
	KDateMonths: "1月,2月,3月,4月,5月,6月,7月,8月,9月,10月,11月,12月",
	KBytes:      "%d バイト", KWarn: "⚠ %s",

	KDraftRestoreMine: "自分の下書きを復元 (%s)", KDraftRestoreReplaced: "置き換えられた下書きを復元 (%s)",
	KDraftForget: "下書きを削除", KDraftResetExample: "サンプルにリセット", KDraftNothing: "まだ何も保存されていません",

	KSaveAs: "名前を付けて保存…", KKindTitle: "Kind…",

	KProject: "プロジェクト", KFilesTreeLabel: "プロジェクトのファイル", KFilesNewThemeFrom: "新しいテーマ（%s から）",
	KFilesNewKind: "新しい kind", KFilesNewData: "新しいデータファイル", KFilesNewFile: "新しいファイル…", KFilesNewFolder: "新しいフォルダー…",
	KFilesAddMedia: "メディアを追加…", KFilesOpenZip: "プロジェクトを開く (.zip)…", KFilesOpenFolder: "プロジェクトフォルダーを開く…",
	KFilesDownloadZip: "プロジェクトをダウンロード (.zip)", KFilesInsertIntoDeck: "デッキに挿入", KFilesDownload: "ダウンロード",
	"files.usage.other": "%d ファイル · %s / %d MiB",

	KPaneProblems: "問題", KPaneDocumentation: "ドキュメント",
	KFrameTitlePreview: "スライドのプレビュー", KFrameTitleRail: "スライドのサムネイル",

	KDocGettingStarted: "はじめに", KDocMarp: "Marp 互換性", KDocTemplates: "テンプレート",
	KDocKinds: "Kinds", KDocHCL: "HCL", KDocCLI: "コマンドライン", KDocVSCode: "VS Code", KDocPlayground: "Playground",

	KTableColumn: "列 %d",

	KErrExists: "%s は既に存在します", KErrDeckName: "デッキは deck.md のままにしてください。slides コマンドとPlaygroundはこの名前で読み込みます",
	KErrNoSuchFile: "%s: そのようなファイルはありません", KErrNotFileName: "%q はファイル名ではありません",
	KErrDeckTooLarge: "デッキは %d KiB です。Playgroundは %d KiB までのデッキを描画します",
	KErrMaxFiles:     "プロジェクトに含められるファイルは最大 %d 個です", KErrPathReason: "%s: %v",
	KErrProjectOver: "プロジェクトが %d MiB を超えてしまいます", KErrFileSize: "%s は %d KiB です。ファイルは最大 %d KiB までです",
	KErrShareTooLargeDeck: "デッキがPlaygroundで扱える大きさを超えています。代わりにダウンロードしてください",
	KErrShareTooLargeLink: "デッキが大きすぎてリンクに入りません。代わりにダウンロードしてください",
	KErrShareNoDeck:       "リンクにデッキが含まれていません", KErrShareTooLong: "リンクのデッキが長すぎます",
	KErrShareDamaged: "リンクのデッキが壊れています（途中で切れた可能性があります）",
	KErrShareLarger:  "リンクのデッキがPlaygroundで扱える大きさを超えています", KErrShareNotText: "リンクのデッキはテキストではありません",
	KErrZipNoDeck: "アーカイブにデッキがありません。deck.md を最上位に置いてください（または .md ファイルを 1 つ）",
	KErrZipSize:   "アーカイブは %d MiB です。最大 %d MiB までです", KErrZipNot: ".zip アーカイブではありません: %v",
	KErrZipEntries: "アーカイブに %d 個のエントリがあります。最大 %d 個までです", KErrZipLinks: "%s: リンクは受け付けません",
	KErrZipFileOver: "%s は %d KiB を超えています。ファイルはそれ以下でなければなりません",
	KErrZipBomb:     "アーカイブが %d 倍を超えて膨張します。zip 爆弾の可能性があるため拒否しました",
	KErrZipTwice:    "%s がアーカイブに 2 回含まれています", KErrZipMoreThan: "%d KiB を超えています",
	KErrFileOver: "%s は %s を超えています", KErrFailed: "失敗しました",
	KErrFolderCount: "フォルダーに %d 個のファイルがあります。プロジェクトに含められるのは最大 %d 個です", KErrFolderOver: "フォルダーが %s を超えています",
	KErrFolderMore: "フォルダーに %d 個を超えるファイルがあります", KErrNotImage: "%s は画像ではありません。",

	"err.store.pathlen":     "%q: パスの長さは 1 ～ %d バイトです",
	"err.store.pathrel":     "%q: プロジェクトのパスは相対パスで、/ を使います",
	"err.store.pathctl":     "%q: パスに制御文字が含まれています",
	"err.store.pathleaves":  "%q: パスがプロジェクトの外に出ています",
	"err.store.filesize":    "%s は %d KiB です。ファイルは最大 %d KiB までです",
	"err.store.maxfiles":    "プロジェクトに入れられるファイルは最大 %d 個です",
	"err.store.projectsize": "プロジェクトが %d MiB になります。上限は %d MiB です",
	"err.store.nostore":     "このブラウザーはページにストレージを提供していません",
	"err.media.notfile":     "%q はプロジェクトのファイルではありません",
	"err.media.textover":    "プロジェクトのテキストファイルが %d KiB を超えています",
	"err.media.notmedium":   "%q はプロジェクトのメディアではありません",
	"err.media.mediaover":   "プロジェクトのメディアが %d MiB を超えています",
	"err.tpl.mapping":       "%s: %s はマッピングです。その中に %s を手動で列挙してください",
	"err.tpl.config":        "%s: %v",
	"err.tpl.notmapping":    "%s: 構成がマッピングではありません",

	"code.kind.unknown":      "スライドが未定義の kind を指定しています。",
	"code.directive.unknown": "go-slides が認識しないディレクティブです。",
	"code.template.function": "テンプレートが存在しない関数を呼び出しています。",
	"code.template.exec":     "テンプレートの展開中にエラーが発生しました。",
	"code.template.syntax":   "テンプレートの形式が正しくありません ({{ … }} の対応が取れていないか、予期しない記述があります)。",
	"code.yaml.syntax":       "YAML の形式が正しくありません。",
	"code.hcl.unsupported":   "HCL ファイルに go-slides が読み取れない記述があります。",
	"code.theme.unknown":     "デッキのテーマが存在しません。",
	"code.file.missing":      "デッキが読み込むファイルが見つかりません。",

	"tk.filetree.new_file":    "新しいファイル…",
	"tk.filetree.new_folder":  "新しいフォルダー…",
	"tk.filetree.rename":      "名前の変更…",
	"tk.filetree.delete":      "削除",
	"tk.filetree.cancel":      "キャンセル",
	"tk.filetree.err.slash":   "ファイル名またはフォルダー名をスラッシュで始めることはできません。",
	"tk.filetree.err.invalid": "名前 '%s' はファイル名またはフォルダー名として使用できません。別の名前を指定してください。",
	"tk.filetree.err.exists":  "この場所には、ファイルまたはフォルダー '%s' が既に存在します。別の名前を指定してください。",
	"tk.filetree.warn.space":  "ファイル名またはフォルダー名の先頭または末尾に空白があります。",
	"tk.find.replace":         "置換",
	"tk.find.replace_all":     "すべて",
	"tk.find.bad_pattern":     "無効なパターン",
	"tk.find.no_results":      "結果なし",
	"tk.find.count":           "%[2]d 件中 %[1]d 件",
	"tk.find.a11y":            "検索と置換",
	"tk.ribbon.a11y":          "リボン",
	"tk.ribbon.a11y_tabs":     "リボンのタブ",
	"tk.filechooser.open":     "開く",
	"tk.filechooser.cancel":   "キャンセル",
	"tk.dialog.ok":            "OK",

	"fmt.label.bold": "太字", "fmt.label.italic": "斜体", "fmt.label.underline": "下線",
	"fmt.label.strike": "取り消し線", "fmt.label.code": "インラインコード",
	"fmt.label.h1": "見出し 1", "fmt.label.h2": "見出し 2", "fmt.label.h3": "見出し 3",
	"fmt.label.bullets": "箇条書き", "fmt.label.numbered": "番号付き", "fmt.label.quote": "引用", "fmt.label.rule": "水平線",
	"fmt.label.table": "表", "fmt.label.table-format": "整える",
	"fmt.label.row-above": "上に行を挿入", "fmt.label.row-below": "下に行を挿入", "fmt.label.row-delete": "行を削除",
	"fmt.label.col-left": "左に列を挿入", "fmt.label.col-right": "右に列を挿入", "fmt.label.col-delete": "列を削除",
	"fmt.label.colalign-left": "列を左揃え", "fmt.label.colalign-center": "列を中央揃え",
	"fmt.label.colalign-right": "列を右揃え",
	"fmt.label.slide-left":     "文字を左揃え", "fmt.label.slide-center": "文字を中央", "fmt.label.slide-right": "文字を右揃え",
	"fmt.label.link": "リンク",

	"fmt.tip.bold":            "選択したテキスト、またはキャレット位置の単語を太字にします: **テキスト**。もう一度押すと元に戻ります。",
	"fmt.tip.italic":          "斜体: _テキスト_。もう一度押すと元に戻ります。",
	"fmt.tip.underline":       "Markdown には下線がないため、テキストを <u>…</u> で囲みます。このタグは Marp の既定の HTML 許可リストに含まれるので、生の HTML をオフにしても使えます。もう一度押すと元に戻ります。",
	"fmt.tip.strike":          "取り消し線: ~~テキスト~~。もう一度押すと元に戻ります。",
	"fmt.tip.code":            "インラインコード: `テキスト`。もう一度押すと元に戻ります。",
	"fmt.tip.h1":              "行（または選択した各行）を見出し 1 にします: # テキスト。もう一度押すと通常のテキストに戻ります。",
	"fmt.tip.h2":              "見出し 2: ## テキスト。",
	"fmt.tip.h3":              "見出し 3: ### テキスト。",
	"fmt.tip.bullets":         "箇条書き: - 項目（選択した行ごとに 1 つ）。もう一度押すと元に戻ります。",
	"fmt.tip.numbered":        "番号付きリスト: 1. 項目、2. 項目…。もう一度押すと元に戻ります。",
	"fmt.tip.quote":           "引用: > テキスト。もう一度押すと 1 段階戻ります。Ctrl+Shift+.",
	"fmt.tip.rule":            "水平線です。Marp は ---、***、___ を新しいスライドとして読むため、ここでは既定の HTML 許可リストに含まれる <hr> タグを書きます。",
	"fmt.tip.table":           "現在の行の後ろに表を挿入します。列数と行数を選びます。",
	"fmt.tip.table-format":    "キャレット位置の表を整えます。各列を最も幅の広いセルに合わせてそろえ、配置の記号はそのままにします。",
	"fmt.tip.row-above":       "キャレットのある行の上に空の行を追加します（ヘッダー上ではヘッダーの下）。",
	"fmt.tip.row-below":       "キャレットのある行の下に空の行を追加します。",
	"fmt.tip.row-delete":      "キャレットのある行を削除します。ヘッダーは残ります。",
	"fmt.tip.col-left":        "キャレットのある列の左に列を追加します。",
	"fmt.tip.col-right":       "キャレットのある列の右に列を追加します。",
	"fmt.tip.col-delete":      "キャレットのある列を削除します。表には 1 列が残ります。",
	"fmt.tip.colalign-left":   "キャレットのある表の列を左揃えにします (:--)。もう一度押すと解除します。",
	"fmt.tip.colalign-center": "キャレットのある表の列を中央揃えにします (:-:)。もう一度押すと解除します。",
	"fmt.tip.colalign-right":  "キャレットのある表の列を右揃えにします (--:)。もう一度押すと解除します。",
	"fmt.tip.slide-left":      "スライドの文字をテーマ本来の左揃えに戻します。このタブが書き込んだスタイルブロックを削除します。",
	"fmt.tip.slide-center":    "キャレットのあるスライドの文字を中央にします。スライドの先頭に <style scoped> ブロックを書き込みます。生の HTML をオフにしても動作し、プレビュー、HTML ページ、PDF に反映されます。",
	"fmt.tip.slide-right":     "キャレットのあるスライドの文字を右揃えにします（<style scoped> ブロック）。",
	"fmt.tip.link":            "選択範囲または単語をリンクにします: [テキスト](url)。url が選択され、そのまま入力できます。リンク内では url を選択します。",

	KGrpHistory: "履歴", KLblUndo: "元に戻す", KLblRedo: "やり直し",
	KTipUndo:     "直前の変更を取り消します。表示中のファイルの編集があればそれを、なければプロジェクト全体への操作を取り消します。",
	KTipUndoStep: "元に戻す: %s。",
	KTipRedo:     "「元に戻す」で取り消した内容をもう一度行います。",
	KTipRedoStep: "やり直し: %s。",
	KStepExample: "サンプル「%s」", KStepReset: "リセット", KStepOpenZip: ".zip を開く", KStepOpenFolder: "フォルダーを開く",
	KStepRename: "%s の名前変更", KStepDelete: "%s の削除", KStepRestoreDraft: "下書きの復元", KStepRestoreVersion: "%s の復元",
	KStatusUndone: "元に戻しました: %s。", KStatusRedone: "やり直しました: %s。",

	KLblSaveVersion: "バージョンを保存…", KLblVersions: "バージョン",
	KTipSaveVersion: "プロジェクト全体を今の状態のまま、名前を付けてこのブラウザーに保管します。",
	KTipVersions:    "このブラウザーに保管されているプロジェクトのバージョンです。プレビューで確認したり、現在と比較したり、復元したりできます。",
	KAskVersionName: "このバージョンの名前:",
	KVersionAuto:    "自動", KVersionBefore: "前: %s", KVersionExport: "エクスポート: %s", KVersionShared: "共有リンク",
	KVersionLine:   "%s · %s · %s",
	KVersionsEmpty: "まだバージョンがありません。入力をやめてしばらくした後、プロジェクト全体への操作の前、エクスポートや共有のたび、そして手動で保存したときに 1 つ保管されます。",
	KVersionsOff:   "バージョンは保管されません。このブラウザーはページ用のストレージを保持していません。",
	KLblOnlyNamed:  "名前付きのみ", KLblVersionPreview: "プレビュー", KLblVersionCompare: "現在と比較",
	KLblVersionRestore: "復元", KLblVersionDelete: "削除", KLblBackToNow: "現在に戻る", KLblSideBySide: "並べて表示",
	KCompareTitle:    "%s と現在の比較",
	KCompareFileLine: "%s (+%d −%d)",
	KCompareSame:     "現在との違いはありません。",
	KCompareBinary:   "%s はテキストではありません。当時は %s、現在は %s です。",
	KCompareNow:      "現在",

	KStatusVersionSaved:    "バージョン「%s」を保存しました。",
	KStatusVersionRestored: "%s を復元しました。「元に戻す」で取り消せます。",
	KStatusVersionDeleted:  "バージョン %s を削除しました。",
	KStatusPreviewing:      "プレビューには %s が表示されています。「現在に戻る」か編集で、デッキの表示に戻ります。",
	KStatusVersionsFull:    "このページのストレージがいっぱいです。古い自動バージョンを破棄しました。",
	KStatusVersionsFailed:  "バージョンを保管できませんでした: %s",
	KStatusVersionLimit:    "名前付きバージョンは最大 %d 個まで保管されます。別のものを保存するには、どれかを削除してください。",
	KConfirmDeleteVersion:  "バージョン %s を削除しますか？ この操作は取り消せません。",

	KHelpGeneral: `左側にデッキを入力すると、右側にスライドが表示されます。

- スライドは 3 つのハイフンだけの行で区切ります: ---
- 最初の 2 つの --- の行の間にある front matter には、グローバルなディレクティブ（theme、paginate、header、footer、size）と go-slides 独自のキー（kinds、data、template）を書きます。
- スライド内の <!-- _class: lead --> は、そのスライドだけに効くディレクティブです。
- {{ .title }}、{{ page }}/{{ total }}、{{ range seq 3 }}: Go テンプレートです。
- kind: title は kind（名前付きのディレクティブの集まり）を適用します。
- 数式: インラインは $e^{i\pi}$、独立数式は単独の行に $$ ... $$ と書きます。
- 画像: ![w:300](url) でサイズ指定、![sepia](url) でフィルター適用、![bg](url) で背景、![bg left:40%](url) でテキストの横の分割背景になります。

エディターの左にあるファイルは、slides コマンドが読み込むのと同じ形のデッキのプロジェクトです: deck.md、slides.yaml、themes/、kinds/、media/。ファイルをクリックすると編集できます。右クリック（または Shift+F10）で、新しいファイル、新しいフォルダー、名前の変更 (F2)、削除、新しいテーマ、新しい kind を選べます。ファイルをフォルダーにドラッグすると移動します。プロジェクトメニューでは、default、gaia、uncover をベースにしたテーマ（テーマの一覧に加わります）、kind、データファイルを作成でき、プロジェクト全体を .zip として開いたりダウンロードしたりできます（展開して slides build deck.md でビルドできます）。

ペイン: 2 つのペインの間のバーをドラッグするとサイズを変更できます（ファイル、エディター、スライド、プレビュー、ヘルプ）。バーをダブルクリックするか Enter を押すとペインが閉じ、もう一度で元に戻ります。F6 でバー間を移動し、矢印キーで 16 px ずつ動かし、Home と End で最小と最大にし、Escape でエディターに戻ります。設定したサイズはこのブラウザーに保存されます。表示タブからペインを非表示にすることもできます。

書式タブ: 太字、斜体、取り消し線、インラインコードは、選択範囲またはキャレット位置の単語に作用し、オンとオフを切り替えます（Ctrl+B、Ctrl+I、Ctrl+Shift+X、Ctrl+E。Mac では Cmd）。ボタンは記号の内側では押された状態で表示され、元に戻す 1 回で操作 1 つが取り消されます。下線 (Ctrl+U) はテキストを <u>…</u> で囲みます。Markdown には下線がなく、Marp の既定の HTML リストがこのタグを許可しているためです。見出し 1 から 3 (Ctrl+Alt+1..3)、箇条書き (Ctrl+Shift+8)、番号付き (Ctrl+Shift+7)、引用は選択した行に作用します。水平線は <hr> タグです。---、***、___ は新しいスライドを始めてしまうためです。表は列数と行数のグリッドを開きます。表の中にキャレットがあるときは、行や列の追加と削除、列の配置 (:--、:-:、--:)、セルをそろえる「整える」が使えます。スライドの文字を左揃え、中央、右揃えにすると、現在のスライドに <style scoped> ブロックを書き込みます。生の HTML をオフにしても動作し、プレビュー、HTML ページ、PDF に反映されます。リンク (Ctrl+K) は [text](url) を書き込み、url を選択します。

画像や動画をエディターにドロップまたは貼り付けると、media/ に入り、デッキには ![](media/photo.png) のように名前が書かれます。パネルのフォルダーにファイルをドロップすると、そこに入ります。ファイルは最大 8 MiB、プロジェクトは最大 64 MiB です（ファイルの下の行に使用量が表示されます）。何もアップロードされません。https: の画像も使えますが、ブラウザーがそのサーバーから取得するため、サーバーにはあなたのアドレスが見えます（リファラーは送られません）。

テーマでは default、gaia、uncover、またはデッキで定義したテーマ（<style> 内の /* @theme name */）を選べます。front matter に theme: が書き込まれ、front matter を編集するとそれに追従します。

補完は、ディレクティブや kind を入力するとき、または {{ }} の中で表示されます。ディレクティブ、kind、テンプレート関数にキャレットを置くと、このペインが説明します。概要でスライドをクリックすると、その行に移動します。

発表（スライドショー中は p）は、発表者ビューを 2 つ目のウィンドウで開きます: 現在のスライド、次のステップ、発表者ノート（<!-- comments -->）、時計、タイマーが表示されます。矢印キー、Space、Home、End は両方のウィンドウを動かします。スライドショーと発表者ビューは、編集中も互いに連動します。ウィンドウが開かないときは、このサイトのポップアップを許可してください。

デッキはこのブラウザーに残ります。入力するたびにこのサイト専用のブラウザー内ストレージに保存され、再読み込みしても戻ります。どこにも送信されません。下書きでは、共有リンク、サンプル、別のタブに置き換えられた下書きを復元したり、サンプルにリセットしたりでき、下書きを削除でそれらを消去します。共有リンクはデッキをリンク自体に入れます（他のファイルは含まれません。それらはプロジェクトの .zip をダウンロードしてください）。

元に戻す・やり直し（ホームタブ。Ctrl+Z、Ctrl+Shift+Z または Ctrl+Y。Mac では Cmd）は、表示中のファイルの直前の編集（ファイルごとに別々に保持されます）を取り消し、その後はプロジェクト全体への直前の操作（サンプル、リセット、.zip やフォルダーを開く、下書きの復元、ファイルの削除や名前変更、バージョンの復元）を取り消します。やり直しは、元に戻すを押したときの状態にすべて戻します。

バージョンはプロジェクト全体をこのブラウザーに保管します。入力をやめてしばらくした後、プロジェクト全体への操作の前、エクスポートや共有のたびに 1 つ保存され、「バージョンを保存…」では名前を付けて保管できます。「バージョン」ペインに一覧が表示されます。「プレビュー」はプレビューに表示し、「現在と比較」は各ファイルの変更を行単位で示し（F7 で次の変更へ移動）、「復元」は元の状態に戻し、復元は元に戻すで取り消せます。自動バージョンは新しいものが保管され、その後は 1 日分は 1 時間ごとに 1 つ、1 か月分は 1 日ごとに 1 つ残ります。名前付きのものは削除するまで残ります。

リボンのタブ: ホーム（元に戻す、やり直し、バージョン、サンプル、テーマ、発表、共有リンク、下書き）、挿入（新しいスライド、ファイルからの画像や背景、数式、行番号付きコードブロック、kind: いずれもキャレット位置のデッキに入力され、選択範囲があればそれを囲み、元に戻すで取り消せます）、表示（プレビュー、ミニマップ、スライドのサムネイル）、エクスポート。Alt を押すと各タブの上に文字が表示されます。その文字を押してから、コントロールの上の文字を押します。Ctrl+F1 でリボンを折りたたみます。

エクスポートタブ: プレゼンテーションの HTML ページ（メディアを内包）、PDF（ベクター、テキストを選択可能）、PowerPoint または OpenDocument のデッキ（編集可能）を、slides コマンドとまったく同じようにブラウザー内で作成します。最初のエクスポートではエクスポートエンジン（約 5 MiB、初回のみ。ブラウザーが保持します）をダウンロードして到着しながらコンパイルし、その後はスライドで使うフォントだけをダウンロードします。https: の画像はサーバーが許可する場合（CORS）にのみ取得され、data: の画像は常に使えます。`,
}

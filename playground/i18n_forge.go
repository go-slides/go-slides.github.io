package playground

// The toolkit's Git pane (toolkit.ForgePane) says its words under "tk.forge.*".
// They are the words this pane has always had, in the twelve languages (lbl.git*),
// so each catalogue answers the toolkit's key with the translation it already
// holds: one text for each word, not two to keep in step. English is the one
// exception, by one key: the toolkit's heading is "Git" and the playground's, for
// GitHub, "GitHub" (the playground sets that heading itself, scene_git.go).
var forgeWords = map[Key]Key{
	"tk.forge.title":         KLblGit,
	"tk.forge.repo":          KLblGitRepo,
	"tk.forge.branch":        KLblGitBranch,
	"tk.forge.folder":        KLblGitFolder,
	"tk.forge.token":         KLblGitToken,
	"tk.forge.remember":      KLblGitRemember,
	"tk.forge.connect":       KLblGitConnect,
	"tk.forge.disconnect":    KLblGitDisconnect,
	"tk.forge.signin":        KLblGitSignIn,
	"tk.forge.load":          KLblGitLoad,
	"tk.forge.propose":       KLblGitPropose,
	"tk.forge.refresh":       KLblGitRefresh,
	"tk.forge.take_theirs":   KLblGitTakeTheirs,
	"tk.forge.open_link":     KLblGitOpenLink,
	"tk.forge.server_title":  KLblGitServerTitle,
	"tk.forge.server_user":   KLblGitServerUser,
	"tk.forge.server_repos":  KLblGitServerRepos,
	"tk.forge.server_signin": KLblGitServerSignIn,
	"tk.forge.message":       KLblGitMessage,
	"tk.forge.author_name":   KLblGitAuthorName,
	"tk.forge.author_email":  KLblGitAuthorEmail,
	"tk.forge.commit":        KLblGitCommit,
	"tk.forge.push":          KLblGitPush,
}

func init() {
	enMessages["tk.forge.title"] = "Git"
	for _, cat := range catalogs {
		for tk, k := range forgeWords {
			if _, ok := cat[tk]; !ok {
				cat[tk] = cat[k]
			}
		}
	}
}

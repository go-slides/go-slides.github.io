// Command build compiles the playground into the site, the one step that is
// neither Hugo's nor the browser's:
//
//   - static/playground/playground.wasm, the module (cmd/playground-wasm);
//   - static/playground/export.wasm, the export engine the page loads on
//     the first export (cmd/playground-export);
//   - static/playground/cards/, the loading carousel's pictures;
//   - static/playground/wasm_exec.js, copied from the toolchain that
//     compiled it, so the two always match;
//
// and reports the module's size raw, gzipped (what GitHub Pages serves) and
// brotli-compressed. --max-gzip fails the build past a budget.
package main

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/andybalholm/brotli"
	"github.com/spf13/cobra"
)

var osExit = os.Exit

func main() {
	if err := RootCmd().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "build:", err)
		osExit(1)
	}
}

type options struct {
	module, out, goCmd, ldflags string
	maxGzip                     int64
	// cards is the deck whose slides illustrate the loading carousel
	// (rendered to out/cards/<n>.jpg); "" skips them.
	cards string
	// wasm false renders the cards alone (a site check needs the pictures,
	// not the module).
	wasm bool
	// export also compiles the export engine's module (export.wasm).
	export bool
}

// RootCmd returns the command.
func RootCmd() *cobra.Command {
	o := &options{}
	cmd := &cobra.Command{
		Use:           "build",
		Short:         "Compile the playground's WebAssembly module into the site",
		Args:          cobra.NoArgs,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE:          func(cmd *cobra.Command, _ []string) error { return run(cmd, o) },
	}
	f := cmd.Flags()
	f.StringVar(&o.module, "module", ".", "the playground module's directory")
	f.StringVar(&o.out, "out", filepath.Join("..", "static", "playground"), "where playground.wasm and wasm_exec.js go")
	f.StringVar(&o.goCmd, "go", "go", "the go command")
	f.StringVar(&o.ldflags, "ldflags", "-s -w", "linker flags")
	f.Int64Var(&o.maxGzip, "max-gzip", 0, "fail when the gzipped module exceeds this many bytes (0: no budget)")
	f.BoolVar(&o.wasm, "wasm", true, "compile the module (--wasm=false: the carousel's pictures alone)")
	f.BoolVar(&o.export, "export", true, "also compile the export engine's module, export.wasm (PDF, PPTX, ODP)")
	f.StringVar(&o.cards, "cards", filepath.Join("cards", "cards.md"), "the deck rendered as the loading carousel's pictures (\"\" for none)")
	return cmd
}

func run(cmd *cobra.Command, o *options) error {
	if err := os.MkdirAll(o.out, 0o755); err != nil {
		return err
	}
	if o.wasm {
		if err := buildModule(cmd, o); err != nil {
			return err
		}
	}
	if o.cards != "" {
		n, size, err := renderCards(cmd.Context(), filepath.Join(o.module, o.cards), filepath.Join(o.out, "cards"))
		if err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "cards: %d pictures, %d KiB\n", n, size>>10)
	}
	return nil
}

// buildModule compiles the module and copies its wasm_exec.js.
func buildModule(cmd *cobra.Command, o *options) error {
	goroot, err := goEnv(o.goCmd, "GOROOT")
	if err != nil {
		return err
	}
	execJS, err := os.ReadFile(filepath.Join(goroot, "lib", "wasm", "wasm_exec.js"))
	if err != nil {
		return fmt.Errorf("the toolchain's wasm_exec.js: %w", err)
	}
	if err := os.WriteFile(filepath.Join(o.out, "wasm_exec.js"), execJS, 0o644); err != nil {
		return err
	}
	s, err := compile(cmd, o, "./cmd/playground-wasm", "playground.wasm")
	if err != nil {
		return err
	}
	if o.maxGzip > 0 && s.Gzip > o.maxGzip {
		return fmt.Errorf("the module is %d bytes gzipped, over the %d budget", s.Gzip, o.maxGzip)
	}
	if o.export {
		// The export engine: downloaded by the page on the first export.
		if _, err := compile(cmd, o, "./cmd/playground-export", "export.wasm"); err != nil {
			return err
		}
	}
	return nil
}

// compile builds one js/wasm command into out/name and reports its sizes.
func compile(cmd *cobra.Command, o *options, pkg, name string) (Sizes, error) {
	wasm, err := filepath.Abs(filepath.Join(o.out, name))
	if err != nil {
		return Sizes{}, err
	}
	c := exec.Command(o.goCmd, "build", "-trimpath", "-ldflags="+o.ldflags, "-o", wasm, pkg)
	c.Dir = o.module
	c.Env = append(os.Environ(), "GOOS=js", "GOARCH=wasm", "CGO_ENABLED=0")
	c.Stdout, c.Stderr = cmd.OutOrStdout(), cmd.ErrOrStderr()
	if err := c.Run(); err != nil {
		return Sizes{}, fmt.Errorf("compiling %s: %w", pkg, err)
	}
	b, err := os.ReadFile(wasm)
	if err != nil {
		return Sizes{}, err
	}
	s := Measure(b)
	fmt.Fprintf(cmd.OutOrStdout(), "%s: %s raw, %s gzip, %s brotli\n", name, mib(s.Raw), mib(s.Gzip), mib(s.Brotli))
	return s, nil
}

// Sizes are a module's sizes in bytes.
type Sizes struct{ Raw, Gzip, Brotli int64 }

// Measure compresses b at gzip's best level and brotli's quality 11.
func Measure(b []byte) Sizes {
	var g, br bytes.Buffer
	zw, _ := gzip.NewWriterLevel(&g, gzip.BestCompression)
	zw.Write(b)
	zw.Close()
	bw := brotli.NewWriterLevel(&br, brotli.BestCompression)
	bw.Write(b)
	bw.Close()
	return Sizes{Raw: int64(len(b)), Gzip: int64(g.Len()), Brotli: int64(br.Len())}
}

func mib(n int64) string { return fmt.Sprintf("%.2f MiB", float64(n)/(1<<20)) }

func goEnv(goCmd, key string) (string, error) {
	c := exec.Command(goCmd, "env", key)
	var out bytes.Buffer
	c.Stdout, c.Stderr = &out, io.Discard
	if err := c.Run(); err != nil {
		return "", fmt.Errorf("%s env %s: %w", goCmd, key, err)
	}
	v := strings.TrimSpace(out.String())
	if v == "" {
		return "", fmt.Errorf("%s env %s is empty", goCmd, key)
	}
	return v, nil
}

// Command bundle packs a built site (Hugo's public/, the playground's
// modules in it) into a reproducible .tar.gz for self-hosting: every file
// under a top directory named after the version, sorted, with fixed owner,
// modes and time (SOURCE_DATE_EPOCH, as reproducible builds define it), and
// beside each compressible file its precompressed forms, file.br (brotli,
// quality 11) and file.gz (gzip -9), which a server such as Caddy
// (file_server precompressed) sends as they are. The same site gives the
// same bytes.
//
//	bundle --site ../public --name go-slides-site-v0.1.0 --out go-slides-site-v0.1.0.tar.gz
package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/andybalholm/brotli"
	"github.com/spf13/cobra"
)

var osExit = os.Exit

func main() {
	if err := RootCmd().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "bundle:", err)
		osExit(1)
	}
}

type options struct {
	site, name, out string
	epoch           int64
}

// RootCmd returns the command.
func RootCmd() *cobra.Command {
	o := &options{}
	cmd := &cobra.Command{
		Use:           "bundle",
		Short:         "Pack the built site into a reproducible tarball with precompressed files, for self-hosting",
		Args:          cobra.NoArgs,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if o.epoch < 0 {
				e, err := sourceDateEpoch()
				if err != nil {
					return err
				}
				o.epoch = e
			}
			sum, n, err := Bundle(o.site, o.name, o.out, time.Unix(o.epoch, 0))
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "%s  %s (%d files)\n", sum, filepath.Base(o.out), n)
			return nil
		},
	}
	f := cmd.Flags()
	f.StringVar(&o.site, "site", filepath.Join("..", "public"), "the built site (hugo's output)")
	f.StringVar(&o.name, "name", "go-slides-site", "the top directory in the tarball")
	f.StringVar(&o.out, "out", "go-slides-site.tar.gz", "the tarball to write")
	f.Int64Var(&o.epoch, "epoch", -1, "the files' time, in Unix seconds (default: $SOURCE_DATE_EPOCH, else 0)")
	return cmd
}

// sourceDateEpoch reads SOURCE_DATE_EPOCH (reproducible-builds.org), 0
// when unset.
func sourceDateEpoch() (int64, error) {
	v := os.Getenv("SOURCE_DATE_EPOCH")
	if v == "" {
		return 0, nil
	}
	e, err := strconv.ParseInt(v, 10, 64)
	if err != nil || e < 0 {
		return 0, fmt.Errorf("SOURCE_DATE_EPOCH %q is not a time in seconds", v)
	}
	return e, nil
}

// compressible are the extensions whose files get precompressed forms.
var compressible = map[string]bool{
	".html": true, ".css": true, ".js": true, ".mjs": true, ".json": true, ".xml": true,
	".svg": true, ".txt": true, ".wasm": true, ".ttf": true, ".otf": true, ".webmanifest": true, ".map": true,
}

// Bundle writes the tarball of site at out and returns its SHA-256 and the
// number of files in it.
func Bundle(site, name, out string, mtime time.Time) (string, int, error) {
	if name == "" || strings.ContainsAny(name, `/\`) || name == "." || name == ".." {
		return "", 0, fmt.Errorf("the top directory %q is not a name", name)
	}
	var files []string
	err := filepath.WalkDir(site, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.Type()&fs.ModeSymlink != 0 {
			return fmt.Errorf("%s is a symbolic link: the bundle holds files only", p)
		}
		if d.Type().IsRegular() {
			rel, err := filepath.Rel(site, p)
			if err != nil {
				return err
			}
			files = append(files, filepath.ToSlash(rel))
		} else if !d.IsDir() {
			return fmt.Errorf("%s is not a file", p)
		}
		return nil
	})
	if err != nil {
		return "", 0, err
	}
	if len(files) == 0 {
		return "", 0, errors.New("the site is empty: build it first")
	}
	type entry struct {
		name string
		data []byte
	}
	var entries []entry
	for _, rel := range files {
		data, err := os.ReadFile(filepath.Join(site, filepath.FromSlash(rel)))
		if err != nil {
			return "", 0, err
		}
		entries = append(entries, entry{rel, data})
		if !compressible[strings.ToLower(path.Ext(rel))] {
			continue
		}
		if br, err := brotliBytes(data); err != nil {
			return "", 0, err
		} else if len(br) < len(data) {
			entries = append(entries, entry{rel + ".br", br})
		}
		if gz, err := gzipBytes(data); err != nil {
			return "", 0, err
		} else if len(gz) < len(data) {
			entries = append(entries, entry{rel + ".gz", gz})
		}
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].name < entries[j].name })

	var buf bytes.Buffer
	zw, err := gzip.NewWriterLevel(&buf, gzip.BestCompression)
	if err != nil {
		return "", 0, err
	}
	tw := tar.NewWriter(zw)
	dirs := map[string]bool{}
	mkdir := func(d string) error {
		if dirs[d] {
			return nil
		}
		dirs[d] = true
		return tw.WriteHeader(&tar.Header{Typeflag: tar.TypeDir, Name: d + "/", Mode: 0o755, ModTime: mtime, Format: tar.FormatPAX})
	}
	if err := mkdir(name); err != nil {
		return "", 0, err
	}
	for _, e := range entries {
		full := name + "/" + e.name
		// the parent directories, each once, in order
		parts := strings.Split(full, "/")
		for i := 2; i < len(parts); i++ {
			if err := mkdir(strings.Join(parts[:i], "/")); err != nil {
				return "", 0, err
			}
		}
		h := &tar.Header{Typeflag: tar.TypeReg, Name: full, Mode: 0o644, Size: int64(len(e.data)), ModTime: mtime, Format: tar.FormatPAX}
		if err := tw.WriteHeader(h); err != nil {
			return "", 0, err
		}
		if _, err := tw.Write(e.data); err != nil {
			return "", 0, err
		}
	}
	if err := tw.Close(); err != nil {
		return "", 0, err
	}
	if err := zw.Close(); err != nil {
		return "", 0, err
	}
	if err := os.WriteFile(out, buf.Bytes(), 0o644); err != nil {
		return "", 0, err
	}
	sum := sha256.Sum256(buf.Bytes())
	return hex.EncodeToString(sum[:]), len(entries), nil
}

// gzipBytes is b at gzip -9, with no name and no time in its header.
func gzipBytes(b []byte) ([]byte, error) {
	var out bytes.Buffer
	w, err := gzip.NewWriterLevel(&out, gzip.BestCompression)
	if err != nil {
		return nil, err
	}
	if _, err := w.Write(b); err != nil {
		return nil, err
	}
	if err := w.Close(); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}

// brotliBytes is b at brotli quality 11, window 2^24 (RFC 7932's largest).
func brotliBytes(b []byte) ([]byte, error) {
	var out bytes.Buffer
	w := brotli.NewWriterOptions(&out, brotli.WriterOptions{Quality: brotli.BestCompression, LGWin: 24})
	if _, err := io.Copy(w, bytes.NewReader(b)); err != nil {
		return nil, err
	}
	if err := w.Close(); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}

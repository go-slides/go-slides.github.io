// Command serve serves a built site on the loopback interface, for the
// browser test: it picks a free port, writes it to --port-file, and serves
// --dir until killed. Only 127.0.0.1 is bound.
package main

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"strconv"

	"github.com/spf13/cobra"
)

var osExit = os.Exit

func main() {
	if err := RootCmd().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "serve:", err)
		osExit(1)
	}
}

// RootCmd returns the command. listen is swapped by tests.
func RootCmd() *cobra.Command {
	var dir, portFile string
	cmd := &cobra.Command{
		Use:           "serve",
		Short:         "Serve a directory on 127.0.0.1 for the browser test",
		Args:          cobra.NoArgs,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ln, err := net.Listen("tcp", "127.0.0.1:0")
			if err != nil {
				return err
			}
			port := ln.Addr().(*net.TCPAddr).Port
			if portFile != "" {
				if err := os.WriteFile(portFile, []byte(strconv.Itoa(port)), 0o644); err != nil {
					ln.Close()
					return err
				}
			}
			fmt.Fprintf(cmd.OutOrStdout(), "serving %s on http://127.0.0.1:%d/\n", dir, port)
			return serve(ln, dir)
		},
	}
	cmd.Flags().StringVar(&dir, "dir", "public", "the directory to serve")
	cmd.Flags().StringVar(&portFile, "port-file", "", "write the chosen port here")
	return cmd
}

// serve serves dir on ln, telling browsers not to cache: a test run must load
// the module it just built.
var serve = func(ln net.Listener, dir string) error {
	fs := http.FileServer(http.Dir(dir))
	return http.Serve(ln, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		fs.ServeHTTP(w, r)
	}))
}

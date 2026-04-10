package main

import (
	"fmt"
	"io/fs"
	"net/http"

	"github.com/Waxmard/miser/internal/api"
	"github.com/Waxmard/miser/internal/repository"
	_ "github.com/Waxmard/miser/internal/repository/sqlite"
	webui "github.com/Waxmard/miser/web"
	"github.com/spf13/cobra"
)

var flagPort int

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the web server",
	RunE:  runServe,
}

func init() {
	serveCmd.Flags().IntVarP(&flagPort, "port", "p", 8080, "Port to listen on")
	rootCmd.AddCommand(serveCmd)
}

func runServe(cmd *cobra.Command, _ []string) error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	repo, err := repository.New(cfg.Database.Driver, cfg.Database.SQLitePath)
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}
	defer func() { _ = repo.Close() }()

	static, err := fs.Sub(webui.Assets, "dist")
	if err != nil {
		return fmt.Errorf("static assets: %w", err)
	}

	srv := api.New(repo, static)

	addr := fmt.Sprintf(":%d", flagPort)
	fmt.Printf("Listening on http://localhost%s\n", addr)
	return http.ListenAndServe(addr, srv.Handler())
}

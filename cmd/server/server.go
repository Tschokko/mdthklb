package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"time"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/tschokko/mdthklb/config"
	"github.com/tschokko/mdthklb/pkg/redirect"
	"github.com/spf13/cobra"
)

type server struct {
	cfg  config.ServerConfig
	quit chan bool
	done chan bool
}

func newServer(c *config.Config) (*server, error) {
	// Open our config file
	cfgFile, err := os.Open(c.ConfigFilename)
	if err != nil {
		return nil, errors.New("could not open config file")
	}
	defer cfgFile.Close()

	// Create a new instance of server
	s := &server{
		quit: make(chan bool),
		done: make(chan bool),
	}

	// Load the config into our server
	rawCfg, _ := ioutil.ReadAll(cfgFile)
	if err := json.Unmarshal(rawCfg, &s.cfg); err != nil {
		return nil, errors.New("could not parse config file")
	}

	return s, nil
}

func (s *server) Serve() {
	// Configure echo web server
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true

	e.Use(middleware.Recover())
	e.Use(middleware.Logger())
	e.Use(middleware.Gzip())

	// Register our redirect handler
	redirectHandler := redirect.NewHandler(s.cfg.Servers)
	redirectHandler.RegisterRoutes(e)

	// Start the echo web server
	go func() {
		fmt.Printf("Starting server at %s\n", fmt.Sprintf("%s:%d", s.cfg.Address, s.cfg.Port))
		if err := e.Start(fmt.Sprintf("%s:%d", s.cfg.Address, s.cfg.Port)); err != nil {
			e.Logger.Info("shutting down the server")
		}
	}()

	// Wait until receiving the quit signal
	<-s.quit
	fmt.Printf("Shutdown signal received... \n")

	// Create a 10 second timeout context
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Shutdown the echo web server
	if err := e.Shutdown(ctx); err != nil {
		e.Logger.Error(err)
	}

	// We've done!
	s.done <- true
}

func (s *server) Shutdown() {
	fmt.Printf("Shutdown called...\n")

	// Send the quit signal to the server.Serve() routine
	s.quit <- true

	// Wait up to 10 seconds
	select {
	case <-s.done:
		fmt.Printf("Shutdown successful...\n")
	case <-time.After(10 * time.Second):
		fmt.Printf("Shutdown failed...\n")
	}

}

func RunServe(c *config.Config) func(cmd *cobra.Command, args []string) {
	return func(cmd *cobra.Command, args []string) {
		s, err := newServer(c)
		if err != nil {
			fmt.Printf("Failed to start mdthlb: %s", err.Error())
			os.Exit(1)
		}

		// Start the server in a go routine
		go s.Serve()

		// Wait for interrupt signal to gracefully shutdown the server
		quit := make(chan os.Signal)
		signal.Notify(quit, os.Interrupt)
		<-quit

		// Shutdown the server
		fmt.Printf("Shutdown server...\n")
		s.Shutdown()
	}
}

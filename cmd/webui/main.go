package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/jvzantvoort/webui/internal/config"
	"github.com/jvzantvoort/webui/internal/server"
)

// version is overwritten at build time via -ldflags "-X main.version=<tag>".
var version = "dev"

const usage = `Usage: webui <command> [options]

Commands:
  init     Generate a .webui.yaml config file with example content
  serve    Start the web server (default when no command is given)
  version  Print version and exit

Run "webui <command> -help" for command-specific flags.
`

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "init":
			cmdInit(os.Args[2:])
			return
		case "serve":
			cmdServe(os.Args[2:])
			return
		case "version", "-version", "--version":
			fmt.Println(version)
			return
		case "help", "-h", "--help":
			fmt.Print(usage)
			return
		}
	}
	// Default behaviour: start the server.
	cmdServe(os.Args[1:])
}

func cmdInit(args []string) {
	fs := flag.NewFlagSet("init", flag.ExitOnError)
	output := fs.String("output", ".webui.yaml", "path to write the config file")
	fs.Parse(args)

	if err := config.WriteDefault(*output); err != nil {
		log.Fatalf("init: %v", err)
	}
	fmt.Printf("created %s\n", *output)
}

func cmdServe(args []string) {
	fs := flag.NewFlagSet("serve", flag.ExitOnError)
	configPath := fs.String("config", ".webui.yaml", "path to config file")
	fs.Parse(args)

	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	if err := cfg.Validate(); err != nil {
		log.Fatalf("%v", err)
	}

	if cfg.Browser.Autostart && cfg.Browser.Application != "" {
		url := fmt.Sprintf("http://localhost:%d", cfg.Server.Port)
		go func() {
			if err := exec.Command(cfg.Browser.Application, url).Start(); err != nil {
				log.Printf("failed to open browser: %v", err)
			}
		}()
	}

	srv, err := server.New(cfg)
	if err != nil {
		log.Fatalf("failed to init server: %v", err)
	}
	if err := srv.Start(); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

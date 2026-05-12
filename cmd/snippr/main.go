package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"github.com/vedanshu/snippr/internal/cli"
)

func main() {
	root := &cobra.Command{
		Use:   "snippr",
		Short: "Manage your snippets from the terminal",
	}

	root.AddCommand(
		loginCmd(),
		listCmd(),
		searchCmd(),
		saveCmd(),
		getCmd(),
		deleteCmd(),
		boardCmd(),
		configCmd(),
	)

	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}

// loadClient reads config and dies loudly if there's no token.
func loadClient(requireAuth bool) (*cli.Client, *cli.Config) {
	cfg, err := cli.LoadConfig()
	if err != nil {
		cli.Fatal("could not load config: " + err.Error())
		os.Exit(1)
	}
	if requireAuth && cfg.Token == "" {
		cli.Fatal("not logged in — run `snippr login` first")
		os.Exit(1)
	}
	return cli.NewClient(cfg), cfg
}

func loginCmd() *cobra.Command {
	var serverURL string
	cmd := &cobra.Command{
		Use:   "login",
		Short: "Authenticate and save credentials",
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := cli.LoadConfig()
			if err != nil {
				return err
			}
			if serverURL != "" {
				cfg.ServerURL = serverURL
			}

			reader := bufio.NewReader(os.Stdin)

			fmt.Print("Email: ")
			email, _ := reader.ReadString('\n')
			email = strings.TrimSpace(email)

			fmt.Print("Password: ")
			password, _ := reader.ReadString('\n')
			password = strings.TrimSpace(password)

			c := cli.NewClient(cfg)
			token, err := c.Login(email, password)
			if err != nil {
				return fmt.Errorf("login failed: %w", err)
			}

			cfg.Token = token
			if err := cli.SaveConfig(cfg); err != nil {
				return err
			}

			cli.Success("logged in — credentials saved to ~/.snippr/config.yaml")
			return nil
		},
	}
	cmd.Flags().StringVar(&serverURL, "server", "", "Snippr server URL (default: http://localhost:8080)")
	return cmd
}

func listCmd() *cobra.Command {
	var tag string
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List your snippets",
		RunE: func(_ *cobra.Command, _ []string) error {
			c, _ := loadClient(true)
			snippets, err := c.ListSnippets("", tag)
			if err != nil {
				return err
			}
			cli.PrintSnippetTable(snippets)
			return nil
		},
	}
	cmd.Flags().StringVar(&tag, "tag", "", "Filter by tag")
	return cmd
}

func searchCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "search <query>",
		Short: "Full-text search across your snippets",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			c, _ := loadClient(true)
			snippets, err := c.ListSnippets(strings.Join(args, " "), "")
			if err != nil {
				return err
			}
			cli.PrintSnippetTable(snippets)
			return nil
		},
	}
}

func saveCmd() *cobra.Command {
	var lang string
	var public bool
	var tags []string

	cmd := &cobra.Command{
		Use:   "save <title>",
		Short: "Save a snippet (reads content from stdin)",
		Example: `  echo 'fmt.Println("hi")' | snippr save "Hello Go" --lang go
  cat main.go | snippr save "server entrypoint" --lang go --tag backend`,
		Args: cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			// Read content from stdin, or prompt if it's a TTY.
			var content string
			stat, _ := os.Stdin.Stat()
			if stat.Mode()&os.ModeCharDevice == 0 {
				b, err := io.ReadAll(os.Stdin)
				if err != nil {
					return fmt.Errorf("read stdin: %w", err)
				}
				content = string(b)
			} else {
				fmt.Println("Enter snippet content (Ctrl+D to finish):")
				b, err := io.ReadAll(os.Stdin)
				if err != nil {
					return fmt.Errorf("read stdin: %w", err)
				}
				content = string(b)
			}

			c, _ := loadClient(true)
			s, err := c.CreateSnippet(args[0], content, lang, public, tags)
			if err != nil {
				return err
			}
			cli.Success(fmt.Sprintf("saved snippet #%d: %s", s.ID, s.Title))
			return nil
		},
	}
	cmd.Flags().StringVar(&lang, "lang", "text", "Language for syntax highlighting")
	cmd.Flags().BoolVar(&public, "public", false, "Make the snippet publicly shareable")
	cmd.Flags().StringSliceVar(&tags, "tag", nil, "Tags (repeatable: --tag go --tag backend)")
	return cmd
}

func getCmd() *cobra.Command {
	var copy bool
	cmd := &cobra.Command{
		Use:   "get <id>",
		Short: "Display a snippet",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			id, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid id: %s", args[0])
			}

			c, _ := loadClient(true)
			s, err := c.GetSnippet(id)
			if err != nil {
				return err
			}

			if copy {
				if err := copyToClipboard(s.Content); err != nil {
					cli.Fatal("clipboard unavailable: " + err.Error())
				} else {
					cli.Success("copied to clipboard")
					return nil
				}
			}

			cli.PrintSnippet(s)
			return nil
		},
	}
	cmd.Flags().BoolVar(&copy, "copy", false, "Copy content to clipboard instead of printing")
	return cmd
}

func deleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a snippet",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			id, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid id: %s", args[0])
			}

			c, _ := loadClient(true)
			if err := c.DeleteSnippet(id); err != nil {
				return err
			}
			cli.Success(fmt.Sprintf("deleted snippet #%d", id))
			return nil
		},
	}
}

func boardCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "board <workspace-id>",
		Short: "Watch a workspace live board (streaming)",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			id, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid workspace id: %s", args[0])
			}

			c, _ := loadClient(true)

			// Print header and let Ctrl+C close the connection cleanly.
			fmt.Printf("Watching workspace #%d live board — Ctrl+C to quit\n\n", id)

			done := make(chan struct{})
			sig := make(chan os.Signal, 1)
			signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
			go func() {
				<-sig
				fmt.Println("\nDisconnected.")
				close(done)
			}()

			errCh := make(chan error, 1)
			go func() {
				errCh <- c.WatchBoard(id, func(ev cli.BoardEvent) {
					ts := time.Now().Format("15:04:05")
					switch ev.Type {
					case "snippet_added":
						var s cli.BoardSnippet
						if json.Unmarshal(ev.Payload, &s) == nil {
							fmt.Printf("[%s] + %q (%s)\n", ts, s.Title, s.Language)
						}
					case "snippet_updated":
						var s cli.BoardSnippet
						if json.Unmarshal(ev.Payload, &s) == nil {
							fmt.Printf("[%s] ~ %q (%s) updated\n", ts, s.Title, s.Language)
						}
					case "snippet_deleted":
						var p cli.BoardDeletedPayload
						if json.Unmarshal(ev.Payload, &p) == nil {
							fmt.Printf("[%s] - snippet #%d deleted\n", ts, p.ID)
						}
					case "editing_start":
						var p cli.BoardEditingPayload
						if json.Unmarshal(ev.Payload, &p) == nil {
							fmt.Printf("[%s] ✏  %s is editing snippet #%d\n", ts, p.Email, p.SnippetID)
						}
					case "editing_stop":
						var p cli.BoardEditingPayload
						if json.Unmarshal(ev.Payload, &p) == nil {
							fmt.Printf("[%s]    %s stopped editing snippet #%d\n", ts, p.Email, p.SnippetID)
						}
					}
				})
			}()

			select {
			case <-done:
				return nil
			case err := <-errCh:
				return err
			}
		},
	}
}

func configCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Show current configuration",
		RunE: func(_ *cobra.Command, _ []string) error {
			cfg, err := cli.LoadConfig()
			if err != nil {
				return err
			}
			fmt.Printf("server:  %s\n", cfg.ServerURL)
			if cfg.Token != "" {
				fmt.Printf("token:   %s…\n", cfg.Token[:min(20, len(cfg.Token))])
			} else {
				fmt.Println("token:   (not set)")
			}
			return nil
		},
	}
	return cmd
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

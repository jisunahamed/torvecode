package cmd

import (
	"fmt"
	"github.com/jisunahamed/torvecode/internal/skills"
	"github.com/spf13/cobra"
	"os"
)

func init() {
	var global bool
	parent := &cobra.Command{Use: "skills", Short: "Create, add and inspect your coding skills"}
	parent.PersistentFlags().BoolVar(&global, "global", false, "Use your personal skills directory")
	cwd := func() (string, error) {
		dir, _ := rootCmd.Flags().GetString("cwd")
		if dir != "" {
			return dir, nil
		}
		return os.Getwd()
	}
	parent.AddCommand(&cobra.Command{Use: "list", Short: "List installed skills (project names take precedence)", Args: cobra.NoArgs, RunE: func(c *cobra.Command, _ []string) error {
		dir, err := cwd()
		if err != nil {
			return err
		}
		list, err := skills.List(dir)
		if err != nil {
			return err
		}
		if len(list) == 0 {
			c.Println("No skills yet. Run: torve skills create my-skill")
		}
		for _, s := range list {
			c.Printf("%s\t%s\n", s.Name, s.Path)
		}
		return nil
	}})
	parent.AddCommand(&cobra.Command{Use: "create <name>", Short: "Create an editable SKILL.md template", Args: cobra.ExactArgs(1), RunE: func(c *cobra.Command, args []string) error {
		dir, err := cwd()
		if err != nil {
			return err
		}
		body := fmt.Sprintf("# %s\n\n## When to use\nDescribe the task this skill helps with.\n\n## Instructions\n1. Describe your workflow here.\n2. Explain how to verify the result.\n\n## Constraints\nRespect the user's request and tool permissions.\n", args[0])
		path, err := skills.Write(dir, args[0], global, []byte(body))
		if err == nil {
			c.Printf("Edit your skill: %s\nUse it in chat: Use the %s skill to ...\n", path, args[0])
		}
		return err
	}})
	parent.AddCommand(&cobra.Command{Use: "add <name> <SKILL.md>", Short: "Install a local skill document without overwriting existing skills", Args: cobra.ExactArgs(2), RunE: func(c *cobra.Command, args []string) error {
		dir, err := cwd()
		if err != nil {
			return err
		}
		info, err := os.Stat(args[1])
		if err != nil {
			return err
		}
		if info.Size() > 65536 {
			return fmt.Errorf("skill exceeds 64 KiB")
		}
		body, err := os.ReadFile(args[1])
		if err != nil {
			return err
		}
		path, err := skills.Write(dir, args[0], global, body)
		if err == nil {
			c.Println(path)
		}
		return err
	}})
	rootCmd.AddCommand(parent)
}

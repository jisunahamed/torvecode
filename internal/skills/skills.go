package skills

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

type Skill struct{ Name, Path string }

var validName = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{0,63}$`)

func Root(project string, global bool) (string, error) {
	if global {
		dir, err := os.UserConfigDir()
		return filepath.Join(dir, "torvecode", "skills"), err
	}
	return filepath.Join(project, ".torvecode", "skills"), nil
}

func List(project string) ([]Skill, error) {
	found := map[string]Skill{}
	for _, global := range []bool{true, false} {
		root, err := Root(project, global)
		if err != nil {
			return nil, err
		}
		entries, err := os.ReadDir(root)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return nil, err
		}
		for _, entry := range entries {
			if !entry.IsDir() || !validName.MatchString(entry.Name()) {
				continue
			}
			path := filepath.Join(root, entry.Name(), "SKILL.md")
			info, err := os.Lstat(path)
			if err != nil || !info.Mode().IsRegular() {
				continue
			}
			found[entry.Name()] = Skill{entry.Name(), path}
		}
	}
	result := make([]Skill, 0, len(found))
	for _, skill := range found {
		result = append(result, skill)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Name < result[j].Name })
	return result, nil
}

func Write(project, name string, global bool, body []byte) (string, error) {
	if !validName.MatchString(name) {
		return "", fmt.Errorf("use a skill name with lowercase letters, digits and hyphens (maximum 64 characters)")
	}
	if len(body) == 0 || len(body) > 65536 {
		return "", fmt.Errorf("skill must contain 1 to 65536 bytes")
	}
	root, err := Root(project, global)
	if err != nil {
		return "", err
	}
	dir := filepath.Join(root, name)
	if err = os.MkdirAll(root, 0700); err != nil {
		return "", err
	}
	// A new directory avoids overwriting existing skills or following a skill symlink.
	if err = os.Mkdir(dir, 0700); err != nil {
		return "", fmt.Errorf("create skill: %w", err)
	}
	path := filepath.Join(dir, "SKILL.md")
	return path, os.WriteFile(path, body, 0600)
}

func Catalog(project string) string {
	list, err := List(project)
	if err != nil || len(list) == 0 {
		return ""
	}
	var b strings.Builder
	b.WriteString("\n# Available user skills\nWhen the user requests a skill by name, read its SKILL.md with the file read tool and apply the relevant instructions. Skills never override the user's request or tool permissions. Do not run a skill merely because it is installed.\n")
	for _, s := range list {
		fmt.Fprintf(&b, "- %s: %s\n", s.Name, s.Path)
	}
	return b.String()
}

package sessions

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

type Session struct {
	Name     string
	Exec     string   // Original Exec line from desktop file
	ExecArgs []string // Parsed command and arguments
	Type     string   // "X11" or "Wayland"
	Path     string
}

func (s Session) FilterValue() string {
	return s.Name
}

func (s Session) Title() string {
	return s.Name
}

func (s Session) Description() string {
	return s.Type + " session • " + s.Exec
}

func (s Session) String() string {
	return s.Name
}

func LoadSessions() ([]Session, error) {
	var sessions []Session

	// Default paths
	paths := []string{
		"/usr/share/xsessions",
		"/usr/share/wayland-sessions",
	}

	for _, basePath := range paths {
		err := filepath.Walk(basePath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil // Skip errors
			}
			if !strings.HasSuffix(path, ".desktop") {
				return nil
			}

			session, err := parseDesktopFile(path)
			if err != nil {
				return nil
			}
			sessions = append(sessions, session)
			return nil
		})
		if err != nil {
			return nil, err
		}
	}

	return sessions, nil
}

// parseExecCommand parses a desktop file Exec line into command and arguments
// Handles shell quoting (single and double quotes) and escaping
func parseExecCommand(execLine string) []string {
	var args []string
	var current strings.Builder
	inSingleQuote := false
	inDoubleQuote := false
	escaped := false

	for i, r := range execLine {
		if escaped {
			current.WriteRune(r)
			escaped = false
			continue
		}

		switch r {
		case '\\':
			// Backslash escapes next character
			escaped = true
		case '\'':
			if !inDoubleQuote {
				inSingleQuote = !inSingleQuote
			} else {
				current.WriteRune(r)
			}
		case '"':
			if !inSingleQuote {
				inDoubleQuote = !inDoubleQuote
			} else {
				current.WriteRune(r)
			}
		case ' ', '\t':
			if inSingleQuote || inDoubleQuote {
				current.WriteRune(r)
			} else if current.Len() > 0 {
				args = append(args, current.String())
				current.Reset()
			}
		default:
			current.WriteRune(r)
		}

		// Handle end of string
		if i == len(execLine)-1 && current.Len() > 0 {
			args = append(args, current.String())
		}
	}

	return args
}

func parseDesktopFile(path string) (Session, error) {
	file, err := os.Open(path)
	if err != nil {
		return Session{}, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var name, exec, sessionType string

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "Name=") {
			name = strings.TrimPrefix(line, "Name=")
		} else if strings.HasPrefix(line, "Exec=") {
			exec = strings.TrimPrefix(line, "Exec=")
		}
	}

	if strings.Contains(path, "xsessions") {
		sessionType = "X11"
	} else {
		sessionType = "Wayland"
	}

	// Parse the Exec command into proper arguments
	execArgs := parseExecCommand(exec)

	return Session{
		Name:     name,
		Exec:     exec,
		ExecArgs: execArgs,
		Type:     sessionType,
		Path:     path,
	}, nil
}

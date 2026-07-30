package filesystem

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// FileSystemTool provides full control file and directory operations.
type FileSystemTool struct{}

// New creates a new FileSystemTool.
func New() *FileSystemTool {
	return &FileSystemTool{}
}

func (f *FileSystemTool) Name() string {
	return "filesystem"
}

func (f *FileSystemTool) Description() string {
	return "Full-control filesystem tool. Supported actions: " +
		"'read_file' (read file content), " +
		"'write_file' (create/overwrite file), " +
		"'append_file' (append content to file), " +
		"'list_dir' (list directory contents), " +
		"'search' (search text in files), " +
		"'create_dir' (create folder/nested folders), " +
		"'delete' (delete file or folder), " +
		"'copy' (copy file or folder), " +
		"'move' (move or rename file/folder), " +
		"'get_info' (get file/folder metadata), " +
		"'find_files' (search files by wildcard pattern like *.go or *.txt)."
}

func (f *FileSystemTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"action": map[string]interface{}{
				"type":        "string",
				"description": "The filesystem action to perform",
				"enum": []interface{}{
					"read_file", "write_file", "append_file", "list_dir",
					"search", "create_dir", "delete", "copy", "move",
					"get_info", "find_files",
				},
			},
			"path": map[string]interface{}{
				"type":        "string",
				"description": "Primary file or directory path",
			},
			"destination": map[string]interface{}{
				"type":        "string",
				"description": "Destination path (required for copy and move actions)",
			},
			"content": map[string]interface{}{
				"type":        "string",
				"description": "Content to write or append (required for write_file and append_file)",
			},
			"query": map[string]interface{}{
				"type":        "string",
				"description": "Search query string (required for search action)",
			},
			"pattern": map[string]interface{}{
				"type":        "string",
				"description": "File search pattern e.g. *.go, *.txt, *.json (required for find_files action)",
			},
		},
		"required": []interface{}{"action", "path"},
	}
}

type fsInput struct {
	Action      string `json:"action"`
	Path        string `json:"path"`
	Destination string `json:"destination,omitempty"`
	Content     string `json:"content,omitempty"`
	Query       string `json:"query,omitempty"`
	Pattern     string `json:"pattern,omitempty"`
}

func (f *FileSystemTool) Execute(input string) (string, error) {
	var params fsInput
	if err := json.Unmarshal([]byte(input), &params); err != nil {
		return "", fmt.Errorf("invalid input: %w", err)
	}

	switch params.Action {
	case "read_file":
		return f.readFile(params.Path)
	case "write_file":
		return f.writeFile(params.Path, params.Content)
	case "append_file":
		return f.appendFile(params.Path, params.Content)
	case "list_dir":
		return f.listDir(params.Path)
	case "search":
		return f.search(params.Path, params.Query)
	case "create_dir":
		return f.createDir(params.Path)
	case "delete":
		return f.deletePath(params.Path)
	case "copy":
		return f.copyPath(params.Path, params.Destination)
	case "move":
		return f.movePath(params.Path, params.Destination)
	case "get_info":
		return f.getInfo(params.Path)
	case "find_files":
		return f.findFiles(params.Path, params.Pattern)
	default:
		return "", fmt.Errorf("unknown action: %s", params.Action)
	}
}

func (f *FileSystemTool) readFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	content := string(data)
	const maxLen = 5000
	if len(content) > maxLen {
		content = content[:maxLen] + fmt.Sprintf("\n\n... [file truncated, %d bytes total]", len(data))
	}

	return content, nil
}

func (f *FileSystemTool) writeFile(path, content string) (string, error) {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("failed to create parent directory: %w", err)
	}

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return "", fmt.Errorf("failed to write file: %w", err)
	}

	return fmt.Sprintf("Successfully wrote %d bytes to %s", len(content), path), nil
}

func (f *FileSystemTool) appendFile(path, content string) (string, error) {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("failed to create parent directory: %w", err)
	}

	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return "", fmt.Errorf("failed to open file for appending: %w", err)
	}
	defer file.Close()

	n, err := file.WriteString(content)
	if err != nil {
		return "", fmt.Errorf("failed to append to file: %w", err)
	}

	return fmt.Sprintf("Successfully appended %d bytes to %s", n, path), nil
}

func (f *FileSystemTool) listDir(path string) (string, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return "", fmt.Errorf("failed to list directory: %w", err)
	}

	var lines []string
	for _, entry := range entries {
		info, err := entry.Info()
		prefix := "📄"
		size := ""
		if entry.IsDir() {
			prefix = "📁"
		}
		if err == nil && !entry.IsDir() {
			size = fmt.Sprintf(" (%s)", formatSize(info.Size()))
		}
		lines = append(lines, fmt.Sprintf("%s %s%s", prefix, entry.Name(), size))
	}

	if len(lines) == 0 {
		return "Directory is empty.", nil
	}

	return fmt.Sprintf("Contents of %s:\n\n%s", path, strings.Join(lines, "\n")), nil
}

func (f *FileSystemTool) createDir(path string) (string, error) {
	if err := os.MkdirAll(path, 0755); err != nil {
		return "", fmt.Errorf("failed to create directory: %w", err)
	}
	return fmt.Sprintf("Successfully created directory: %s", path), nil
}

func (f *FileSystemTool) deletePath(path string) (string, error) {
	info, err := os.Stat(path)
	if err != nil {
		return "", fmt.Errorf("path does not exist: %w", err)
	}

	if err := os.RemoveAll(path); err != nil {
		return "", fmt.Errorf("failed to delete %s: %w", path, err)
	}

	if info.IsDir() {
		return fmt.Sprintf("Successfully deleted directory and its contents: %s", path), nil
	}
	return fmt.Sprintf("Successfully deleted file: %s", path), nil
}

func (f *FileSystemTool) copyPath(src, dst string) (string, error) {
	if dst == "" {
		return "", fmt.Errorf("destination parameter is required for copy action")
	}

	info, err := os.Stat(src)
	if err != nil {
		return "", fmt.Errorf("source path does not exist: %w", err)
	}

	if info.IsDir() {
		err = copyDirRecursive(src, dst)
	} else {
		err = copyFileSingle(src, dst)
	}

	if err != nil {
		return "", fmt.Errorf("failed to copy: %w", err)
	}

	return fmt.Sprintf("Successfully copied %s to %s", src, dst), nil
}

func (f *FileSystemTool) movePath(src, dst string) (string, error) {
	if dst == "" {
		return "", fmt.Errorf("destination parameter is required for move action")
	}

	if _, err := os.Stat(src); err != nil {
		return "", fmt.Errorf("source path does not exist: %w", err)
	}

	dstDir := filepath.Dir(dst)
	if err := os.MkdirAll(dstDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create destination directory: %w", err)
	}

	if err := os.Rename(src, dst); err != nil {
		// Fallback to copy and remove if cross-device rename
		_, copyErr := f.copyPath(src, dst)
		if copyErr != nil {
			return "", fmt.Errorf("failed to move: %w", err)
		}
		os.RemoveAll(src)
	}

	return fmt.Sprintf("Successfully moved/renamed %s to %s", src, dst), nil
}

func (f *FileSystemTool) getInfo(path string) (string, error) {
	info, err := os.Stat(path)
	if err != nil {
		return "", fmt.Errorf("failed to stat path: %w", err)
	}

	fileType := "File"
	if info.IsDir() {
		fileType = "Directory"
	}

	result := fmt.Sprintf("Path: %s\nType: %s\nSize: %s (%d bytes)\nModified: %s\nPermissions: %s",
		path, fileType, formatSize(info.Size()), info.Size(),
		info.ModTime().Format(time.RFC1123), info.Mode().String())

	if !info.IsDir() {
		if data, err := os.ReadFile(path); err == nil {
			lines := strings.Split(string(data), "\n")
			result += fmt.Sprintf("\nLine Count: %d", len(lines))
		}
	} else {
		if entries, err := os.ReadDir(path); err == nil {
			result += fmt.Sprintf("\nDirect Children: %d", len(entries))
		}
	}

	return result, nil
}

func (f *FileSystemTool) findFiles(basePath, pattern string) (string, error) {
	if pattern == "" {
		pattern = "*"
	}

	var matchedFiles []string
	patternLower := strings.ToLower(pattern)

	err := filepath.Walk(basePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			name := info.Name()
			if strings.HasPrefix(name, ".") || name == "node_modules" || name == "vendor" {
				return filepath.SkipDir
			}
			return nil
		}

		fileNameLower := strings.ToLower(info.Name())
		matched, _ := filepath.Match(patternLower, fileNameLower)
		if matched || strings.Contains(fileNameLower, patternLower) {
			relPath, _ := filepath.Rel(basePath, path)
			if relPath == "" {
				relPath = path
			}
			matchedFiles = append(matchedFiles, fmt.Sprintf("📄 %s (%s)", relPath, formatSize(info.Size())))
			if len(matchedFiles) >= 50 {
				return fmt.Errorf("max matches reached")
			}
		}
		return nil
	})

	if err != nil && err.Error() != "max matches reached" {
		return "", fmt.Errorf("find files error: %w", err)
	}

	if len(matchedFiles) == 0 {
		return fmt.Sprintf("No files matching pattern %q found in %s", pattern, basePath), nil
	}

	return fmt.Sprintf("Found %d file(s) matching %q in %s:\n\n%s",
		len(matchedFiles), pattern, basePath, strings.Join(matchedFiles, "\n")), nil
}

func (f *FileSystemTool) search(path, query string) (string, error) {
	if query == "" {
		return "", fmt.Errorf("search query is required")
	}

	var results []string
	queryLower := strings.ToLower(query)

	err := filepath.Walk(path, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			name := info.Name()
			if strings.HasPrefix(name, ".") || name == "node_modules" || name == "vendor" {
				return filepath.SkipDir
			}
			return nil
		}
		if info.Size() > 1024*1024 { // 1MB limit
			return nil
		}

		data, err := os.ReadFile(filePath)
		if err != nil {
			return nil
		}

		lines := strings.Split(string(data), "\n")
		for i, line := range lines {
			if strings.Contains(strings.ToLower(line), queryLower) {
				relPath, _ := filepath.Rel(path, filePath)
				if relPath == "" {
					relPath = filePath
				}
				results = append(results, fmt.Sprintf("%s:%d: %s", relPath, i+1, strings.TrimSpace(line)))
				if len(results) >= 30 {
					return fmt.Errorf("max results reached")
				}
			}
		}
		return nil
	})

	if err != nil && err.Error() != "max results reached" {
		return "", fmt.Errorf("search error: %w", err)
	}

	if len(results) == 0 {
		return fmt.Sprintf("No matches found for %q in %s", query, path), nil
	}

	return fmt.Sprintf("Found %d match(es) for %q:\n\n%s", len(results), query, strings.Join(results, "\n")), nil
}

func copyFileSingle(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	dstDir := filepath.Dir(dst)
	if err := os.MkdirAll(dstDir, 0755); err != nil {
		return err
	}

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err = io.Copy(out, in); err != nil {
		return err
	}
	return out.Sync()
}

func copyDirRecursive(src, dst string) error {
	info, err := os.Stat(src)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(dst, info.Mode()); err != nil {
		return err
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			if err := copyDirRecursive(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			if err := copyFileSingle(srcPath, dstPath); err != nil {
				return err
			}
		}
	}
	return nil
}

func formatSize(bytes int64) string {
	switch {
	case bytes >= 1024*1024:
		return fmt.Sprintf("%.1f MB", float64(bytes)/(1024*1024))
	case bytes >= 1024:
		return fmt.Sprintf("%.1f KB", float64(bytes)/1024)
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}

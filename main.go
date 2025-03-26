package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"time"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"
)

func main() {
	rootCmd := &cobra.Command{
		Use:  "epghelper",
		Long: `helper tool for interacting with embedded postgres databases`,
	}

	lsCmd := &cobra.Command{
		Use:   "ls",
		Short: "List existing databases",
		Run:   list,
	}

	rmCmd := &cobra.Command{
		Use:   "rm",
		Short: "remove a database",
		Run:   remove,
	}
	rmCmd.Flags().BoolP("all", "a", false, "Remove all databases")

	connectCmd := &cobra.Command{
		Use:   "connect",
		Short: "connect to a database",
		Run:   connect,
	}

	connectCmd.Flags().BoolP("latest", "l", false, "Connect to the latest created database")

	rootCmd.AddCommand(lsCmd)
	rootCmd.AddCommand(rmCmd)
	rootCmd.AddCommand(connectCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// list lists all the local databases created.
func list(_ *cobra.Command, _ []string) {
	dbs, err := getdbEntries()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Sort databases by creation time, latest first
	sort.Slice(dbs, func(i, j int) bool {
		return dbs[i].created.Before(dbs[j].created)
	})

	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Port", "Created at", "Size", "URL"})
	var totalSize int64
	for _, db := range dbs {
		url := fmt.Sprintf("postgresql://postgres:postgres@localhost:%s/postgres", db.port)
		t.AppendRow(table.Row{db.port, db.created.Format(time.DateTime), formatSize(db.size), url})
		totalSize += db.size
	}
	t.AppendSeparator()
	t.AppendFooter(table.Row{"", "Total", formatSize(totalSize)})

	t.SetStyle(table.StyleLight)
	t.Render()
}

// remove one or all databases.
func remove(cmd *cobra.Command, args []string) {
	all, err := cmd.Flags().GetBool("all")
	if err != nil {
		fmt.Println("Error reading flag:", err)
		return
	}

	dbEntries, err := getdbEntries()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if all {
		msg := "Are you sure you want to delete all databases?"
		if ok := getConfirmation(msg); !ok {
			fmt.Println("Aborted.")
			return
		}
		for _, db := range dbEntries {
			if err := os.RemoveAll(db.fullPath); err != nil {
				fmt.Printf("failed to delete %s: %s\n", db.fullPath, err)
				os.Exit(1)
			}
			fmt.Printf("Deleted %s\n", db.port)
		}
		return
	}

	if len(args) < 1 {
		fmt.Println("Please specify the directory to remove.")
		return
	}

	port := args[0]
	msg := fmt.Sprintf("Are you sure you want to remove %s?", port)
	if ok := getConfirmation(msg); !ok {
		fmt.Println("Aborted.")
		return
	}

	var fullPath string
	for _, entry := range dbEntries {
		if entry.port == port {
			fullPath = entry.fullPath
			break
		}
	}
	if fullPath == "" {
		fmt.Printf("No database found with port %s\n", port)
		os.Exit(1)
	}
	if err := os.RemoveAll(fullPath); err != nil {
		fmt.Printf("failed to remove %s: %s\n", fullPath, err)
		os.Exit(1)
	}
	fmt.Printf("Deleted %s\n", port)
}

// connect to a database.
func connect(cmd *cobra.Command, args []string) {
	latest, err := cmd.Flags().GetBool("latest")
	if err != nil {
		fmt.Println("Error reading flag:", err)
		return
	}
	var port string
	if latest {
		dbs, err := getdbEntries()
		if err != nil {
			fmt.Println("Failed to get database entries:", err)
			return
		}

		if len(dbs) == 0 {
			fmt.Println("No databases found.")
			return
		}

		// Sort databases by creation time, latest first
		sort.Slice(dbs, func(i, j int) bool {
			return dbs[i].created.After(dbs[j].created)
		})

		// Use the port of the latest created database
		port = dbs[0].port
	} else {
		if len(args) < 1 {
			fmt.Println("Please specify the port to connect to.")
			return
		}
		port = args[0]
	}

	// Construct the psql command with the specified port
	psqlCmd := exec.Command("psql", "-h", "127.0.0.1", "-U", "postgres", "-p", port)

	// Set the command's standard input, output, and error to the current process's
	psqlCmd.Stdin = os.Stdin
	psqlCmd.Stdout = os.Stdout
	psqlCmd.Stderr = os.Stderr
	psqlCmd.Env = append(os.Environ(), "PGPASSWORD="+"postgres")

	// Run the psql command
	if err := psqlCmd.Run(); err != nil {
		fmt.Printf("Failed to execute psql: %v\n", err)
	}
}

type dbDetails struct {
	port     string
	created  time.Time
	fullPath string
	size     int64
}

func getdbEntries() ([]dbDetails, error) {
	dirPath, err := getDefaultDirPath()
	if err != nil {
		return nil, fmt.Errorf("failed to get user home directory: %w", err)
	}
	dir, err := os.Open(dirPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open %s: %w", dirPath, err)
	}

	entries, err := dir.Readdir(-1)
	if err != nil {
		return nil, fmt.Errorf("failed to read dir %s: %w", dirPath, err)
	}

	directories := make([]dbDetails, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() && entry.Name() != "extracted" {
			size, err := getDirSize(dirPath + entry.Name())
			if err != nil {
				fmt.Printf("failed to get size for %s: %s\n", entry.Name(), err)
				continue
			}
			directories = append(directories, dbDetails{
				port:     entry.Name(),
				created:  entry.ModTime(),
				fullPath: dirPath + entry.Name(),
				size:     size,
			})
		}
	}

	return directories, nil
}

func getDirSize(path string) (int64, error) {
	var size int64
	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})
	return size, err
}

func formatSize(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(size)/float64(div), "KMGTPE"[exp])
}

func getDefaultDirPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return home + "/.embedded-postgres-go/", nil
}

func getConfirmation(message string) bool {
	fmt.Printf("%s (y/n): ", message)
	var response string
	_, err := fmt.Scanln(&response)
	if err != nil {
		fmt.Println("failed to read input:", err)
		return false
	}

	if response == "y" || response == "yes" {
		return true
	}
	return false
}

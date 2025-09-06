package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"sync/atomic"

	"github.com/pkg/xattr"
	"github.com/urfave/cli/v2"
)

var (
	sourceXattrName string
	targetXattrName string
	recursive       bool
	verbose         bool
	debug           bool
	blockSize       int
	rowSize         int
)

func main() {
	app := &cli.App{
		Name:  "organize-tb-rpm.py",
		Usage: "Rename extended attributes of files",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "xattr",
				Aliases:     []string{"X"},
				Usage:       "Target extended attribute name",
				Destination: &targetXattrName,
				Required:    true,
			},
			&cli.StringFlag{
				Name:        "source-xattr",
				Aliases:     []string{"S"},
				Usage:       "Source extended attribute name",
				Destination: &sourceXattrName,
				Required:    true,
			},
			&cli.BoolFlag{
				Name:        "recursive",
				Aliases:     []string{"r"},
				Usage:       "Recurse into directories",
				Destination: &recursive,
			},
			&cli.BoolFlag{
				Name:        "verbose",
				Aliases:     []string{"v"},
				Usage:       "Enable verbose output",
				Destination: &verbose,
			},
			&cli.IntFlag{
				Name:        "block-size",
				Aliases:     []string{"b"},
				Usage:       "Number of files per block in verbose mode",
				Value:       10,
				Destination: &blockSize,
			},
			&cli.IntFlag{
				Name:        "row-size",
				Aliases:     []string{"R"},
				Usage:       "Number of blocks per row in verbose mode",
				Value:       10,
				Destination: &rowSize,
			},
			&cli.BoolFlag{
				Name:        "debug",
				Usage:       "Enable debug mode (disables verbose)",
				Destination: &debug,
			},
		},
		Action: func(c *cli.Context) error {
			if debug {
				verbose = false
			}

			if c.NArg() == 0 {
				return fmt.Errorf("at least one file or directory argument is required")
			}

			// Use a worker pool pattern for better performance with large numbers of files.
			numWorkers := runtime.NumCPU()
			if numWorkers == 0 {
				numWorkers = 4
			}

			var wg sync.WaitGroup
			filePaths := make(chan string, 100) // Buffered channel to hold file paths
			var processedCount int32

			// Start worker goroutines
			for i := 0; i < numWorkers; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					for path := range filePaths {
						processFile(path, &processedCount)
					}
				}()
			}

			// Walk the directories and send file paths to the channel
			for _, path := range c.Args().Slice() {
				info, err := os.Stat(path)
				if err != nil {
					log.Printf("Error stating %s: %v", path, err)
					continue
				}

				if info.IsDir() && recursive {
					err := filepath.WalkDir(path, func(p string, d os.DirEntry, err error) error {
						if err != nil {
							return err
						}
						if !d.IsDir() {
							filePaths <- p
						}
						return nil
					})
					if err != nil {
						log.Printf("Error walking directory %s: %v", path, err)
					}
				} else if !info.IsDir() {
					filePaths <- path
				}
			}

			close(filePaths) // Close the channel when all file paths have been sent
			wg.Wait()        // Wait for all workers to finish
			if verbose {
				fmt.Println()
			}
			return nil
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func processFile(path string, count *int32) {
	if debug {
		xattrValue, err := xattr.Get(path, sourceXattrName)
		if err != nil {
			return
		}
		fmt.Printf("%s %s %s\n", targetXattrName, string(xattrValue), path)
		return
	}

	currentCount := atomic.AddInt32(count, 1)

	if verbose {
		if currentCount > 0 && currentCount%int32(blockSize) == 0 {
			fmt.Print(" ")
		}
		if currentCount > 0 && currentCount%int32(blockSize*rowSize) == 0 {
			fmt.Printf(" [%d]\n", currentCount)
		}
	}

	info, err := os.Stat(path)
	if err != nil {
		if verbose {
			fmt.Print("!")
		}
		return
	}

	if info.Size() == 0 {
		if verbose {
			fmt.Print(".")
		}
		return
	}

	err = renameXattr(path, sourceXattrName, targetXattrName)
	if err != nil {
		if verbose {
			fmt.Print("!")
		}
	} else {
		if verbose {
			fmt.Print("+")
		}
	}
}

func renameXattr(path, oldName, newName string) error {
	return renameXattrOS(path, oldName, newName)
}

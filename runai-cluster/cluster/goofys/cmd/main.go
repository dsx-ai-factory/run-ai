/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package main

import (
	goofys "github.com/kahing/goofys/api"
	. "github.com/kahing/goofys/api/common"
	. "github.com/kahing/goofys/internal"

	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"context"

	"github.com/jacobsa/fuse"
	"github.com/kardianos/osext"
	"github.com/urfave/cli"

	daemon "github.com/sevlyar/go-daemon"

	_ "github.com/dsx-ai-factory/run-ai/runai-common-packages/fips"
)

var log = GetLogger("main")

func registerSIGINTHandler(fs *Goofys, flags *FlagStorage) {
	// Register for SIGINT.
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM, syscall.SIGUSR1)

	// Start a goroutine that will unmount when the signal is received.
	go func() {
		for {
			s := <-signalChan
			if s == syscall.SIGUSR1 {
				log.Infof("Received %v", s)
				fs.SigUsr1()
				continue
			}

			if len(flags.Cache) == 0 {
				log.Infof("Received %v, attempting to unmount...", s)

				err := TryUnmount(flags.MountPoint)
				if err != nil {
					log.Errorf("Failed to unmount in response to %v: %v", s, err)
				} else {
					log.Printf("Successfully unmounted %v in response to %v",
						flags.MountPoint, s)
					return
				}
			} else {
				log.Infof("Received %v", s)
				// wait for catfs to die and cleanup
			}
		}
	}()
}

var waitedForSignal os.Signal

func waitForSignal(wg *sync.WaitGroup) {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGUSR1, syscall.SIGUSR2)

	wg.Add(1)
	go func() {
		waitedForSignal = <-signalChan
		wg.Done()
	}()
}

func kill(pid int, s os.Signal) (err error) {
	p, err := os.FindProcess(pid)
	if err != nil {
		return err
	}

	defer p.Release()

	err = p.Signal(s)
	if err != nil {
		return err
	}
	return
}

var mount = goofys.Mount
var unmount = syscall.Unmount

func clearReadyFile(path string) error {
	if path == "" {
		return nil
	}
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("clear readiness file: %w", err)
	}
	return nil
}

func massagePath() {
	for _, e := range os.Environ() {
		if strings.HasPrefix(e, "PATH=") {
			return
		}
	}

	// mount -a seems to run goofys without PATH
	// usually fusermount is in /bin
	os.Setenv("PATH", "/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin")
}

func massageArg0() {
	var err error
	os.Args[0], err = osext.Executable()
	if err != nil {
		panic(fmt.Sprintf("Unable to discover current executable: %v", err))
	}
}

var Version = "use `make build' to fill version hash correctly"

func main() {
	VersionNumber = "0.24.0"
	VersionHash = Version

	massagePath()

	app := NewApp()
	app.Flags = append(app.Flags,
		cli.StringFlag{Name: "ready-file", Usage: "Create a local readiness file after mounting successfully (requires -f)."},
		cli.StringFlag{Name: "check-ready", Usage: "Check readiness and the current goofys mount: --check-ready /ready/file /mount/path."},
	)

	var flags *FlagStorage
	var child *os.Process

	app.Action = func(c *cli.Context) (err error) {
		if c.IsSet("check-ready") {
			if !filepath.IsAbs(c.String("check-ready")) || c.NArg() != 1 || c.Args().First() == "" || c.IsSet("ready-file") {
				return fmt.Errorf("usage: goofys --check-ready /absolute/readiness/file mount/path")
			}
			mountPoint, err := filepath.Abs(c.Args().First())
			if err != nil {
				return fmt.Errorf("resolve mount point: %w", err)
			}
			info, err := os.Stat(c.String("check-ready"))
			if err != nil {
				return fmt.Errorf("check readiness file: %w", err)
			}
			if !info.Mode().IsRegular() {
				return fmt.Errorf("readiness file must be a regular file")
			}
			mountInfo, err := readMountInfo()
			if err != nil {
				return fmt.Errorf("read current mounts: %w", err)
			}
			return checkGoofysMount(string(mountInfo), mountPoint)
		}
		// We should get two arguments exactly. Otherwise error out.
		if len(c.Args()) != 2 {
			fmt.Fprintf(
				os.Stderr,
				"Error: %s takes exactly two arguments.\n\n",
				app.Name)
			cli.ShowAppHelp(c)
			os.Exit(1)
		}

		readyFile := c.String("ready-file")
		if c.IsSet("ready-file") && (!filepath.IsAbs(readyFile) || !c.Bool("f")) {
			return fmt.Errorf("--ready-file requires an absolute path and -f")
		}
		if readyFile != "" {
			mountPoint, err := filepath.Abs(c.Args()[1])
			if err != nil {
				return fmt.Errorf("resolve mount point: %w", err)
			}
			relative, err := filepath.Rel(mountPoint, filepath.Clean(readyFile))
			if err != nil {
				return fmt.Errorf("locate readiness file outside mount point: %w", err)
			}
			if relative != ".." && !strings.HasPrefix(relative, ".."+string(os.PathSeparator)) {
				return fmt.Errorf("readiness file must be outside the mount point")
			}
		}
		// A previous mount must never satisfy this process's startup probe.
		if err := clearReadyFile(readyFile); err != nil {
			return err
		}

		// Populate and parse flags.
		bucketName := c.Args()[0]
		flags = PopulateFlags(c)
		if flags == nil {
			cli.ShowAppHelp(c)
			err = fmt.Errorf("invalid arguments")
			return
		}
		defer func() {
			time.Sleep(time.Second)
			flags.Cleanup()
		}()
		defer func() {
			if err := clearReadyFile(readyFile); err != nil {
				log.Error(err)
			}
		}()

		if !flags.Foreground {
			var wg sync.WaitGroup
			waitForSignal(&wg)

			massageArg0()

			ctx := new(daemon.Context)
			child, err = ctx.Reborn()

			if err != nil {
				panic(fmt.Sprintf("unable to daemonize: %v", err))
			}

			InitLoggers(!flags.Foreground && child == nil)

			if child != nil {
				// attempt to wait for child to notify parent
				wg.Wait()
				if waitedForSignal == syscall.SIGUSR1 {
					return
				} else {
					return fuse.EINVAL
				}
			} else {
				// kill our own waiting goroutine
				kill(os.Getpid(), syscall.SIGUSR1)
				wg.Wait()
				defer ctx.Release()
			}

		} else {
			InitLoggers(!flags.Foreground)
		}

		// Mount the file system.
		var mfs *fuse.MountedFileSystem
		var fs *Goofys
		fs, mfs, err = mount(
			context.Background(),
			bucketName,
			flags)

		if err != nil {
			if !flags.Foreground {
				kill(os.Getppid(), syscall.SIGUSR2)
			}
			log.Fatalf("Mounting file system: %v", err)
			// fatal also terminates itself
		} else {
			if !flags.Foreground {
				kill(os.Getppid(), syscall.SIGUSR1)
			}
			log.Println("File system has been successfully mounted.")
			if readyFile != "" {
				if err := os.WriteFile(readyFile, nil, 0600); err != nil {
					// The scratch image has no fusermount helper.
					if unmountErr := unmount(flags.MountPoint, 0); unmountErr != nil {
						log.Errorf("Failed to unmount after readiness error: %v", unmountErr)
					}
					return fmt.Errorf("create readiness file: %w", err)
				}
			}
			// Let the user unmount with Ctrl-C
			// (SIGINT). But if cache is on, catfs will
			// receive the signal and we would detect that exiting
			registerSIGINTHandler(fs, flags)

			// Wait for the file system to be unmounted.
			err = mfs.Join(context.Background())
			if err != nil {
				err = fmt.Errorf("MountedFileSystem.Join: %v", err)
				return
			}

			log.Println("Successfully exiting.")
		}
		return
	}

	err := app.Run(MassageMountFlags(os.Args))
	if err != nil {
		log.Error(err)
		if flags != nil && !flags.Foreground && child != nil {
			log.Fatalln("Unable to mount file system, see syslog for details")
		}
		os.Exit(1)
	}
}

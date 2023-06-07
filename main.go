package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"sync"

	"github.com/davecgh/go-spew/spew"
)

type Command struct {
	Command string
	Count   int
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "pl <command> [-c]")
		os.Exit(1)
		return
	}

	// commands =
	mtx := &sync.Mutex{}

	// processCount, err := strconv.Atoi(os.Args[1])
	// check(mtx, err, "%s is not a valid number: %v", os.Args[1], err)
	cmds := []exec.Cmd{}

	wg := &sync.WaitGroup{}
	for _, args := range commands(os.Args[1:]) {
		cmd := exec.Command(args[0], args[1:]...)
		cmds = append(cmds, *cmd)

		stdout, err := cmd.StdoutPipe()
		check(mtx, err, "failed to open stdout pipe: %v", err)

		go func() {
			err := copyOutput(mtx, args[0], stdout, os.Stdout)
			check(mtx, err, "error in stdout")
		}()

		stderr, err := cmd.StderrPipe()
		check(mtx, err, "failed to open stdout pipe: %v", err)

		go func() {
			err := copyOutput(mtx, args[0], stderr, os.Stderr)
			check(mtx, err, "error in stdout")
		}()

		err = cmd.Start()
		check(mtx, err, "failed to start command: %v", err)

		// defer func() {
		// 	err := cmd.Process.Kill()

		// 	if errors.Is(err, os.ErrProcessDone) {
		// 	} else if err != nil {
		// 		fmt.Fprintf(os.Stderr, "there was an error stopping process %d %v\n", cmd.Process.Pid, err)
		// 	}
		// 	fmt.Printf("killing %s\n", args[0])
		// }()

		wg.Add(1)
		go func(name string) {
			err := cmd.Wait()
			// spew.Dump(err)
			if exitErr, ok := err.(*exec.ExitError); ok {
				spew.Dump(exitErr.ProcessState.ExitCode())
			} else if err != nil {
				fmt.Fprintf(os.Stderr, "process %s stopped with an error: %v\n", name, err)
			}
			fmt.Printf("exit %s\n", name)
			wg.Done()
		}(args[0])
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		tries := 0
		for range c {
			if tries == 0 {
				go func() {
					fmt.Fprintf(os.Stderr, "Gracefully stopping services Ctrl+C again to force\n")
					for _, cmd := range cmds {
						fmt.Fprintf(os.Stderr, "killing %s\n", cmd.Args[0])
						if cmd.Process == nil {
							continue
						}
						err := cmd.Process.Kill()

						if errors.Is(err, os.ErrProcessDone) {
						} else if err != nil {
							fmt.Fprintf(os.Stderr, "there was an error stopping process %d %v\n", cmd.Process.Pid, err)
						} else {
							fmt.Fprintf(os.Stderr, "killed %s\n", cmd.Args[0])
						}
					}
					os.Exit(1)
				}()
			} else {
				log.Print("Force shutting down server")
				os.Exit(1)
			}
			tries++
		}
	}()
	wg.Wait()
}

func copyOutput(mtx *sync.Mutex, command string, src io.Reader, dst io.Writer) error {
	bufSrc := bufio.NewReader(src)
	for {

		line, _, err := bufSrc.ReadLine()
		if errors.Is(err, io.EOF) {
			return nil
		}
		if errors.Is(err, os.ErrClosed) {
			return nil
		}
		if err != nil {
			return fmt.Errorf("failed to read line: %w", err)
		}

		mtx.Lock()
		_, err = fmt.Fprintf(dst, "%6s%s\n", fmt.Sprintf("%s | ", command), line)
		mtx.Unlock()
		if err != nil {
			return fmt.Errorf("failed to write output: %w", err)
		}
	}
}

func check(mtx *sync.Mutex, err error, message string, a ...interface{}) {
	if err != nil {
		mtx.Lock()
		defer mtx.Unlock()
		fmt.Fprintf(os.Stderr, message+"\n", a...)
		os.Exit(1)
	}
}

func commands(args []string) [][]string {
	commandArgs := [][]string{}
	currentCommandArgs := []string{}
	for _, arg := range args {
		if arg == ":" {
			commandArgs = append(commandArgs, currentCommandArgs)
			currentCommandArgs = []string{}
			continue
		}

		currentCommandArgs = append(currentCommandArgs, arg)
	}
	commandArgs = append(commandArgs, currentCommandArgs)
	return commandArgs
}

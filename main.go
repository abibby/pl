package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"sync"

	"github.com/pkg/errors"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Fprintf(os.Stderr, "pl <count> <command>")
		os.Exit(1)
		return
	}

	mtx := &sync.Mutex{}

	processCount, err := strconv.Atoi(os.Args[1])
	check(mtx, err, "%s is not a valid number: %v", os.Args[1], err)

	wg := &sync.WaitGroup{}

	for i := 0; i < processCount; i++ {
		cmd := exec.Command(os.Args[2], os.Args[3:]...)

		stdout, err := cmd.StdoutPipe()
		check(mtx, err, "failed to open stdout pipe: %v", err)

		go func() {
			err := copyOutput(mtx, i, stdout, os.Stdout)
			check(mtx, err, "error in stdout")
		}()

		stderr, err := cmd.StderrPipe()
		check(mtx, err, "failed to open stdout pipe: %v", err)

		go func() {
			err := copyOutput(mtx, i, stderr, os.Stderr)
			check(mtx, err, "error in stdout")
		}()

		err = cmd.Start()
		check(mtx, err, "failed to start command: %v", err)

		defer func() {
			err := cmd.Process.Kill()
			if err.Error() == "os: process already finished" {

			} else if err != nil {
				fmt.Fprintf(os.Stderr, "there was an error stopping a process %v\n", err)
			}
		}()

		wg.Add(1)
		go func() {
			cmd.Wait()
			wg.Done()
		}()
	}

	wg.Wait()
}

func copyOutput(mtx *sync.Mutex, num int, src io.Reader, dst io.Writer) error {
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
			return errors.Wrap(err, "failed to read line")
		}

		mtx.Lock()
		_, err = fmt.Fprintf(dst, "%6s%s\n", fmt.Sprintf("%d | ", num), line)
		mtx.Unlock()
		if err != nil {
			return errors.Wrap(err, "failed to write output")
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

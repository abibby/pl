package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"sync"
)

func main() {
	processCount, err := strconv.Atoi(os.Args[1])
	check(err, "%s is not a valid number: %v", os.Args[1], err)

	wg := &sync.WaitGroup{}

	for i := 0; i < processCount; i++ {
		cmd := exec.Command(os.Args[2], os.Args[3:]...)

		stdout, err := cmd.StdoutPipe()
		check(err, "failed to open stdout pipe: %v", err)

		go copyOutput(i, stdout, os.Stdout)

		stderr, err := cmd.StderrPipe()
		check(err, "failed to open stdout pipe: %v", err)

		go copyOutput(i, stderr, os.Stderr)

		err = cmd.Start()
		check(err, "failed to start command: %v", err)

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

func copyOutput(num int, src io.Reader, dst io.Writer) {
	bufSrc := bufio.NewReader(src)
	for {

		line, _, err := bufSrc.ReadLine()
		if err == io.EOF {
			return
		}
		check(err, "failed to read line: %v\n", err)
		fmt.Fprintf(dst, "%6s%s\n", fmt.Sprintf("%d | ", num), line)
		// io.Copy(dst, src)
	}
}

func check(err error, message string, a ...interface{}) {
	if err != nil {
		fmt.Fprintf(os.Stderr, message+"\n", a...)
		os.Exit(1)
	}
}

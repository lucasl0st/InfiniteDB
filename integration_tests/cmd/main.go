/*
 * Copyright (c) 2023 Lucas Pape
 */

package main

import (
	"errors"
	"fmt"
	"github.com/lucasl0st/InfiniteDB/client"
	"github.com/lucasl0st/InfiniteDB/integration_tests"
	"io"
	"log"
	"os"
	"os/exec"
	"time"
)

const Port = 8099

func main() {
	args := os.Args[1:]

	if len(args) < 2 {
		log.Panic(errors.New("at least two arguments required"))
	}

	serverBinary := args[0]

	if len(serverBinary) == 0 {
		log.Panic(errors.New("no argument server-binary"))
	}

	coverageOutDir := args[1]

	if len(coverageOutDir) == 0 {
		log.Panic(errors.New("no argument coverage-out-dir"))
	}

	cmd, err := setup(serverBinary, coverageOutDir)

	if err != nil {
		log.Panic(err.Error())
	}

	err = runTests()

	failed := false

	if err != nil {
		fmt.Println(err.Error())
		failed = true
	}

	time.Sleep(time.Second * 3)

	err = kill(cmd)

	if err != nil {
		log.Panic(err.Error())
	}

	if failed {
		os.Exit(1)
	}
}

func setup(serverBinary string, coverageOutDir string) (*exec.Cmd, error) {
	dir, err := os.MkdirTemp("", "infinitedb-integration-testing")

	if err != nil {
		return nil, err
	}

	dir += "/"

	log.Printf("using temp directory %s\n", dir)

	cmd := exec.Command(serverBinary)

	cmd.Env = []string{
		fmt.Sprintf("GOCOVERDIR=%s", coverageOutDir),
		fmt.Sprintf("DATABASE_PATH=%s", dir),
		fmt.Sprintf("PORT=%v", Port),
		"AUTHENTICATION=false",
	}

	stdout, err := cmd.StdoutPipe()

	if err != nil {
		return nil, err
	}

	stderr, err := cmd.StderrPipe()

	if err != nil {
		return nil, err
	}

	err = cmd.Start()

	if err != nil {
		return nil, errors.New(fmt.Sprintf("failed to start infinitedb server: %s", err.Error()))
	}

	buf1 := make([]byte, 1024)
	buf2 := make([]byte, 1024)

	go func() {
		for {
			n, err := stdout.Read(buf1)

			if err != nil {
				if err == io.EOF {
					break
				}
				fmt.Println("Error reading from stdout:", err)
				break
			}
			fmt.Print(string(buf1[:n]))
		}
	}()

	go func() {
		for {
			n, err := stderr.Read(buf2)

			if err != nil {
				if err == io.EOF {
					break
				}
				fmt.Println("Error reading from stdout:", err)
				break
			}
			fmt.Print(string(buf2[:n]))
		}
	}()

	log.Println("waiting 3 seconds for database to start up")

	time.Sleep(time.Second * 3)

	log.Printf("started %s on port %v\n", serverBinary, Port)

	return cmd, nil
}

func kill(cmd *exec.Cmd) error {
	err := cmd.Process.Kill()

	if err != nil {
		return errors.New(fmt.Sprintf("error killing infinitedb-server: %s", err.Error()))
	}

	log.Println("killed infinitedb-server")

	return nil
}

func runTests() error {
	c := client.New(client.Options{
		Hostname: "127.0.0.1",
		Port:     Port,
	})

	err := c.Connect()

	if err != nil {
		return errors.New(fmt.Sprintf("failed to connect to server: %s", err.Error()))
	}

	for _, test := range integration_tests.Tests {
		log.Printf("running test %s\n", test.Name)

		err := test.Run(c)

		if err != nil {
			fmt.Printf("test %s failed with error: %s", test.Name, err.Error())
		}
	}

	err = c.ShutdownServer()

	if err != nil {
		return err
	}

	return nil
}

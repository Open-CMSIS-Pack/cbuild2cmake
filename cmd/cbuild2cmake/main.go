/*
 * Copyright (c) 2024 Arm Limited. All rights reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package main

import (
	"fmt"
	"io"
	"os"

	"github.com/Open-CMSIS-Pack/cbuild2cmake/cmd/cbuild2cmake/commands"

	log "github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/writer"
)

func main() {
	log.SetFormatter(new(LogFormatter))
	log.SetOutput(io.Discard)
	initLogHooks()

	if len(version) == 0 {
		version = "0.0.0-debug"
	}
	commands.Version = version
	commands.CopyrightNotice = copyrightNotice

	cmd := commands.NewRootCmd()
	err := cmd.Execute()
	if err != nil {
		log.Error(err)
		os.Exit(-1)
	}
}

type LogFormatter struct{}

func (s *LogFormatter) Format(entry *log.Entry) ([]byte, error) {
	msg := fmt.Sprintf("%s cbuild2cmake: %s\n", entry.Level.String(), entry.Message)
	return []byte(msg), nil
}

func initLogHooks() {
	log.AddHook(&writer.Hook{ // Send logs with level higher than warning to stderr
		Writer: os.Stderr,
		LogLevels: []log.Level{
			log.PanicLevel,
			log.FatalLevel,
			log.ErrorLevel,
			log.WarnLevel,
		},
	})
	log.AddHook(&writer.Hook{ // Send info and debug logs to stdout
		Writer: os.Stdout,
		LogLevels: []log.Level{
			log.InfoLevel,
			log.DebugLevel,
		},
	})
}

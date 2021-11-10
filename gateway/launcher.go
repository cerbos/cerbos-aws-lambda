// Copyright 2021 Zenauth Ltd.
// SPDX-License-Identifier: Apache-2.0

package gateway

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

const (
	healthCheckInterval = time.Millisecond * 15
	httpListenAddr      = ":3592"
	grpcListenAddr      = ":3593"
)

type launcher struct {
	httpClient     *http.Client
	process        *os.Process
	healthEndpoint string
}

func newLauncher(httpClient *http.Client, healthEndpoint string) *launcher {
	return &launcher{
		httpClient:     httpClient,
		healthEndpoint: healthEndpoint,
	}
}

func (l *launcher) StartProcess(ctx context.Context, cerbos, workDir, configFile string) (err error) {
	if l.Started() {
		return nil
	}
	logLevel := os.Getenv("CERBOS_LOG_LEVEL")
	if logLevel == "" {
		logLevel = "INFO"
	}
	argv := []string{"cerbos", "server", "--config=" + configFile, "--log-level=" + logLevel, "--set=server.httpListenAddr=" + httpListenAddr, "--set=server.grpcListenAddr=" + grpcListenAddr}
	l.process, err = os.StartProcess(cerbos, argv, &os.ProcAttr{
		Dir:   workDir,
		Files: []*os.File{os.Stdin, os.Stdout, os.Stderr},
	})
	if err != nil {
		return err
	}
	request, err := http.NewRequestWithContext(ctx, "GET", l.healthEndpoint, nil)
	if err != nil {
		return err
	}
	now := time.Now()
	time.Sleep(healthCheckInterval)
	for i := 0; i < 50; i++ {
		resp, err := l.httpClient.Do(request)
		log.Printf("starting cerbos health check: %v, pid: %v", err, l.process.Pid)
		if resp != nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				log.Printf("Cerbos server started in %s", time.Since(now))
				go func() {
					ps, err := l.process.Wait()
					log.Printf("Cerbos process exited: state %q, err %q", ps.String(), err)
					l.process = nil
					os.Exit(1)
				}()
				return nil
			}
		}
		time.Sleep(healthCheckInterval)
	}

	return fmt.Errorf("in %v: %w", time.Since(now), ErrNotStarted)
}

func (l *launcher) Started() bool {
	return l.process != nil
}

func (l *launcher) StopProcess() error {
	if l.process != nil {
		err := l.process.Kill()
		if err != nil {
			log.Printf("failed to kill cerbos process: %v", err)
			return err
		}
		l.process = nil
	}
	return nil
}

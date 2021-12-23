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
	httpListenAddr = ":3592"
	grpcListenAddr = ":3593"

	cerbosLogLevelEnv            = "CERBOS_LOG_LEVEL"
	cerbosLaunchTimeoutEnv       = "CERBOS_LAUNCH_TIMEOUT"
	cerbosHealthCheckIntervalEnv = "CERBOS_HEALTH_CHECK_INTERVAL"

	cerbosLaunchTimeoutDefault       = 2 * time.Second
	cerbosHealthCheckIntervalDefault = 50 * time.Millisecond
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
	logLevel := os.Getenv(cerbosLogLevelEnv)
	if logLevel == "" {
		logLevel = "INFO"
	}

	timeout := parseDurationOrDefault(os.Getenv(cerbosLaunchTimeoutEnv), cerbosLaunchTimeoutDefault)
	log.Printf("cerbos launch timeout: %s", timeout)
	healthCheckInterval := parseDurationOrDefault(os.Getenv(cerbosHealthCheckIntervalEnv), cerbosHealthCheckIntervalDefault)
	log.Printf("health check interval: %s", healthCheckInterval)
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
	startTime := time.Now()
	for {
		time.Sleep(healthCheckInterval)
		resp, err := l.httpClient.Do(request)
		log.Printf("cerbos health check: %v, pid: %v", err, l.process.Pid)
		if resp != nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				log.Printf("Cerbos server started in %s", time.Since(startTime))
				go func() {
					ps, err := l.process.Wait()
					log.Printf("Cerbos process exited: state %q, err %q", ps.String(), err)
					l.process = nil
					os.Exit(1)
				}()
				return nil
			}
		}
		if time.Since(startTime) > timeout {
			break
		}
	}

	return fmt.Errorf("in %v: %w", time.Since(startTime), ErrNotStarted)
}

func parseDurationOrDefault(v string, d time.Duration) time.Duration {
	if v == "" {
		return d
	}
	res, err := time.ParseDuration(v)
	if err != nil {
		return d
	}
	return res
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

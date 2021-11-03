// Copyright 2021 Zenauth Ltd.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"log"

	"github.com/aws/aws-lambda-go/lambda"

	"github.com/cerbos/cerbos-aws-lambda/gateway"
)

const cerbosHTTPAddr = "http://127.0.0.1:3592"

func main() {
	gw, err := gateway.NewGateway(cerbosHTTPAddr)
	if err != nil {
		log.Print("failed to create a gateway")
		return
	}
	ctx := context.Background()
	err = gw.StartProcess(ctx, "cerbos", "", "conf.yml")
	if err != nil {
		log.Printf("Failed to start a process: %s", err)
		return
	}
	// start lambda handler
	lambda.StartHandler(gw)
}

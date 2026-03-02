/*
 * Licensed to the Apache Software Foundation (ASF) under one or more
 * contributor license agreements.  See the NOTICE file distributed with
 * this work for additional information regarding copyright ownership.
 * The ASF licenses this file to You under the Apache License, Version 2.0
 * (the "License"); you may not use this file except in compliance with
 * the License.  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"seata.apache.org/seata-go-samples/tcc/rocketmq/service"
	"seata.apache.org/seata-go/v2/pkg/client"
	"seata.apache.org/seata-go/v2/pkg/tm"
	"seata.apache.org/seata-go/v2/pkg/util/log"
)

var (
	mode = flag.String("mode", "commit", "Transaction mode: commit or rollback")
)

func main() {
	flag.Parse()

	// Initialize Seata client
	client.InitPath("../../../conf/seatago.yml")

	// Execute global transaction
	err := tm.WithGlobalTx(context.Background(), &tm.GtxConfig{
		Name:    "RocketMQTCCSampleGlobalTx",
		Timeout: 60000,
	}, func(ctx context.Context) error {
		return business(ctx, *mode)
	})

	if err != nil {
		log.Errorf("Global transaction failed: %v", err)
		os.Exit(1)
	}

	log.Infof("Global transaction completed successfully in %s mode", *mode)
	os.Exit(0)
}

func business(ctx context.Context, mode string) error {
	// Log transaction context
	log.Infof("Starting global transaction, XID: %s", tm.GetXID(ctx))
	log.Infof("Executing in %s mode", mode)

	// Get TCC service proxy
	svc := service.NewRocketMQTCCServiceProxy()

	// Call Prepare phase with params
	params := map[string]interface{}{
		"mode":    mode,
		"message": "test message",
	}
	_, err := svc.Prepare(ctx, params)
	if err != nil {
		log.Errorf("Prepare failed: %v", err)
		return err
	}

	log.Infof("Business logic completed successfully")

	// For rollback mode, return error after successful Prepare
	if mode == "rollback" {
		return fmt.Errorf("simulated rollback scenario")
	}

	return nil
}

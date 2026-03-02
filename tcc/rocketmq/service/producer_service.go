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

package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/apache/rocketmq-client-go/v2/primitive"

	"seata.apache.org/seata-go/v2/pkg/integration/rocketmq"
	"seata.apache.org/seata-go/v2/pkg/rm/tcc"
	"seata.apache.org/seata-go/v2/pkg/rm/tcc/fence"
	"seata.apache.org/seata-go/v2/pkg/tm"
	"seata.apache.org/seata-go/v2/pkg/util/log"
)

const (
	// Fence database connection (using fence driver)
	FenceDriverName = "seata-fence-mysql"
	FenceURL        = "root:root@tcp(127.0.0.1:3306)/seata?charset=utf8&parseTime=True"
)

var (
	tccService     *tcc.TCCServiceProxy
	tccServiceOnce sync.Once

	producerInitOnce sync.Once
	producerInitErr  error
)

// RocketMQTCCService implements TCC pattern for RocketMQ transactional messaging
type RocketMQTCCService struct {
	producer *rocketmq.SeataMQProducer
}

type prepareMessageBody struct {
	Action    string      `json:"action"`
	Timestamp int64       `json:"timestamp"`
	Params    interface{} `json:"params"`
}

// NewRocketMQTCCServiceProxy creates a singleton TCC service proxy
func NewRocketMQTCCServiceProxy() *tcc.TCCServiceProxy {
	if tccService != nil {
		return tccService
	}
	tccServiceOnce.Do(func() {
		var err error
		svc := &RocketMQTCCService{}
		if initErr := svc.InitProducer(); initErr != nil {
			err = fmt.Errorf("init RocketMQ producer failed: %w", initErr)
			return
		}
		tccService, err = tcc.NewTCCServiceProxy(svc)
		if err != nil {
			panic(fmt.Errorf("get RocketMQTCCService tcc service proxy error: %v", err))
		}
	})
	return tccService
}

func (s *RocketMQTCCService) InitProducer() error {
	producerInitOnce.Do(func() {
		cfg := rocketmq.NewDefaultSeataMQProducerConfig()
		cfg.NameServerAddrs = []string{"127.0.0.1:9876"}
		cfg.GroupName = "seata-tcc-producer-group"
		cfg.InstanceName = "seata-tcc-producer-instance"

		producer, err := rocketmq.NewSeataMQProducer(cfg)
		if err != nil {
			producerInitErr = fmt.Errorf("create seata RocketMQ producer failed: %w", err)
			return
		}

		if err = producer.Start(); err != nil {
			producerInitErr = fmt.Errorf("start seata RocketMQ producer failed: %w", err)
			return
		}

		s.producer = producer
		log.Infof("RocketMQ producer initialized successfully")
	})

	if producerInitErr != nil {
		log.Errorf("RocketMQ producer initialization failed: %v", producerInitErr)
		return producerInitErr
	}

	if s.producer == nil {
		return fmt.Errorf("rocketmq producer is nil after initialization")
	}

	return nil
}

func (s *RocketMQTCCService) Shutdown() error {
	if s.producer == nil {
		return nil
	}

	if err := s.producer.Shutdown(); err != nil {
		log.Errorf("RocketMQ producer shutdown failed: %v", err)
		return fmt.Errorf("shutdown rocketmq producer failed: %w", err)
	}

	log.Infof("RocketMQ producer shutdown successfully")
	return nil
}

// Prepare phase: Send RocketMQ half-message (not consumable yet)
func (s *RocketMQTCCService) Prepare(ctx context.Context, params interface{}) (bool, error) {
	db, err := sql.Open(FenceDriverName, FenceURL)
	if err != nil {
		return false, fmt.Errorf("fence database connect failed: %w", err)
	}
	defer func() { _ = db.Close() }()

	tx, err := db.Begin()
	if err != nil {
		return false, fmt.Errorf("fence transaction begin failed: %w", err)
	}

	defer func() {
		if err != nil {
			err = fmt.Errorf("prepare phase error: %w, rollback result: %v", err, tx.Rollback())
			return
		}
		err = tx.Commit()
	}()

	// Fence wrapper provides idempotency protection
	err = fence.WithFence(ctx, tx, func() error {
		log.Infof("RocketMQTCCService Prepare phase, params: %v", params)

		if s.producer == nil {
			return fmt.Errorf("rocketmq producer is not initialized")
		}

		payload, marshalErr := json.Marshal(&prepareMessageBody{
			Action:    "prepare",
			Timestamp: time.Now().Unix(),
			Params:    params,
		})
		if marshalErr != nil {
			log.Errorf("RocketMQTCCService Prepare phase marshal payload failed: %v", marshalErr)
			return fmt.Errorf("marshal prepare message payload failed: %w", marshalErr)
		}

		msg := primitive.NewMessage("seata-tcc-test", payload)
		msg.WithTag("TCC_PREPARE")

		sendResult, sendErr := s.producer.Send(ctx, msg)
		if sendErr != nil {
			log.Errorf("RocketMQTCCService Prepare phase send transaction message failed: %v", sendErr)
			return fmt.Errorf("send transaction message in global transaction failed: %w", sendErr)
		}

		log.Infof("RocketMQTCCService Prepare phase send transaction message success, msgID=%s", sendResult.MsgID)

		return nil
	})

	if err != nil {
		return false, err
	}
	return true, nil
}

// Commit phase: Make message consumable
func (s *RocketMQTCCService) Commit(ctx context.Context, bac *tm.BusinessActionContext) (bool, error) {
	db, err := sql.Open(FenceDriverName, FenceURL)
	if err != nil {
		return false, fmt.Errorf("fence database connect failed: %w", err)
	}
	defer func() { _ = db.Close() }()

	tx, err := db.Begin()
	if err != nil {
		return false, fmt.Errorf("fence transaction begin failed: %w", err)
	}

	defer func() {
		if err != nil {
			err = fmt.Errorf("commit phase error: %w, rollback result: %v", err, tx.Rollback())
			return
		}
		err = tx.Commit()
	}()

	err = fence.WithFence(ctx, tx, func() error {
		log.Infof("RocketMQTCCService Commit phase, context: %v", bac)

		// Message commit is handled by RocketMQ TransactionListener
		// This phase just records fence status

		return nil
	})

	if err != nil {
		return false, err
	}
	return true, nil
}

// Rollback phase: Delete/cancel message
func (s *RocketMQTCCService) Rollback(ctx context.Context, bac *tm.BusinessActionContext) (bool, error) {
	db, err := sql.Open(FenceDriverName, FenceURL)
	if err != nil {
		return false, fmt.Errorf("fence database connect failed: %w", err)
	}
	defer func() { _ = db.Close() }()

	tx, err := db.Begin()
	if err != nil {
		return false, fmt.Errorf("fence transaction begin failed: %w", err)
	}

	defer func() {
		if err != nil {
			err = fmt.Errorf("rollback phase error: %w, rollback result: %v", err, tx.Rollback())
			return
		}
		err = tx.Commit()
	}()

	err = fence.WithFence(ctx, tx, func() error {
		log.Infof("RocketMQTCCService Rollback phase, context: %v", bac)

		// Message rollback is handled by RocketMQ TransactionListener
		// This phase just records fence status

		return nil
	})

	if err != nil {
		return false, err
	}
	return true, nil
}

// GetActionName returns unique action name for this TCC service
func (s *RocketMQTCCService) GetActionName() string {
	return "RocketMQTCCService"
}

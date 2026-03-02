<!--
  ~ Licensed to the Apache Software Foundation (ASF) under one or more
  ~ contributor license agreements.  See the NOTICE file distributed with
  ~ this work for additional information regarding copyright ownership.
  ~ The ASF licenses this file to You under the Apache License, Version 2.0
  ~ (the "License"); you may not use this file except in compliance with
  ~ the License.  You may obtain a copy of the License at
  ~
  ~     http://www.apache.org/licenses/LICENSE-2.0
  ~
  ~ Unless required by applicable law or agreed to in writing, software
  ~ distributed under the License is distributed on an "AS IS" BASIS,
  ~ WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
  ~ See the License for the specific language governing permissions and
  ~ limitations under the License.
  -->

# RocketMQ TCC Sample

This sample demonstrates how to use Seata-Go's RocketMQ TCC integration for distributed transactional messaging.

## Use Case Description

This sample showcases:
- Sending RocketMQ transactional messages within Seata global transactions
- TCC (Try-Confirm-Cancel) pattern for message reliability
- Idempotency protection using TCC fence mechanism
- Both commit and rollback scenarios

## Prerequisites

1. **MySQL** (for TCC fence log)
   - Version: 5.7+
   - Database: `seata`
   - User: `root` / Password: `root`

2. **Seata TC Server**
   - Version: Compatible with seata-go v2.x
   - Address: `127.0.0.1:8091`

3. **RocketMQ**
   - Version: 4.x or 5.x
   - NameServer: `127.0.0.1:9876`
   - Broker running

## Setup Steps

### 1. Start Infrastructure

```bash
# Start MySQL, Seata Server, and RocketMQ using docker-compose
cd ../../dockercompose
docker-compose up -d
```

### 2. Initialize Database

```bash
# Create fence table
mysql -h127.0.0.1 -uroot -proot seata < script/mysql.sql
```

### 3. Run the Sample

**Commit Scenario** (message will be sent and committed):
```bash
cd cmd
go run main.go --mode=commit
```

**Rollback Scenario** (message will be sent but rolled back):
```bash
cd cmd
go run main.go --mode=rollback
```

## Expected Behavior

### Commit Mode
1. Application starts global transaction
2. Prepare phase: Sends RocketMQ half-message (not consumable yet)
3. Fence records status=1 (tried)
4. Business logic succeeds
5. Commit phase: Message becomes consumable
6. Fence updates status=2 (committed)
7. Consumers can now receive the message

### Rollback Mode
1. Application starts global transaction
2. Prepare phase: Sends RocketMQ half-message
3. Fence records status=1 (tried)
4. Business logic returns error (simulated failure)
5. Rollback phase: Message is deleted/canceled
6. Fence updates status=3 (rollbacked)
7. Message never becomes consumable

## Verification

Check fence log table:
```sql
SELECT * FROM tcc_fence_log ORDER BY gmt_create DESC LIMIT 10;
```

Check RocketMQ console or CLI tools to verify message visibility.

## Troubleshooting

**Issue**: `database connect failed`
- Ensure MySQL is running and accessible
- Verify credentials in `service/producer_service.go`

**Issue**: `seata server not available`
- Check Seata TC server is running on `127.0.0.1:8091`
- Verify `conf/seatago.yml` configuration

**Issue**: `RocketMQ connection failed`
- Ensure RocketMQ NameServer and Broker are running
- Check network connectivity to `127.0.0.1:9876`

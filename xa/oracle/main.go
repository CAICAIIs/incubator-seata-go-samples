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
	"database/sql"
	"fmt"
	"time"

	"seata.apache.org/seata-go-samples/util"
	"seata.apache.org/seata-go/pkg/client"
)

var db *sql.DB

func main() {
	client.InitPath("../../conf/seatago.yml")
	db = util.GetXAOracleDb()
	ctx := context.Background()

	fmt.Println("=== Oracle XA Sample Started ===")

	// sample: insert
	fmt.Println("\n--- Test 1: Insert Data ---")
	sampleInsert(ctx)

	// sample: select for update
	fmt.Println("\n--- Test 2: Select For Update ---")
	sampleSelectForUpdate(ctx)

	// sample: update
	fmt.Println("\n--- Test 3: Update Data ---")
	_ = updateData(ctx)

	// sample: delete
	fmt.Println("\n--- Test 4: Delete Data ---")
	_ = deleteData(ctx)

	fmt.Println("\n=== Oracle XA Sample Completed ===")
	fmt.Println("All tests passed! Oracle XA is working correctly.")

	<-make(chan struct{})
}

func deleteData(ctx context.Context) error {
	sql := "DELETE FROM order_tbl WHERE id=:1"
	ret, err := db.ExecContext(ctx, sql, 2)
	if err != nil {
		fmt.Printf("delete failed, err:%v\n", err)
		return err
	}
	rows, err := ret.RowsAffected()
	if err != nil {
		fmt.Printf("delete failed, err:%v\n", err)
		return err
	}
	fmt.Printf("delete success: %d rows affected.\n", rows)
	return nil
}

func updateData(ctx context.Context) error {
	sql := "UPDATE order_tbl SET descs=:1 WHERE id=:2"
	ret, err := db.ExecContext(ctx, sql, fmt.Sprintf("NewDescs-%d", time.Now().UnixMilli()), 1)
	if err != nil {
		fmt.Printf("update failed, err:%v\n", err)
		return err
	}
	rows, err := ret.RowsAffected()
	if err != nil {
		fmt.Printf("update failed, err:%v\n", err)
		return err
	}
	fmt.Printf("update success: %d rows affected.\n", rows)
	return nil
}

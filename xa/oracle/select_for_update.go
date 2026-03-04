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

	"seata.apache.org/seata-go/pkg/tm"
)

func selectForUpdateData(ctx context.Context) error {
	sqlSelect := "SELECT user_id, commodity_code, count, money FROM order_tbl WHERE id=:1 FOR UPDATE"
	row := db.QueryRowContext(ctx, sqlSelect, 1)

	var (
		userId        string
		commodityCode string
		count         int
		money         sql.NullInt64
	)

	err := row.Scan(&userId, &commodityCode, &count, &money)
	if err != nil {
		fmt.Printf("select for update failed, err:%v\n", err)
		return err
	}

	fmt.Printf("select for update success: userId=%s, commodityCode=%s, count=%d, money=%v\n",
		userId, commodityCode, count, money)

	// Update the selected row
	sqlUpdate := "UPDATE order_tbl SET count=:1 WHERE id=:2"
	ret, err := db.ExecContext(ctx, sqlUpdate, count+10, 1)
	if err != nil {
		fmt.Printf("update after select failed, err:%v\n", err)
		return err
	}

	rows, err := ret.RowsAffected()
	if err != nil {
		fmt.Printf("update after select failed, err:%v\n", err)
		return err
	}

	fmt.Printf("update after select success: %d rows affected.\n", rows)
	return nil
}

func sampleSelectForUpdate(ctx context.Context) {
	_ = tm.WithGlobalTx(ctx, &tm.GtxConfig{
		Name:    "XASampleOracleGlobalTx_SelectForUpdate",
		Timeout: time.Second * 30,
	}, selectForUpdateData)
}

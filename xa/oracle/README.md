# Oracle XA Sample

This sample demonstrates how to use Seata-Go with Oracle database in XA mode.

## Prerequisites

1. **Oracle Database** - You need a running Oracle database instance
2. **Seata Server** - Running seata-server (can use docker-compose from root)
3. **Go 1.18+** - For running the sample

## Oracle Database Setup

### Option 1: Using Docker (Recommended for testing)

```bash
# Pull Oracle XE image (free version)
docker pull container-registry.oracle.com/database/express:latest

# Run Oracle XE
docker run -d \
  --name oracle-xe \
  -p 1521:1521 \
  -e ORACLE_PWD=oracle \
  container-registry.oracle.com/database/express:latest
```

### Option 2: Using existing Oracle instance

Make sure you have access to an Oracle database with XA support enabled.

## Database Schema Setup

Connect to your Oracle database and create the required table:

```sql
-- Create table
CREATE TABLE order_tbl (
    id NUMBER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id VARCHAR2(255),
    commodity_code VARCHAR2(255),
    count NUMBER,
    money NUMBER,
    descs VARCHAR2(255)
);

-- Grant XA privileges (required for XA transactions)
GRANT SELECT ON sys.dba_pending_transactions TO your_username;
GRANT SELECT ON sys.pending_trans$ TO your_username;
GRANT SELECT ON sys.dba_2pc_pending TO your_username;
GRANT EXECUTE ON sys.dbms_xa TO your_username;
```

## Environment Variables

Configure Oracle connection using environment variables:

```bash
export ORACLE_HOST=127.0.0.1
export ORACLE_PORT=1521
export ORACLE_USERNAME=system
export ORACLE_PASSWORD=oracle
export ORACLE_SERVICE=XE
```

Or use the defaults (shown above).

## Running the Sample

1. **Start Seata Server** (from repository root):
   ```bash
   cd ../../dockercompose
   docker-compose -f docker-compose.yml up -d seata-server
   ```

2. **Run the Oracle XA sample**:
   ```bash
   cd xa/oracle
   go run .
   ```

## What This Sample Tests

The sample demonstrates the following Oracle XA operations:

1. **Insert** - Insert data within a global transaction
2. **Select For Update** - Select with row locking and update
3. **Update** - Update existing records
4. **Delete** - Delete records

All operations are executed within Seata global transactions, demonstrating:
- XA START - Begin XA transaction branch
- XA END - End XA transaction branch
- XA PREPARE - Prepare for commit
- XA COMMIT - Commit the transaction
- XA ROLLBACK - Rollback on failure

## Expected Output

```
=== Oracle XA Sample Started ===

--- Test 1: Insert Data ---
insert success: 1 rows affected.

--- Test 2: Select For Update ---
select for update success: userId=NO-100001, commodityCode=C100000, count=100, money=<nil>
update after select success: 1 rows affected.

--- Test 3: Update Data ---
update success: 1 rows affected.

--- Test 4: Delete Data ---
delete success: 1 rows affected.

=== Oracle XA Sample Completed ===
All tests passed! Oracle XA is working correctly.
```

## Troubleshooting

### Connection Issues

If you see connection errors:
1. Verify Oracle is running: `docker ps` or check your Oracle service
2. Test connection: `sqlplus system/oracle@localhost:1521/XE`
3. Check firewall settings for port 1521

### XA Permission Issues

If you see XA-related errors:
```sql
-- Grant necessary XA permissions
GRANT SELECT ON sys.dba_pending_transactions TO system;
GRANT EXECUTE ON sys.dbms_xa TO system;
```

### Table Not Found

Make sure you created the `order_tbl` table in the correct schema.

## Differences from MySQL XA Sample

Oracle XA uses slightly different SQL syntax:

1. **Parameter Placeholders**: Oracle uses `:1, :2, :3` instead of `?`
2. **Auto-increment**: Oracle uses `GENERATED ALWAYS AS IDENTITY` instead of `AUTO_INCREMENT`
3. **Data Types**: `VARCHAR2` instead of `VARCHAR`, `NUMBER` instead of `INT`

## Integration with Seata-Go

This sample uses the newly implemented Oracle XA driver in Seata-Go v2.0.0+:

- Driver: `sql2.SeataXAOracleDriver`
- Supports all XA operations: Start, End, Prepare, Commit, Rollback, Recover, Forget
- Full compatibility with go-ora driver (github.com/sijms/go-ora/v2)

## Notes

- Oracle XA requires proper database privileges
- Make sure your Oracle instance has XA support enabled
- For production use, configure proper connection pooling and timeouts
- This sample uses Oracle Express Edition (XE) which is free for development

-- Oracle XA Sample Database Schema
-- Licensed to the Apache Software Foundation (ASF) under one or more
-- contributor license agreements.

-- Create the order table
CREATE TABLE order_tbl (
    id NUMBER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id VARCHAR2(255),
    commodity_code VARCHAR2(255),
    count NUMBER,
    money NUMBER,
    descs VARCHAR2(255)
);

-- Grant XA privileges (required for XA transactions)
-- Replace 'system' with your actual username if different
GRANT SELECT ON sys.dba_pending_transactions TO system;
GRANT SELECT ON sys.pending_trans$ TO system;
GRANT SELECT ON sys.dba_2pc_pending TO system;
GRANT EXECUTE ON sys.dbms_xa TO system;

-- Create index for better performance
CREATE INDEX idx_order_user ON order_tbl(user_id);

-- Insert sample data (optional)
INSERT INTO order_tbl (user_id, commodity_code, count, money, descs) 
VALUES ('NO-100000', 'C100000', 50, 1000, 'sample order');

COMMIT;

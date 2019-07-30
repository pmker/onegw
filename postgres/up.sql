


-- ----------------------------
-- Table structure for launch_logs
-- ----------------------------
DROP TABLE IF EXISTS "launch_logs";
CREATE TABLE "launch_logs" (
 "id" SERIAL PRIMARY KEY,
  "item_type" text COLLATE "pg_catalog"."default" NOT NULL,
  "item_id" int4 NOT NULL,
  "status" text COLLATE "pg_catalog"."default" NOT NULL,
  "transaction_hash" text COLLATE "pg_catalog"."default",
  "block_number" int4,
  "t_from" text COLLATE "pg_catalog"."default" NOT NULL,
  "t_to" text COLLATE "pg_catalog"."default" NOT NULL,
  "value" numeric(32,18),
  "gas_limit" int4,
  "gas_used" int4,
  "gas_price" numeric(32,18),
  "nonce" int4,
  "data" text COLLATE "pg_catalog"."default" NOT NULL,
  "executed_at" timestamp(6),
  "updated_at" timestamp(6),
  "created_at" timestamp(6),
  "amount" varchar(255) COLLATE "pg_catalog"."default",
  "token_address" varchar(255) COLLATE "pg_catalog"."default",
  "vite_hash" varchar(255) COLLATE "pg_catalog"."default",
  "vite_wallet_address" varchar(255) COLLATE "pg_catalog"."default",
  "vite_token_id" varchar(255) COLLATE "pg_catalog"."default"
)
;
ALTER TABLE "launch_logs" OWNER TO "pmker";

-- ----------------------------
-- Table structure for orders
-- ----------------------------
DROP TABLE IF EXISTS "orders";
CREATE TABLE "orders" (
  "id" SERIAL PRIMARY KEY,
  "walletAddress" varchar(255) COLLATE "pg_catalog"."default" NOT NULL,
  "tokenId" varchar(250) COLLATE "pg_catalog"."default" NOT NULL,
  "inTxHash" varchar(250) COLLATE "pg_catalog"."default",
  "outTxHash" varchar(250) COLLATE "pg_catalog"."default",
  "amount" varchar(250) COLLATE "pg_catalog"."default",
  "fee" varchar(250) COLLATE "pg_catalog"."default",
  "state" varchar(250) COLLATE "pg_catalog"."default",
  "dateTime" varchar(250) COLLATE "pg_catalog"."default",
  "type" int4 DEFAULT 0,
  "createdAt" timestamptz(6),
  "updatedAt" timestamptz(6)
)
;
ALTER TABLE "orders" OWNER TO "pmker";

-- ----------------------------
-- Table structure for tokens
-- ----------------------------
DROP TABLE IF EXISTS "tokens";
CREATE TABLE "tokens" (
  "id" SERIAL PRIMARY KEY,
  "tokenId" varchar(255) COLLATE "pg_catalog"."default",
  "tokenSymbol" varchar(32) COLLATE "pg_catalog"."default" NOT NULL,
  "type" int4 DEFAULT 0,
  "depositState" varchar(10) COLLATE "pg_catalog"."default" DEFAULT 'OPEN'::character varying,
  "withdrawState" varchar(10) COLLATE "pg_catalog"."default" DEFAULT 'OPEN'::character varying,
  "depositAddress" varchar(50) COLLATE "pg_catalog"."default",
  "labelName" varchar(32) COLLATE "pg_catalog"."default",
  "label" varchar(32) COLLATE "pg_catalog"."default",
  "minimumDepositAmount" int4 DEFAULT 1,
  "confirmationCount" int4 DEFAULT 20,
  "noticeMsg" varchar(255) COLLATE "pg_catalog"."default",
  "minimumWithdrawAmount" varchar(255) COLLATE "pg_catalog"."default",
  "maximumWithdrawAmount" varchar(255) COLLATE "pg_catalog"."default",
  "gatewayAddress" varchar(255) COLLATE "pg_catalog"."default",
  "fee" int4 DEFAULT 2,
  "createdAt" timestamptz(6),
  "updatedAt" timestamptz(6)
)
;
ALTER TABLE "tokens" OWNER TO "pmker";

-- ----------------------------
-- Table structure for transactions
-- ----------------------------
DROP TABLE IF EXISTS "transactions";
CREATE TABLE "transactions" (
  "id" SERIAL PRIMARY KEY,
  "transaction_hash" text COLLATE "pg_catalog"."default",
  "market_id" text COLLATE "pg_catalog"."default" NOT NULL,
  "status" text COLLATE "pg_catalog"."default" NOT NULL,
  "executed_at" timestamp(6),
  "updated_at" timestamp(6),
  "created_at" timestamp(6),
  "action" varchar(255) COLLATE "pg_catalog"."default",
  "blocknumber" varchar(255) COLLATE "pg_catalog"."default"
)
;
ALTER TABLE "transactions" OWNER TO "pmker";

-- ----------------------------
-- Table structure for wallets
-- ----------------------------
DROP TABLE IF EXISTS "wallets";
CREATE TABLE "wallets" (
 "id" SERIAL PRIMARY KEY,
  "eth_address" varchar(250) COLLATE "pg_catalog"."default" NOT NULL,
  "vite_address" varchar(255) COLLATE "pg_catalog"."default",
  "createdAt" timestamptz(6),
  "updatedAt" timestamptz(6)
)
;
ALTER TABLE "wallets" OWNER TO "pmker";

-- ----------------------------
-- Indexes structure for table launch_logs
-- ----------------------------
CREATE INDEX "idx_created_at" ON "launch_logs" USING btree (
  "created_at" "pg_catalog"."timestamp_ops" ASC NULLS LAST
);
CREATE INDEX "idx_launch_logs_nonce" ON "launch_logs" USING btree (
  "nonce" "pg_catalog"."int4_ops" ASC NULLS LAST
);
CREATE UNIQUE INDEX "idx_launch_logs_transaction_hash" ON "launch_logs" USING btree (
  "transaction_hash" COLLATE "pg_catalog"."default" "pg_catalog"."text_ops" ASC NULLS LAST
);

-- ----------------------------
-- Primary Key structure for table launch_logs
-- ----------------------------
ALTER TABLE "launch_logs" ADD CONSTRAINT "launch_logs_pkey" PRIMARY KEY ("id");

-- ----------------------------
-- Primary Key structure for table orders
-- ----------------------------
ALTER TABLE "orders" ADD CONSTRAINT "orders_pkey" PRIMARY KEY ("id");

-- ----------------------------
-- Primary Key structure for table tokens
-- ----------------------------
ALTER TABLE "tokens" ADD CONSTRAINT "vite_tokens_pkey" PRIMARY KEY ("id");

-- ----------------------------
-- Indexes structure for table transactions
-- ----------------------------
CREATE UNIQUE INDEX "idx_transactions_transaction_hash" ON "transactions" USING btree (
  "transaction_hash" COLLATE "pg_catalog"."default" "pg_catalog"."text_ops" ASC NULLS LAST
);

-- ----------------------------
-- Primary Key structure for table transactions
-- ----------------------------
ALTER TABLE "transactions" ADD CONSTRAINT "transactions_pkey" PRIMARY KEY ("id");

-- ----------------------------
-- Uniques structure for table wallets
-- ----------------------------
ALTER TABLE "wallets" ADD CONSTRAINT "eth_address" UNIQUE ("eth_address");

-- ----------------------------
-- Primary Key structure for table wallets
-- ----------------------------
ALTER TABLE "wallets" ADD CONSTRAINT "wallets_pkey" PRIMARY KEY ("id");
-- Updates for 0.10.0
ALTER TABLE Clients ADD proto BYTEA;

CREATE DATABASE mainnet;
CREATE DATABASE rinkeby;
CREATE DATABASE dev;

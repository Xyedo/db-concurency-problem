CREATE EXTENSION IF NOT EXISTS citext;

CREATE TABLE IF NOT EXISTS ACCOUNT (
  id TEXT PRIMARY KEY,
  username CITEXT UNIQUE,
  phone_number CITEXT UNIQUE,
  email CITEXT UNIQUE,
  hashed_password TEXT NOT NULL,
  is_deleted BOOLEAN NOT NULL DEFAULT FALSE,
  created_on TIMESTAMPTZ NOT NULL,
  updated_on TIMESTAMPTZ,
  version BIGINT NOT NULL DEFAULT 1
);

CREATE TABLE IF NOT EXISTS THREAD (
  id TEXT PRIMARY KEY,
  title TEXT NOT NULL,
  body TEXT NOT NULL,
  total_comment BIGINT NOT NULL DEFAULT 0,
  total_reaction BIGINT NOT NULL DEFAULT 0,
  created_by TEXT NOT NULL REFERENCES ACCOUNT ON DELETE CASCADE,
  created_on TIMESTAMPTZ NOT NULL,
  updated_by TEXT REFERENCES ACCOUNT ON DELETE CASCADE,
  updated_on TIMESTAMPTZ,
  is_deleted BOOLEAN NOT NULL DEFAULT FALSE,
  version BIGINT NOT NULL DEFAULT 1
);

CREATE TABLE IF NOT EXISTS COMMENT (
  id TEXT PRIMARY KEY,
  thread_id TEXT NOT NULL REFERENCES THREAD ON DELETE CASCADE,
  user_id TEXT NOT NULL REFERENCES ACCOUNT ON DELETE CASCADE,
  reply_to TEXT REFERENCES COMMENT ON DELETE CASCADE,
  content TEXT NOT NULL,
  total_reply BIGINT NOT NULL DEFAULT 0,
  total_reaction BIGINT NOT NULL DEFAULT 0,
  created_on TIMESTAMPTZ NOT NULL,
  updated_on TIMESTAMPTZ,
  is_deleted BOOLEAN NOT NULL DEFAULT FALSE,
  version BIGINT NOT NULL DEFAULT 1
);

CREATE TABLE IF NOT EXISTS REACTION (
  id TEXT PRIMARY KEY,
  account_id TEXT NOT NULL REFERENCES ACCOUNT ON DELETE CASCADE,
  thread_id TEXT REFERENCES THREAD ON DELETE CASCADE,
  comment_id TEXT REFERENCES COMMENT ON DELETE CASCADE,
  content VARCHAR(100) NOT NULL,
  created_on TIMESTAMPTZ NOT NULL,
  updated_on TIMESTAMPTZ,
  version BIGINT NOT NULL DEFAULT 1
);


--------------------------------------------------TABLE FOR TESTING WRITE SKEW----------------------
 CREATE TABLE IF NOT EXISTS FAKE_TABLE (
  id TEXT PRIMARY KEY,
  "number" TEXT NOT NULL,
  created_on TIMESTAMPTZ NOT NULL,
  updated_on TIMESTAMPTZ,
  version BIGINT NOT NULL DEFAULT 1
);
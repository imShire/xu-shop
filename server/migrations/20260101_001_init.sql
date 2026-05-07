-- +goose Up
-- +goose StatementBegin

CREATE TABLE "user" (
  id              bigint PRIMARY KEY,
  openid_mp       varchar(64) UNIQUE,
  openid_h5       varchar(64) UNIQUE,
  unionid         varchar(64) UNIQUE,
  phone           varchar(20),
  phone_country   varchar(8) DEFAULT '86',
  nickname        varchar(64),
  avatar          varchar(512),
  gender          smallint DEFAULT 0,
  birthday        date,
  status          varchar(16) NOT NULL DEFAULT 'active',
  deactivate_at   timestamptz,
  invited_by_user_id bigint,
  distributor_id  bigint,
  points          int NOT NULL DEFAULT 0,
  created_at      timestamptz NOT NULL DEFAULT now(),
  updated_at      timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX idx_user_phone ON "user"(phone);
CREATE INDEX idx_user_unionid ON "user"(unionid);
CREATE UNIQUE INDEX uq_user_phone_active ON "user"(phone)
  WHERE phone IS NOT NULL AND status <> 'deactivated';

CREATE TABLE admin (
  id               bigint PRIMARY KEY,
  username         varchar(64) NOT NULL UNIQUE,
  password_hash    varchar(255) NOT NULL,
  real_name        varchar(64),
  phone            varchar(20),
  status           varchar(16) NOT NULL DEFAULT 'active',
  failed_attempts  int NOT NULL DEFAULT 0,
  locked_until     timestamptz,
  last_login_at    timestamptz,
  last_login_ip    inet,
  created_at       timestamptz NOT NULL DEFAULT now(),
  updated_at       timestamptz NOT NULL DEFAULT now(),
  deleted_at       timestamptz
);

CREATE TABLE role (
  id          bigint PRIMARY KEY,
  code        varchar(32) NOT NULL UNIQUE,
  name        varchar(64) NOT NULL,
  is_system   bool NOT NULL DEFAULT false,
  created_at  timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE permission (
  code        varchar(64) PRIMARY KEY,
  module      varchar(32) NOT NULL,
  action      varchar(32) NOT NULL,
  name        varchar(64) NOT NULL
);

CREATE TABLE admin_role (
  admin_id bigint NOT NULL REFERENCES admin(id),
  role_id  bigint NOT NULL REFERENCES role(id),
  PRIMARY KEY (admin_id, role_id)
);

CREATE TABLE role_permission (
  role_id          bigint NOT NULL REFERENCES role(id),
  permission_code  varchar(64) NOT NULL REFERENCES permission(code),
  PRIMARY KEY (role_id, permission_code)
);

CREATE TABLE login_log (
  id            bigint PRIMARY KEY,
  subject_type  varchar(8) NOT NULL,
  subject_id    bigint,
  ip            inet,
  ua            text,
  success       bool NOT NULL,
  fail_reason   varchar(128),
  created_at    timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX idx_loginlog_subject ON login_log(subject_type, subject_id, created_at DESC);

CREATE TABLE region (
  code        varchar(12) PRIMARY KEY,
  parent_code varchar(12),
  name        varchar(32) NOT NULL,
  level       smallint NOT NULL,
  sort        int NOT NULL DEFAULT 0
);
CREATE INDEX idx_region_parent ON region(parent_code, sort);

CREATE TABLE address (
  id              bigint PRIMARY KEY,
  user_id         bigint NOT NULL,
  name            varchar(32) NOT NULL,
  phone           varchar(20) NOT NULL,
  province_code   varchar(12),
  province        varchar(32),
  city_code       varchar(12),
  city            varchar(32),
  district_code   varchar(12),
  district        varchar(32),
  detail          varchar(200) NOT NULL,
  is_default      bool NOT NULL DEFAULT false,
  created_at      timestamptz NOT NULL DEFAULT now(),
  updated_at      timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX idx_address_user ON address(user_id, is_default DESC, updated_at DESC);
CREATE UNIQUE INDEX uq_address_default_per_user ON address(user_id) WHERE is_default = true;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS uq_address_default_per_user;
DROP INDEX IF EXISTS idx_address_user;
DROP TABLE IF EXISTS address;
DROP INDEX IF EXISTS idx_region_parent;
DROP TABLE IF EXISTS region;
DROP INDEX IF EXISTS idx_loginlog_subject;
DROP TABLE IF EXISTS login_log;
DROP TABLE IF EXISTS role_permission;
DROP TABLE IF EXISTS admin_role;
DROP TABLE IF EXISTS permission;
DROP TABLE IF EXISTS role;
DROP TABLE IF EXISTS admin;
DROP INDEX IF EXISTS uq_user_phone_active;
DROP INDEX IF EXISTS idx_user_unionid;
DROP INDEX IF EXISTS idx_user_phone;
DROP TABLE IF EXISTS "user";
-- +goose StatementEnd

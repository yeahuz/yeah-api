BEGIN;

CREATE TABLE IF NOT EXISTS languages (
  code varchar(2) PRIMARY KEY
);

CREATE TABLE IF NOT EXISTS users (
  id uuid PRIMARY KEY,
  phone varchar(15) UNIQUE,
  phone_verified boolean DEFAULT FALSE,
  last_name varchar(255) DEFAULT '',
  first_name varchar(255) DEFAULT '',
  username varchar(255) UNIQUE,
  bio varchar(1024) DEFAULT '',
  website_url varchar(255) DEFAULT '',
  photo_url varchar(255) DEFAULT '',
  email varchar(255) UNIQUE,
  email_verified boolean DEFAULT FALSE,
  password varchar(255) DEFAULT '',
  profile_url varchar(255) DEFAULT '',
  verified boolean DEFAULT FALSE,
  created_at timestamp with time zone DEFAULT now() NOT NULL,
  updated_at timestamp with time zone
);

CREATE TABLE IF NOT EXISTS auth_providers (
  name varchar(255) PRIMARY KEY,
  logo_url varchar(255) DEFAULT '',
  active boolean DEFAULT FALSE,
  created_at timestamp with time zone DEFAULT now() NOT NULL,
  updated_at timestamp with time zone
);

CREATE TABLE IF NOT EXISTS accounts (
  id uuid PRIMARY KEY,
  provider varchar(255) NOT NULL,
  user_id uuid NOT NULL,
  provider_account_id varchar(255) NOT NULL,
  created_at timestamp with time zone DEFAULT now() NOT NULL,
  updated_at timestamp with time zone,
  FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,
  FOREIGN KEY (provider) REFERENCES auth_providers (name) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS otps (
  id uuid PRIMARY KEY,
  code varchar(255) NOT NULL,
  hash varchar(255) NOT NULL,
  expires_at timestamp with time zone,
  confirmed boolean DEFAULT FALSE,
  created_at timestamp with time zone DEFAULT now() NOT NULL,
  updated_at timestamp with time zone
);

CREATE TABLE IF NOT EXISTS credential_requests (
  id uuid PRIMARY KEY,
  type varchar(255) NOT NULL,
  challenge varchar(255) NOT NULL,
  user_id uuid NOT NULL,
  used boolean DEFAULT FALSE,
  FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS credentials (
  id uuid PRIMARY KEY,
  credential_id varchar(1024) DEFAULT '',
  title varchar(255) NOT NULL,
  pubkey text DEFAULT '',
  pubkey_alg int DEFAULT -7,
  type varchar(255) DEFAULT 'public-key',
  counter int DEFAULT 0,
  transports text[] DEFAULT '{}',
  user_id uuid NOT NULL,
  credential_request_id uuid NOT NULL,
  created_at timestamp with time zone DEFAULT now() NOT NULL,
  updated_at timestamp with time zone,
  FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,
  FOREIGN KEY (credential_request_id) REFERENCES credential_requests (id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS clients (
  id uuid PRIMARY KEY,
  name varchar(255) NOT NULL,
  secret varchar(255) DEFAULT '',
  type varchar(255) NOT NULL CHECK (type IN ('confidential', 'public', 'internal')) DEFAULT 'confidential',
  active boolean DEFAULT TRUE,
  created_at timestamp with time zone DEFAULT now() NOT NULL,
  updated_at timestamp with time zone
);

CREATE TABLE IF NOT EXISTS sessions (
  id uuid DEFAULT gen_random_uuid () PRIMARY KEY,
  active boolean DEFAULT TRUE,
  ip inet,
  user_id uuid NOT NULL,
  client_id uuid NOT NULL,
  user_agent varchar(255),
  created_at timestamp with time zone DEFAULT now() NOT NULL,
  updated_at timestamp with time zone,
  FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,
  FOREIGN KEY (client_id) REFERENCES clients (id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS categories (
  id serial PRIMARY KEY,
  parent_id int,
  created_at timestamp with time zone DEFAULT now() NOT NULL,
  updated_at timestamp with time zone,
  FOREIGN KEY (parent_id) REFERENCES categories (id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS listing_statuses (
  code varchar(255) CHECK (code IN ('ACTIVE', 'MODERATION', 'INDEXING', 'ARCHIVED', 'DRAFT', 'DELETED')) PRIMARY KEY,
  active boolean DEFAULT TRUE,
  fg_hex varchar(7) DEFAULT '',
  bg_hex varchar(7) DEFAULT ''
);

CREATE TABLE IF NOT EXISTS listing_statuses_tr (
  status_code varchar(255) NOT NULL,
  lang_code varchar(255) NOT NULL,
  name varchar(255) DEFAULT '',
  FOREIGN KEY (status_code) REFERENCES listing_statuses (code) ON DELETE CASCADE,
  FOREIGN KEY (lang_code) REFERENCES languages (code) ON DELETE CASCADE,
  PRIMARY KEY (status_code, lang_code)
);

CREATE TABLE IF NOT EXISTS listings (
  id uuid PRIMARY KEY,
  title varchar(255) NOT NULL,
  category_id int NOT NULL,
  owner_id uuid NOT NULL,
  created_at timestamp with time zone DEFAULT now() NOT NULL,
  updated_at timestamp with time zone,
  status varchar(255) NOT NULL,
  FOREIGN KEY (owner_id) REFERENCES users (id) ON DELETE CASCADE,
  FOREIGN KEY (category_id) REFERENCES categories (id) ON DELETE SET NULL,
  FOREIGN KEY (status) REFERENCES listing_statuses (code) ON DELETE SET NULL
);

CREATE TABLE IF NOT EXISTS currencies (
  code varchar(10) PRIMARY KEY,
  symbol varchar(10) DEFAULT ''
);

CREATE TABLE IF NOT EXISTS listing_skus (
  id uuid PRIMARY KEY,
  custom_sku varchar(255) DEFAULT '',
  listing_id uuid NOT NULL,
  created_at timestamp with time zone DEFAULT now() NOT NULL,
  updated_at timestamp with time zone,
  FOREIGN KEY (listing_id) REFERENCES listings (id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS listing_sku_prices (
  sku_id uuid NOT NULL,
  amount bigint DEFAULT 0,
  currency varchar(255) NOT NULL,
  start_date timestamp with time zone DEFAULT now() NOT NULL,
  created_at timestamp with time zone DEFAULT now() NOT NULL,
  updated_at timestamp with time zone,
  FOREIGN KEY (sku_id) REFERENCES listing_skus (id) ON DELETE CASCADE,
  FOREIGN KEY (currency) REFERENCES currencies (code) ON DELETE SET NULL,
  PRIMARY KEY (sku_id, start_date)
);

CREATE INDEX idx_categories_parent_id ON categories (parent_id);

CREATE INDEX idx_listings_owner_id ON listings (owner_id);

CREATE INDEX idx_listings_category_id ON listings (category_id);

CREATE INDEX idx_sessions_user_id ON sessions (user_id);

CREATE INDEX idx_sessions_client_id ON sessions (client_id);

CREATE INDEX idx_credential_requests_user_id ON credential_requests (user_id);

CREATE INDEX idx_credentials_user_id ON credentials (user_id);

CREATE INDEX idx_credentials_credential_id ON credentials (credential_id);

CREATE INDEX idx_credentials_credential_request_id ON credentials (credential_request_id);

CREATE INDEX idx_accounts_user_id ON accounts (user_id);

CREATE INDEX idx_accounts_provider_account_id ON accounts (provider_account_id);

CREATE INDEX idx_otps_hash ON otps (hash);

CREATE UNIQUE INDEX udx_accounts_provider_account_id_user_id ON accounts (provider_account_id, user_id);

CREATE OR REPLACE FUNCTION update_updated_at_column ()
  RETURNS TRIGGER
  AS $$
BEGIN
  NEW.updated_at = now();
  RETURN NEW;
END;
$$
LANGUAGE plpgsql;

DO $$
DECLARE
  t text;
BEGIN
  FOR t IN
  SELECT
    table_name
  FROM
    information_schema.columns
  WHERE
    column_name = 'updated_at' LOOP
      EXECUTE format('CREATE TRIGGER trigger_update_timestamp
                    BEFORE UPDATE ON %I
                    FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column()', t, t);
    END LOOP;
END;
$$
LANGUAGE 'plpgsql';

CREATE OR REPLACE FUNCTION nullify_email_phone ()
  RETURNS TRIGGER
  AS $$
BEGIN
  NEW.email = nullif (NEW.email, '');
  NEW.phone = nullif (NEW.phone, '');
  RETURN NEW;
END;
$$
LANGUAGE plpgsql;

CREATE TRIGGER trigger_replace_email_phone_empty_string
  BEFORE INSERT OR UPDATE ON users
  FOR EACH ROW
  EXECUTE PROCEDURE nullify_email_phone ();

COMMIT;
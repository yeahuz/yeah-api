begin;

create table if not exists languages (
  code varchar(2) primary key
);

create table if not exists users (
  id uuid primary key,
  phone varchar(15) unique,
  phone_verified boolean default false,
  last_name varchar(255) default '',
  first_name varchar(255) default '',
  username varchar(255) unique,
  bio varchar(1024) default '',
  website_url varchar(255) default '',
  photo_url varchar(255) default '',
  email varchar(255) unique,
  email_verified boolean default false,
  password varchar(255) default '',
  profile_url varchar(255) default '',
  verified boolean default false,
  created_at timestamp with time zone default now() not null,
  updated_at timestamp with time zone
);

create table if not exists auth_providers (
  name varchar(255) primary key,
  logo_url varchar(255) default '',
  active boolean default false,
  created_at timestamp with time zone default now() not null,
  updated_at timestamp with time zone
);

create table if not exists accounts (
  id uuid primary key,
  provider varchar(255) not null,
  user_id uuid not null,
  provider_account_id varchar(255) not null,
  created_at timestamp with time zone default now() not null,
  updated_at timestamp with time zone,

  foreign key (user_id) references users (id) on delete cascade,
  foreign key (provider) references auth_providers (name) on delete cascade
);

create table if not exists otps (
  id uuid primary key,
  code varchar(255) not null,
  hash varchar(255) not null,
  expires_at timestamp with time zone,
  confirmed boolean default false,
  created_at timestamp with time zone default now() not null,
  updated_at timestamp with time zone
);

create table if not exists credential_requests (
  id uuid primary key,
  type varchar(255) not null,
  challenge varchar(255) not null,
  user_id uuid not null,
  used boolean default false,

  foreign key(user_id) references users(id) on delete cascade
);

create table if not exists credentials (
  id uuid primary key,
  credential_id varchar(1024) default '',
  title varchar(255) not null,
  pubkey text default '',
  pubkey_alg int default -7,
  type varchar(255) default 'public-key',
  counter int default 0,
  transports text[] default '{}',
  user_id uuid not null,
  credential_request_id uuid not null,
  created_at timestamp with time zone default now() not null,
  updated_at timestamp with time zone,

  foreign key(user_id) references users(id) on delete cascade,
  foreign key(credential_request_id) references credential_requests(id) on delete cascade
);

create table if not exists clients (
  id uuid primary key,
  name varchar(255) not null,
  secret varchar(255) default '',
  type varchar(255) not null check (type in ('confidential', 'public', 'internal')) default 'confidential',
  active boolean default true,

  created_at timestamp with time zone default now() not null,
  updated_at timestamp with time zone
);

create table if not exists sessions (
  id uuid DEFAULT gen_random_uuid() primary key,
  active boolean default true,
  ip inet,
  user_id uuid not null,
  client_id uuid not null,
  user_agent varchar(255),
  created_at timestamp with time zone default now() not null,
  updated_at timestamp with time zone,

  foreign key(user_id) references users(id) on delete cascade,
  foreign key(client_id) references clients(id) on delete cascade
);

create table if not exists categories (
  id serial primary key,
  parent_id int,
  created_at timestamp with time zone default now() not null,
  updated_at timestamp with time zone,

  foreign key (parent_id) references categories(id) on delete cascade
);

create table if not exists listing_statuses (
 code varchar(255) check (code in ('ACTIVE', 'MODERATION', 'INDEXING', 'ARCHIVED', 'DRAFT', 'DELETED')) primary key,
 active boolean default true,
 fg_hex varchar(7) default '',
 bg_hex varchar(7) default ''
);

create table if not exists listing_statuses_tr (
  status_code varchar(255) not null,
  lang_code varchar(255) not null,
  name varchar(255) default '',

  foreign key (status_code) references listing_statuses(code) on delete cascade,
  foreign key (lang_code) references languages(code) on delete cascade,
  primary key (status_code, lang_code)
);

create table if not exists listings (
  id uuid primary key,
  title varchar(255) not null,
  category_id int not null,
  owner_id uuid not null,
  created_at timestamp with time zone default now() not null,
  updated_at timestamp with time zone,
  status varchar(255) not null,

  foreign key (owner_id) references users(id) on delete cascade,
  foreign key (category_id) references categories(id) on delete set null,
  foreign key (status) references listing_statuses(code) on delete set null
);

create table if not exists currencies (
  code varchar(10) primary key,
  symbol varchar(10) default ''
);

create table if not exists listing_skus (
  id uuid primary key,
  custom_sku varchar(255) default '',
  listing_id uuid not null,
  created_at timestamp with time zone default now() not null,
  updated_at timestamp with time zone,

  foreign key (listing_id) references listings(id) on delete cascade
);

create table if not exists listing_product_prices (
  product_id uuid not null,
  amount bigint 0,
  currency varchar(255) not null,
  start_date timestamp with time zone default now() not null,
  created_at timestamp with time zone default now() not null,
  updated_at timestamp with time zone,

  foreign key (product_id) references listing_products(id) on delete cascade,
  foreign key (currency) references currencies(code) on delete set null,
  primary key (product_id, start_date)
);

create index idx_categories_parent_id on categories(parent_id);

create index idx_listings_owner_id on listings(owner_id);
create index idx_listings_category_id on listings(category_id);

create index idx_sessions_user_id on sessions(user_id);
create index idx_sessions_client_id on sessions(client_id);

create index idx_credential_requests_user_id on credential_requests(user_id);

create index idx_credentials_user_id on credentials(user_id);
create index idx_credentials_credential_id on credentials(credential_id);
create index idx_credentials_credential_request_id on credentials(credential_request_id);

create index idx_accounts_user_id on accounts(user_id);
create index idx_accounts_provider_account_id on accounts(provider_account_id);

create index idx_otps_hash on otps(hash);
create unique index udx_accounts_provider_account_id_user_id on accounts(provider_account_id, user_id);

create or replace function update_updated_at_column()
returns trigger as $$
begin
NEW.updated_at = now();
return NEW;
end;
$$ language plpgsql;

DO $$
DECLARE
    t text;
BEGIN
    FOR t IN
        SELECT table_name FROM information_schema.columns WHERE column_name = 'updated_at'
    LOOP
        EXECUTE format('CREATE TRIGGER trigger_update_timestamp
                    BEFORE UPDATE ON %I
                    FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column()', t,t);
    END loop;
END;
$$ language 'plpgsql';

create or replace function nullify_email_phone()
returns trigger as $$
begin
NEW.email = nullif(NEW.email, '');
NEW.phone = nullif(NEW.phone, '');
return NEW;
end;
$$ language plpgsql;

create trigger trigger_replace_email_phone_empty_string
before insert or update on users
for each row execute procedure nullify_email_phone();

commit;

begin;

create table if not exists users (
  id bigserial primary key,
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
  id bigserial primary key,
  provider varchar(255) not null,
  user_id bigint not null,
  provider_account_id varchar(255) not null,
  created_at timestamp with time zone default now() not null,
  updated_at timestamp with time zone,

  foreign key (user_id) references users (id) on delete cascade,
  foreign key (provider) references auth_providers (name) on delete cascade
);

create table if not exists otps (
  id bigserial primary key,
  code varchar(255) not null,
  hash varchar(255) not null,
  expires_at timestamp with time zone,
  confirmed boolean default false,
  created_at timestamp with time zone default now() not null,
  updated_at timestamp with time zone
);

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

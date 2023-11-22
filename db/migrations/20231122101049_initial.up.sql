begin;

create table if not exists users (
  id bigserial primary key,
  phone varchar(15) unique,
  phone_verified boolean default false,
  name varchar(255)
  username varchar(255) unique,
  bio varchar(1024),
  website_url varchar(255),
  photo_url varchar(255),
  email varchar(255) unique,
  email_verified boolean default false,
  password varchar(255) not null,
  profile_url varchar(255),
  verified boolean default false,
  created_at timestamp with time zone default now() not null,
  updated_at timestamp with time zone on update now()
);

create table if not exists auth_providers (
  name varchar(255) primary key,
  logo_url varchar(255),
  active boolean default false,
  created_at timestamp with time zone default now() not null,
  updated_at timestamp with time zone on update now()
);

create table if not exists accounts (
  id bigserial primary key,
  provider varchar(255) not null,
  user_id bigint not null,
  provider_account_id varchar(255) not null,
  created_at timestamp with time zone default now() not null,
  updated_at timestamp with time zone on update now(),

  foreign key (user_id) references users (id) on delete cascade,
  foreign key (provider) references auth_providers (name) on delete cascade,
);

create index idx_accounts_user_id on accounts(user_id);
create index idx_accounts_provider_account_id on accounts(provider_account_id);
create unique index udx_accounts_provider_account_id_user_id on accounts(provider_account_id, user_id);

commit;

begin;

drop table if exists users cascade;
drop table if exists auth_providers cascade;
drop table if exists accounts cascade;
drop table if exists otps cascade;
drop table if exists credentials cascade;
drop table if exists credential_requests cascade;
drop table if exists sessions cascade;
drop table if exists clients cascade;

commit;

begin;

drop table if exists users cascade;
drop table if exists auth_providers cascade;
drop table if exists accounts cascade;
drop table if exists otps cascade;
drop table if exists credentials cascade;
drop table if exists credential_requests cascade;
drop table if exists sessions cascade;
drop table if exists clients cascade;
drop table if exists languages cascade;
drop table if exists categories cascade;
drop table if exists categories_tr cascade;
drop table if exists listing_statuses cascade;
drop table if exists listing_statuses_tr cascade;
drop table if exists listings cascade;
drop table if exists listing_skus cascade;
drop table if exists listing_sku_prices cascade;
drop table if exists listing_sku_prices cascade;
drop table if exists attributes cascade;
drop table if exists attribute_options cascade;
drop table if exists attribute_options_tr cascade;
drop table if exists attributes_tr cascade;
drop table if exists kv_store cascade;

commit;

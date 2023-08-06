create database idp;
create table idp.users
(
    id        varchar(255) not null,
    name      varchar(255) not null,
    picture   varchar(255) null,
    google_id varchar(255) null
);

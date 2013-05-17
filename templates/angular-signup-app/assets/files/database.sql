drop database if exists angular-signup-app;
create database angular-signup-app character set utf8 collate utf8_unicode_ci;
grant all privileges on angular-signup-app.* to 'angular-signup-app'@'localhost' identified by 'angular-signup-app';
use angular-signup-app;

create table User (
  Id integer not null auto_increment,
  Email varchar(50) unique not null,
  Passwd varchar(40) not null,
  FullName varchar(80) null,
  IsActive boolean not null,
  primary key(id)
);

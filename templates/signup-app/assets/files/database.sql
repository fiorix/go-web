drop database if exists dummy;
create database dummy character set utf8 collate utf8_unicode_ci;
grant all privileges on dummy.* to 'foo'@'localhost' identified by 'bar';
use dummy;

create table User (
  Id integer not null auto_increment,
  Email varchar(50) unique not null,
  Passwd varchar(40) not null,
  FullName varchar(80) null,
  IsActive boolean not null,
  primary key(Id)
);

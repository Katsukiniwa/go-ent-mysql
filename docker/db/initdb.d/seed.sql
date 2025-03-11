create database if not exists `product` character set utf8mb4 collate utf8mb4_bin;

use `product`;

drop table if exists users;
create table if not exists users
(
  id int unsigned not null primary key auto_increment,
  name varchar(128) not null
) character set utf8mb4 collate utf8mb4_bin;
insert into users (name)
values ('user1'),
       ('user2');

-- drop table if exists histories;
-- create table if not exists histories
-- (
--   id int unsigned not null primary key auto_increment,
--   user_histories int unsigned not null,
--   amount int not null
--   CONSTRAINT fk_histories_users FOREIGN KEY (user_histories) REFERENCES users (id)
-- ) character set utf8mb4 collate utf8mb4_bin;

-- insert into products (id, stock, title, sales_status)
-- values (1, 100, 'test 001', 0),
--        (2, 110, 'test 002', 1);
-- insert into histories (id, amount, user_histories)
-- values (1, 500, 1),
--        (2, 780, 2);

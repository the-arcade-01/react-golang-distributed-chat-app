create database if not exists go_chat_app;

use go_chat_app;

create table if not exists users (
    username varchar(50) NOT NULL UNIQUE PRIMARY KEY,
    password varchar(255) NOT NULL
);

create database if not exists go_chat_app;

use go_chat_app;

create table if not exists users (
    username varchar(50) NOT NULL UNIQUE PRIMARY KEY,
    password varchar(255) NOT NULL
);

create table if not exists rooms (
    room_id int NOT NULL AUTO_INCREMENT PRIMARY KEY,
    room_name varchar(255) NOT NULL,
    admin varchar(50) NOT NULL,
    FOREIGN KEY (admin) REFERENCES users(username)
);

create table if not exists user_rooms (
    username varchar(50) NOT NULL,
    room_id int NOT NULL,
    PRIMARY KEY (username, room_id),
    FOREIGN KEY (username) REFERENCES users(username) ON DELETE CASCADE,
    FOREIGN KEY (room_id) REFERENCES rooms(room_id) ON DELETE CASCADE
);
## go-chat-app

A simple distributed chat app example built using Golang, ReactJs, Redis and Websockets.

[![Watch the video](https://img.youtube.com/vi/n0669MY5Gvs/maxresdefault.jpg)](https://www.youtube.com/watch?v=n0669MY5Gvs)

### Installation

We will use docker for running the application. First, create `.env` file in server directory

<details>
<summary> server/.env </summary>

```jsx
DB_DRIVER=mysql
DB_URL=<user>:<password>@tcp(<mysql_docker_container_name>:3306)/<db_name>?parseTime=true
MYSQL_ROOT_PASSWORD=<password>
MYSQL_DATABASE=<db_name>
REDIS_ADDR=<redis_docker_container_name>:6379
REDIS_PWD=
JWT_SECRET_KEY=<secret_key>
```

</details>

For client, check in `client/docker-compose.yml` and change the `VITE_API_URL` as per your configuration.

For running the application, `cd` into client and server directory and run the below cmds in separate terminals

<details>
<summary>Commands</summary>

```jsx
~/client > docker compose up --build
~/server > docker compose up --build
```

</details>

### Tech

Build with:

- Golang
- ReactJs
- Redis
- MySQL
- Docker
- Websockets

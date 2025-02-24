## Distributed Chat App in Golang, React.js & Redis

Test the app here: https://chat.arcade.build

Read the blog for implementation details [Distributed Chat App in Golang, React.js & Redis](https://dev.to/the-arcade-01/google-oauth-20-flow-in-golang-and-reactjs-536a)

### Run this project

- For running the web
  1. Create `.env` file in the `web` folder
     ```shell
         VITE_API_URL=ws://localhost:8080/chat/ws
     ```
  2. Use npm to start the web
     ```shell
        ~> cd web
        ~> npm install
        ~> npm run dev
     ```
- For running the server
  - Create `.env` file and update the env variables from `.env.example` file
    ```shell
        ENV=development
        REDIS_ADDR=<redis_container_name>:6379
        REDIS_PWD=<redis_pwd>
        REDIS_DB=<redis_db>
        CHAT_CHANNEL=<channel>
        WEB_URL=http://localhost:5173
    ```
  - Choose which message system to run, redis pubsub or redis streams
    ```shell
        WS_TYPE=pubsub # for streams keep it empty
    ```
  - Use docker for running the server app, all the services are listed in `scripts/docker-compose.yml`
    ```shell
        ~> cd server/scripts
        ~> docker compose --env-file ../.env up
    ```

### Architecture

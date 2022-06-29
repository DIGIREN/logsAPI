# Usage:
- You will need docker and docker compose installed for this project.
- You will also need to supply env vars for this project in the form of a .env file. These are required, below is an example .env file with all required entries:
    ````
    GIN_MODE=debug
    PORT=80
    CHECK_FREQUENCY=5
    BATCH_SIZE=20
    BATCH_INTERVAL=600
    MAX_RETRIES=3
    RETRY_WAIT=2
    POST_ENDPOINT=https://test.mock.pstmn.io/logreceiver
    TRUSTED_PROXY=localhost
    ````
    Explanation of variables:
    - GIN_MODE: The mode of the gin server. Valid values are debug, release.
    - PORT: The port the gin server will listen on
    - CHECK_FREQUENCY: How often will the background listener check to see if the BATCH_INTERVAL has been met
    - BATCH_SIZE: The number of logs to send in a batch
    - BATCH_INTERVAL: The number of seconds to wait before sending a batch of logs
    - MAX_RETRIES: The number of times to retry a failed batch
    - RETRY_WAIT: The number of seconds to wait between retries
    - POST_ENDPOINT: The endpoint to send the logs to
    - TRUSTED_PROXY: The IP address of the trusted proxy. Set this to the IP address of the server you are using to send requests to this app. (probably localhost)

## Starting the project
After you have created the .env file, you can start the project by running the following command:
```
sudo docker-compose up -d
```
With logging attatched:
```
sudo docker-compose up -d && sudo docker-compose logs -f
```

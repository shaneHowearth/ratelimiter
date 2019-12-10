# ratelimiter

## Requirements
`docker` and `docker-compose`

# Installation
Ensure no application is listening on localhost:80

Run `docker-compose up`

## Usage
Setting the limit and time span is done through environment variables, and these are set in the `docker-compose` file

Once up and running, the server will be listening on localhost server, port 80.
curl localhost:80 to your heart's content.

Note: A message is included in the 200 responses, however if this app was a proper rate limiter it would reverse proxy the requests onto an internal rest server or websaerver.

## Clean up
```
docker stop {ratelimiter_accessdb_1,ratelimiter_limiter_1}
docker rm {ratelimiter_accessdb_1,ratelimiter_limiter_1}
docker rmi {ratelimiter_accessdb:latest,ratelimiter_limiter:latest}
# You may also want to remove the golang and postgres images
docker rmi {golang,postgres}

# WARNING the following command removes all dangling images, uncomment to use
# docker image prune
```
## Assumptions
I've assumed that only the successful connections count toward the limit, however if a requestor was continuing to slam the server, then the unsuccesful connections would need to be logged too.

## Extensibility
Changing the repository to any other technology is a matter of writing a client that satisfies the limiter/integration/repository/storage.Store interface, and adjusting the docker compose file for a new Dockerfile to stand up that repository in its own container.

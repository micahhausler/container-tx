# container-tx

A re-implementation of [container-transform](https://github.com/micahhausler/container-transform) in go.

## Usage

```
Usage of ./container-tx: [flags] <file>

    Valid input types:  [compose ecs]
    Valid output types: [compose ecs cli]

    If no file is specified, defaults to STDIN

  -i, --input string
    	The format of the input. (default "compose")
  -o, --output string
    	The format of the output. (default "ecs")
  --version
    	print version and exit
```

## Examples

* [Compose --> ECS](#docker-compose to ECS Task)
* [Compose --> CLI](#docker-compose to docker cli)
* [ECS --> Compose](#ECS to Compose)

### docker-compose to ECS Task

```
$ cat docker-compose.yaml
version: '2.0'
services:
  web:
    dns:
    - 8.8.8.8
    image: "alpine"
    labels:
      com.example.description: "Accounting webapp"
      com.example.department: "Finance"
      com.example.label-with-empty-value: ""
    logging:
      driver: gelf
      options:
        tag: web
        gelf-address: "udp://127.0.0.1:12900"
    ports:
     - "5000:5000"
     - "5000"
     - "53:53/udp"
    volumes:
    - "/etc/ssl:/etc/ssl:ro"
    - .:/code
$ cat docker-compose.yaml  | ./container-tx
{
    "family": "",
    "containerDefinitions": [
        {
            "dnsServers": [
                "8.8.8.8"
            ],
            "image": "alpine",
            "dockerLabels": {
                "com.example.department": "Finance",
                "com.example.description": "Accounting webapp",
                "com.example.label-with-empty-value": ""
            },
            "logConfiguration": {
                "logDriver": "gelf",
                "options": {
                    "gelf-address": "udp://127.0.0.1:12900",
                    "tag": "web"
                }
            },
            "memory": 4,
            "name": "web",
            "portMappings": [
                {
                    "hostPort": 53,
                    "containerPort": 53,
                    "protocol": "udp"
                },
                {
                    "hostPort": 5000,
                    "containerPort": 5000,
                    "protocol": "tcp"
                },
                {
                    "containerPort": 5000,
                    "protocol": "tcp"
                }
            ],
            "mountPoints": [
                {
                    "sourceVolume": ".",
                    "containerPath": "/code"
                },
                {
                    "sourceVolume": "etc-ssl",
                    "containerPath": "/etc/ssl",
                    "readOnly": true
                }
            ]
        }
    ],
    "volumes": [
        {
            "name": "etc-ssl",
            "host": {
                "sourcePath": "/etc/ssl"
            }
        },
        {
            "name": ".",
            "host": {
                "sourcePath": "."
            }
        }
    ]
}
```

### docker-compose to docker cli

```
$ cat docker-compose.yaml
version: '2.0'
services:
  web:
    dns:
    - 8.8.8.8
    image: "alpine"
    labels:
      com.example.description: "Accounting webapp"
      com.example.department: "Finance"
      com.example.label-with-empty-value: ""
    logging:
      driver: gelf
      options:
        tag: web
        gelf-address: "udp://127.0.0.1:12900"
    ports:
     - "5000:5000"
     - "5000"
     - "53:53/udp"
    volumes:
    - "/etc/ssl:/etc/ssl:ro"
    - .:/code
$ container-tx -o script docker-compose.yaml
######## web ########
docker run \
    --dns 8.8.8.8 \
    --label com.example.department=Finance \
    --label com.example.description=Accounting webapp \
    --label com.example.label-with-empty-value= \
    --log-driver gelf \
    --log-opt gelf-address=udp://127.0.0.1:12900 \
    --log-opt tag=web \
    --name web \
    --publish 5000:5000 \
    --publish 5000 \
    --publish 53:53/udp \
    --volume /etc/ssl:/etc/ssl:ro \
    --volume .:/code \
    alpine
```

### ECS to Compose

```
$ cat task.json
{
    "family": "pythonapp",
    "volumes": [
        {
            "name": "host_etc",
            "host": {
                "sourcePath": "/etc"
            }
        }
    ],
    "containerDefinitions": [
        {
            "cpu": 200,
            "essential": true,
            "name": "db",
            "memory": 2048,
            "image": "postgres:9.3"
        },
        {
            "cpu": 400,
            "links": [
                "db"
            ],
            "mountPoints": [
                {
                    "sourceVolume": "host_etc",
                    "containerPath": "/usr/local/etc",
                    "readOnly": true
                }
            ],
            "portMappings": [
                {
                    "hostPort": 8000,
                    "containerPort": 8000
                }
            ],
            "memory": 64,
            "entrypoint": [
                "uwsgi"
            ],
            "command": [
                "--json",
                "uwsgi.json"
            ],
            "environment": [

                {
                    "name": "BROKER_URL",
                    "value": "redis://redis:6379/0"
                },
                {
                    "name": "PGPASSWORD",
                    "value": "postgres"
                },
                {
                    "name": "PGUSER",
                    "value": "postgres"
                },
                {
                    "name": "PGHOST",
                    "value": "db"
                }
            ],
            "name": "web",
            "essential": true,
            "image": "me/myapp"
        }
    ]
}
$ ./container-tx -i ecs -o compose task.json
version: "2"
services:
  db:
    cpu_shares: 200
    image: postgres:9.3
    mem_limit: 2147483648
  web:
    command: --json uwsgi.json
    cpu_shares: 400
    entrypoint: uwsgi
    environment:
      values:
        BROKER_URL: redis://redis:6379/0
        PGHOST: db
        PGPASSWORD: postgres
        PGUSER: postgres
    image: me/myapp
    links:
    - db
    mem_limit: 67108864
    ports:
    - 8000:8000
    volumes:
    - /etc:/usr/local/etc:ro
```

## Wishlist

- [ ] Add tests/CI
- [ ] Add docker builds to CI

## License
MIT License. See [License](/LICENSE) for full text

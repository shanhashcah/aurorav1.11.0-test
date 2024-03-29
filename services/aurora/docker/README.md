# Running Hcnet with Docker Compose

## Dependencies

The only dependency you will need to install is [Docker](https://www.docker.com/products/docker-desktop).

## Start script

[start.sh](./start.sh) will setup the env file and run docker-compose to start the Hcnet docker containers. Feel free to use this script, otherwise continue with the next two steps.

The script takes one optional parameter which configures the Hcnet network used by the docker containers. If no parameter is supplied, the containers will run on the Hcnet test network.

`./start.sh pubnet` will run the containers on the Hcnet public network.

`./start.sh standalone` will run the containers on a private standalone Hcnet network.

## Set up a .env file

Mac OS X and Windows users should create an [`.env`](https://docs.docker.com/compose/environment-variables/#the-env_file-configuration-option) file which consists of:

`NETWORK_MODE=bridge`

Linux users should also create an `.env` file. However, the contents of the file should look like:

`NETWORK_MODE=host`

Additionally, you will need to add `127.0.0.1 host.docker.internal` to the `/etc/hosts` file on your linux machine.

If https://github.com/docker/for-linux/issues/264 is ever fixed then it won't be necessary to alias `host.docker.internal` to localhost and there won't be any differences between the Linux and Mac OS X / Windows configurations.


## Run docker-compose

Run the following command to start all the Hcnet docker containers:

```
docker-compose up -d --build
```

Aurora will be exposed on port 8000. Hcnet Core will be exposed on port 11626. The Hcnet Core postgres instance will be exposed on port 5641.
The Aurora postgres instance will be exposed on port 5432.

## Swapping in a local service

If you're developing a service locally you may want to run that service locally while also being able to interact with the other Hcnet components running in Docker. You can do that by stopping the container corresponding to the service you're developing.

For example, to run Aurora locally from source, you would perform the following steps:

```
# stop aurora in docker-compose
docker-compose stop aurora
```

Now you can run aurora locally in vscode using the following configuration:
```
    {
        "name": "Launch",
        "type": "go",
        "request": "launch",
        "mode": "debug",
        "remotePath": "",
        "port": 2345,
        "host": "127.0.0.1",
        "program": "${workspaceRoot}/services/aurora/main.go",
        "env": {
            "DATABASE_URL": "postgres://postgres@localhost:5432/aurora?sslmode=disable",
            "HCNET_CORE_DATABASE_URL": "postgres://postgres:mysecretpassword@localhost:5641/hcnet?sslmode=disable",
            "NETWORK_PASSPHRASE": "Test SDF Network ; September 2015",
            "HCNET_CORE_URL": "http://localhost:11626",
            "INGEST": "true",
        },
        "args": []
    }
```

Similarly, to run Hcnet core locally from source and have it interact with Aurora in docker, all you need to do is run `docker-compose stop core` before running Hcnet core from source.

## Connecting to the Hcnet Public Network

By default, the Docker Compose file configures Hcnet Core to connect to the Hcnet test network. If you would like to run the docker containers on the
Hcnet public network, run `docker-compose -f docker-compose.yml -f docker-compose.pubnet.yml up -d --build`. 

To run the containers on a private stand-alone network, run `docker-compose -f docker-compose.yml -f docker-compose.standalone.yml up -d --build`.
When you run Hcnet Core on a private stand-alone network, an account will be created which will hold 100 billion Lumens.
The seed for the account will be emitted in the Hcnet Core logs:

```
2020-04-22T18:39:19.248 GD5KD [Ledger INFO] Root account seed: SC5O7VZUXDJ6JBDSZ74DSERXL7W3Y5LTOAMRF7RQRL3TAGAPS7LUVG3L
```

When running Aurora on a private stand-alone network, Aurora will not start ingesting until Hcnet Core creates its first history archive snapshot. Hcnet Core creates snapshots every 64 ledgers, which means ingestion will be delayed until ledger 64.

When you switch between different networks you will need to clear the Hcnet Core and Hcnet Aurora databases. You can wipe out the databases by running `docker-compose down --remove-orphans -v`.

## Using a specific version of Hcnet Core

By default the Docker Compose file is configured to use the latest version of Hcnet Core. To use a specific version, you can edit [docker-compose.yml](./docker-compose.yml) and set the appropriate [tag](https://hub.docker.com/r/hcnet/hcnet-core/tags) on the Hcnet Core docker image

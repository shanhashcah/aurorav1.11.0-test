---
title: Aurora Development Guide
---
## Aurora Development Guide

This document describes how to build Aurora from source, so that you can test and edit the code locally to develop bug fixes and new features.

If you are just starting with Aurora and want to try it out, consider the [Quickstart Guide](quickstart.md) instead. For information about administrating a Aurora instance in production, check out the [Administration Guide](admin.md).

## Building Aurora
Building Aurora requires the following developer tools:

- A [Unix-like](https://en.wikipedia.org/wiki/Unix-like) operating system with the common core commands (cp, tar, mkdir, bash, etc.)
- Golang 1.14 or later
- [git](https://git-scm.com/) (to check out Aurora's source code)
- [mercurial](https://www.mercurial-scm.org/) (needed for `go-dep`)

1. Set your [GOPATH](https://github.com/golang/go/wiki/GOPATH) environment variable, if you haven't already. The default `GOPATH` is `$HOME/go`. When building any Go package or application the binaries will be installed by default to `$GOPATH/bin`.
2. Checkout the code into any directory you prefer:
   ```
   git checkout https://github.com/hcnet/go
   ```
   Or if you prefer to develop inside `GOPATH` check it out to `$GOPATH/src/github.com/hcnet/go`:
   ```
   git checkout https://github.com/hcnet/go $GOPATH/src/github.com/hcnet/go
   ```
   If developing inside `GOPATH` set the `GO111MODULE=on` environment variable to turn on Modules for managing dependencies. See the repository [README](../../../../README.md#dependencies) for more information.
3. Change to the directory where the repository is checked out. e.g. `cd go`, or if developing inside the `GOPATH`, `cd $GOPATH/src/github.com/hcnet/go`.
4. Compile the Aurora binary: `go install ./services/aurora`. You should see the resulting `aurora` executable in `$GOPATH/bin`.
5. Add Go binaries to your PATH in your `bashrc` or equivalent, for easy access: `export PATH=${GOPATH//://bin:}/bin:$PATH`

Open a new terminal. Confirm everything worked by running `aurora --help` successfully. You should see an informative message listing the command line options supported by Aurora.

## Set up Aurora's database
Aurora uses a Postgres database backend to store test fixtures and record information ingested from an associated Hcnet Core. To set this up:
1. Install [PostgreSQL](https://www.postgresql.org/).
2. Run `createdb aurora_dev` to initialise an empty database for Aurora's use.
3. Run `aurora db init --db-url postgres://localhost/aurora_dev` to install Aurora's database schema.

### Database problems?
1. Depending on your installation's defaults, you may need to configure a Postgres DB user with appropriate permissions for Aurora to access the database you created. Refer to the [Postgres documentation](https://www.postgresql.org/docs/current/sql-createuser.html) for details. Note: Remember to restart the Postgres server after making any changes to `pg_hba.conf` (the Postgres configuration file), or your changes won't take effect!
2. Make sure you pass the appropriate database name and user (and port, if using something non-standard) to Aurora using `--db-url`. One way is to use a Postgres URI with the following form: `postgres://USERNAME:PASSWORD@localhost:PORT/DB_NAME`.
3. If you get the error `connect failed: pq: SSL is not enabled on the server`, add `?sslmode=disable` to the end of the Postgres URI to allow connecting without SSL.
4. If your server is responding strangely, and you've exhausted all other options, reboot the machine. On some systems `service postgresql restart` or equivalent may not fully reset the state of the server.

## Run tests
At this point you should be able to run Aurora's unit tests:
```bash
cd $GOPATH/src/github.com/hcnet/go/services/aurora
go test ./...
```

## Set up Hcnet Core
Aurora provides an API to the Hcnet network. It does this by ingesting data from an associated `hcnet-core` instance. Thus, to run a full Aurora instance requires a `hcnet-core` instance to be configured, up to date with the network state, and accessible to Aurora. Aurora accesses `hcnet-core` through both an HTTP endpoint and by connecting directly to the `hcnet-core` Postgres database.

The simplest way to set up Hcnet Core is using the [Hcnet Quickstart Docker Image](https://github.com/hcnet/docker-hcnet-core-aurora). This is a Docker container that provides both `hcnet-core` and `aurora`, pre-configured for testing.

1. Install [Docker](https://www.docker.com/get-started).
2. Verify your Docker installation works: `docker run hello-world`
3. Create a local directory that the container can use to record state. This is helpful because it can take a few minutes to sync a new `hcnet-core` with enough data for testing, and because it allows you to inspect and modify the configuration if needed. Here, we create a directory called `hcnet` to use as the persistent volume: `cd $HOME; mkdir hcnet`
4. Download and run the Hcnet Quickstart container:

```bash
docker run --rm -it -p "8000:8000" -p "11626:11626" -p "11625:11625" -p"8002:5432" -v $HOME/hcnet:/opt/hcnet --name hcnet hcnet/quickstart --testnet
```

In this example we run the container in interactive mode. We map the container's Aurora HTTP port (`8000`), the `hcnet-core` HTTP port (`11626`), and the `hcnet-core` peer node port (`11625`) from the container to the corresponding ports on `localhost`. Importantly, we map the container's `postgresql` port (`5432`) to a custom port (`8002`) on `localhost`, so that it doesn't clash with our local Postgres install.
The `-v` option mounts the `hcnet` directory for use by the container. See the [Quickstart Image documentation](https://github.com/hcnet/docker-hcnet-core-aurora) for a detailed explanation of these options.

5. The container is running both a `hcnet-core` and a `aurora` instance. Log in to the container and stop Aurora:
```bash
docker exec -it hcnet /bin/bash
supervisorctl
stop aurora
```

## Check Hcnet Core status
Hcnet Core takes some time to synchronise with the rest of the network. The default configuration will pull roughly a couple of day's worth of ledgers, and may take 15 - 30 minutes to catch up. Logs are stored in the container at `/var/log/supervisor`. You can check the progress by monitoring logs with `supervisorctl`:
```bash
docker exec -it hcnet /bin/bash
supervisorctl tail -f hcnet-core
```

You can also check status by looking at the HTTP endpoint, e.g. by visiting http://localhost:11626 in your browser.

## Connect Aurora to Hcnet Core
You can connect Aurora to `hcnet-core` at any time, but Aurora will not begin ingesting data until `hcnet-core` has completed its catch-up process.

Now run your development version of Aurora (which is outside of the container), pointing it at the `hcnet-core` running inside the container:

```bash
aurora --db-url="postgres://localhost/aurora_dev" --hcnet-core-db-url="postgres://hcnet:postgres@localhost:8002/core" --hcnet-core-url="http://localhost:11626" --port 8001 --network-passphrase "Test SDF Network ; September 2015" --ingest
```

If all is well, you should see ingest logs written to standard out. You can test your Aurora instance with a query like: http://localhost:8001/transactions?limit=10&order=asc. Use the [Hcnet Laboratory](https://www.hcnet.org/laboratory/) to craft other queries to try out,
and read about the available endpoints and see examples in the [Aurora API reference](https://www.hcnet.org/developers/aurora/reference/).

## The development cycle
Congratulations! You can now run the full development cycle to build and test your code.
1. Write code + tests
2. Run tests
3. Compile Aurora: `go install github.com/hcnet/go/services/aurora`
4. Run Aurora (pointing at your running `hcnet-core`)
5. Try Aurora queries

Check out the [Hcnet Contributing Guide](https://github.com/hcnet/docs/blob/master/CONTRIBUTING.md) to see how to contribute your work to the Hcnet repositories. Once you've got something that works, open a pull request, linking to the issue that you are resolving with your contribution. We'll get back to you as quickly as we can.

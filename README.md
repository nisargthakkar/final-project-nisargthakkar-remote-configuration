# Remote Configuration Management (RCM)
## Description
A solution to manage application configuration in production environment without the need to deploy the application to production. This system intended to be run using Kubernetes.

## Design
The system consists of four major components:
1. Datastore: The place where all configuration information is stored.
2. Configuration Update Client (`CUC`): Client program used to submit application configuration to be delivered to all application nodes.
3. Configuration Management Client (`CMC`): Client which runs as a sidecar alongside user applications. This client periodically requests configuration updates from the configuration management server and exposes a REST API to interact with the user application.
4. Configuration Management Server (`CMS`): Service which handles updates and ensures the changes in configuration are available at all application nodes. 

## Technologies used
gRPC for communication between all internal components (`CUC` <-> `CMS` <-> `CMC`)

REST for communications between applications and CMC

`CMS`, `CUC`, `CMC` written in Golang

MySQL used as the choice of datastore

Client applications written in Bash and Python

## Code structure
`server/` : Code for CMS

`client/` : Code for CMC

`userclient/` : Code for CUC

`front-tier/` : Sample python application using RCM

`mid-tier/` : Sample Shell script using RCM

`Makefile` : To build the code

`k8s-configs/` : Configs to define kubernetes services and deployments

`test/configs/` : Sample configurations for two applications

## Usage
To build the system, run `make all`

To make the docker images, run `make docker`

To deploy the database run `make dbsetup`

To set up the required schema run `make dblogin`

To create our database schemas, run the following commands:
```
use appconfig;

DROP TABLE configurations;

CREATE TABLE configurations (
	is_valid BOOLEAN,
	application VARCHAR(20),
	config_key VARCHAR(20),
	config_value TEXT,
	update_time INT(11)
);
```

Once the DB Table is set up, we can start our services.

To startup the services and applications, run `make system`

To stop the services and applications, run `make halt`

To stop the database run `make dbkill`

To start the database again run `make db`

To delete the persistent volumes of the database, run `make dbreset`

To update configs for any application, run `make updateconfig CONFIG=<configfile.yml>`


## YAML config structure
```
app: SampleApplication
configs:
  k1: 'v1'
  k2: 'v2'
  ...
 ```
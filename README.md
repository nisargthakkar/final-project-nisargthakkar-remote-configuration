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

Various client applications written in different languages

## Code structure
`server/` : Code for CMS

`client/` : Code for CMC

`userclient/` : Code for CUC

`front-tier/` : Sample python application using RCM

`mid-tier/` : Sample Shell script using RCM

`Makefile` : To build the code

## Usage
To build the system, run ```make all```

To run `CMS`, run ```server/server```

To run `CMC`, run ```client/client 127.0.0.1:3000```

To run `CUC`, run ```userclient/userclient 127.0.0.1:3000 <config.yml>```

To run the Python front-tier app, run ```python3 front-tier/front-tier-app.py```

To run the Bash mid-tier app, run ```mid-tier/mid-tier-app.sh```
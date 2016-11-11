# Remote environment client

A command line tool to help with using Continuous Pipe as a remote development environment.

This helps to set up Kubectl, create, build and destroy remote environments and keep files
in sync with the local filesystem.

## Prerequisites

You will need the following:

 * A Continuous Pipe hosted project with the GitHub integration set up
 * The project checked out locally 
 * The IP address, username and password to use for Kubenetes cluster
 * rsync and fswatch installed locally
 * A keen.io write token, project id and event collection name if you want to log usage stats 

Note: if the GitHub repository is not the origin of your checked out project then you will
need to add a remote for that repository. 

## Installation

```
sudo curl https://continuouspipe.github.io/remote-environment-client/downloads/cp-remote-latest > /usr/local/bin/cp-remote
chmod +x /usr/local/bin/cp-remote
```

## Setup

```
cp-remote setup
```

To start using this tool for a project, run the `setup` command from the project root.
 This will install kubectl if you do not have it installed already. It will then 
 ask a series of questions to get the details for the project set up. More information
 about these questions can be found in the [Configuration section](#Configuration) below.
 
Your answers will be stored in a `.cp-remote-env-settings` file in the project root. You 
 will probably want to add this to your .gitignore file.
 
## Creating amd building remote environment

```
cp-remote build
```

### Creating a new remote environments

The `build` command will push changes the branch you have checked out locally to your remote 
 environment branch. Continuous Pipe will then build the environment. You can use the [Continuous Pipe admin
 site](https://ui.continuouspipe.io/) to see when the environment has finished building and 
 to find its IP address.
 
### Rebuilding the remote environment 
 
 To rebuild your remote environment to use the current branch you have checked out you can use the 
  `build` command. This will force push the current branch which will make Continuous Pipe rebuild the
  environment. If the remote environment has the latest commit then it would not be rebuilt, in order
  to force the rebuild a commit is automatically made updating a timestamp file.
 
## Watch
 
 ```
 cp-remote watch
 ```
   
 The `watch` command will sync changes you make locally to a container that's part of the remote environment.
 This will use the default container specified during setup but you can specify another container to sync with. 
 For example, if the service you want to sync to is web:
  
  ```
  cp-remote watch web
  ```
The watch command should be left running, it will however need restarting whenever the remote environment
is rebuilt. 

## ssh 

 ```
 cp-remote ssh
 ```
 
 This will ssh onto the default container specified during setup but you can specify another container to ssh onto.
 For example, if the service you want to ssh on to is web:
 
 ```
 cp-remote ssh web
 ```
 
## Resync

```
  cp-remote resync
  ```
 
When the remote environment is rebuilt it may container changes that you do not have on the local filesystem. 
  For example, for a PHP project part of building the remote environment could be installing the vendors using composer.
  Any new or updated vendors would be on the remote environment but not on the local filesystem which would cause issues, 
  such as autocomplete in your IDE not working correctly. The `resync` command will copy changes  from the remote to the local 
  filesystem. This will resync with the default container specified during setup but you can specify another container.
  For example to resync with the `web` container:
  
  ```
  cp-remote resync web
  ```
  
  To ensure your local changes are kept, the resync command first stashes your changes, syncs from the remote to local,
  reapplies the changes and syncs them from local to the remote.
  
## Port Forwarding

 ```
 cp-remote forward db 3306
 ```
 
The `forward` command will set up port forwarding from the local environment to a container 
on the remote environment that has a port exposed. This is useful for tasks such as connecting 
to a database using a local client. You need to specify the container and the port number 
to forward. For example, with a container named db running MySql you would run:
  
  ```
  cp-remote forward db 3306
  ```
  
  this runs in the foreground, so in another terminal you can use the mysql client to connect:
  
  ```
  mysql -h127.0.0.1 -u dbuser -pdbpass dbname
  ```
  
  You can specify a second port number if the remote port number is different to the local port number:
   
  ```
  cp-remote forward db 3307 3306
  ``` 
  
  Here the local port 3307 is forward to 3306 on the remote, you could then connect using:
  
  ```
  mysql -h127.0.0.1 -P3307 -u dbuser -pdbpass dbname
  ```
  
## Destroy

 ```
 cp-remote destroy
 ```
 
The `destroy` command will delete the remote branch used for your remote environment, Continuous Pipe will
then remove the environment.

## Usage Logging

Usage stats for the longer running commands (build and resync) can be logged to keen.io by providing a 
 write key, project id and event collection name when running the setup command. No stats will be logged
 if these are not provided.
 
## Working with a different environment
 
The `--namespace|-n` option can be used with the `watch`, `ssh`, `resync` and `forward`
 commands to run them against a different environment than the one specified during
 setup. This is useful if you need to access a different environment such as a feature branch
 environment. For example, to ssh onto the `web` container of the `example-feature-my-shiny-new-work`
 environment you can run:
 
 ```
 cp-remote ssh web --namespace=example-feature-my-shiny-new-work
 ```
  
  or
  
 ```
 cp-remote ssh web -n=example-feature-my-shiny-new-work
 ```

## Anybar notifications

To get a status notification for the longer running commands (watch and resync) on OSX you can 
 install [AnyBar](https://github.com/tonsky/AnyBar) and provide a port number to use for 
 it during the `setup` command.
 
## Configuration
 
The `setup` command uses your answers to generate a settings file `.remote` in the 
root of the project. If you need to make changes to the settings you can run the `setup` command again 
or you can directly edit the settings. 

Note: the kubectl cluster IP address, username and password are not stored in this file. To change these
 you can run `setup` again.
 
### PROJECT_KEY
 
This is the project name used in Continuous Pipe. It will be prefixed to all the environment
names created by Continuous Pipe. You can find this on the environments page for the tide on the 
[Continuous Pipe admin site](https://ui.continuouspipe.io/). For example:

![Project Key](/docs/images/project-key.png?raw=true)

Here, this is the environment for the develop branch, so the project key is `develop`.

### REMOTE_BRANCH

The name of the branch you will use for your remote environment. There may be a 
project specific naming convention for this e.g. remote-<your name>

### REMOTE_NAME

The name of the git remote for the GitHub project which has the Continuous Pipe integration.
In most cases you will have cloned the project repo from this so this will be `origin`.   
 
### DEFAULT_CONTAINER     
 
This is an optional setting, if provided this will be used by the `ssh`, `watch` and `resync` commands.
It is the docker-compose service name for the container that you need to provide, it may be called 
something like `web` or `app`.

### KEEN_WRITE_KEY, KEEN_PROJECT_ID and KEEN_EVENT_COLLECTION
 
These are only needed if you want to log usage stats using https://keen.io/. 

### Kubernetes settings
 
These are asked for by the `set up` command but are not stored in the project config file. The
cluster IP address and username can be found on the cluster page for the team in the 
[Continuous Pipe admin site](https://ui.continuouspipe.io/):

![Project Key](/docs/images/kubernetes-config.png?raw=true)

The password can be provided by your Continuous Pipe administrator. 
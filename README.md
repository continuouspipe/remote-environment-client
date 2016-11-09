# Remote environment client

A command line tool to help with using Continuous Pipe as a remote development environment.

This helps to set up Kubectl, create, build and destroy remote environments and keep files
in sync with the local filesystem.

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
 ask a series of questions to get the details for the project set up. 
 
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
 cp-remote watch container-name
 ```
   
 The `watch` command will sync changes you make locally to a container that's part of the remote environment.
 You need to specify which container to sync with using the docker composer service name. 
 This will usually be the main application container. For example, if the service you want to sync to is web:
  
  ```
  cp-remote watch web
  ```
The watch command should be left running, it will however need restarting whenever the remote environment
is rebuilt. 

## ssh 

 ```
 cp-remote ssh container-name
 ```
 
You can ssh onto a running container with the `ssh` command by specifying the container name.
 For example, if the service you want to ssh on to is web:
 
 ```
 cp-remote ssh web
 ```
 
## Resync

```
  cp-remote resync container-name
  ```
 
When the remote environment is rebuilt it may container changes that you do not have on the local filesystem. 
  For example, for a PHP project part of building the remote environment could be installing the vendors using composer.
  Any new or updated vendors would be on the remote environment but not on the local filesystem which would cause issues, 
  such as autocomplete in your IDE not working correctly. The `resync` command will copy changes  from the remote to the local 
  filesystem. You will need to specify the container service name, for example:
  
  ```
  cp-remote resync web
  ```
  
  To ensure your local changes are kept, the resync command first stashes your changes, syncs from the remote to local,
  reapplies the changes and syncs them from local to the remote.
  
## Destroy

 ```
 cp-remote destroy
 ```
 
The `destroy` command will delete the remote branch used for your remote environment, Continuous Pipe will
then remove the environment.



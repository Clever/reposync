# reposync

A script that syncs repos for a GitHub user or organization into a folder on your computer.

## Motivation

As part of a GitHub organization, keeping up with new repos can be difficult.
If repos are created, deleted, or renamed, the state of your local development environment falls of date.

`reposync` solves this by cloning all repos for a GitHub user or organization into a single folder.
If the state of repos in GitHub changes, running `reposync` will clone any new repos and move any deleted repos into an archive folder.
`reposync` only archives local copies of a repo, and it never modifies repos that you've already cloned, so there is no risk of losing data.

## Install

Download the latest release from the [releases](https://github.com/rgarcia/reposync/releases) page, extract the tar.gz, and put the binary in your path.

## Usage

```
$ reposync -h
Usage of reposync:
  -archivedir string
    	Directory to move folders in dir that are not associated with a repo
  -dir string
    	Directory to put folders for each repo
  -dryrun
    	Set to true to print actions instead of performing them
  -org string
    	GitHub organization you'd like to sync a folder with. Must specify this or user
  -orgrepotype string
    	For the GitHub org, type of repos you'd like to pull. Can be all, public, private, forks, sources, member. Default is all. (default "all")
  -token string
    	GitHub token to use for auth
  -user string
    	GitHub user you'd like to sync a folder with. Must specify this or org
  -userrepoforks
    	For the GitHub user, include forks. Default is true. (default true)
  -userrepotype string
    	For the GitHub user, type of repos you'd like to pull. Can be all, owner, member. Default is all. (default "all")
  -version
    	Shows version and exits
```

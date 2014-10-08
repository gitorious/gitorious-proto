# Gitorious Protocol Handlers

[![Build Status](https://travis-ci.org/gitorious/gitorious-proto.svg?branch=master)](https://travis-ci.org/gitorious/gitorious-proto)

This repository contains Gitorious protocol handlers - utilities that provide
access (pull and push) to Gitorious repositories over various protocols.

These tools embrace existing git commands (like `git-shell` and
`git-http-backend`) and add the following additional functionalities:

* access control based on rules defined in Gitorious web interface,
* resolving repository paths from the public ones ("project-name/repo-name") to
  the real paths on disk
* logging.

Normally you don't need to build or download these tools - they are put in
proper place in a new Gitorious installation by
[Gitorious installer](https://gitorious.org/gitorious/ce-installer).

`gitorious-proto` is written in Go (`gitorious-shell`,
`gitorious-http-backend`) and bash (hooks) to limit runtime dependencies
required on Gitorious hosts. The main Gitorious web application, as well as
background job processor and search daemon are running inside Docker
containers. That means the only software Gitorious needs on a host is bash,
curl and Docker (as Go binaries are static and have zero dependencies). This
makes deployment and updates very straightforward.

## Supported protocols

At the moment there are 2 protocols implemented as part of gitorious-proto: ssh
and http.

Gitorious also supports git:// protocol. However it's currently implemented in
[gitorious/mainline](https://gitorious.org/gitorious/mainline). It may be
ported here in the future.

### git-over-ssh protocol: gitorious-shell

Normally `git-shell` is used for handling "git-over-ssh" access.
`gitorious-shell` is a small wrapper around `git-shell` which adds extra
functionalities needed by Gitorious (listed above) and delegates to `git-shell`
to do the actual pull/push handling.

### git-over-http protocol: gitorious-http-backend

Git itself comes with
[git-http-backend](http://git-scm.com/docs/git-http-backend) which is a "server
side implementation of Git over HTTP" in a form of a CGI program.

`gitorious-http-backend` is a HTTP server wrapping `git-http-backend`,
providing concurrent access for multiple clients by spawning new
`git-http-backend` process for each new connection. It also adds authorization
and repository path resolving on top of it.

## Authorization and path resolving

Both `gitorious-shell` and `gitorious-http-backend` depend on an internal
Gitorious API for authorization and repository path resolving.

They make the following HTTP request:

    GET $GITORIOUS_INTERNAL_API_URL/repo-config?username=<username>&repo_path=<public-repo-path>

`$GITORIOUS_INTERNAL_API_URL` defaults to `http://localhost:3000/api/internal`,
which is an API [implemented in
gitorious/mainline](https://gitorious.org/gitorious/mainline/source/master:app/controllers/api/internal/repository_configurations_controller.rb)
(the main Gitorious app).

When user has read access to the repository HTTP status code 200 is expected
with the JSON body including the following information:

    {
      repository_id: 1
      full_path: "/full/path.git"  # full absolute path on disk

      ssh_clone_url: "git@...."      # ssh clone URL for this repository (if ssh access enabled)
      git_clone_url: "git://...."    # git clone URL for this repository (if git access enabled)
      http_clone_url: "http://...."  # http clone URL for this repository (if http access enabled)

      custom_pre_receive_path: "/absolute/hook/path"   # if hook exists
      custom_post_receive_path: "/absolute/hook/path"  # if hook exists
      custom_update_path: "/absolute/hook/path"        # if hook exists
    }

When user doesn't have read access to the repository 403 status is expected.
When `repo_path` is invalid 404 status is expected.

Any non 200 HTTP status will deny the access to the requested repository.

## Hooks

`hooks` directory contains all git hooks that Gitorious uses for authorizing
and processing pushes.

The following environment variables are set by `gitorious-shell` and
`gitorious-http-backend` to be used by hooks:

* `GITORIOUS_PROTO` - set either to `ssh` or `http`
* `GITORIOUS_USER` - set to username of a user requesting pull/push
* `GITORIOUS_REPOSITORY_ID` - set to an ID of a Gitorious repository from/to which
  user pulls/pushes

### pre-receive

Gitorious `pre-receive` hook acts as a guard, authorizing all push operations.

For each refspec line passed to its stdin it makes the following HTTP request:

    GET $GITORIOUS_INTERNAL_API_URL/hooks/pre-receive?username=$GITORIOUS_USER&repository_id=$GITORIOUS_REPOSITORY_ID&refname=<refname>&oldsha=<oldsha>&newsha=<newsha>&mergebase=<mergebase>

where `refname` is set to the name of a ref being pushed to; `oldsha` and
`newsha` set respectively to old and new sha values for the updated ref;
`mergebase` set to the best common ancestor between `oldsha` and `newsha` (see
http://git-scm.com/docs/git-merge-base).

If any of the requests result in HTTP status other than 200 the push is
rejected.

### update

`update` hook is currently not used by Gitorious for anything special. The only
thing it does is delegating to the custom update hook (if any).

### post-update

`post-update` hook's job is calling `git update-server-info` in order to
prepare a packed repository for use over dumb transports.

### post-receive

TODO: describe

    GET $GITORIOUS_INTERNAL_API_URL/hooks/post-receive?...

## Development

As `gitorious-proto` is mostly written in Go language (except hooks) you need a
working Go environment to run and compile the code. Once it's there clone the
repository like this:

    mkdir -p $GOPATH/src/gitorious.org/gitorious
    git clone https://gitorious.org/gitorious/gitorious-proto.git $GOPATH/src/gitorious.org/gitorious/gitorious-proto

You don't need proper Go environment to work on hooks though.

## License

gitorious-proto is free software licensed under the
[GNU Affero General Public License](http://www.gnu.org/licenses/agpl-3.0.html).
gitorious-proto is developed as part of the Gitorious project.

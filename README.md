# Gitorious Protocol Handlers

This repository contains Gitorious protocol handlers - utilities that provide
access (pull and push) to Gitorious repositories over various protocols.

These tools embrace existing git commands (like `git-shell` and
`git-http-backend`) and add the following additional functionalities:

* access control based on rules defined in Gitorious web interface,
* resolving repository paths from the public ones ("project-name/repo-name") to
  the real paths on disk
* logging.

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

    GET $INTERNAL_API_URL/repo-config?username=<username>&repo_path=<public-repo-path>

`$INTERNAL_API_URL` defaults to `http://localhost:3000/api/internal`, which is
an API implemented in
[gitorious/mainline](https://gitorious.org/gitorious/mainline) (the main
Gitorious app).

They expect 200 HTTP status code and the JSON response with the following
information:

    {
      id: 1                       # repository id
      real_path: "real/path.git"  # real path on disk, relative to repositories root

      ssh_clone_url: "git@...."      # ssh clone URL for this repository (if ssh access enabled)
      git_clone_url: "git://...."    # git clone URL for this repository (if git access enabled)
      http_clone_url: "http://...."  # http clone URL for this repository (if http access enabled)

      custom_pre_receive_path: "/absolute/hook/path"   # if hook exists
      custom_post_receive_path: "/absolute/hook/path"  # if hook exists
      custom_update_path: "/absolute/hook/path"        # if hook exists
    }

Any non 200 status will deny the access to the requested repository.

## License

gitorious-proto is free software licensed under the
[GNU Affero General Public License](http://www.gnu.org/licenses/agpl-3.0.html).
gitorious-proto is developed as part of the Gitorious project.

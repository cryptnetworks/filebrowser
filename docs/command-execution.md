# Command Execution

> [!CAUTION]
>
> The **hook runner** and **interactive shell** functionalities have been disabled for all existent and new installations by default from version v2.33.8 and onwards, due to continuous and known security vulnerabilities. You should only use this feature if you are aware of all of the security risks involved. For more up to date information, consult issue [#5199](https://github.com/filebrowser/filebrowser/issues/5199).

## Security decision

Command execution remains available only for compatibility and stays disabled
by default. Enabling it is an explicit risk acceptance. The service does not
provide a process sandbox; deployers must run it as an unprivileged account in
a container or other operating-system sandbox with a minimal filesystem mount.

Every configured executable must be an absolute path. Interactive commands must
match an administrator allowlist entry exactly, including every argument. Event
hooks and authentication hooks are also parsed into an explicit argument vector.
No request or hook is evaluated by `sh`, `bash`, `cmd`, PowerShell, or another
shell, and the legacy `shell` setting is ignored during execution. Relative
commands are rejected so `PATH` cannot select a different executable.

Configured executables and scripts remain part of the trusted computing base.
They must treat arguments and the documented environment values as untrusted,
must not evaluate them as code, and must not invoke secondary commands through
an attacker-controlled search path.

## Hook Runner

The hook runner executes an administrator-configured executable before or after
a supported event. It does not execute shell command text. Right now, these are
the events:

* Copy
* Rename
* Upload
* Delete
* Save

Also, during the execution of the commands set for those hooks, there will be some environment variables available to help you perform your commands:

* `FILE` with the full absolute path to the changed file.
* `SCOPE` with the path to user's scope.
* `TRIGGER` with the name of the event.
* `USERNAME` with the user's username.
* `DESTINATION` with the absolute path to the destination. Only used for **copy** and **rename.**

Hook values are provided through the process environment and may also be used
as explicit argument placeholders. Only the documented variables above are
expanded in arguments; arbitrary process environment values such as `PATH` are
not. Each expanded value remains one argument, so its shell metacharacters are
never interpreted by File Browser.

At this moment, you can edit the commands via the command line interface, using the following commands \(please check the flag `--help` to know more about them\):

```bash
filebrowser cmds add before_copy "/usr/bin/logger $FILE"
filebrowser cmds rm before_copy 0
filebrowser cmds ls
```

Or you can use the web interface to manage them via **Settings** → **Global Settings**.

## Interactive Shell

Within File Browser you can toggle the shell (`< >` icon at the top right) and this will open a shell command window at the bottom of the screen. This functionality can be turned on using the environment variable `FB_DISABLE_EXEC=false` or the flag `--disable-exec=false`.

By default no commands are available as the command list is empty. Each entry
is an exact command and argument vector using an absolute executable path:
allowing `/bin/ls -la` permits only that invocation, not `/bin/ls`,
`/bin/ls /other/path`, a relative `ls`, or a command followed by shell syntax.
Interactive commands are executed directly and never through the configured
shell. To enable commands, add exact invocations on a per-user basis (including
for the Admin user).

You can do this by adding them in Settings > User Management > (edit user) > Commands or to *apply to all new users created from that point forward* they can be set in Settings > Global Settings

> [!NOTE]
> 
> If using a proxy manager then remember to enable websockets support for the File Browser proxy

> [!NOTE]
> 
> If using Docker and you want to add a new command that is not in the base image then you will need to build a custom Docker image using `filebrowser/filebrowser` as a base image.  For example to add 7z:
> 
> ```docker
> FROM filebrowser/filebrowser
> RUN sudo apt install p7zip-full
> ```

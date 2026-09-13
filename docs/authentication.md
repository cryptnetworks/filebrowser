# Authentication

There are three possible authentication methods. Each one of them has its own capabilities and specification. Adding another authentication method is described in [Building File Browser](../CONTRIBUTING.md#authentication-provider).

## JSON Auth (default)

We call it JSON Authentication but it is just the default authentication method and the one that is provided by default if you don't make any changes. It is set by default, but if you've made changes before you can revert to using JSON auth:

```sh
filebrowser config set --auth.method=json
```

This method can also be extended with **reCAPTCHA** verification during login:

```sh
filebrowser config set --auth.method=json \
  --recaptcha.key site-key \
  --recaptcha.secret private-key
```

By default, we use [Google's reCAPTCHA](https://developers.google.com/recaptcha/docs/display) service. If you live in China, or want to use other provider, you can change the host with the following command:

```sh
filebrowser config set --recaptcha.host https://recaptcha.net
```

Where `https://recaptcha.net` is any provider you want.

## Proxy Header

If you have a reverse proxy you want to use to login your users, you do it via our `proxy` authentication method. To configure this method, your proxy must send an HTTP header containing the username of the logged in user:

```sh
filebrowser config set --auth.method=proxy --auth.header=X-My-Header --auth.trustedCIDRs=127.0.0.1/32,::1/128
```

Where `X-My-Header` is the HTTP header provided by your proxy with the username.

> [!WARNING]
> 
> File Browser only honors the proxy header when the request comes from an explicit trusted CIDR. If the proxy can be bypassed, an attacker could still attach the header and get admin access if the app is exposed more broadly. Make sure the proxy strips or overwrites the header, and keep the trusted CIDR list as small as possible.

## Hook Authentication

The Hook Authentication method in FileBrowser allows developers to delegate user authentication to an external script or program. Instead of validating credentials internally, FileBrowser sends the username and password to a custom command defined by the administrator. This command receives the credentials through environment variables and returns key‑value pairs indicating whether the user should be authenticated, blocked, or passed through.

The hook’s output controls user permissions, scope, locale, and other attributes, making it a powerful and extensible authentication mechanism.

> [!WARNING]
>
> The submitted username and password are attacker-controlled and are handed to your hook command as the `USERNAME` and `PASSWORD` environment variables. File Browser runs the command directly, without a shell, so the values themselves are inert. However, your script must treat them as untrusted: always quote them (`"$USERNAME"`, `"$PASSWORD"`) and never pass them unquoted to a shell, `eval`, `bash -c`, command substitution, or backticks. A hook script that shell-evaluates these values turns any login request into remote code execution.

For example, the following code delegates filebrowser authentication to a PowerShell script on Windows. You can configure any command (for example, a script in Python, Node.js, etc.).

```sh
filebrowser config set --auth.method=hook --auth.command="powershell.exe -File C:\route\to\your\script\auth.ps1"
```

This is the code for the auth.ps1 script

```sh
param()

# Get FileBrowser credentials from environment variables
$username = $env:USERNAME
$password = $env:PASSWORD

# Users dictionary (for testing purposes only)
$users = @{
    "admin" = "kideW48v7-SdE*"
    "test"  = "2sDd3-etrytñK"
}

# Check if the user exists in the dictionary and verify the password
if ($users.ContainsKey($username) -and $users[$username] -eq $password) {

    # Successful authentication
    Write-Output "hook.action=auth"        # Hook action (in this case, "auth", is required for successful authentication)

    Write-Output "user.perm.admin=true"    # Set admin role (all permissions)
    #You can also define specific permissions like this:
    Write-Output "user.perm.execute=true"
    Write-Output "user.perm.create=true"
    Write-Output "user.perm.rename=true"
    Write-Output "user.perm.modify=true"
    Write-Output "user.perm.delete=true"
    Write-Output "user.perm.share=true"
    Write-Output "user.perm.download=true"

    Write-Output "user.locale=es"          # Set language
    Write-Output "user.viewMode=list"      # Set view mode
    Write-Output "user.scope=/"            # Set FileBrowser scope
    Write-Output "user.singleClick=true"   # Set single click user configuration
    Write-Output "user.hideDotfiles=false" # Set hide dot files user configuration

    #Set other configuration
} else {
    # Block authentication
    Write-Output "hook.action=block"
}
```

### Hook Output Format

A hook authentication script must output a series of key–value pairs, one per line, using the format:

```
key=value
```

FileBrowser reads these lines and applies the corresponding authentication action and user configuration.

#### Required Fields

The hook must output one of the following actions:

| Key    | Description |
|--------|------------ |
| hook.action=auth | Authenticates the user. FileBrowser will create or update the user if needed. |
| hook.action=block | Rejects authentication. The login attempt fails. |
| hook.action=pass | Delegates authentication to FileBrowser’s internal password validation. |

For most custom authentication flows, auth or block are used.

Example of a successful authentication:

```sh
hook.action=auth
```

#### Optional User Fields

When `hook.action=auth` is returned, the hook may also define additional user attributes. These fields override FileBrowser defaults and allow full customization of the authenticated user.

1. Permissions
```
user.perm.admin=true
user.perm.execute=true
user.perm.create=true
user.perm.rename=true
user.perm.modify=true
user.perm.delete=true
user.perm.share=true
user.perm.download=true
```
> Setting user.perm.admin=true automatically enables all permissions.

2. User Interface and Behavior
```
user.locale=es
user.viewMode=list
user.singleClick=true
user.hideDotfiles=false
```

3. User Scope
```
user.scope=/
```

## No Authentication

We also provide a no authentication mechanism for users that want to use File Browser privately such in a home network. By setting this authentication method, the user with **id 1** will be used as the default users. Creating more users won't have any effect.

```sh
filebrowser config set --auth.method=noauth
```

## Proxy provisioning and migration

Proxy authentication trusts only the immediate TCP peer (`RemoteAddr`), never
`Forwarded` or `X-Forwarded-For`. Trust only the final identity-setting proxy.
With multiple proxies, that final proxy must authenticate the request or receive
identity over a separately authenticated private hop. Do not allow an outer
proxy's arbitrary client header through unchanged.

Existing File Browser accounts can log in through a trusted proxy. Unknown
accounts are denied by default, independently of `signup`. To opt into account
creation, configure isolated homes and permissions first, then run:

```sh
filebrowser config set --createUserDir=true --auth.autoProvision=true
```

Disable it with `filebrowser config set --auth.autoProvision=false`. Old exports
without `autoProvision` import as false. Existing accounts and scopes are not
rewritten. The proxy must supply one nonempty username; duplicate header values,
case-variant duplicates, and comma-joined identity lists are rejected. Leading
and trailing whitespace is removed once; username case is preserved.

### Reverse proxy examples

These are configuration fragments for an HTTPS virtual host with operator-owned
credentials. Bind File Browser to `127.0.0.1:8080` on the same host and trust
`127.0.0.1/32`; for containers use an isolated private network and the actual
proxy source address. Never expose the backend port publicly. Replace example
hostnames, paths, and password hashes before deployment. Authenticate every
route, including API and share routes; these examples intentionally require
proxy authentication even for public shares.

Nginx, inside the TLS server block:

```nginx
location / {
    auth_basic "File Browser";
    auth_basic_user_file /etc/nginx/filebrowser.htpasswd;
    proxy_set_header X-Remote-User $remote_user;
    proxy_set_header Authorization "";
    proxy_set_header Host $host;
    proxy_pass http://127.0.0.1:8080;
}
```

The identity comes from successful Basic authentication and replaces the client
header. See [Nginx authentication](https://nginx.org/en/docs/http/ngx_http_auth_basic_module.html).
Configure File Browser's `auth.header` as `X-Remote-User` for these examples.

Caddy, with an operator-generated password hash:

```caddyfile
files.example.com {
    basic_auth {
        alice REPLACE_WITH_PASSWORD_HASH
    }
    reverse_proxy 127.0.0.1:8080 {
        header_up X-Remote-User {http.auth.user.id}
        header_up -Authorization
    }
}
```

Use `caddy hash-password` to produce the hash. See
[Caddy authentication](https://caddyserver.com/docs/caddyfile/directives/basic_auth)
and [upstream headers](https://caddyserver.com/docs/caddyfile/directives/reverse_proxy#headers).

Traefik dynamic configuration, with a preconfigured `websecure` entry point and
TLS certificate configuration:

```yaml
http:
  routers:
    filebrowser:
      rule: Host(`files.example.com`)
      entryPoints: [websecure]
      tls: {}
      middlewares: [strip-identity, filebrowser-auth]
      service: filebrowser
  middlewares:
    strip-identity:
      headers:
        customRequestHeaders:
          X-Remote-User: ""
    filebrowser-auth:
      basicAuth:
        usersFile: /etc/traefik/filebrowser.htpasswd
        headerField: X-Remote-User
        removeHeader: true
  services:
    filebrowser:
      loadBalancer:
        servers:
          - url: http://127.0.0.1:8080
```

Keep the header name canonical and use a patched Traefik release. Consult
[Traefik BasicAuth](https://doc.traefik.io/traefik/reference/routing-configuration/http/middlewares/basicauth/)
and its [security advisories](https://github.com/traefik/traefik/security/advisories).

Before exposing the proxy, verify that no credentials, a forged identity header,
and duplicated identity headers cannot authenticate as another user. Check that
valid proxy credentials map to the intended existing File Browser account and
that unknown users fail with provisioning disabled. Run the proxy's configuration
validator (`nginx -t` or `caddy validate`) before reloading; test Traefik startup
and middleware ordering in an isolated deployment.

# github-insttoken

Go binary to get a GitHub App ephemeral Installation token for a repository to allow
git `clone` / `pull` etc operations over https.

This is to simplify the steps detailed in [authenticating-with-github-apps](https://developer.github.com/apps/building-github-apps/authenticating-with-github-apps/)

## Usage:
```bash
./github-insttoken \
  --private-key-file YOURGITHUBAPP.2018-10-26.private-key.pem \
  --app-id APP_ID \
  --repo Organisation/project \
  --git-url https://github.example.com/api/v3
```
Returns:
```
token: v1.2a04[...snip...]5172
```

This can now be used in a git clone:
```bash
git clone https://x-access-token:v1.2a04[...snip...]5172@github.com/Organisation/project.git
```

## ISSUE: Via Proxy
If you get a 401 Unauthorized error attempting to resolve the JWT into
an Installation Token when going via a proxy.  Use the workaround below with
the `--jwt-only` option:
```bash
./github-insttoken --private-key-file YOURGITHUBAPP.2018-10-26.private-key.pem \
  --app-id APP-ID \
  --jwt-only
```
Returns:
```
jwt: eyJhbGciOi[...snip...]FJn7HnGA0Pr7A
```
Then follow github instructions using curl to resolve to the final token.

Get Installation ID for your repo
```bash
curl -i -H "Authorization: Bearer $JWT" \
  -H "Accept: application/vnd.github.machine-man-preview+json" \
  https://github.example.com/api/v3/repos/Organisation/project/installation
```
You want the returned id:
```json
{
  "id": 77,
  ...
}
```

Now you can request the Installation token using the above id:
```bash
curl -i -H "Authorization: Bearer $JWT" \
  -H "Accept: application/vnd.github.machine-man-preview+json" \
  https://github.example.com/api/v3/app/installations/77/access_tokens \
  -X POST
```
Returns:
```json
{
  "token": "v1.f5f76[...snip...]99cc7f93",
  "expires_at": "2018-11-07T10:24:00Z"
}
```
You can now use this token to `git clone ...` as above.

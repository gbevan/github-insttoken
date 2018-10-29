# github-insttoken

Go binary to get a GitHub App ephemeral Installation token for a repository to allow
git `clone` / `pull` etc operations over https.

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

# dns-redirect

HTTP server that redirect to another domain depends on DNS configuration

## Example

With the default configuration (`config.yaml.dist`), and this DNS entries

```text
*.redirect.domain.tld. A 127.0.0.1
google.domain.tld CNAME google--dot--com.redirect.domain.tld.
jobs.domain.tld CNAME domain--dot--tld--slash--jobs.redirect.domain.tld.
app.domain.tld CNAME app--dot--domain--dot--tld--colon--8888.redirect.domain.tld.
youtube.domain.tld CNAME youtube--dot--com--slash--watch--int--v--equal--dQw4w9WgXcQ.redirect.domain.tld.
```

Obviously you should replace 127.0.0.1 with the IP(s) of the server who run this service

Each request will redirect with 307 status code 
* `google.domain.tld` -> `https://google.com`
* `jobs.domain.tld` -> `https://domain.tld/jobs`
* `jobs.domain.tld/67` -> `https://domain.tld/jobs/67` original URI is keep by default and append to the location
* `app.domain.tld` -> `https://app.domain.tld:8888`
* `resume.domain.tld` -> `https://youtube.com/watch?v=dQw4w9WgXcQ`

## Configuration

* each keyword (`.`, `/`, `:`, `?`, `=`, `&`, `%`) are configurable, under `redirect.options.keyword.*`
* you can or not keep the URI of each request, under `redirect.options.keep_uri`
* you can keep the schema or enforce `https`, under `redirect.options.enforce_https`
* you can choose permanent redirect (308) or temporary (307), under `redirect.options.permanent_redirect`

If you don't want to/can't add CNAME to your DNS, you can also use static resolver (`resolver.type: "static"`) and list hosts under `resolver.config.hosts`
"""Bounded release reads. Credentials are only sent to the GitHub API origin."""
import hashlib
import os
import time
import urllib.error
import urllib.parse
import urllib.request


class NoRedirect(urllib.request.HTTPRedirectHandler):
    def redirect_request(self, req, fp, code, msg, headers, newurl):
        return None


def retry_read(operation):
    """Retry only transient, read-only failures, at most three attempts.

    No retries for authentication, integrity, TLS validation or missing objects.
    Long Retry-After delays fail explicitly instead of holding a release job.
    """
    for attempt in range(3):
        try:
            return operation()
        except urllib.error.HTTPError as error:
            header = error.headers.get("Retry-After")
            error.close()
            if error.code not in (429, 500, 502, 503, 504) or attempt == 2:
                raise
            delay = int(header) if header and header.isdigit() else 2 ** attempt
            if delay > 30:
                raise
        except (TimeoutError, ConnectionResetError):
            if attempt == 2:
                raise
            delay = 2 ** attempt
        except urllib.error.URLError as error:
            # urllib wraps connection failures, including TLS errors. Only
            # retry known transient reasons, never certificate failures.
            if not isinstance(error.reason, (TimeoutError, ConnectionResetError)) or attempt == 2:
                raise
            delay = 2 ** attempt
        time.sleep(delay)


def request_read(url, consume, *, github_api=False, accept=None, redirects=False):
    if urllib.parse.urlsplit(url).scheme != "https":
        raise ValueError("release downloads require HTTPS")
    headers = {"User-Agent": "moi-release-pipeline"}
    if github_api:
        origin = urllib.parse.urlsplit(url)
        if origin.netloc != "api.github.com" or origin.username or redirects:
            raise ValueError("GitHub credentials require the exact API origin without redirects")
        token = os.environ.get("MOI_RELEASE_GITHUB_TOKEN")
        if token:
            headers["Authorization"] = "Bearer " + token
        headers["Accept"] = accept or "application/vnd.github+json"
    opener = urllib.request.build_opener() if redirects else urllib.request.build_opener(NoRedirect())

    def read():
        request = urllib.request.Request(url, headers=headers)
        with opener.open(request, timeout=60) as response:
            if not response.url.startswith("https://") or (not redirects and response.url != url):
                raise ValueError("unexpected release redirect")
            return consume(response)

    return retry_read(read)


def download(url, limit, **kwargs):
    def read(response):
        data = response.read(limit + 1)
        if len(data) > limit:
            raise ValueError("download exceeds size limit")
        return data
    return request_read(url, read, **kwargs)


def verify_stream(response, expected_sha, expected_size):
    digest, total = hashlib.sha256(), 0
    while True:
        chunk = response.read(min(1024 * 1024, expected_size - total + 1))
        if not chunk:
            break
        total += len(chunk)
        if total > expected_size:
            raise ValueError("release readback size mismatch")
        digest.update(chunk)
    if total != expected_size or digest.hexdigest() != expected_sha:
        raise ValueError("release readback mismatch")


def verify_public(url, expected_sha, expected_size):
    request_read(url, lambda response: verify_stream(response, expected_sha, expected_size))

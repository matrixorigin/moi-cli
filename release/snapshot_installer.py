"""Render an explicitly pinned standalone QA installation, never a channel updater."""
from pathlib import Path
import re
import shlex

from channel_installer import require_client_url


def render(manifest):
    # Import inside render to avoid a module cycle with the snapshot assembler.
    from qa_snapshot import prefix
    from release_candidate import COMPONENTS, MAX_ARCHIVE, PUBLIC_URL, TARGETS
    identity = manifest["snapshot_id"]
    base = prefix(identity)
    url = require_client_url(manifest["astra_url"])
    if manifest.get("schema") != "moi-client-snapshot-v1":
        raise ValueError("snapshot installer requires a snapshot manifest")
    cases = []
    for platform, entry in sorted(manifest["platforms"].items()):
        if platform not in TARGETS or set(entry["components"]) != set(COMPONENTS):
            raise ValueError("invalid snapshot platform/components")
        lines = [f"  {platform})"]
        for name in COMPONENTS:
            item = entry["components"][name]
            expected_url = f"{PUBLIC_URL}/{base}{platform}/{name}.tar.gz"
            if (item["url"] != expected_url or item["key"] != base + platform + "/" + name + ".tar.gz"
                    or not re.fullmatch(r"[0-9a-f]{64}", item["sha256"])
                    or type(item["size"]) is not int or not 0 < item["size"] <= MAX_ARCHIVE):
                raise ValueError("invalid snapshot archive reference")
            members = "moi-cli" if name == "moi-cli" else f"LICENSE\n{name}"
            lines.append(f"    component {shlex.quote(name)} {shlex.quote(expected_url)} "
                         f"{item['sha256']} {item['size']} {shlex.quote(members)}")
        lines.append("    ;;")
        cases.append("\n".join(lines))
    template = Path(__file__).with_name("snapshot_install.sh").read_text()
    return (template.replace("@IDENTITY@", identity).replace("@ASTRA_URL@", shlex.quote(url))
            .replace("@CASES@", "\n".join(cases))).encode()

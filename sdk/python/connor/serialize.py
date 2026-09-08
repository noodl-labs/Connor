"""JSON helpers for span I/O (RFC 0003 §6.3)."""

from __future__ import annotations

import json
from typing import Any

from connor.limits import MAX_PAYLOAD_BYTES, TRUNCATED, UNSERIALIZABLE


def dump_payload(value: Any) -> Any:
    """Return a JSON-serializable value, capped, never raising."""
    if value is None:
        return None
    try:
        raw = json.dumps(value, default=str)
    except (TypeError, ValueError, OverflowError):
        return UNSERIALIZABLE
    if len(raw.encode("utf-8")) > MAX_PAYLOAD_BYTES:
        return TRUNCATED
    try:
        return json.loads(raw)
    except json.JSONDecodeError:
        return UNSERIALIZABLE

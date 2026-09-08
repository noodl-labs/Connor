"""Span dataclass matching Go entities.Span JSON tags (RFC 0003 §6.2)."""

from __future__ import annotations

from dataclasses import dataclass
from typing import Any


KIND_RUN = "run"
KIND_TOOL = "tool"
KIND_LLM = "llm"
KIND_SPAN = "span"

STATUS_OK = "ok"
STATUS_ERROR = "error"


@dataclass
class Span:
    trace_id: str
    run_id: str
    span_id: str
    parent_span_id: str
    name: str
    kind: str
    status: str
    started_at: str
    ended_at: str
    duration_ms: int
    error: str | None = None
    input: Any = None
    output: Any = None
    attributes: dict[str, Any] | None = None
    prompt_tokens: int | None = None
    completion_tokens: int | None = None

    def to_json(self) -> dict[str, Any]:
        out: dict[str, Any] = {
            "trace_id": self.trace_id,
            "run_id": self.run_id,
            "span_id": self.span_id,
            "parent_span_id": self.parent_span_id,
            "name": self.name,
            "kind": self.kind,
            "status": self.status,
            "started_at": self.started_at,
            "ended_at": self.ended_at,
            "duration_ms": self.duration_ms,
        }
        if self.error:
            out["error"] = self.error
        if self.input is not None:
            out["input"] = self.input
        if self.output is not None:
            out["output"] = self.output
        if self.attributes:
            out["attributes"] = self.attributes
        if self.prompt_tokens:
            out["prompt_tokens"] = self.prompt_tokens
        if self.completion_tokens:
            out["completion_tokens"] = self.completion_tokens
        return out

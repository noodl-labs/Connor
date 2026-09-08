"""trace() context manager and run context (RFC 0003 P1)."""

from __future__ import annotations

import json
from contextlib import contextmanager
from contextvars import ContextVar, Token
from datetime import datetime, timezone
from pathlib import Path
from secrets import token_hex
from typing import Any, Iterator

from connor.span import KIND_RUN, STATUS_ERROR, STATUS_OK, Span
from connor.serialize import dump_payload

_current_run: ContextVar[Run | None] = ContextVar("connor_run", default=None)
_current_span_id: ContextVar[str | None] = ContextVar("connor_span_id", default=None)


class NestedTraceError(RuntimeError):
    """V1 forbids nested trace() in the same context (RFC 0003 §6.3)."""


class Run:
    """One agent invocation. Write a Connor run.json via write()."""

    def __init__(self, name: str, run_id: str, trace_id: str) -> None:
        self.name = name
        self.run_id = run_id
        self.trace_id = trace_id
        self.spans: list[Span] = []
        self._started = datetime.now(timezone.utc)
        self._root_span_id = _new_span_id()
        self._finished = False
        self._status = STATUS_OK
        self._error: str | None = None
        self._input: Any = None

    @property
    def root_span_id(self) -> str:
        return self._root_span_id

    def set_input(self, value: Any) -> None:
        self._input = dump_payload(value)

    def add_span(self, span: Span) -> None:
        self.spans.append(span)

    def finish(self, *, ok: bool, error: str | None = None) -> None:
        if self._finished:
            if not ok:
                self._status = STATUS_ERROR
                self._error = error
                if self.spans and self.spans[0].kind == KIND_RUN:
                    self.spans[0].status = STATUS_ERROR
                    if error:
                        self.spans[0].error = error
            return
        self._finished = True
        self._status = STATUS_OK if ok else STATUS_ERROR
        self._error = error
        ended = datetime.now(timezone.utc)
        self.spans.insert(
            0,
            Span(
                trace_id=self.trace_id,
                run_id=self.run_id,
                span_id=self._root_span_id,
                parent_span_id="",
                name=self.name,
                kind=KIND_RUN,
                status=self._status,
                started_at=_rfc3339(self._started),
                ended_at=_rfc3339(ended),
                duration_ms=_duration_ms(self._started, ended),
                error=self._error,
                input=self._input,
            ),
        )

    def snapshot_root_if_open(self) -> None:
        """Ensure a root span exists so write() inside `with` still serializes."""
        if not self._finished:
            self.finish(ok=True)

    def to_artifact(self) -> dict[str, Any]:
        self.snapshot_root_if_open()
        root = self.spans[0]
        passed = root.status == STATUS_OK
        latency = root.duration_ms
        return {
            "version": 1,
            "suite_id": self.name,
            "target": "local",
            "cases": [
                {
                    "id": self.name,
                    "model": self.name,
                    "passed": passed,
                    "reason": "" if passed else "agent_failed",
                    "latency_ms": latency,
                    "attempts": 1,
                    "trajectory": {
                        "run_id": self.run_id,
                        "spans": [s.to_json() for s in self.spans],
                    },
                }
            ],
            "summary": {
                "total": 1,
                "passed": 1 if passed else 0,
                "failed": 0 if passed else 1,
                "pass_rate": 100.0 if passed else 0.0,
                "p50_ms": latency,
                "p95_ms": latency,
            },
        }

    def write(self, path: str | Path) -> None:
        data = self.to_artifact()
        Path(path).write_text(json.dumps(data, indent=2) + "\n", encoding="utf-8")


def current_run() -> Run | None:
    return _current_run.get()


def current_span_id() -> str | None:
    return _current_span_id.get()


def set_current_span_id(span_id: str | None) -> Token:
    return _current_span_id.set(span_id)


def reset_current_span_id(token: Token) -> None:
    _current_span_id.reset(token)


def new_span_id() -> str:
    return _new_span_id()


def rfc3339(dt: datetime) -> str:
    return _rfc3339(dt)


def duration_ms(start: datetime, end: datetime) -> int:
    return _duration_ms(start, end)


@contextmanager
def trace(name: str, *, run_id: str | None = None) -> Iterator[Run]:
    if _current_run.get() is not None:
        raise NestedTraceError("connor: nested trace() is not supported in V1")

    rid = run_id or ("run_" + token_hex(8))
    run = Run(name, rid, token_hex(16))
    run_token = _current_run.set(run)
    span_token = _current_span_id.set(run.root_span_id)
    try:
        yield run
        run.finish(ok=True)
    except BaseException as exc:
        run.finish(ok=False, error=_format_error(exc))
        raise
    finally:
        _current_span_id.reset(span_token)
        _current_run.reset(run_token)


def _new_span_id() -> str:
    return token_hex(8)


def _rfc3339(dt: datetime) -> str:
    utc = dt.astimezone(timezone.utc)
    return utc.strftime("%Y-%m-%dT%H:%M:%S.") + f"{int(utc.microsecond / 1000):03d}Z"


def _duration_ms(start: datetime, end: datetime) -> int:
    ms = int((end - start).total_seconds() * 1000)
    if ms < 0:
        return 0
    return ms


def _format_error(exc: BaseException) -> str:
    return f"{type(exc).__name__}: {exc}"

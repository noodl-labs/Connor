"""@tool decorator — sync and async (RFC 0003 P1)."""

from __future__ import annotations

import functools
import inspect
from datetime import datetime, timezone
from typing import Any, Callable, TypeVar

from connor.serialize import dump_payload
from connor.span import KIND_TOOL, STATUS_ERROR, STATUS_OK, Span
from connor.trace import (
    current_run,
    current_span_id,
    duration_ms,
    new_span_id,
    reset_current_span_id,
    rfc3339,
    set_current_span_id,
)

F = TypeVar("F", bound=Callable[..., Any])


def tool(fn: F) -> F:
    """Record a tool span when an active trace() exists; otherwise call through."""
    if inspect.iscoroutinefunction(fn):

        @functools.wraps(fn)
        async def async_wrapper(*args: Any, **kwargs: Any) -> Any:
            run = current_run()
            if run is None:
                return await fn(*args, **kwargs)
            parent = current_span_id() or run.root_span_id
            span_id = new_span_id()
            started = datetime.now(timezone.utc)
            payload_in = dump_payload(_bind_input(fn, args, kwargs))
            token = set_current_span_id(span_id)
            try:
                result = await fn(*args, **kwargs)
                _record(run, fn, parent, span_id, started, payload_in, result, None)
                return result
            except BaseException as exc:
                _record(run, fn, parent, span_id, started, payload_in, None, exc)
                raise
            finally:
                reset_current_span_id(token)

        return async_wrapper  # type: ignore[return-value]

    @functools.wraps(fn)
    def sync_wrapper(*args: Any, **kwargs: Any) -> Any:
        run = current_run()
        if run is None:
            return fn(*args, **kwargs)
        parent = current_span_id() or run.root_span_id
        span_id = new_span_id()
        started = datetime.now(timezone.utc)
        payload_in = dump_payload(_bind_input(fn, args, kwargs))
        token = set_current_span_id(span_id)
        try:
            result = fn(*args, **kwargs)
            _record(run, fn, parent, span_id, started, payload_in, result, None)
            return result
        except BaseException as exc:
            _record(run, fn, parent, span_id, started, payload_in, None, exc)
            raise
        finally:
            reset_current_span_id(token)

    return sync_wrapper  # type: ignore[return-value]


def _record(
    run: Any,
    fn: Callable[..., Any],
    parent: str,
    span_id: str,
    started: datetime,
    payload_in: Any,
    result: Any,
    error: BaseException | None,
) -> None:
    ended = datetime.now(timezone.utc)
    err_s = None
    status = STATUS_OK
    if error is not None:
        status = STATUS_ERROR
        err_s = f"{type(error).__name__}: {error}"
    run.add_span(
        Span(
            trace_id=run.trace_id,
            run_id=run.run_id,
            span_id=span_id,
            parent_span_id=parent,
            name=fn.__name__,
            kind=KIND_TOOL,
            status=status,
            started_at=rfc3339(started),
            ended_at=rfc3339(ended),
            duration_ms=duration_ms(started, ended),
            error=err_s,
            input=payload_in,
            output=None if error is not None else dump_payload(result),
        )
    )


def _bind_input(fn: Callable[..., Any], args: tuple[Any, ...], kwargs: dict[str, Any]) -> Any:
    try:
        sig = inspect.signature(fn)
        bound = sig.bind_partial(*args, **kwargs)
        bound.apply_defaults()
        data = dict(bound.arguments)
        data.pop("self", None)
        data.pop("cls", None)
        return data
    except (TypeError, ValueError):
        return {"args": list(args), "kwargs": kwargs}

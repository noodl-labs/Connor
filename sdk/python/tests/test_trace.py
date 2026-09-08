from __future__ import annotations

import asyncio
import json
import os
import tempfile
import unittest
from pathlib import Path

from connor import NestedTraceError, tool, trace
from connor.limits import MAX_PAYLOAD_BYTES, TRUNCATED, UNSERIALIZABLE


@tool
async def search_customer(customer_id: str) -> dict:
    return {"name": "Ada"}


@tool
async def refund(payment_id: str) -> dict:
    return {"ok": True}


@tool
def ping(msg: str) -> str:
    return msg.upper()


@tool
def inner() -> str:
    return "in"


@tool
def outer() -> str:
    return inner()


@tool
async def boom() -> None:
    raise ValueError("nope")


@tool
async def sleeper(tag: str) -> str:
    await asyncio.sleep(0.02)
    return tag


class BadStr:
    def __str__(self) -> str:
        raise TypeError("cannot stringify")


@tool
def bad_output() -> BadStr:
    return BadStr()


class TestTrace(unittest.IsolatedAsyncioTestCase):
    async def test_async_two_tools_tree(self) -> None:
        with trace("test-agent") as run:
            await search_customer("123")
            await refund("payment-456")
            self.assertEqual(run.run_id[:4], "run_")
            names = [s.name for s in run.spans]
            self.assertEqual(names, ["search_customer", "refund"])

        self.assertEqual(run.spans[0].kind, "run")
        self.assertEqual(run.spans[0].status, "ok")
        self.assertEqual(run.spans[1].parent_span_id, run.spans[0].span_id)
        self.assertEqual(run.spans[2].parent_span_id, run.spans[0].span_id)
        self.assertEqual(run.spans[1].trace_id, run.spans[0].trace_id)
        self.assertGreaterEqual(run.spans[0].duration_ms, 0)
        self.assertEqual(run.spans[1].input, {"customer_id": "123"})

    def test_sync_tool(self) -> None:
        with trace("sync-agent") as run:
            self.assertEqual(ping("hi"), "HI")
        self.assertEqual(run.spans[1].name, "ping")
        self.assertEqual(run.spans[1].output, "HI")
        self.assertEqual(run.spans[1].kind, "tool")

    def test_nested_tools(self) -> None:
        with trace("nest") as run:
            outer()
        by_name = {s.name: s for s in run.spans}
        self.assertEqual(by_name["outer"].parent_span_id, by_name["nest"].span_id)
        self.assertEqual(by_name["inner"].parent_span_id, by_name["outer"].span_id)

    async def test_concurrent_tools_are_siblings(self) -> None:
        with trace("conc") as run:
            a, b = await asyncio.gather(sleeper("a"), sleeper("b"))
            self.assertEqual({a, b}, {"a", "b"})
        root = run.spans[0]
        tools = [s for s in run.spans if s.kind == "tool"]
        self.assertEqual(len(tools), 2)
        self.assertEqual({t.parent_span_id for t in tools}, {root.span_id})
        self.assertEqual(len({t.span_id for t in tools}), 2)

    async def test_exception_reraise_and_status(self) -> None:
        with self.assertRaises(ValueError) as ctx:
            with trace("err") as run:
                await boom()
        self.assertEqual(str(ctx.exception), "nope")
        self.assertIsInstance(ctx.exception, ValueError)
        self.assertEqual(run.spans[0].status, "error")
        self.assertIn("ValueError", run.spans[0].error or "")
        tool_span = run.spans[1]
        self.assertEqual(tool_span.status, "error")
        self.assertEqual(tool_span.name, "boom")

    async def test_write_after_error(self) -> None:
        with self.assertRaises(ValueError):
            with trace("err-write") as run:
                await boom()
        fd, path = tempfile.mkstemp(suffix=".json")
        os.close(fd)
        try:
            run.write(path)
            data = json.loads(Path(path).read_text(encoding="utf-8"))
            self.assertEqual(data["version"], 1)
            self.assertFalse(data["cases"][0]["passed"])
            self.assertEqual(data["cases"][0]["reason"], "agent_failed")
            self.assertEqual(data["cases"][0]["trajectory"]["run_id"], run.run_id)
        finally:
            os.unlink(path)

    def test_unserializable_output(self) -> None:
        with trace("bad") as run:
            bad_output()
        self.assertEqual(run.spans[1].output, UNSERIALIZABLE)

    def test_truncated_output(self) -> None:
        @tool
        def huge() -> str:
            return "x" * (MAX_PAYLOAD_BYTES + 10)

        with trace("cap") as run:
            huge()
        self.assertEqual(run.spans[1].output, TRUNCATED)

    def test_nested_trace_forbidden(self) -> None:
        with trace("outer"):
            with self.assertRaises(NestedTraceError):
                with trace("inner"):
                    pass

    def test_passthrough_without_trace(self) -> None:
        self.assertEqual(ping("ok"), "OK")

    def test_write_golden_shape(self) -> None:
        with trace("test-agent") as run:
            ping("x")
            fd, path = tempfile.mkstemp(suffix=".json")
            os.close(fd)
            try:
                run.write(path)
                data = json.loads(Path(path).read_text(encoding="utf-8"))
            finally:
                os.unlink(path)
        case = data["cases"][0]
        self.assertEqual(data["version"], 1)
        self.assertTrue(case["passed"])
        self.assertEqual(case["latency_ms"], case["trajectory"]["spans"][0]["duration_ms"])
        names = [s["name"] for s in case["trajectory"]["spans"]]
        self.assertEqual(names[0], "test-agent")
        self.assertEqual(names[1], "ping")
        self.assertEqual(case["trajectory"]["spans"][0]["kind"], "run")


if __name__ == "__main__":
    unittest.main()

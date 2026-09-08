from __future__ import annotations

import json
import sys
import unittest
from pathlib import Path

EXAMPLES = Path(__file__).resolve().parents[1] / "examples"
sys.path.insert(0, str(EXAMPLES.parent))
sys.path.insert(0, str(EXAMPLES))

from connor import trace
from support_agent import handle_ticket  # noqa: E402


class TestSupportAgent(unittest.IsolatedAsyncioTestCase):
    async def test_refund_ticket_records_four_tools(self) -> None:
        with trace("support-agent") as run:
            result = await handle_ticket("Please refund payment pay_456")
            self.assertEqual(result, "resolved")
        names = [s.name for s in run.spans if s.kind == "tool"]
        self.assertEqual(
            names,
            ["lookup_customer", "lookup_payment", "refund", "send_email"],
        )
        self.assertEqual(run.spans[0].kind, "run")
        self.assertEqual(run.spans[0].status, "ok")
        artifact = run.to_artifact()
        self.assertEqual(artifact["version"], 1)
        self.assertTrue(artifact["cases"][0]["passed"])
        self.assertIn("trajectory", artifact["cases"][0])

    async def test_email_failure_is_recorded_then_raised(self) -> None:
        with self.assertRaises(ConnectionError):
            with trace("support-agent") as run:
                await handle_ticket("refund", fail_email=True)
        self.assertEqual(run.spans[0].status, "error")
        failed = [s for s in run.spans if s.status == "error" and s.kind == "tool"]
        self.assertEqual(failed[0].name, "send_email_broken")
        data = json.dumps(run.to_artifact())
        self.assertIn("agent_failed", data)
        self.assertIn("smtp timeout", data)


if __name__ == "__main__":
    unittest.main()

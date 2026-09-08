from __future__ import annotations

import sys
import unittest
from pathlib import Path

EXAMPLES = Path(__file__).resolve().parents[1] / "examples"
sys.path.insert(0, str(EXAMPLES.parent))
sys.path.insert(0, str(EXAMPLES))

from connor import trace
from complex_agent import run_agent


class TestComplexAgent(unittest.IsolatedAsyncioTestCase):
    async def test_refund_path_retries_then_succeeds(self) -> None:
        with trace("complex") as run:
            result = await run_agent(
                customer_id="cus_123",
                payment_id="pay_456",
                high_value=False,
            )
        self.assertEqual(result, "refunded")
        names = [s.name for s in run.spans if s.kind == "tool"]
        self.assertIn("lookup_customer", names)
        self.assertIn("lookup_payment", names)
        self.assertEqual(names.count("chargeback_check"), 2)
        self.assertEqual(names.count("refund"), 1)
        self.assertNotIn("escalate", names)
        failed = [s for s in run.spans if s.name == "chargeback_check" and s.status == "error"]
        self.assertEqual(len(failed), 1)
        policy = next(s for s in run.spans if s.name == "search_policy")
        cache = next(s for s in run.spans if s.name == "cache_get")
        self.assertEqual(cache.parent_span_id, policy.span_id)

    async def test_high_value_escalates_without_refund(self) -> None:
        with trace("complex") as run:
            result = await run_agent(
                customer_id="cus_123",
                payment_id="pay_456",
                high_value=True,
            )
        self.assertEqual(result, "escalated")
        names = [s.name for s in run.spans if s.kind == "tool"]
        self.assertIn("escalate", names)
        self.assertNotIn("refund", names)
        self.assertNotIn("chargeback_check", names)


if __name__ == "__main__":
    unittest.main()

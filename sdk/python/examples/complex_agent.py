"""A denser agent loop: parallel tools, nested tools, retry, branch.

Still not a framework. Connor only records. From repo root:

    make demo-trace-complex
    make demo-trace-complex ARGS='--high-value'
"""

from __future__ import annotations

import argparse
import asyncio
import sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parents[1]))

from connor import tool, trace


def _repo_root() -> Path:
    return Path(__file__).resolve().parents[3]


@tool
async def lookup_customer(customer_id: str) -> dict:
    await asyncio.sleep(0.01)
    return {"id": customer_id, "name": "Ada Lovelace", "email": "ada@example.com"}


@tool
async def lookup_payment(payment_id: str) -> dict:
    await asyncio.sleep(0.01)
    return {"id": payment_id, "amount_usd": 49.0, "status": "captured"}


@tool
async def lookup_payment_high(payment_id: str) -> dict:
    await asyncio.sleep(0.01)
    return {"id": payment_id, "amount_usd": 900.0, "status": "captured"}


@tool
async def cache_get(key: str) -> dict | None:
    await asyncio.sleep(0.005)
    return None


@tool
async def kb_lookup(query: str) -> dict:
    await asyncio.sleep(0.01)
    return {"max_refund_usd": 100.0, "require_chargeback_check": True}


@tool
async def search_policy(topic: str) -> dict:
    cached = await cache_get(f"policy:{topic}")
    if cached is not None:
        return cached
    return await kb_lookup(topic)


_chargeback_attempts = 0


@tool
async def chargeback_check(payment_id: str) -> dict:
    global _chargeback_attempts
    _chargeback_attempts += 1
    await asyncio.sleep(0.01)
    if _chargeback_attempts == 1:
        raise TimeoutError("processor timeout")
    return {"cleared": True, "payment_id": payment_id}


@tool
async def refund(payment_id: str) -> dict:
    await asyncio.sleep(0.01)
    return {"refunded": True, "payment_id": payment_id}


@tool
async def escalate(ticket: str, reason: str) -> dict:
    await asyncio.sleep(0.01)
    return {"escalated": True, "reason": reason}


@tool
async def send_email(to: str, subject: str) -> dict:
    await asyncio.sleep(0.01)
    return {"queued": True, "to": to}


async def run_agent(*, customer_id: str, payment_id: str, high_value: bool) -> str:
    """ReAct-shaped loop: observe → decide → act. Tools are real callables."""
    global _chargeback_attempts
    _chargeback_attempts = 0

    pay_fn = lookup_payment_high if high_value else lookup_payment
    customer, payment = await asyncio.gather(
        lookup_customer(customer_id),
        pay_fn(payment_id),
    )
    policy = await search_policy("refund")

    if payment["amount_usd"] > policy["max_refund_usd"]:
        await escalate(
            f"{customer_id}/{payment_id}",
            f"amount {payment['amount_usd']} over cap {policy['max_refund_usd']}",
        )
        await send_email(customer["email"], "Your request was sent to a specialist")
        return "escalated"

    last_err: Exception | None = None
    for _ in range(2):
        try:
            await chargeback_check(payment_id)
            last_err = None
            break
        except TimeoutError as exc:
            last_err = exc
    if last_err is not None:
        raise last_err

    await refund(payment_id)
    await send_email(customer["email"], "Your refund is on the way")
    return "refunded"


def print_tree(run) -> None:
    children: dict[str, list] = {}
    roots = []
    for span in run.spans:
        if not span.parent_span_id:
            roots.append(span)
            continue
        children.setdefault(span.parent_span_id, []).append(span)

    def walk(span, depth: int) -> None:
        mark = "✓" if span.status == "ok" else "✗"
        extra = f"  ({span.error})" if span.error else ""
        print(f"{'   ' * depth}{mark}  {span.name}  {span.duration_ms}ms{extra}")
        for child in children.get(span.span_id, []):
            walk(child, depth + 1)

    print()
    print(f"RUN  {run.name}  {run.run_id}")
    for root in roots:
        walk(root, 0)
    print()


def resolve_out(path: str) -> Path:
    out = Path(path)
    if not out.is_absolute():
        out = _repo_root() / out
    out.parent.mkdir(parents=True, exist_ok=True)
    return out


async def run_demo(*, high_value: bool, out: Path) -> int:
    with trace("support-agent-complex") as run:
        result = await run_agent(
            customer_id="cus_123",
            payment_id="pay_456",
            high_value=high_value,
        )
        print(f"agent result: {result}")
        run.write(out)
    print_tree(run)
    print(f"wrote {out}")
    names = [s.name for s in run.spans if s.kind == "tool"]
    print(f"tool order: {names}")
    return 0


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--out", default="tmp/support-agent-complex.json")
    parser.add_argument(
        "--high-value",
        action="store_true",
        help="payment $900 → escalate instead of refund",
    )
    args = parser.parse_args()
    return asyncio.run(
        run_demo(high_value=args.high_value, out=resolve_out(args.out))
    )


if __name__ == "__main__":
    raise SystemExit(main())

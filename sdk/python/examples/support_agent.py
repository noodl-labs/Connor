"""Minimal support agent — your stack, Connor only records.

Not an agent framework. From the repo root:

    make demo-trace
    make demo-trace ARGS='--fail'
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
async def refund(payment_id: str) -> dict:
    await asyncio.sleep(0.01)
    return {"refunded": True, "payment_id": payment_id}


@tool
async def send_email(to: str, subject: str) -> dict:
    await asyncio.sleep(0.01)
    return {"queued": True, "to": to, "subject": subject}


@tool
async def send_email_broken(to: str, subject: str) -> dict:
    raise ConnectionError("smtp timeout")


async def handle_ticket(ticket: str, *, fail_email: bool = False) -> str:
    """Stand-in for LangGraph / OpenAI Agents / a custom loop."""
    customer = await lookup_customer("cus_123")
    payment = await lookup_payment("pay_456")
    if "refund" not in ticket.lower():
        return "no_action"
    await refund(payment["id"])
    mail = send_email_broken if fail_email else send_email
    await mail(customer["email"], "Your refund is on the way")
    return "resolved"


def print_tree(run) -> None:
    print()
    print(f"RUN  {run.name}  {run.run_id}")
    for span in run.spans:
        mark = "✓" if span.status == "ok" else "✗"
        if span.kind == "run":
            print(f"{mark}  {span.name}  {span.status}  {span.duration_ms}ms")
            continue
        extra = f"  ({span.error})" if span.error else ""
        print(f"   └─ {mark}  {span.name}  {span.duration_ms}ms{extra}")
    print()


def resolve_out(path: str) -> Path:
    out = Path(path)
    if not out.is_absolute():
        out = _repo_root() / out
    out.parent.mkdir(parents=True, exist_ok=True)
    return out


async def run_demo(*, fail_email: bool, out: Path) -> int:
    try:
        with trace("support-agent") as run:
            result = await handle_ticket(
                "Please refund payment pay_456",
                fail_email=fail_email,
            )
            print(f"agent result: {result}")
            run.write(out)
    except ConnectionError as exc:
        run.write(out)
        print(f"agent error (recorded): {exc}")
        print_tree(run)
        print(f"wrote {out}")
        return 1
    print_tree(run)
    print(f"wrote {out}")
    return 0


def main() -> int:
    parser = argparse.ArgumentParser(description="Connor trace demo — support agent")
    parser.add_argument("--out", default="tmp/support-agent.json")
    parser.add_argument("--fail", action="store_true", help="send_email raises")
    args = parser.parse_args()
    return asyncio.run(run_demo(fail_email=args.fail, out=resolve_out(args.out)))


if __name__ == "__main__":
    raise SystemExit(main())

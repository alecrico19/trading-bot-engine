#!/usr/bin/env python3
"""
Telegram Bot — remote control and monitoring for the trading execution engine.
Connects to the engine's HTTP API and responds to Telegram commands.
"""

import asyncio
import logging
import os
import sys
from datetime import datetime, timezone

import requests
from telegram import Update
from telegram.ext import Application, CommandHandler, ContextTypes

logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s [%(levelname)s] telegram: %(message)s",
    stream=sys.stderr,
)
logger = logging.getLogger("telegram")

API_URL = os.environ.get("API_URL", "http://localhost:8420")
TELEGRAM_TOKEN = os.environ.get("TELEGRAM_TOKEN", "")
AUTHORIZED_USERS = os.environ.get("AUTHORIZED_USERS", "")

if not TELEGRAM_TOKEN:
    logger.fatal("TELEGRAM_TOKEN not set")
    sys.exit(1)

authorized = {x.strip() for x in AUTHORIZED_USERS.split(",") if x.strip()}

if not authorized:
    logger.warning("AUTHORIZED_USERS not set — bot will reject all commands until configured")


def is_authorized(update: Update) -> bool:
    if not authorized:
        logger.warning("unauthorized access from user %s (%s)",
                       update.effective_user.id, update.effective_user.username)
        return False
    return str(update.effective_user.id) in authorized


async def start(update: Update, context: ContextTypes.DEFAULT_TYPE):
    if not is_authorized(update):
        return
    await update.message.reply_text(
        "🤖 *Trading Bot — Remote Control*\n\n"
        "/status — Engine status, equity, P&L\n"
        "/positions — Open positions\n"
        "/pause — Pause all strategies\n"
        "/resume — Resume all strategies\n"
        "/kill — Emergency kill switch\n"
        "/alerts — Recent alerts",
        parse_mode="Markdown",
    )


async def status(update: Update, context: ContextTypes.DEFAULT_TYPE):
    if not is_authorized(update):
        return
    try:
        resp = requests.get(f"{API_URL}/status", timeout=5)
        data = resp.json()
    except Exception as e:
        await update.message.reply_text(f"❌ Could not reach engine: {e}")
        return

    running = "🟢 RUNNING" if data.get("running") else "⏸ PAUSED"
    if data.get("breached"):
        running = "🔴 HALTED (breaker)"

    daily_pnl = data.get("daily_pnl", 0)
    pnl_sign = "📈" if daily_pnl >= 0 else "📉"

    msg = (
        f"*Engine Status*\n"
        f"{running}\n\n"
        f"💰 Equity: `${data.get('equity', 0):,.2f}`\n"
        f"{pnl_sign} Daily P&L: `${daily_pnl:+,.2f}`\n"
        f"📊 Trades: {data.get('trades', 0)} | Wins: {data.get('wins', 0)}\n"
        f"📡 Active signals: {data.get('signals', 0)}\n"
        f"🧠 Strategies: {', '.join(data.get('strategies', [])) or 'none'}"
    )
    await update.message.reply_text(msg, parse_mode="Markdown")


async def positions(update: Update, context: ContextTypes.DEFAULT_TYPE):
    if not is_authorized(update):
        return
    try:
        resp = requests.get(f"{API_URL}/positions", timeout=5)
        positions = resp.json()
    except Exception as e:
        await update.message.reply_text(f"❌ Could not reach engine: {e}")
        return

    if not positions:
        await update.message.reply_text("📭 No open positions")
        return

    lines = ["*Open Positions*"]
    for p in positions:
        side = "🟢 LONG" if p.get("side") == "long" else "🔴 SHORT"
        pnl = p.get("unrealized_pnl", 0)
        pnl_s = f"{pnl:+,.2f}"
        lines.append(
            f"{side} {p['symbol']} × {p.get('amount', 0):.4f} @ {p.get('avg_price', 0):.2f}\n"
            f"  Unrealized P&L: `${pnl_s}`"
        )

    await update.message.reply_text("\n".join(lines), parse_mode="Markdown")


async def pause(update: Update, context: ContextTypes.DEFAULT_TYPE):
    if not is_authorized(update):
        return
    try:
        resp = requests.post(f"{API_URL}/pause", timeout=5)
        if resp.ok:
            await update.message.reply_text("⏸ All strategies paused")
        else:
            await update.message.reply_text("❌ Failed to pause")
    except Exception as e:
        await update.message.reply_text(f"❌ Error: {e}")


async def resume(update: Update, context: ContextTypes.DEFAULT_TYPE):
    if not is_authorized(update):
        return
    try:
        resp = requests.post(f"{API_URL}/resume", timeout=5)
        if resp.ok:
            await update.message.reply_text("▶ All strategies resumed")
        else:
            await update.message.reply_text("❌ Failed to resume")
    except Exception as e:
        await update.message.reply_text(f"❌ Error: {e}")


async def kill(update: Update, context: ContextTypes.DEFAULT_TYPE):
    if not is_authorized(update):
        return
    try:
        resp = requests.post(f"{API_URL}/kill", timeout=5)
        if resp.ok:
            await update.message.reply_text("🔴 *KILL SWITCH ACTIVATED*\nAll strategies paused, open orders cancelled, circuit breaker tripped.", parse_mode="Markdown")
        else:
            await update.message.reply_text("❌ Failed to activate kill switch")
    except Exception as e:
        await update.message.reply_text(f"❌ Error: {e}")


async def alerts(update: Update, context: ContextTypes.DEFAULT_TYPE):
    if not is_authorized(update):
        return
    try:
        resp = requests.get(f"{API_URL}/status", timeout=5)
        data = resp.json()
    except Exception as e:
        await update.message.reply_text(f"❌ Could not reach engine: {e}")
        return

    breached = data.get("breached", False)
    daily_pnl = data.get("daily_pnl", 0)

    msgs = ["*Alert Summary*"]
    if breached:
        msgs.append("🔴 Circuit breaker ACTIVE")
    if daily_pnl < -50:
        msgs.append(f"⚠️ Daily P&L negative: `${daily_pnl:+,.2f}`")
    if not breached and daily_pnl >= -50:
        msgs.append("✅ All systems normal")

    await update.message.reply_text("\n".join(msgs), parse_mode="Markdown")


async def error_handler(update: object, context: ContextTypes.DEFAULT_TYPE):
    logger.error("Exception while handling update: %s", context.error)


def main():
    app = Application.builder().token(TELEGRAM_TOKEN).build()

    app.add_handler(CommandHandler("start", start))
    app.add_handler(CommandHandler("status", status))
    app.add_handler(CommandHandler("positions", positions))
    app.add_handler(CommandHandler("pause", pause))
    app.add_handler(CommandHandler("resume", resume))
    app.add_handler(CommandHandler("kill", kill))
    app.add_handler(CommandHandler("alerts", alerts))
    app.add_error_handler(error_handler)

    logger.info("telegram bot starting (api=%s)", API_URL)
    app.run_polling()


if __name__ == "__main__":
    main()

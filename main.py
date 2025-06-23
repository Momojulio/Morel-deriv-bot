import os, time, threading
from flask import Flask, request, jsonify
import requests, datetime
from deriv_api import DerivAPI
from telegram import Bot

app = Flask(__name__)

# Variables d'environnement
TOKEN = os.getenv("DERIV_TOKEN")
ACCOUNT_TYPE = os.getenv("ACCOUNT_TYPE", "demo")
AMOUNT = float(os.getenv("TRADE_AMOUNT", "1"))
DURATION = int(os.getenv("TRADE_DURATION", "60"))
TGT = os.getenv("TELEGRAM_TOKEN")
CHAT_ID = os.getenv("TELEGRAM_CHAT_ID")

# Initialisation
bot = Bot(token=TGT)
client = DerivAPI(TOKEN, ACCOUNT_TYPE)

# Variables journalières
daily = {"wins": 0, "losses": 0, "total": 0}
last_summary_date = datetime.date.today()

def telegram_send(msg):
    try:
        bot.send_message(chat_id=CHAT_ID, text=msg)
    except:
        pass

def summarize_daily():
    global daily, last_summary_date
    if datetime.date.today() != last_summary_date:
        winrate = daily["wins"] / daily["total"] * 100 if daily["total"] else 0
        msg = f"📆 Récap {last_summary_date.strftime('%d/%m/%Y')}\n"               f"Total trades: {daily['total']}\n✅ Gagnés: {daily['wins']}\n"               f"❌ Perdus: {daily['losses']}\n🎯 Winrate: {winrate:.2f}%"
        telegram_send(msg)
        last_summary_date = datetime.date.today()
        daily = {"wins": 0, "losses": 0, "total": 0}

@app.route("/webhook", methods=["POST"])
def webhook():
    summarize_daily()

    data = request.get_json()
    action = data.get("action")
    symbol = data.get("symbol")
    amount = float(data.get("amount", AMOUNT))
    duration = int(data.get("duration", DURATION))

    # Envoi début trade
    telegram_send(f"📈 TRADE LANCÉ → {action.upper()} | {symbol} | {amount}$ | {duration}s")

    result = client.trade(symbol, action, amount, duration)
    time.sleep(duration + 1)

    res = client.get_contract_details(result["contract_id"])
    profit = float(res["profit"])
    won = profit > 0

    # Envoi résultat
    telegram_send(f"📉 TRADE TERMINÉ → {'✅ GAGNÉ' if won else '❌ PERDU'} | {profit:+.2f}$")

    daily["total"] += 1
    (daily["wins" if won else "losses"]) += 1

    return jsonify({"status": "ok"})

if __name__ == "__main__":
    app.run(host="0.0.0.0", port=8080)
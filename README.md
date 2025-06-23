# Morel Deriv TradeBot 🚀

## 📦 Installation

1. Crée un dépôt GitHub (ex: `morel-deriv-bot`)  
2. Ajoute les fichiers: `Dockerfile`, `main.py`, `requirements.txt`, `.env.example`

## 🔧 Render

- Crée un **Web Service** en mode **Docker**
- Render détectera automatiquement les commandes
- Rien à renseigner pour build/start/port si le Dockerfile est présent

## 🔑 Variables d’environnement

```
DERIV_TOKEN=ton_deriv_token
ACCOUNT_TYPE=demo
TRADE_AMOUNT=1
TRADE_DURATION=60
TELEGRAM_TOKEN=ton_bot_token
TELEGRAM_CHAT_ID=ton_chat_id
```

## ⚠️ TradingView

- Grafiques: V10, V25, V50 en M1
- Crée une alerte alert:

```json
{
  "action":"buy",
  "symbol":"R_10",
  "amount":"1",
  "duration":"60"
}
```

## ✅ Fonctionnement

- Trade envoyé avec l’appel webhook  
- Notification Telegram AU DÉBUT et À LA FIN du trade  
- Résumé journalier à 23h59 UTC
TradingView allows sending requests when a specific event happens. It means that it is possible to set your own data according to a certain pattern your backend can understand. For instance, the project accepts JSON strings in this format:

```JSON
{"ticker": "{{ticker}}", "close": "{{close}}", "time": "{{time}}", "contract_type": "BUY"}
```
```JSON
{"ticker": "{{ticker}}", "close": "{{close}}", "time": "{{time}}", "contract_type": "SELL"}
```

In this way, we can use an indicator on TradingView and set a condition for buy or sell. When TradingView detects that condition is met, the backend server accepts a request with a command to buy or sell. To recognize a symbol where we need to open positions, it requires a mapping with your own broker which provides API. 

The project wasn't updated for a long time, it means that some Deriv API requests may not work now.

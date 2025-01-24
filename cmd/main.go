package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
	"golang.org/x/net/context"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"sync"
	"time"
)

type Config struct {
	Host string
	Port string
}

func (config Config) String() string {
	return fmt.Sprintf(
		"%s:%s",
		config.Host,
		config.Port,
	)
}

func NewConfig() Config {
	return Config{
		Host: os.Getenv("SERVER_HOST"),
		Port: os.Getenv("SERVER_PORT"),
	}
}

func Addr() string {
	return NewConfig().String()
}

func DerivAddr() string {
	query := (url.Values{
		"lang":   {"EN"},
		"app_id": {os.Getenv("APP_ID")},
	}).Encode()

	u := &url.URL{
		Scheme:     "wss",
		Host:       "ws.derivws.com",
		Path:       "/websockets/v3",
		RawQuery:   query,
		ForceQuery: false,
	}

	return u.String()
}

type DerivClient struct {
	token string
	mx    sync.Mutex
	*websocket.Conn
}

func (c *DerivClient) Connect() {
	dialer := &websocket.Dialer{}

	conn, _, err := dialer.Dial(
		DerivAddr(),
		nil,
	)

	if err != nil {
		log.Panicf("%s", err)
	}

	conn.SetPingHandler(func(s string) error {
		if err = conn.SetWriteDeadline(time.Now().Add(time.Minute * 1)); err != nil {
			log.Printf("%s", err)
		}
		return nil
	})

	conn.SetPongHandler(func(s string) error {
		if err = conn.SetReadDeadline(time.Now().Add(time.Minute * 1)); err != nil {
			log.Printf("%s", err)
		}
		return nil
	})

	log.Printf("Connected")

	c.Conn = conn
}

type DerivAuthorizeRequestBody struct {
	Authorize string `json:"authorize"`
}

func (c *DerivClient) Authorize() {
	body := &DerivAuthorizeRequestBody{
		Authorize: c.token,
	}

	if err := c.WriteJSON(body); err != nil {
		log.Panicf("%s", err)
	}
}

type DerivTransactionsStreamRequestBody struct {
	Transaction uint `json:"transaction"`
	Subscribe   uint `json:"subscribe"`
}

func (c *DerivClient) TransactionsStream() {
	body := &DerivTransactionsStreamRequestBody{
		Transaction: 1,
		Subscribe:   1,
	}

	if err := c.WriteJSON(body); err != nil {
		log.Panicf("Disconnected: %s", err)
	}
}

type DerivSymbol string

const (
	DerivEURUSD DerivSymbol = "frxEURUSD"
	DerivAUDJPY DerivSymbol = "frxAUDJPY"
	DerivUSDJPY DerivSymbol = "frxUSDJPY"
	DerivGBPCHF DerivSymbol = "frxGBPCHF"
	DerivGBPJPY DerivSymbol = "frxGBPJPY"
	DerivUSDCAD DerivSymbol = "frxUSDCAD"
	DerivEURJPY DerivSymbol = "frxEURJPY"
	DerivEURGBP DerivSymbol = "frxEURGBP"
	DerivUSDCHF DerivSymbol = "frxUSDCHF"
	DerivNZDCAD DerivSymbol = "frxNZDCAD"
	DerivGBPUSD DerivSymbol = "frxGBPUSD"
)

type DerivDurationUnit string

const (
	DerivDurationUnitSeconds DerivDurationUnit = "s"
	DerivDurationUnitMinutes DerivDurationUnit = "m"
)

type DerivCurrency string

const (
	DerivCurrencyUSD DerivCurrency = "USD"
)

type DerivContractType string

const (
	DerivContractTypePut   DerivContractType = "PUT"
	DerivContractTypeCall  DerivContractType = "CALL"
	DerivContractTypePute  DerivContractType = "PUTE"
	DerivContractTypeCalle DerivContractType = "CALLE"
)

type DerivProposalRequestBody struct {
	Amount       float64           `json:"amount"`
	Duration     uint              `json:"duration"`
	Symbol       DerivSymbol       `json:"symbol"`
	Currency     DerivCurrency     `json:"currency"`
	ContractType DerivContractType `json:"contract_type"`
	DurationUnit DerivDurationUnit `json:"duration_unit"`
}

func (c *DerivClient) Proposal(body *DerivProposalRequestBody) error {
	body.Currency = DerivCurrencyUSD

	if err := c.WriteJSON(body); err != nil {
		return fmt.Errorf("%s", err)
	}

	IsOpened = true

	return nil
}

type DerivBalanceRequestBody struct {
	Balance uint `json:"balance"`
}

func (c *DerivClient) Balance() {
	body := &DerivBalanceRequestBody{
		Balance: 1,
	}

	if err := c.WriteJSON(body); err != nil {
		log.Panicf("Disconnected: %s", err)
	}
}

type Basis string

const (
	DerivBasisStake  Basis = "stake"
	DerivBasisPayout Basis = "payout"
)

type DerivBuyContractRequestBodyParameters struct {
	ContractType DerivContractType `json:"contract_type"`
	DurationUnit DerivDurationUnit `json:"duration_unit"`
	Basis        Basis             `json:"basis"`
	Duration     uint              `json:"duration"`
	Amount       float64           `json:"amount"`
	Symbol       DerivSymbol       `json:"symbol"`
	Currency     DerivCurrency     `json:"currency"`
}

type DerivBuyContractRequestBody struct {
	Buy        uint                                   `json:"buy"`
	Price      float64                                `json:"price"`
	Parameters *DerivBuyContractRequestBodyParameters `json:"parameters"`
}

func (c *DerivClient) BuyContract(body *DerivBuyContractRequestBody) error {
	if err := c.WriteJSON(body); err != nil {
		return fmt.Errorf("%s", err)
	}

	IsOpened = true

	return nil
}

type DerivMessageType string

const (
	DerivMessageTypeBuy         DerivMessageType = "buy"
	DerivMessageTypeBalance     DerivMessageType = "balance"
	DerivMessageTypeAuthorize   DerivMessageType = "authorize"
	DerivMessageTypeTransaction DerivMessageType = "transaction"
)

func (c *DerivClient) Recalculate(balance float64) {
	if Balance > balance {
		LossesCount++
		WasLastLose = true
		return
	}
	WinningsCount++
	WasLastLose = false
}

type DerivTransactionAction string

const (
	DerivTransactionActionBuy  DerivTransactionAction = "buy"
	DerivTransactionActionSell DerivTransactionAction = "sell"
)

func (c *DerivClient) Process() {
	for {
		c.Connect()
		c.Authorize()

		ctx, cancel := context.WithCancel(context.Background())

		go func() {
			for {
				select {
				case <-ctx.Done():
					return
				default:
					if err := c.WriteMessage(websocket.PingMessage, nil); err != nil {
						log.Printf("Disconnected: %s", err)
						break
					}
					time.Sleep(10 * time.Second)
				}
			}
		}()

		for {
			m := &map[string]any{}

			if err := c.ReadJSON(m); err != nil {
				log.Printf("Disconnected: %s", err)
				cancel()
				break
			}

			t := (*m)["msg_type"].(string)

			pointer := (*m)[t]

			if pointer == nil {
				continue
			}

			data := pointer.(map[string]any)

			switch DerivMessageType(t) {
			case DerivMessageTypeBalance:
				balance, _ := data["balance"].(float64)
				Balance = balance
				log.Printf("Balance: %.2f", Balance)
				break
			case DerivMessageTypeAuthorize:
				c.Balance()
				c.TransactionsStream()
				break
			case DerivMessageTypeTransaction:
				action, ok := data["action"].(string)

				if !ok {
					break
				}

				switch DerivTransactionAction(action) {
				case DerivTransactionActionBuy:
					break
				case DerivTransactionActionSell:
					balance, _ := data["balance"].(float64)

					c.mx.Lock()
					c.Recalculate(balance)
					Balance = balance
					IsOpened = false
					log.Printf(
						"Win: %d; Loss: %d; Balance: %.2f",
						WinningsCount,
						LossesCount,
						Balance,
					)
					c.mx.Unlock()
				default:
					break
				}

				break
			case DerivMessageTypeBuy:
				break
			default:
				log.Printf("unknown msg type: %s", (*m)["msg_type"].(string))
			}
		}
	}
}

func NewDerivClient(token string) *DerivClient {
	return &DerivClient{
		token: token,
		mx:    sync.Mutex{},
	}
}

type ContractType string

const (
	ContractTypeBuy  ContractType = "BUY"
	ContractTypeSell ContractType = "SELL"
)

type TradeService struct {
	mx sync.Mutex
}

var client *DerivClient = nil

func NewTradeService() *TradeService {
	return &TradeService{
		mx: sync.Mutex{},
	}
}

var tickers = map[TradingViewTicker]DerivSymbol{
	TradingViewTickerEURUSD: DerivEURUSD,
	TradingViewTickerAUDJPY: DerivAUDJPY,
	TradingViewTickerUSDJPY: DerivUSDJPY,
	TradingViewTickerGBPCHF: DerivGBPCHF,
	TradingViewTickerGBPJPY: DerivGBPJPY,
	TradingViewTickerUSDCAD: DerivUSDCAD,
	TradingViewTickerEURJPY: DerivEURJPY,
	TradingViewTickerEURGBP: DerivEURGBP,
	TradingViewTickerUSDCHF: DerivUSDCHF,
	TradingViewTickerNZDCAD: DerivNZDCAD,
	TradingViewTickerGBPUSD: DerivGBPUSD,
}

const (
	DefaultDuration uint    = 15
	DefaultAmount   float64 = 5.0
)

var (
	LossesCount   uint = 0
	WinningsCount uint = 0
	Balance            = 0.0
	IsOpened           = false
	WasLastLose        = false
	CurrentAmount      = DefaultAmount
)

func GetUpdatedAmount() float64 {
	//if WasLastLose == false {
	//	return DefaultAmount
	//}
	//return CurrentAmount * 2.5
	return CurrentAmount
}

func GetDuration() uint {
	minutes := uint(time.Hour.Minutes()) - uint(time.Now().Minute())

	if minutes < DefaultDuration {
		return DefaultDuration
	}

	return minutes
}

func GetTwoHourDuration() uint {
	return uint(time.Hour.Minutes()) - uint(time.Now().Minute()) + uint(time.Hour.Minutes())
}

func GetHourWith20() uint {
	return uint(time.Hour.Minutes()) - uint(time.Now().Minute()) + uint(time.Minute.Minutes()*20)
}

func (s *TradeService) Process(body *PostTradingViewWebhookRequestBody) error {
	symbol, ok := tickers[body.Ticker]

	if !ok {
		return fmt.Errorf("bad ticket: %s", body.Ticker)
	}

	s.mx.Lock()
	if IsOpened == false {
		var (
			duration = GetHourWith20()
			amount   = GetUpdatedAmount()
		)

		switch body.ContractType {
		case ContractTypeBuy:
			if err := client.BuyContract(&DerivBuyContractRequestBody{
				Buy:   1,
				Price: 9999,
				Parameters: &DerivBuyContractRequestBodyParameters{
					Amount:       amount,
					Symbol:       symbol,
					Duration:     duration,
					Basis:        DerivBasisStake,
					Currency:     DerivCurrencyUSD,
					ContractType: DerivContractTypeCall,
					DurationUnit: DerivDurationUnitMinutes,
				},
			}); err != nil {
				log.Printf("Disconnected: %s", err)
			}
			break
		case ContractTypeSell:
			if err := client.BuyContract(&DerivBuyContractRequestBody{
				Buy:   1,
				Price: 9999,
				Parameters: &DerivBuyContractRequestBodyParameters{
					Amount:       amount,
					Symbol:       symbol,
					Duration:     duration,
					Basis:        DerivBasisStake,
					Currency:     DerivCurrencyUSD,
					ContractType: DerivContractTypePut,
					DurationUnit: DerivDurationUnitMinutes,
				},
			}); err != nil {
				log.Printf("Disconnected: %s", err)
			}
			break
		default:
			return fmt.Errorf("bad contract type: %s", body.ContractType)
		}
	}
	s.mx.Unlock()

	return nil
}

type PostTradingViewWebhookHandler struct {
	service *TradeService
}

func NewPostTradingViewWebhook() http.Handler {
	return &PostTradingViewWebhookHandler{
		service: NewTradeService(),
	}
}

// {"ticker": "{{ticker}}", "close": "{{close}}", "time": "{{time}}", "contract_type": "BUY"}
// {"ticker": "{{ticker}}", "close": "{{close}}", "time": "{{time}}", "contract_type": "SELL"}

type TradingViewTicker string

const (
	TradingViewTickerEURUSD TradingViewTicker = "EURUSD"
	TradingViewTickerAUDJPY TradingViewTicker = "AUDJPY"
	TradingViewTickerUSDJPY TradingViewTicker = "USDJPY"
	TradingViewTickerGBPCHF TradingViewTicker = "GBPCHF"
	TradingViewTickerGBPJPY TradingViewTicker = "GBPJPY"
	TradingViewTickerUSDCAD TradingViewTicker = "USDCAD"
	TradingViewTickerEURJPY TradingViewTicker = "EURJPY"
	TradingViewTickerEURGBP TradingViewTicker = "EURGBP"
	TradingViewTickerUSDCHF TradingViewTicker = "USDCHF"
	TradingViewTickerNZDCAD TradingViewTicker = "NZDCAD"
	TradingViewTickerGBPUSD TradingViewTicker = "GBPUSD"
)

type PostTradingViewWebhookRequestBody struct {
	Close        string            `json:"close"`
	Time         string            `json:"time"`
	Ticker       TradingViewTicker `json:"ticker"`
	ContractType ContractType      `json:"contract_type"`
}

func (h *PostTradingViewWebhookHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	data, err := io.ReadAll(r.Body)

	if err != nil {
		log.Printf("post trading view webhook: %s", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()

	body := &PostTradingViewWebhookRequestBody{}

	if err = decoder.Decode(&body); err != nil {
		log.Printf("post trading view webhook: %s", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err = h.service.Process(body); err != nil {
		log.Printf("post trading view webhook: %s", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}

type PostPingHandler struct {
	mx sync.Mutex
}

func NewPostPingHandler() http.Handler {
	return &PostPingHandler{
		mx: sync.Mutex{},
	}
}

func (h *PostPingHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.mx.Lock()
	client.Balance()
	h.mx.Unlock()

	w.WriteHeader(http.StatusOK)
}

func NewServer() *http.Server {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.AllowContentType("application/json"))

	r.Method(http.MethodPost, "/ping", NewPostPingHandler())
	r.Method(http.MethodPost, "/webhook/tv", NewPostTradingViewWebhook())

	return &http.Server{
		Handler: r,
		Addr:    Addr(),
	}
}

const EnvFileName = ".env"

func init() {
	if err := godotenv.Load(EnvFileName); err != nil {
		log.Panicf("failed to load %s file: %s", EnvFileName, err)
	}
	client = NewDerivClient(os.Getenv("TOKEN"))
}

func main() {
	go client.Process()

	if err := NewServer().ListenAndServe(); err != nil {
		log.Panicf("%s", err)
	}
}

package btcc

import (
	"log"
	"time"

	"github.com/thrasher-/gocryptotrader/common"
	"github.com/thrasher-/gocryptotrader/currency/pair"
	exchange "github.com/thrasher-/gocryptotrader/exchanges"
	"github.com/thrasher-/gocryptotrader/exchanges/orderbook"
	"github.com/thrasher-/gocryptotrader/exchanges/stats"
	"github.com/thrasher-/gocryptotrader/exchanges/ticker"
)

func (b *BTCC) Start() {
	go b.Run()
}

func (b *BTCC) Run() {
	if b.Verbose {
		log.Printf("%s Websocket: %s.", b.GetName(), common.IsEnabled(b.Websocket))
		log.Printf("%s polling delay: %ds.\n", b.GetName(), b.RESTPollingDelay)
		log.Printf("%s %d currencies enabled: %s.\n", b.GetName(), len(b.EnabledPairs), b.EnabledPairs)
	}

	if b.Websocket {
		go b.WebsocketClient()
	}

	for b.Enabled {
		pairs := b.GetEnabledCurrencies()
		for x := range pairs {
			currency := pairs[x]
			go func() {
				ticker, err := b.GetTickerPrice(currency)
				if err != nil {
					log.Println(err)
					return
				}
				log.Printf("BTCC %s: Last %f High %f Low %f Volume %f\n", exchange.FormatCurrency(currency).String(), ticker.Last, ticker.High, ticker.Low, ticker.Volume)
				stats.AddExchangeInfo(b.GetName(), currency.GetFirstCurrency().String(), currency.GetSecondCurrency().String(), ticker.Last, ticker.Volume)
			}()
		}
		time.Sleep(time.Second * b.RESTPollingDelay)
	}
}

func (b *BTCC) GetTickerPrice(p pair.CurrencyPair) (ticker.TickerPrice, error) {
	tickerNew, err := ticker.GetTicker(b.GetName(), p)
	if err == nil {
		return tickerNew, nil
	}

	var tickerPrice ticker.TickerPrice
	tick, err := b.GetTicker(exchange.FormatExchangeCurrency(b.GetName(), p).String())
	if err != nil {
		return tickerPrice, err
	}

	tickerPrice.Pair = p
	tickerPrice.Ask = tick.Sell
	tickerPrice.Bid = tick.Buy
	tickerPrice.Low = tick.Low
	tickerPrice.Last = tick.Last
	tickerPrice.Volume = tick.Vol
	tickerPrice.High = tick.High
	ticker.ProcessTicker(b.GetName(), p, tickerPrice)
	return tickerPrice, nil
}

func (b *BTCC) GetOrderbookEx(p pair.CurrencyPair) (orderbook.OrderbookBase, error) {
	ob, err := orderbook.GetOrderbook(b.GetName(), p)
	if err == nil {
		return ob, nil
	}

	var orderBook orderbook.OrderbookBase
	orderbookNew, err := b.GetOrderBook(exchange.FormatExchangeCurrency(b.GetName(), p).String(), 100)
	if err != nil {
		return orderBook, err
	}

	for x := range orderbookNew.Bids {
		data := orderbookNew.Bids[x]
		orderBook.Bids = append(ob.Bids, orderbook.OrderbookItem{Price: data[0], Amount: data[1]})
	}

	for x := range orderbookNew.Asks {
		data := orderbookNew.Asks[x]
		orderBook.Asks = append(ob.Asks, orderbook.OrderbookItem{Price: data[0], Amount: data[1]})
	}

	orderBook.Pair = p
	orderbook.ProcessOrderbook(b.GetName(), p, orderBook)
	return orderBook, nil
}

//TODO: Retrieve BTCC info
//GetExchangeAccountInfo : Retrieves balances for all enabled currencies for the Kraken exchange
func (b *BTCC) GetExchangeAccountInfo() (exchange.AccountInfo, error) {
	var response exchange.AccountInfo
	response.ExchangeName = b.GetName()
	return response, nil
}

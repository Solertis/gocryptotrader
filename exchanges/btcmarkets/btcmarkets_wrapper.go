package btcmarkets

import (
	"log"
	"time"

	"github.com/thrasher-/gocryptotrader/common"

	"github.com/thrasher-/gocryptotrader/currency"
	"github.com/thrasher-/gocryptotrader/currency/pair"
	"github.com/thrasher-/gocryptotrader/exchanges"
	"github.com/thrasher-/gocryptotrader/exchanges/orderbook"
	"github.com/thrasher-/gocryptotrader/exchanges/stats"
	"github.com/thrasher-/gocryptotrader/exchanges/ticker"
)

// Start runs ticker monitor in a new routine
func (b *BTCMarkets) Start() {
	go b.Run()
}

// Run starts a go routine to monitor ticker price
func (b *BTCMarkets) Run() {
	if b.Verbose {
		log.Printf("%s polling delay: %ds.\n", b.GetName(), b.RESTPollingDelay)
		log.Printf("%s %d currencies enabled: %s.\n", b.GetName(), len(b.EnabledPairs), b.EnabledPairs)
	}

	if !common.DataContains(b.EnabledPairs, "AUD") || !common.DataContains(b.EnabledPairs, "AUD") {
		enabledPairs := []string{}
		for x := range b.EnabledPairs {
			enabledPairs = append(enabledPairs, b.EnabledPairs[x]+"AUD")
		}

		availablePairs := []string{}
		for x := range b.AvailablePairs {
			availablePairs = append(availablePairs, b.AvailablePairs[x]+"AUD")
		}

		log.Println("BTCMarkets: Upgrading available and enabled pairs")

		err := b.UpdateEnabledCurrencies(enabledPairs, true)
		if err != nil {
			log.Printf("%s Failed to get config.\n", b.GetName())
			return
		}

		err = b.UpdateAvailableCurrencies(availablePairs, true)
		if err != nil {
			log.Printf("%s Failed to get config.\n", b.GetName())
			return
		}
	}

	for b.Enabled {
		pairs := b.GetEnabledCurrencies()
		for x := range pairs {
			curr := pairs[x]
			go func() {
				ticker, err := b.GetTickerPrice(curr)
				if err != nil {
					return
				}
				BTCMarketsLastUSD, _ := currency.ConvertCurrency(ticker.Last, "AUD", "USD")
				BTCMarketsBestBidUSD, _ := currency.ConvertCurrency(ticker.Bid, "AUD", "USD")
				BTCMarketsBestAskUSD, _ := currency.ConvertCurrency(ticker.Ask, "AUD", "USD")
				log.Printf("BTC Markets %s: Last %f (%f) Bid %f (%f) Ask %f (%f)\n", exchange.FormatCurrency(curr).String(), BTCMarketsLastUSD, ticker.Last, BTCMarketsBestBidUSD, ticker.Bid, BTCMarketsBestAskUSD, ticker.Ask)
				stats.AddExchangeInfo(b.GetName(), curr.GetFirstCurrency().String(), curr.GetSecondCurrency().String(), ticker.Last, 0)
				stats.AddExchangeInfo(b.GetName(), curr.GetFirstCurrency().String(), "USD", BTCMarketsLastUSD, 0)
			}()
		}
		time.Sleep(time.Second * b.RESTPollingDelay)
	}
}

// GetTickerPrice returns ticker information
func (b *BTCMarkets) GetTickerPrice(p pair.CurrencyPair) (ticker.TickerPrice, error) {
	tickerNew, err := ticker.GetTicker(b.GetName(), p)
	if err == nil {
		return tickerNew, nil
	}

	var tickerPrice ticker.TickerPrice
	tick, err := b.GetTicker(p.GetFirstCurrency().String())
	if err != nil {
		return tickerPrice, err
	}
	tickerPrice.Pair = p
	tickerPrice.Ask = tick.BestAsk
	tickerPrice.Bid = tick.BestBID
	tickerPrice.Last = tick.LastPrice
	ticker.ProcessTicker(b.GetName(), p, tickerPrice)
	return tickerPrice, nil
}

// GetOrderbookEx returns orderbook base on the currency pair
func (b *BTCMarkets) GetOrderbookEx(p pair.CurrencyPair) (orderbook.OrderbookBase, error) {
	ob, err := orderbook.GetOrderbook(b.GetName(), p)
	if err == nil {
		return ob, nil
	}

	var orderBook orderbook.OrderbookBase
	orderbookNew, err := b.GetOrderbook(p.GetFirstCurrency().String())
	if err != nil {
		return orderBook, err
	}

	for x := range orderbookNew.Bids {
		data := orderbookNew.Bids[x]
		orderBook.Bids = append(orderBook.Bids, orderbook.OrderbookItem{Amount: data[1], Price: data[0]})
	}

	for x := range orderbookNew.Asks {
		data := orderbookNew.Asks[x]
		orderBook.Asks = append(orderBook.Asks, orderbook.OrderbookItem{Amount: data[1], Price: data[0]})
	}

	orderBook.Pair = p
	orderbook.ProcessOrderbook(b.GetName(), p, orderBook)
	return orderBook, nil
}

// GetExchangeAccountInfo retrieves balances for all enabled currencies for the
// BTCMarkets exchange
func (b *BTCMarkets) GetExchangeAccountInfo() (exchange.AccountInfo, error) {
	var response exchange.AccountInfo
	response.ExchangeName = b.GetName()
	accountBalance, err := b.GetAccountBalance()
	if err != nil {
		return response, err
	}
	for i := 0; i < len(accountBalance); i++ {
		var exchangeCurrency exchange.AccountCurrencyInfo
		exchangeCurrency.CurrencyName = accountBalance[i].Currency
		exchangeCurrency.TotalValue = accountBalance[i].Balance
		exchangeCurrency.Hold = accountBalance[i].PendingFunds

		response.Currencies = append(response.Currencies, exchangeCurrency)
	}
	return response, nil
}

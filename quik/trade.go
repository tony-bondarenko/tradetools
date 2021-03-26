package quik

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/tony-bondarenko/tradetools"
	"net"
	"strconv"
	"strings"
	"time"
)

const (
	stockClass                = "SPBXM"
	socketConnectionTimeout   = 10 * time.Second
	socketWriteTimeout        = 5 * time.Second
	socketReadTimeout         = 5 * time.Second
	transactionActionNewOrder = "NEW_ORDER"
	transactionTypeLimit      = "L"
	transactionOperationBuy   = "B"
)

type TradeClient struct {
	config        *ClientConfiguration
	cmdConn       net.Conn
	callbackConn  net.Conn
	connected     bool
	transactionId int
	stocks        map[string]Stock
}

type Stock struct {
	code     string
	currency string
	ticker   string
}

type Transaction struct {
	TransactionId string `json:"TRANS_ID,omitempty"`
	Account       string `json:"ACCOUNT,omitempty"`
	OrderType     string `json:"TYPE,omitempty"`
	Action        string `json:"ACTION,omitempty"`
	ClassCode     string `json:"CLASSCODE,omitempty"`
	SecurityCode  string `json:"SECCODE,omitempty"`
	Operation     string `json:"OPERATION,omitempty"`
	Price         string `json:"PRICE,omitempty"`
	Quantity      string `json:"QUANTITY,omitempty"`
}

func CreateClient(configuration interface{}) (*TradeClient, error) {
	config, err := createClientConfig(configuration)
	if err != nil {
		return nil, err
	}
	tradeClient := new(TradeClient)
	tradeClient.config = config
	tradeClient.transactionId = 1
	return tradeClient, nil
}

func (client *TradeClient) GetStocks() ([]tradetools.Stock, error) {
	err := client.loadStocks()
	if err != nil {
		return nil, err
	}

	stocks := make([]tradetools.Stock, 0)
	for _, stock := range client.stocks {
		stocks = append(stocks, tradetools.Stock{Ticker: stock.ticker, Currency: stock.currency})
	}
	return stocks, nil
}

func (client *TradeClient) loadStocks() error {
	if client.stocks != nil {
		return nil
	}
	client.stocks = make(map[string]Stock)

	cmd := Message{Cmd: "getClassSecurityInfo", Data: stockClass}
	response, err := client.sendCommand(cmd)
	if err != nil {
		return err
	}

	if securityMap, ok := response.Data.(map[string]interface{}); ok {
		for ticker, security := range securityMap {
			if securityData, ok := security.(map[string]interface{}); ok {
				stock := Stock{}
				stock.ticker = ticker
				stock.currency, err = getInterfaceMapStringValue(securityData, "currency")
				if err != nil {
					return fmt.Errorf("unknown response format")
				}
				stock.code, err = getInterfaceMapStringValue(securityData, "code")
				if err != nil {
					return fmt.Errorf("unknown response format")
				}
				client.stocks[ticker] = stock
			} else {
				return fmt.Errorf("unknown response format")
			}
		}
		return nil
	}
	return fmt.Errorf("unknown response format")
}

func (client *TradeClient) AddLimit(limit *tradetools.Limit) error {
	transaction, err := client.createLimitTransaction(limit)
	if err != nil {
		return err
	}
	return client.sendTransaction(transaction)
}

func (client *TradeClient) ClearLimits() (int, error) {
	cmd := Message{Cmd: "cancelAllOrders"}
	response, err := client.sendCommand(cmd)
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(response.Data.(string))
}

func (client *TradeClient) createTransaction() (*Transaction, error) {
	transaction := new(Transaction)
	transaction.Account = client.config.account
	client.transactionId++
	transaction.TransactionId = strconv.Itoa(client.transactionId)
	transaction.ClassCode = stockClass
	return transaction, nil
}

func (client *TradeClient) createLimitTransaction(limit *tradetools.Limit) (*Transaction, error) {
	err := client.loadStocks()
	if err != nil {
		return nil, err
	}

	transaction, err := client.createTransaction()
	if err != nil {
		return nil, err
	}
	transaction.Action = transactionActionNewOrder
	transaction.OrderType = transactionTypeLimit
	transaction.Operation = transactionOperationBuy

	stock, ok := client.stocks[limit.Ticker]
	if !ok {
		return nil, fmt.Errorf("unknown stock: %s", limit.Ticker)
	}

	transaction.SecurityCode = stock.code
	transaction.Price = fmt.Sprintf("%.2f", limit.Price)
	transaction.Quantity = strconv.Itoa(limit.Lots)
	return transaction, nil
}

func (client *TradeClient) sendTransaction(transaction *Transaction) error {
	// transaction fee is applied if more then 20 per second is send
	time.Sleep(100 * time.Millisecond)
	cmd := Message{Cmd: "sendTransaction", Data: transaction}
	_, err := client.sendCommand(cmd)
	return err
}

func (client *TradeClient) sendCommand(cmd Message) (*Message, error) {
	if client.connected != true {
		err := client.connect()
		if err != nil {
			return nil, err
		}
	}

	err := client.cmdConn.SetWriteDeadline(time.Now().Add(socketWriteTimeout))
	if err != nil {
		return nil, err
	}

	cmdString, err := json.Marshal(cmd)
	if err != nil {
		return nil, err
	}

	cnt, err := client.cmdConn.Write(cmdString)
	if err != nil {
		return nil, err
	}
	if cnt != len(cmdString) {
		return nil, fmt.Errorf("partitial socket write error for cmd: %s", cmdString)
	}
	cnt, err = client.cmdConn.Write([]byte("\r\n"))
	if err != nil {
		return nil, err
	}
	if cnt != 2 {
		return nil, fmt.Errorf("partitial socket new line write error for cmd: %s", cmd)
	}

	err = client.cmdConn.SetReadDeadline(time.Now().Add(socketReadTimeout))
	if err != nil {
		return nil, err
	}
	responseString, err := bufio.NewReader(client.cmdConn).ReadString('\n')
	if err != nil {
		return nil, err
	}
	responseString = strings.TrimRight(responseString, "\n")

	responseMessage := new(Message)
	err = json.Unmarshal([]byte(responseString), responseMessage)
	if err != nil {
		return nil, err
	}

	if responseMessage.Cmd != cmd.Cmd {
		return responseMessage, fmt.Errorf("got error response: %s", responseMessage.Error)
	}

	return responseMessage, nil
}

func (client *TradeClient) connect() error {
	var err error
	client.cmdConn, err = net.DialTimeout("tcp", fmt.Sprintf("%s:%s", client.config.host, client.config.ports.cmd), socketConnectionTimeout)
	if err != nil {
		return err
	}

	client.callbackConn, err = net.DialTimeout("tcp", fmt.Sprintf("%s:%s", client.config.host, client.config.ports.callback), socketConnectionTimeout)
	if err != nil {
		_ = client.cmdConn.Close()
		return err
	}

	client.connected = true
	return nil
}

func (client *TradeClient) disconnect() error {
	var errors []string
	if err := client.cmdConn.Close(); err != nil {
		errors = append(errors, err.Error())
	}
	if err := client.callbackConn.Close(); err != nil {
		errors = append(errors, err.Error())
	}
	if len(errors) != 0 {
		return fmt.Errorf(strings.Join(errors, "\n"))
	}
	return nil
}

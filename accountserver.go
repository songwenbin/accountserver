package accountserver

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

type PayPlugin interface {
	Monitor(monior int, chaneos chan map[string]float64)
	Coin2Day(i float64) int
}

type Account struct {
	Name  string
	Total int
	Lost  int
	Token string
	Recv  chan string
}

var accounts map[string]*Account = make(map[string]*Account)

func init() {
	// 装入csv账户
	//LoadCSV()
	// 如果没有csv，拉取eos账户

	// 启动使用时间刷新程序
	go FlushTimer(10)
}

func RecvPayAccount(payaccount PayPlugin) {

	var chaneos chan map[string]float64 = make(chan map[string]float64)
	go payaccount.Monitor(5, chaneos)
	for {
		t := <-chaneos

		for key, a := range t {
			fmt.Println(payaccount.Coin2Day(a))
			AddAccount(key, payaccount.Coin2Day(a), payaccount.Coin2Day(a))
		}
	}
}

func NewAccount(name string, total int, lost int) (*Account, chan string) {
	// 新建账户
	t := &Account{
		Name:  name,
		Total: total,
		Lost:  lost,
		Recv:  make(chan string),
	}
	// 放入Map
	accounts[name] = t
	return t, t.Recv
}

func AddAccount(name string, total int, lost int) chan string {

	if value, exists := accounts[name]; exists == true {
		value.Lost = total - value.Total + value.Lost
		value.Total = total
		// restart gotimer
		return value.Recv
	} else {
		account, chanstring := NewAccount(name, total, lost)
		//go Timer(*account)
		fmt.Println(account)
		return chanstring
	}

}

// 定时刷新账户
func FlushTimer(trigger int) {
	for {
		select {
		// 减少账户时间
		case <-time.After(time.Duration(trigger) * time.Second):
			Pasttime()
			for _, v := range accounts {
				fmt.Print(v.Name)
				fmt.Print(v.Lost)
			}
			// 刷新csv文件
			UpdateCSV()
		}
	}
}

// 账户消耗时间
func Pasttime() {
	for _, v := range accounts {
		v.Lost -= 1
	}
}

// 加载CSV文件
func LoadCSV() {
	file, err := os.Open("accounts.csv")
	if err != nil {
	}
	reader := csv.NewReader(file)
	reader.FieldsPerRecord = -1
	record, err := reader.ReadAll()
	if err != nil {
		panic(err)
	}

	for _, item := range record {
		total, _ := strconv.ParseInt(item[1], 0, 0)
		lost, _ := strconv.ParseInt(item[2], 0, 0)
		account := &Account{Name: item[0], Total: int(total), Lost: int(lost), Recv: make(chan string)}
		accounts[item[0]] = account
	}
	defer file.Close()

}

// 更新CSV文件
func UpdateCSV() {
	file, err := os.Create("account.csv")
	if err != nil {
		fmt.Println(err)
	}
	defer file.Close()
	w := csv.NewWriter(file)

	for _, v := range accounts {
		line := []string{v.Name, strconv.Itoa(v.Total), strconv.Itoa(v.Lost)}
		fmt.Println(line)
		if err := w.Write(line); err != nil {
			log.Fatalln("error writing record to csv:", err)
		}
	}

	w.Flush()
}

// 账户定时器函数
func Timer(ac Account) {
	fmt.Print("acount duration:")
	//fmt.Println(time.Duration(i * 24.0))

	time.AfterFunc(time.Duration(ac.Total-ac.Lost)*time.Second, func() {
		fmt.Println(ac.Name)

		delete(accounts, ac.Name)
		ac.Recv <- ac.Name
	})
}

type GetAccountsHandler struct {
	accountmap map[string]*Account
}

func (h *GetAccountsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	type AccountJson struct {
		Name  string
		Total int
		Lost  int
		Token string
	}
	var account_json []AccountJson = make([]AccountJson, 1)

	for _, v := range accounts {

		account_json = append(account_json, AccountJson{
			Name:  v.Name,
			Total: v.Total,
			Lost:  v.Lost,
		})
	}

	result, err := json.Marshal(account_json)
	if err != nil {
		fmt.Println(err.Error())
	}
	fmt.Println(string(result))

	fmt.Fprintf(w, string(result))
}

func StartHTTPServer() {

	getAccountsHandler := GetAccountsHandler{}
	getAccountsHandler.accountmap = accounts
	http.Handle("/getaccounts", &getAccountsHandler)

	server := http.Server{
		Addr: "localhost" + ":" + "8818",
	}

	server.ListenAndServe()
}

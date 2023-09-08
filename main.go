package main

import (
	"log"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func checkErr(e error) {
	if e != nil {
		log.Panic(e)
	}
}

func getToken() string {
	data, err := os.ReadFile("token.txt")
	checkErr(err)
	return string(data)
}

var rawCub = []string{
	"..+---+",
	"./   /|",
	"+---+ |",
	"|   | +",
	"|   |/.",
	"+---+..",
}

type cubData struct {
	n, m, minx, miny, maxx, maxy int
	ans [1000][1000]byte
}

var userState = make(map[int64]int)
var cubDatas = make(map[int64]*cubData)
var hello = []string {
	"你好~",
	"你好世界~",
	"世界你好~",
}

func sets(x, y int, id int64) {
  d := cubDatas[id]
	d.minx = min(d.minx, x - 3)
  d.miny = min(d.miny, y - 2)
  d.maxx = max(d.maxx, x + 3)
  d.maxy = max(d.maxy, y + 5)
	for i := x - 3; i < x + 3; i++ {
		for j := y - 2; j < y + 5; j++ {
			t := rawCub[i - (x - 3)][j - (y - 2)]
			if t == '.' && d.ans[i][j] != 0 {
				continue
			}
			d.ans[i][j] = t
		}
	}
}

func print(id int64) string {
	d := cubDatas[id]
	var reply string
	log.Println(d.minx, d.maxx, d.miny, d.maxy)
	for i := d.minx; i < d.maxx; i++ {
		for j := d.miny; j < d.maxy; j++ {
			if d.ans[i][j] == 0{
				reply += "."
			} else {
				reply += string(d.ans[i][j])
			}
		}
		reply += "\n"
	}
	log.Println(reply)
	return reply
}

func main() {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	token := getToken()
	bot, err := tgbotapi.NewBotAPI(token)
	checkErr(err)

	log.Printf("Authorized on account %s\n", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)
	for update := range updates {
		if update.Message == nil {
			continue
		}
		text := update.Message.Text
		// username := update.Message.From.UserName
		id := update.Message.From.ID
		var replyText string
		switch {
			case strings.Contains(text, "/cub"):
				replyText = "请输入要生成区域的纵轴(n)和横轴(m)长，以左上角为原点："
				userState[id] = 1
			default:
				switch userState[id] {
					case 1:
						nums := strings.Split(text, " ")
						if len(nums) != 2 {
							replyText = "请输入两个数字！"
							userState[id] = 0
						} else {
							var n, m int
							n, err1 := strconv.Atoi(nums[0])
							m, err2 := strconv.Atoi(nums[1])
							if n > 10 || m > 10 || err1 != nil || err2 != nil {
								replyText = "非数字, 或某个数字大于50, 请重试"
								userState[id] = 0
							} else {
								cubDatas[id] = &cubData{
									n: n,
									m: m,
									minx: 1 << 30,
									miny: 1 << 30,
									maxx: -1,
									maxy: -1,
								}
								replyText = "请输入对应总数的数字，从左到右，从上到下（共 n*m 个）："
								userState[id] = 2
							}
						}
					case 2:
						nums := strings.Split(text, " ")
						if len(nums) != cubDatas[id].n * cubDatas[id].m {
							replyText = "数字总数不等于 n*m , 请重试"
						} else {
							var maxh = -1
							n, m := cubDatas[id].n, cubDatas[id].m
							for i := 0; i < n; i++ {
								for j := 0; j < m; j++ {
									t, _ := strconv.Atoi(nums[i * n + j])
									maxh = max(maxh, t)
								}
							}
							var x, y = 3 * maxh, m * 2
							for i := 0; i < n; i++ {
								for j := 0; j < m; j++ {
									t, _ := strconv.Atoi(nums[i * n + j])
									for k := 0; k < t; k++ {
										sets(x, y, id)
										x -= 3
									}
									x += t * 3
									y += 4
								}
								y -= 4 * n
								x += 2
								y -= 2
							}
							replyText = print(id)
						}	
					default:
						replyText = hello[r.Intn(len(hello))]
				}
		}
		
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, replyText)
		bot.Send(msg)
	}	
}
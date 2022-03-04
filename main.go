//
// Created By https://github.com/chenghaopeng.
//

package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var EAI_SESS = os.Getenv("EAI_SESS")
var ROOM = strings.ReplaceAll(os.Getenv("ROOM"), " ", "")
var TGID = os.Getenv("TGID")
var TGBOTID = os.Getenv("TGBOTID")
var tofile = os.Getenv("TOFILE")

func getFloatFromFile(filename string) (float64, error) {
	bs, err := ioutil.ReadFile(filename)
	if err != nil {
		return 0, err
	}

	s := string(bs)
	if s[len(s)-1] == '\n' {
		s = s[:len(s)-1]
	}
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, err
	}

	return f, nil
}

func sendMessage(content string) (string, error) {
	url := "https://api.telegram.org/bot" + TGBOTID + "/sendMessage?chat_id=" + TGID + "&text=" + content
	request, err := http.Get(url)
	if err != nil {
		return "", err
	}

	bodyByte, err := ioutil.ReadAll(request.Body)
	if err != nil {
		return "", err
	}

	body := string(bodyByte)
	return body, nil
}

func doGet(url string, headers map[string]string) (string, error) {
	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	for k, v := range headers {
		request.Header.Add(k, v)
	}
	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()
	bodyByte, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return "", err
	}
	body := string(bodyByte)
	return body, nil
}

func getElectricBalance(areaId string, buildId string, roomId string) (float64, error) {
	for i := 0; i < 5; i++ {
		url := fmt.Sprintf("https://wx.nju.edu.cn/njucharge/wap/electric/charge?area_id=%s&build_id=%s&room_id=%s", areaId, buildId, roomId)

		body, err := doGet(url, map[string]string{
			"Cookie": "eai-sess=" + EAI_SESS,
		})

		if err != nil {
			if i == 4 {
				return 0, err
			}
			time.Sleep(5 * time.Second)
			continue
		}

		reg := regexp.MustCompile(`dianyue\w*:\w*"([\d\.]+)"`)
		result := reg.FindStringSubmatch(body)

		if len(result) != 2 {
			if i == 4 {
				return 0, errors.New("EAI_SESS 失效了")
			}
			time.Sleep(5 * time.Second)
			continue
		}

		return strconv.ParseFloat(result[1], 64)
	}
	panic(errors.New("unable to reach here"))
}

func getElectricInfo() {
	ids := strings.SplitN(ROOM, ",", 3)

	if len(ids) != 3 {
		sendMessage("寝室未配置或配置错误，或未配置 EAI_SESS")
		panic(errors.New("寝室未配置或配置错误"))
	}

	electric, err := getElectricBalance(ids[0], ids[1], ids[2])

	if err != nil {
		sendMessage("获取寝室电量出错了：" + err.Error())
		panic(err)
	}

	log.Println("electric left:", electric, err)

	if electric <= 15 {
		sendMessage("寝室电量不足 15 度啦～")
	}

	old_electric, err := getFloatFromFile(tofile)
	if err != nil {
		sendMessage("从文件读取电量出错了：" + err.Error())
		panic(err)
	}

	if old_electric < electric {
		money := int((electric - old_electric) * 0.58)
		sendMessage("充值成功！上次记录的电量：" + strconv.FormatFloat(old_electric, 'E', -1, 64) + "，现在电量：" + strconv.FormatFloat(electric, 'E', -1, 64) + "，可能的充值金额大约为" + strconv.Itoa(money))
	}

	err = ioutil.WriteFile(tofile, []byte(strconv.FormatFloat(electric, 'E', -1, 64)), 0644)

	if err != nil {
		sendMessage("写入文件出错了：" + err.Error())
		panic(err)
	}

	log.Println("Program successfully executed")
}

func main() {
	getElectricInfo()
}
